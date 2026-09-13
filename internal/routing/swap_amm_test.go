package routing

import (
	"context"
	"errors"
	"math/big"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixedReserves returns a reserveReader with the canonical test pool:
// 10 TXDC (1e19 wei) paired with 35,000 tUSDC (6 decimals, raw 3.5e10).
func fixedReserves(txdc *big.Int, token *big.Int) reserveReader {
	return func(context.Context) (*big.Int, *big.Int, error) {
		return new(big.Int).Set(txdc), new(big.Int).Set(token), nil
	}
}

func newTestAMMRoute(txdcReserve, tokenReserve *big.Int, tokenDec uint8) *AMMSwapRoute {
	r := NewAMMSwapRoute("", "0x1111111111111111111111111111111111111111",
		"0x2222222222222222222222222222222222222222", "deadbeef")
	r.tokenDec = tokenDec
	r.readReserves = fixedReserves(txdcReserve, tokenReserve)
	return r
}

var (
	testReserveTXDC  = new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18)) // 10 TXDC
	testReserveToken = big.NewInt(35000 * 1e6)                            // 35,000 tUSDC (6 dec)
)

func TestAMMAmountOut_ConstantProductWithFee(t *testing.T) {
	// 1 TXDC in against (10 TXDC, 35,000 USDC): expected 3,173.138128 USDC.
	// Verified against the Solidity formula:
	// out = in*997*reserveOut / (reserveIn*1000 + in*997)
	out := AMMAmountOut(big.NewInt(1e18), testReserveTXDC, testReserveToken)
	assert.Equal(t, big.NewInt(3173138128), out)

	// Reverse direction: 100 USDC in -> 0.028404801180636871 TXDC.
	out = AMMAmountOut(big.NewInt(100*1e6), testReserveToken, testReserveTXDC)
	assert.Equal(t, new(big.Int).SetUint64(28404801180636871), out)
}

func TestAMMAmountOut_ZeroOutput(t *testing.T) {
	// A 1-wei input rounds to zero — callers must reject.
	assert.Equal(t, int64(0), AMMAmountOut(big.NewInt(1), testReserveTXDC, testReserveToken).Int64())
}

func TestMinOutWithSlippage(t *testing.T) {
	out := big.NewInt(3173138128)
	assert.Equal(t, big.NewInt(3141406746), minOutWithSlippage(out, 100)) // 1%
	assert.Equal(t, big.NewInt(3157272437), minOutWithSlippage(out, 50))  // 0.5%
	assert.Equal(t, big.NewInt(0), minOutWithSlippage(out, 10000))        // 100%
}

func TestAMMSwapRoute_Quote_TXDCtoUSDC(t *testing.T) {
	r := newTestAMMRoute(testReserveTXDC, testReserveToken, 6)

	q, err := r.Quote(context.Background(), "TXDC", "USDC", decimal.NewFromInt(1))
	require.NoError(t, err)

	assert.Equal(t, RouteAMMSwap, q.RouteID)
	assert.Equal(t, "TXDC", q.SourceAsset)
	assert.Equal(t, "USDC", q.DestAsset)
	// 3,173.138128 tUSDC out (6 decimals preserved).
	assert.True(t, q.DestAmount.Equal(decimal.NewFromFloat(3173.138128)),
		"dest amount: got %s", q.DestAmount)
	// 0.3% fee in the source asset.
	assert.True(t, q.Fee.Equal(decimal.NewFromFloat(0.003)), "fee: got %s", q.Fee)
	assert.Equal(t, "TXDC", q.FeeAsset)
	assert.Equal(t, 0, q.SpreadBps)
	assert.Equal(t, "flowx_amm", q.Provider)
}

