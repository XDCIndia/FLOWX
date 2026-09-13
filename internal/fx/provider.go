package fx

import (
	"context"

	"github.com/shopspring/decimal"
)

// Provider fetches mid-market FX rates for a set of currency pairs.
// Swapping the implementation (e.g. CoinGecko → third-party oracle) requires
// only changing the concrete type passed to NewService — no handler changes.
type Provider interface {
	// GetRate returns the mid-market rate: units of `to` per one unit of `from`.
	GetRate(ctx context.Context, from, to, amount string) (decimal.Decimal, error)
	// SupportedPairs returns supported pairs in "FROM-TO" format.
	SupportedPairs() []string
}
