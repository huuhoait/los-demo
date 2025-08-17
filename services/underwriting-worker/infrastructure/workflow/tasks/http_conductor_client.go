package tasks

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

	"underwriting_worker/pkg/config"
)

// HTTPConductorClient implements a simple HTTP client for Conductor
type HTTPConductorClient struct {
	logger     *zap.Logger
	config     *config.Config
	httpClient *http.Client
	baseURL    string
	workers    map[string]TaskHandler
	isRunning  bool
	stopChan   chan struct{}
}

// WorkflowDefinition represents a Conductor workflow definition
type WorkflowDefinition struct {
	Name             string                 `json:"name"`
	Description      string                 `json:"description"`
	Version          int                    `json:"version"`
	Tasks            []WorkflowTask         `json:"tasks"`
	InputParameters  []string               `json:"inputParameters,omitempty"`
	OutputParameters map[string]interface{} `json:"outputParameters,omitempty"`
	SchemaVersion    int                    `json:"schemaVersion"`
}

// WorkflowTask represents a task in a workflow
type WorkflowTask struct {
	Name              string                 `json:"name"`
	TaskReferenceName string                 `json:"taskReferenceName"`
	Type              string                 `json:"type"`
	InputParameters   map[string]interface{} `json:"inputParameters,omitempty"`
}

// TaskDefinition represents a Conductor task definition
type TaskDefinition struct {
	Name                   string   `json:"name"`
	Description            string   `json:"description"`
	TimeoutSeconds         int      `json:"timeoutSeconds"`
	ResponseTimeoutSeconds int      `json:"responseTimeoutSeconds"`
	RetryCount             int      `json:"retryCount"`
	InputKeys              []string `json:"inputKeys,omitempty"`
	OutputKeys             []string `json:"outputKeys,omitempty"`
}

// ConductorTask represents a task from Conductor
type ConductorTask struct {
	TaskID             string                 `json:"taskId"`
	TaskType           string                 `json:"taskType"`
	WorkflowInstanceID string                 `json:"workflowInstanceId"`
	InputData          map[string]interface{} `json:"inputData"`
	Status             string                 `json:"status"`
}

// ConductorTaskResult represents a task result for Conductor
type ConductorTaskResult struct {
	TaskID                string                 `json:"taskId"`
	ReferenceTaskName     string                 `json:"referenceTaskName"`
	WorkflowInstanceID    string                 `json:"workflowInstanceId"`
	Status                string                 `json:"status"`
	OutputData            map[string]interface{} `json:"outputData"`
	ReasonForIncompletion string                 `json:"reasonForIncompletion,omitempty"`
	WorkerID              string                 `json:"workerId"`
}

// NewHTTPConductorClient creates a new HTTP-based Conductor client
func NewHTTPConductorClient(logger *zap.Logger, cfg *config.Config) (*HTTPConductorClient, error) {
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Parse and validate the base URL
	baseURL := cfg.Conductor.ServerURL
	if _, err := url.Parse(baseURL); err != nil {
		return nil, fmt.Errorf("invalid conductor URL: %w", err)
	}

	client := &HTTPConductorClient{
		logger:     logger,
		config:     cfg,
		httpClient: httpClient,
		baseURL:    baseURL,
		workers:    make(map[string]TaskHandler),
		isRunning:  false,
		stopChan:   make(chan struct{}),
	}

	return client, nil
}

// RegisterWorker registers a worker for a specific task type
func (c *HTTPConductorClient) RegisterWorker(taskType string, handler TaskHandler) {
	c.workers[taskType] = handler
	c.logger.Info("Registered worker for task type with HTTP Conductor client",
		zap.String("task_type", taskType),
		zap.String("conductor_url", c.baseURL))
}

