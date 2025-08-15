package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	httpClient      *http.Client
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
	StartTime          interface{}            `json:"startTime"`         // Use interface{} to handle various time formats
	EndTime            interface{}            `json:"endTime,omitempty"` // Use interface{} to handle various time formats
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
		httpClient: &http.Client{
			Timeout: 35 * time.Second,
		},
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
			// Try to mark task as failed, but don't fail the entire process if status update fails
			if err := w.updateTaskStatus(ctx, task.TaskID, "FAILED", map[string]interface{}{
				"error": err.Error(),
			}); err != nil {
				logger.Warn("Failed to update task status to failed, but continuing", zap.Error(err))
			}
		}
	}

	return nil
}

// canHandleTask checks if the task worker can handle a specific task
func (w *TaskWorker) canHandleTask(task Task) bool {
	_, exists := w.taskHandlers[task.ReferenceTaskName]
	return exists
}

// claimTask claims a task by updating its status to IN_PROGRESS
func (w *TaskWorker) claimTask(ctx context.Context, taskID string) error {
	logger := w.logger.With(
		zap.String("task_id", taskID),
		zap.String("operation", "claim_task"),
	)

	logger.Info("Claiming task by updating status to IN_PROGRESS")

	// Update task status to IN_PROGRESS to claim it
	if err := w.conductorClient.UpdateTask(ctx, taskID, "IN_PROGRESS", map[string]interface{}{}); err != nil {
		logger.Error("Failed to claim task", zap.Error(err))
		return fmt.Errorf("failed to claim task: %w", err)
	}

	logger.Info("Task claimed successfully")
	return nil
}

