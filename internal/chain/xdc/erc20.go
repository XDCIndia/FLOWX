package xdc

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

	"github.com/fluxa/fluxa/internal/chain"
)

// erc20ABI is a minimal hand-written ERC-20 binding covering the surface
// FlowX needs (balanceOf / transfer / decimals). Using abi.JSON directly
// avoids a code-generation step (abigen/solc) in the build.
const erc20ABI = `[
	{"name":"balanceOf","type":"function","stateMutability":"view",
	 "inputs":[{"name":"account","type":"address"}],
	 "outputs":[{"name":"","type":"uint256"}]},
	{"name":"transfer","type":"function","stateMutability":"nonpayable",
	 "inputs":[{"name":"to","type":"address"},{"name":"amount","type":"uint256"}],
	 "outputs":[{"name":"","type":"bool"}]},
	{"name":"decimals","type":"function","stateMutability":"view",
	 "inputs":[],
	 "outputs":[{"name":"","type":"uint8"}]}
]`

var parsedERC20 = mustParseABI()

func mustParseABI() abi.ABI {
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		panic(fmt.Sprintf("xdc: embedded ERC-20 ABI is invalid: %v", err))
	}
	return parsed
}

// encodeBalanceOf returns the calldata for balanceOf(addr).
func encodeBalanceOf(addr string) ([]byte, error) {
	data, err := parsedERC20.Pack("balanceOf", common.HexToAddress(normalizeAddress(addr)))
	if err != nil {
		return nil, fmt.Errorf("xdc: encode balanceOf: %w", err)
	}
	return data, nil
}

// decodeBalanceResult decodes an eth_call result into base units.
func decodeBalanceResult(data []byte) (*big.Int, error) {
	out, err := parsedERC20.Unpack("balanceOf", data)
	if err != nil {
		return nil, fmt.Errorf("xdc: decode balanceOf result: %w", err)
	}
	if len(out) != 1 {
		return nil, fmt.Errorf("xdc: balanceOf returned %d values, want 1", len(out))
	}
	bal, ok := out[0].(*big.Int)
	if !ok {
		return nil, fmt.Errorf("xdc: balanceOf result has unexpected type %T", out[0])
	}
	return bal, nil
}

// encodeTransfer returns the calldata for transfer(to, amount).
func encodeTransfer(toAddr string, amount *big.Int) ([]byte, error) {
	data, err := parsedERC20.Pack("transfer", common.HexToAddress(normalizeAddress(toAddr)), amount)
	if err != nil {
		return nil, fmt.Errorf("xdc: encode transfer: %w", err)
	}
	return data, nil
}

// decodeTransferResult decodes a transfer receipt status flag.
func decodeTransferResult(data []byte) (bool, error) {
	out, err := parsedERC20.Unpack("transfer", data)
	if err != nil {
		return false, fmt.Errorf("xdc: decode transfer result: %w", err)
	}
	if len(out) != 1 {
		return false, fmt.Errorf("xdc: transfer returned %d values, want 1", len(out))
	}
	ok, isBool := out[0].(bool)
	if !isBool {
		return false, fmt.Errorf("xdc: transfer result has unexpected type %T", out[0])
	}
	return ok, nil
}

// encodeDecimals returns the calldata for decimals().
func encodeDecimals() ([]byte, error) {
	data, err := parsedERC20.Pack("decimals")
	if err != nil {
		return nil, fmt.Errorf("xdc: encode decimals: %w", err)
	}
	return data, nil
}

// decodeDecimalsResult decodes an eth_call result into a decimals value.
func decodeDecimalsResult(data []byte) (uint8, error) {
	out, err := parsedERC20.Unpack("decimals", data)
	if err != nil {
		return 0, fmt.Errorf("xdc: decode decimals result: %w", err)
	}
	if len(out) != 1 {
		return 0, fmt.Errorf("xdc: decimals returned %d values, want 1", len(out))
	}
	d, ok := out[0].(uint8)
	if !ok {
		return 0, fmt.Errorf("xdc: decimals result has unexpected type %T", out[0])
	}
	return d, nil
}

// transferERC20 sends amount base units of an ERC-20 token from the wallet
// unlocked by privateKey to toAddr, returning the chain tx hash. Gas is
// estimated against pending state with a fixed ceiling; the signed tx is
// submitted as a plain legacy contract call (value 0, data = transfer()).
func (c *Client) transferERC20(ctx context.Context, privateKey, toAddr string, asset chain.AssetRef, amount *big.Int) (string, error) {
	data, err := encodeTransfer(toAddr, amount)
	if err != nil {
		return "", err
	}
	sk, err := crypto.ToECDSA(common.FromHex(privateKey))
	if err != nil {
		return "", fmt.Errorf("xdc: parse private key: %w", err)
	}
	from := crypto.PubkeyToAddress(sk.PublicKey)
	to := common.HexToAddress(normalizeAddress(asset.ContractAddress))

	gas, err := c.ec.EstimateGas(ctx, ethereum.CallMsg{
		From: from,
		To:   &to,
		Data: data,
	})
	if err != nil {
		// Conservative fallback: a plain ERC-20 transfer fits in ~65k.
		gas = 100000
	}
	gp, err := c.ec.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("xdc: gas price: %w", err)
	}
	nonce, err := c.ec.PendingNonceAt(ctx, from)
	if err != nil {
		return "", fmt.Errorf("xdc: nonce: %w", err)
	}

	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &to,
		Value:    big.NewInt(0),
		Gas:      gas,
		GasPrice: gp,
		Data:     data,
	})
	signed, err := types.SignTx(tx, c.signer, sk)
	if err != nil {
		return "", fmt.Errorf("xdc: sign: %w", err)
	}
	if err := c.ec.SendTransaction(ctx, signed); err != nil {
		return "", fmt.Errorf("xdc: submit: %w", err)
	}
	return signed.Hash().Hex(), nil
}

// Decimals queries the token contract at contractAddr for its decimals.
// Returns 0 if the contract does not expose the (optional) ERC-20 decimals
// method, so callers can fall back to a configured value.
func (c *Client) Decimals(ctx context.Context, contractAddr string) (uint8, error) {
	data, err := encodeDecimals()
	if err != nil {
		return 0, err
	}
	to := common.HexToAddress(normalizeAddress(contractAddr))
	res, err := c.ec.CallContract(ctx, ethereum.CallMsg{To: &to, Data: data}, nil)
	if err != nil {
		return 0, fmt.Errorf("xdc: decimals call %s: %w", contractAddr, err)
	}
	if len(res) == 0 {
		return 0, nil // optional method not implemented
	}
	return decodeDecimalsResult(res)
}
