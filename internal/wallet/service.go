package wallet

import (
	"context"

	"github.com/fluxa/fluxa/internal/domain"
	"github.com/shopspring/decimal"
)

type TenantGetter interface {
	GetByID(ctx context.Context, id string) (*domain.Tenant, error)
}

type FXRateGetter interface {
	GetRates(ctx context.Context, from, to string) (*domain.RateResponse, error)
}

// FaucetResult holds the outcome of a faucet request.
type FaucetResult struct {
	Balance string `json:"balance"`
	TxHash  string `json:"tx_hash,omitempty"`
}

type Balance struct {
	AssetCode     string `json:"asset_code"`
	Issuer        string `json:"issuer"`
	Balance       string `json:"balance"`
	USDEquivalent string `json:"usd_equivalent,omitempty"`
}

type Service interface {
	// CreateWallet provisions a custodial wallet: a fresh on-chain keypair
	// whose secret is encrypted with the platform master key.
	CreateWallet(ctx context.Context, ownerPublicKey ...string) (*domain.Wallet, error)
	GetWalletForHandler(ctx context.Context, walletID string) (*domain.Wallet, error)
	GetBalances(ctx context.Context, walletID string, includeFX ...string) ([]Balance, error)
	// AddTrustline is a legacy Stellar concept; on XDC it always returns an
	// error explaining trustlines do not exist on EVM chains.
	AddTrustline(ctx context.Context, walletID, assetCode, issuer, limit string) (string, error)
	// ExecuteTransfer moves an asset out of the wallet and returns the
	// transaction hash.
	ExecuteTransfer(ctx context.Context, walletID, destination, assetCode, issuer string, amount decimal.Decimal, memo string) (string, error)
	// VerifyDeposit checks an on-chain tx hash, confirms it was sent to this
	// wallet, and records it as a deposit transaction. Returns the verified
	// deposit details.
	VerifyDeposit(ctx context.Context, walletID, txHash string) (*domain.Transaction, error)
	WithFXService(fxSvc FXRateGetter) Service
	WithIssuers(usdcIssuer, eurcIssuer string) Service
	Delete(ctx context.Context, walletID string) error
	List(ctx context.Context) ([]*domain.Wallet, error)
	Faucet(ctx context.Context, walletID, assetCode string, amount decimal.Decimal) (*FaucetResult, error)
}
