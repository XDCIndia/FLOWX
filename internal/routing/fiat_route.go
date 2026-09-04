package routing

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// FiatRoute handles fiat ↔ crypto conversions via payment processors.
type FiatRoute struct {
	id              RouteID
	name            string
	supportedPairs  map[string]bool
	provider        string
	rate            decimal.Decimal // fiat per 1 USDC/TXDC
	spreadBps       int
	settlementTime  time.Duration
	minAmount       decimal.Decimal
	maxAmount       decimal.Decimal
}

// NewFiatNGNRoute creates a Nigerian Naira route via Flutterwave.
func NewFiatNGNRoute() *FiatRoute {
	return &FiatRoute{
		id:             RouteFiatNGN,
		name:           "NGN (Flutterwave)",
		supportedPairs: map[string]bool{"NGN-TXDC": true, "TXDC-NGN": true, "NGN-USDC": true, "USDC-NGN": true},
		provider:       "flutterwave",
		rate:           decimal.NewFromInt(1500), // 1500 NGN per USDC
		spreadBps:      100,                      // 1% spread
		settlementTime: 5 * time.Minute,
		minAmount:      decimal.NewFromInt(500),    // 500 NGN min
		maxAmount:      decimal.NewFromInt(5000000), // 5M NGN max
	}
}

func (r *FiatRoute) ID() RouteID   { return r.id }
func (r *FiatRoute) Name() string  { return r.name }

func (r *FiatRoute) Supports(from, to, _, _ string) bool {
	pair := from + "-" + to
	return r.supportedPairs[pair]
}

func (r *FiatRoute) Quote(_ context.Context, from, to string, amount decimal.Decimal) (*RouteQuote, error) {
	if !r.Supports(from, to, "", "") {
		return nil, fmt.Errorf("fiat route: unsupported pair %s-%s", from, to)
	}

	// Determine direction
	isDeposit := from == "NGN" || from == "KES" // fiat → crypto

	var rate decimal.Decimal
	if isDeposit {
		// Deposit: 1 USDC = rate fiat
		rate = r.rate
	} else {
		// Withdrawal: 1 USDC = rate fiat (inverse)
		rate = r.rate
	}

	// Apply spread
	spreadFactor := decimal.NewFromInt(int64(r.spreadBps)).Div(decimal.NewFromInt(10000))
	effectiveRate := rate.Mul(decimal.NewFromInt(1).Add(spreadFactor))

	var sourceAmt, destAmt decimal.Decimal
	if isDeposit {
		sourceAmt = amount
		destAmt = amount.Div(effectiveRate)
	} else {
		sourceAmt = amount
		destAmt = amount.Mul(effectiveRate)
	}

	return &RouteQuote{
		RouteID:        r.ID(),
		RouteName:      r.Name(),
		SourceAsset:    from,
		DestAsset:      to,
		SourceAmount:   sourceAmt,
		DestAmount:     destAmt,
		Rate:           effectiveRate,
		SpreadBps:      r.spreadBps,
		Fee:            sourceAmt.Mul(decimal.NewFromFloat(0.015)), // 1.5% fee
		FeeAsset:       from,
		SettlementTime: r.settlementTime,
		ExpiresAt:      time.Now().Add(30 * time.Second),
		Provider:       r.provider,
	}, nil
}

func (r *FiatRoute) Execute(_ context.Context, _ PaymentRequest, _ *RouteQuote) (string, error) {
	return "", fmt.Errorf("fiat execute: use POST /v1/fiat/deposit or /withdraw directly")
}

func (r *FiatRoute) Status(_ context.Context, _ string) (string, error) {
	return "", fmt.Errorf("fiat status: check via fiat webhook")
}
