package fx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fluxa/fluxa/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCoinGeckoProvider_SupportedPairs(t *testing.T) {
	p := NewCoinGeckoProvider("")
	assert.Equal(t, []string{"USDC-XDC", "XDC-USDC"}, p.SupportedPairs())
}

func TestCoinGeckoProvider_GetRate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/simple/price", r.URL.Path)
		assert.Contains(t, r.URL.RawQuery, "ids=xdce-crowd-sale,usd-coin")
		_, _ = w.Write([]byte(`{"xdce-crowd-sale":{"usd":0.024},"usd-coin":{"usd":1.002}}`))
	}))
	defer srv.Close()

	p := NewCoinGeckoProvider(srv.URL)
	ctx := context.Background()

	// XDC-USDC: 1 XDC buys 0.024/1.002 USDC.
	rate, err := p.GetRate(ctx, "XDC", "USDC", "1")
	require.NoError(t, err)
	want := decimal.RequireFromString("0.024").Div(decimal.RequireFromString("1.002"))
	assert.True(t, rate.Equal(want), "XDC-USDC rate = %s, want %s", rate, want)

	// USDC-XDC: 1 USDC buys 1.002/0.024 XDC.
	rate, err = p.GetRate(ctx, "USDC", "XDC", "1")
	require.NoError(t, err)
	want, _ = decimal.NewFromString("41.75")
	assert.True(t, rate.Equal(want), "USDC-XDC rate = %s, want %s", rate, want)
}

func TestCoinGeckoProvider_GetRate_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("missing prices", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{}`))
		}))
		defer srv.Close()
		_, err := NewCoinGeckoProvider(srv.URL).GetRate(ctx, "USDC", "XDC", "1")
		require.Error(t, err)
	})

	t.Run("non-200", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer srv.Close()
		_, err := NewCoinGeckoProvider(srv.URL).GetRate(ctx, "USDC", "XDC", "1")
		require.Error(t, err)
	})

	t.Run("unsupported pair", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"xdce-crowd-sale":{"usd":0.024},"usd-coin":{"usd":1.002}}`))
		}))
		defer srv.Close()
		_, err := NewCoinGeckoProvider(srv.URL).GetRate(ctx, "BTC", "ETH", "1")
		require.ErrorIs(t, err, domain.ErrUnsupportedPair)
	})
}