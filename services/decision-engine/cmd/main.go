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

	"github.com/huuhoait/los-demo/services/decision-engine/application"
	"github.com/huuhoait/los-demo/services/decision-engine/domain"
	"github.com/huuhoait/los-demo/services/decision-engine/infrastructure"
	"github.com/huuhoait/los-demo/services/decision-engine/interfaces"
	"github.com/huuhoait/los-demo/services/decision-engine/pkg/config"
	"github.com/huuhoait/los-demo/services/decision-engine/pkg/logger"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	zapLogger, err := logger.New(cfg.Logger.Level, cfg.Environment)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	logger := zapLogger.With(zap.String("service", "decision-engine"))

	// Initialize database
	db, err := setupDatabase(cfg.Database.URL, logger)
	if err != nil {
		logger.Fatal("Failed to setup database", zap.Error(err))
	}
	defer db.Close()

	// Initialize services
	decisionService, err := setupServices(db, cfg, logger)
	if err != nil {
		logger.Fatal("Failed to setup services", zap.Error(err))
	}

	// Initialize HTTP handler
	handler := interfaces.NewDecisionHandler(decisionService, logger)

	// Setup router
	router := setupRouter(handler, cfg, logger)

	// Start server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", zap.String("port", cfg.Server.Port))
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

// setupDatabase initializes database connection and runs migrations
func setupDatabase(databaseURL string, logger *zap.Logger) (*sql.DB, error) {
	logger.Info("Connecting to database")

	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize repository (this will create tables if needed)
	decisionRepo := infrastructure.NewDecisionRepository(db, logger)
	if err := decisionRepo.InitializeDatabase(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	logger.Info("Database connection established")
	return db, nil
}

// setupServices initializes all application services
func setupServices(db *sql.DB, cfg *config.Config, logger *zap.Logger) (*application.DecisionEngineService, error) {
	// Initialize repositories
	decisionRepo := infrastructure.NewDecisionRepository(db, logger)

	// Initialize services
	riskService := application.NewRiskAssessmentService(logger)
	
	// Create a mock rules service
	rulesService := NewMockRulesService(logger)
	
	decisionService := application.NewDecisionEngineService(
		riskService,
		rulesService,
		decisionRepo,
		logger,
	)

	return decisionService, nil
}

// setupRouter configures the HTTP router
func setupRouter(handler *interfaces.DecisionHandler, cfg *config.Config, logger *zap.Logger) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	
	// Add middleware
	router.Use(gin.Recovery())
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		
		logger.Info("HTTP Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("ip", c.ClientIP()),
		)
	})

	// Setup routes
	handler.RegisterRoutes(router)

	return router
}

// MockRulesService provides a mock implementation of domain.RulesEngineService
type MockRulesService struct {
	logger *zap.Logger
}

// NewMockRulesService creates a new mock rules service
func NewMockRulesService(logger *zap.Logger) *MockRulesService {
	return &MockRulesService{
		logger: logger,
	}
}

// EvaluateRules implements domain.RulesEngineService
func (m *MockRulesService) EvaluateRules(request *domain.DecisionRequest, assessment *domain.RiskAssessment) (*domain.DecisionResponse, error) {
	// Mock implementation - always return approval for testing
	return &domain.DecisionResponse{
		ApplicationID: request.ApplicationID,
		Decision:      domain.DecisionApprove,
		RiskScore:     assessment.OverallScore,
		RiskCategory:  domain.RiskLow,
	}, nil
}

// GetActiveRules implements domain.RulesEngineService
func (m *MockRulesService) GetActiveRules() ([]domain.DecisionRule, error) {
	return []domain.DecisionRule{}, nil
}

// AddRule implements domain.RulesEngineService
func (m *MockRulesService) AddRule(rule *domain.DecisionRule) error {
	return nil
}

// UpdateRule implements domain.RulesEngineService
func (m *MockRulesService) UpdateRule(rule *domain.DecisionRule) error {
	return nil
}

// DeleteRule implements domain.RulesEngineService
func (m *MockRulesService) DeleteRule(ruleID string) error {
	return nil
}
