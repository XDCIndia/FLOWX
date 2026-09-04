package routing

import (
	"encoding/json"
	"net/http"

	"github.com/fluxa/fluxa/internal/api"
	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"
)

// Handler exposes the routing engine as HTTP endpoints.
type Handler struct {
	evaluator *Evaluator
	scorer    *Scorer
	checker   ComplianceChecker
}

// NewHandler creates a routing handler with default components.
func NewHandler() *Handler {
	evaluator := NewEvaluator(
		NewOnChainXDCRoute(),
		NewFiatNGNRoute(),
	)
	return &Handler{
		evaluator: evaluator,
		scorer:    NewScorer(DefaultWeights(), RankingBalanced),
		checker:   NewDefaultComplianceChecker(),
	}
}

// Routes registers the payment routing endpoints.
func (h *Handler) Routes() func(r chi.Router) {
	return func(r chi.Router) {
		r.Post("/quote", h.quote)
		r.Post("/send", h.send)
		r.Get("/routes", h.listRoutes)
	}
}

type quoteRequest struct {
	SourceAsset  string `json:"source_asset"`
	DestAsset    string `json:"dest_asset"`
	Amount       string `json:"amount"`
	SourceRegion string `json:"source_region,omitempty"`
	DestRegion   string `json:"dest_region,omitempty"`
	RiskProfile  string `json:"risk_profile,omitempty"`
	RankingMode  string `json:"ranking_mode,omitempty"`
}

type routeOption struct {
	RouteID      string  `json:"route_id"`
	RouteName    string  `json:"route_name"`
	Score        float64 `json:"score"`
	CostScore    float64 `json:"cost_score"`
	SpeedScore   float64 `json:"speed_score"`
	Reliability  float64 `json:"reliability"`
	Compliance   float64 `json:"compliance"`
	Liquidity    float64 `json:"liquidity"`
	Recommended  bool    `json:"recommended"`
	SourceAsset  string  `json:"source_asset"`
	DestAsset    string  `json:"dest_asset"`
	SourceAmount string  `json:"source_amount"`
	DestAmount   string  `json:"dest_amount"`
	Rate         string  `json:"rate"`
	Fee          string  `json:"fee"`
	FeeAsset     string  `json:"fee_asset"`
	Settlement   string  `json:"settlement_time"`
	Provider     string  `json:"provider"`
	Warnings     []string `json:"warnings,omitempty"`
}

// quote returns ranked route options for a payment corridor.
func (h *Handler) quote(w http.ResponseWriter, r *http.Request) {
	var req quoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		api.BadRequest(w, "amount must be a positive number")
		return
	}

	paymentReq := PaymentRequest{
		SourceAsset:  req.SourceAsset,
		DestAsset:    req.DestAsset,
		Amount:       amount,
		SourceRegion: req.SourceRegion,
		DestRegion:   req.DestRegion,
		RiskProfile:  req.RiskProfile,
	}

	// Step 1: Collect quotes from all viable routes
	quotes, err := h.evaluator.Evaluate(r.Context(), paymentReq)
	if err != nil {
		api.HandleDomainError(w, err)
		return
	}

	// Step 2: Score and rank
	mode := RankingMode(req.RankingMode)
	if mode == "" {
		mode = RankingBalanced
	}
	h.scorer = NewScorer(DefaultWeights(), mode)
	scores := h.scorer.ScoreAndRank(quotes)

	// Step 3: Compliance check per route
	options := make([]routeOption, 0, len(scores))
	for _, s := range scores {
		compliance, _ := h.checker.CheckRoute(r.Context(), paymentReq, s.Quote.RouteID)

		if compliance != nil && compliance.Blocked {
		s.Warnings = append(s.Warnings, compliance.Warnings...)
			continue
		}

		opt := routeOption{
			RouteID:      string(s.Quote.RouteID),
			RouteName:    s.Quote.RouteName,
			Score:        round2(s.Score),
			CostScore:    round2(s.CostScore),
			SpeedScore:   round2(s.SpeedScore),
			Reliability:  round2(s.Reliability),
			Compliance:   round2(s.Compliance),
			Liquidity:    round2(s.Liquidity),
			Recommended:  s.Recommended,
			SourceAsset:  s.Quote.SourceAsset,
			DestAsset:    s.Quote.DestAsset,
			SourceAmount: s.Quote.SourceAmount.StringFixed(7),
			DestAmount:   s.Quote.DestAmount.StringFixed(7),
			Rate:         s.Quote.Rate.StringFixed(7),
			Fee:          s.Quote.Fee.StringFixed(7),
			FeeAsset:     s.Quote.FeeAsset,
			Settlement:   s.Quote.SettlementTime.String(),
			Provider:     s.Quote.Provider,
			Warnings:     s.Warnings,
		}

		if compliance != nil {
			opt.Warnings = append(opt.Warnings, compliance.Warnings...)
		}
		options = append(options, opt)
	}

	api.JSON(w, http.StatusOK, map[string]interface{}{
		"source_asset": req.SourceAsset,
		"dest_asset":   req.DestAsset,
		"amount":       amount.StringFixed(7),
		"ranking_mode": string(mode),
		"routes":       options,
		"total_routes": len(options),
	})
}

