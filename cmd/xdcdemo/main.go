// xdcdemo is the XDC Apothem working model: a cross-border-style
// wallet-to-wallet transfer using Fluxa's custody pattern (encrypted
// secrets via internal/crypto) and the chain abstraction from
// internal/chain + internal/chain/xdc.
//
// Usage:
//
//	go run ./cmd/xdcdemo
//
// Flow:
//  1. generate sender + receiver wallets (secp256k1)
//  2. encrypt both private keys with MASTER_ENCRYPTION_KEY (Fluxa custody model)
//  3. wait for faucet funding (one manual captcha step at faucet.apothem.network)
//  4. decrypt sender key in memory, transfer native XDC
//  5. wait 6 confirmations, print on-chain proof links
package main

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/spf13/viper"

	"github.com/fluxa/fluxa/internal/chain"
	"github.com/fluxa/fluxa/internal/chain/xdc"
	"github.com/fluxa/fluxa/internal/crypto"
)

const (
	faucetURL         = "https://faucet.apothem.network/"
	fundingPoll       = 10 * time.Second
	defaultConfirmations = 6 // migration plan §6 testnet default (~12s)
)

var (
	rpcURL        = flag.String("rpc", xdc.DefaultRPC, "XDC Apothem RPC endpoint")
	amountTXDC     = flag.Float64("amount", 1.0, "TXDC to send sender -> receiver")
	waitMinutes   = flag.Int("wait", 15, "minutes to wait for manual faucet funding")
	confirmations = flag.Uint64("confirmations", defaultConfirmations, "required confirmations")
	toAddr        = flag.String("to", "", "optional existing 0x receiver address (default: generate one)")
)

func fatal(err error) { fmt.Fprintln(os.Stderr, "FATAL:", err); os.Exit(1) }

func main() {
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*waitMinutes+10)*time.Minute)
	defer cancel()

	// Config: reuse Fluxa's .env conventions (viper) so the demo runs with
	// the same environment as the API server.
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	_ = viper.ReadInConfig()
	masterKeyHex := viper.GetString("MASTER_ENCRYPTION_KEY")
	if masterKeyHex == "" {
		masterKeyHex = os.Getenv("MASTER_ENCRYPTION_KEY")
	}
	if masterKeyHex == "" {
		fatal(fmt.Errorf("MASTER_ENCRYPTION_KEY not set (put it in .env or export it)"))
	}
	masterKey, err := hex.DecodeString(masterKeyHex)
	if err != nil {
		fatal(fmt.Errorf("MASTER_ENCRYPTION_KEY must be hex: %w", err))
	}

	fmt.Println("=== Fluxa XDC Apothem — wallet-to-wallet transfer model ===")
	client, err := xdc.New(ctx, *rpcURL, xdc.ApothemChainID)
	if err != nil {
		fatal(err)
	}
	fmt.Println("chain OK | chainId:", xdc.ApothemChainID, "| rpc:", *rpcURL)

	// [1] wallets — Fluxa custody model: private keys never stored raw,
	// only AES-encrypted with the master key (same as wallet/service.go).
	fmt.Println("\n[1] wallets (Fluxa custody: secrets AES-encrypted)")
	senderAddr, senderPK, err := client.GenerateKeypair()
	if err != nil {
		fatal(err)
	}
	receiverAddr := *toAddr
	if receiverAddr == "" {
		receiverAddr, _, err = client.GenerateKeypair()
		if err != nil {
			fatal(err)
		}
	}
	senderEnc, err := crypto.Encrypt([]byte(senderPK), masterKey)
	if err != nil {
		fatal(err)
	}
	fmt.Println("  sender   :", senderAddr, "|", xdc.ExplorerAddressURL(senderAddr))
	fmt.Println("  receiver :", receiverAddr, "|", xdc.ExplorerAddressURL(receiverAddr))
	fmt.Println("  sender secret encrypted:", len(senderEnc), "bytes (AES-256-GCM, master key) — raw key discarded")

	// [2] funding — Apothem faucet has a captcha: one manual step.
	fmt.Println("\n[2] funding")
	bal, err := client.Balance(ctx, senderAddr, chain.NativeTXDC)
	if err != nil {
		fatal(err)
	}
	fmt.Println("  sender balance:", weiToTXDC(bal), "TXDC")
	if bal.Sign() == 0 {
		fmt.Println("  ACTION NEEDED:")
		fmt.Println("    1. open:", faucetURL)
		fmt.Println("    2. paste address:", senderAddr)
		fmt.Println("    3. complete the captcha — waiting (up to", *waitMinutes, "min)...")
		if err := waitFunded(ctx, client, senderAddr); err != nil {
			fatal(err)
		}
	}

	// [3] transfer — decrypt sender secret in memory only, sign, submit.
	fmt.Println("\n[3] transfer")
	senderPKBytes, err := crypto.Decrypt(senderEnc, masterKey)
	if err != nil {
		fatal(err)
	}
	amountWei, _ := new(big.Float).Mul(big.NewFloat(*amountTXDC), big.NewFloat(1e18)).Int(nil)
	txHash, err := client.Transfer(ctx, string(senderPKBytes), receiverAddr, chain.NativeTXDC, amountWei)
	if err != nil {
		fatal(err)
	}
	fmt.Println("  submitted:", txHash)
	fmt.Println("  explorer :", xdc.ExplorerTxURL(txHash))

	// [4] confirmations + proof.
	fmt.Printf("\n[4] waiting for %d confirmations...\n", *confirmations)
	if err := client.WaitConfirmations(ctx, txHash, *confirmations); err != nil {
		fatal(err)
	}
	sBal, _ := client.Balance(ctx, senderAddr, chain.NativeTXDC)
	rBal, _ := client.Balance(ctx, receiverAddr, chain.NativeTXDC)
	fmt.Println("\n=== SUCCESS — cross-border transfer settled on XDC testnet ===")
	fmt.Println("  tx hash      :", txHash)
	fmt.Println("  proof        :", xdc.ExplorerTxURL(txHash))
	fmt.Println("  sender bal   :", weiToTXDC(sBal), "TXDC")
	fmt.Println("  receiver bal :", weiToTXDC(rBal), "TXDC")
}

func waitFunded(ctx context.Context, cli chain.ChainClient, addr string) error {
	for {
		b, err := cli.Balance(ctx, addr, chain.NativeTXDC)
		if err == nil && b.Sign() > 0 {
			fmt.Println("  funded! balance:", weiToTXDC(b), "TXDC")
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for faucet funding")
		case <-time.After(fundingPoll):
		}
	}
}

func weiToTXDC(v *big.Int) string {
	f := new(big.Float).Quo(new(big.Float).SetInt(v), big.NewFloat(1e18))
	return f.Text('f', 6)
}
