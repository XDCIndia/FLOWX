package routing

import (
	"context"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEvaluator_PicksSwapRouteWhenCheaper verifies the AMM swap leg competes
// on cost: with reserves priced at a tighter rate than the bridge's flat
// fee + spread, the scorer must recommend amm_swap for TXDC->USDC.
func TestEvaluator_PicksSwapRouteWhenCheaper(t *testing.T) {
	ctx := context.Background()

	// Pool: 10 TXDC / 35,000 USDC. Swapping 1 TXDC yields ~3,173 USDC
	// after fee+impact — far better than any fiat rail for this corridor.
	swap := newTestAMMRoute(testReserveTXDC, testReserveToken, 6)

	// Bridge-style route quoting a deliberately worse TXDC->USDC deal:
	// lower effective FX rate AND a higher all-in cost (0.5% fee + 5bps
	// spread = 0.55% vs the AMM's flat 0.3%).
	bridge := &stubRoute{
		id:   "xdc_bridge",
		name: "Blockchain (XDC Stablecoin)",
		pair: "TXDC-USDC",
		quote: &RouteQuote{
			RouteID: "xdc_bridge", RouteName: "Blockchain (XDC Stablecoin)",
			SourceAsset: "TXDC", DestAsset: "USDC",
			SourceAmount: decimal.NewFromInt(1), DestAmount: decimal.NewFromInt(2500),
			Rate: decimal.NewFromInt(2500), SpreadBps: 5,
			Fee: decimal.NewFromFloat(0.005), FeeAsset: "TXDC",
			SettlementTime: 12 * time.Second, Provider: "xdc_stablecoin",
		},
	}

	ev := NewEvaluator(swap, bridge)
	quotes, err := ev.Evaluate(ctx, PaymentRequest{
		SourceAsset: "TXDC", DestAsset: "USDC", Amount: decimal.NewFromInt(1),
	})
	require.NoError(t, err)
	require.Len(t, quotes, 2)

	scores := NewScorer(DefaultWeights(), RankingBalanced).ScoreAndRank(quotes)
	require.NotEmpty(t, scores)
	assert.Equal(t, RouteAMMSwap, scores[0].Quote.RouteID,
		"amm_swap should win on cost+liquidity; got %s", scores[0].Quote.RouteID)
	assert.True(t, scores[0].Recommended)

	// The swap quote's cost score must beat the bridge's.
	var swapScore, bridgeScore *RouteScore
	for i := range scores {
		switch scores[i].Quote.RouteID {
		case RouteAMMSwap:
			swapScore = &scores[i]
		case "xdc_bridge":
			bridgeScore = &scores[i]
		}
	}
	require.NotNil(t, swapScore)
	require.NotNil(t, bridgeScore)
	assert.Greater(t, swapScore.CostScore, bridgeScore.CostScore)
	assert.Greater(t, swapScore.Liquidity, bridgeScore.Liquidity)
}

// TestEvaluator_SwapRouteDisabledWithoutConfig verifies the route degrades
// cleanly: with no AMM_POOL_ADDRESS / XDC_USDC_CONTRACT_ADDRESS its Quote
// errors and the evaluator skips it, surfacing ErrNoRoutes when nothing
// else supports the pair.
func TestEvaluator_SwapRouteDisabledWithoutConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("quote error surfaces missing config", func(t *testing.T) {
		r := NewAMMSwapRoute("", "", "", "key")
		assert.True(t, r.Supports("TXDC", "USDC", "", "")) // pair still supported
		_, err := r.Quote(ctx, "TXDC", "USDC", decimal.NewFromInt(1))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not configured")
	})

	t.Run("evaluator skips unconfigured route", func(t *testing.T) {
		ev := NewEvaluator(NewAMMSwapRoute("", "", "", "key"))
		_, err := ev.Evaluate(ctx, PaymentRequest{
			SourceAsset: "TXDC", DestAsset: "USDC", Amount: decimal.NewFromInt(1),
		})
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNoRoutes)
	})

	t.Run("configured route produces a quote", func(t *testing.T) {
		ev := NewEvaluator(newTestAMMRoute(testReserveTXDC, testReserveToken, 6))
		quotes, err := ev.Evaluate(ctx, PaymentRequest{
			SourceAsset: "TXDC", DestAsset: "USDC", Amount: decimal.NewFromInt(1),
		})
		require.NoError(t, err)
		require.Len(t, quotes, 1)
		assert.Equal(t, RouteAMMSwap, quotes[0].RouteID)
	})
}

// stubRoute is a minimal PaymentRoute for evaluator tests.
type stubRoute struct {
	id    RouteID
	name  string
	pair  string
	quote *RouteQuote
}

func (s *stubRoute) ID() RouteID  { return s.id }
func (s *stubRoute) Name() string { return s.name }
func (s *stubRoute) Supports(from, to, _, _ string) bool {
	return from+"-"+to == s.pair
}
func (s *stubRoute) Quote(_ context.Context, from, to string, _ decimal.Decimal) (*RouteQuote, error) {
	if from+"-"+to != s.pair {
		return nil, ErrNoRoutes
	}
	return s.quote, nil
}
func (s *stubRoute) Execute(_ context.Context, _ PaymentRequest, _ *RouteQuote) (string, error) {
	return "", nil
}
func (s *stubRoute) Status(_ context.Context, _ string) (string, error) { return "", nil }
