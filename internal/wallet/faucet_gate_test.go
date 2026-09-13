package wallet_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/shopspring/decimal"

	"github.com/fluxa/fluxa/internal/domain"
	"github.com/fluxa/fluxa/internal/wallet"
)

// faucetMockSvc implements wallet.Service well enough for the faucet route;
// Faucet records calls and returns a canned result.
type faucetMockSvc struct {
	calls []faucetCall
}

type faucetCall struct {
	walletID  string
	assetCode string
	amount    decimal.Decimal
}

func (m *faucetMockSvc) Faucet(_ context.Context, walletID, assetCode string, amount decimal.Decimal) (*wallet.FaucetResult, error) {
	m.calls = append(m.calls, faucetCall{walletID, assetCode, amount})
	return &wallet.FaucetResult{Balance: amount.String()}, nil
}

func (m *faucetMockSvc) CreateWallet(_ context.Context, _ ...string) (*domain.Wallet, error) {
	return nil, nil
}
func (m *faucetMockSvc) GetWalletForHandler(_ context.Context, id string) (*domain.Wallet, error) {
	return &domain.Wallet{ID: id}, nil
}
func (m *faucetMockSvc) GetBalances(_ context.Context, _ string, _ ...string) ([]wallet.Balance, error) {
	return nil, nil
}
func (m *faucetMockSvc) AddTrustline(_ context.Context, _, _, _, _ string) (string, error) {
	return "", nil
}
func (m *faucetMockSvc) ExecuteTransfer(_ context.Context, _, _, _, _ string, _ decimal.Decimal, _ string) (string, error) {
	return "", nil
}
func (m *faucetMockSvc) VerifyDeposit(_ context.Context, _, _ string) (*domain.Transaction, error) {
	return nil, nil
}
func (m *faucetMockSvc) WithFXService(_ wallet.FXRateGetter) wallet.Service { return m }
func (m *faucetMockSvc) WithIssuers(_, _ string) wallet.Service      { return m }
func (m *faucetMockSvc) Delete(_ context.Context, _ string) error    { return nil }
func (m *faucetMockSvc) List(_ context.Context) ([]*domain.Wallet, error) {
	return nil, nil
}

func faucetRequest(t *testing.T, h *wallet.Handler, walletID, body string) *httptest.ResponseRecorder {
	t.Helper()
	r := chi.NewRouter()
	h.Routes()(r)
	req := httptest.NewRequest(http.MethodPost, "/"+walletID+"/faucet", strings.NewReader(body))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestFaucetDisabledReturns404(t *testing.T) {
	svc := &faucetMockSvc{}
	h := wallet.NewHandler(svc).WithFaucet(false, 1000)

	rec := faucetRequest(t, h, "w1", `{"asset_code":"USDC","amount":10}`)
	if rec.Code != http.StatusNotFound {
		t.Errorf("disabled faucet: status = %d, want 404", rec.Code)
	}
	if len(svc.calls) != 0 {
		t.Errorf("disabled faucet must not reach the service; got %d calls", len(svc.calls))
	}
}

func TestFaucetOverCapReturns400(t *testing.T) {
	svc := &faucetMockSvc{}
	h := wallet.NewHandler(svc).WithFaucet(true, 1000)

	rec := faucetRequest(t, h, "w1", `{"asset_code":"USDC","amount":1001}`)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("over-cap request: status = %d, want 400 (body %s)", rec.Code, rec.Body.String())
	}
	if len(svc.calls) != 0 {
		t.Errorf("over-cap request must not reach the service; got %d calls", len(svc.calls))
	}
}

func TestFaucetEnabledWithinCapSucceeds(t *testing.T) {
	svc := &faucetMockSvc{}
	h := wallet.NewHandler(svc).WithFaucet(true, 1000)

	rec := faucetRequest(t, h, "w1", `{"asset_code":"USDC","amount":250}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid request: status = %d, want 200 (body %s)", rec.Code, rec.Body.String())
	}
	if len(svc.calls) != 1 {
		t.Fatalf("service calls = %d, want 1", len(svc.calls))
	}
	if !svc.calls[0].amount.Equal(decimal.NewFromInt(250)) {
		t.Errorf("amount passed to service = %s, want 250", svc.calls[0].amount)
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not JSON: %v", err)
	}
	if resp["wallet_id"] != "w1" {
		t.Errorf("wallet_id = %v, want w1", resp["wallet_id"])
	}
}

func TestFaucetDailyLimitPerWallet(t *testing.T) {
	svc := &faucetMockSvc{}
	h := wallet.NewHandler(svc).WithFaucet(true, 1000) // max doubles as daily limit

	// First draw of 600 succeeds; second draw of 500 would exceed the 1000
	// daily budget for the same wallet.
	if rec := faucetRequest(t, h, "w1", `{"amount":600}`); rec.Code != http.StatusOK {
		t.Fatalf("first draw: status = %d, want 200", rec.Code)
	}
	if rec := faucetRequest(t, h, "w1", `{"amount":500}`); rec.Code != http.StatusTooManyRequests {
		t.Errorf("second over-budget draw: status = %d, want 429", rec.Code)
	}
	// A different wallet has its own budget.
	if rec := faucetRequest(t, h, "w2", `{"amount":500}`); rec.Code != http.StatusOK {
		t.Errorf("other wallet draw: status = %d, want 200", rec.Code)
	}
	if len(svc.calls) != 2 {
		t.Errorf("service calls = %d, want 2 (rejected draws must not reach the service)", len(svc.calls))
	}
}
