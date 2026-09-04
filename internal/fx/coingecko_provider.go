package fx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/fluxa/fluxa/internal/domain"
	"github.com/shopspring/decimal"
)

// CoinGeckoProvider prices XDC against USDC using CoinGecko's public
// simple/price endpoint. XDC is not a Stellar asset, so the Horizon order
// book cannot price it; this provider is the XDC-backend counterpart.
type CoinGeckoProvider struct {
	client  *http.Client
	baseURL string
}

const coingeckoBaseURL = "https://api.coingecko.com/api/v3"

// NewCoinGeckoProvider creates a Provider backed by CoinGecko. baseURL is
// overridable for tests.
func NewCoinGeckoProvider(baseURL string) *CoinGeckoProvider {
	if baseURL == "" {
		baseURL = coingeckoBaseURL
	}
	return &CoinGeckoProvider{
		client:  &http.Client{Timeout: 5 * time.Second},
		baseURL: baseURL,
	}
}

func (p *CoinGeckoProvider) SupportedPairs() []string {
	return []string{"USDC-XDC", "XDC-USDC"}
}

// GetRate returns units of `to` per one unit of `from`, derived from
// CoinGecko USD spot prices for XDC and USDC.
func (p *CoinGeckoProvider) GetRate(ctx context.Context, from, to, _ string) (decimal.Decimal, error) {
	// CoinGecko's id for XDC Network is "xdce-crowd-sale" (legacy listing).
	url := fmt.Sprintf("%s/simple/price?ids=xdce-crowd-sale,usd-coin&vs_currencies=usd", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return decimal.Zero, fmt.Errorf("build coingecko request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return decimal.Zero, fmt.Errorf("coingecko price for %s-%s: %w", from, to, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("coingecko price for %s-%s: status %d", from, to, resp.StatusCode)
	}

	var prices map[string]map[string]json.Number
	if err := json.NewDecoder(resp.Body).Decode(&prices); err != nil {
		return decimal.Zero, fmt.Errorf("decode coingecko prices: %w", err)
	}

	xdcUsd, err := usdPrice(prices, "xdce-crowd-sale")
	if err != nil {
		return decimal.Zero, fmt.Errorf("coingecko price for %s-%s: %w", from, to, err)
	}
	usdcUsd, err := usdPrice(prices, "usd-coin")
	if err != nil {
		return decimal.Zero, fmt.Errorf("coingecko price for %s-%s: %w", from, to, err)
	}

	switch from + "-" + to {
	case "XDC-USDC":
		if usdcUsd.IsZero() {
			return decimal.Zero, fmt.Errorf("%w: no liquidity for %s-%s", domain.ErrNoLiquidity, from, to)
		}
		return xdcUsd.Div(usdcUsd), nil
	case "USDC-XDC":
		if xdcUsd.IsZero() {
			return decimal.Zero, fmt.Errorf("%w: no liquidity for %s-%s", domain.ErrNoLiquidity, from, to)
		}
		return usdcUsd.Div(xdcUsd), nil
	default:
		return decimal.Zero, fmt.Errorf("%w: no provider for pair %s-%s", domain.ErrUnsupportedPair, from, to)
	}
}

func usdPrice(prices map[string]map[string]json.Number, id string) (decimal.Decimal, error) {
	entry, ok := prices[id]
	if !ok {
		return decimal.Zero, fmt.Errorf("missing price for %s", id)
	}
	raw, ok := entry["usd"]
	if !ok {
		return decimal.Zero, fmt.Errorf("missing usd price for %s", id)
	}
	d, err := decimal.NewFromString(raw.String())
	if err != nil {
		return decimal.Zero, fmt.Errorf("parse usd price for %s: %w", id, err)
	}
	if d.Sign() < 0 {
		return decimal.Zero, fmt.Errorf("negative price for %s", id)
	}
	return d, nil
}