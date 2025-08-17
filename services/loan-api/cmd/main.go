package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/huuhoait/los-demo/services/loan-api/application"
	"github.com/huuhoait/los-demo/services/loan-api/domain"
	"github.com/huuhoait/los-demo/services/loan-api/infrastructure/database/postgres"
	"github.com/huuhoait/los-demo/services/loan-api/infrastructure/workflow"
	"github.com/huuhoait/los-demo/services/loan-api/interfaces"
	"github.com/huuhoait/los-demo/services/loan-api/interfaces/middleware"
	"github.com/huuhoait/los-demo/services/loan-api/pkg/config"
	"github.com/huuhoait/los-demo/services/shared/pkg/i18n"
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

	logger.Info("Starting loan API service",
		zap.String("version", cfg.Application.Version),
		zap.String("environment", cfg.Application.Environment),
		zap.Int("port", cfg.Server.Port),
	)

	// Initialize i18n localizer
	localizer, err := i18n.NewLocalizer()
	if err != nil {
		logger.Warn("Failed to initialize i18n localizer, using default", zap.Error(err))
		localizer = &i18n.Localizer{}
	}

	// Initialize database connection
	dbConnection, err := postgres.NewConnection(&postgres.Config{
		Host:            cfg.Database.Host,
		Port:            fmt.Sprintf("%d", cfg.Database.Port),
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Name,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}, logger)
	if err != nil {
		logger.Warn("Failed to initialize database connection, using mock repositories", zap.Error(err))
		dbConnection = nil
	}

	// Initialize repositories
	var userRepo application.UserRepository
	var loanRepo application.LoanRepository
	if dbConnection != nil {
		factory := postgres.NewFactory(dbConnection, logger)
		userRepo = factory.GetUserRepository()
		loanRepo = factory.GetLoanRepository()
	} else {
		// Use mock repositories for now
		userRepo = &MockUserRepository{}
		loanRepo = &MockLoanRepository{}
	}

	// Initialize workflow orchestrator
	conductorClient := workflow.NewConductorClientImpl(cfg.Conductor.BaseURL, logger)
	workflowOrchestrator := workflow.NewLoanWorkflowOrchestrator(conductorClient, logger, localizer)

	// Initialize services
	loanService := application.NewLoanService(userRepo, loanRepo, workflowOrchestrator, logger, localizer)

	// Initialize handlers
	loanHandler := interfaces.NewLoanHandler(loanService, logger, localizer)

	// Setup HTTP server
	router := setupRouter(logger, loanHandler, localizer)

	server := &http.Server{
		Addr:         cfg.GetServerAddr(),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting HTTP server",
			zap.String("addr", server.Addr),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.GracefulShutdownTimeout)*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// Mock repositories for when database is not available
type MockUserRepository struct{}
type MockLoanRepository struct{}

func (m *MockUserRepository) CreateUser(ctx context.Context, user *domain.User) (string, error) {
	return "mock-user-123", nil
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	return &domain.User{ID: id}, nil
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return &domain.User{Email: email}, nil
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	return nil
}

func (m *MockUserRepository) DeleteUser(ctx context.Context, id string) error {
	return nil
}

func (m *MockLoanRepository) CreateApplication(ctx context.Context, app *domain.LoanApplication) error {
	return nil
}

func (m *MockLoanRepository) GetApplicationByID(ctx context.Context, id string) (*domain.LoanApplication, error) {
	return &domain.LoanApplication{ID: id}, nil
}

func (m *MockLoanRepository) GetApplicationsByUserID(ctx context.Context, userID string) ([]*domain.LoanApplication, error) {
	return []*domain.LoanApplication{}, nil
}

func (m *MockLoanRepository) UpdateApplication(ctx context.Context, app *domain.LoanApplication) error {
	return nil
}

func (m *MockLoanRepository) DeleteApplication(ctx context.Context, id string) error {
	return nil
}

func (m *MockLoanRepository) CreateOffer(ctx context.Context, offer *domain.LoanOffer) error {
	return nil
}

func (m *MockLoanRepository) GetOfferByApplicationID(ctx context.Context, applicationID string) (*domain.LoanOffer, error) {
	return nil, fmt.Errorf("not found")
}

func (m *MockLoanRepository) UpdateOffer(ctx context.Context, offer *domain.LoanOffer) error {
	return nil
}

func (m *MockLoanRepository) CreateStateTransition(ctx context.Context, transition *domain.StateTransition) error {
	return nil
}

func (m *MockLoanRepository) GetStateTransitions(ctx context.Context, applicationID string) ([]*domain.StateTransition, error) {
	return []*domain.StateTransition{}, nil
}

func (m *MockLoanRepository) SaveWorkflowExecution(ctx context.Context, execution *domain.WorkflowExecution) error {
	return nil
}

func (m *MockLoanRepository) GetWorkflowExecutionByApplicationID(ctx context.Context, applicationID string) (*domain.WorkflowExecution, error) {
	return nil, fmt.Errorf("not found")
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

// setupRouter sets up the Gin router with middleware and routes
func setupRouter(logger *zap.Logger, loanHandler *interfaces.LoanHandler, localizer *i18n.Localizer) *gin.Engine {
	// Set Gin mode
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(corsMiddleware())
	router.Use(loggerMiddleware(logger))

	// Add i18n middleware to set localizer in context
	i18nMiddleware := middleware.NewI18nMiddleware(localizer, logger)
	router.Use(i18nMiddleware.Handler())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "loan-api",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// API routes
	v1 := router.Group("/v1")
	{
		// Register loan routes
		loanHandler.RegisterRoutes(v1)
	}

	return router
}

// corsMiddleware handles CORS
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Language, X-Request-ID")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// loggerMiddleware logs HTTP requests
func loggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithWriter(gin.DefaultWriter)
}
