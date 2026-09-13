package routing

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/rs/zerolog/log"
	"github.com/shopspring/decimal"

	"github.com/fluxa/fluxa/internal/chain/xdc"
	"github.com/fluxa/fluxa/internal/routing/amm"
)

// RouteAMMSwap is the on-chain swap route backed by our own constant-product
// AMM (contracts/amm/FlowXPool.sol) on XDC Apothem. It replaces the
// previously fabricated "Ripple ODL" leg with a real, verifiable on-chain
// leg: quotes come from live pool reserves (eth_call), Execute submits a
// real signed swap transaction, and Status polls the chain for the actual
// receipt — no fabricated references anywhere.
const RouteAMMSwap RouteID = "amm_swap"

// AMM fee/slippage constants. Keep FEE_NUM/FEE_DEN in sync with
// FlowXPool.sol (997/1000 = 0.3%).
const (
	ammFeeNum          = 997
	ammFeeDen          = 1000
	ammSlippageBps     = int64(100) // 1% default execution slippage
	ammGasLimitSwap    = uint64(300000)
	ammQuoteTTL        = 30 * time.Second
	ammSwapDeadline    = 5 * time.Minute
	ammReceiptTimeout  = 90 * time.Second
	ammReceiptPoll     = 2 * time.Second
	ammTokenDecDefault = uint8(6) // tUSDC assumption until chain resolution succeeds
)

// ammDirection is the swap direction within the pool.
type ammDirection int

const (
	dirTXDCToToken ammDirection = iota // TXDC (native, 18 dec) -> token
	dirTokenToTXDC                     // token -> TXDC
)

// reserveReader reads the pool's reserves. Injectable so unit tests can
// exercise the quote math without any network.
type reserveReader func(ctx context.Context) (reserveTXDC, reserveToken *big.Int, err error)

// AMMSwapRoute swaps TXDC <-> tUSDC via the FlowXPool AMM contract.
type AMMSwapRoute struct {
	rpcURL      string
	poolAddr    string
	tokenAddr   string
	treasuryKey string

	client   *ethclient.Client
	signer   types.Signer
	chainID  *big.Int
	tokenDec uint8 // 0 = resolve lazily from chain at quote time

	slippageBps  int64
	readReserves reserveReader
}

// NewAMMSwapRoute creates the AMM swap route. Empty poolAddr or tokenAddr
// leaves the route unconfigured: Quote returns an error and the evaluator
// skips it cleanly (cmd/api only registers the route when both are set).
func NewAMMSwapRoute(rpcURL, poolAddr, tokenAddr, treasuryKey string) *AMMSwapRoute {
	return &AMMSwapRoute{
		rpcURL:      strings.TrimSpace(rpcURL),
		poolAddr:    strings.TrimSpace(poolAddr),
		tokenAddr:   strings.TrimSpace(tokenAddr),
		treasuryKey: strings.TrimSpace(treasuryKey),
		chainID:     big.NewInt(xdc.ApothemChainID),
		slippageBps: ammSlippageBps,
	}
}

func (r *AMMSwapRoute) ID() RouteID  { return RouteAMMSwap }
func (r *AMMSwapRoute) Name() string { return "AMM Swap (FlowXPool)" }

func (r *AMMSwapRoute) Supports(from, to, _, _ string) bool {
	pair := from + "-" + to
	return pair == "TXDC-USDC" || pair == "USDC-TXDC"
}

// configErr reports whether the route is usable at all.
func (r *AMMSwapRoute) configErr() error {
	if r.poolAddr == "" {
		return fmt.Errorf("amm swap: AMM_POOL_ADDRESS is not configured")
	}
	if r.tokenAddr == "" {
		return fmt.Errorf("amm swap: XDC_USDC_CONTRACT_ADDRESS is not configured")
	}
	return nil
}

func ammDirectionFor(from, to string) (ammDirection, error) {
	switch from + "-" + to {
	case "TXDC-USDC":
		return dirTXDCToToken, nil
	case "USDC-TXDC":
		return dirTokenToTXDC, nil
	default:
		return dirTXDCToToken, fmt.Errorf("amm swap: unsupported pair %s-%s", from, to)
	}
}

// AMMAmountOut computes the constant-product output with the 0.3% fee:
// amountOut = amountIn*997*reserveOut / (reserveIn*1000 + amountIn*997).
// Mirrors FlowXPool.getAmountOut exactly — keep in sync with the contract.
func AMMAmountOut(amountIn, reserveIn, reserveOut *big.Int) *big.Int {
	amountInWithFee := new(big.Int).Mul(amountIn, big.NewInt(ammFeeNum))
	numerator := new(big.Int).Mul(amountInWithFee, reserveOut)
	denominator := new(big.Int).Add(
		new(big.Int).Mul(reserveIn, big.NewInt(ammFeeDen)),
		amountInWithFee,
	)
	return new(big.Int).Quo(numerator, denominator)
}

