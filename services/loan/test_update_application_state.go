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

	// Create a test application
	testApp := &domain.LoanApplication{
		ID:           "test-app-123",
		UserID:       "test-user-456",
		LoanAmount:   25000.0,
		CurrentState: domain.StateInitiated,
		Status:       domain.StatusDraft,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Add it to the mock repository
	mockRepo.applications[testApp.ID] = testApp

	// Create task handler with repository
	handler := tasks.NewUpdateApplicationStateTaskHandlerWithRepository(logger, mockRepo)

	// Test case 1: Valid state transition
	fmt.Println("=== Test Case 1: Valid State Transition (initiated -> pre_qualified) ===")
	input1 := map[string]interface{}{
		"applicationId": "test-app-123",
		"fromState":     "initiated",
		"toState":       "pre_qualified",
		"reason":        "Pre-qualification completed successfully",
		"userId":        "test-user-456",
		"automated":     true,
	}

	result1, err1 := handler.Execute(context.Background(), input1)
	if err1 != nil {
		fmt.Printf("ERROR: %v\n", err1)
	} else {
		fmt.Printf("SUCCESS: %+v\n", result1)
		// Check if application state was updated
		updatedApp, _ := mockRepo.GetApplicationByID(context.Background(), "test-app-123")
		fmt.Printf("Updated Application State: %s\n", updatedApp.CurrentState)
		fmt.Printf("Updated Application Status: %s\n", updatedApp.Status)
		fmt.Printf("State Transitions Created: %d\n", len(mockRepo.stateTransitions))
	}

	fmt.Println("\n=== Test Case 2: Invalid State Transition (pre_qualified -> funded) ===")
	input2 := map[string]interface{}{
		"applicationId": "test-app-123",
		"fromState":     "pre_qualified",
		"toState":       "funded", // Invalid transition - skipping required steps
		"reason":        "Attempting invalid transition",
		"automated":     false,
	}

	result2, err2 := handler.Execute(context.Background(), input2)
	if err2 != nil {
		fmt.Printf("EXPECTED ERROR: %v\n", err2)
	} else {
		fmt.Printf("UNEXPECTED SUCCESS: %+v\n", result2)
	}

	fmt.Println("\n=== Test Case 3: Valid Sequential Transitions ===")
	// pre_qualified -> documents_submitted
	input3 := map[string]interface{}{
		"applicationId": "test-app-123",
		"toState":       "documents_submitted",
		"reason":        "Documents uploaded and submitted",
		"automated":     true,
	}

	result3, err3 := handler.Execute(context.Background(), input3)
	if err3 != nil {
		fmt.Printf("ERROR: %v\n", err3)
	} else {
		fmt.Printf("SUCCESS: %+v\n", result3)
	}

	// documents_submitted -> identity_verified
	input4 := map[string]interface{}{
		"applicationId": "test-app-123",
		"toState":       "identity_verified",
		"reason":        "Identity verification completed",
		"automated":     true,
	}

	result4, err4 := handler.Execute(context.Background(), input4)
	if err4 != nil {
		fmt.Printf("ERROR: %v\n", err4)
	} else {
		fmt.Printf("SUCCESS: %+v\n", result4)
	}

	// identity_verified -> underwriting
	input5 := map[string]interface{}{
		"applicationId": "test-app-123",
		"toState":       "underwriting",
		"reason":        "Application moved to underwriting",
		"automated":     true,
	}

	result5, err5 := handler.Execute(context.Background(), input5)
	if err5 != nil {
		fmt.Printf("ERROR: %v\n", err5)
	} else {
		fmt.Printf("SUCCESS: %+v\n", result5)
	}

	// underwriting -> approved
	input6 := map[string]interface{}{
		"applicationId": "test-app-123",
		"toState":       "approved",
		"reason":        "Application approved after underwriting",
		"automated":     false,
		"userId":        "underwriter-789",
	}

	result6, err6 := handler.Execute(context.Background(), input6)
	if err6 != nil {
		fmt.Printf("ERROR: %v\n", err6)
	} else {
		fmt.Printf("SUCCESS: %+v\n", result6)
		// Check final state
		finalApp, _ := mockRepo.GetApplicationByID(context.Background(), "test-app-123")
		fmt.Printf("Final Application State: %s\n", finalApp.CurrentState)
		fmt.Printf("Final Application Status: %s\n", finalApp.Status)
		fmt.Printf("Total State Transitions: %d\n", len(mockRepo.stateTransitions))
	}

	fmt.Println("\n=== Test Case 4: Missing Required Fields ===")
	input7 := map[string]interface{}{
		"applicationId": "test-app-123",
		// Missing toState
		"reason": "Missing target state",
	}

	result7, err7 := handler.Execute(context.Background(), input7)
	if err7 != nil {
		fmt.Printf("EXPECTED ERROR: %v\n", err7)
	} else {
		fmt.Printf("UNEXPECTED SUCCESS: %+v\n", result7)
	}

	fmt.Println("\n=== Test Case 5: Simulation Mode (No Repository) ===")
	handlerNoRepo := tasks.NewUpdateApplicationStateTaskHandler(logger)

	input8 := map[string]interface{}{
		"applicationId": "sim-app-123",
		"fromState":     "initiated",
		"toState":       "pre_qualified",
		"reason":        "Simulation test",
		"automated":     true,
	}

	result8, err8 := handlerNoRepo.Execute(context.Background(), input8)
	if err8 != nil {
		fmt.Printf("ERROR: %v\n", err8)
	} else {
		fmt.Printf("SIMULATION SUCCESS: %+v\n", result8)
	}

	fmt.Println("\n=== Summary ===")
	fmt.Printf("Total test applications in repository: %d\n", len(mockRepo.applications))
	fmt.Printf("Total state transitions recorded: %d\n", len(mockRepo.stateTransitions))

	// Print all state transitions
	fmt.Println("\nState Transition History:")
	for i, transition := range mockRepo.stateTransitions {
		fmt.Printf("%d. %s -> %s (Reason: %s, Automated: %t)\n",
			i+1,
			func() string {
				if transition.FromState != nil {
					return string(*transition.FromState)
				}
				return "nil"
			}(),
			transition.ToState,
			transition.TransitionReason,
			transition.Automated)
	}
}
