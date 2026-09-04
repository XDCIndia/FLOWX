package fx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

// CoinGeckoProvider prices the live asset universe against CoinGecko spot
// data. One memoized /simple/price call (crypto IDs in USD + NGN) backs
// every supported pair, so quotes track the real market instead of the
// static testnet rates.
//
// Asset codes: TXDC (Apothem testnet XDC) is priced as XDC
// ("xdce-crowd-sale"); USDC and XLM map to their CoinGecko ids. USD and
// NGN are fiat legs priced through the same call.
type CoinGeckoProvider struct {
	client  *http.Client
	baseURL string

	mu      sync.Mutex
	prices  map[string]map[string]decimal.Decimal // id -> fiat -> price
	fetched time.Time
}

const coingeckoBaseURL = "https://api.coingecko.com/api/v3"

// coinGeckoIDs maps Fluxa asset codes to CoinGecko coin ids.
var coinGeckoIDs = map[string]string{
	"XDC":  "xdce-crowd-sale",
	"TXDC": "xdce-crowd-sale",
	"USDC": "usd-coin",
	"XLM":  "stellar",
}

// coinGeckoFiat are fiat codes usable as quote legs via vs_currencies.
var coinGeckoFiat = map[string]bool{
	"USD": true,
	"NGN": true,
}

// coinGeckoAssets is the full pair universe advertised by SupportedPairs.
var coinGeckoAssets = []string{"USDC", "XDC", "TXDC", "XLM", "USD", "NGN"}

const coinGeckoPriceTTL = 30 * time.Second

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
	pairs := make([]string, 0, len(coinGeckoAssets)*(len(coinGeckoAssets)-1))
	for _, from := range coinGeckoAssets {
		for _, to := range coinGeckoAssets {
			if from != to {
				pairs = append(pairs, from+"-"+to)
			}
		}
	}
	return pairs
}

// GetRate returns units of `to` per one unit of `from` from live spot data.
func (p *CoinGeckoProvider) GetRate(ctx context.Context, from, to, _ string) (decimal.Decimal, error) {
	fromID, fromCrypto := coinGeckoIDs[from]
	toID, toCrypto := coinGeckoIDs[to]
	_, fromFiat := coinGeckoFiat[from]
	_, toFiat := coinGeckoFiat[to]
	if !((fromCrypto || fromFiat) && (toCrypto || toFiat)) {
		return decimal.Zero, fmt.Errorf("coingecko: unsupported pair %s-%s", from, to)
	}

	prices, err := p.priceMatrix(ctx)
	if err != nil {
		return decimal.Zero, err
	}

	// crypto -> crypto: cross via USD spot.
	if fromCrypto && toCrypto {
		fp, ok1 := prices[fromID]["usd"]
		tp, ok2 := prices[toID]["usd"]
		if !ok1 || !ok2 || tp.IsZero() {
			return decimal.Zero, fmt.Errorf("coingecko: missing usd price for %s-%s", from, to)
		}
		return fp.Div(tp), nil
	}
	// crypto -> fiat: direct vs_currencies price.
	if fromCrypto && toFiat {
		fp, ok := prices[fromID][strings.ToLower(to)]
		if !ok || fp.IsZero() {
			return decimal.Zero, fmt.Errorf("coingecko: missing %s price for %s", to, from)
		}
		return fp, nil
	}
	// fiat -> crypto: inverse of the crypto's fiat price.
	if fromFiat && toCrypto {
		tp, ok := prices[toID][strings.ToLower(from)]
		if !ok || tp.IsZero() {
			return decimal.Zero, fmt.Errorf("coingecko: missing %s price for %s", from, to)
		}
		return decimal.NewFromInt(1).Div(tp), nil
	}
	// fiat -> fiat: cross both fiats through a crypto quote. USDC is
	// preferred as the anchor (USD-pegged, so it tracks fiat feeds best).
	for _, id := range []string{"usd-coin", "xdce-crowd-sale", "stellar"} {
		pFrom, ok1 := prices[id][strings.ToLower(from)]
		pTo, ok2 := prices[id][strings.ToLower(to)]
		if ok1 && ok2 && !pFrom.IsZero() {
			return pTo.Div(pFrom), nil
		}
	}
	return decimal.Zero, fmt.Errorf("coingecko: cannot price %s-%s", from, to)
}

// priceMatrix fetches (and briefly memoizes) the USD+NGN spot matrix for
// every known crypto id.
func (p *CoinGeckoProvider) priceMatrix(ctx context.Context) (map[string]map[string]decimal.Decimal, error) {
	p.mu.Lock()
	if len(p.prices) > 0 && time.Since(p.fetched) < coinGeckoPriceTTL {
		defer p.mu.Unlock()
		return p.prices, nil
	}
	p.mu.Unlock()

	ids := make([]string, 0, len(coinGeckoIDs))
	seen := map[string]bool{}
	for _, id := range coinGeckoIDs {
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	url := fmt.Sprintf("%s/simple/price?ids=%s&vs_currencies=usd,ngn", p.baseURL, strings.Join(ids, ","))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build coingecko request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("coingecko price fetch: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("coingecko price fetch: status %d", resp.StatusCode)
	}

	var raw map[string]map[string]json.Number
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode coingecko response: %w", err)
	}
	matrix := make(map[string]map[string]decimal.Decimal, len(raw))
	for id, fiats := range raw {
		matrix[id] = make(map[string]decimal.Decimal, len(fiats))
		for fiat, num := range fiats {
			d, err := decimal.NewFromString(num.String())
			if err != nil {
				continue
			}
			matrix[id][fiat] = d
		}
	}

	p.mu.Lock()
	p.prices = matrix
	p.fetched = time.Now()
	p.mu.Unlock()
	return matrix, nil
}
