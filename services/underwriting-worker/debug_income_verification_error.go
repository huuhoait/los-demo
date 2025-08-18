package main

import (
	"context"
	"log"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"underwriting_worker/infrastructure/workflow/tasks"

	"github.com/huuhoait/los-demo/services/shared/pkg/config"
)

func debugIncomeMain() {
	// Initialize logger
	logger, err := initIncomeLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize task worker
	taskWorker := tasks.NewUnderwritingTaskWorker(logger, cfg)

	// Test specific error case
	testIncomeVerificationError(logger, taskWorker)
}

func initIncomeLogger() (*zap.Logger, error) {
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
			"service": "underwriting-worker-income-debug",
			"version": "1.0.0",
		},
	}

	return zapConfig.Build()
}

func testIncomeVerificationError(logger *zap.Logger, taskWorker *tasks.UnderwritingTaskWorker) {
	logger.Info("Testing income verification error scenarios")

	// Test case 1: Missing applicationId
	logger.Info("=== Test Case 1: Missing applicationId ===")
	testInput1 := map[string]interface{}{
		"userId": "test-user-123",
		// Missing applicationId
	}
	testIncomeVerificationWithInput(logger, taskWorker, testInput1)

	// Test case 2: Missing userId
	logger.Info("=== Test Case 2: Missing userId ===")
	testInput2 := map[string]interface{}{
		"applicationId": "test-app-123",
		// Missing userId
	}
	testIncomeVerificationWithInput(logger, taskWorker, testInput2)

	// Test case 3: Empty applicationId
	logger.Info("=== Test Case 3: Empty applicationId ===")
	testInput3 := map[string]interface{}{
		"applicationId": "",
		"userId":        "test-user-123",
	}
	testIncomeVerificationWithInput(logger, taskWorker, testInput3)

	// Test case 4: Empty userId
	logger.Info("=== Test Case 4: Empty userId ===")
	testInput4 := map[string]interface{}{
		"applicationId": "test-app-123",
		"userId":        "",
	}
	testIncomeVerificationWithInput(logger, taskWorker, testInput4)

	// Test case 5: Valid input (should succeed with mock data)
	logger.Info("=== Test Case 5: Valid input ===")
	testInput5 := map[string]interface{}{
		"applicationId": "test-app-123",
		"userId":        "test-user-123",
	}
	testIncomeVerificationWithInput(logger, taskWorker, testInput5)

	// Test case 6: Valid input with verification method
	logger.Info("=== Test Case 6: Valid input with verification method ===")
	testInput6 := map[string]interface{}{
		"applicationId":      "test-app-123",
		"userId":             "test-user-123",
		"verificationMethod": "paystub_verification",
	}
	testIncomeVerificationWithInput(logger, taskWorker, testInput6)

	// Test case 7: Nil input
	logger.Info("=== Test Case 7: Nil input ===")
	testIncomeVerificationWithInput(logger, taskWorker, nil)

	// Test case 8: Wrong data types
	logger.Info("=== Test Case 8: Wrong data types ===")
	testInput8 := map[string]interface{}{
		"applicationId": 123, // Should be string
		"userId":        456, // Should be string
	}
	testIncomeVerificationWithInput(logger, taskWorker, testInput8)
}

func testIncomeVerificationWithInput(logger *zap.Logger, taskWorker *tasks.UnderwritingTaskWorker, input map[string]interface{}) {
	ctx := context.Background()

	logger.Info("Testing income verification with input", zap.Any("input", input))

	// Get the income verification handler
	incomeVerificationHandler := taskWorker.GetIncomeVerificationHandler()
	if incomeVerificationHandler == nil {
		logger.Error("Income verification handler is nil")
		return
	}

	// Execute the income verification
	startTime := time.Now()
	output, err := incomeVerificationHandler.Execute(ctx, input)
	processingTime := time.Since(startTime)

	if err != nil {
		logger.Error("Income verification failed",
			zap.Error(err),
			zap.Duration("processing_time", processingTime),
			zap.Any("input", input))
	} else {
		logger.Info("Income verification completed",
			zap.Duration("processing_time", processingTime),
			zap.Any("output", output),
			zap.Any("input", input))
	}

	logger.Info("--- End of test case ---")
}
