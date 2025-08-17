package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"underwriting-worker/domain"
)

// UpdateApplicationStateTaskHandler handles application state update tasks
type UpdateApplicationStateTaskHandler struct {
	logger              *zap.Logger
	loanApplicationRepo domain.LoanApplicationRepository
}

// NewUpdateApplicationStateTaskHandler creates a new update application state task handler
func NewUpdateApplicationStateTaskHandler(
	logger *zap.Logger,
	loanApplicationRepo domain.LoanApplicationRepository,
) *UpdateApplicationStateTaskHandler {
	return &UpdateApplicationStateTaskHandler{
		logger:              logger,
		loanApplicationRepo: loanApplicationRepo,
	}
}

// Execute updates the application state
func (h *UpdateApplicationStateTaskHandler) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	startTime := time.Now()
	logger := h.logger.With(zap.String("operation", "update_application_state"))

	logger.Info("Starting application state update task")

	// Extract input parameters
	applicationID, ok := input["applicationId"].(string)
	if !ok || applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}

	newState, ok := input["newState"].(string)
	if !ok || newState == "" {
		return nil, fmt.Errorf("new state is required")
	}

	// Optional parameters
	userID, _ := input["userId"].(string)
	reason, _ := input["reason"].(string)
	metadata, _ := input["metadata"].(map[string]interface{})

	logger.Info("Updating application state",
		zap.String("application_id", applicationID),
		zap.String("new_state", newState),
		zap.String("user_id", userID),
		zap.String("reason", reason))

	// Update application state
	if h.loanApplicationRepo != nil {
		// Get current application
		application, err := h.loanApplicationRepo.GetByID(ctx, applicationID)
		if err != nil {
			logger.Error("Failed to get application", zap.Error(err))
			return h.createFailureResponse(applicationID, err), nil
		}

		oldState := application.CurrentState

		// Update state
		application.CurrentState = newState
		application.UpdatedAt = time.Now()

		// Update in repository
		if err := h.loanApplicationRepo.Update(ctx, application); err != nil {
			logger.Error("Failed to update application state", zap.Error(err))
			return h.createFailureResponse(applicationID, err), nil
		}

		logger.Info("Application state updated successfully",
			zap.String("application_id", applicationID),
			zap.String("old_state", oldState),
			zap.String("new_state", newState))

		return h.createSuccessResponse(applicationID, oldState, newState, time.Since(startTime), metadata), nil
	}

	// Mock successful update when repository is not available
	logger.Info("Mock application state update completed",
		zap.String("application_id", applicationID),
		zap.String("new_state", newState))

	return h.createMockSuccessResponse(applicationID, newState, time.Since(startTime), metadata), nil
}

// createSuccessResponse creates a successful response
func (h *UpdateApplicationStateTaskHandler) createSuccessResponse(
	applicationID, oldState, newState string,
	processingTime time.Duration,
	metadata map[string]interface{},
) map[string]interface{} {
	response := map[string]interface{}{
		"success":       true,
		"applicationId": applicationID,
		"stateTransition": map[string]interface{}{
			"previousState": oldState,
			"currentState":  newState,
			"transitionAt":  time.Now().UTC().Format(time.RFC3339),
		},
		"processingTime": processingTime.String(),
		"completedAt":    time.Now().UTC().Format(time.RFC3339),
	}

	if metadata != nil {
		response["metadata"] = metadata
	}

	return response
}

// createMockSuccessResponse creates a mock successful response
func (h *UpdateApplicationStateTaskHandler) createMockSuccessResponse(
	applicationID, newState string,
	processingTime time.Duration,
	metadata map[string]interface{},
) map[string]interface{} {
	response := map[string]interface{}{
		"success":       true,
		"applicationId": applicationID,
		"stateTransition": map[string]interface{}{
			"previousState": "unknown",
			"currentState":  newState,
			"transitionAt":  time.Now().UTC().Format(time.RFC3339),
		},
		"mock":           true,
		"processingTime": processingTime.String(),
		"completedAt":    time.Now().UTC().Format(time.RFC3339),
	}

	if metadata != nil {
		response["metadata"] = metadata
	}

	return response
}

// createFailureResponse creates a failure response
func (h *UpdateApplicationStateTaskHandler) createFailureResponse(applicationID string, err error) map[string]interface{} {
	return map[string]interface{}{
		"success":       false,
		"applicationId": applicationID,
		"error":         err.Error(),
		"completedAt":   time.Now().UTC().Format(time.RFC3339),
	}
}
