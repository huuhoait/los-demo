package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/conductor-sdk/conductor-go/sdk/client"
	"github.com/conductor-sdk/conductor-go/sdk/settings"
	"go.uber.org/zap"
)

// ConductorClientImpl implements the ConductorClient interface for Netflix Conductor
type ConductorClientImpl struct {
	workflowClient client.WorkflowClient
	taskClient     client.TaskClient
	logger         *zap.Logger
	baseURL        string
}

// NewConductorClientImpl creates a new Conductor client implementation
func NewConductorClientImpl(baseURL string, logger *zap.Logger) *ConductorClientImpl {
	// Create authentication settings (no auth for local development)
	authSettings := settings.NewAuthenticationSettings("", "")

	// Create HTTP settings
	httpSettings := settings.NewHttpSettings(baseURL)

	// Create the API client
	apiClient := client.NewAPIClient(authSettings, httpSettings)

	// Create workflow and task clients
	workflowClient := client.NewWorkflowClient(apiClient)
	taskClient := client.NewTaskClient(apiClient)

	return &ConductorClientImpl{
		workflowClient: workflowClient,
		taskClient:     taskClient,
		logger:         logger,
		baseURL:        baseURL,
	}
}

// StartWorkflow starts a new workflow execution
func (c *ConductorClientImpl) StartWorkflow(
	ctx context.Context,
	workflowName string,
	version int,
	input map[string]interface{},
) (*WorkflowExecution, error) {
	logger := c.logger.With(
		zap.String("workflow_name", workflowName),
		zap.Int("version", version),
		zap.String("operation", "start_workflow"),
	)

	logger.Debug("Starting workflow via HTTP API",
		zap.String("workflow_name", workflowName),
		zap.Int("version", version),
		zap.Any("input", input))

	// Create the workflow start request
	startRequest := map[string]interface{}{
		"name":    workflowName,
		"version": version,
		"input":   input,
	}

	// Marshal the request
	jsonData, err := json.Marshal(startRequest)
	if err != nil {
		logger.Error("Failed to marshal workflow start request", zap.Error(err))
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/workflow", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Failed to create HTTP request", zap.Error(err))
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/plain")

	// Execute the request
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to execute workflow start request", zap.Error(err))
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", zap.Error(err))
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	workflowId := string(responseBody)

	// Check if the request was successful
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logger.Error("Workflow start failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", workflowId))
		return nil, fmt.Errorf("workflow start failed with status %d: %s", resp.StatusCode, workflowId)
	}

	logger.Info("Workflow started successfully via HTTP API",
		zap.String("workflow_id", workflowId))

	// Create the execution response
	execution := &WorkflowExecution{
		WorkflowID:    workflowId,
		Status:        "RUNNING",
		Input:         input,
		CorrelationID: workflowId, // Use workflow ID as correlation ID
		StartTime:     time.Now(),
	}

	return execution, nil
}

// GetWorkflowStatus retrieves the status of a workflow execution
func (c *ConductorClientImpl) GetWorkflowStatus(
	ctx context.Context,
	workflowID string,
) (*WorkflowStatus, error) {
	logger := c.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("operation", "get_workflow_status"),
	)

	// Get workflow execution using the SDK
	execution, _, err := c.workflowClient.GetExecutionStatus(ctx, workflowID, nil)
	if err != nil {
		logger.Error("Failed to get workflow status", zap.Error(err))
		return nil, fmt.Errorf("failed to get workflow status: %w", err)
	}

	// Convert SDK response to our format
	status := &WorkflowStatus{
		WorkflowID: execution.WorkflowId,
		Status:     string(execution.Status),
		Input:      execution.Input,
		Output:     execution.Output,
		Tasks:      make([]TaskStatus, 0, len(execution.Tasks)),
	}

	// Convert tasks
	for _, task := range execution.Tasks {
		taskStatus := TaskStatus{
			TaskID:            task.TaskId,
			TaskType:          task.TaskType,
			Status:            string(task.Status),
			ReferenceTaskName: task.ReferenceTaskName,
			Input:             task.InputData,
			Output:            task.OutputData,
		}

		// Handle start time
		if task.StartTime > 0 {
			taskStatus.StartTime = time.Unix(task.StartTime/1000, 0)
		}

		// Handle end time
		if task.EndTime > 0 {
			endTimeParsed := time.Unix(task.EndTime/1000, 0)
			taskStatus.EndTime = &endTimeParsed
		}

		status.Tasks = append(status.Tasks, taskStatus)
	}

	logger.Debug("Retrieved workflow status",
		zap.String("status", status.Status),
		zap.Int("task_count", len(status.Tasks)))

	return status, nil
}

