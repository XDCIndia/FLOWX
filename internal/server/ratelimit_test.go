package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// AuthRateLimit must engage once a single IP exhausts its burst: requests
// beyond the burst are rejected with 429 while the first burst passes through.
func TestAuthRateLimitEngagesPerIP(t *testing.T) {
	handler := AuthRateLimit(5.0/60.0, 10)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	const burst = 10
	for i := 0; i < burst; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil))
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: got %d, want 200 (within burst)", i+1, rec.Code)
		}
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil))
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("request beyond burst: got %d, want 429", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "RATE_LIMITED") {
		t.Fatalf("429 body = %q, want RATE_LIMITED error", rec.Body.String())
	}
}

// A different source IP must have its own bucket: one IP being limited must
// not throttle unrelated clients.
func TestAuthRateLimitIsPerIP(t *testing.T) {
	handler := AuthRateLimit(5.0/60.0, 10)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	flood := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
	for i := 0; i < 15; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, flood)
	}

	other := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
	other.RemoteAddr = "198.51.100.7:5555"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, other)
	if rec.Code != http.StatusOK {
		t.Fatalf("different IP throttled by another client's flood: got %d, want 200", rec.Code)
	}
}

// Wiring: the strict limiter applies to the public /v1/auth group only.
// Hammering login must trip 429, while a non-auth route (authenticated
// /v1/usage, unauthenticated here) must still answer 401 — not 429 — proving
// the authenticated group's limits were not tightened.
func TestAuthRateLimitAppliesOnlyToAuthRoutes(t *testing.T) {
	srv := newAuthzTestServerWithValidator(t, newMockMembershipValidator())

	// 10 requests inside the burst: auth handler rejects the empty body with
	// 400 (the nil service is never reached), proving the limiter passed them
	// through.
	for i := 0; i < 10; i++ {
		rec := httptest.NewRecorder()
		srv.router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil))
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("auth request %d: got %d, want 400 (limiter should pass burst through)", i+1, rec.Code)
		}
	}

	// Burst exhausted: next auth request is rate limited before the handler.
	rec := httptest.NewRecorder()
	srv.router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil))
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("auth request beyond burst: got %d, want 429", rec.Code)
	}

	// Non-auth route from the same IP: must NOT be affected by the auth
	// limiter (401 from the auth middleware, not 429).
	rec = httptest.NewRecorder()
	srv.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/v1/usage", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("non-auth route: got %d, want 401 (auth limiter leaked outside /v1/auth)", rec.Code)
	}
}
