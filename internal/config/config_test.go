package config

import (
	"os"
	"strings"
	"testing"
)

// setRequiredEnv provides the env keys config.Load refuses to boot without.
// STELLAR_* keys are deliberately NOT set: with the xdc default the API must
// boot without any Stellar configuration.
func setRequiredEnv(t *testing.T) {
	t.Helper()
	keys := map[string]string{
		"DATABASE_URL":          "postgres://localhost/fluxa_test",
		"REDIS_URL":             "redis://localhost:6379/0",
		"MASTER_ENCRYPTION_KEY": strings.Repeat("ab", 32),
	}
	for k, v := range keys {
		k, v := k, v
		t.Setenv(k, v)
	}
}

func TestLoadDefaultsToXDCBackend(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ChainBackend != "xdc" {
		t.Errorf("ChainBackend = %q, want %q", cfg.ChainBackend, "xdc")
	}
	// XDC defaults must be in place without any env config.
	if cfg.XDCChainID != 51 {
		t.Errorf("XDCChainID = %d, want 51", cfg.XDCChainID)
	}
	if cfg.XDCRPCURL == "" {
		t.Error("XDCRPCURL default must be set")
	}
}

func TestLoadFaucetDefaultsDisabled(t *testing.T) {
	setRequiredEnv(t)
	os.Unsetenv("TESTNET_FAUCET_ENABLED")
	os.Unsetenv("FAUCET_MAX_AMOUNT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.TestnetFaucetEnabled {
		t.Error("TestnetFaucetEnabled must default to false")
	}
	if cfg.FaucetMaxAmount != 1000 {
		t.Errorf("FaucetMaxAmount = %v, want 1000", cfg.FaucetMaxAmount)
	}
}

func TestLoadFaucetOverrides(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("TESTNET_FAUCET_ENABLED", "true")
	t.Setenv("FAUCET_MAX_AMOUNT", "5000.5")
	t.Setenv("XDC_USDC_CONTRACT_ADDRESS", "xdc8f3cf7ad23cd3cadbd9735aff958023239c6a063")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !cfg.TestnetFaucetEnabled {
		t.Error("TestnetFaucetEnabled must honor TESTNET_FAUCET_ENABLED=true")
	}
	if cfg.FaucetMaxAmount != 5000.5 {
		t.Errorf("FaucetMaxAmount = %v, want 5000.5", cfg.FaucetMaxAmount)
	}
	if cfg.XDCUSDCContractAddress == "" {
		t.Error("XDCUSDCContractAddress must be read from XDC_USDC_CONTRACT_ADDRESS")
	}
}

func TestLoadBootsWithoutStellarEnv(t *testing.T) {
	setRequiredEnv(t)
	// Belt-and-braces: none of the Stellar keys may be set for this test.
	for _, k := range []string{"STELLAR_NETWORK", "STELLAR_HORIZON_URL", "STELLAR_USDC_ISSUER", "STELLAR_EURC_ISSUER"} {
		os.Unsetenv(k)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load must not require STELLAR_* env keys with xdc backend: %v", err)
	}
	if cfg.ChainBackend != "xdc" {
		t.Errorf("ChainBackend = %q, want xdc", cfg.ChainBackend)
	}
}