// StartPolling starts polling for tasks from Conductor
func (c *HTTPConductorClient) StartPolling() error {
	if c.isRunning {
		return fmt.Errorf("HTTP conductor client already running")
	}

	c.logger.Info("Starting HTTP Conductor client",
		zap.String("conductor_url", c.baseURL),
		zap.Int("worker_pool_size", c.config.Conductor.WorkerPoolSize),
		zap.Int("polling_interval_ms", c.config.Conductor.PollingInterval))

	// Test connection to Conductor
	if err := c.testConnection(); err != nil {
		c.logger.Error("Failed to connect to Conductor", zap.Error(err))
		return fmt.Errorf("failed to connect to Conductor: %w", err)
	}

	c.isRunning = true

	// Start polling workers
	for i := 0; i < c.config.Conductor.WorkerPoolSize; i++ {
		go c.pollingWorker(fmt.Sprintf("worker-%d", i))
	}

	c.logger.Info("HTTP Conductor client started successfully")
	return nil
}

// StopPolling stops polling for tasks
func (c *HTTPConductorClient) StopPolling() {
	if !c.isRunning {
		return
	}

	c.logger.Info("Stopping HTTP Conductor client")
	c.isRunning = false
	close(c.stopChan)
	c.logger.Info("HTTP Conductor client stopped")
}

// testConnection tests the connection to Conductor server
func (c *HTTPConductorClient) testConnection() error {
	healthURL := fmt.Sprintf("%s/health", c.baseURL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Conductor health endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("conductor health check failed with status: %d", resp.StatusCode)
	}

	c.logger.Info("Successfully connected to Conductor server",
		zap.String("conductor_url", c.baseURL))

	return nil
}

// pollingWorker polls for tasks and executes them
func (c *HTTPConductorClient) pollingWorker(workerID string) {
	logger := c.logger.With(zap.String("worker_id", workerID))
	pollInterval := time.Duration(c.config.Conductor.PollingInterval) * time.Millisecond

	for {
		select {
		case <-c.stopChan:
			logger.Info("Polling worker stopped")
			return
		default:
			// Poll for tasks
			for taskType := range c.workers {
				if !c.isRunning {
					return
				}

				task, err := c.pollTask(taskType, workerID)
				if err != nil {
					logger.Debug("Failed to poll task",
						zap.String("task_type", taskType),
						zap.Error(err))
					continue
				}

				if task != nil {
					c.executeTask(task, workerID, logger)
				}
			}

			// Wait before next poll
			time.Sleep(pollInterval)
		}
	}
}

// pollTask polls for a specific task type
func (c *HTTPConductorClient) pollTask(taskType, workerID string) (*ConductorTask, error) {
	pollURL := fmt.Sprintf("%s/api/tasks/poll/%s?workerid=%s", c.baseURL, taskType, workerID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", pollURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create poll request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to poll task: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent || resp.StatusCode == 204 {
		// No tasks available
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Debug("Poll task failed",
			zap.String("task_type", taskType),
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)))
		return nil, fmt.Errorf("poll task failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read poll response: %w", err)
	}

	// Handle empty response body
	if len(body) == 0 {
		return nil, nil
	}

	var task ConductorTask
	if err := json.Unmarshal(body, &task); err != nil {
		c.logger.Error("Failed to unmarshal task",
			zap.String("task_type", taskType),
			zap.String("response_body", string(body)),
			zap.Error(err))
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	// Validate task has required fields
	if task.TaskID == "" || task.TaskType == "" {
		c.logger.Warn("Received invalid task",
			zap.String("task_id", task.TaskID),
			zap.String("task_type", task.TaskType))
		return nil, nil
	}

	return &task, nil
}

