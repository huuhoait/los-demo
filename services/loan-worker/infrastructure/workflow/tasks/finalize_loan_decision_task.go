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
	currentState, _ := input["currentState"].(string)

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

	// Validate that the task is being called in the correct workflow sequence
	// This task should only be called after identity verification and underwriting
	if currentState != "" {
		validStatesForFinalization := []string{
			"identity_verified",
			"approved",
			"denied",
			"manual_review",
		}

		isValidState := false
		for _, validState := range validStatesForFinalization {
			if currentState == validState {
				isValidState = true
				break
			}
		}

		if !isValidState {
			logger.Error("Invalid application state for loan decision finalization",
				zap.String("current_state", currentState),
				zap.String("required_states", fmt.Sprintf("%v", validStatesForFinalization)))
			return nil, fmt.Errorf("invalid state for loan decision finalization: current state '%s' is not eligible for finalization. Required states: %v",
				currentState, validStatesForFinalization)
		}
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
		zap.String("current_state", currentState),
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
