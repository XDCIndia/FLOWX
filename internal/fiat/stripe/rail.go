// Package stripe implements fiat.Rail on Stripe. Deposits are collected
// with a Stripe Checkout Session (card, bank debit via Financial
// Connections where available); withdrawals settle as Stripe Connect
// transfers to a destination account.
//
// Status: sandbox scaffold. What works end-to-end locally:
//   - Deposit → Checkout Session URL (open in browser, pay with Stripe
//     test cards) → checkout.session.completed webhook → crypto credit.
//   - Webhook signature verification with STRIPE_WEBHOOK_SECRET.
//
// What needs real Stripe work before production: withdrawal bank payouts
// require Connect destination accounts with KYC (the sandbox path treats
// WithdrawRequest.AccountNumber as a connected account id, acct_…), and
// per-corridor availability of bank-debit payment methods must be checked.
package stripe

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
	"github.com/stripe/stripe-go/v81/transfer"
	"github.com/stripe/stripe-go/v81/webhook"

	"github.com/fluxa/fluxa/internal/fiat"
)

// Event types emitted by this rail (matched against fiat.Event* constants
// by value in the fiat service).
const (
	eventDepositConfirmed  = "deposit.confirmed"
	eventDepositFailed     = "deposit.failed"
	eventWithdrawalSent    = "withdrawal.sent"
	eventWithdrawalFailed  = "withdrawal.failed"
	statusCompleted        = "completed"
	statusFailed           = "failed"
	defaultSuccessURL      = "http://localhost:3001/wallets?deposit=success"
	defaultCancelURL       = "http://localhost:3001/fiat?deposit=cancelled"
)

// Rail is a fiat.Rail backed by Stripe.
type Rail struct {
	webhookSecret string
	successURL    string
	cancelURL     string
}

// NewRail builds the Stripe rail. secretKey may be empty (the rail then
// refuses Deposit/Withdraw with a clear error, which keeps the API bootable
// without keys); webhookSecret is required for HandleWebhook.
func NewRail(secretKey, webhookSecret, successURL string) *Rail {
	if secretKey != "" {
		stripe.Key = secretKey
	}
	if successURL == "" {
		successURL = defaultSuccessURL
	}
	return &Rail{
		webhookSecret: webhookSecret,
		successURL:    successURL,
		cancelURL:     defaultCancelURL,
	}
}

// Deposit creates a Checkout Session for the fiat amount and returns its
// hosted URL as the payment link. The merchant reference rides in
// client_reference_id so the webhook can map back to the fiat_deposits row.
func (r *Rail) Deposit(_ context.Context, req fiat.DepositRequest) (*fiat.DepositResponse, error) {
	if stripe.Key == "" {
		return nil, fmt.Errorf("stripe: STRIPE_SECRET_KEY not configured")
	}
	// Stripe amounts are in the currency's minor units.
	units := req.FiatAmount.Mul(decimal.NewFromInt(100)).IntPart()
	if units <= 0 {
		return nil, fmt.Errorf("stripe: fiat amount %s too small for %s", req.FiatAmount, req.FiatCurrency)
	}

	params := &stripe.CheckoutSessionParams{
		Mode:              stripe.String(string(stripe.CheckoutSessionModePayment)),
		ClientReferenceID: stripe.String(req.Reference),
		SuccessURL:        stripe.String(r.successURL + "&ref=" + req.Reference),
		CancelURL:         stripe.String(r.cancelURL),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Quantity: stripe.Int64(1),
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency:   stripe.String(lowerASCII(req.FiatCurrency)),
					UnitAmount: stripe.Int64(units),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name: stripe.String("FlowX wallet deposit (" + req.WalletID + ")"),
					},
				},
			},
		},
	}
	if req.CustomerEmail != "" {
		params.CustomerEmail = stripe.String(req.CustomerEmail)
	}

	sess, err := session.New(params)
	if err != nil {
		return nil, fmt.Errorf("stripe: create checkout session: %w", err)
	}
	return &fiat.DepositResponse{
		PaymentLink: sess.URL,
		Reference:   req.Reference,
	}, nil
}

