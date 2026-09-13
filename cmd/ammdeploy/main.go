// Command ammdeploy deploys the FlowXPool constant-product AMM contract to
// XDC Apothem and optionally seeds it with initial liquidity.
//
// Usage:
//
//	go run ./cmd/ammdeploy -token <tUSDC-address> [-txdc 10] [-token-amount 35000]
//
// Env fallback: XDC_RPC_URL, XDC_TREASURY_SECRET_KEY, XDC_CHAIN_ID,
// XDC_USDC_CONTRACT_ADDRESS. The treasury key is the pool operator (lp):
// it deploys the pool, approves the token, and supplies initial liquidity.
//
// This tool only prints the deployed pool address — it does NOT write any
// config. Set AMM_POOL_ADDRESS (and XDC_USDC_CONTRACT_ADDRESS) in .env after
// the token worktree has merged and the pool is seeded.
package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"

	"github.com/fluxa/fluxa/internal/chain/xdc"
	"github.com/fluxa/fluxa/internal/routing/amm"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "ammdeploy: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		rpcURL      = flag.String("rpc", envOr("XDC_RPC_URL", xdc.DefaultRPC), "XDC RPC URL")
		chainID     = flag.Int64("chain-id", int64EnvOr("XDC_CHAIN_ID", xdc.ApothemChainID), "expected chain ID")
		key         = flag.String("key", os.Getenv("XDC_TREASURY_SECRET_KEY"), "operator private key (hex)")
		tokenAddr   = flag.String("token", os.Getenv("XDC_USDC_CONTRACT_ADDRESS"), "tUSDC ERC-20 address (required unless seeding later)")
		txdcAmount  = flag.String("txdc", "0", "TXDC liquidity to add (whole units; 0 = deploy only)")
		tokenAmount = flag.String("token-amount", "0", "token liquidity to add (whole units; 0 = deploy only)")
		tokenDec    = flag.Int("token-decimals", 6, "token decimals (for whole-unit seed amounts)")
		waitTimeout = flag.Duration("wait", 90*time.Second, "receipt wait timeout")
	)
	flag.Parse()

	if *key == "" {
		return fmt.Errorf("treasury key required (-key or XDC_TREASURY_SECRET_KEY)")
	}
	if *tokenAddr == "" {
		return fmt.Errorf("token address required (-token or XDC_USDC_CONTRACT_ADDRESS)")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	ec, err := ethclient.Dial(*rpcURL)
	if err != nil {
		return fmt.Errorf("dial %s: %w", *rpcURL, err)
	}
	id, err := ec.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("chainId: %w", err)
	}
	if id.Int64() != *chainID {
		return fmt.Errorf("chainId mismatch: RPC says %d, want %d — refusing to deploy", id.Int64(), *chainID)
	}

	sk, err := crypto.ToECDSA(common.FromHex(*key))
	if err != nil {
		return fmt.Errorf("parse key: %w", err)
	}
	op := crypto.PubkeyToAddress(sk.PublicKey)
	fmt.Printf("operator (lp): %s\n", op.Hex())

	bal, err := ec.BalanceAt(ctx, op, nil)
	if err != nil {
		return fmt.Errorf("operator balance: %w", err)
	}
	fmt.Printf("operator TXDC balance: %s\n", decimal.NewFromBigInt(bal, -18).String())

	auth, err := bind.NewKeyedTransactorWithChainID(sk, id)
	if err != nil {
		return fmt.Errorf("transactor: %w", err)
	}
	auth.Context = ctx
	auth.GasLimit = 2000000

	fmt.Println("deploying FlowXPool…")
	addr, tx, pool, err := amm.DeployFlowXPool(auth, ec, common.HexToAddress(normalize(*tokenAddr)), op)
	if err != nil {
		return fmt.Errorf("deploy: %w", err)
	}
	fmt.Printf("deploy tx: %s\n", tx.Hash().Hex())
	if err := waitReceipt(ctx, ec, tx.Hash(), *waitTimeout); err != nil {
		return fmt.Errorf("deploy receipt: %w", err)
	}
	fmt.Printf("FlowXPool deployed: %s\n", addr.Hex())
	fmt.Printf("explorer: %s\n", xdc.ExplorerAddressURL(addr.Hex()))

	txdcLiquidity := decimal.RequireFromString(*txdcAmount)
	tokenLiquidity := decimal.RequireFromString(*tokenAmount)
	if txdcLiquidity.LessThanOrEqual(decimal.Zero) || tokenLiquidity.LessThanOrEqual(decimal.Zero) {
		fmt.Println("no liquidity requested (-txdc / -token-amount) — done.")
		fmt.Printf("\nSet AMM_POOL_ADDRESS=%s (xdc-prefixed also accepted)\n", addr.Hex())
		return nil
	}

	// Approve the pool to pull the token side, wait, then addLiquidity.
	tokenRaw := tokenLiquidity.Shift(int32(*tokenDec)).BigInt()
	gp, err := ec.SuggestGasPrice(ctx)
	if err != nil {
		return fmt.Errorf("gas price: %w", err)
	}
	nonce, err := ec.PendingNonceAt(ctx, op)
	if err != nil {
		return fmt.Errorf("nonce: %w", err)
	}
	approveTx, err := amm.Approve(ctx, ec, types.NewEIP155Signer(id), *key, *tokenAddr, addr.Hex(), tokenRaw, nonce)
	if err != nil {
		return fmt.Errorf("approve: %w", err)
	}
	fmt.Printf("approve tx: %s\n", approveTx)
	if err := waitReceipt(ctx, ec, common.HexToHash(approveTx), *waitTimeout); err != nil {
		return fmt.Errorf("approve receipt: %w", err)
	}

	auth.Nonce = new(big.Int).SetUint64(nonce + 1)
	auth.Value = txdcLiquidity.Shift(18).BigInt()
	auth.GasLimit = 300000
	auth.GasPrice = gp
	addTx, err := pool.AddLiquidity(auth, tokenRaw)
	if err != nil {
		return fmt.Errorf("addLiquidity: %w", err)
	}
	fmt.Printf("addLiquidity tx: %s\n", addTx.Hash().Hex())
	if err := waitReceipt(ctx, ec, addTx.Hash(), *waitTimeout); err != nil {
		return fmt.Errorf("addLiquidity receipt: %w", err)
	}

	res0, res1, err := pool.Reserves(&bind.CallOpts{Context: ctx})
	if err != nil {
		return fmt.Errorf("reserves: %w", err)
	}
	fmt.Printf("reserves: %s TXDC / %s token (kLast=%s)\n",
		decimal.NewFromBigInt(res0, -18).String(), res1, new(big.Int).Mul(res0, res1))

	fmt.Printf("\nSet AMM_POOL_ADDRESS=%s (xdc-prefixed also accepted)\n", addr.Hex())
	return nil
}

func waitReceipt(ctx context.Context, ec *ethclient.Client, hash common.Hash, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		receipt, err := ec.TransactionReceipt(ctx, hash)
		if err == nil {
			if receipt.Status != types.ReceiptStatusSuccessful {
				return fmt.Errorf("tx %s reverted", hash.Hex())
			}
			fmt.Printf("confirmed in block %d\n", receipt.BlockNumber)
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("tx %s not mined within %s", hash.Hex(), timeout)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func int64EnvOr(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		var n int64
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			return n
		}
	}
	return def
}

func normalize(addr string) string {
	if len(addr) >= 3 && (addr[:3] == "xdc" || addr[:3] == "XDC") {
		return "0x" + addr[3:]
	}
	return addr
}
