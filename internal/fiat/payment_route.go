package fiat

import (
	"context"
	"fmt"
	"time"

	"github.com/fluxa/fluxa/internal/routing"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

type PaymentNetworkRoute struct {
	corridor string
	rate     decimal.Decimal
	fee      decimal.Decimal
	feeAsset string
}

func NewPaymentNetworkRoute(corridor string) *PaymentNetworkRoute {
	rates := map[string]decimal.Decimal{
		"INR-EUR": decimal.NewFromFloat(90.20),
		"EUR-INR": decimal.NewFromFloat(0.01109),
	}
	fees := map[string]decimal.Decimal{
		"INR-EUR": decimal.NewFromInt(1200),
		"EUR-INR": decimal.NewFromFloat(8),
	}
	rate := rates[corridor]
	if rate.IsZero() { rate = decimal.NewFromInt(90) }
	fee := fees[corridor]
	if fee.IsZero() { fee = decimal.NewFromInt(1200) }
	return &PaymentNetworkRoute{corridor: corridor, rate: rate, fee: fee, feeAsset: "INR"}
}

func (r *PaymentNetworkRoute) ID() routing.RouteID { return "payment_network" }
func (r *PaymentNetworkRoute) Name() string { return "Payment Network (Ripple ODL)" }
func (r *PaymentNetworkRoute) Supports(from, to, _, _ string) bool { return from+"-"+to == r.corridor }

func (r *PaymentNetworkRoute) Quote(_ context.Context, from, to string, amount decimal.Decimal) (*routing.RouteQuote, error) {
	if !r.Supports(from, to, "", "") {
		return nil, fmt.Errorf("payment network: unsupported pair %s-%s", from, to)
	}
	isInrToEur := from == "INR"
	var rate, destAmt decimal.Decimal
	if isInrToEur {
		rate = decimal.NewFromInt(1).Div(r.rate)
	} else {
		rate = r.rate
	}
	markup := decimal.NewFromFloat(0.0015)
	effectiveRate := rate.Mul(decimal.NewFromInt(1).Sub(markup))
	destAmt = amount.Mul(effectiveRate)

	log.Debug().Str("route", "payment_network").Str("from", from).Str("dest_amount", destAmt.String()).Msg("payment network quote generated")

	return &routing.RouteQuote{
		RouteID: r.ID(), RouteName: r.Name(), SourceAsset: from, DestAsset: to,
		SourceAmount: amount, DestAmount: destAmt.Round(2), Rate: effectiveRate.Round(4),
		SpreadBps: 15, Fee: r.fee, FeeAsset: r.feeAsset,
		SettlementTime: 5 * time.Hour, ExpiresAt: time.Now().Add(30 * time.Second), Provider: "ripple_odl",
	}, nil
}

func (r *PaymentNetworkRoute) Execute(_ context.Context, _ routing.PaymentRequest, _ *routing.RouteQuote) (string, error) {
	return fmt.Sprintf("RIPPLE-ODL-%d", time.Now().UnixNano()%1000000), nil
}
func (r *PaymentNetworkRoute) Status(_ context.Context, _ string) (string, error) { return "processing", nil }
