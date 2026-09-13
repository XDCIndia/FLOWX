package xdc

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

func TestERC20ABIParses(t *testing.T) {
	parsed, err := abi.JSON(bytes.NewReader([]byte(erc20ABI)))
	if err != nil {
		t.Fatalf("embedded ABI must parse: %v", err)
	}
	for _, name := range []string{"balanceOf", "transfer", "decimals"} {
		if _, ok := parsed.Methods[name]; !ok {
			t.Errorf("ABI missing method %s", name)
		}
	}
}

func TestEncodeDecodeBalanceOfRoundtrip(t *testing.T) {
	addr := "xdcA0b86a33E6441e0c40e4bC5d9d7e2F4a81C2dE34"
	data, err := encodeBalanceOf(addr)
	if err != nil {
		t.Fatalf("encodeBalanceOf: %v", err)
	}

	// Method selector (4 bytes) + 32-byte address word.
	if len(data) != 4+32 {
		t.Fatalf("calldata length = %d, want %d", len(data), 4+32)
	}
	wantSelector := parsedERC20.Methods["balanceOf"].ID
	if !bytes.Equal(data[:4], wantSelector[:]) {
		t.Fatalf("selector mismatch: got %x want %x", data[:4], wantSelector)
	}

	// Roundtrip: decode the result the chain would return for 12345 units.
	bal, err := decodeBalanceResult(abiEncodeUint256(big.NewInt(12345)))
	if err != nil {
		t.Fatalf("decodeBalanceResult: %v", err)
	}
	if bal.Cmp(big.NewInt(12345)) != 0 {
		t.Fatalf("balance = %s, want 12345", bal)
	}
}

func TestEncodeDecodeTransferRoundtrip(t *testing.T) {
	to := "0xB1e92B4b7F0a2C3d4E5F60718293a4B5C6d7E8F90"
	amount := big.NewInt(999_000_000)
	data, err := encodeTransfer(to, amount)
	if err != nil {
		t.Fatalf("encodeTransfer: %v", err)
	}
	if len(data) != 4+32+32 {
		t.Fatalf("calldata length = %d, want %d", len(data), 4+64)
	}

	// Roundtrip the return value: transfer() returns bool.
	ok, err := decodeTransferResult(abiEncodeBool(true))
	if err != nil {
		t.Fatalf("decodeTransferResult(true): %v", err)
	}
	if !ok {
		t.Fatal("decoded transfer success flag = false, want true")
	}
	ok, err = decodeTransferResult(abiEncodeBool(false))
	if err != nil {
		t.Fatalf("decodeTransferResult(false): %v", err)
	}
	if ok {
		t.Fatal("decoded transfer success flag = true, want false")
	}
}

func TestEncodeDecodeDecimalsRoundtrip(t *testing.T) {
	if _, err := encodeDecimals(); err != nil {
		t.Fatalf("encodeDecimals: %v", err)
	}
	// ABI encodes uint8 as a full 32-byte word.
	d, err := decodeDecimalsResult(abiEncodeUint256(big.NewInt(6)))
	if err != nil {
		t.Fatalf("decodeDecimalsResult: %v", err)
	}
	if d != 6 {
		t.Fatalf("decimals = %d, want 6", d)
	}
}

func TestNormalizeAddressUsedInEncoding(t *testing.T) {
	xdcAddr := "xdcA0b86a33E6441e0c40e4bC5d9d7e2F4a81C2dE34"
	hexAddr := "0xA0b86a33E6441e0c40e4bC5d9d7e2F4a81C2dE34"

	a, err := encodeBalanceOf(xdcAddr)
	if err != nil {
		t.Fatalf("encodeBalanceOf(xdc): %v", err)
	}
	b, err := encodeBalanceOf(hexAddr)
	if err != nil {
		t.Fatalf("encodeBalanceOf(0x): %v", err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("xdc- and 0x-prefixed addresses must encode identically")
	}
	if !common.IsHexAddress(hexAddr) {
		t.Fatal("fixture address invalid")
	}
}

// abiEncodeUint256 ABI-encodes v as a single uint256 word.
func abiEncodeUint256(v *big.Int) []byte {
	packed, err := parsedERC20.Methods["balanceOf"].Outputs.Pack(v)
	if err != nil {
		panic(err)
	}
	return packed
}

// abiEncodeBool ABI-encodes b as the bool return of transfer().
func abiEncodeBool(b bool) []byte {
	packed, err := parsedERC20.Methods["transfer"].Outputs.Pack(b)
	if err != nil {
		panic(err)
	}
	return packed
}
