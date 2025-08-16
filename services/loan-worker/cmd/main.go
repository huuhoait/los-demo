package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"loan-worker/infrastructure/database/postgres"
	"loan-worker/infrastructure/workflow"
	"loan-worker/pkg/config"
	"loan-worker/pkg/i18n"
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

	logger.Info("Starting loan worker service",
		zap.String("version", cfg.Application.Version),
		zap.String("environment", cfg.Application.Environment),
	)

	// Initialize i18n
	localizer, err := initI18n()
	if err != nil {
		logger.Fatal("Failed to initialize i18n", zap.Error(err))
	}

	// Initialize database connection
	dbConfig := &postgres.Config{
		Host:            cfg.Database.Host,
		Port:            strconv.Itoa(cfg.Database.Port),
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Name,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
	}

	dbConnection, err := postgres.NewConnection(dbConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbConnection.Close()

	// Initialize repositories with real database implementation
	dbFactory := postgres.NewFactory(dbConnection, logger)
	loanRepo := dbFactory.GetLoanRepository()

	// Initialize workflow orchestrator with real Conductor client
	conductorClient := workflow.NewConductorClientImpl(
		cfg.Conductor.BaseURL,
		logger,
	)

	// Initialize and start task worker with repository
	taskWorker := workflow.NewTaskWorkerWithRepository(conductorClient, logger, localizer, loanRepo)

	// Start task worker in a goroutine
	go func() {
		logger.Info("Starting task worker")
		if err := taskWorker.Start(context.Background()); err != nil {
			logger.Error("Task worker stopped with error", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down worker...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Stop the task worker gracefully
	logger.Info("Worker exited")
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

// initI18n initializes the internationalization
func initI18n() (*i18n.Localizer, error) {
	return i18n.NewLocalizer()
}
