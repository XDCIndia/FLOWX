// Package chain defines the chain-agnostic interface that FlowX settlement
// backends implement. See docs/xdc-migration-plan.md.
//
// Status: seeded by the XDC Apothem migration (Phase 0/2). The existing
// internal/stellar package is the reference implementation to be adapted
// onto this interface in Phase 1.
package chain

import (
	"context"
	"math/big"
)

// AssetRef identifies an asset on a chain. For EVM chains, ContractAddress
// is the ERC-20 address; the zero address means the native token. For
// Stellar, Issuer+Code identify the asset. Decimals is authoritative for
// all amount math at the interface boundary (see risk ENG-R3 in the
// migration plan).
type AssetRef struct {
	Symbol          string
	Decimals        uint8
	ContractAddress string // EVM: 0x… ("" = native). Stellar: "".
	Issuer          string // Stellar issuer; empty for native XLM/XDC.
}

// NativeTXDC is the Apothem testnet native asset (test XDC).
var NativeTXDC = AssetRef{Symbol: "TXDC", Decimals: 18}

// NativeXDC is the XDC mainnet native asset.
var NativeXDC = AssetRef{Symbol: "XDC", Decimals: 18}

// ChainClient is the minimal surface FlowX's wallet/transfer/reconcile
// services need from a settlement chain.
type ChainClient interface {
	// GenerateKeypair returns (address, privateKey, error). Address is the
	// chain-native form (EVM 0x…, Stellar G…).
	GenerateKeypair() (string, string, error)

	// Balance returns the balance of addr for asset, in the asset's base
	// units (10^-Decimals of one whole token).
	Balance(ctx context.Context, addr string, asset AssetRef) (*big.Int, error)

	// Transfer sends amount (base units) of asset from the wallet unlocked
	// by privateKey to toAddr. Returns the chain tx hash.
	Transfer(ctx context.Context, privateKey, toAddr string, asset AssetRef, amount *big.Int) (string, error)

	// Confirmations returns the number of confirmations txHash has, or an
	// error if it is not yet mined.
	Confirmations(ctx context.Context, txHash string) (uint64, error)

	// ChainID of the connected network. Callers must verify it matches
	// their configured expectation before signing anything.
	ChainID(ctx context.Context) (*big.Int, error)
}
