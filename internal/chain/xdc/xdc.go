// Package xdc implements chain.ChainClient for XDC Network (EVM, XDPoS)
// via JSON-RPC. Developed against Apothem testnet (chain ID 51).
package xdc

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/fluxa/fluxa/internal/chain"
)

// ApothemChainID is the XDC Apothem testnet chain ID.
const ApothemChainID = 51

const (
	// DefaultRPC is the public Apothem endpoint. Rate-limited — fine for
	// dev/spike work (see migration plan §9).
	DefaultRPC = "https://rpc.apothem.network"
	// ExplorerBase for human-readable proof links.
	ExplorerBase = "https://apothem.blocksscan.io"
)

// Client is an XDC chain client satisfying chain.ChainClient.
type Client struct {
	rpc    string
	ec     *ethclient.Client
	chain  *big.Int
	signer types.Signer
}

// New dials rpcURL and verifies the remote chain ID matches expectedChainID.
// This guard exists so a misconfigured RPC can never get a signed tx for the
// wrong network.
func New(ctx context.Context, rpcURL string, expectedChainID int64) (*Client, error) {
	ec, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("xdc: dial %s: %w", rpcURL, err)
	}
	id, err := ec.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("xdc: chainId: %w", err)
	}
	if id.Int64() != expectedChainID {
		return nil, fmt.Errorf("xdc: chainId mismatch: RPC says %d, want %d — refusing to use this endpoint", id.Int64(), expectedChainID)
	}
	return &Client{
		rpc:    rpcURL,
		ec:     ec,
		chain:  id,
		signer: types.NewEIP155Signer(id),
	}, nil
}

// normalizeAddress converts an XDC-prefixed address (xdc1f3c…) or an
// Ethereum-prefixed one (0x1f3c…) into the 0x form go-ethereum expects.
func normalizeAddress(addr string) string {
	if len(addr) >= 3 && strings.EqualFold(addr[:3], "xdc") {
		return "0x" + addr[3:]
	}
	return addr
}

// toXDCAddress renders any 0x/xdc address as a lowercase, XDC-prefixed
// address (xdc1f3c…). XDC Network uses lowercase addresses, not EIP-55
// checksummed ones.
func toXDCAddress(addr string) string {
	checksummed := common.HexToAddress(normalizeAddress(addr)).Hex()
	return "xdc" + strings.ToLower(checksummed[2:])
}

// GenerateKeypair generates a secp256k1 keypair. Returns (xdcAddress,
// privateKeyHex) — the public key carries the XDC prefix per network
// convention (same EIP-55 checksum as 0x, prefix swapped).
func (c *Client) GenerateKeypair() (string, string, error) {
	sk, err := crypto.GenerateKey()
	if err != nil {
		return "", "", fmt.Errorf("xdc: generate key: %w", err)
	}
	return toXDCAddress(crypto.PubkeyToAddress(sk.PublicKey).Hex()), fmt.Sprintf("%x", crypto.FromECDSA(sk)), nil
}

// Balance returns the balance in base units (wei for native XDC).
// ERC-20 support (asset.ContractAddress != "") is Phase 4 scope; only the
// native asset is implemented for the working model. Accepts xdc- or
// 0x-prefixed addresses.
func (c *Client) Balance(ctx context.Context, addr string, asset chain.AssetRef) (*big.Int, error) {
	if asset.ContractAddress != "" {
		return nil, fmt.Errorf("xdc: ERC-20 balances not implemented yet (asset %s)", asset.Symbol)
	}
	b, err := c.ec.BalanceAt(ctx, common.HexToAddress(normalizeAddress(addr)), nil)
	if err != nil {
		return nil, fmt.Errorf("xdc: balance %s: %w", addr, err)
	}
	return b, nil
}

// Transfer sends amount base units of the native asset. toAddr may be xdc-
// or 0x-prefixed.
func (c *Client) Transfer(ctx context.Context, privateKey, toAddr string, asset chain.AssetRef, amount *big.Int) (string, error) {
	if asset.ContractAddress != "" {
		return "", fmt.Errorf("xdc: ERC-20 transfers not implemented yet (asset %s)", asset.Symbol)
	}
	sk, err := crypto.ToECDSA(common.FromHex(privateKey))
	if err != nil {
		return "", fmt.Errorf("xdc: parse private key: %w", err)
	}
	from := crypto.PubkeyToAddress(sk.PublicKey)

	nonce, err := c.ec.PendingNonceAt(ctx, from)
	if err != nil {
		return "", fmt.Errorf("xdc: nonce: %w", err)
	}
	gp, err := c.ec.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("xdc: gas price: %w", err)
	}
	to := common.HexToAddress(normalizeAddress(toAddr))
	tx := types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		To:       &to,
		Value:    amount,
		Gas:      21000,
		GasPrice: gp,
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

// Confirmations returns how many blocks bury txHash (0 when just mined).
func (c *Client) Confirmations(ctx context.Context, txHash string) (uint64, error) {
	r, err := c.ec.TransactionReceipt(ctx, common.HexToHash(txHash))
	if err != nil {
		return 0, fmt.Errorf("xdc: receipt %s: %w", txHash, err)
	}
	if r.Status != types.ReceiptStatusSuccessful {
		return 0, fmt.Errorf("xdc: tx %s reverted", txHash)
	}
	head, err := c.ec.HeaderByNumber(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("xdc: head: %w", err)
	}
	if head.Number.Uint64() < r.BlockNumber.Uint64() {
		return 0, nil
	}
	return head.Number.Uint64() - r.BlockNumber.Uint64() + 1, nil
}

// ChainID returns the verified chain ID.
func (c *Client) ChainID(ctx context.Context) (*big.Int, error) { return c.chain, nil }

// WaitConfirmations polls until txHash has at least n confirmations.
// Testnet default per migration plan §6: 6 confirmations (~12s).
func (c *Client) WaitConfirmations(ctx context.Context, txHash string, n uint64) error {
	for {
		confs, err := c.Confirmations(ctx, txHash)
		if err == nil {
			if confs >= n {
				return nil
			}
			fmt.Printf("  confirmations: %d/%d\n", confs, n)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(4 * time.Second):
		}
	}
}

// ExplorerTxURL is a human proof link for logs and the demo output.
func ExplorerTxURL(txHash string) string {
	return ExplorerBase + "/tx/" + strings.TrimPrefix(txHash, "0x")
}

// ExplorerAddressURL is a human proof link for an address.
func ExplorerAddressURL(addr string) string {
	return ExplorerBase + "/address/" + strings.TrimPrefix(strings.TrimPrefix(strings.ToLower(addr), "xdc"), "0x")
}

// sanity: Client satisfies the chain interface at compile time.
var _ chain.ChainClient = (*Client)(nil)

// silence unused warning for ecdsa (kept for future signing variants).
var _ = ecdsa.PublicKey{}
