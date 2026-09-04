package settlement

import (
	"context"
	"encoding/hex"
	"errors"
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
	"github.com/fluxa/fluxa/internal/fees"
	"github.com/fluxa/fluxa/internal/transfer"
	"github.com/fluxa/fluxa/internal/wallet"
)

// xdcRequiredConfirmations is the testnet confirmation policy from
// docs/xdc-migration-plan.md §6 (validated empirically in Phase 0; ~12s at
// 2s blocks).
const xdcRequiredConfirmations = 6

// XDCEngine settles transfers on XDC (EVM) instead of Stellar. It mirrors
// Engine.SubmitTransfer's state machine (claim -> submit -> confirmed) but
// settles the native TXDC asset with an EIP-155 signed EVM transaction.
//
// Documented deviations from the Stellar engine (testnet working model):
//   - The hash is recorded at confirmation time, not before submission —
//     the chain client's Transfer builds, signs and submits atomically, so
//     there is no pre-submit point where the hash is known. Ambiguous
//     outcomes (RPC timeout after a possibly-accepted tx) are marked failed
//     and logged; a production version must poll the mempool by nonce
//     (see migration plan, ENG-R4/R6).
//   - Confirmation is finalized only after xdcRequiredConfirmations blocks
//     (Stellar/Horizon treats horizon-accept as confirmed).
//   - Fee collection is a separate EVM transfer to the platform fee wallet,
//     which must be a 0x address on this backend.
type XDCEngine struct {
	txRepo     transfer.Repository
	walletRepo wallet.Repository
	feeSvc     fees.Service
	chain      *xdc.Client
	masterKey  []byte
	feeWallet  string // platform fee wallet — must be a 0x address on XDC
}

// NewXDCEngine builds the XDC settlement engine.
func NewXDCEngine(
	txRepo transfer.Repository,
	walletRepo wallet.Repository,
	feeSvc fees.Service,
	chainClient *xdc.Client,
	masterKey []byte,
	feeWallet string,
) *XDCEngine {
	return &XDCEngine{
		txRepo:     txRepo,
		walletRepo: walletRepo,
		feeSvc:     feeSvc,
		chain:      chainClient,
		masterKey:  masterKey,
		feeWallet:  feeWallet,
	}
}

// SubmitTransfer claims and settles a pending transfer on XDC.
func (e *XDCEngine) SubmitTransfer(ctx context.Context, txID string) error {
	tx, err := e.txRepo.GetByID(ctx, txID)
	if err != nil {
		return fmt.Errorf("load transaction: %w", err)
	}

	// Same claim semantics as the Stellar engine: exactly one worker may
	// settle a given transfer id.
	if err := e.txRepo.ClaimForSubmission(ctx, txID); err != nil {
		if errors.Is(err, domain.ErrConcurrentUpdate) {
			log.Info().Str("tx_id", txID).Str("status", string(tx.Status)).
				Msg("settlement: transaction already claimed or not eligible for submission")
			return nil
		}
		return fmt.Errorf("claim transaction for submission: %w", err)
	}

	if !isNativeAsset(tx.Asset) {
		if err := e.txRepo.UpdateStatus(ctx, txID, domain.StatusFailed, ""); err != nil {
			log.Error().Err(err).Str("tx_id", txID).Msg("xdc settlement: failed to update failed status")
		}
		return fmt.Errorf("%w: asset %s is not supported on the XDC backend (native TXDC only)", domain.ErrInvalidAsset, tx.Asset)
	}

	srcWallet, err := e.walletRepo.GetByID(ctx, tx.FromWallet)
	if err != nil {
		return fmt.Errorf("load source wallet: %w", err)
	}
	dstWallet, err := e.walletRepo.GetByID(ctx, tx.ToWallet)
	if err != nil {
		return fmt.Errorf("load destination wallet: %w", err)
	}

	srcSecret, err := e.decryptSecret(srcWallet.EncryptedSecret)
	if err != nil {
		return err
	}

	netAmount := tx.NetAmount() // whole TXDC units
	hash, err := e.chain.Transfer(ctx, srcSecret, dstWallet.PublicKey, chain.NativeTXDC, txdcToWeiX(netAmount))
	if err != nil {
		if uErr := e.txRepo.UpdateStatus(ctx, txID, domain.StatusFailed, ""); uErr != nil {
			log.Error().Err(uErr).Str("tx_id", txID).Msg("xdc settlement: failed to update failed status")
		}
		return fmt.Errorf("xdc submit transfer: %w", err)
	}
	log.Info().Str("tx_id", txID).Str("tx_hash", hash).
		Msg("xdc settlement: transfer submitted")

	// Wait for the confirmation policy before marking confirmed.
	confCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	if err := e.chain.WaitConfirmations(confCtx, hash, xdcRequiredConfirmations); err != nil {
		if uErr := e.txRepo.UpdateStatus(ctx, txID, domain.StatusSubmitted, hash); uErr != nil {
			log.Error().Err(uErr).Str("tx_id", txID).Msg("xdc settlement: failed to record submitted status")
		}
		return fmt.Errorf("xdc wait confirmations: %w", err)
	}

	if err := e.txRepo.UpdateStatus(ctx, txID, domain.StatusConfirmed, hash); err != nil {
		log.Error().Err(err).Str("tx_id", txID).Str("tx_hash", hash).
			Msg("xdc settlement: failed to update confirmed status")
	}

	e.syncBalance(ctx, srcWallet)
	e.syncBalance(ctx, dstWallet)

	if tx.Fee.GreaterThan(decimal.Zero) {
		e.collectFee(ctx, tx, txID, srcSecret)
	}

	return nil
}

