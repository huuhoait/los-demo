package workflow

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/huuhoait/los-demo/services/loan-worker/infrastructure/workflow/tasks"
	"github.com/huuhoait/los-demo/services/shared/pkg/i18n"
)

// LoanProcessingTaskHandler handles loan processing tasks using the task factory
type LoanProcessingTaskHandler struct {
	logger      *zap.Logger
	localizer   *i18n.Localizer
	taskFactory *tasks.TaskFactory
}

// NewLoanProcessingTaskHandler creates a new loan processing task handler
func NewLoanProcessingTaskHandler(logger *zap.Logger, localizer *i18n.Localizer) *LoanProcessingTaskHandler {
	return &LoanProcessingTaskHandler{
		logger:      logger,
		localizer:   localizer,
		taskFactory: tasks.NewTaskFactory(logger),
	}
}

// NewLoanProcessingTaskHandlerWithRepository creates a new loan processing task handler with repository
func NewLoanProcessingTaskHandlerWithRepository(logger *zap.Logger, localizer *i18n.Localizer, loanRepository tasks.LoanRepository) *LoanProcessingTaskHandler {
	return &LoanProcessingTaskHandler{
		logger:      logger,
		localizer:   localizer,
		taskFactory: tasks.NewTaskFactoryWithRepository(logger, loanRepository),
	}
}

// HandleTask handles loan processing tasks by delegating to the appropriate task handler
func (h *LoanProcessingTaskHandler) HandleTask(
	ctx context.Context,
	taskType string,
	input map[string]interface{},
) (map[string]interface{}, error) {
	logger := h.logger.With(
		zap.String("task_type", taskType),
		zap.String("operation", "handle_task"),
	)

	logger.Info("Handling loan processing task",
		zap.String("task_type", taskType),
		zap.Int("input_keys", len(input)))

	// Execute the task using the task factory
	result, err := h.taskFactory.ExecuteTask(ctx, taskType, input)
	if err != nil {
		logger.Error("Task execution failed",
			zap.String("task_type", taskType),
			zap.Error(err))
		return nil, fmt.Errorf("task execution failed: %w", err)
	}

	logger.Info("Task completed successfully",
		zap.String("task_type", taskType))

	return result, nil
}

// Execute implements the TaskHandler interface
func (h *LoanProcessingTaskHandler) Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	// Extract the task type from the input
	taskType, _ := input["taskType"].(string)

	// Use the HandleTask method to execute the task
	return h.HandleTask(ctx, taskType, input)
}

// GetSupportedTaskTypes returns a list of all supported task types
func (h *LoanProcessingTaskHandler) GetSupportedTaskTypes() []string {
	return h.taskFactory.GetSupportedTaskTypes()
}
