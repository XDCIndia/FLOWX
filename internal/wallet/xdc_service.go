package wallet

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
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
//     Phase 4+ scope (FlowXUSD ERC-20).
//   - The Stellar-specific signer hook is unused; signing happens inside
//     the chain client with the decrypted wallet secret.
// DepositRecorder is the narrow view of the transfer repository that
// VerifyDeposit needs — avoids an import cycle with the transfer package.
type DepositRecorder interface {
	Create(ctx context.Context, tx *domain.Transaction) error
	ExistsByTxHash(ctx context.Context, txHash string) (bool, error)
	GetByID(ctx context.Context, id string) (*domain.Transaction, error)
}

type XDCService struct {
	repo       Repository
	txRepo     DepositRecorder
	chain      *xdc.Client
	masterKey  []byte
	treasurySK string // hex private key; funds new wallets. Empty = manual faucet per wallet.
	tenantRepo TenantGetter
}

// NewXDCService builds the XDC-backed wallet service.
func NewXDCService(repo Repository, txRepo DepositRecorder, chainClient *xdc.Client, masterKey []byte, treasurySecretKey string, tenantRepo ...TenantGetter) Service {
	s := &XDCService{
		repo:       repo,
		txRepo:     txRepo,
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

	// Fetch on-chain TXDC balance
	balances := []Balance{}
	wei, err := s.chain.Balance(ctx, w.PublicKey, chain.NativeTXDC)
	if err == nil {
		amt := weiToTXDC(wei)
		if err := s.repo.UpsertBalance(ctx, walletID, chain.NativeTXDC.Symbol, "", amt); err != nil {
			log.Debug().Err(err).Str("wallet_id", walletID).Msg("xdc: cache balance upsert failed")
		}
		balances = append(balances, Balance{AssetCode: chain.NativeTXDC.Symbol, Balance: amt.String()})
	}

	// Also include DB-only balances (faucet test tokens like USDC)
	dbBalances, err := s.repo.GetBalances(ctx, walletID)
	if err == nil {
		for _, db := range dbBalances {
			// Skip if already added from on-chain
			found := false
			for _, b := range balances {
				if b.AssetCode == db.AssetCode {
					found = true
					break
				}
			}
			if !found {
				balances = append(balances, Balance{AssetCode: db.AssetCode, Balance: db.Balance})
			}
		}
	}

	return balances, nil
}

// AddTrustline is Stellar-specific (trustlines do not exist on EVM chains).
func (s *XDCService) AddTrustline(_ context.Context, _, assetCode, _, _ string) (string, error) {
	return "", fmt.Errorf("%w: trustlines are a Stellar concept; asset %s needs no trustline on XDC", domain.ErrInvalidAsset, assetCode)
}

// VerifyDeposit checks an on-chain tx hash, confirms it was sent to this
// wallet, and records it as a deposit transaction in the database.
func (s *XDCService) VerifyDeposit(ctx context.Context, walletID, txHash string) (*domain.Transaction, error) {
	w, err := s.repo.GetByID(ctx, walletID)
	if err != nil {
		return nil, fmt.Errorf("load wallet: %w", err)
	}

	deposit, err := s.chain.GetDeposit(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("fetch deposit: %w", err)
	}

	if !deposit.Successful {
		return nil, fmt.Errorf("transaction %s reverted on-chain", txHash)
	}

	// Verify the deposit was sent to this wallet.
	walletAddr := normalizeAddr(w.PublicKey)
	if strings.ToLower(deposit.To) != strings.ToLower(walletAddr) {
		return nil, fmt.Errorf("transaction %s was sent to %s, not to wallet %s", txHash, deposit.To, walletAddr)
	}

	// Check for duplicate.
	exists, _ := s.txRepo.ExistsByTxHash(ctx, txHash)
	if exists {
		existing, _ := s.txRepo.GetByID(ctx, txHash)
		if existing != nil {
			return existing, nil
		}
	}

	// Convert wei to whole TXDC.
	amount := decimal.NewFromBigInt(deposit.Value, 0).Div(decimal.NewFromInt(1e18))

	fromAddr := deposit.From
	tx := &domain.Transaction{
		ID:         uuid.New().String(),
		Type:       domain.TypeTransfer,
		Status:     domain.StatusConfirmed,
		FromWallet: walletID, // external deposit — use the wallet ID as both from/to
		ToWallet:   walletID,
		Asset:      "TXDC",
		Amount:     amount,
		Fee:        decimal.Zero,
		FeeBps:     0,
		TxHash:     txHash,
		CreatedAt:  time.Now().UTC(),
	}
	_ = fromAddr // available for future use (external sender tracking)

	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("record deposit: %w", err)
	}

	// Sync balance after recording.
	wei, balErr := s.chain.Balance(ctx, w.PublicKey, chain.NativeTXDC)
	if balErr == nil {
		amt := weiToTXDC(wei)
		_ = s.repo.UpsertBalance(ctx, walletID, chain.NativeTXDC.Symbol, "", amt)
	}

	log.Info().Str("wallet_id", walletID).Str("tx_hash", txHash).Str("amount", amount.String()).
		Msg("xdc deposit verified and recorded")

	return tx, nil
}

// normalizeAddr strips xdc/0x prefix and returns lowercase hex.
func normalizeAddr(addr string) string {
	a := strings.ToLower(addr)
	if strings.HasPrefix(a, "xdc") {
		a = "0x" + a[3:]
	}
	return a
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

func (s *XDCService) List(ctx context.Context) ([]*domain.Wallet, error) {
	return s.repo.List(ctx, 100, 0)
}

func (s *XDCService) Delete(ctx context.Context, walletID string) error {
	return s.repo.Delete(ctx, walletID)
}


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

func (s *XDCService) Faucet(ctx context.Context, walletID, assetCode string, amount decimal.Decimal) (*FaucetResult, error) {
	// Verify wallet exists
	w, err := s.repo.GetByID(ctx, walletID)
	if err != nil || w == nil {
		return nil, fmt.Errorf("wallet not found")
	}

	// For TXDC: send real on-chain tokens from treasury
	if assetCode == "TXDC" || assetCode == "XDC" {
		if s.treasurySK == "" {
			return nil, fmt.Errorf("treasury key not configured")
		}
		wei := txdcToWei(amount)
		hash, err := s.chain.Transfer(ctx, s.treasurySK, w.PublicKey, chain.NativeTXDC, wei)
		if err != nil {
			return nil, fmt.Errorf("treasury transfer failed: %w", err)
		}
		log.Info().Str("wallet_id", walletID).Str("tx_hash", hash).Msg("faucet: sent real TXDC from treasury")
		// Return the on-chain balance
		onChainWei, _ := s.chain.Balance(ctx, w.PublicKey, chain.NativeTXDC)
		return &FaucetResult{Balance: weiToTXDC(onChainWei).String(), TxHash: hash}, nil
	}

	// For other assets (USDC, EURC, etc.): add to DB balance
	balances, err := s.repo.GetBalances(ctx, walletID)
	if err != nil {
		balances = nil
	}

	var currentBalance decimal.Decimal
	for _, b := range balances {
		if b.AssetCode == assetCode {
			currentBalance, _ = decimal.NewFromString(b.Balance)
			break
		}
	}

	newBalance := currentBalance.Add(amount)
	if err := s.repo.UpsertBalance(ctx, walletID, assetCode, "", newBalance); err != nil {
		return nil, fmt.Errorf("failed to update balance: %w", err)
	}

	return &FaucetResult{Balance: newBalance.String()}, nil
}

