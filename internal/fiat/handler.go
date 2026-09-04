package fiat

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/fluxa/fluxa/internal/api"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) DepositRoutes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/fiat", h.handleDeposit)
		r.Post("/fiat/simulate", h.handleSimulateDeposit)
	}
}

func (h *Handler) WithdrawRoutes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/fiat", h.handleWithdrawal)
	}
}

func (h *Handler) WebhookRoutes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/{provider}", h.handleWebhook)
	}
}

// PublicWebhook exposes the provider webhook handler on the PUBLIC route
// (/webhooks/fiat/{provider}). Real provider callbacks carry no FlowX auth;
// the rail verifies per-provider signatures (verif-hash / Stripe-Signature).
func (h *Handler) PublicWebhook(w http.ResponseWriter, r *http.Request) {
	h.handleWebhook(w, r)
}

type depositReq struct {
	Amount   string `json:"amount" validate:"required"`
	Currency string `json:"currency" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required"`
}

func (h *Handler) handleDeposit(w http.ResponseWriter, r *http.Request) {
	walletID := chi.URLParam(r, "id")
	if walletID == "" {
		api.BadRequest(w, "wallet id is required")
		return
	}

	var req depositReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	if err := api.Validate(req); err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		api.BadRequest(w, "invalid amount")
		return
	}

	dr := DepositRequest{
		WalletID:      walletID,
		Reference:     "DEP-" + uuid.New().String()[:8],
		FiatAmount:    amount,
		FiatCurrency:  req.Currency,
		CustomerEmail: req.Email,
		CustomerName:  req.Name,
	}

	resp, err := h.svc.InitiateDeposit(r.Context(), dr)
	if err != nil {
		log.Error().Err(err).Str("wallet_id", walletID).Msg("initiate deposit failed")
		api.InternalError(w, err)
		return
	}

	api.JSON(w, http.StatusOK, resp)
}

type withdrawReq struct {
	Amount        string `json:"amount" validate:"required"`
	Currency      string `json:"currency" validate:"required"`
	AccountBank   string `json:"account_bank" validate:"required"`
	AccountNumber string `json:"account_number" validate:"required"`
}

func (h *Handler) handleWithdrawal(w http.ResponseWriter, r *http.Request) {
	walletID := chi.URLParam(r, "id")
	if walletID == "" {
		api.BadRequest(w, "wallet id is required")
		return
	}

	var req withdrawReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	if err := api.Validate(req); err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		api.BadRequest(w, "invalid amount")
		return
	}

	wr := WithdrawRequest{
		WalletID:      walletID,
		Reference:     "WIT-" + uuid.New().String()[:8],
		FiatAmount:    amount,
		FiatCurrency:  req.Currency,
		AccountBank:   req.AccountBank,
		AccountNumber: req.AccountNumber,
	}

	resp, err := h.svc.InitiateWithdrawal(r.Context(), wr)
	if err != nil {
		log.Error().Err(err).Str("wallet_id", walletID).Msg("initiate withdrawal failed")
		api.InternalError(w, err)
		return
	}

	api.JSON(w, http.StatusOK, resp)
}

func (h *Handler) handleWebhook(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider == "" {
		api.BadRequest(w, "provider is required")
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		api.BadRequest(w, "read payload error")
		return
	}

	// Flutterwave sends its signature in "verif-hash"; Stripe in
	// "Stripe-Signature". The service drives one rail, so trying both is
	// safe: the inactive rail's verify will simply fail.
	signature := r.Header.Get("verif-hash")
	if signature == "" {
		signature = r.Header.Get("Stripe-Signature")
	}

	if err := h.svc.HandleWebhook(r.Context(), payload, signature); err != nil {
		log.Error().Err(err).Str("provider", provider).Msg("webhook handling failed")
		// Do not return 500 so provider won't keep retrying if it's a fatal validation error
		api.BadRequest(w, "webhook validation failed")
		return
	}

	api.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}


func (h *Handler) handleSimulateDeposit(w http.ResponseWriter, r *http.Request) {
	walletID := chi.URLParam(r, "id")
	if walletID == "" {
		api.BadRequest(w, "wallet id is required")
		return
	}

	var req struct {
		Amount   string `json:"amount" validate:"required"`
		Currency string `json:"currency" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		api.BadRequest(w, "invalid amount")
		return
	}

	// For demo: directly credit the wallet with equivalent crypto
	// Using a fixed rate: 1 USDC = 1500 NGN (or 1 TXDC = 1500 NGN)
	creditAsset := "USDC"
	creditAmount := amount.Div(decimal.NewFromInt(1500))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"wallet_id":   walletID,
		"fiat_amount": req.Amount,
		"fiat_currency": req.Currency,
		"credit_amount": creditAmount.String(),
		"credit_asset": creditAsset,
		"rate": "1500",
		"status": "simulated",
		"message": "Demo mode: deposit simulated. In production, webhook confirms payment.",
	})
}

