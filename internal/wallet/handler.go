package wallet

import (
	"encoding/json"
	"net/http"

	"github.com/shopspring/decimal"

	"github.com/fluxa/fluxa/internal/api"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	svc  Service
	idem func(http.Handler) http.Handler
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// WithIdempotency attaches the idempotency-key middleware to the
// state-mutating routes (POST / and POST /{id}/trustlines) only.
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
		post("/", h.createWallet)
		r.Get("/", h.listWallets)
		r.Get("/{id}", h.getWallet)
		r.Get("/{id}/balances", h.getBalances)
		r.Delete("/{id}", h.deleteWallet)
		r.Post("/{id}/faucet", h.faucet)
		post("/{id}/trustlines", h.addTrustline)
		r.Post("/{id}/verify-deposit", h.verifyDeposit)
	}
}

type addTrustlineRequest struct {
	Asset  string `json:"asset" validate:"required"`
	Issuer string `json:"issuer,omitempty"`
	Limit  string `json:"limit,omitempty"`
}

func (h *Handler) getWallet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	wallet, err := h.svc.GetWalletForHandler(r.Context(), id)
	if err != nil {
		api.HandleDomainError(w, err)
		return
	}
	api.JSON(w, http.StatusOK, map[string]interface{}{
		"id":           wallet.ID,
		"public_key":   wallet.PublicKey,
		"custody_type": wallet.CustodyType,
		"created_at":   wallet.CreatedAt,
	})
}

func (h *Handler) createWallet(w http.ResponseWriter, r *http.Request) {
	wallet, err := h.svc.CreateWallet(r.Context())
	if err != nil {
		api.HandleDomainError(w, err)
		return
	}

	api.JSON(w, http.StatusCreated, map[string]interface{}{
		"id":           wallet.ID,
		"public_key":   wallet.PublicKey,
		"custody_type": wallet.CustodyType,
		"created_at":   wallet.CreatedAt,
	})
}

func (h *Handler) getBalances(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	includeFX := r.URL.Query().Get("include_fx")

	balances, err := h.svc.GetBalances(r.Context(), id, includeFX)
	if err != nil {
		api.HandleDomainError(w, err)
		return
	}

	api.JSON(w, http.StatusOK, map[string]interface{}{
		"wallet_id": id,
		"balances":  balances,
	})
}

func (h *Handler) addTrustline(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req addTrustlineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}
	if err := api.Validate(req); err != nil {
		api.BadRequest(w, err.Error())
		return
	}

	txHash, err := h.svc.AddTrustline(r.Context(), id, req.Asset, req.Issuer, req.Limit)
	if err != nil {
		api.HandleDomainError(w, err)
		return
	}

	api.JSON(w, http.StatusOK, map[string]interface{}{
		"status":    "confirmed",
		"wallet_id": id,
		"asset":     req.Asset,
		"tx_hash":   txHash,
	})
}
func (h *Handler) listWallets(w http.ResponseWriter, r *http.Request) {
	wallets, err := h.svc.List(r.Context())
	if err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	type walletResp struct {
		ID        string `json:"id"`
		PublicKey string `json:"public_key"`
	}
	resp := make([]walletResp, 0, len(wallets))
	for _, wl := range wallets {
		resp = append(resp, walletResp{ID: wl.ID, PublicKey: wl.PublicKey})
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"wallets": resp})
}

func (h *Handler) deleteWallet(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) faucet(w http.ResponseWriter, r *http.Request) {
	walletID := chi.URLParam(r, "id")

	var req struct {
		AssetCode string  `json:"asset_code"`
		Amount    float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.AssetCode == "" {
		req.AssetCode = "USDC"
	}
	if req.Amount <= 0 {
		req.Amount = 1000
	}

	amt := decimal.NewFromFloat(req.Amount)
	result, err := h.svc.Faucet(r.Context(), walletID, req.AssetCode, amt)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resp := map[string]interface{}{
		"wallet_id":   walletID,
		"asset_code":  req.AssetCode,
		"added":       req.Amount,
		"new_balance": result.Balance,
	}
	if result.TxHash != "" {
		resp["tx_hash"] = result.TxHash
	}
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) verifyDeposit(w http.ResponseWriter, r *http.Request) {
	walletID := chi.URLParam(r, "id")

	var req struct {
		TxHash string `json:"tx_hash"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.TxHash == "" {
		http.Error(w, `{"error":"tx_hash is required"}`, http.StatusBadRequest)
		return
	}

	tx, err := h.svc.VerifyDeposit(r.Context(), walletID, req.TxHash)
	if err != nil {
		api.HandleDomainError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    tx.Status,
		"wallet_id": walletID,
		"tx_hash":   tx.TxHash,
		"amount":    tx.Amount,
		"asset":     tx.Asset,
	})
}
