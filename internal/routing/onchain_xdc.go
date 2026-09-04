package routing

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

// OnChainXDCRoute handles direct XDC transfers between wallets.
type OnChainXDCRoute struct {
	settlementTime time.Duration
	spreadBps      int
}

// NewOnChainXDCRoute creates an on-chain XDC route.
func NewOnChainXDCRoute() *OnChainXDCRoute {
	return &OnChainXDCRoute{
		settlementTime: 12 * time.Second, // ~6 confirmations at 2s blocks
		spreadBps:      0,                // no spread for native transfers
	}
}

func (r *OnChainXDCRoute) ID() RouteID   { return RouteOnChainXDC }
func (r *OnChainXDCRoute) Name() string  { return "On-Chain XDC" }

func (r *OnChainXDCRoute) Supports(from, to, _, _ string) bool {
	pair := from + "-" + to
	return pair == "TXDC-TXDC" || pair == "XDC-TXDC" || pair == "TXDC-XDC" ||
		pair == "XLM-TXDC" || pair == "USDC-TXDC" // cross-chain via bridge (future)
}

func (r *OnChainXDCRoute) Quote(_ context.Context, from, to string, amount decimal.Decimal) (*RouteQuote, error) {
	if !r.Supports(from, to, "", "") {
		return nil, fmt.Errorf("on-chain XDC: unsupported pair %s-%s", from, to)
	}

	return &RouteQuote{
		RouteID:        r.ID(),
		RouteName:      r.Name(),
		SourceAsset:    from,
		DestAsset:      to,
		SourceAmount:   amount,
		DestAmount:     amount, // 1:1 for native asset
		Rate:           decimal.NewFromInt(1),
		SpreadBps:      r.spreadBps,
		Fee:            decimal.NewFromFloat(0.001), // gas cost ~0.001 TXDC
		FeeAsset:       "TXDC",
		SettlementTime: r.settlementTime,
		ExpiresAt:      time.Now().Add(30 * time.Second),
		Provider:       "xdc_rpc",
	}, nil
}

func (r *OnChainXDCRoute) Execute(_ context.Context, _ PaymentRequest, _ *RouteQuote) (string, error) {
	return "", fmt.Errorf("on-chain execute: use POST /v1/transfers/ directly")
}

func (r *OnChainXDCRoute) Status(_ context.Context, _ string) (string, error) {
	return "", fmt.Errorf("on-chain status: use GET /v1/transfers/{id} directly")
}
