package workflow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"loan-service/pkg/i18n"
)

// TaskWorker polls Netflix Conductor for tasks and executes them
type TaskWorker struct {
	conductorClient ConductorClient
	taskHandlers    map[string]TaskHandler
	logger          *zap.Logger
	localizer       *i18n.Localizer
	workerID        string
	pollInterval    time.Duration
}

// TaskHandler interface for executing workflow tasks
type TaskHandler interface {
	Execute(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)
}

// Task represents a task from Netflix Conductor
type Task struct {
	TaskID             string                 `json:"taskId"`
	TaskType           string                 `json:"taskType"`
	Status             string                 `json:"status"`
	ReferenceTaskName  string                 `json:"referenceTaskName"`
	Input              map[string]interface{} `json:"inputData"`
	Output             map[string]interface{} `json:"outputData"`
	StartTime          time.Time              `json:"startTime"`
	EndTime            *time.Time             `json:"endTime,omitempty"`
	WorkflowInstanceId string                 `json:"workflowInstanceId"`
}

// NewTaskWorker creates a new task worker
func NewTaskWorker(
	conductorClient ConductorClient,
	logger *zap.Logger,
	localizer *i18n.Localizer,
) *TaskWorker {
	worker := &TaskWorker{
		conductorClient: conductorClient,
		taskHandlers:    make(map[string]TaskHandler),
		logger:          logger,
		localizer:       localizer,
		workerID:        fmt.Sprintf("worker_%d", time.Now().UnixNano()),
		pollInterval:    5 * time.Second,
	}

	// Register task handlers
	worker.registerTaskHandlers()

	return worker
}

// Start starts the task worker
func (w *TaskWorker) Start(ctx context.Context) error {
	logger := w.logger.With(
		zap.String("worker_id", w.workerID),
		zap.String("operation", "start_task_worker"),
	)

	logger.Info("Starting task worker")

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Task worker stopped due to context cancellation")
			return ctx.Err()

		case <-ticker.C:
			if err := w.pollAndExecuteTasks(ctx); err != nil {
				logger.Error("Error polling and executing tasks", zap.Error(err))
				// Continue running, don't stop the worker
			}
		}
	}
}

// pollAndExecuteTasks polls for available tasks and executes them
func (w *TaskWorker) pollAndExecuteTasks(ctx context.Context) error {
	logger := w.logger.With(
		zap.String("worker_id", w.workerID),
		zap.String("operation", "poll_and_execute_tasks"),
	)

	// Poll for available tasks
	tasks, err := w.pollForTasks(ctx)
	if err != nil {
		logger.Error("Failed to poll for tasks", zap.Error(err))
		return err
	}

	if len(tasks) == 0 {
		logger.Debug("No tasks available")
		return nil
	}

	logger.Info("Found tasks to execute", zap.Int("task_count", len(tasks)))

	// Execute each task
	for _, task := range tasks {
		if err := w.executeTask(ctx, task); err != nil {
			logger.Error("Failed to execute task",
				zap.String("task_id", task.TaskID),
				zap.String("task_type", task.TaskType),
				zap.Error(err),
			)
			// Mark task as failed
			if err := w.updateTaskStatus(ctx, task.TaskID, "FAILED", map[string]interface{}{
				"error": err.Error(),
			}); err != nil {
				logger.Error("Failed to update task status to failed", zap.Error(err))
			}
		}
	}

	return nil
}

// pollForTasks polls Netflix Conductor for available tasks
func (w *TaskWorker) pollForTasks(ctx context.Context) ([]Task, error) {
	logger := w.logger.With(zap.String("operation", "poll_for_tasks"))

	// Create HTTP request to poll for tasks
	url := fmt.Sprintf("%s/api/tasks/poll", w.getConductorBaseURL())
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create poll request: %w", err)
	}

	// Add query parameters for task types we can handle
	q := req.URL.Query()
	q.Add("taskType", "SIMPLE")
	q.Add("workerid", w.workerID)
	q.Add("timeout", "30")
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Accept", "application/json")

	// Execute request
	client := &http.Client{Timeout: 35 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute poll request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == http.StatusNoContent {
		// No tasks available
		return []Task{}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("poll request failed with status: %s", resp.Status)
	}

	// Parse response
	var tasks []Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks response: %w", err)
	}

	logger.Debug("Polled for tasks", zap.Int("task_count", len(tasks)))
	return tasks, nil
}

// executeTask executes a single task
func (w *TaskWorker) executeTask(ctx context.Context, task Task) error {
	logger := w.logger.With(
		zap.String("task_id", task.TaskID),
		zap.String("task_type", task.TaskType),
		zap.String("reference_task_name", task.ReferenceTaskName),
		zap.String("operation", "execute_task"),
	)

	logger.Info("Executing task")

	// Find the appropriate task handler
	handler, exists := w.taskHandlers[task.ReferenceTaskName]
	if !exists {
		return fmt.Errorf("no handler found for task type: %s", task.ReferenceTaskName)
	}

	// Execute the task
	output, err := handler.Execute(ctx, task.Input)
	if err != nil {
		logger.Error("Task execution failed", zap.Error(err))
		return err
	}

	// Update task status to completed
	if err := w.updateTaskStatus(ctx, task.TaskID, "COMPLETED", output); err != nil {
		logger.Error("Failed to update task status to completed", zap.Error(err))
		return err
	}

	logger.Info("Task executed successfully")
	return nil
}

// updateTaskStatus updates the status and output of a task
func (w *TaskWorker) updateTaskStatus(
	ctx context.Context,
	taskID string,
	status string,
	output map[string]interface{},
) error {
	return w.conductorClient.UpdateTask(ctx, taskID, status, output)
}

// registerTaskHandlers registers all available task handlers
func (w *TaskWorker) registerTaskHandlers() {
	// Register prequalification task handlers
	w.taskHandlers["validate_prequalify_input"] = &PreQualificationTaskHandler{
		logger:    w.logger,
		localizer: w.localizer,
	}
	w.taskHandlers["calculate_dti_ratio"] = &PreQualificationTaskHandler{
		logger:    w.logger,
		localizer: w.localizer,
	}
	w.taskHandlers["assess_prequalify_risk"] = &PreQualificationTaskHandler{
		logger:    w.logger,
		localizer: w.localizer,
	}
	w.taskHandlers["generate_prequalify_terms"] = &PreQualificationTaskHandler{
		logger:    w.logger,
		localizer: w.localizer,
	}
	w.taskHandlers["finalize_prequalification"] = &PreQualificationTaskHandler{
		logger:    w.logger,
		localizer: w.localizer,
	}
	w.taskHandlers["update_application_state"] = &PreQualificationTaskHandler{
		logger:    w.logger,
		localizer: w.localizer,
	}

	w.logger.Info("Task handlers registered", zap.Int("handler_count", len(w.taskHandlers)))
}

// getConductorBaseURL extracts the base URL from the conductor client
// This is a helper method to get the base URL for polling
func (w *TaskWorker) getConductorBaseURL() string {
	// For now, return a default URL - in a real implementation,
	// you would extract this from the conductor client
	return "http://localhost:8080"
}

// SetPollInterval sets the polling interval for the worker
func (w *TaskWorker) SetPollInterval(interval time.Duration) {
	w.pollInterval = interval
}

// GetWorkerID returns the worker ID
func (w *TaskWorker) GetWorkerID() string {
	return w.workerID
}
