// swapsmoke performs a real end-to-end swap through the amm_swap route on
// XDC Apothem: Quote -> Execute -> Status, printing on-chain evidence.
// Dev tool; not part of the server wiring.
package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"

	"github.com/fluxa/fluxa/internal/chain/xdc/tokens"
	"github.com/fluxa/fluxa/internal/routing"
)

func main() {
	rpc := os.Getenv("XDC_RPC_URL")
	pool := os.Getenv("AMM_POOL_ADDRESS")
	token := os.Getenv("XDC_USDC_CONTRACT_ADDRESS")
	key := os.Getenv("XDC_TREASURY_SECRET_KEY")
	if rpc == "" || pool == "" || token == "" || key == "" {
		fmt.Println("need XDC_RPC_URL, AMM_POOL_ADDRESS, XDC_USDC_CONTRACT_ADDRESS, XDC_TREASURY_SECRET_KEY")
		os.Exit(1)
	}
	pk, err := crypto.ToECDSA(common.Hex2Bytes(key))
	if err != nil {
		fmt.Println("bad key:", err)
		os.Exit(1)
	}
	self := crypto.PubkeyToAddress(pk.PublicKey)

	route := routing.NewAMMSwapRoute(rpc, pool, token, key)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	if !route.Supports("TXDC", "USDC", "", "") {
		fmt.Println("route does not support TXDC->USDC")
		os.Exit(1)
	}

	amount := decimal.RequireFromString("0.5")
	q, err := route.Quote(ctx, "TXDC", "USDC", amount)
	if err != nil {
		fmt.Println("quote err:", err)
		os.Exit(1)
	}
	fmt.Printf("QUOTE: %s TXDC -> %s USDC (rate %s, fee %s %s, provider %s)\n",
		q.SourceAmount, q.DestAmount, q.Rate, q.Fee, q.FeeAsset, q.Provider)

	ref, err := route.Execute(ctx, routing.PaymentRequest{
		SourceAsset: "TXDC", DestAsset: "USDC", Amount: amount,
		DestinationAddress: self.Hex(),
	}, q)
	if err != nil {
		fmt.Println("execute err:", err)
		os.Exit(1)
	}
	fmt.Println("EXECUTE tx:", ref)

	deadline := time.Now().Add(90 * time.Second)
	for {
		st, err := route.Status(ctx, ref)
		if err != nil {
			fmt.Println("status err:", err)
			break
		}
		fmt.Println("STATUS:", st)
		if st == "confirmed" || st == "failed" {
			break
		}
		if time.Now().After(deadline) {
			fmt.Println("STATUS: timeout waiting for confirmation")
			break
		}
		time.Sleep(3 * time.Second)
	}

	// On-chain proof: token balance of treasury after the swap.
	cl, err := ethclient.Dial(rpc)
	if err != nil {
		fmt.Println("client err:", err)
		os.Exit(1)
	}
	caller, err := tokens.NewTestUSDCCaller(common.HexToAddress(strings.TrimPrefix(token, "xdc")), cl)
	if err != nil {
		fmt.Println("caller err:", err)
		os.Exit(1)
	}
	bal, err := caller.BalanceOf(nil, self)
	if err != nil {
		fmt.Println("balance check err:", err)
		os.Exit(1)
	}
	fmt.Println("TREASURY tUSDC BALANCE (raw 6dp):", bal.String())
}
