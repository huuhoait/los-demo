package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"

	"loan-service/domain"
	"loan-service/infrastructure/workflow/tasks"
)

// MockLoanRepository is a mock implementation for testing
type MockLoanRepository struct {
	applications     map[string]*domain.LoanApplication
	stateTransitions []domain.StateTransition
}

func NewMockLoanRepository() *MockLoanRepository {
	return &MockLoanRepository{
		applications:     make(map[string]*domain.LoanApplication),
		stateTransitions: make([]domain.StateTransition, 0),
	}
}

func (m *MockLoanRepository) GetApplicationByID(ctx context.Context, id string) (*domain.LoanApplication, error) {
	app, exists := m.applications[id]
	if !exists {
		return nil, fmt.Errorf("application not found: %s", id)
	}
	return app, nil
}

func (m *MockLoanRepository) UpdateApplication(ctx context.Context, app *domain.LoanApplication) error {
	m.applications[app.ID] = app
	return nil
}

func (m *MockLoanRepository) CreateStateTransition(ctx context.Context, transition *domain.StateTransition) error {
	m.stateTransitions = append(m.stateTransitions, *transition)
	return nil
}

func main() {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}
	defer logger.Sync()

	// Create mock repository
	mockRepo := NewMockLoanRepository()

	// Create a test application already in documents_submitted state
	testApp := &domain.LoanApplication{
		ID:           "test-app-123",
		UserID:       "test-user-456",
		LoanAmount:   25000.0,
		CurrentState: domain.StateDocumentsSubmitted, // Already in target state
		Status:       domain.StatusUnderReview,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Add it to the mock repository
	mockRepo.applications[testApp.ID] = testApp

	// Create task handler with repository
	handler := tasks.NewUpdateApplicationStateTaskHandlerWithRepository(logger, mockRepo)

	// Test case: Idempotent operation (already in target state)
	fmt.Println("=== Test Case: Idempotent State Transition (documents_submitted -> documents_submitted) ===")
	input := map[string]interface{}{
		"applicationId": "test-app-123",
		"fromState":     "pre_qualified",       // This doesn't match current state
		"toState":       "documents_submitted", // Already in this state
		"reason":        "Required documents collected",
		"automated":     true,
	}

	result, err := handler.Execute(context.Background(), input)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("SUCCESS: %+v\n", result)

		// Check if the idempotent flag is set
		if idempotent, ok := result["idempotent"].(bool); ok && idempotent {
			fmt.Println("✅ Idempotent operation detected and handled correctly!")
		} else {
			fmt.Println("❌ Idempotent operation not properly handled")
		}

		// Verify application state remains unchanged
		updatedApp, _ := mockRepo.GetApplicationByID(context.Background(), "test-app-123")
		fmt.Printf("Application State After: %s (should remain documents_submitted)\n", updatedApp.CurrentState)
		fmt.Printf("State Transitions Created: %d (should be 0 for idempotent operation)\n", len(mockRepo.stateTransitions))
	}

	fmt.Println("\n=== Test Case 2: Valid State Transition (documents_submitted -> identity_verified) ===")
	input2 := map[string]interface{}{
		"applicationId": "test-app-123",
		"fromState":     "documents_submitted", // Matches current state
		"toState":       "identity_verified",   // Valid next state
		"reason":        "Identity verification completed",
		"automated":     true,
	}

	result2, err2 := handler.Execute(context.Background(), input2)
	if err2 != nil {
		fmt.Printf("ERROR: %v\n", err2)
	} else {
		fmt.Printf("SUCCESS: %+v\n", result2)

		// Check if the idempotent flag is NOT set
		if idempotent, ok := result2["idempotent"].(bool); ok && idempotent {
			fmt.Println("❌ This should not be idempotent!")
		} else {
			fmt.Println("✅ Normal state transition handled correctly!")
		}

		// Verify application state changed
		updatedApp, _ := mockRepo.GetApplicationByID(context.Background(), "test-app-123")
		fmt.Printf("Application State After: %s (should be identity_verified)\n", updatedApp.CurrentState)
		fmt.Printf("State Transitions Created: %d (should be 1 for normal transition)\n", len(mockRepo.stateTransitions))
	}
}
