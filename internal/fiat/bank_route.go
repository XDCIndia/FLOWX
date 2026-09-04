package fiat

import (
	"context"
	"fmt"
	"time"

	"github.com/fluxa/fluxa/internal/routing"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
)

type BankRoute struct {
	corridor string
	rate     decimal.Decimal
	fee      decimal.Decimal
	feeAsset string
}

func NewBankRoute(corridor string) *BankRoute {
	rates := map[string]decimal.Decimal{
		"INR-EUR": decimal.NewFromFloat(90.50),
		"EUR-INR": decimal.NewFromFloat(0.01105),
	}
	fees := map[string]decimal.Decimal{
		"INR-EUR": decimal.NewFromInt(4500),
		"EUR-INR": decimal.NewFromFloat(25),
	}
	rate := rates[corridor]
	if rate.IsZero() { rate = decimal.NewFromInt(90) }
	fee := fees[corridor]
	if fee.IsZero() { fee = decimal.NewFromInt(4500) }
	return &BankRoute{corridor: corridor, rate: rate, fee: fee, feeAsset: "INR"}
}

func (r *BankRoute) ID() routing.RouteID { return "bank_swift" }
func (r *BankRoute) Name() string { return "Bank (SWIFT/SEPA)" }
func (r *BankRoute) Supports(from, to, _, _ string) bool { return from+"-"+to == r.corridor }

func (r *BankRoute) Quote(_ context.Context, from, to string, amount decimal.Decimal) (*routing.RouteQuote, error) {
	if !r.Supports(from, to, "", "") {
		return nil, fmt.Errorf("bank SWIFT: unsupported pair %s-%s", from, to)
	}
	isInrToEur := from == "INR"
	var rate, destAmt decimal.Decimal
	if isInrToEur {
		rate = decimal.NewFromInt(1).Div(r.rate)
	} else {
		rate = r.rate
	}
	markup := decimal.NewFromFloat(0.003)
	effectiveRate := rate.Mul(decimal.NewFromInt(1).Sub(markup))
	destAmt = amount.Mul(effectiveRate)

	log.Debug().Str("route", "bank_swift").Str("from", from).Str("dest_amount", destAmt.String()).Msg("bank quote generated")

	return &routing.RouteQuote{
		RouteID: r.ID(), RouteName: r.Name(), SourceAsset: from, DestAsset: to,
		SourceAmount: amount, DestAmount: destAmt.Round(2), Rate: effectiveRate.Round(4),
		SpreadBps: 30, Fee: r.fee, FeeAsset: r.feeAsset,
		SettlementTime: 48 * time.Hour, ExpiresAt: time.Now().Add(60 * time.Second), Provider: "swift_gpi",
	}, nil
}

func (r *BankRoute) Execute(_ context.Context, _ routing.PaymentRequest, _ *routing.RouteQuote) (string, error) {
	return fmt.Sprintf("SWIFT-GPI-%d", time.Now().UnixNano()%1000000), nil
}
func (r *BankRoute) Status(_ context.Context, _ string) (string, error) { return "processing", nil }
