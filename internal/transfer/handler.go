package transfer

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/fluxa/fluxa/internal/api"
	"github.com/fluxa/fluxa/internal/domain"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

type Handler struct {
	svc  Service
	idem func(http.Handler) http.Handler
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// WithIdempotency attaches the idempotency-key middleware to the
// state-mutating route (POST /) only; reads (GET /{id}) are unaffected.
func (h *Handler) WithIdempotency(mw func(http.Handler) http.Handler) *Handler {
	h.idem = mw
	return h
}

func (h *Handler) Routes() func(r chi.Router) {
	return func(r chi.Router) {
		post := r.Post
		if h.idem != nil {
			post = r.With(h.idem).Post
		}
		post("/", h.initiateTransfer)
		r.Get("/{id}", h.getTransaction)
	}
}

func (h *Handler) TransactionRoutes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", h.listTransactions)
	}
}

type createTransferRequest struct {
	FromWalletID string `json:"from_wallet_id" validate:"required,uuid"`
	// Exactly one of ToWalletID / ToAddress must be set: a FlowX wallet
	// UUID, or a raw on-chain address (0x…/xdc…) for external payouts.
	ToWalletID string `json:"to_wallet_id" validate:"omitempty,uuid"`
	ToAddress  string `json:"to_address"   validate:"omitempty"`
	Asset      string `json:"asset"        validate:"required"`
	Amount     string `json:"amount"       validate:"required"`
}

type transferResponse struct {
	ID         string `json:"id"`
	TxHash     string `json:"tx_hash,omitempty"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	FromWallet string `json:"from_wallet_id"`
	ToWallet   string `json:"to_wallet_id,omitempty"`
	ToAddress  string `json:"to_address,omitempty"`
	Asset      string `json:"asset"`
	Amount     string `json:"amount"`
	FeeAmount  string `json:"fee_amount"`
	NetAmount  string `json:"net_amount"`
	FeeBps     int    `json:"fee_bps"`
	CreatedAt  string `json:"created_at"`
}

func toTransferResponse(tx *domain.Transaction) transferResponse {
	return transferResponse{
		ID:         tx.ID,
		TxHash:     tx.TxHash,
		Type:       string(tx.Type),
		Status:     string(tx.Status),
		FromWallet: tx.FromWallet,
		ToWallet:   tx.ToWallet,
		ToAddress:  tx.ToAddress,
		Asset:      tx.Asset,
		Amount:     tx.Amount.StringFixed(7),
		FeeAmount:  tx.Fee.StringFixed(7),
		NetAmount:  tx.NetAmount().StringFixed(7),
		FeeBps:     tx.FeeBps,
		CreatedAt:  tx.CreatedAt.Format(time.RFC3339),
	}
}

func (h *Handler) initiateTransfer(w http.ResponseWriter, r *http.Request) {
	var req createTransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}
	if err := api.Validate(req); err != nil {
		api.BadRequest(w, err.Error())
		return
	}
	if (req.ToWalletID == "") == (req.ToAddress == "") {
		api.BadRequest(w, "exactly one of to_wallet_id or to_address is required")
		return
	}
	if req.ToAddress != "" && !IsValidDestinationAddress(req.ToAddress) {
		api.BadRequest(w, "to_address must be a 0x or xdc address of 40 hex characters")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		api.BadRequest(w, "amount must be a positive number")
		return
	}

	var tx *domain.Transaction
	if req.ToAddress != "" {
		tx, err = h.svc.InitiatePayoutIdempotent(r.Context(), req.FromWalletID, req.ToAddress, req.Asset, amount, r.Header.Get("Idempotency-Key"))
	} else {
		tx, err = h.svc.InitiateTransferIdempotent(r.Context(), req.FromWalletID, req.ToWalletID, req.Asset, amount, r.Header.Get("Idempotency-Key"))
	}
	if err != nil {
		api.HandleDomainError(w, err)
		return
	}

	api.JSON(w, http.StatusAccepted, toTransferResponse(tx))
}

func (h *Handler) getTransaction(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	tx, err := h.svc.GetTransaction(r.Context(), id)
	if err != nil {
		api.HandleDomainError(w, err)
		return
	}
	api.JSON(w, http.StatusOK, toTransferResponse(tx))
}

func (h *Handler) listTransactions(w http.ResponseWriter, r *http.Request) {
	walletID := r.URL.Query().Get("wallet_id")
	if walletID == "" {
		api.BadRequest(w, "wallet_id query param is required")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	txs, err := h.svc.ListTransactions(r.Context(), walletID, limit, offset)
	if err != nil {
		api.HandleDomainError(w, err)
		return
	}

	responses := make([]transferResponse, len(txs))
	for i, tx := range txs {
		responses[i] = toTransferResponse(tx)
	}

	api.JSON(w, http.StatusOK, map[string]interface{}{
		"transactions": responses,
	})
}