type sendRequest struct {
	SourceAsset  string `json:"source_asset"`
	DestAsset    string `json:"dest_asset"`
	Amount       string `json:"amount"`
	RouteID      string `json:"route_id,omitempty"`    // explicit route, or auto-select
	AutoSelect   bool   `json:"auto_select,omitempty"` // pick best route
	SourceRegion string `json:"source_region,omitempty"`
	DestRegion   string `json:"dest_region,omitempty"`
}

// send executes a payment with auto-selected or explicit route.
func (h *Handler) send(w http.ResponseWriter, r *http.Request) {
	var req sendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.BadRequest(w, "invalid request body")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil || amount.LessThanOrEqual(decimal.Zero) {
		api.BadRequest(w, "amount must be a positive number")
		return
	}

	paymentReq := PaymentRequest{
		SourceAsset:  req.SourceAsset,
		DestAsset:    req.DestAsset,
		Amount:       amount,
		SourceRegion: req.SourceRegion,
		DestRegion:   req.DestRegion,
	}

	// Get quotes
	quotes, err := h.evaluator.Evaluate(r.Context(), paymentReq)
	if err != nil {
		api.HandleDomainError(w, err)
		return
	}

	// Score and rank
	scores := h.scorer.ScoreAndRank(quotes)

	// Select route
	var selected *RouteQuote
	if req.RouteID != "" {
		// Explicit route requested
		for _, s := range scores {
			if string(s.Quote.RouteID) == req.RouteID {
				selected = &s.Quote
				break
			}
		}
		if selected == nil {
			api.BadRequest(w, "route not available for this corridor")
			return
		}
	} else if req.AutoSelect || len(scores) > 0 {
		// Auto-select best route
		selected = &scores[0].Quote
	}

	if selected == nil {
		api.BadRequest(w, "no routes available")
		return
	}

	// Execute via the route
	for _, route := range h.evaluator.routes {
		if route.ID() == selected.RouteID {
			ref, err := route.Execute(r.Context(), paymentReq, selected)
			if err != nil {
				api.HandleDomainError(w, err)
				return
			}
			api.JSON(w, http.StatusAccepted, map[string]interface{}{
				"status":       "initiated",
				"route_id":     string(selected.RouteID),
				"route_name":   selected.RouteName,
				"reference":    ref,
				"source_asset": selected.SourceAsset,
				"dest_asset":   selected.DestAsset,
				"amount":       selected.SourceAmount.StringFixed(7),
				"dest_amount":  selected.DestAmount.StringFixed(7),
				"fee":          selected.Fee.StringFixed(7),
				"fee_asset":    selected.FeeAsset,
			})
			return
		}
	}

	api.BadRequest(w, "route execution not available")
}

// listRoutes returns all registered routes and their capabilities.
func (h *Handler) listRoutes(w http.ResponseWriter, _ *http.Request) {
	routes := h.evaluator.ListRoutes()
	api.JSON(w, http.StatusOK, map[string]interface{}{
		"routes": routes,
	})
}

func round2(v float64) float64 {
	return float64(int(v*100)) / 100
}

// RegisterRoute adds a payment route to the evaluator.
func (h *Handler) RegisterRoute(route PaymentRoute) {
	h.evaluator.Register(route)
}
