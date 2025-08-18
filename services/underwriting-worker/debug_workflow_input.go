package main

import (
	"context"
	"log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"underwriting_worker/infrastructure/workflow/tasks"

	"github.com/huuhoait/los-demo/services/shared/pkg/config"
)

func debugWorkflowInputMain() {
	// Initialize logger
	logger, err := initWorkflowLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Test workflow input scenarios
	testWorkflowInputScenarios(logger, cfg)
}

func initWorkflowLogger() (*zap.Logger, error) {
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
			"service": "underwriting-worker-workflow-debug",
			"version": "1.0.0",
		},
	}

	return zapConfig.Build()
}

func testWorkflowInputScenarios(logger *zap.Logger, cfg *config.BaseConfig) {
	logger.Info("Testing workflow input scenarios")

	// Test 1: Create HTTP Conductor client
	logger.Info("=== Test 1: Creating HTTP Conductor Client ===")
	httpClient, err := tasks.NewHTTPConductorClient(logger, cfg)
	if err != nil {
		logger.Error("Failed to create HTTP Conductor client", zap.Error(err))
		return
	}
	logger.Info("Successfully created HTTP Conductor client")

	// Test 2: Test different input scenarios
	logger.Info("=== Test 2: Testing Input Scenarios ===")

	// Scenario 1: Valid input with both parameters
	logger.Info("--- Scenario 1: Valid input with both parameters ---")
	validInput := map[string]interface{}{
		"applicationId": "test-app-123",
		"userId":        "test-user-123",
	}
	testWorkflowStart(logger, httpClient, "underwriting_workflow", validInput)

	// Scenario 2: Missing userId
	logger.Info("--- Scenario 2: Missing userId ---")
	missingUserIdInput := map[string]interface{}{
		"applicationId": "test-app-123",
		// Missing userId
	}
	testWorkflowStart(logger, httpClient, "underwriting_workflow", missingUserIdInput)

	// Scenario 3: Missing applicationId
	logger.Info("--- Scenario 3: Missing applicationId ---")
	missingAppIdInput := map[string]interface{}{
		"userId": "test-user-123",
		// Missing applicationId
	}
	testWorkflowStart(logger, httpClient, "underwriting_workflow", missingAppIdInput)

	// Scenario 4: Empty userId
	logger.Info("--- Scenario 4: Empty userId ---")
	emptyUserIdInput := map[string]interface{}{
		"applicationId": "test-app-123",
		"userId":        "",
	}
	testWorkflowStart(logger, httpClient, "underwriting_workflow", emptyUserIdInput)

	// Scenario 5: Empty applicationId
	logger.Info("--- Scenario 5: Empty applicationId ---")
	emptyAppIdInput := map[string]interface{}{
		"applicationId": "",
		"userId":        "test-user-123",
	}
	testWorkflowStart(logger, httpClient, "underwriting_workflow", emptyAppIdInput)

	// Scenario 6: Nil input
	logger.Info("--- Scenario 6: Nil input ---")
	testWorkflowStart(logger, httpClient, "underwriting_workflow", nil)

	// Scenario 7: Wrong data types
	logger.Info("--- Scenario 7: Wrong data types ---")
	wrongTypesInput := map[string]interface{}{
		"applicationId": 123, // Should be string
		"userId":        456, // Should be string
	}
	testWorkflowStart(logger, httpClient, "underwriting_workflow", wrongTypesInput)

	// Test 3: Test task execution with different inputs
	logger.Info("=== Test 3: Testing Task Execution ===")
	testTaskExecution(logger, cfg)
}

func testWorkflowStart(logger *zap.Logger, client *tasks.HTTPConductorClient, workflowName string, input map[string]interface{}) {
	logger.Info("Testing workflow start",
		zap.String("workflow_name", workflowName),
		zap.Any("input", input))

	// Try to start the workflow
	workflowID, err := client.StartWorkflow(workflowName, input)
	if err != nil {
		logger.Error("Failed to start workflow",
			zap.String("workflow_name", workflowName),
			zap.Any("input", input),
			zap.Error(err))
	} else {
		logger.Info("Successfully started workflow",
			zap.String("workflow_id", workflowID),
			zap.String("workflow_name", workflowName),
			zap.Any("input", input))
	}
}

func testTaskExecution(logger *zap.Logger, cfg *config.BaseConfig) {
	logger.Info("Testing task execution with different inputs")

	// Create task worker
	taskWorker := tasks.NewUnderwritingTaskWorker(logger, cfg)

	// Test income verification task with different inputs
	testCases := []struct {
		name  string
		input map[string]interface{}
	}{
		{
			name: "Valid input",
			input: map[string]interface{}{
				"applicationId": "test-app-123",
				"userId":        "test-user-123",
			},
		},
		{
			name: "Missing userId",
			input: map[string]interface{}{
				"applicationId": "test-app-123",
			},
		},
		{
			name: "Missing applicationId",
			input: map[string]interface{}{
				"userId": "test-user-123",
			},
		},
		{
			name: "Empty userId",
			input: map[string]interface{}{
				"applicationId": "test-app-123",
				"userId":        "",
			},
		},
		{
			name: "Empty applicationId",
			input: map[string]interface{}{
				"applicationId": "",
				"userId":        "test-user-123",
			},
		},
		{
			name:  "Nil input",
			input: nil,
		},
	}

	for _, testCase := range testCases {
		logger.Info("Testing income verification task", zap.String("case", testCase.name))

		// Get the income verification handler
		handler := taskWorker.GetIncomeVerificationHandler()
		if handler == nil {
			logger.Error("Income verification handler is nil")
			continue
		}

		// Execute the task
		ctx := context.Background()
		output, err := handler.Execute(ctx, testCase.input)

		if err != nil {
			logger.Error("Task execution failed",
				zap.String("case", testCase.name),
				zap.Any("input", testCase.input),
				zap.Error(err))
		} else {
			logger.Info("Task execution succeeded",
				zap.String("case", testCase.name),
				zap.Any("input", testCase.input),
				zap.Any("output", output))
		}
	}
}
