package main

import (
	"bufio"
	"context"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	f, err := os.Open("/home/dhiraj/Fluxa/.env")
	if err != nil {
		fmt.Println("ENV_ERR:", err)
		return
	}
	defer f.Close()
	var key, rpc string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		l := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(l, "XDC_TREASURY_SECRET_KEY=") {
			key = strings.Trim(strings.TrimPrefix(l, "XDC_TREASURY_SECRET_KEY="), `"'`)
		}
		if strings.HasPrefix(l, "XDC_RPC_URL=") {
			rpc = strings.Trim(strings.TrimPrefix(l, "XDC_RPC_URL="), `"'`)
		}
	}
	if key == "" {
		fmt.Println("NO_TREASURY_KEY")
		return
	}
	if rpc == "" {
		rpc = "https://rpc.apothem.network"
	}
	pk, err := crypto.ToECDSA(common.Hex2Bytes(strings.TrimPrefix(key, "0x")))
	if err != nil {
		fmt.Println("BAD_KEY:", err)
		return
	}
	addr := crypto.PubkeyToAddress(pk.PublicKey)
	fmt.Println("ADDRESS:", "xdc"+strings.TrimPrefix(addr.Hex(), "0x"))
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cl, err := ethclient.DialContext(ctx, rpc)
	if err != nil {
		fmt.Println("RPC_ERR:", err)
		return
	}
	bal, err := cl.BalanceAt(ctx, addr, nil)
	if err != nil {
		fmt.Println("BAL_ERR:", err)
		return
	}
	fmt.Printf("TXDC_BALANCE: %s\n", new(big.Float).Quo(new(big.Float).SetInt(bal), big.NewFloat(1e18)).Text('f', 4))
}
