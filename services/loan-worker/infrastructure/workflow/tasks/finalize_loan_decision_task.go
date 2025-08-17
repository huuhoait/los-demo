package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// FinalizeLoanDecisionTaskHandler handles loan decision finalization tasks
type FinalizeLoanDecisionTaskHandler struct {
	logger *zap.Logger
}

// NewFinalizeLoanDecisionTaskHandler creates a new finalize loan decision task handler
func NewFinalizeLoanDecisionTaskHandler(logger *zap.Logger) *FinalizeLoanDecisionTaskHandler {
	return &FinalizeLoanDecisionTaskHandler{
		logger: logger,
	}
}

// Execute finalizes loan processing decision and updates application
func (h *FinalizeLoanDecisionTaskHandler) Execute(
	ctx context.Context,
	input map[string]interface{},
) (map[string]interface{}, error) {
	logger := h.logger.With(zap.String("operation", "finalize_loan_decision"))

	logger.Info("Finalizing loan decision")

	// Extract input parameters
	applicationID, _ := input["applicationId"].(string)
	finalState, _ := input["finalState"].(string)
	decision, _ := input["decision"].(string)
	interestRate, _ := input["interestRate"].(float64)

	// Validate required fields
	if applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}

	if finalState == "" {
		return nil, fmt.Errorf("final state is required")
	}

	if decision == "" {
		return nil, fmt.Errorf("decision is required")
	}

	// Simulate finalization process
	completedAt := time.Now()
	nextSteps := []string{}

	switch decision {
	case "approved":
		nextSteps = []string{
			"Document signing",
			"Fund disbursement",
			"Loan activation",
		}
	case "denied":
		nextSteps = []string{
			"Send denial letter",
			"Archive application",
		}
	case "manual_review":
		nextSteps = []string{
			"Manual review",
			"Additional documentation request",
		}
	default:
		nextSteps = []string{"Review decision"}
	}

	logger.Info("Loan decision finalized",
		zap.String("application_id", applicationID),
		zap.String("final_state", finalState),
		zap.String("decision", decision),
		zap.Float64("interest_rate", interestRate),
		zap.Time("completed_at", completedAt))

	return map[string]interface{}{
		"finalState":   finalState,
		"decision":     decision,
		"interestRate": interestRate,
		"completedAt":  completedAt,
		"nextSteps":    nextSteps,
		"summary": map[string]interface{}{
			"applicationId": applicationID,
			"status":        finalState,
			"decision":      decision,
			"timestamp":     completedAt,
		},
	}, nil
}
