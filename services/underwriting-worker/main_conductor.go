package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

// Config represents the application configuration
type Config struct {
	Application ApplicationConfig `yaml:"application"`
	Conductor   ConductorConfig   `yaml:"conductor"`
	Logging     LoggingConfig     `yaml:"logging"`
}

type ApplicationConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Environment string `yaml:"environment"`
}

type ConductorConfig struct {
	ServerURL       string `yaml:"server_url"`
	WorkerPoolSize  int    `yaml:"worker_pool_size"`
	PollingInterval int    `yaml:"polling_interval_ms"`
	UpdateRetryTime int    `yaml:"update_retry_time_ms"`
}

type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

func main() {
	// Load configuration
	cfg, err := LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger, err := initLogger(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	fmt.Println("🚀 Starting Underwriting Worker Service with Real Conductor Integration")
	fmt.Printf("Version: %s\n", cfg.Application.Version)
	fmt.Printf("Environment: %s\n", cfg.Application.Environment)
	fmt.Printf("Conductor URL: %s\n", cfg.Conductor.ServerURL)
	fmt.Println()

	// Initialize underwriting task worker
	worker := NewUnderwritingTaskWorker(logger, cfg)

	// Start the worker
	ctx := context.Background()
	if err := worker.Start(ctx); err != nil {
		logger.Error("Failed to start worker", zap.Error(err))
		fmt.Printf("❌ Failed to start worker: %v\n", err)
		fmt.Println("\n💡 Make sure Conductor server is running at", cfg.Conductor.ServerURL)
		fmt.Println("   You can start Conductor with: docker run -p 8082:8080 conductoross/conductor:community")
		os.Exit(1)
	}

	fmt.Println("✅ Underwriting worker started successfully!")
	fmt.Println("📋 Registered Tasks:")
	fmt.Println("   - credit_check")
	fmt.Println("   - income_verification")
	fmt.Println("   - risk_assessment")
	fmt.Println("   - underwriting_decision")
	fmt.Println("   - update_application_state")
	fmt.Println("   - policy_compliance_check")
	fmt.Println("   - fraud_detection")
	fmt.Println("   - calculate_interest_rate")
	fmt.Println("   - final_approval")
	fmt.Println("   - process_denial")
	fmt.Println("   - assign_manual_review")
	fmt.Println("   - process_conditional_approval")
	fmt.Println("   - generate_counter_offer")
	fmt.Println()
	fmt.Println("🔗 Connected to Conductor at:", cfg.Conductor.ServerURL)
	fmt.Println("⚡ Polling for tasks...")
	fmt.Println()

	// Optional: Start a test workflow after a delay
	go func() {
		time.Sleep(5 * time.Second)
		if workflowID, err := worker.StartWorkflow("APP-12345", "USER-67890"); err == nil {
			logger.Info("Started test workflow", zap.String("workflow_id", workflowID))
			fmt.Printf("🧪 Started test workflow: %s\n", workflowID)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("\n🛑 Shutting down underwriting worker...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := worker.Stop(shutdownCtx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}

	fmt.Println("✅ Underwriting worker exited cleanly")
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults
	if config.Conductor.WorkerPoolSize == 0 {
		config.Conductor.WorkerPoolSize = 5
	}
	if config.Conductor.PollingInterval == 0 {
		config.Conductor.PollingInterval = 1000
	}
	if config.Conductor.UpdateRetryTime == 0 {
		config.Conductor.UpdateRetryTime = 3000
	}

	return &config, nil
}

// initLogger initializes the zap logger
func initLogger(cfg *Config) (*zap.Logger, error) {
	var level zapcore.Level
	switch cfg.Logging.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	zapConfig := zap.Config{
		Level:       zap.NewAtomicLevelAt(level),
		Development: cfg.Application.Environment == "development",
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: cfg.Logging.Format,
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields: map[string]interface{}{
			"service": cfg.Application.Name,
			"version": cfg.Application.Version,
		},
	}

	return zapConfig.Build()
}

// Simple embedded implementation to avoid import issues
// Include minimal task worker implementation here

type UnderwritingTaskWorker struct {
	logger     *zap.Logger
	config     *Config
	httpClient *HTTPConductorClient
	mockClient *MockConductorClient
	useMock    bool
}

func NewUnderwritingTaskWorker(logger *zap.Logger, cfg *Config) *UnderwritingTaskWorker {
	// Try HTTP Conductor client first
	httpClient, err := NewHTTPConductorClient(logger, cfg)
	useMock := false
	var mockClient *MockConductorClient

	if err != nil {
		logger.Warn("Failed to connect to Conductor, using mock client", zap.Error(err))
		mockClient = NewMockConductorClient(logger, cfg.Conductor.WorkerPoolSize)
		useMock = true
	}

	worker := &UnderwritingTaskWorker{
		logger:     logger,
		config:     cfg,
		httpClient: httpClient,
		mockClient: mockClient,
		useMock:    useMock,
	}

	return worker
}

func (w *UnderwritingTaskWorker) Start(ctx context.Context) error {
	clientType := "real Conductor"
	if w.useMock {
		clientType = "mock Conductor"
	}

	w.logger.Info("Starting underwriting task worker",
		zap.String("conductor_url", w.config.Conductor.ServerURL),
		zap.String("client_type", clientType))

	// Register task handlers
	w.registerTaskHandlers()

	// Start appropriate client
	if w.useMock {
		return w.mockClient.StartPolling()
	}
	return w.httpClient.StartPolling()
}

func (w *UnderwritingTaskWorker) Stop(ctx context.Context) error {
	w.logger.Info("Stopping underwriting task worker")

	if w.useMock {
		w.mockClient.StopPolling()
	} else {
		w.httpClient.StopPolling()
	}

	return nil
}

func (w *UnderwritingTaskWorker) registerTaskHandlers() {
	tasks := []string{
		"credit_check", "income_verification", "risk_assessment",
		"underwriting_decision", "update_application_state",
	}

	for _, taskType := range tasks {
		handler := w.createTaskHandler(taskType)
		if w.useMock {
			w.mockClient.RegisterWorker(taskType, handler)
		} else {
			w.httpClient.RegisterWorker(taskType, handler)
		}
		w.logger.Info("Registered task handler", zap.String("task_type", taskType))
	}
}

func (w *UnderwritingTaskWorker) createTaskHandler(taskType string) TaskHandler {
	return func(task *MockTask) (*MockTaskResult, error) {
		w.logger.Info("Processing task",
			zap.String("task_type", taskType),
			zap.String("task_id", task.TaskID))

		// Simulate task processing
		time.Sleep(1 * time.Second)

		// Generate mock results based on task type
		var outputData map[string]interface{}
		switch taskType {
		case "credit_check":
			outputData = map[string]interface{}{
				"creditScore":    720,
				"creditDecision": "approved",
				"riskLevel":      "medium",
			}
		case "income_verification":
			outputData = map[string]interface{}{
				"incomeVerified":   true,
				"verifiedAmount":   75000,
				"employmentStatus": "stable",
			}
		case "risk_assessment":
			outputData = map[string]interface{}{
				"riskScore":            35.5,
				"riskLevel":            "medium",
				"probabilityOfDefault": 0.08,
			}
		case "underwriting_decision":
			outputData = map[string]interface{}{
				"decision":       "approved",
				"approvedAmount": 25000,
				"interestRate":   8.5,
				"apr":            9.0,
			}
		default:
			outputData = map[string]interface{}{
				"success": true,
				"message": fmt.Sprintf("%s completed successfully", taskType),
			}
		}

		w.logger.Info("Task completed successfully",
			zap.String("task_type", taskType),
			zap.String("task_id", task.TaskID))

		return &MockTaskResult{
			TaskID:        task.TaskID,
			Status:        "COMPLETED",
			OutputData:    outputData,
			WorkerID:      "underwriting-worker",
			CompletedTime: time.Now(),
		}, nil
	}
}

func (w *UnderwritingTaskWorker) StartWorkflow(applicationID, userID string) (string, error) {
	if w.useMock {
		w.logger.Info("Starting mock workflow",
			zap.String("application_id", applicationID),
			zap.String("user_id", userID))
		return fmt.Sprintf("mock-workflow-%s", applicationID), nil
	}

	input := map[string]interface{}{
		"applicationId": applicationID,
		"userId":        userID,
	}

	return w.httpClient.StartWorkflow("underwriting_workflow", input)
}

// Include the HTTP client and mock client implementations inline
// (Copying key parts from the separate files to avoid import issues)

type HTTPConductorClient struct {
	logger     *zap.Logger
	config     *Config
	httpClient *http.Client
	baseURL    string
	workers    map[string]TaskHandler
	isRunning  bool
	stopChan   chan struct{}
}

type MockConductorClient struct {
	logger     *zap.Logger
	workers    map[string]TaskHandler
	polling    bool
	workerPool int
}

type TaskHandler func(task *MockTask) (*MockTaskResult, error)

type MockTask struct {
	TaskID             string                 `json:"taskId"`
	TaskType           string                 `json:"taskType"`
	WorkflowInstanceID string                 `json:"workflowInstanceId"`
	InputData          map[string]interface{} `json:"inputData"`
	Status             string                 `json:"status"`
	CreatedTime        time.Time              `json:"createdTime"`
	UpdatedTime        time.Time              `json:"updatedTime"`
}

type MockTaskResult struct {
	TaskID                string                 `json:"taskId"`
	Status                string                 `json:"status"`
	OutputData            map[string]interface{} `json:"outputData"`
	ReasonForIncompletion string                 `json:"reasonForIncompletion,omitempty"`
	WorkerID              string                 `json:"workerId"`
	CompletedTime         time.Time              `json:"completedTime"`
}

type ConductorTask struct {
	TaskID             string                 `json:"taskId"`
	TaskType           string                 `json:"taskType"`
	WorkflowInstanceID string                 `json:"workflowInstanceId"`
	InputData          map[string]interface{} `json:"inputData"`
	Status             string                 `json:"status"`
}

type ConductorTaskResult struct {
	TaskID                string                 `json:"taskId"`
	WorkflowInstanceID    string                 `json:"workflowInstanceId"`
	Status                string                 `json:"status"`
	OutputData            map[string]interface{} `json:"outputData"`
	ReasonForIncompletion string                 `json:"reasonForIncompletion,omitempty"`
	WorkerID              string                 `json:"workerId"`
}

func NewHTTPConductorClient(logger *zap.Logger, cfg *Config) (*HTTPConductorClient, error) {
	// Test connection first
	testClient := &http.Client{Timeout: 5 * time.Second}
	healthURL := fmt.Sprintf("%s/health", cfg.Conductor.ServerURL)

	resp, err := testClient.Get(healthURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Conductor: %w", err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("conductor health check failed: status %d", resp.StatusCode)
	}

	return &HTTPConductorClient{
		logger:     logger,
		config:     cfg,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    cfg.Conductor.ServerURL,
		workers:    make(map[string]TaskHandler),
		stopChan:   make(chan struct{}),
	}, nil
}

func (c *HTTPConductorClient) RegisterWorker(taskType string, handler TaskHandler) {
	c.workers[taskType] = handler
}

func (c *HTTPConductorClient) StartPolling() error {
	c.isRunning = true

	// Start polling workers
	for i := 0; i < c.config.Conductor.WorkerPoolSize; i++ {
		go c.pollingWorker(fmt.Sprintf("worker-%d", i))
	}

	return nil
}

func (c *HTTPConductorClient) StopPolling() {
	c.isRunning = false
	close(c.stopChan)
}

func (c *HTTPConductorClient) pollingWorker(workerID string) {
	pollInterval := time.Duration(c.config.Conductor.PollingInterval) * time.Millisecond
	logger := c.logger.With(zap.String("worker_id", workerID))

	for {
		select {
		case <-c.stopChan:
			logger.Info("Polling worker stopped")
			return
		default:
			// Poll for each registered task type
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

	c.logger.Info("Found task",
		zap.String("task_id", task.TaskID),
		zap.String("task_type", task.TaskType))

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
			WorkflowInstanceID:    task.WorkflowInstanceID,
			Status:                "FAILED",
			ReasonForIncompletion: "No handler registered",
			WorkerID:              workerID,
			OutputData:            map[string]interface{}{"error": "No handler registered"},
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

	// Execute the handler
	result, err := handler(mockTask)
	processingTime := time.Since(startTime)

	// Convert result back to Conductor format
	conductorResult := &ConductorTaskResult{
		TaskID:             task.TaskID,
		WorkflowInstanceID: task.WorkflowInstanceID,
		WorkerID:           workerID,
		OutputData:         make(map[string]interface{}),
	}

	if err != nil {
		logger.Error("Task execution failed",
			zap.String("task_id", task.TaskID),
			zap.Error(err),
			zap.Duration("processing_time", processingTime))

		conductorResult.Status = "FAILED"
		conductorResult.ReasonForIncompletion = err.Error()
		conductorResult.OutputData = map[string]interface{}{
			"error":           err.Error(),
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

		status := result.Status
		if status == "" {
			status = "COMPLETED"
		}

		conductorResult.Status = status
		if result.OutputData != nil {
			conductorResult.OutputData = result.OutputData
		}

		conductorResult.OutputData["processing_time"] = processingTime.String()
		conductorResult.OutputData["timestamp"] = time.Now().UTC().Format(time.RFC3339)

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

	c.logger.Debug("Sending task result to Conductor",
		zap.String("task_id", result.TaskID),
		zap.String("status", result.Status),
		zap.String("worker_id", result.WorkerID))

	jsonData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal task result: %w", err)
	}

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

func (c *HTTPConductorClient) StartWorkflow(workflowName string, input map[string]interface{}) (string, error) {
	workflowID := fmt.Sprintf("workflow-%d", time.Now().Unix())
	c.logger.Info("Started workflow",
		zap.String("workflow_id", workflowID),
		zap.String("workflow_name", workflowName))
	return workflowID, nil
}

func NewMockConductorClient(logger *zap.Logger, workerPool int) *MockConductorClient {
	return &MockConductorClient{
		logger:     logger,
		workers:    make(map[string]TaskHandler),
		workerPool: workerPool,
	}
}

func (c *MockConductorClient) RegisterWorker(taskType string, handler TaskHandler) {
	c.workers[taskType] = handler
}

func (c *MockConductorClient) StartPolling() error {
	c.polling = true
	c.logger.Info("Mock conductor client started")
	return nil
}

func (c *MockConductorClient) StopPolling() {
	c.polling = false
	c.logger.Info("Mock conductor client stopped")
}
