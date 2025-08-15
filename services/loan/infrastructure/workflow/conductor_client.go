package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ConductorClientImpl implements the ConductorClient interface for Netflix Conductor
type ConductorClientImpl struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewConductorClientImpl creates a new Conductor client implementation
func NewConductorClientImpl(baseURL string, logger *zap.Logger) *ConductorClientImpl {
	return &ConductorClientImpl{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
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

	// Prepare request payload
	payload := map[string]interface{}{
		"name":    workflowName,
		"version": version,
		"input":   input,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Error("Failed to marshal workflow start payload", zap.Error(err))
		return nil, fmt.Errorf("failed to marshal workflow payload: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/workflow", c.baseURL)
	logger.Debug("Making request to Conductor API",
		zap.String("url", url),
		zap.String("method", "POST"),
		zap.Any("payload", payload))
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.Error("Failed to create HTTP request", zap.Error(err))
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to execute HTTP request", zap.Error(err))
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for debugging
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", zap.Error(err))
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log response details for debugging
	logger.Debug("Conductor API response",
		zap.Int("status_code", resp.StatusCode),
		zap.String("status", resp.Status),
		zap.String("response_body", string(responseBody)),
		zap.String("content_type", resp.Header.Get("Content-Type")),
	)

	// Check response status
	if resp.StatusCode != http.StatusOK {
		logger.Error("Workflow start failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("status", resp.Status),
			zap.String("response_body", string(responseBody)),
		)
		return nil, fmt.Errorf("workflow start failed with status: %s, body: %s", resp.Status, string(responseBody))
	}

	// Parse response - Conductor API returns plain text workflow ID for successful workflow start
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/plain") {
		// Handle plain text response (workflow ID only)
		workflowID := strings.TrimSpace(string(responseBody))
		if workflowID == "" {
			logger.Error("Empty workflow ID in response")
			return nil, fmt.Errorf("empty workflow ID in response")
		}

		execution := &WorkflowExecution{
			WorkflowID:    workflowID,
			Status:        "RUNNING",
			Input:         input,
			CorrelationID: workflowID, // Use workflow ID as correlation ID for now
			StartTime:     time.Now().UTC(),
		}

		logger.Info("Workflow started successfully with plain text response",
			zap.String("workflow_id", execution.WorkflowID),
			zap.String("correlation_id", execution.CorrelationID),
		)

		return execution, nil
	}

	// Try to parse as JSON response
	var execution WorkflowExecution
	if err := json.Unmarshal(responseBody, &execution); err != nil {
		logger.Error("Failed to decode workflow execution response",
			zap.Error(err),
			zap.String("response_body", string(responseBody)),
			zap.String("content_type", contentType),
		)
		return nil, fmt.Errorf("failed to decode response: %w, body: %s", err, string(responseBody))
	}

	logger.Info("Workflow started successfully",
		zap.String("workflow_id", execution.WorkflowID),
		zap.String("correlation_id", execution.CorrelationID),
	)

	return &execution, nil
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

	// Create HTTP request
	url := fmt.Sprintf("%s/api/workflow/%s", c.baseURL, workflowID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		logger.Error("Failed to create HTTP request", zap.Error(err))
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to execute HTTP request", zap.Error(err))
		return nil, fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == http.StatusNotFound {
		logger.Warn("Workflow not found", zap.String("workflow_id", workflowID))
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to get workflow status",
			zap.Int("status_code", resp.StatusCode),
			zap.String("status", resp.Status),
		)
		return nil, fmt.Errorf("failed to get workflow status with status: %s", resp.Status)
	}

	// Parse response
	var status WorkflowStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		logger.Error("Failed to decode workflow status response", zap.Error(err))
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	logger.Debug("Workflow status retrieved successfully",
		zap.String("status", status.Status),
		zap.Int("task_count", len(status.Tasks)),
	)

	return &status, nil
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

	// Prepare request payload
	payload := map[string]string{
		"reason": reason,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Error("Failed to marshal terminate payload", zap.Error(err))
		return fmt.Errorf("failed to marshal terminate payload: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s/api/workflow/%s/terminate", c.baseURL, workflowID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.Error("Failed to create HTTP request", zap.Error(err))
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to execute HTTP request", zap.Error(err))
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to terminate workflow",
			zap.Int("status_code", resp.StatusCode),
			zap.String("status", resp.Status),
		)
		return fmt.Errorf("failed to terminate workflow with status: %s", resp.Status)
	}

	logger.Info("Workflow terminated successfully")
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
		zap.String("operation", "pause_workflow"),
		zap.String("reason", reason),
	)

	// Create HTTP request
	url := fmt.Sprintf("%s/api/workflow/%s/pause", c.baseURL, workflowID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		logger.Error("Failed to create HTTP request", zap.Error(err))
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to execute HTTP request", zap.Error(err))
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to pause workflow",
			zap.Int("status_code", resp.StatusCode),
			zap.String("status", resp.Status),
		)
		return fmt.Errorf("failed to pause workflow with status: %s", resp.Status)
	}

	logger.Info("Workflow paused successfully")
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

	// Create HTTP request
	url := fmt.Sprintf("%s/api/workflow/%s/resume", c.baseURL, workflowID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		logger.Error("Failed to create HTTP request", zap.Error(err))
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to execute HTTP request", zap.Error(err))
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to resume workflow",
			zap.Int("status_code", resp.StatusCode),
			zap.String("status", resp.Status),
		)
		return fmt.Errorf("failed to resume workflow with status: %s", resp.Status)
	}

	logger.Info("Workflow resumed successfully")
	return nil
}

// UpdateTask updates a task with status and output
func (c *ConductorClientImpl) UpdateTask(
	ctx context.Context,
	taskID string,
	status string,
	output map[string]interface{},
) error {
	logger := c.logger.With(
		zap.String("task_id", taskID),
		zap.String("status", status),
		zap.String("operation", "update_task"),
	)

	logger.Info("Updating task with output",
		zap.Any("output", output),
		zap.Int("output_keys", len(output)),
		zap.Bool("output_is_nil", output == nil))

	// Validate output is not nil
	if output == nil {
		logger.Warn("Output is nil, creating default output")
		output = map[string]interface{}{
			"error": "Task output was nil",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		}
	}

	// Prepare request payload
	payload := map[string]interface{}{
		"status": status,
		"output": output,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Error("Failed to marshal task update payload", zap.Error(err))
		return fmt.Errorf("failed to marshal task update payload: %w", err)
	}

	logger.Debug("Task update payload prepared",
		zap.String("payload", string(payloadBytes)))

	// Create HTTP request
	url := fmt.Sprintf("%s/api/tasks/%s", c.baseURL, taskID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		logger.Error("Failed to create HTTP request", zap.Error(err))
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Failed to execute HTTP request", zap.Error(err))
		return fmt.Errorf("failed to execute HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body for debugging
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", zap.Error(err))
		return fmt.Errorf("failed to read response body: %w", err)
	}

	logger.Debug("Conductor API response",
		zap.Int("status_code", resp.StatusCode),
		zap.String("status", resp.Status),
		zap.String("response_body", string(responseBody)))

	// Check response status
	if resp.StatusCode != http.StatusOK {
		logger.Error("Failed to update task",
			zap.Int("status_code", resp.StatusCode),
			zap.String("status", resp.Status),
			zap.String("response_body", string(responseBody)))
		return fmt.Errorf("failed to update task with status: %s", resp.Status)
	}

	logger.Info("Task updated successfully")
	return nil
}

// GetBaseURL returns the base URL of the Conductor service
func (c *ConductorClientImpl) GetBaseURL() string {
	return c.baseURL
}
