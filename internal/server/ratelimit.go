package server

import (
	"net/http"
	"sync"
	"time"

	"github.com/fluxa/fluxa/internal/tenant"
	"golang.org/x/time/rate"
)

const maxIdleTime = 5 * time.Minute

type visitorEntry struct {
	rateLimiter *rate.Limiter
	lastSeen    time.Time
}

type rateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitorEntry
	rate     rate.Limit
	burst    int
}

func newRateLimiter(rps float64, burst int) *rateLimiter {
	rl := &rateLimiter{
		visitors: make(map[string]*visitorEntry),
		rate:     rate.Limit(rps),
		burst:    burst,
	}
	go rl.cleanup()
	return rl
}

func (rl *rateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	entry, exists := rl.visitors[key]
	if !exists {
		entry = &visitorEntry{
			rateLimiter: rate.NewLimiter(rl.rate, rl.burst),
			lastSeen:    time.Now(),
		}
		rl.visitors[key] = entry
	} else {
		entry.lastSeen = time.Now()
	}
	return entry.rateLimiter
}

func (rl *rateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.visitors {
			if now.Sub(entry.lastSeen) > maxIdleTime {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit returns a middleware that limits requests per-tenant (or per-IP for unauthenticated).
func RateLimit(rps float64, burst int) func(http.Handler) http.Handler {
	globalLimiter := newRateLimiter(rps, burst)
	tenantLimiters := newRateLimiter(rps*10, burst*10) // 10x for authenticated tenants

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check global rate first
			if !globalLimiter.getLimiter("global").Allow() {
				http.Error(w, `{"error":{"code":"RATE_LIMITED","message":"global rate limit exceeded"}}`, http.StatusTooManyRequests)
				return
			}

			// Per-tenant or per-IP limiting
			tid := tenant.IDFromContext(r.Context())
			key := tid
			if key == "" {
				key = r.RemoteAddr
			}

			if !tenantLimiters.getLimiter(key).Allow() {
				http.Error(w, `{"error":{"code":"RATE_LIMITED","message":"rate limit exceeded"}}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// AuthRateLimit returns a strict per-IP limiter intended ONLY for
// unauthenticated public auth endpoints (register/login/refresh), which are
// cheap to attack and carry no tenant identity to key on. Unlike RateLimit it
// has no global bucket and no tenant branch — every client IP gets its own
// tight bucket (e.g. 5 req/min with burst 10) so credential-stuffing and
// registration floods are throttled per source.
func AuthRateLimit(rps float64, burst int) func(http.Handler) http.Handler {
	limiters := newRateLimiter(rps, burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if ip == "" {
				ip = "unknown"
			}

			if !limiters.getLimiter(ip).Allow() {
				http.Error(w, `{"error":{"code":"RATE_LIMITED","message":"rate limit exceeded"}}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
