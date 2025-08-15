package tasks

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// TaskHandler defines the interface for all task handlers
type TaskHandler interface {
	Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
}

// TaskFactory creates and manages task handlers
type TaskFactory struct {
	logger *zap.Logger
	handlers map[string]TaskHandler
}

// NewTaskFactory creates a new task factory
func NewTaskFactory(logger *zap.Logger) *TaskFactory {
	factory := &TaskFactory{
		logger: logger,
		handlers: make(map[string]TaskHandler),
	}

	// Register all task handlers
	factory.registerHandlers()

	return factory
}

// registerHandlers registers all available task handlers
func (f *TaskFactory) registerHandlers() {
	f.handlers["validate_application"] = NewValidateApplicationTaskHandler(f.logger)
	f.handlers["document_collection"] = NewDocumentCollectionTaskHandler(f.logger)
	f.handlers["identity_verification"] = NewIdentityVerificationTaskHandler(f.logger)
	f.handlers["update_application_state"] = NewUpdateApplicationStateTaskHandler(f.logger)
	f.handlers["finalize_loan_decision"] = NewFinalizeLoanDecisionTaskHandler(f.logger)
}

// GetHandler returns a task handler for the given task type
func (f *TaskFactory) GetHandler(taskType string) (TaskHandler, error) {
	if f == nil {
		return nil, fmt.Errorf("task factory is nil")
	}
	if f.handlers == nil {
		return nil, fmt.Errorf("task handlers map is nil")
	}
	
	handler, exists := f.handlers[taskType]
	if !exists {
		// Log available handlers for debugging
		availableHandlers := make([]string, 0, len(f.handlers))
		for handlerType := range f.handlers {
			availableHandlers = append(availableHandlers, handlerType)
		}
		return nil, fmt.Errorf("unknown task type: %s. Available handlers: %v", taskType, availableHandlers)
	}
	return handler, nil
}

// ExecuteTask executes a task with the given type and input
func (f *TaskFactory) ExecuteTask(ctx context.Context, taskType string, input map[string]interface{}) (map[string]interface{}, error) {
	handler, err := f.GetHandler(taskType)
	if err != nil {
		return nil, err
	}

	return handler.Execute(ctx, input)
}

// GetSupportedTaskTypes returns a list of all supported task types
func (f *TaskFactory) GetSupportedTaskTypes() []string {
	types := make([]string, 0, len(f.handlers))
	for taskType := range f.handlers {
		types = append(types, taskType)
	}
	return types
}