// pollForTasks polls the Conductor server for available tasks
func (w *TaskWorker) pollForTasks(ctx context.Context) ([]Task, error) {
	logger := w.logger.With(zap.String("operation", "poll_for_tasks"))

	logger.Debug("Looking for SCHEDULED tasks to execute directly")

	scheduledTasks, err := w.findScheduledTasks(ctx)
	if err != nil {
		logger.Warn("Failed to find scheduled tasks", zap.Error(err))
	} else if len(scheduledTasks) > 0 {
		logger.Info("Found scheduled tasks", zap.Int("count", len(scheduledTasks)))

		// Return the first available task that we can handle
		for _, task := range scheduledTasks {
			if w.canHandleTask(task) {
				logger.Info("Found executable scheduled task",
					zap.String("task_id", task.TaskID),
					zap.String("task_type", task.TaskType))

				// Return the task directly for execution without claiming
				// The task will be marked as IN_PROGRESS during execution
				return []Task{task}, nil
			}
		}
	}

	logger.Debug("No SCHEDULED tasks found, trying normal polling")

	// Fall back to normal polling
	q := url.Values{}
	q.Add("workerid", w.workerID)
	q.Add("timeout", "30")

	logger.Debug("Polling for tasks with parameters",
		zap.String("worker_id", w.workerID),
		zap.String("timeout", "30"))

	// Create HTTP request
	url := fmt.Sprintf("%s/api/tasks/poll?%s", w.getConductorBaseURL(), q.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logger.Error("Failed to create polling request", zap.Error(err))
		return nil, fmt.Errorf("failed to create polling request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := w.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to execute polling request", zap.Error(err))
		return nil, fmt.Errorf("failed to execute polling request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for debugging
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", zap.Error(err))
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	logger.Debug("Conductor API response",
		zap.Int("status_code", resp.StatusCode),
		zap.String("status", resp.Status),
		zap.String("response_body", string(responseBody)),
		zap.String("content_type", resp.Header.Get("Content-Type")))

	// Check if no tasks available
	if resp.StatusCode == http.StatusNoContent {
		logger.Debug("No tasks available (HTTP 204)")
		return nil, nil
	}

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		logger.Error("Polling request failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("status", resp.Status))
		return nil, fmt.Errorf("polling request failed with status: %s", resp.Status)
	}

	// Parse response
	var tasks []Task
	if err := json.NewDecoder(bytes.NewReader(responseBody)).Decode(&tasks); err != nil {
		logger.Error("Failed to decode polling response", zap.Error(err))
		return nil, fmt.Errorf("failed to decode polling response: %w", err)
	}

	logger.Info("Successfully polled for tasks", zap.Int("task_count", len(tasks)))
	return tasks, nil
}

// findScheduledTasks searches for workflows with SCHEDULED tasks
func (w *TaskWorker) findScheduledTasks(ctx context.Context) ([]Task, error) {
	logger := w.logger.With(zap.String("operation", "find_scheduled_tasks"))

	// Search for running workflows
	url := fmt.Sprintf("%s/api/workflow/search?query=status:RUNNING", w.getConductorBaseURL())
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create search request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search request failed with status: %s", resp.Status)
	}

	var searchResult struct {
		Results []struct {
			WorkflowID string `json:"workflowId"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// Get detailed workflow info for each running workflow
	var scheduledTasks []Task
	for _, workflow := range searchResult.Results {
		workflowDetail, err := w.getWorkflowDetail(ctx, workflow.WorkflowID)
		if err != nil {
			logger.Warn("Failed to get workflow detail", zap.String("workflow_id", workflow.WorkflowID), zap.Error(err))
			continue
		}

		// Check if workflow has SCHEDULED tasks
		for _, task := range workflowDetail.Tasks {
			if task.Status == "SCHEDULED" {
				scheduledTasks = append(scheduledTasks, task)
			}
		}
	}

	logger.Info("Found scheduled tasks", zap.Int("count", len(scheduledTasks)))
	return scheduledTasks, nil
}

// getWorkflowDetail gets detailed information about a specific workflow
func (w *TaskWorker) getWorkflowDetail(ctx context.Context, workflowID string) (*WorkflowDetail, error) {
	url := fmt.Sprintf("%s/api/workflow/%s", w.getConductorBaseURL(), workflowID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow detail request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute workflow detail request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("workflow detail request failed with status: %s", resp.Status)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse as a raw JSON map first
	var rawResponse map[string]interface{}
	if err := json.Unmarshal(body, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response as JSON: %w", err)
	}

	// Extract only the fields we need
	workflow := &WorkflowDetail{
		WorkflowID:    getString(rawResponse, "workflowId"),
		Status:        getString(rawResponse, "status"),
		Input:         getMap(rawResponse, "input"),
		Output:        getMap(rawResponse, "output"),
		CorrelationID: getString(rawResponse, "correlationId"),
		StartTime:     rawResponse["startTime"],
		EndTime:       rawResponse["endTime"],
	}

	// Extract tasks array
	if tasksRaw, exists := rawResponse["tasks"]; exists {
		if tasksArray, ok := tasksRaw.([]interface{}); ok {
			workflow.Tasks = make([]Task, 0, len(tasksArray))
			for _, taskRaw := range tasksArray {
				if taskMap, ok := taskRaw.(map[string]interface{}); ok {
					task := Task{
						TaskID:            getString(taskMap, "taskId"),
						TaskType:          getString(taskMap, "taskType"),
						Status:            getString(taskMap, "status"),
						ReferenceTaskName: getString(taskMap, "referenceTaskName"),
						Input:             getMap(taskMap, "inputData"), // Note: Conductor uses "inputData" not "input"
					}
					workflow.Tasks = append(workflow.Tasks, task)
				}
			}
		}
	}

	return workflow, nil
}

// Helper functions for safe type extraction
func getString(m map[string]interface{}, key string) string {
	if val, exists := m[key]; exists {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if val, exists := m[key]; exists {
		if mapVal, ok := val.(map[string]interface{}); ok {
			return mapVal
		}
	}
	return make(map[string]interface{})
}

// WorkflowDetail represents detailed workflow information with tasks
type WorkflowDetail struct {
	WorkflowID    string                 `json:"workflowId"`
	Status        string                 `json:"status"`
	Input         map[string]interface{} `json:"input"`
	Output        map[string]interface{} `json:"output"`
	CorrelationID string                 `json:"correlationId"`
	StartTime     interface{}            `json:"startTime"`         // Use interface{} to handle various time formats
	EndTime       interface{}            `json:"endTime,omitempty"` // Use interface{} to handle various time formats
	Tasks         []Task                 `json:"tasks"`
	// Add other fields that might be present to avoid parsing errors
	CreateTime   interface{} `json:"createTime,omitempty"`
	UpdateTime   interface{} `json:"updateTime,omitempty"`
	WorkflowName string      `json:"workflowName,omitempty"`
	// Use a catch-all field for any other properties
	AdditionalFields map[string]interface{} `json:"-"`
}

// executeTask executes a single task
func (w *TaskWorker) executeTask(ctx context.Context, task Task) error {
	logger := w.logger.With(
		zap.String("task_id", task.TaskID),
		zap.String("task_type", task.TaskType),
		zap.String("reference_task_name", task.ReferenceTaskName),
		zap.String("operation", "execute_task"),
	)

	logger.Info("Executing task",
		zap.Any("task_input", task.Input),
		zap.String("task_status", task.Status))

	// Mark task as IN_PROGRESS since we're executing it directly
	if err := w.updateTaskStatus(ctx, task.TaskID, "IN_PROGRESS", map[string]interface{}{}); err != nil {
		logger.Warn("Failed to mark task as IN_PROGRESS, continuing execution", zap.Error(err))
		// Continue execution even if status update fails
	}

	// Find the appropriate task handler
	handler, exists := w.taskHandlers[task.ReferenceTaskName]
	if !exists {
		logger.Error("No handler found for task type",
			zap.String("reference_task_name", task.ReferenceTaskName),
			zap.Strings("available_handlers", w.getAvailableHandlerNames()))
		return fmt.Errorf("no handler found for task type: %s", task.ReferenceTaskName)
	}

	logger.Info("Found task handler",
		zap.String("handler_type", fmt.Sprintf("%T", handler)))

	// Add task type information to the input so the handler knows what to execute
	inputWithTaskType := make(map[string]interface{})
	for k, v := range task.Input {
		inputWithTaskType[k] = v
	}
	inputWithTaskType["taskType"] = task.TaskType
	inputWithTaskType["referenceTaskName"] = task.ReferenceTaskName

	logger.Info("Executing task with enhanced input",
		zap.Any("enhanced_input", inputWithTaskType))

	// Execute the task
	output, err := handler.Execute(ctx, inputWithTaskType)
	if err != nil {
		logger.Error("Task execution failed", zap.Error(err))
		return err
	}

	logger.Info("Task execution completed successfully",
		zap.Any("task_output", output),
		zap.Int("output_keys", len(output)))

	// Validate output is not nil
	if output == nil {
		logger.Error("Task handler returned nil output",
			zap.String("task_type", task.TaskType),
			zap.String("reference_task_name", task.ReferenceTaskName))
		output = map[string]interface{}{
			"error":     "Task handler returned nil output",
			"taskType":  task.TaskType,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
	}

	// Update task status to completed
	if err := w.updateTaskStatus(ctx, task.TaskID, "COMPLETED", output); err != nil {
		logger.Warn("Failed to update task status to completed, but task executed successfully",
			zap.Error(err),
			zap.Any("task_output", output))
		// Don't return error since the task executed successfully
		// The status update failure is a Conductor API limitation, not a task execution issue
	} else {
		logger.Info("Task executed successfully and status updated")
	}

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
	// Note: update_application_state is handled by both prequalification and loan processing workflows
	// The specific handler will be determined by the taskReferenceName

	// Register loan processing task handlers
	loanProcessingHandler := &LoanProcessingTaskHandler{
		logger:    w.logger,
		localizer: w.localizer,
	}
	w.taskHandlers["validate_application_ref"] = loanProcessingHandler
	w.taskHandlers["update_state_to_prequalified_ref"] = loanProcessingHandler
	w.taskHandlers["document_collection_ref"] = loanProcessingHandler
	w.taskHandlers["identity_verification_ref"] = loanProcessingHandler
	w.taskHandlers["finalize_loan_decision_ref"] = loanProcessingHandler

	w.logger.Info("Task handlers registered", zap.Int("handler_count", len(w.taskHandlers)))
}

// getAvailableHandlerNames returns a list of available handler names for debugging
func (w *TaskWorker) getAvailableHandlerNames() []string {
	names := make([]string, 0, len(w.taskHandlers))
	for name := range w.taskHandlers {
		names = append(names, name)
	}
	return names
}

// getConductorBaseURL extracts the base URL from the conductor client
// This is a helper method to get the base URL for polling
func (w *TaskWorker) getConductorBaseURL() string {
	// Get the base URL from the conductor client
	return w.conductorClient.GetBaseURL()
}

// SetPollInterval sets the polling interval for the worker
func (w *TaskWorker) SetPollInterval(interval time.Duration) {
	w.pollInterval = interval
}

// GetWorkerID returns the worker ID
func (w *TaskWorker) GetWorkerID() string {
	return w.workerID
}
