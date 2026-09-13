package config

import "testing"

// TestLoadAMMPoolAddress verifies the AMM pool config key round-trips.
// Kept in a separate file from config_test.go (owned by the token worktree)
// to minimize merge conflicts.
func TestLoadAMMPoolAddress(t *testing.T) {
	setRequiredEnv(t)
	// Isolate from ambient env / repo .env: the default must be empty.
	t.Setenv("AMM_POOL_ADDRESS", "")

	// Default: unset must stay empty (route disabled).
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AMMPoolAddress != "" {
		t.Errorf("AMMPoolAddress = %q, want empty by default", cfg.AMMPoolAddress)
	}

	t.Setenv("AMM_POOL_ADDRESS", "xdc967f5f9d1a1d8b5c8d1f0a2b3c4d5e6f7890abcd")
	cfg, err = Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.AMMPoolAddress != "xdc967f5f9d1a1d8b5c8d1f0a2b3c4d5e6f7890abcd" {
		t.Errorf("AMMPoolAddress = %q, want value from AMM_POOL_ADDRESS", cfg.AMMPoolAddress)
	}
}
