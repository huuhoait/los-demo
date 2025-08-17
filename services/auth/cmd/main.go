package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/huuhoait/los-demo/services/auth/application"
	"github.com/huuhoait/los-demo/services/auth/infrastructure"
	"github.com/huuhoait/los-demo/services/auth/interfaces"
	"github.com/huuhoait/los-demo/services/shared/pkg/config"
	"github.com/huuhoait/los-demo/services/shared/pkg/logger"
	sharedMiddleware "github.com/huuhoait/los-demo/services/shared/pkg/middleware"
)

// Config holds application configuration
type Config struct {
	config.BaseConfig
	JWT struct {
		SigningKey string        `yaml:"signing_key" json:"signing_key"`
		Issuer     string        `yaml:"issuer" json:"issuer"`
		TTL        time.Duration `yaml:"ttl" json:"ttl"`
	} `yaml:"jwt" json:"jwt"`
}

func main() {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
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
		log.Fatal("Failed to initialize logger:", err)
	}
	defer appLogger.Sync()

	appLogger.Info("Starting authentication service",
		zap.String("version", "v1.0.0"),
		zap.String("port", cfg.Server.Port))

	// Initialize database
	db, err := initDatabase(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis
	redisClient := initRedis(cfg, appLogger)
	defer redisClient.Close()

	// Initialize services
	authService := initAuthService(db, redisClient, cfg, appLogger)

	// Initialize HTTP server
	server := initServer(cfg, authService, appLogger)

	// Start server
	go func() {
		appLogger.Info("Server starting", zap.String("address", ":"+cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			appLogger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	appLogger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		appLogger.Error("Server forced to shutdown", zap.Error(err))
	}

	appLogger.Info("Server exited")
}

// loadConfig loads configuration from environment variables
func loadConfig() (*Config, error) {
	cfg := &Config{}

	// Load base configuration with defaults
	cfg.Environment = getEnv("ENVIRONMENT", "development")
	cfg.Service.Name = getEnv("SERVICE_NAME", "auth-service")
	cfg.Service.Version = getEnv("SERVICE_VERSION", "1.0.0")
	cfg.Server.Port = getEnv("PORT", "8080")

	// Database URL or individual fields
	if dbURL := getEnv("DATABASE_URL", ""); dbURL != "" {
		cfg.Database.URL = dbURL
	} else {
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", "postgres")
		password := getEnv("DB_PASSWORD", "password")
		database := getEnv("DB_NAME", "los_auth")
		sslMode := getEnv("DB_SSL_MODE", "disable")
		cfg.Database.URL = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, database, sslMode)
	}

	// Redis configuration
	cfg.Redis.Host = getEnv("REDIS_HOST", "localhost")
	cfg.Redis.Port = getEnv("REDIS_PORT", "6379")
	cfg.Redis.Password = getEnv("REDIS_PASSWORD", "")
	cfg.Redis.DB = config.GetInt("REDIS_DB", 0)

	// Logging configuration
	cfg.Logging.Level = getEnv("LOG_LEVEL", "info")
	cfg.Logging.Format = getEnv("LOG_FORMAT", "json")
	cfg.Logging.Output = getEnv("LOG_OUTPUT", "stdout")

	// JWT configuration
	cfg.JWT.SigningKey = getEnv("JWT_SIGNING_KEY", "your-secret-key")
	cfg.JWT.Issuer = getEnv("JWT_ISSUER", "los-auth-service")
	if ttlStr := getEnv("JWT_TTL", "15m"); ttlStr != "" {
		if ttl, err := time.ParseDuration(ttlStr); err == nil {
			cfg.JWT.TTL = ttl
		} else {
			cfg.JWT.TTL = 15 * time.Minute
		}
	}

	return cfg, nil
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// initDatabase initializes the PostgreSQL database connection
func initDatabase(config *Config, logger *logger.Logger) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", config.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established")

	return db, nil
}

// initRedis initializes the Redis client
func initRedis(config *Config, logger *logger.Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Redis.Host, config.Redis.Port),
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	logger.Info("Redis connection established",
		zap.String("host", config.Redis.Host),
		zap.Int("database", config.Redis.DB))

	return client
}

// initAuthService initializes the authentication service with all dependencies
func initAuthService(db *sqlx.DB, redisClient *redis.Client, config *Config, logger *logger.Logger) *application.AuthService {
	// Initialize repositories
	userRepo := infrastructure.NewPostgresUserRepository(db, logger.Logger)
	sessionRepo := infrastructure.NewPostgresSessionRepository(db, logger.Logger)

	// Initialize cache service
	cacheService := infrastructure.NewRedisCacheService(redisClient, logger.Logger)

	// Initialize token manager
	tokenManager := infrastructure.NewJWTTokenManager(
		[]byte(config.JWT.SigningKey),
		config.JWT.Issuer,
		config.JWT.TTL,
		cacheService,
		logger.Logger,
		nil, // temporarily remove localizer
	)

	// Initialize audit logger (placeholder)
	auditLogger := infrastructure.NewAuditLogger(logger.Logger)

	// Create auth service
	authService := application.NewAuthService(
		userRepo,
		sessionRepo,
		tokenManager,
		cacheService,
		auditLogger,
		logger.Logger,
		nil, // temporarily remove localizer
	)

	logger.Info("Authentication service initialized")
	return authService
}

// initServer initializes the HTTP server with routes and middleware
func initServer(config *Config, authService *application.AuthService, appLogger *logger.Logger) *http.Server {
	// Set Gin mode
	if config.Logging.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(sharedMiddleware.RequestIDMiddleware())
	router.Use(sharedMiddleware.CORSMiddleware())

	// Add logger middleware
	loggerMiddleware := logger.NewLoggerMiddleware(appLogger)
	router.Use(loggerMiddleware.Handler())

	// Initialize handlers and middleware
	authHandler := interfaces.NewAuthHandler(authService, appLogger.Logger, nil)
	authMiddleware := interfaces.NewAuthMiddleware(authService, appLogger.Logger, nil)

	// Register routes
	v1 := router.Group("/v1")
	{
		auth := v1.Group("/auth")
		authHandler.RegisterRoutes(auth, authMiddleware)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "auth-service",
			"version":   "v1.0.0",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	return &http.Server{
		Addr:         ":" + config.Server.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}
