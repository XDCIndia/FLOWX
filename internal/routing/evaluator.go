package routing

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
)

// Evaluator collects quotes from all registered routes and returns viable options.
type Evaluator struct {
	routes []PaymentRoute
	mu     sync.RWMutex
}

// NewEvaluator creates a route evaluator with the given routes.
func NewEvaluator(routes ...PaymentRoute) *Evaluator {
	return &Evaluator{routes: routes}
}

// Register adds a route to the evaluator.
func (e *Evaluator) Register(route PaymentRoute) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.routes = append(e.routes, route)
	log.Info().Str("route_id", string(route.ID())).Str("name", route.Name()).Msg("routing: route registered")
}

// Evaluate collects quotes from all routes that support the requested pair.
func (e *Evaluator) Evaluate(ctx context.Context, req PaymentRequest) ([]RouteQuote, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.routes) == 0 {
		return nil, fmt.Errorf("no payment routes registered")
	}

	var (
		quotes []RouteQuote
		mu     sync.Mutex
		wg     sync.WaitGroup
	)

	pair := req.SourceAsset + "-" + req.DestAsset
	for _, route := range e.routes {
		supported := route.Supports(req.SourceAsset, req.DestAsset, req.SourceRegion, req.DestRegion)
		log.Debug().Str("route_id", string(route.ID())).Str("pair", pair).Bool("supports", supported).Msg("evaluating route")
		if !supported {
			continue
		}
		wg.Add(1)
		go func(r PaymentRoute) {
			defer wg.Done()
			log.Debug().Str("route_id", string(r.ID())).Msg("calling Quote")
			quote, err := r.Quote(ctx, req.SourceAsset, req.DestAsset, req.Amount)
			if err != nil {
				log.Debug().Err(err).Str("route", string(r.ID())).Msg("routing: quote failed")
				return
			}
			mu.Lock()
			quotes = append(quotes, *quote)
			mu.Unlock()
		}(route)
	}
	wg.Wait()
log.Debug().Int("quotes_count", len(quotes)).Msg("quotes collected")

	if len(quotes) == 0 {
		return nil, fmt.Errorf("no routes available for %s", pair)
	}

	return quotes, nil
}

// SupportsPair checks if any registered route handles this pair.
func (e *Evaluator) SupportsPair(from, to string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, r := range e.routes {
		if r.Supports(from, to, "", "") {
			return true
		}
	}
	return false
}

// ListRoutes returns capabilities of all registered routes.
func (e *Evaluator) ListRoutes() []RouteCapability {
	e.mu.RLock()
	defer e.mu.RUnlock()
	caps := make([]RouteCapability, 0, len(e.routes))
	for _, r := range e.routes {
		caps = append(caps, RouteCapability{
			ID:   r.ID(),
			Name: r.Name(),
		})
	}
	return caps
}
