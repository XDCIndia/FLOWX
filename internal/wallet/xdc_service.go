package wallet

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"github.com/fluxa/fluxa/internal/chain"
	"github.com/fluxa/fluxa/internal/chain/xdc"
	"github.com/fluxa/fluxa/internal/crypto"
	"github.com/fluxa/fluxa/internal/domain"
	"github.com/fluxa/fluxa/internal/stellar"
	"github.com/fluxa/fluxa/internal/tenant"
)

// XDC funding parameters for the Apothem testnet model. The treasury funds
// every newly created wallet so it can pay gas and receive/settle transfers
// without a manual faucet step per wallet.
const (
	// xdcTreasuryFundAmount is how much TXDC a new wallet receives.
	xdcTreasuryFundAmount = 2 // TXDC
	// xdcFundGasBuffer is headroom kept in the funding tx for gas.
	xdcFundGasBufferTXDC = 0.01
)

// XDCService implements Service for the XDC (EVM) backend. It mirrors the
// custodial Stellar service but replaces keypair generation, balance reads
// and settlement with the chain abstraction client, and funds new wallets
// from a treasury key instead of Stellar's friendbot/create-account flow.
//
// Status: testnet working model (see docs/xdc-migration-plan.md). Notable
// deviations from the Stellar service:
//   - AddTrustline is a Stellar concept; it returns an error here.
//   - Only the native asset (TXDC) is supported; USDC/EURC settlement is
//     Phase 4+ scope (FluxaUSD ERC-20).
//   - The Stellar-specific signer hook is unused; signing happens inside
//     the chain client with the decrypted wallet secret.
type XDCService struct {
	repo       Repository
	chain      *xdc.Client
	masterKey  []byte
	treasurySK string // hex private key; funds new wallets. Empty = manual faucet per wallet.
	tenantRepo TenantGetter
}

// NewXDCService builds the XDC-backed wallet service.
func NewXDCService(repo Repository, chainClient *xdc.Client, masterKey []byte, treasurySecretKey string, tenantRepo ...TenantGetter) Service {
	s := &XDCService{
		repo:       repo,
		chain:      chainClient,
		masterKey:  masterKey,
		treasurySK: treasurySecretKey,
	}
	if len(tenantRepo) > 0 {
		s.tenantRepo = tenantRepo[0]
	}
	return s
}

// CreateWallet generates an XDC keypair, encrypts the secret with the master
// key (same custody model as the Stellar backend), persists the wallet, then
// funds it from the treasury so it can transact immediately.
func (s *XDCService) CreateWallet(ctx context.Context, _ ...string) (*domain.Wallet, error) {
	tenantID := tenant.IDFromContext(ctx)
	if tenantID != "" && s.tenantRepo != nil {
		t, err := s.tenantRepo.GetByID(ctx, tenantID)
		if err == nil && t != nil {
			limit := t.GetWalletLimit()
			count, err := s.repo.CountByTenant(ctx, tenantID)
			if err == nil && count >= limit {
				return nil, domain.ErrWalletLimitReached
			}
		}
	}

	pubKey, secretKey, err := s.chain.GenerateKeypair()
	if err != nil {
		return nil, fmt.Errorf("generate keypair: %w", err)
	}

	encryptedBytes, err := crypto.Encrypt([]byte(secretKey), s.masterKey)
	if err != nil {
		return nil, fmt.Errorf("encrypt secret: %w", err)
	}

	w := &domain.Wallet{
		ID:              uuid.New().String(),
		PublicKey:       pubKey,
		EncryptedSecret: hex.EncodeToString(encryptedBytes),
		CustodyType:     domain.CustodyCustodial,
		CreatedAt:       time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, w); err != nil {
		return nil, fmt.Errorf("persist wallet: %w", err)
	}

	if err := s.fundFromTreasury(ctx, w); err != nil {
		// Funding failure must not roll back wallet creation — the wallet is
		// valid and can be funded later (manual faucet at
		// https://faucet.apothem.network/). Log loudly for ops.
		log.Warn().Err(err).Str("wallet_id", w.ID).Str("address", w.PublicKey).
			Msg("xdc wallet created but treasury funding failed; fund manually via Apothem faucet")
	}

	return w, nil
}

