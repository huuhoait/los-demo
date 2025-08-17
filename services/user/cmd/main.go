package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/huuhoait/los-demo/services/shared/pkg/config"
	"github.com/huuhoait/los-demo/services/shared/pkg/i18n"
	"github.com/huuhoait/los-demo/services/shared/pkg/logger"
	sharedMiddleware "github.com/huuhoait/los-demo/services/shared/pkg/middleware"
	"github.com/huuhoait/los-demo/services/user/application"
	"github.com/huuhoait/los-demo/services/user/domain"
	"github.com/huuhoait/los-demo/services/user/infrastructure"
	"github.com/huuhoait/los-demo/services/user/interfaces"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("./configs")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	loggerConfig := logger.Config{
		Level:       cfg.Logging.Level,
		Format:      cfg.Logging.Format,
		Output:      cfg.Logging.Output,
		Environment: cfg.Environment,
	}

	appLogger, err := logger.New(loggerConfig)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer appLogger.Sync()

	appLogger.Info("Starting User Service", zap.String("version", "1.0.0"))

	// Initialize localizer
	i18nConfig := &i18n.Config{
		DefaultLanguage: cfg.I18n.DefaultLanguage,
		SupportedLangs:  cfg.I18n.SupportedLanguages,
		FallbackLang:    cfg.I18n.FallbackLanguage,
	}
	localizer, err := i18n.NewLocalizer(i18nConfig)
	if err != nil {
		appLogger.Fatal("Failed to initialize localizer", zap.Error(err))
	}

	// Initialize database
	db, err := initializeDatabase(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis
	redisClient := initializeRedis(cfg, appLogger)
	defer redisClient.Close()

	// Initialize services and repositories
	app, err := initializeApplication(db, redisClient, cfg, appLogger, localizer)
	if err != nil {
		appLogger.Fatal("Failed to initialize application", zap.Error(err))
	}

	// Initialize HTTP server
	server := initializeHTTPServer(app, cfg, appLogger, localizer)

	// Start server in goroutine
	go func() {
		appLogger.Info("Starting HTTP server", zap.String("address", server.Addr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down User Service...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	appLogger.Info("User Service shutdown complete")
}

func initializeDatabase(cfg *config.Config, appLogger *logger.Logger) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	appLogger.Info("Database connection established",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
		zap.String("database", cfg.Database.Name),
	)

	return db, nil
}

func initializeRedis(cfg *config.Config, appLogger *logger.Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
		MaxRetries:   cfg.Redis.MaxRetries,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		appLogger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	appLogger.Info("Redis connection established",
		zap.String("host", cfg.Redis.Host),
		zap.Int("port", cfg.Redis.Port),
	)

	return client
}

type Application struct {
	UserService domain.UserService
	UserHandler *interfaces.UserHandler
	Logger      *zap.Logger
}

func initializeApplication(
	db *sqlx.DB,
	redisClient *redis.Client,
	cfg *config.Config,
	appLogger *logger.Logger,
	localizer *i18n.Localizer,
) (*Application, error) {
	// Initialize repositories
	userRepo := infrastructure.NewPostgresUserRepository(db, appLogger.Logger)
	kycRepo := infrastructure.NewPostgresKYCRepository(db, appLogger.Logger)
	documentRepo := infrastructure.NewPostgresDocumentRepository(db, appLogger.Logger)

	// Initialize infrastructure services
	cacheService := infrastructure.NewRedisCacheService(redisClient, appLogger.Logger)
	validationService := infrastructure.NewValidationService(appLogger.Logger)
	encryptionService := infrastructure.NewAESEncryptionService(cfg.Encryption.MasterKey, appLogger.Logger)

	// Mock services for development (replace with real implementations in production)
	kycProvider := infrastructure.NewMockKYCProviderService(appLogger.Logger)

	// TODO: Initialize real services
	var storageService domain.DocumentStorageService
	var notificationService domain.NotificationService
	var auditService domain.AuditService

	// For now, use mock implementations
	storageService = NewMockStorageService(appLogger.Logger)
	notificationService = NewMockNotificationService(appLogger.Logger)
	auditService = NewMockAuditService(appLogger.Logger)

	// Initialize user service
	userService := application.NewUserService(
		userRepo,
		kycRepo,
		documentRepo,
		storageService,
		encryptionService,
		kycProvider,
		notificationService,
		validationService,
		auditService,
		cacheService,
		appLogger.Logger,
		localizer,
	)

	// Initialize handlers
	userHandler := interfaces.NewUserHandler(userService, appLogger.Logger, localizer)

	return &Application{
		UserService: userService,
		UserHandler: userHandler,
		Logger:      appLogger.Logger,
	}, nil
}

func initializeHTTPServer(app *Application, cfg *config.Config, appLogger *logger.Logger, localizer *i18n.Localizer) *http.Server {
	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(sharedMiddleware.CORSMiddleware())
	router.Use(sharedMiddleware.RequestIDMiddleware())

	// Add logger middleware
	loggerMiddleware := logger.NewLoggerMiddleware(appLogger)
	router.Use(loggerMiddleware.Handler())

	router.Use(timestampMiddleware())

	// Add i18n middleware
	i18nMiddleware := sharedMiddleware.NewI18nMiddleware(localizer, appLogger.Logger)
	router.Use(i18nMiddleware.Handler())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "user-service",
			"version":   "1.0.0",
			"timestamp": time.Now().UTC(),
		})
	})

	// API routes
	v1 := router.Group("/api/v1")
	app.UserHandler.RegisterRoutes(v1)

	return &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}
}

func timestampMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("timestamp", time.Now().UTC())
		c.Next()
	}
}

func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Mock service implementations for development

type MockStorageService struct {
	logger *zap.Logger
}

func NewMockStorageService(logger *zap.Logger) domain.DocumentStorageService {
	return &MockStorageService{logger: logger}
}

func (m *MockStorageService) UploadFile(ctx context.Context, key string, content io.Reader, contentType string, metadata map[string]string) error {
	m.logger.Info("Mock file upload", zap.String("key", key), zap.String("content_type", contentType))
	return nil
}

func (m *MockStorageService) DownloadFile(ctx context.Context, key string) (io.ReadCloser, error) {
	m.logger.Info("Mock file download", zap.String("key", key))
	return io.NopCloser(strings.NewReader("mock file content")), nil
}

func (m *MockStorageService) DeleteFile(ctx context.Context, key string) error {
	m.logger.Info("Mock file deletion", zap.String("key", key))
	return nil
}

func (m *MockStorageService) GetFileMetadata(ctx context.Context, key string) (map[string]string, error) {
	return map[string]string{"mock": "metadata"}, nil
}

func (m *MockStorageService) GeneratePresignedURL(ctx context.Context, key string, expiration int) (string, error) {
	return fmt.Sprintf("https://mock-storage.example.com/%s", key), nil
}

func (m *MockStorageService) EncryptContent(content []byte) ([]byte, string, error) {
	return content, "mock-encryption-key", nil
}

func (m *MockStorageService) DecryptContent(encryptedContent []byte, key string) ([]byte, error) {
	return encryptedContent, nil
}

type MockNotificationService struct {
	logger *zap.Logger
}

func NewMockNotificationService(logger *zap.Logger) domain.NotificationService {
	return &MockNotificationService{logger: logger}
}

func (m *MockNotificationService) SendWelcomeEmail(ctx context.Context, userID, email, firstName string) error {
	m.logger.Info("Mock welcome email sent", zap.String("user_id", userID), zap.String("email", email))
	return nil
}

func (m *MockNotificationService) SendEmailVerification(ctx context.Context, userID, email, verificationCode string) error {
	m.logger.Info("Mock email verification sent", zap.String("user_id", userID), zap.String("code", verificationCode))
	return nil
}

func (m *MockNotificationService) SendPasswordReset(ctx context.Context, userID, email, resetToken string) error {
	m.logger.Info("Mock password reset sent", zap.String("user_id", userID))
	return nil
}

func (m *MockNotificationService) SendPhoneVerification(ctx context.Context, userID, phone, verificationCode string) error {
	m.logger.Info("Mock phone verification sent", zap.String("user_id", userID), zap.String("code", verificationCode))
	return nil
}

func (m *MockNotificationService) SendSecurityAlert(ctx context.Context, userID, phone, alertMessage string) error {
	m.logger.Info("Mock security alert sent", zap.String("user_id", userID))
	return nil
}

func (m *MockNotificationService) SendPushNotification(ctx context.Context, userID, title, message string, data map[string]interface{}) error {
	m.logger.Info("Mock push notification sent", zap.String("user_id", userID), zap.String("title", title))
	return nil
}

type MockAuditService struct {
	logger *zap.Logger
}

func NewMockAuditService(logger *zap.Logger) domain.AuditService {
	return &MockAuditService{logger: logger}
}

func (m *MockAuditService) LogUserCreated(ctx context.Context, userID, email string, metadata map[string]interface{}) error {
	m.logger.Info("Mock audit: user created", zap.String("user_id", userID), zap.String("email", email))
	return nil
}

func (m *MockAuditService) LogUserUpdated(ctx context.Context, userID string, changes map[string]interface{}) error {
	m.logger.Info("Mock audit: user updated", zap.String("user_id", userID))
	return nil
}

func (m *MockAuditService) LogProfileUpdated(ctx context.Context, userID string, changes map[string]interface{}) error {
	m.logger.Info("Mock audit: profile updated", zap.String("user_id", userID))
	return nil
}

func (m *MockAuditService) LogDocumentUploaded(ctx context.Context, userID, documentID, documentType string) error {
	m.logger.Info("Mock audit: document uploaded", zap.String("user_id", userID), zap.String("document_id", documentID))
	return nil
}

func (m *MockAuditService) LogKYCStatusChanged(ctx context.Context, userID, verificationType string, oldStatus, newStatus domain.KYCStatus) error {
	m.logger.Info("Mock audit: KYC status changed", zap.String("user_id", userID), zap.String("type", verificationType))
	return nil
}

func (m *MockAuditService) LogSecurityEvent(ctx context.Context, userID, eventType string, metadata map[string]interface{}) error {
	m.logger.Info("Mock audit: security event", zap.String("user_id", userID), zap.String("event_type", eventType))
	return nil
}

func (m *MockAuditService) LogDataAccess(ctx context.Context, userID, accessedBy, dataType string) error {
	m.logger.Info("Mock audit: data access", zap.String("user_id", userID), zap.String("data_type", dataType))
	return nil
}
