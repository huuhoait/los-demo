package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"decision-engine/application"
	"decision-engine/infrastructure"
	"decision-engine/interfaces"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds application configuration
type Config struct {
	Port               string
	DatabaseURL        string
	LogLevel           string
	Environment        string
	CreditBureauConfig infrastructure.CreditBureauConfig
}

// loadConfig loads configuration from environment variables
func loadConfig() *Config {
	return &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://user:password@localhost/decision_engine?sslmode=disable"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Environment: getEnv("ENVIRONMENT", "development"),
		CreditBureauConfig: infrastructure.CreditBureauConfig{
			ExperianEndpoint:   getEnv("EXPERIAN_ENDPOINT", "https://api.experian.com"),
			EquifaxEndpoint:    getEnv("EQUIFAX_ENDPOINT", "https://api.equifax.com"),
			TransUnionEndpoint: getEnv("TRANSUNION_ENDPOINT", "https://api.transunion.com"),
			APITimeout:         30 * time.Second,
			RetryAttempts:      3,
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// setupLogger creates a configured zap logger
func setupLogger(logLevel, environment string) (*zap.Logger, error) {
	var config zap.Config

	if environment == "production" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Parse log level
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	config.Level = zap.NewAtomicLevelAt(level)

	// Build logger
	logger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}

// setupDatabase initializes database connection and runs migrations
func setupDatabase(databaseURL string, logger *zap.Logger) (*sql.DB, error) {
	logger.Info("Connecting to database")

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize database tables
	decisionRepo := infrastructure.NewDecisionRepository(db, logger)
	if err := decisionRepo.InitializeDatabase(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	logger.Info("Database connection established and tables initialized")
	return db, nil
}

// setupServices initializes all application services
func setupServices(db *sql.DB, config *Config, logger *zap.Logger) (*application.DecisionEngineService, error) {
	// Initialize repositories
	decisionRepo := infrastructure.NewDecisionRepository(db, logger)
	creditRepo := infrastructure.NewCreditBureauRepository(logger, config.CreditBureauConfig)

	// Initialize services
	riskService := application.NewRiskAssessmentService(logger)
	decisionService := application.NewDecisionEngineService(
		decisionRepo,
		creditRepo,
		riskService,
		logger,
	)

	logger.Info("Services initialized successfully")
	return decisionService, nil
}

// setupHTTPServer creates and configures the HTTP server
func setupHTTPServer(decisionService *application.DecisionEngineService, config *Config, logger *zap.Logger) *http.Server {
	// Set Gin mode based on environment
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Initialize handlers
	decisionHandler := interfaces.NewDecisionHandler(decisionService, logger)
	decisionHandler.SetupRoutes(router)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	return server
}

func main() {
	// Load configuration
	config := loadConfig()

	// Setup logger
	logger, err := setupLogger(config.LogLevel, config.Environment)
	if err != nil {
		log.Fatal("Failed to setup logger:", err)
	}
	defer logger.Sync()

	logger.Info("Starting Decision Engine Service",
		zap.String("environment", config.Environment),
		zap.String("port", config.Port),
		zap.String("log_level", config.LogLevel),
	)

	// Setup database
	db, err := setupDatabase(config.DatabaseURL, logger)
	if err != nil {
		logger.Fatal("Failed to setup database", zap.Error(err))
	}
	defer db.Close()

	// Setup services
	decisionService, err := setupServices(db, config, logger)
	if err != nil {
		logger.Fatal("Failed to setup services", zap.Error(err))
	}

	// Setup HTTP server
	server := setupHTTPServer(decisionService, config, logger)

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	} else {
		logger.Info("Server shutdown completed")
	}
}