// fundFromTreasury sends a small amount of native TXDC from the treasury key
// to a freshly created wallet so it has gas money and can receive transfers.
func (s *XDCService) fundFromTreasury(ctx context.Context, w *domain.Wallet) error {
	if s.treasurySK == "" {
		log.Info().Str("address", w.PublicKey).
			Msg("XDC_TREASURY_SECRET_KEY not set; fund this wallet manually at https://faucet.apothem.network/")
		return nil
	}
	fundAmt := decimal.NewFromFloat(xdcTreasuryFundAmount - xdcFundGasBufferTXDC)
	wei := txdcToWei(fundAmt)
	hash, err := s.chain.Transfer(ctx, s.treasurySK, w.PublicKey, chain.NativeTXDC, wei)
	if err != nil {
		return fmt.Errorf("treasury funding transfer: %w", err)
	}
	log.Info().Str("wallet_id", w.ID).Str("address", w.PublicKey).Str("fund_tx", hash).
		Msg("xdc wallet funded from treasury")
	return nil
}

func (s *XDCService) GetWalletForHandler(ctx context.Context, walletID string) (*domain.Wallet, error) {
	return s.repo.GetByID(ctx, walletID)
}

// GetBalances returns the on-chain native TXDC balance. Stellar-style
// multi-asset balances and FX conversion are not applicable on the XDC
// backend yet (single native asset on the testnet model).
func (s *XDCService) GetBalances(ctx context.Context, walletID string, _ ...string) ([]Balance, error) {
	w, err := s.repo.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}
	wei, err := s.chain.Balance(ctx, w.PublicKey, chain.NativeTXDC)
	if err != nil {
		return nil, fmt.Errorf("xdc balance: %w", err)
	}
	amt := weiToTXDC(wei)
	if err := s.repo.UpsertBalance(ctx, walletID, chain.NativeTXDC.Symbol, "", amt); err != nil {
		log.Debug().Err(err).Str("wallet_id", walletID).Msg("xdc: cache balance upsert failed")
	}
	return []Balance{{AssetCode: chain.NativeTXDC.Symbol, Balance: amt.String()}}, nil
}

// AddTrustline is Stellar-specific (trustlines do not exist on EVM chains).
func (s *XDCService) AddTrustline(_ context.Context, _, assetCode, _, _ string) (string, error) {
	return "", fmt.Errorf("%w: trustlines are a Stellar concept; asset %s needs no trustline on XDC", domain.ErrInvalidAsset, assetCode)
}

// ExecuteTransfer settles a payment directly on-chain: decrypt the wallet
// secret in memory, sign, submit. Amount is in whole TXDC units.
func (s *XDCService) ExecuteTransfer(
	ctx context.Context,
	walletID, destination, assetCode, _ string,
	amount decimal.Decimal,
	_ string,
) (string, error) {
	w, err := s.repo.GetByID(ctx, walletID)
	if err != nil {
		return "", err
	}
	if !isNativeTXDC(assetCode) {
		return "", fmt.Errorf("%w: asset %s is not supported on the XDC backend yet (native TXDC only)", domain.ErrInvalidAsset, assetCode)
	}

	encryptedSecret, err := hex.DecodeString(w.EncryptedSecret)
	if err != nil {
		return "", fmt.Errorf("decode encrypted secret: %w", err)
	}
	secretBytes, err := crypto.Decrypt(encryptedSecret, s.masterKey)
	if err != nil {
		return "", fmt.Errorf("decrypt wallet secret: %w", err)
	}

	return s.chain.Transfer(ctx, string(secretBytes), destination, chain.NativeTXDC, txdcToWei(amount))
}

// WithSigner is Stellar-specific signing hooks; unused on XDC.
func (s *XDCService) WithSigner(_ stellar.Signer) Service { return s }

// WithFXService keeps interface parity; FX conversion is not wired on XDC yet.
func (s *XDCService) WithFXService(_ FXRateGetter) Service { return s }

// WithIssuers keeps interface parity; issuer resolution is Stellar-specific.
func (s *XDCService) WithIssuers(_, _ string) Service { return s }

// isNativeTXDC reports whether an asset code refers to the XDC native asset.
func isNativeTXDC(code string) bool {
	return code == "" || code == "TXDC" || code == "XDC"
}

// txdcToWei converts whole-unit TXDC decimal to base units (1e18).
func txdcToWei(d decimal.Decimal) *big.Int {
	v := d.Mul(decimal.NewFromInt(1e18)).BigInt()
	return v
}

// weiToTXDC converts base units to a whole-unit decimal.
func weiToTXDC(wei *big.Int) decimal.Decimal {
	return decimal.NewFromBigInt(wei, 0).Div(decimal.NewFromInt(1e18))
}