// minOutWithSlippage applies the slippage floor to a quoted output.
func minOutWithSlippage(out *big.Int, slippageBps int64) *big.Int {
	return new(big.Int).Quo(
		new(big.Int).Mul(out, big.NewInt(10000-slippageBps)),
		big.NewInt(10000),
	)
}

// ensureClient dials the RPC once and pins the EIP-155 signer to the
// verified chain ID, so a misconfigured RPC can never get a signed tx for
// the wrong network (same guard as internal/chain/xdc).
func (r *AMMSwapRoute) ensureClient(ctx context.Context) (*ethclient.Client, error) {
	if r.client != nil {
		return r.client, nil
	}
	if r.rpcURL == "" {
		return nil, fmt.Errorf("amm swap: XDC_RPC_URL is not configured")
	}
	ec, err := ethclient.Dial(r.rpcURL)
	if err != nil {
		return nil, fmt.Errorf("amm swap: dial %s: %w", r.rpcURL, err)
	}
	id, err := ec.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("amm swap: chainId: %w", err)
	}
	if r.chainID != nil && id.Cmp(r.chainID) != 0 {
		return nil, fmt.Errorf("amm swap: chainId mismatch: RPC says %d, want %d — refusing to use this endpoint", id.Int64(), r.chainID.Int64())
	}
	r.chainID = id
	r.signer = types.NewEIP155Signer(id)
	r.client = ec
	return ec, nil
}

// reserves reads (reserveTXDC wei, reserveToken raw) via the injected reader
// or, by default, an eth_call to FlowXPool.reserves().
func (r *AMMSwapRoute) reserves(ctx context.Context) (*big.Int, *big.Int, error) {
	if r.readReserves != nil {
		return r.readReserves(ctx)
	}
	ec, err := r.ensureClient(ctx)
	if err != nil {
		return nil, nil, err
	}
	pool, err := amm.NewFlowXPool(common.HexToAddress(ammNormalize(r.poolAddr)), ec)
	if err != nil {
		return nil, nil, fmt.Errorf("amm swap: bind pool: %w", err)
	}
	res0, res1, err := pool.Reserves(&bind.CallOpts{Context: ctx})
	if err != nil {
		return nil, nil, fmt.Errorf("amm swap: reserves call: %w", err)
	}
	return res0, res1, nil
}

// decimals resolves the token's decimals: from the configured/test value if
// set, else from the token contract itself. Falls back to 6 (tUSDC
// convention) with a warning rather than failing the whole quote when the
// RPC is flaky — amounts are still computed consistently off live reserves.
func (r *AMMSwapRoute) decimals(ctx context.Context) (uint8, error) {
	if r.tokenDec != 0 {
		return r.tokenDec, nil
	}
	ec, err := r.ensureClient(ctx)
	if err != nil {
		return ammTokenDecDefault, err
	}
	data := common.FromHex("0x313ce567") // decimals()
	to := common.HexToAddress(ammNormalize(r.tokenAddr))
	res, err := ec.CallContract(ctx, ethereum.CallMsg{To: &to, Data: data}, nil)
	if err != nil || len(res) < 32 {
		log.Warn().Err(err).Str("token", r.tokenAddr).
			Msg("amm swap: token decimals call failed, assuming 6")
		return ammTokenDecDefault, nil
	}
	d := big.NewInt(0).SetBytes(res).Uint64()
	if d > 77 { // absurd value; treat as malformed
		log.Warn().Uint64("decimals", d).Str("token", r.tokenAddr).
			Msg("amm swap: implausible token decimals, assuming 6")
		return ammTokenDecDefault, nil
	}
	r.tokenDec = uint8(d)
	return r.tokenDec, nil
}

// toBaseUnits converts a whole-unit decimal amount into chain base units.
func toBaseUnits(amount decimal.Decimal, decimals int32) *big.Int {
	return amount.Shift(decimals).BigInt()
}

// fromBaseUnits converts chain base units into a whole-unit decimal.
func fromBaseUnits(base *big.Int, decimals int32) decimal.Decimal {
	return decimal.NewFromBigInt(base, -decimals)
}

