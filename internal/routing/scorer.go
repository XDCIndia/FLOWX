package routing

import (
	"sort"

	"github.com/shopspring/decimal"
)

// RankingMode selects which dimension to optimize for.
type RankingMode string

const (
	RankingBalanced  RankingMode = "balanced"
	RankingCheapest  RankingMode = "cheapest"
	RankingFastest   RankingMode = "fastest"
	RankingReliable  RankingMode = "most_reliable"
)

// ScoringWeights configure how each dimension contributes to the composite score.
type ScoringWeights struct {
	Cost       float64 `json:"cost"`        // default 0.35
	Speed      float64 `json:"speed"`       // default 0.25
	Reliability float64 `json:"reliability"` // default 0.20
	Compliance float64 `json:"compliance"`  // default 0.10
	Liquidity  float64 `json:"liquidity"`   // default 0.10
}

// DefaultWeights returns balanced scoring weights.
func DefaultWeights() ScoringWeights {
	return ScoringWeights{
		Cost:        0.35,
		Speed:       0.25,
		Reliability: 0.20,
		Compliance:  0.10,
		Liquidity:   0.10,
	}
}

// RouteScore is a scored route option with breakdown.
type RouteScore struct {
	Quote       RouteQuote    `json:"quote"`
	Score       float64       `json:"score"`        // 0-100 composite
	CostScore   float64       `json:"cost_score"`   // 0-100
	SpeedScore  float64       `json:"speed_score"`  // 0-100
	Reliability float64       `json:"reliability"`  // 0-100
	Compliance  float64       `json:"compliance"`   // 0-100
	Liquidity   float64       `json:"liquidity"`    // 0-100
	Warnings    []string      `json:"warnings,omitempty"`
	Recommended bool          `json:"recommended"`
}

// Scorer scores and ranks route quotes.
type Scorer struct {
	weights  ScoringWeights
	mode     RankingMode
	history  map[RouteID]float64 // historical success rates (0-1)
}

// NewScorer creates a scorer with the given weights and ranking mode.
func NewScorer(weights ScoringWeights, mode RankingMode) *Scorer {
	if weights.Cost == 0 && weights.Speed == 0 {
		weights = DefaultWeights()
	}
	return &Scorer{weights: weights, mode: mode, history: make(map[RouteID]float64)}
}

// UpdateSuccessRate records a route's success rate for reliability scoring.
func (s *Scorer) UpdateSuccessRate(routeID RouteID, success bool) {
	current := s.history[routeID]
	if success {
		s.history[routeID] = current*0.9 + 0.1 // exponential moving average
	} else {
		s.history[routeID] = current * 0.9
	}
}

// ScoreAndRank scores each quote and returns them sorted by composite score.
func (s *Scorer) ScoreAndRank(quotes []RouteQuote) []RouteScore {
	if len(quotes) == 0 {
		return nil
	}

	scores := make([]RouteScore, 0, len(quotes))
	for _, q := range quotes {
		scores = append(scores, s.scoreQuote(q))
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Score > scores[j].Score
	})

	// Mark recommended
	if len(scores) > 0 {
		scores[0].Recommended = true
	}

	return scores
}

func (s *Scorer) scoreQuote(q RouteQuote) RouteScore {
	rs := RouteScore{Quote: q}

	// Cost score: lower total cost (fee + spread) = higher score
	totalCostPct := q.Fee.Div(q.SourceAmount).Mul(decimal.NewFromInt(100)).InexactFloat64()
	spreadCost := float64(q.SpreadBps) / 100.0
	combinedCost := totalCostPct + spreadCost
	rs.CostScore = clamp(100 - combinedCost*10) // 10% cost = 0 score

	// Speed score: faster settlement = higher score (30s=100, 5min=50, 1hr=10)
	minutes := q.SettlementTime.Minutes()
	if minutes <= 0.5 {
		rs.SpeedScore = 100
	} else if minutes <= 5 {
		rs.SpeedScore = 80
	} else if minutes <= 30 {
		rs.SpeedScore = 50
	} else {
		rs.SpeedScore = 20
	}

	// Reliability score: from historical success rate
	if rate, ok := s.history[q.RouteID]; ok {
		rs.Reliability = rate * 100
	} else {
		rs.Reliability = 85 // default for unknown routes
	}

	// Compliance score: simpler = higher (no KYC = 100, KYC required = 60)
	rs.Compliance = 85

	// Liquidity score: higher dest amount relative to source = better
	if q.SourceAmount.GreaterThan(decimal.Zero) {
		efficiency := q.DestAmount.Div(q.SourceAmount).InexactFloat64()
		rs.Liquidity = clamp(efficiency * 50) // normalized
	}

	// Composite score
	w := s.weights
	rs.Score = rs.CostScore*w.Cost +
		rs.SpeedScore*w.Speed +
		rs.Reliability*w.Reliability +
		rs.Compliance*w.Compliance +
		rs.Liquidity*w.Liquidity

	// Apply ranking mode bias
	switch s.mode {
	case RankingCheapest:
		rs.Score = rs.Score*0.6 + rs.CostScore*0.4
	case RankingFastest:
		rs.Score = rs.Score*0.6 + rs.SpeedScore*0.4
	case RankingReliable:
		rs.Score = rs.Score*0.6 + rs.Reliability*0.4
	}

	return rs
}

func clamp(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
