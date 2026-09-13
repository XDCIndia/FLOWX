// Command tokendeploy deploys the TestUSDC (tUSDC) ERC-20 contract to XDC
// Apothem testnet from the treasury key and prints the contract address.
//
// It reads .env from the repository root (walking up from the working
// directory until a go.mod is found), using:
//
//	XDC_RPC_URL               (default https://rpc.apothem.network)
//	XDC_CHAIN_ID              (default 51)
//	XDC_TREASURY_SECRET_KEY   (hex private key of the deployer)
//	TUSDC_INITIAL_SUPPLY      (optional, whole units; default 1000000)
//
// Usage: go run ./cmd/tokendeploy
package main

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/fluxa/fluxa/internal/chain/xdc/tokens"
)

// findRepoRoot walks up from dir until it finds a directory containing
// go.mod, and returns that directory.
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found in any parent of the working directory")
		}
		dir = parent
	}
}

// loadDotEnv reads KEY=VALUE pairs from path into env (without overriding
// variables already set in the process environment).
func loadDotEnv(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		// Strip surrounding quotes if present.
		if len(value) >= 2 && (value[0] == '"' && value[len(value)-1] == '"' || value[0] == '\'' && value[len(value)-1] == '\'') {
			value = value[1 : len(value)-1]
		}
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	root, err := findRepoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, "tokendeploy:", err)
		os.Exit(1)
	}
	// Load .env from the repo root; a missing file is not fatal (env vars
	// may be provided by the shell instead).
	if err := loadDotEnv(filepath.Join(root, ".env")); err != nil && !os.IsNotExist(err) {
		fmt.Fprintln(os.Stderr, "tokendeploy: reading .env:", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	rpcURL := envOr("XDC_RPC_URL", "https://rpc.apothem.network")
	chainID := new(big.Int)
	chainID.SetString(envOr("XDC_CHAIN_ID", "51"), 10)

	secret := os.Getenv("XDC_TREASURY_SECRET_KEY")
	if secret == "" {
		fmt.Fprintln(os.Stderr, "tokendeploy: XDC_TREASURY_SECRET_KEY is not set")
		os.Exit(1)
	}
	key, err := crypto.HexToECDSA(strings.TrimPrefix(strings.TrimSpace(secret), "0x"))
	if err != nil {
		fmt.Fprintln(os.Stderr, "tokendeploy: parsing treasury key:", err)
		os.Exit(1)
	}
	deployer := crypto.PubkeyToAddress(key.PublicKey)

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tokendeploy: dial:", err)
		os.Exit(1)
	}
	remoteID, err := client.ChainID(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tokendeploy: chainId:", err)
		os.Exit(1)
	}
	if remoteID.Cmp(chainID) != 0 {
		fmt.Fprintf(os.Stderr, "tokendeploy: chainId mismatch: RPC says %s, want %s — refusing to deploy\n", remoteID, chainID)
		os.Exit(1)
	}

	balance, err := client.BalanceAt(ctx, deployer, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tokendeploy: balance:", err)
		os.Exit(1)
	}
	fmt.Printf("deployer: %s\n", deployer.Hex())
	fmt.Printf("chainId:  %s (%s)\n", chainID, rpcURL)
	fmt.Printf("balance:  %s wei\n", balance)

	// Initial supply: whole units of tUSDC, converted to 6-decimal base
	// units. Default 1,000,000 tUSDC.
	supplyWhole := envOr("TUSDC_INITIAL_SUPPLY", "1000000")
	supply, ok := new(big.Int).SetString(supplyWhole, 10)
	if !ok {
		fmt.Fprintln(os.Stderr, "tokendeploy: invalid TUSDC_INITIAL_SUPPLY:", supplyWhole)
		os.Exit(1)
	}
	supply.Mul(supply, new(big.Int).Exp(big.NewInt(10), big.NewInt(6), nil)) // 6 decimals
	fmt.Printf("supply:   %s base units (6 decimals)\n", supply)

	nonce, err := client.PendingNonceAt(ctx, deployer)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tokendeploy: nonce:", err)
		os.Exit(1)
	}
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tokendeploy: gasPrice:", err)
		os.Exit(1)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tokendeploy: transactor:", err)
		os.Exit(1)
	}
	auth.Nonce = new(big.Int).SetUint64(nonce)
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(1_500_000)
	auth.GasPrice = gasPrice

	address, tx, _, err := tokens.DeployTestUSDC(auth, client, supply)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tokendeploy: deploy:", err)
		os.Exit(1)
	}
	fmt.Printf("tx:       %s\n", tx.Hash().Hex())
	fmt.Printf("contract: %s\n", address.Hex())

	fmt.Println("waiting for confirmation...")
	receipt, err := bind.WaitMined(ctx, client, tx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "tokendeploy: wait mined:", err)
		os.Exit(1)
	}
	if receipt.Status != 1 {
		fmt.Fprintln(os.Stderr, "tokendeploy: deployment tx failed (status 0)")
		os.Exit(1)
	}
	fmt.Printf("confirmed in block %d\n", receipt.BlockNumber)
	fmt.Printf("xdc address: xdc%s\n", strings.ToLower(address.Hex()[2:]))
}
