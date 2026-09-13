package fiat

import (
	"context"
	"strings"
	"testing"

	"github.com/shopspring/decimal"

	"github.com/fluxa/fluxa/internal/routing"
)

// TestPaymentNetworkRoute_ExecuteRefuses verifies the fabricated Ripple ODL
// execution is gone: Execute must fail loudly instead of minting a fake
// "RIPPLE-ODL-%d" reference.
func TestPaymentNetworkRoute_ExecuteRefuses(t *testing.T) {
	r := NewPaymentNetworkRoute("INR-EUR", nil)
	ref, err := r.Execute(context.Background(), routing.PaymentRequest{
		SourceAsset: "INR", DestAsset: "EUR", Amount: decimal.NewFromInt(1000),
	}, nil)
	if err == nil {
		t.Fatalf("Execute must error, got reference %q", ref)
	}
	if !strings.Contains(err.Error(), "liquidity-provider partnership") {
		t.Errorf("unexpected error: %v", err)
	}
	if strings.Contains(ref, "RIPPLE-ODL") {
		t.Errorf("fabricated reference leaked: %q", ref)
	}
}
