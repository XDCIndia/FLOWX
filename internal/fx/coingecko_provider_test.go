package fx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoinGeckoProvider_SupportedPairs(t *testing.T) {
	p := NewCoinGeckoProvider("")
	pairs := p.SupportedPairs()
	assert.Contains(t, pairs, "USDC-XDC")
	assert.Contains(t, pairs, "XDC-USDC")
	assert.Contains(t, pairs, "USDC-TXDC")
	assert.Contains(t, pairs, "TXDC-USDC")
	assert.Contains(t, pairs, "NGN-TXDC")
	assert.Contains(t, pairs, "TXDC-NGN")
	assert.Contains(t, pairs, "USD-NGN")
	assert.Contains(t, pairs, "XLM-TXDC")
	assert.NotContains(t, pairs, "USDC-USDC")
	assert.NotContains(t, pairs, "BTC-USDC")
}

func TestCoinGeckoProvider_GetRate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/simple/price", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "ids=xdce-crowd-sale")
		assert.Contains(t, r.URL.RawQuery, "vs_currencies=usd")
		assert.Contains(t, r.URL.RawQuery, "ngn")
		_, _ = w.Write([]byte(`{"xdce-crowd-sale":{"usd":0.024,"ngn":38},"usd-coin":{"usd":1.002,"ngn":1590},"stellar":{"usd":0.1,"ngn":160}}`))
	}))
	defer srv.Close()

	p := NewCoinGeckoProvider(srv.URL)
	ctx := context.Background()

	cases := map[string]string{
		// crypto -> crypto via USD spot
		"XDC-USDC":  "0.0239520958083832335329",
		"USDC-XDC":  "41.75",
		"TXDC-USDC": "0.0239520958083832335329",
		"XLM-TXDC":  "4.1666666666666666667",
		// crypto -> fiat direct
		"TXDC-NGN": "38",
		"XDC-USD":  "0.024",
		// fiat -> crypto inverse
		"NGN-TXDC": "0.02631578947368421053",
		"USD-XDC":  "41.666666666666666667",
		// fiat -> fiat via USDC anchor
		"USD-NGN": "1586.8263473053892215569",
		"NGN-USD": "0.0006301886792453",
	}
	for pair, wantStr := range cases {
		from, to := splitPair(pair)
		rate, err := p.GetRate(ctx, from, to, "1")
		require.NoError(t, err, pair)
		want, err := decimal.NewFromString(wantStr)
		require.NoError(t, err, pair)
		diff := rate.Sub(want).Abs()
		assert.True(t, diff.LessThan(decimal.NewFromFloat(1e-8)),
			"%s rate = %s, want %s", pair, rate, want)
	}
}

func TestCoinGeckoProvider_UnknownPair(t *testing.T) {
	p := NewCoinGeckoProvider("")
	_, err := p.GetRate(context.Background(), "BTC", "ETH", "1")
	assert.Error(t, err)
}

func splitPair(pair string) (string, string) {
	for i := 0; i < len(pair); i++ {
		if pair[i] == '-' {
			return pair[:i], pair[i+1:]
		}
	}
	return pair, ""
}
