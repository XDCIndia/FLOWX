package routing

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
)

// StripeBankRoute implements PaymentRoute using Stripe for bank/card payments.
// Supports fiat→fiat corridors (INR-EUR, EUR-INR, NGN-USDC, etc.)
type StripeBankRoute struct {
	stripeKey string
	corridors map[string]corridorConfig
}

type corridorConfig struct {
	feePercent   float64 // Stripe fee percentage
	fixedFee     int64   // Fixed fee in minor units (cents/paise)
	settlement   time.Duration
	feeAsset     string
	rate         float64 // mock rate: how many dest units per 1 source unit
	rateNote     string
	paymentTypes []string // "card", "sepa_debit", "bank_transfer"
}

func NewStripeBankRoute(stripeKey string) *StripeBankRoute {
	return &StripeBankRoute{
		stripeKey: stripeKey,
		corridors: map[string]corridorConfig{
			"INR-EUR": {
				feePercent: 2.0, fixedFee: 300, // 2% + ₹3 (~$0.03)
				settlement: 3 * 24 * time.Hour, // 3 days (bank transfer)
				feeAsset:   "INR", rate: 0.0109,
				paymentTypes: []string{"bank_transfer", "card"},
			},
			"EUR-INR": {
				feePercent: 1.5, fixedFee: 20, // 1.5% + €0.20
				settlement: 2 * 24 * time.Hour,
				feeAsset:   "EUR", rate: 91.5,
				paymentTypes: []string{"bank_transfer", "card"},
			},
			"NGN-USDC": {
				feePercent: 1.4, fixedFee: 100, // 1.4% + ₦1
				settlement: 2 * 24 * time.Hour,
				feeAsset:   "NGN", rate: 1500,
				paymentTypes: []string{"card", "bank_transfer"},
			},
			"USDC-NGN": {
				feePercent: 1.0, fixedFee: 50,
				settlement: 1 * 24 * time.Hour,
				feeAsset:   "USDC", rate: 0.000667,
				paymentTypes: []string{"bank_transfer"},
			},
		},
	}
}

func (r *StripeBankRoute) ID() RouteID   { return "stripe_bank" }
func (r *StripeBankRoute) Name() string  { return "Bank (Stripe)" }

func (r *StripeBankRoute) Supports(from, to, _, _ string) bool {
	_, ok := r.corridors[from+"-"+to]
	return ok
}

func (r *StripeBankRoute) Quote(_ context.Context, from, to string, amount decimal.Decimal) (*RouteQuote, error) {
	key := from + "-" + to
	cfg, ok := r.corridors[key]
	if !ok {
		return nil, fmt.Errorf("stripe bank: unsupported pair %s-%s", from, to)
	}

	// Calculate fee
	feePct := decimal.NewFromFloat(cfg.feePercent).Div(decimal.NewFromInt(100))
	feeFromAmount := amount.Mul(feePct)
	fixedFee := decimal.NewFromInt(cfg.fixedFee).Div(decimal.NewFromInt(100))
	fee := feeFromAmount.Add(fixedFee).Round(2)

	// Calculate destination amount after fee and rate
	amountAfterFee := amount.Sub(fee)
	rate := decimal.NewFromFloat(cfg.rate)
	destAmt := amountAfterFee.Mul(rate).Round(2)

	// Effective rate (including fees)
	effectiveRate := destAmt.Div(amount).Round(6)

	spreadBps := int(cfg.feePercent * 100) // approximate spread in bps

	log.Debug().
		Str("route", "stripe_bank").
		Str("from", from).Str("to", to).
		Str("dest_amount", destAmt.String()).
		Str("fee", fee.String()).
		Msg("stripe bank quote generated")

	return &RouteQuote{
		RouteID:        r.ID(),
		RouteName:      r.Name(),
		SourceAsset:    from,
		DestAsset:      to,
		SourceAmount:   amount,
		DestAmount:     destAmt,
		Rate:           effectiveRate,
		SpreadBps:      spreadBps,
		Fee:            fee,
		FeeAsset:       cfg.feeAsset,
		SettlementTime: cfg.settlement,
		ExpiresAt:      time.Now().Add(60 * time.Second),
		Provider:       "stripe",
	}, nil
}

func (r *StripeBankRoute) Execute(ctx context.Context, req PaymentRequest, quote *RouteQuote) (string, error) {
	key := req.SourceAsset + "-" + req.DestAsset
	cfg, ok := r.corridors[key]
	if !ok {
		return "", fmt.Errorf("stripe bank: unsupported corridor %s", key)
	}

	// Create a Stripe Checkout Session so the user can actually pay
	if stripe.Key == "" {
		stripe.Key = r.stripeKey
	}
	if stripe.Key != "" {
		minorUnits := req.Amount.Mul(decimal.NewFromInt(100)).IntPart()

		params := &stripe.CheckoutSessionParams{
			Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
			PaymentIntentData: &stripe.CheckoutSessionPaymentIntentDataParams{
				Description: stripe.String(fmt.Sprintf("FlowX: %s %s → %s via Stripe", req.SourceAsset, req.Amount.StringFixed(2), req.DestAsset)),
				Metadata: map[string]string{
					"source_asset": req.SourceAsset,
					"dest_asset":   req.DestAsset,
					"amount":       req.Amount.StringFixed(2),
					"route":        "stripe_bank",
				},
			},
			LineItems: []*stripe.CheckoutSessionLineItemParams{
				{
					Quantity: stripe.Int64(1),
					PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
						Currency:   stripe.String(lowerStripe(string(req.SourceAsset))),
						UnitAmount: stripe.Int64(minorUnits),
						ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
							Name: stripe.String(fmt.Sprintf("Cross-border: %s → %s", req.SourceAsset, req.DestAsset)),
						},
					},
				},
			},
			SuccessURL: stripe.String("http://localhost:3001/payments?stripe=success&session_id={CHECKOUT_SESSION_ID}"),
			CancelURL:  stripe.String("http://localhost:3001/payments?stripe=cancelled"),
		}

		sess, err := session.New(params)
		if err != nil {
			log.Warn().Err(err).Msg("stripe checkout session failed, falling back to mock")
		} else {
			log.Info().Str("session_id", sess.ID).Str("url", sess.URL).Msg("stripe checkout session created")
			return sess.URL, nil
		}
	}

	// Fallback: generate a mock reference with realistic Stripe format
	ref := fmt.Sprintf("STRIPE-%d", time.Now().UnixNano()%1000000)
	log.Info().Str("reference", ref).Str("corridor", key).Msg("stripe bank payment initiated (sandbox)")

	_ = cfg // used for config but not needed in mock path
	return ref, nil
}

func (r *StripeBankRoute) Status(_ context.Context, reference string) (string, error) {
	return "processing", nil
}

func lowerStripe(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}
