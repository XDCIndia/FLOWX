package routing

import (
	"context"
	"fmt"

	"github.com/shopspring/decimal"
)

// RiskLevel classifies the risk of a payment.
type RiskLevel string

const (
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
	RiskBlocked RiskLevel = "blocked"
)

// ComplianceResult holds the outcome of a compliance check for a route.
type ComplianceResult struct {
	RouteID     RouteID  `json:"route_id"`
	RiskLevel   RiskLevel `json:"risk_level"`
	RiskScore   int       `json:"risk_score"` // 0-100
	Blocked     bool      `json:"blocked"`
	Warnings    []string  `json:"warnings,omitempty"`
	ChecksRun   []string  `json:"checks_run"`
}

// ComplianceChecker runs per-route compliance checks.
type ComplianceChecker interface {
	CheckRoute(ctx context.Context, req PaymentRequest, route RouteID) (*ComplianceResult, error)
}

// DefaultComplianceChecker is a simple compliance checker that applies
// basic rules based on amount, route type, and risk profile.
type DefaultComplianceChecker struct {
	maxAmount     decimal.Decimal
velocityLimit int
}

// NewDefaultComplianceChecker creates a checker with sensible defaults.
func NewDefaultComplianceChecker() *DefaultComplianceChecker {
	return &DefaultComplianceChecker{
		maxAmount:     decimal.NewFromInt(10000000), // 10M default limit (covers fiat currencies)
		velocityLimit: 10,                        // max 10 transfers per hour
	}
}

// CheckRoute evaluates compliance for a specific route.
func (c *DefaultComplianceChecker) CheckRoute(ctx context.Context, req PaymentRequest, route RouteID) (*ComplianceResult, error) {
	result := &ComplianceResult{
		RouteID:   route,
		RiskLevel: RiskLow,
		RiskScore: 10,
		ChecksRun: []string{},
	}

	// Check 1: Amount limit
	result.ChecksRun = append(result.ChecksRun, "amount_limit")
	if req.Amount.GreaterThan(c.maxAmount) {
		result.RiskScore += 30
		result.Warnings = append(result.Warnings, fmt.Sprintf("amount %s exceeds limit %s", req.Amount, c.maxAmount))
		if req.Amount.GreaterThan(c.maxAmount.Mul(decimal.NewFromInt(5))) {
			result.RiskLevel = RiskBlocked
			result.Blocked = true
			result.Warnings = append(result.Warnings, "amount exceeds 5x limit — blocked")
			return result, nil
		}
		result.RiskLevel = RiskMedium
	}

	// Check 2: Risk profile
	result.ChecksRun = append(result.ChecksRun, "risk_profile")
	switch req.RiskProfile {
	case "high":
		result.RiskScore += 25
		result.Warnings = append(result.Warnings, "high risk profile flagged")
	case "medium":
		result.RiskScore += 10
	}

	// Check 3: Fiat route requires higher scrutiny
	result.ChecksRun = append(result.ChecksRun, "route_type")
	if route == RouteFiatNGN || route == RouteFiatKES {
		result.RiskScore += 10
		result.Warnings = append(result.Warnings, "fiat route — additional KYC may apply")
	}

	// Final risk classification
	switch {
	case result.RiskScore >= 70:
		result.RiskLevel = RiskBlocked
		result.Blocked = true
	case result.RiskScore >= 40:
		result.RiskLevel = RiskHigh
	case result.RiskScore >= 20:
		result.RiskLevel = RiskMedium
	default:
		result.RiskLevel = RiskLow
	}

	return result, nil
}
