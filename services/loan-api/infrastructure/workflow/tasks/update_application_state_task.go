package tasks

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/lendingplatform/los/services/loan-api/domain"
)

// UpdateApplicationStateTaskHandler handles application state update tasks
type UpdateApplicationStateTaskHandler struct {
	logger         *zap.Logger
	loanRepository LoanRepository
}

// NewUpdateApplicationStateTaskHandler creates a new update application state task handler
func NewUpdateApplicationStateTaskHandler(logger *zap.Logger) *UpdateApplicationStateTaskHandler {
	return &UpdateApplicationStateTaskHandler{
		logger: logger,
	}
}

// NewUpdateApplicationStateTaskHandlerWithRepository creates a new update application state task handler with repository
func NewUpdateApplicationStateTaskHandlerWithRepository(logger *zap.Logger, loanRepository LoanRepository) *UpdateApplicationStateTaskHandler {
	return &UpdateApplicationStateTaskHandler{
		logger:         logger,
		loanRepository: loanRepository,
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
	userID, _ := input["userId"].(string)
	automated, _ := input["automated"].(bool)

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
		zap.String("reason", reason),
		zap.String("user_id", userID),
		zap.Bool("automated", automated))

	// If no repository is configured, fall back to simulation mode
	if h.loanRepository == nil {
		logger.Warn("No loan repository configured, running in simulation mode")
		return h.simulateStateUpdate(applicationID, fromState, toState, reason, automated)
	}

	// Get current application from database
	application, err := h.loanRepository.GetApplicationByID(ctx, applicationID)
	if err != nil {
		logger.Error("Failed to get application by ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	// Validate state transition
	targetState := domain.ApplicationState(toState)

	// Check if already in target state (idempotent operation)
	if application.CurrentState == targetState {
		logger.Info("Application already in target state, treating as successful idempotent operation",
			zap.String("current_state", string(application.CurrentState)),
			zap.String("target_state", toState))

		// Return success response for idempotent operation
		return map[string]interface{}{
			"success":       true,
			"newState":      string(targetState),
			"previousState": string(application.CurrentState),
			"newStatus":     string(application.Status),
			"updatedAt":     time.Now().UTC().Format(time.RFC3339),
			"idempotent":    true,
			"message":       "Application already in target state",
		}, nil
	}

	if !application.CanTransitionTo(targetState) {
		logger.Error("Invalid state transition",
			zap.String("current_state", string(application.CurrentState)),
			zap.String("target_state", toState))
		return nil, fmt.Errorf("invalid state transition from %s to %s", application.CurrentState, toState)
	}

	// Store previous state for transition record
	previousState := application.CurrentState

	// Update application state
	application.CurrentState = targetState
	application.UpdatedAt = time.Now().UTC()

	// Update status based on state
	switch targetState {
	case domain.StateApproved:
		application.Status = domain.StatusApproved
	case domain.StateDenied:
		application.Status = domain.StatusDenied
	case domain.StateFunded:
		application.Status = domain.StatusFunded
	case domain.StateActive:
		application.Status = domain.StatusActive
	case domain.StateClosed:
		application.Status = domain.StatusClosed
	default:
		// For other states, keep status as submitted or under_review
		if application.Status == domain.StatusDraft {
			application.Status = domain.StatusSubmitted
		} else if application.Status != domain.StatusApproved &&
			application.Status != domain.StatusDenied &&
			application.Status != domain.StatusFunded &&
			application.Status != domain.StatusActive &&
			application.Status != domain.StatusClosed {
			application.Status = domain.StatusUnderReview
		}
	}

	// Save updated application to database
	if err := h.loanRepository.UpdateApplication(ctx, application); err != nil {
		logger.Error("Failed to update application state", zap.Error(err))
		return nil, fmt.Errorf("failed to update application state: %w", err)
	}

	// Create state transition record
	transition := &domain.StateTransition{
		ID:               uuid.New().String(),
		ApplicationID:    applicationID,
		FromState:        &previousState,
		ToState:          targetState,
		TransitionReason: reason,
		Automated:        automated,
		CreatedAt:        time.Now().UTC(),
	}

	if userID != "" {
		transition.UserID = &userID
	}

	if err := h.loanRepository.CreateStateTransition(ctx, transition); err != nil {
		logger.Error("Failed to create state transition record", zap.Error(err))
		// Don't fail the entire operation if transition record creation fails
		logger.Warn("Continuing despite state transition record creation failure")
	}

	updatedAt := time.Now().UTC()

	logger.Info("Application state updated successfully",
		zap.String("application_id", applicationID),
		zap.String("previous_state", string(previousState)),
		zap.String("new_state", string(targetState)),
		zap.String("new_status", string(application.Status)),
		zap.Time("updated_at", updatedAt))

	return map[string]interface{}{
		"success":       true,
		"updatedAt":     updatedAt,
		"previousState": string(previousState),
		"newState":      string(targetState),
		"newStatus":     string(application.Status),
		"transition": map[string]interface{}{
			"id":        transition.ID,
			"fromState": string(previousState),
			"toState":   string(targetState),
			"reason":    reason,
			"automated": automated,
			"timestamp": updatedAt,
			"userId":    userID,
		},
	}, nil
}

// simulateStateUpdate simulates state update when no repository is available
func (h *UpdateApplicationStateTaskHandler) simulateStateUpdate(
	applicationID, fromState, toState, reason string, automated bool,
) (map[string]interface{}, error) {
	logger := h.logger.With(zap.String("operation", "simulate_state_update"))

	// Simulate state update process
	updatedAt := time.Now().UTC()

	// Basic validation of state transitions
	targetState := domain.ApplicationState(toState)

	// Simulate previous state if not provided
	var previousState domain.ApplicationState
	if fromState != "" {
		previousState = domain.ApplicationState(fromState)
	} else {
		// Default to initiated if no previous state
		previousState = domain.StateInitiated
	}

	// Create a mock application to test transition validity
	mockApp := &domain.LoanApplication{
		CurrentState: previousState,
	}

	if !mockApp.CanTransitionTo(targetState) {
		logger.Error("Invalid state transition in simulation",
			zap.String("from_state", string(previousState)),
			zap.String("to_state", toState))
		return nil, fmt.Errorf("invalid state transition from %s to %s", previousState, toState)
	}

	logger.Info("Application state simulation completed successfully",
		zap.String("application_id", applicationID),
		zap.String("previous_state", string(previousState)),
		zap.String("new_state", toState),
		zap.Time("updated_at", updatedAt))

	return map[string]interface{}{
		"success":       true,
		"updatedAt":     updatedAt,
		"previousState": string(previousState),
		"newState":      toState,
		"simulated":     true,
		"transition": map[string]interface{}{
			"fromState": string(previousState),
			"toState":   toState,
			"reason":    reason,
			"automated": automated,
			"timestamp": updatedAt,
		},
	}, nil
}