// collectFee settles the platform fee as a second EVM transfer and records
// the collection. The fee wallet must be a 0x address on this backend.
func (e *XDCEngine) collectFee(ctx context.Context, tx *domain.Transaction, txID, srcSecret string) {
	if !strings.HasPrefix(e.feeWallet, "0x") && !strings.HasPrefix(e.feeWallet, "xdc") {
		log.Warn().Str("tx_id", txID).Str("fee_wallet", e.feeWallet).
			Msg("xdc settlement: platform fee wallet is not a 0x address; skipping on-chain fee collection")
		return
	}
	feeHash, err := e.chain.Transfer(ctx, srcSecret, e.feeWallet, chain.NativeTXDC, txdcToWeiX(tx.Fee))
	if err != nil {
		log.Error().Err(err).Str("tx_id", txID).Msg("xdc settlement: fee transfer failed")
		return
	}
	collection := &domain.FeeCollection{
		ID:            uuid.New().String(),
		TransactionID: txID,
		TenantID:      tx.TenantID,
		FeeAmount:     tx.Fee,
		Asset:         tx.Asset,
		FeeBps:        tx.FeeBps,
		CollectedAt:   time.Now().UTC(),
	}
	if err := e.feeSvc.RecordCollection(ctx, collection); err != nil {
		log.Error().Err(err).Str("tx_id", txID).Msg("xdc settlement: failed to record fee collection")
		return
	}
	log.Info().Str("tx_id", txID).Str("fee_tx", feeHash).Msg("xdc settlement: fee collected")
}

// syncBalance refreshes the cached TXDC balance for a wallet from chain.
func (e *XDCEngine) syncBalance(ctx context.Context, w *domain.Wallet) {
	if w == nil {
		return
	}
	wei, err := e.chain.Balance(ctx, w.PublicKey, chain.NativeTXDC)
	if err != nil {
		log.Debug().Err(err).Str("wallet_id", w.ID).Msg("xdc settlement: balance sync failed")
		return
	}
	if err := e.walletRepo.UpsertBalance(ctx, w.ID, chain.NativeTXDC.Symbol, "", weiToTXDCX(wei)); err != nil {
		log.Debug().Err(err).Str("wallet_id", w.ID).Msg("xdc settlement: balance upsert failed")
	}
}

func (e *XDCEngine) decryptSecret(encryptedHex string) (string, error) {
	encryptedSecret, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return "", fmt.Errorf("decode encrypted secret: %w", err)
	}
	secretBytes, err := crypto.Decrypt(encryptedSecret, e.masterKey)
	if err != nil {
		return "", fmt.Errorf("decrypt wallet secret: %w", err)
	}
	return string(secretBytes), nil
}

// isNativeAsset reports whether an asset code is the chain-native asset.
func isNativeAsset(code string) bool {
	return code == "" || code == "TXDC" || code == "XDC" || code == "XLM"
}

func txdcToWeiX(d decimal.Decimal) *big.Int {
	return d.Mul(decimal.NewFromInt(1e18)).BigInt()
}

func weiToTXDCX(wei *big.Int) decimal.Decimal {
	return decimal.NewFromBigInt(wei, 0).Div(decimal.NewFromInt(1e18))
}
