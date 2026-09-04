package fx

import (
	"context"
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

// StaticProvider serves fixed, operator-configured rates. It exists so the
// fiat on/off-ramp can quote fiat→crypto pairs on the XDC Apothem testnet
// model, where no live market for TXDC exists (CoinGecko lists XDC spot,
// but the testnet asset has no price feed, and fiat pairs like NGN-USDC are
// not served by any live provider in this deployment).
//
// Rates are configured as a map of "FROM-TO" → mid-market rate (units of
// `to` per one unit of `from`), e.g. "NGN-USDC": 0.000625.
//
// Status: testnet scaffolding only. A production deployment must replace
// this with live rate feeds per corridor.
type StaticProvider struct {
	rates map[string]decimal.Decimal
}

// NewStaticProvider builds a provider from a fixed rate table. Unknown
// pairs return domain-unsupported errors via GetRate.
func NewStaticProvider(rates map[string]decimal.Decimal) *StaticProvider {
	cp := make(map[string]decimal.Decimal, len(rates))
	for k, v := range rates {
		cp[k] = v
	}
	return &StaticProvider{rates: cp}
}

// ParseStaticRates parses "FROM-TO=rate,FROM-TO=rate" (rate as decimal)
// into the map NewStaticProvider expects. Invalid entries are skipped.
func ParseStaticRates(spec string) map[string]decimal.Decimal {
	out := map[string]decimal.Decimal{}
	for _, entry := range strings.Split(spec, ",") {
		pair, raw, ok := strings.Cut(strings.TrimSpace(entry), "=")
		if !ok {
			continue
		}
		r, err := decimal.NewFromString(strings.TrimSpace(raw))
		if err != nil || !r.GreaterThan(decimal.Zero) {
			continue
		}
		out[strings.ToUpper(strings.TrimSpace(pair))] = r
	}
	return out
}

func (p *StaticProvider) SupportedPairs() []string {
	pairs := make([]string, 0, len(p.rates))
	for k := range p.rates {
		pairs = append(pairs, k)
	}
	return pairs
}

func (p *StaticProvider) GetRate(_ context.Context, from, to, _ string) (decimal.Decimal, error) {
	r, ok := p.rates[from+"-"+to]
	if !ok {
		return decimal.Zero, fmt.Errorf("static fx: no rate for pair %s-%s", from, to)
	}
	return r, nil
}