func (r *AMMSwapRoute) Quote(ctx context.Context, from, to string, amount decimal.Decimal) (*RouteQuote, error) {
	if err := r.configErr(); err != nil {
		return nil, err
	}
	dir, err := ammDirectionFor(from, to)
	if err != nil {
		return nil, err
	}
	if amount.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("amm swap: amount must be positive")
	}

	res0, res1, err := r.reserves(ctx)
	if err != nil {
		return nil, err
	}
	if res0.Sign() <= 0 || res1.Sign() <= 0 {
		return nil, fmt.Errorf("amm swap: pool has no liquidity (reserves %s TXDC / %s token)", res0, res1)
	}

	dec, err := r.decimals(ctx)
	if err != nil && r.tokenDec == 0 {
		return nil, err
	}

	var (
		amountIn, resIn, resOut *big.Int
		outDec                  int32
	)
	if dir == dirTXDCToToken {
		amountIn, resIn, resOut = toBaseUnits(amount, 18), res0, res1
		outDec = int32(dec)
	} else {
		amountIn, resIn, resOut = toBaseUnits(amount, int32(dec)), res1, res0
		outDec = 18
	}

	out := AMMAmountOut(amountIn, resIn, resOut)
	if out.Sign() <= 0 {
		return nil, fmt.Errorf("amm swap: output rounds to zero for %s %s", amount.String(), from)
	}
	destAmount := fromBaseUnits(out, outDec)
	rate := destAmount.Div(amount)

	// 0.3% AMM fee, expressed in the source asset. Gas (~0.001 TXDC) is
	// negligible on XDPoS and already covered by the fixed-fee accounting in
	// the other routes; the fee here is the dominant, real cost.
	fee := amount.Mul(decimal.NewFromInt(3)).Div(decimal.NewFromInt(1000))

	log.Debug().
		Str("route", string(RouteAMMSwap)).
		Str("from", from).Str("to", to).
		Str("reserve_txdc", fromBaseUnits(res0, 18).String()).
		Str("dest_amount", destAmount.String()).
		Msg("amm swap quote generated")

	return &RouteQuote{
		RouteID:        r.ID(),
		RouteName:      r.Name(),
		SourceAsset:    from,
		DestAsset:      to,
		SourceAmount:   amount,
		DestAmount:     destAmount,
		Rate:           rate,
		SpreadBps:      0,
		Fee:            fee,
		FeeAsset:       from,
		SettlementTime: 10 * time.Second, // ~5 XDPoS blocks at 2s
		ExpiresAt:      time.Now().Add(ammQuoteTTL),
		Provider:       "flowx_amm",
	}, nil
}

// Execute submits a real swap transaction signed by the treasury key. For
// USDC->TXDC it first submits approve() on the token and waits for that
// receipt before the swap (so the pool's transferFrom cannot race).
// Returns the real on-chain tx hash, which doubles as the Status reference.
func (r *AMMSwapRoute) Execute(ctx context.Context, req PaymentRequest, quote *RouteQuote) (string, error) {
	if err := r.configErr(); err != nil {
		return "", err
	}
	if r.treasuryKey == "" {
		return "", fmt.Errorf("amm swap: XDC_TREASURY_SECRET_KEY is not configured")
	}
	dir, err := ammDirectionFor(req.SourceAsset, req.DestAsset)
	if err != nil {
		return "", err
	}
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return "", fmt.Errorf("amm swap: amount must be positive")
	}

	ec, err := r.ensureClient(ctx)
	if err != nil {
		return "", err
	}
	sk, err := crypto.ToECDSA(common.FromHex(r.treasuryKey))
	if err != nil {
		return "", fmt.Errorf("amm swap: parse treasury key: %w", err)
	}
	fromAddr := crypto.PubkeyToAddress(sk.PublicKey)

	// Beneficiary: caller-provided destination wins; otherwise the swap pays
	// out to the treasury wallet itself (self-swap).
	toAddr := fromAddr
	if req.DestinationAddress != "" {
		if !chainAddressPattern.MatchString(req.DestinationAddress) {
			return "", fmt.Errorf("amm swap: invalid destination address %q", req.DestinationAddress)
		}
		toAddr = common.HexToAddress(ammNormalize(req.DestinationAddress))
	}

	dec, derr := r.decimals(ctx)
	if derr != nil && r.tokenDec == 0 {
		return "", derr
	}
	var amountIn *big.Int
	if dir == dirTXDCToToken {
		amountIn = toBaseUnits(req.Amount, 18)
	} else {
		amountIn = toBaseUnits(req.Amount, int32(dec))
	}

	// Slippage floor: prefer the quoted destination amount (what the user
	// was shown); fall back to a live recompute when the quote is missing.
	var minOut *big.Int
	if quote != nil && quote.DestAmount.GreaterThan(decimal.Zero) {
		outDec := int32(dec)
		if dir == dirTokenToTXDC {
			outDec = 18
		}
		minOut = minOutWithSlippage(toBaseUnits(quote.DestAmount, outDec), r.slippageBps)
	} else {
		res0, res1, rerr := r.reserves(ctx)
		if rerr != nil {
			return "", rerr
		}
		var resIn, resOut *big.Int
		if dir == dirTXDCToToken {
			resIn, resOut = res0, res1
		} else {
			resIn, resOut = res1, res0
		}
		minOut = minOutWithSlippage(AMMAmountOut(amountIn, resIn, resOut), r.slippageBps)
	}

	pool, err := amm.NewFlowXPool(common.HexToAddress(ammNormalize(r.poolAddr)), ec)
	if err != nil {
		return "", fmt.Errorf("amm swap: bind pool: %w", err)
	}
	gp, err := ec.SuggestGasPrice(ctx)
	if err != nil {
		return "", fmt.Errorf("amm swap: gas price: %w", err)
	}
	nonce, err := ec.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return "", fmt.Errorf("amm swap: nonce: %w", err)
	}
	deadline := big.NewInt(time.Now().Add(ammSwapDeadline).Unix())

	auth, err := bind.NewKeyedTransactorWithChainID(sk, r.chainID)
	if err != nil {
		return "", fmt.Errorf("amm swap: transactor: %w", err)
	}
	auth.Context = ctx
	auth.GasPrice = gp
	auth.GasLimit = ammGasLimitSwap

	var tx *types.Transaction
	if dir == dirTXDCToToken {
		auth.Nonce = new(big.Int).SetUint64(nonce)
		auth.Value = amountIn
		tx, err = pool.SwapExactTXDCForTokens(auth, minOut, toAddr, deadline)
	} else {
		// Approve the pool to pull the token input, wait for the receipt,
		// then swap with the following nonce.
		approveHash, aerr := amm.Approve(ctx, ec, r.signer, r.treasuryKey, r.tokenAddr, r.poolAddr, amountIn, nonce)
		if aerr != nil {
			return "", fmt.Errorf("amm swap: approve token: %w", aerr)
		}
		log.Info().Str("tx_hash", approveHash).Msg("amm swap: approve submitted, waiting for confirmation")
		if werr := waitReceipt(ctx, ec, approveHash, ammReceiptTimeout); werr != nil {
			return "", fmt.Errorf("amm swap: approve not confirmed: %w", werr)
		}
		auth.Nonce = new(big.Int).SetUint64(nonce + 1)
		tx, err = pool.SwapExactTokensForTXDC(auth, amountIn, minOut, toAddr, deadline)
	}
	if err != nil {
		return "", fmt.Errorf("amm swap: submit: %w", err)
	}

	hash := tx.Hash().Hex()
	log.Info().
		Str("tx_hash", hash).
		Str("explorer", xdc.ExplorerTxURL(hash)).
		Str("direction", fmt.Sprint(dir)).
		Str("to", toAddr.Hex()).
		Msg("amm swap: real swap submitted on Apothem")
	return hash, nil
}

