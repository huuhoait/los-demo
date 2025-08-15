package tasks

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// UpdateApplicationStateTaskHandler handles application state update tasks
type UpdateApplicationStateTaskHandler struct {
	logger *zap.Logger
}

// NewUpdateApplicationStateTaskHandler creates a new update application state task handler
func NewUpdateApplicationStateTaskHandler(logger *zap.Logger) *UpdateApplicationStateTaskHandler {
	return &UpdateApplicationStateTaskHandler{
		logger: logger,
	}
}

// Execute updates loan application state in database
func (h *UpdateApplicationStateTaskHandler) Execute(
	ctx context.Context,
	input map[string]interface{},
) (map[string]interface{}, error) {
	logger := h.logger.With(zap.String("operation", "update_application_state"))

	logger.Info("Updating application state")

	// Extract input parameters
	applicationID, _ := input["applicationId"].(string)
	fromState, _ := input["fromState"].(string)
	toState, _ := input["toState"].(string)
	reason, _ := input["reason"].(string)

	// Validate required fields
	if applicationID == "" {
		return nil, fmt.Errorf("application ID is required")
	}

	if toState == "" {
		return nil, fmt.Errorf("target state is required")
	}

	logger.Info("Application state transition",
		zap.String("application_id", applicationID),
		zap.String("from_state", fromState),
		zap.String("to_state", toState),
		zap.String("reason", reason))

	// Simulate state update process
	// In real implementation, this would:
	// 1. Validate state transition is allowed
	// 2. Update application state in database
	// 3. Create state transition record
	// 4. Trigger any state-specific actions

	updatedAt := time.Now()

	logger.Info("Application state updated successfully",
		zap.String("application_id", applicationID),
		zap.String("new_state", toState),
		zap.Time("updated_at", updatedAt))

	return map[string]interface{}{
		"success":    true,
		"updatedAt":  updatedAt,
		"newState":   toState,
		"transition": map[string]interface{}{
			"fromState": fromState,
			"toState":   toState,
			"reason":    reason,
			"timestamp": updatedAt,
		},
	}, nil
}