// TerminateWorkflow terminates a running workflow
func (c *ConductorClientImpl) TerminateWorkflow(
	ctx context.Context,
	workflowID string,
	reason string,
) error {
	logger := c.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("reason", reason),
		zap.String("operation", "terminate_workflow"),
	)

	// Terminate workflow using the SDK
	_, err := c.workflowClient.Terminate(ctx, workflowID, nil)
	if err != nil {
		logger.Error("Failed to terminate workflow", zap.Error(err))
		return fmt.Errorf("failed to terminate workflow: %w", err)
	}

	logger.Debug("Workflow terminated successfully")
	return nil
}

// PauseWorkflow pauses a running workflow
func (c *ConductorClientImpl) PauseWorkflow(
	ctx context.Context,
	workflowID string,
	reason string,
) error {
	logger := c.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("reason", reason),
		zap.String("operation", "pause_workflow"),
	)

	// Pause workflow using the SDK
	_, err := c.workflowClient.PauseWorkflow(ctx, workflowID)
	if err != nil {
		logger.Error("Failed to pause workflow", zap.Error(err))
		return fmt.Errorf("failed to pause workflow: %w", err)
	}

	logger.Debug("Workflow paused successfully")
	return nil
}

// ResumeWorkflow resumes a paused workflow
func (c *ConductorClientImpl) ResumeWorkflow(
	ctx context.Context,
	workflowID string,
) error {
	logger := c.logger.With(
		zap.String("workflow_id", workflowID),
		zap.String("operation", "resume_workflow"),
	)

	// Resume workflow using the SDK
	_, err := c.workflowClient.ResumeWorkflow(ctx, workflowID)
	if err != nil {
		logger.Error("Failed to resume workflow", zap.Error(err))
		return fmt.Errorf("failed to resume workflow: %w", err)
	}

	logger.Debug("Workflow resumed successfully")
	return nil
}

// UpdateTask updates a task with status and output
func (c *ConductorClientImpl) UpdateTask(
	ctx context.Context,
	taskID string,
	workflowInstanceId string,
	referenceTaskName string,
	status string,
	output map[string]interface{},
) error {
	logger := c.logger.With(
		zap.String("task_id", taskID),
		zap.String("workflow_instance_id", workflowInstanceId),
		zap.String("reference_task_name", referenceTaskName),
		zap.String("status", status),
		zap.String("operation", "update_task"),
	)

	logger.Debug("Updating task with output",
		zap.Int("output_keys", len(output)),
		zap.Bool("output_is_nil", output == nil))

	// Validate output is not nil
	if output == nil {
		logger.Warn("Output is nil, creating default output")
		output = map[string]interface{}{
			"error":     "Task output was nil",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
	}

	// Use the correct Conductor API endpoint format
	logger.Debug("Using correct Conductor API endpoint format",
		zap.String("task_id", taskID),
		zap.String("reference_task_name", referenceTaskName),
		zap.String("workflow_instance_id", workflowInstanceId),
		zap.String("status", status))

	// Create the correct request payload
	requestBody := map[string]interface{}{
		"taskId":             taskID,
		"referenceTaskName":  referenceTaskName,
		"workflowInstanceId": workflowInstanceId,
		"status":             status,
		"outputData":         output,
	}

	// Marshal request body
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		logger.Error("Failed to marshal request body", zap.Error(err))
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP client
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/tasks", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		logger.Error("Failed to create HTTP request", zap.Error(err))
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to execute HTTP request", zap.Error(err))
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", zap.Error(err))
		return fmt.Errorf("failed to read response body: %w", err)
	}

	logger.Debug("Conductor API response",
		zap.String("url", url),
		zap.Int("status_code", resp.StatusCode),
		zap.String("response", string(responseBody)))

	// Check if successful
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logger.Debug("Task updated successfully via Conductor API")
		return nil
	}

	logger.Error("Task update failed",
		zap.String("url", url),
		zap.Int("status_code", resp.StatusCode),
		zap.String("response", string(responseBody)))
	return fmt.Errorf("task update failed with status %d: %s", resp.StatusCode, string(responseBody))
}

// GetBaseURL returns the base URL of the Conductor service
func (c *ConductorClientImpl) GetBaseURL() string {
	return c.baseURL
}