// Status polls the chain for the swap tx receipt and reports the ACTUAL
// state: pending (no receipt yet), confirmed (status 1), or failed (status 0).
// This replaces the constants-smell of the other routes' Status methods.
func (r *AMMSwapRoute) Status(ctx context.Context, reference string) (string, error) {
	if err := r.configErr(); err != nil {
		return "", err
	}
	if !common.IsHexAddress(strings.TrimPrefix(reference, "0x")) && !common.IsHexAddress(reference) {
		return "", fmt.Errorf("amm swap: reference %q is not a tx hash", reference)
	}
	ec, err := r.ensureClient(ctx)
	if err != nil {
		return "", err
	}
	receipt, err := ec.TransactionReceipt(ctx, common.HexToHash(reference))
	if errors.Is(err, ethereum.NotFound) {
		return "pending", nil
	}
	if err != nil {
		return "", fmt.Errorf("amm swap: receipt %s: %w", reference, err)
	}
	if receipt.Status == types.ReceiptStatusSuccessful {
		return "confirmed", nil
	}
	return "failed", nil
}

// waitReceipt polls until txHash is mined, reverting is reported as an error.
func waitReceipt(ctx context.Context, ec *ethclient.Client, txHash string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	hash := common.HexToHash(txHash)
	for {
		receipt, err := ec.TransactionReceipt(ctx, hash)
		if err == nil {
			if receipt.Status != types.ReceiptStatusSuccessful {
				return fmt.Errorf("tx %s reverted", txHash)
			}
			return nil
		}
		if !errors.Is(err, ethereum.NotFound) {
			return fmt.Errorf("receipt %s: %w", txHash, err)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("tx %s not mined within %s", txHash, timeout)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(ammReceiptPoll):
		}
	}
}

// ammNormalize converts an xdc-prefixed address into the 0x form go-ethereum
// expects (kept local to avoid coupling to the token binding tree).
func ammNormalize(addr string) string {
	if len(addr) >= 3 && strings.EqualFold(addr[:3], "xdc") {
		return "0x" + addr[3:]
	}
	return addr
}

// sanity: AMMSwapRoute satisfies PaymentRoute at compile time.
var _ PaymentRoute = (*AMMSwapRoute)(nil)
