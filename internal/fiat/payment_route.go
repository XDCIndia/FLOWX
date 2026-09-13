package fiat

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"github.com/fluxa/fluxa/internal/fx"
	"github.com/fluxa/fluxa/internal/routing"
)

// PaymentNetworkRoute quotes fiat→fiat corridors ("Ripple ODL" style rail).
// Rates come live from the FX service (units of `to` per 1 `from`); the
// static map is only a resilience fallback when the FX feed is unreachable.
//
// Execution is NOT wired to a real Ripple/provider account: there is no
// liquidity-provider partnership, so Execute fails loudly rather than
// fabricating a reference. For a real executable TXDC<->USDC leg use the
// amm_swap route (our own on-chain AMM) or stripe_bank.
type PaymentNetworkRoute struct {
	corridor string
	rate     decimal.Decimal // static fallback only
	fee      decimal.Decimal
	feeAsset string
	fxSvc    fx.Service
}

func NewPaymentNetworkRoute(corridor string, fxSvc fx.Service) *PaymentNetworkRoute {
	fees := map[string]decimal.Decimal{
		"INR-EUR": decimal.NewFromInt(1200),
		"EUR-INR": decimal.NewFromFloat(8),
		"INR-USDC": decimal.NewFromFloat(6.5), // ₹6.5 flat
		"USDC-INR": decimal.NewFromFloat(0.05), // $0.05 flat
		"INR-TXDC": decimal.NewFromFloat(5),    // ₹5 flat
		"TXDC-INR": decimal.NewFromFloat(0.03),  // $0.03 flat
	}
	fee := fees[corridor]
	if fee.IsZero() {
		fee = decimal.NewFromInt(1200)
	}
	return &PaymentNetworkRoute{corridor: corridor, fee: fee, feeAsset: "INR", fxSvc: fxSvc}
}

func (r *PaymentNetworkRoute) ID() routing.RouteID { return "payment_network" }
func (r *PaymentNetworkRoute) Name() string        { return "Payment Network (Ripple ODL)" }
func (r *PaymentNetworkRoute) Supports(from, to, _, _ string) bool {
	return from+"-"+to == r.corridor
}

func (r *PaymentNetworkRoute) Quote(ctx context.Context, from, to string, amount decimal.Decimal) (*routing.RouteQuote, error) {
	if !r.Supports(from, to, "", "") {
		return nil, fmt.Errorf("payment network: unsupported pair %s-%s", from, to)
	}

	// Live mid-market rate: units of `to` per 1 `from`.
	rate := r.rate
	if r.fxSvc != nil {
		if resp, err := r.fxSvc.GetRates(ctx, from, to); err == nil && resp.MidMarketRate.GreaterThan(decimal.Zero) {
			rate = resp.MidMarketRate
		} else {
			log.Warn().Err(err).Str("pair", r.corridor).
				Msg("payment network: live FX unavailable, quote skipped")
			return nil, fmt.Errorf("payment network: FX rate unavailable for %s", r.corridor)
		}
	}
	if rate.IsZero() {
		return nil, fmt.Errorf("payment network: no rate for %s", r.corridor)
	}

	markup := decimal.NewFromFloat(0.0015)
	effectiveRate := rate.Mul(decimal.NewFromInt(1).Sub(markup))
	destAmt := amount.Mul(effectiveRate)

	log.Debug().Str("route", "payment_network").Str("from", from).
		Str("rate", rate.String()).Str("dest_amount", destAmt.String()).Msg("payment network quote generated")

	return &routing.RouteQuote{
		RouteID: r.ID(), RouteName: r.Name(), SourceAsset: from, DestAsset: to,
		SourceAmount: amount, DestAmount: destAmt.Round(2), Rate: effectiveRate.Round(6),
		SpreadBps: 15, Fee: r.fee, FeeAsset: r.feeAsset,
		SettlementTime: 5 * time.Hour, ExpiresAt: time.Now().Add(30 * time.Second), Provider: "ripple_odl",
	}, nil
}

// Execute refuses to run: the Ripple ODL leg has no liquidity-provider
// partnership behind it, so there is nothing real to execute against.
// Previously this fabricated a "RIPPLE-ODL-%d" reference with zero real
// payment activity — that behavior is removed. Use the amm_swap route
// (real on-chain AMM swap) or stripe_bank for executable corridors.
func (r *PaymentNetworkRoute) Execute(_ context.Context, _ routing.PaymentRequest, _ *routing.RouteQuote) (string, error) {
	return "", fmt.Errorf("ripple ODL requires a Ripple liquidity-provider partnership; use amm_swap or stripe_bank routes")
}

func (r *PaymentNetworkRoute) Status(_ context.Context, _ string) (string, error) {
	return "processing", nil
}
