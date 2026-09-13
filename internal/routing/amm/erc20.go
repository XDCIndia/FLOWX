package amm

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// erc20ApproveABI is a minimal hand-rolled ABI covering approve() — the one
// ERC-20 surface FlowXPool needs beyond what the pool binding itself does
// (treasury must approve the pool to pull tUSDC on swapExactTokensForTXDC).
const erc20ApproveABI = `[
	{"name":"approve","type":"function","stateMutability":"nonpayable",
	 "inputs":[{"name":"spender","type":"address"},{"name":"amount","type":"uint256"}],
	 "outputs":[{"name":"","type":"bool"}]}
]`

var parsedApproveABI = mustParseApproveABI()

func mustParseApproveABI() abi.ABI {
	parsed, err := abi.JSON(strings.NewReader(erc20ApproveABI))
	if err != nil {
		panic(fmt.Sprintf("amm: embedded approve ABI is invalid: %v", err))
	}
	return parsed
}

// Approve submits an ERC-20 approve(spender, amount) from the wallet unlocked
// by privateKeyHex (raw hex, optional 0x prefix), using the given nonce.
// Returns the tx hash. Gas is capped conservatively (a plain approve fits in
// ~50k) with a suggested gas price; the tx is a plain legacy call.
func Approve(ctx context.Context, ec *ethclient.Client, signer types.Signer, privateKeyHex, tokenAddr, spender string, amount *big.Int, nonce uint64) (string, error) {
	sk, err := crypto.ToECDSA(common.FromHex(privateKeyHex))
	if err != nil {
		return "", fmt.Errorf("amm approve: parse private key: %w", err)
	}
	from := crypto.PubkeyToAddress(sk.PublicKey)

	data, err := parsedApproveABI.Pack("approve", common.HexToAddress(normalize(spender)), amount)
	if err != nil {
		return "", fmt.Errorf("amm approve: pack: %w", err)
	}
	to := common.HexToAddress(normalize(tokenAddr))

	gas, err := ec.EstimateGas(ctx, ethereum.CallMsg{From: from, To: &to, Data: data})
	if err != nil {
		gas = 100000 // conservative fallback; approve fits in ~50k
	}
	gp, err := ec.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("amm approve: gas price: %w", err)
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &to,
		Value:    big.NewInt(0),
		Gas:      gas,
		GasPrice: gp,
		Data:     data,
	})
	signed, err := types.SignTx(tx, signer, sk)
	if err != nil {
		return "", fmt.Errorf("amm approve: sign: %w", err)
	}
	if err := ec.SendTransaction(ctx, signed); err != nil {
		return "", fmt.Errorf("amm approve: submit: %w", err)
	}
	return signed.Hash().Hex(), nil
}

// normalize converts an xdc-prefixed address to the 0x form go-ethereum
// expects. Duplicated deliberately: this package must not import the token
// agent's internal/chain/xdc/tokens tree.
func normalize(addr string) string {
	if len(addr) >= 3 && strings.EqualFold(addr[:3], "xdc") {
		return "0x" + addr[3:]
	}
	return addr
}