func TestAMMSwapRoute_Quote_USDCtoTXDC(t *testing.T) {
	r := newTestAMMRoute(testReserveTXDC, testReserveToken, 6)

	q, err := r.Quote(context.Background(), "USDC", "TXDC", decimal.NewFromInt(100))
	require.NoError(t, err)

	// 0.028404801180636871 TXDC out (18 decimals preserved).
	assert.True(t, q.DestAmount.Equal(decimal.RequireFromString("0.028404801180636871")),
		"dest amount: got %s", q.DestAmount)
	assert.True(t, q.Fee.Equal(decimal.NewFromFloat(0.3)), "fee: got %s", q.Fee)
	assert.Equal(t, "USDC", q.FeeAsset)
	// Rate = dest/source ≈ 0.000284048...
	assert.True(t, q.Rate.Mul(decimal.NewFromInt(100)).Sub(q.DestAmount).Abs().LessThan(decimal.NewFromFloat(1e-12)))
}

func TestAMMSwapRoute_Quote_Errors(t *testing.T) {
	ctx := context.Background()

	t.Run("unsupported pair", func(t *testing.T) {
		r := newTestAMMRoute(testReserveTXDC, testReserveToken, 6)
		_, err := r.Quote(ctx, "NGN", "KES", decimal.NewFromInt(10))
		require.Error(t, err)
	})

	t.Run("zero amount", func(t *testing.T) {
		r := newTestAMMRoute(testReserveTXDC, testReserveToken, 6)
		_, err := r.Quote(ctx, "TXDC", "USDC", decimal.Zero)
		require.Error(t, err)
	})

	t.Run("empty pool", func(t *testing.T) {
		r := newTestAMMRoute(big.NewInt(0), big.NewInt(0), 6)
		_, err := r.Quote(ctx, "TXDC", "USDC", decimal.NewFromInt(1))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no liquidity")
	})

	t.Run("output rounds to zero", func(t *testing.T) {
		r := newTestAMMRoute(testReserveTXDC, testReserveToken, 6)
		// 1e-17 TXDC is below the pool's precision.
		_, err := r.Quote(ctx, "TXDC", "USDC", decimal.NewFromFloat(1e-17))
		require.Error(t, err)
	})
}

func TestAMMSwapRoute_MissingConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("pool unset", func(t *testing.T) {
		r := NewAMMSwapRoute("", "", "0x2222222222222222222222222222222222222222", "key")
		_, err := r.Quote(ctx, "TXDC", "USDC", decimal.NewFromInt(1))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "AMM_POOL_ADDRESS")
	})

	t.Run("token unset", func(t *testing.T) {
		r := NewAMMSwapRoute("", "0x1111111111111111111111111111111111111111", "", "key")
		_, err := r.Quote(ctx, "TXDC", "USDC", decimal.NewFromInt(1))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "XDC_USDC_CONTRACT_ADDRESS")
	})

	t.Run("both unset", func(t *testing.T) {
		r := NewAMMSwapRoute("", "", "", "key")
		_, err := r.Execute(ctx, PaymentRequest{SourceAsset: "TXDC", DestAsset: "USDC"}, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "AMM_POOL_ADDRESS")
	})

	t.Run("status missing config", func(t *testing.T) {
		r := NewAMMSwapRoute("", "", "", "key")
		_, err := r.Status(ctx, "0xabc")
		require.Error(t, err)
	})
}

func TestAMMSwapRoute_Supports(t *testing.T) {
	r := NewAMMSwapRoute("", "0x1111", "0x2222", "key")
	assert.True(t, r.Supports("TXDC", "USDC", "", ""))
	assert.True(t, r.Supports("USDC", "TXDC", "", ""))
	assert.False(t, r.Supports("NGN", "KES", "", ""))
	assert.False(t, r.Supports("TXDC", "INR", "", ""))
}

func TestAMMSwapRoute_Quote_ReserveReaderError(t *testing.T) {
	r := NewAMMSwapRoute("", "0x1111111111111111111111111111111111111111",
		"0x2222222222222222222222222222222222222222", "key")
	r.tokenDec = 6
	r.readReserves = func(context.Context) (*big.Int, *big.Int, error) {
		return nil, nil, errors.New("rpc down")
	}
	_, err := r.Quote(context.Background(), "TXDC", "USDC", decimal.NewFromInt(1))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rpc down")
}