// executeTask executes a task
func (c *HTTPConductorClient) executeTask(task *ConductorTask, workerID string, logger *zap.Logger) {
	startTime := time.Now()

	logger.Info("Executing task",
		zap.String("task_id", task.TaskID),
		zap.String("task_type", task.TaskType))

	handler, exists := c.workers[task.TaskType]
	if !exists {
		logger.Error("No handler registered for task type", zap.String("task_type", task.TaskType))
		c.updateTaskResult(&ConductorTaskResult{
			TaskID:                task.TaskID,
			Status:                "FAILED",
			ReasonForIncompletion: "No handler registered",
			WorkerID:              workerID,
		})
		return
	}

	// Convert to our internal format
	mockTask := &MockTask{
		TaskID:             task.TaskID,
		TaskType:           task.TaskType,
		WorkflowInstanceID: task.WorkflowInstanceID,
		InputData:          task.InputData,
		Status:             task.Status,
		CreatedTime:        time.Now(),
		UpdatedTime:        time.Now(),
	}

	// Execute the handler with recovery
	var result *MockTaskResult
	var handlerErr error

	func() {
		defer func() {
			if r := recover(); r != nil {
				handlerErr = fmt.Errorf("task handler panicked: %v", r)
				logger.Error("Task handler panic recovered",
					zap.String("task_id", task.TaskID),
					zap.Any("panic", r))
			}
		}()
		result, handlerErr = handler(mockTask)
	}()

	processingTime := time.Since(startTime)

	// Convert result back to Conductor format
	conductorResult := &ConductorTaskResult{
		TaskID:             task.TaskID,
		ReferenceTaskName:  task.TaskType, // Use task type as reference task name
		WorkflowInstanceID: task.WorkflowInstanceID,
		WorkerID:           workerID,
	}

	// Ensure we always have valid output data
	if conductorResult.OutputData == nil {
		conductorResult.OutputData = make(map[string]interface{})
	}

	if handlerErr != nil {
		logger.Error("Task execution failed",
			zap.String("task_id", task.TaskID),
			zap.Error(handlerErr),
			zap.Duration("processing_time", processingTime))

		conductorResult.Status = "FAILED"
		conductorResult.ReasonForIncompletion = handlerErr.Error()
		conductorResult.OutputData = map[string]interface{}{
			"error":           handlerErr.Error(),
			"processing_time": processingTime.String(),
			"timestamp":       time.Now().UTC().Format(time.RFC3339),
		}
	} else if result == nil {
		logger.Error("Task handler returned nil result",
			zap.String("task_id", task.TaskID),
			zap.Duration("processing_time", processingTime))

		conductorResult.Status = "FAILED"
		conductorResult.ReasonForIncompletion = "Task handler returned nil result"
		conductorResult.OutputData = map[string]interface{}{
			"error":           "Handler returned nil result",
			"processing_time": processingTime.String(),
			"timestamp":       time.Now().UTC().Format(time.RFC3339),
		}
	} else {
		logger.Info("Task execution completed",
			zap.String("task_id", task.TaskID),
			zap.String("status", result.Status),
			zap.Duration("processing_time", processingTime))

		// Validate result status
		status := result.Status
		if status == "" {
			status = "COMPLETED"
		}

		conductorResult.Status = status

		// Ensure output data is not nil
		if result.OutputData != nil {
			conductorResult.OutputData = result.OutputData
		}

		// Add processing metadata
		conductorResult.OutputData["processing_time"] = processingTime.String()
		conductorResult.OutputData["timestamp"] = time.Now().UTC().Format(time.RFC3339)

		// Only set reason for incompletion if provided and status is FAILED
		if result.ReasonForIncompletion != "" && (status == "FAILED" || status == "TIMED_OUT") {
			conductorResult.ReasonForIncompletion = result.ReasonForIncompletion
		}
	}

	// Update task result in Conductor
	if err := c.updateTaskResult(conductorResult); err != nil {
		logger.Error("Failed to update task result", zap.Error(err))
	}
}