// Withdraw sends a Connect transfer to the destination account.
//
// Sandbox mapping: WithdrawRequest.AccountNumber carries the Stripe
// connected account id (acct_…). Production must replace this with bank
// payout via a verified Connect account with KYC.
func (r *Rail) Withdraw(_ context.Context, req fiat.WithdrawRequest) (*fiat.WithdrawResponse, error) {
	if stripe.Key == "" {
		return nil, fmt.Errorf("stripe: STRIPE_SECRET_KEY not configured")
	}
	destination := req.AccountNumber
	if destination == "" {
		return nil, fmt.Errorf("stripe: withdrawal requires a destination connected account (acct_…)")
	}
	units := req.FiatAmount.Mul(decimal.NewFromInt(100)).IntPart()
	if units <= 0 {
		return nil, fmt.Errorf("stripe: fiat amount %s too small for %s", req.FiatAmount, req.FiatCurrency)
	}

	t, err := transfer.New(&stripe.TransferParams{
		Amount:      stripe.Int64(units),
		Currency:    stripe.String(lowerASCII(req.FiatCurrency)),
		Destination: stripe.String(destination),
		TransferGroup: stripe.String(req.Reference),
	})
	if err != nil {
		return nil, fmt.Errorf("stripe: create transfer: %w", err)
	}
	return &fiat.WithdrawResponse{
		Reference: t.ID,
		Status:    statusCompleted,
	}, nil
}

// HandleWebhook verifies the Stripe signature and maps the event to a
// fiat.RailEvent. Unsupported event types return an error so the fiat
// service ignores them deterministically.
func (r *Rail) HandleWebhook(_ context.Context, payload []byte, signature string) (*fiat.RailEvent, error) {
	if r.webhookSecret == "" {
		return nil, fmt.Errorf("stripe: STRIPE_WEBHOOK_SECRET not configured")
	}
	event, err := webhook.ConstructEvent(payload, signature, r.webhookSecret)
	if err != nil {
		return nil, fmt.Errorf("stripe: webhook signature: %w", err)
	}

	switch event.Type {
	case "checkout.session.completed":
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			return nil, fmt.Errorf("stripe: decode session: %w", err)
		}
		if sess.ClientReferenceID == "" {
			return nil, fmt.Errorf("stripe: session %s has no client_reference_id", sess.ID)
		}
		if sess.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
			return nil, fmt.Errorf("stripe: session %s payment status %s", sess.ID, sess.PaymentStatus)
		}
		return &fiat.RailEvent{
			Type:        eventDepositConfirmed,
			ProviderRef: sess.ClientReferenceID,
			EventID:     sess.ID,
			Status:      statusCompleted,
			Amount:      decimal.NewFromInt(sess.AmountTotal).Div(decimal.NewFromInt(100)),
			Currency:    upperASCII(string(sess.Currency)),
		}, nil

	case "checkout.session.async_payment_failed", "checkout.session.expired":
		var sess stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &sess); err != nil {
			return nil, fmt.Errorf("stripe: decode session: %w", err)
		}
		if sess.ClientReferenceID == "" {
			return nil, fmt.Errorf("stripe: session %s has no client_reference_id", sess.ID)
		}
		return &fiat.RailEvent{
			Type:        eventDepositFailed,
			ProviderRef: sess.ClientReferenceID,
			EventID:     sess.ID,
			Status:      statusFailed,
			Amount:      decimal.NewFromInt(sess.AmountTotal).Div(decimal.NewFromInt(100)),
			Currency:    upperASCII(string(sess.Currency)),
		}, nil

	case "transfer.paid":
		var tr stripe.Transfer
		if err := json.Unmarshal(event.Data.Raw, &tr); err != nil {
			return nil, fmt.Errorf("stripe: decode transfer: %w", err)
		}
		return &fiat.RailEvent{
			Type:        eventWithdrawalSent,
			ProviderRef: tr.TransferGroup,
			EventID:     tr.ID,
			Status:      statusCompleted,
			Amount:      decimal.NewFromInt(tr.Amount).Div(decimal.NewFromInt(100)),
			Currency:    upperASCII(string(tr.Currency)),
		}, nil

	case "transfer.failed":
		var tr stripe.Transfer
		if err := json.Unmarshal(event.Data.Raw, &tr); err != nil {
			return nil, fmt.Errorf("stripe: decode transfer: %w", err)
		}
		return &fiat.RailEvent{
			Type:        eventWithdrawalFailed,
			ProviderRef: tr.TransferGroup,
			EventID:     tr.ID,
			Status:      statusFailed,
			Amount:      decimal.NewFromInt(tr.Amount).Div(decimal.NewFromInt(100)),
			Currency:    upperASCII(string(tr.Currency)),
		}, nil

	default:
		return nil, fmt.Errorf("stripe: unhandled event type %s", event.Type)
	}
}

func lowerASCII(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		}
	}
	return string(b)
}

func upperASCII(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'a' && c <= 'z' {
			b[i] = c - 32
		}
	}
	return string(b)
}
