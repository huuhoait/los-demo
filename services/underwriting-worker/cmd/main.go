package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/huuhoait/los-demo/services/underwriting-worker/infrastructure/workflow/tasks"
	"github.com/huuhoait/los-demo/services/underwriting-worker/pkg/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("Invalid configuration: %v", err))
	}

	// Initialize logger
	logger, err := initLogger(cfg)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting underwriting worker service",
		zap.String("version", cfg.Application.Version),
		zap.String("environment", cfg.Application.Environment),
	)

	// Initialize task worker with enhanced underwriting tasks
	taskWorker := tasks.NewUnderwritingTaskWorker(logger, cfg)

	// Start task worker in a goroutine
	go func() {
		logger.Info("Starting underwriting task worker")
		if err := taskWorker.Start(context.Background()); err != nil {
			logger.Error("Underwriting task worker stopped with error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down underwriting worker...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop the task worker gracefully
	if err := taskWorker.Stop(ctx); err != nil {
		logger.Error("Error stopping task worker", zap.Error(err))
	}

	logger.Info("Underwriting worker exited")
}

// initLogger initializes the zap logger
func initLogger(cfg *config.Config) (*zap.Logger, error) {
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
		Development: cfg.IsDevelopment(),
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