// updateTaskResult updates the task result in Conductor
func (c *HTTPConductorClient) updateTaskResult(result *ConductorTaskResult) error {
	updateURL := fmt.Sprintf("%s/api/tasks", c.baseURL)

	// Log the task result being sent
	c.logger.Debug("Sending task result to Conductor",
		zap.String("task_id", result.TaskID),
		zap.String("status", result.Status),
		zap.String("worker_id", result.WorkerID))

	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal task result: %w", err)
	}

	// Log the JSON payload for debugging
	c.logger.Debug("Task result JSON payload",
		zap.String("task_id", result.TaskID),
		zap.String("json", string(jsonData)))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", updateURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create update request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update task result: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Warn("Failed to read response body", zap.Error(err))
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		c.logger.Error("Update task result failed",
			zap.String("task_id", result.TaskID),
			zap.Int("status_code", resp.StatusCode),
			zap.String("response_body", string(body)))
		return fmt.Errorf("update task result failed with status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.Debug("Task result updated successfully",
		zap.String("task_id", result.TaskID),
		zap.Int("status_code", resp.StatusCode))

	return nil
}

// StartWorkflow starts a workflow execution
func (c *HTTPConductorClient) StartWorkflow(workflowName string, input map[string]interface{}) (string, error) {
	startURL := fmt.Sprintf("%s/api/workflow/%s", c.baseURL, workflowName)

	jsonData, err := json.Marshal(input)
	if err != nil {
		return "", fmt.Errorf("failed to marshal workflow input: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", startURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create start workflow request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to start workflow: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read start workflow response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("start workflow failed with status %d: %s", resp.StatusCode, string(body))
	}

	// The response should be the workflow ID
	workflowID := string(bytes.Trim(body, "\""))

	c.logger.Info("Started workflow",
		zap.String("workflow_id", workflowID),
		zap.String("workflow_name", workflowName))

	return workflowID, nil
}

// GetWorkflowStatus gets the status of a workflow
func (c *HTTPConductorClient) GetWorkflowStatus(workflowID string) (map[string]interface{}, error) {
	statusURL := fmt.Sprintf("%s/api/workflow/%s", c.baseURL, workflowID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get workflow status request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow status response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get workflow status failed with status %d: %s", resp.StatusCode, string(body))
	}

	var workflow map[string]interface{}
	if err := json.Unmarshal(body, &workflow); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow status: %w", err)
	}

	return workflow, nil
}

