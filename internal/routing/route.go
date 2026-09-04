package routing

import (
	"context"
	"time"

	"github.com/shopspring/decimal"
)

// RouteID uniquely identifies a payment route.
type RouteID string

const (
	RouteOnChainXDC  RouteID = "onchain_xdc"
	RouteFiatNGN     RouteID = "fiat_ngn"
	RouteFiatKES     RouteID = "fiat_kes"
)

// RouteCapability describes what a route can do.
type RouteCapability struct {
	ID              RouteID
	Name            string
	Description     string
	SupportedPairs  []string // e.g. ["NGN-TXDC", "TXDC-NGN"]
	SupportedRegions []string // e.g. ["NG", "KE"]
	MinAmount       decimal.Decimal
	MaxAmount       decimal.Decimal
	RequiresKYC     bool
}

// RouteQuote is a price quote from a specific route.
type RouteQuote struct {
	RouteID        RouteID
	RouteName      string
	SourceAsset    string
	DestAsset      string
	SourceAmount   decimal.Decimal
	DestAmount     decimal.Decimal
	Rate           decimal.Decimal
	SpreadBps      int
	Fee            decimal.Decimal
	FeeAsset       string
	SettlementTime time.Duration // estimated time to finality
	ExpiresAt      time.Time
	Provider       string        // e.g. "xdc_rpc", "flutterwave"
}

// RouteEvaluator evaluates routes for a given payment request.
type RouteEvaluator interface {
	// Evaluate returns quotes from all viable routes for a corridor.
	Evaluate(ctx context.Context, req PaymentRequest) ([]RouteQuote, error)
	// SupportsPair checks if any route handles this pair.
	SupportsPair(from, to string) bool
}

// PaymentRequest is the input for route evaluation.
type PaymentRequest struct {
	SourceAsset   string          `json:"source_asset"`
	DestAsset     string          `json:"dest_asset"`
	Amount        decimal.Decimal `json:"amount"`
	SourceRegion  string          `json:"source_region,omitempty"`
	DestRegion    string          `json:"dest_region,omitempty"`
	RiskProfile   string          `json:"risk_profile,omitempty"` // "low", "medium", "high"
}

// PaymentRoute is the interface all routes must implement.
type PaymentRoute interface {
	// ID returns the route identifier.
	ID() RouteID
	// Name returns a human-readable name.
	Name() string
	// Supports checks if this route handles the given pair and region.
	Supports(from, to, sourceRegion, destRegion string) bool
	// Quote returns a price quote for the requested amount.
	Quote(ctx context.Context, from, to string, amount decimal.Decimal) (*RouteQuote, error)
	// Execute initiates the payment and returns a tracking reference.
	Execute(ctx context.Context, req PaymentRequest, quote *RouteQuote) (string, error)
	// Status returns the current status of a payment.
	Status(ctx context.Context, reference string) (string, error)
}
