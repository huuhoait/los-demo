package main

import (
	"log"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"underwriting_worker/infrastructure/workflow/tasks"

	"github.com/huuhoait/los-demo/services/shared/pkg/config"
)

func debugTaskRegistrationMain() {
	// Initialize logger
	logger, err := initTaskLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Test Conductor connection and task registration
	testConductorConnection(logger, cfg)
}

func initTaskLogger() (*zap.Logger, error) {
	zapConfig := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.DebugLevel),
		Development: true,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "console",
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
			"service": "underwriting-worker-task-debug",
			"version": "1.0.0",
		},
	}

	return zapConfig.Build()
}

func testConductorConnection(logger *zap.Logger, cfg *config.BaseConfig) {
	logger.Info("Testing Conductor connection and task registration")

	// Test 1: Check if we can create HTTP Conductor client
	logger.Info("=== Test 1: Creating HTTP Conductor Client ===")
	httpClient, err := tasks.NewHTTPConductorClient(logger, cfg)
	if err != nil {
		logger.Error("Failed to create HTTP Conductor client", zap.Error(err))
		logger.Info("This might indicate a configuration issue or Conductor is not running")
		return
	}
	logger.Info("Successfully created HTTP Conductor client")

	// Test 2: Check Conductor health (skip for now since testConnection is private)
	logger.Info("=== Test 2: Skipping Conductor Health Check (method is private) ===")
	logger.Info("Conductor health check skipped - will be tested during task registration")

	// Test 3: Create task definitions
	logger.Info("=== Test 3: Creating Task Definitions ===")
	taskDefs := httpClient.CreateTaskDefinitions()
	logger.Info("Created task definitions", zap.Int("count", len(taskDefs)))

	for _, taskDef := range taskDefs {
		logger.Info("Task definition",
			zap.String("name", taskDef.Name),
			zap.String("description", taskDef.Description),
			zap.Int("timeout", taskDef.TimeoutSeconds))
	}

	// Test 4: Register task definitions
	logger.Info("=== Test 4: Registering Task Definitions ===")
	successfulRegistrations := 0
	totalTasks := len(taskDefs)

	for _, taskDef := range taskDefs {
		if err := httpClient.RegisterTaskDefinition(taskDef); err != nil {
			logger.Error("Failed to register task definition",
				zap.String("task_name", taskDef.Name),
				zap.Error(err))
		} else {
			logger.Info("Successfully registered task definition", zap.String("task_name", taskDef.Name))
			successfulRegistrations++
		}
	}

	logger.Info("Task registration summary",
		zap.Int("successful", successfulRegistrations),
		zap.Int("total", totalTasks),
		zap.Int("failed", totalTasks-successfulRegistrations))

	// Test 5: Test polling for a specific task
	logger.Info("=== Test 5: Testing Task Polling ===")
	testTaskPolling(logger, httpClient, "credit_check")

	// Test 6: Create and register workflow definition
	logger.Info("=== Test 6: Creating Workflow Definition ===")
	workflowDef := httpClient.CreateUnderwritingWorkflowDefinition()
	logger.Info("Created workflow definition",
		zap.String("name", workflowDef.Name),
		zap.Int("version", workflowDef.Version),
		zap.Int("task_count", len(workflowDef.Tasks)))

	if err := httpClient.RegisterWorkflowDefinition(workflowDef); err != nil {
		logger.Error("Failed to register workflow definition", zap.Error(err))
	} else {
		logger.Info("Successfully registered workflow definition", zap.String("name", workflowDef.Name))
	}

	// Test 7: Test workflow execution
	logger.Info("=== Test 7: Testing Workflow Execution ===")
	testWorkflowExecution(logger, httpClient)
}

func testTaskPolling(logger *zap.Logger, client *tasks.HTTPConductorClient, taskType string) {
	logger.Info("Testing task polling", zap.String("task_type", taskType))

	// Register a simple worker for testing
	client.RegisterWorker(taskType, func(task *tasks.MockTask) (*tasks.MockTaskResult, error) {
		logger.Info("Mock task executed", zap.String("task_id", task.TaskID))
		return &tasks.MockTaskResult{
			TaskID:     task.TaskID,
			Status:     "COMPLETED",
			OutputData: map[string]interface{}{"result": "mock_success"},
			WorkerID:   "test-worker",
		}, nil
	})

	// Try to start polling (this will fail if no tasks are available, which is expected)
	if err := client.StartPolling(); err != nil {
		logger.Error("Failed to start polling", zap.Error(err))
	} else {
		logger.Info("Started polling successfully")
		// Stop polling after a short time
		time.Sleep(2 * time.Second)
		client.StopPolling()
	}
}

func testWorkflowExecution(logger *zap.Logger, client *tasks.HTTPConductorClient) {
	logger.Info("Testing workflow execution")

	// Test input
	input := map[string]interface{}{
		"applicationId": "test-app-123",
		"userId":        "test-user-123",
	}

	// Try to start a workflow
	workflowID, err := client.StartWorkflow("underwriting_workflow", input)
	if err != nil {
		logger.Error("Failed to start workflow", zap.Error(err))
		logger.Info("This might be expected if Conductor is not fully configured")
	} else {
		logger.Info("Successfully started workflow", zap.String("workflow_id", workflowID))
	}
}