// RegisterWorkflowDefinition registers a workflow definition with Conductor
func (c *HTTPConductorClient) RegisterWorkflowDefinition(workflow *WorkflowDefinition) error {
	registerURL := fmt.Sprintf("%s/api/metadata/workflow", c.baseURL)

	jsonData, err := json.Marshal(workflow)
	if err != nil {
		return fmt.Errorf("failed to marshal workflow definition: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", registerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create register workflow request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to register workflow definition: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Accept 200, 201, or 409 (conflict) as success
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		return fmt.Errorf("register workflow definition failed with status %d: %s", resp.StatusCode, string(body))
	}

	if resp.StatusCode == http.StatusConflict {
		c.logger.Info("Workflow definition already exists",
			zap.String("workflow_name", workflow.Name),
			zap.Int("workflow_version", workflow.Version))
	} else {
		c.logger.Info("Registered workflow definition",
			zap.String("workflow_name", workflow.Name),
			zap.Int("workflow_version", workflow.Version))
	}

	return nil
}

// RegisterTaskDefinition registers a task definition with Conductor
func (c *HTTPConductorClient) RegisterTaskDefinition(task *TaskDefinition) error {
	registerURL := fmt.Sprintf("%s/api/metadata/taskdefs", c.baseURL)

	// Conductor expects an array of task definitions
	taskDefs := []*TaskDefinition{task}

	jsonData, err := json.Marshal(taskDefs)
	if err != nil {
		return fmt.Errorf("failed to marshal task definition: %w", err)
	}

	c.logger.Debug("Registering task definition",
		zap.String("task_name", task.Name),
		zap.String("json", string(jsonData)))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", registerURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create register task request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to register task definition: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	// Accept 200, 201, or 204 as success, and also 409 (conflict) if task already exists
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusConflict {
		c.logger.Error("Register task definition failed",
			zap.String("task_name", task.Name),
			zap.Int("status_code", resp.StatusCode),
			zap.String("response_body", string(body)))
		return fmt.Errorf("register task definition failed with status %d: %s", resp.StatusCode, string(body))
	}

	if resp.StatusCode == http.StatusConflict {
		c.logger.Info("Task definition already exists",
			zap.String("task_name", task.Name))
	} else {
		c.logger.Info("Registered task definition",
			zap.String("task_name", task.Name),
			zap.Int("status_code", resp.StatusCode))
	}

	return nil
}

// CreateUnderwritingWorkflowDefinition creates the underwriting workflow definition
func (c *HTTPConductorClient) CreateUnderwritingWorkflowDefinition() *WorkflowDefinition {
	return &WorkflowDefinition{
		Name:        "underwriting_workflow",
		Description: "Complete loan underwriting workflow",
		Version:     1,
		Tasks: []WorkflowTask{
			{
				Name:              "credit_check",
				TaskReferenceName: "credit_check_task",
				Type:              "SIMPLE",
				InputParameters: map[string]interface{}{
					"applicationId": "${workflow.input.applicationId}",
					"userId":        "${workflow.input.userId}",
				},
			},
			{
				Name:              "income_verification",
				TaskReferenceName: "income_verification_task",
				Type:              "SIMPLE",
				InputParameters: map[string]interface{}{
					"applicationId": "${workflow.input.applicationId}",
					"userId":        "${workflow.input.userId}",
				},
			},
			{
				Name:              "risk_assessment",
				TaskReferenceName: "risk_assessment_task",
				Type:              "SIMPLE",
				InputParameters: map[string]interface{}{
					"applicationId": "${workflow.input.applicationId}",
					"userId":        "${workflow.input.userId}",
				},
			},
			{
				Name:              "underwriting_decision",
				TaskReferenceName: "underwriting_decision_task",
				Type:              "SIMPLE",
				InputParameters: map[string]interface{}{
					"applicationId": "${workflow.input.applicationId}",
					"userId":        "${workflow.input.userId}",
				},
			},
			{
				Name:              "update_application_state",
				TaskReferenceName: "update_state_task",
				Type:              "SIMPLE",
				InputParameters: map[string]interface{}{
					"applicationId": "${workflow.input.applicationId}",
					"newState":      "underwriting_completed",
				},
			},
		},
		InputParameters: []string{"applicationId", "userId"},
		OutputParameters: map[string]interface{}{
			"decision":       "${underwriting_decision_task.output.decision}",
			"approvedAmount": "${underwriting_decision_task.output.approvedAmount}",
			"interestRate":   "${underwriting_decision_task.output.interestRate}",
		},
		SchemaVersion: 2,
	}
}

// CreateTaskDefinitions creates all task definitions for underwriting
func (c *HTTPConductorClient) CreateTaskDefinitions() []*TaskDefinition {
	return []*TaskDefinition{
		{
			Name:                   "credit_check",
			Description:            "Performs credit check and analysis",
			TimeoutSeconds:         300,
			ResponseTimeoutSeconds: 280,
			RetryCount:             3,
			InputKeys:              []string{"applicationId", "userId"},
			OutputKeys:             []string{"creditScore", "creditDecision", "riskAnalysis"},
		},
		{
			Name:                   "income_verification",
			Description:            "Verifies applicant income and employment",
			TimeoutSeconds:         300,
			ResponseTimeoutSeconds: 280,
			RetryCount:             3,
			InputKeys:              []string{"applicationId", "userId"},
			OutputKeys:             []string{"incomeVerification", "incomeAnalysis"},
		},
		{
			Name:                   "risk_assessment",
			Description:            "Performs comprehensive risk assessment",
			TimeoutSeconds:         180,
			ResponseTimeoutSeconds: 160,
			RetryCount:             3,
			InputKeys:              []string{"applicationId", "userId"},
			OutputKeys:             []string{"riskAssessment", "riskLevel", "riskScore"},
		},
		{
			Name:                   "underwriting_decision",
			Description:            "Makes final underwriting decision",
			TimeoutSeconds:         120,
			ResponseTimeoutSeconds: 100,
			RetryCount:             2,
			InputKeys:              []string{"applicationId", "userId"},
			OutputKeys:             []string{"decision", "approvedAmount", "interestRate", "conditions"},
		},
		{
			Name:                   "update_application_state",
			Description:            "Updates loan application state",
			TimeoutSeconds:         60,
			ResponseTimeoutSeconds: 50,
			RetryCount:             3,
			InputKeys:              []string{"applicationId", "newState"},
			OutputKeys:             []string{"success", "stateTransition"},
		},
		{
			Name:                   "policy_compliance_check",
			Description:            "Checks policy compliance",
			TimeoutSeconds:         120,
			ResponseTimeoutSeconds: 100,
			RetryCount:             2,
			InputKeys:              []string{"applicationId"},
			OutputKeys:             []string{"compliant", "violations"},
		},
		{
			Name:                   "fraud_detection",
			Description:            "Performs fraud detection analysis",
			TimeoutSeconds:         180,
			ResponseTimeoutSeconds: 160,
			RetryCount:             2,
			InputKeys:              []string{"applicationId"},
			OutputKeys:             []string{"fraudRiskScore", "fraudIndicators"},
		},
		{
			Name:                   "calculate_interest_rate",
			Description:            "Calculates interest rate based on risk",
			TimeoutSeconds:         60,
			ResponseTimeoutSeconds: 50,
			RetryCount:             2,
			InputKeys:              []string{"applicationId", "creditScore", "riskLevel"},
			OutputKeys:             []string{"interestRate", "apr", "rateFactors"},
		},
		{
			Name:                   "final_approval",
			Description:            "Processes final loan approval",
			TimeoutSeconds:         120,
			ResponseTimeoutSeconds: 100,
			RetryCount:             1,
			InputKeys:              []string{"applicationId", "approvedAmount"},
			OutputKeys:             []string{"loanNumber", "approvalDetails"},
		},
		{
			Name:                   "process_denial",
			Description:            "Processes loan denial",
			TimeoutSeconds:         60,
			ResponseTimeoutSeconds: 50,
			RetryCount:             1,
			InputKeys:              []string{"applicationId", "denialReasons"},
			OutputKeys:             []string{"denialProcessed", "nextSteps"},
		},
		{
			Name:                   "assign_manual_review",
			Description:            "Assigns application for manual review",
			TimeoutSeconds:         60,
			ResponseTimeoutSeconds: 50,
			RetryCount:             2,
			InputKeys:              []string{"applicationId", "reviewReason"},
			OutputKeys:             []string{"assignedTo", "reviewPriority"},
		},
		{
			Name:                   "process_conditional_approval",
			Description:            "Processes conditional approval",
			TimeoutSeconds:         120,
			ResponseTimeoutSeconds: 100,
			RetryCount:             1,
			InputKeys:              []string{"applicationId", "conditions"},
			OutputKeys:             []string{"conditionalApproval", "nextSteps"},
		},
		{
			Name:                   "generate_counter_offer",
			Description:            "Generates counter offer terms",
			TimeoutSeconds:         90,
			ResponseTimeoutSeconds: 80,
			RetryCount:             1,
			InputKeys:              []string{"applicationId", "requestedAmount"},
			OutputKeys:             []string{"counterOffer", "offerTerms"},
		},
	}
}

// IsRunning returns whether the client is currently running
func (c *HTTPConductorClient) IsRunning() bool {
	return c.isRunning
}

// GetServerURL returns the Conductor server URL
func (c *HTTPConductorClient) GetServerURL() string {
	return c.baseURL
}
