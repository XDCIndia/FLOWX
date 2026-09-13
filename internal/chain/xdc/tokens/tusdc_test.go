package tokens

import (
	"context"
	"math/big"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBalanceOfPackUnpack checks the balanceOf calldata encoding and result
// decoding round-trip at the ABI level (no network).
func TestBalanceOfPackUnpack(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(TestUSDCMetaData.ABI))
	require.NoError(t, err)

	addr := common.HexToAddress("0x340904D06fcA218aDDc5aEd7601E500EA26a04E6")
	calldata, err := parsed.Pack("balanceOf", addr)
	require.NoError(t, err)
	// Method selector 0x70a08231 + 32-byte padded address.
	require.Len(t, calldata, 4+32)
	assert.Equal(t, "70a08231", common.Bytes2Hex(calldata[:4]))
	assert.Equal(t, strings.Repeat("00", 12)+common.Bytes2Hex(addr.Bytes()), common.Bytes2Hex(calldata[4:36]))

	want := new(big.Int).Mul(big.NewInt(1_000_000), big.NewInt(1_000_000)) // 1M with 6 decimals
	result, err := parsed.Unpack("balanceOf", abiEncodeUint256(want))
	require.NoError(t, err)
	require.Len(t, result, 1)
	got, ok := result[0].(*big.Int)
	require.True(t, ok, "balanceOf output should be *big.Int")
	assert.Equal(t, 0, want.Cmp(got))
}

// TestTransferPackUnpack checks the transfer calldata encoding round-trip
// (no network).
func TestTransferPackUnpack(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(TestUSDCMetaData.ABI))
	require.NoError(t, err)

	to := common.HexToAddress("0xe069a90d55fdECa2cBec7793712aA3D807fFEF9b")
	amount := big.NewInt(123_456_789)
	calldata, err := parsed.Pack("transfer", to, amount)
	require.NoError(t, err)
	require.Len(t, calldata, 4+32+32)
	assert.Equal(t, "a9059cbb", common.Bytes2Hex(calldata[:4]))
	// First arg: address right-aligned in a 32-byte word (12 bytes padding).
	assert.Equal(t, strings.Repeat("00", 12)+common.Bytes2Hex(to.Bytes()), common.Bytes2Hex(calldata[4:36]))
	assert.Equal(t, amount.String(), new(big.Int).SetBytes(calldata[36:68]).String())

	// transfer returns bool: decoding "0x...1" must give true.
	result, err := parsed.Unpack("transfer", abiEncodeUint256(big.NewInt(1)))
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, true, result[0])
}

// TestDecimalsPackUnpack checks decimals() calldata and uint8 decoding.
func TestDecimalsPackUnpack(t *testing.T) {
	parsed, err := abi.JSON(strings.NewReader(TestUSDCMetaData.ABI))
	require.NoError(t, err)

	calldata, err := parsed.Pack("decimals")
	require.NoError(t, err)
	assert.Equal(t, "313ce567", common.Bytes2Hex(calldata)) // selector only

	// uint8 6 encoded as 32-byte word.
	result, err := parsed.Unpack("decimals", abiEncodeUint256(big.NewInt(6)))
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, uint8(6), result[0])
}

func abiEncodeUint256(v *big.Int) []byte {
	out := make([]byte, 32)
	vb := v.Bytes()
	copy(out[32-len(vb):], vb)
	return out
}

// TestDeployedOnChain is a live verification against Apothem. It only runs
// when TUSDC_VERIFY_ADDR and XDC_RPC_URL (or the defaults) are provided:
//
//	TUSDC_VERIFY_ADDR=0x...  (deployed TestUSDC address)
//	XDC_RPC_URL              (default https://rpc.apothem.network)
//	XDC_DEPLOYER_ADDR        (address expected to hold the full supply)
//	TUSDC_EXPECTED_SUPPLY    (base units; default 1e12 = 1M tUSDC)
func TestDeployedOnChain(t *testing.T) {
	addrHex := os.Getenv("TUSDC_VERIFY_ADDR")
	if addrHex == "" {
		t.Skip("TUSDC_VERIFY_ADDR not set; skipping live verification")
	}
	rpcURL := os.Getenv("XDC_RPC_URL")
	if rpcURL == "" {
		rpcURL = "https://rpc.apothem.network"
	}
	deployer := os.Getenv("XDC_DEPLOYER_ADDR")
	if deployer == "" {
		deployer = "0x340904D06fcA218aDDc5aEd7601E500EA26a04E6"
	}
	wantSupply := new(big.Int)
	if s := os.Getenv("TUSDC_EXPECTED_SUPPLY"); s != "" {
		wantSupply.SetString(s, 10)
	} else {
		wantSupply.Mul(big.NewInt(1_000_000), big.NewInt(1_000_000))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := ethclient.Dial(rpcURL)
	require.NoError(t, err)
	id, err := client.ChainID(ctx)
	require.NoError(t, err)
	assert.EqualValues(t, 51, id.Int64(), "must verify on Apothem (chain ID 51)")

	token, err := NewTestUSDC(common.HexToAddress(addrHex), client)
	require.NoError(t, err)

	code, err := client.CodeAt(ctx, common.HexToAddress(addrHex), nil)
	require.NoError(t, err)
	require.NotEmpty(t, code, "no contract code at %s", addrHex)

	name, err := token.Name(nil)
	require.NoError(t, err)
	assert.Equal(t, "FlowX Test USD", name)

	symbol, err := token.Symbol(nil)
	require.NoError(t, err)
	assert.Equal(t, "tUSDC", symbol)

	decimals, err := token.Decimals(nil)
	require.NoError(t, err)
	assert.Equal(t, uint8(6), decimals)

	total, err := token.TotalSupply(nil)
	require.NoError(t, err)
	assert.Equal(t, 0, wantSupply.Cmp(total), "totalSupply %s != expected %s", total, wantSupply)

	bal, err := token.BalanceOf(nil, common.HexToAddress(deployer))
	require.NoError(t, err)
	assert.Equal(t, 0, wantSupply.Cmp(bal), "deployer balance %s != supply %s", bal, wantSupply)

	// Also verify via a raw eth_call through the Call contract to exercise
	// the pack path end-to-end.
	caller, err := NewTestUSDCCaller(common.HexToAddress(addrHex), client)
	require.NoError(t, err)
	rawBal, err := caller.BalanceOf(&bind.CallOpts{Context: ctx}, common.HexToAddress(deployer))
	require.NoError(t, err)
	assert.Equal(t, 0, wantSupply.Cmp(rawBal))
}
