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
	"go.uber.org/zap/zapcore"

	"github.com/huuhoait/los-demo/services/auth/application"
	"github.com/huuhoait/los-demo/services/auth/infrastructure"
	"github.com/huuhoait/los-demo/services/auth/interfaces"
	"github.com/huuhoait/los-demo/services/auth/interfaces/middleware"
	"github.com/huuhoait/los-demo/services/auth/pkg/i18n"
)

// Config holds application configuration
type Config struct {
	Server struct {
		Port         string        `env:"PORT" envDefault:"8080"`
		ReadTimeout  time.Duration `env:"READ_TIMEOUT" envDefault:"10s"`
		WriteTimeout time.Duration `env:"WRITE_TIMEOUT" envDefault:"10s"`
		IdleTimeout  time.Duration `env:"IDLE_TIMEOUT" envDefault:"60s"`
	}
	Database struct {
		Host         string `env:"DB_HOST" envDefault:"localhost"`
		Port         string `env:"DB_PORT" envDefault:"5432"`
		Name         string `env:"DB_NAME" envDefault:"los_auth"`
		User         string `env:"DB_USER" envDefault:"postgres"`
		Password     string `env:"DB_PASSWORD" envDefault:"password"`
		MaxOpenConns int    `env:"DB_MAX_OPEN_CONNS" envDefault:"25"`
		MaxIdleConns int    `env:"DB_MAX_IDLE_CONNS" envDefault:"5"`
		SSLMode      string `env:"DB_SSL_MODE" envDefault:"disable"`
	}
	Redis struct {
		Host     string `env:"REDIS_HOST" envDefault:"localhost"`
		Port     string `env:"REDIS_PORT" envDefault:"6379"`
		Password string `env:"REDIS_PASSWORD" envDefault:""`
		DB       int    `env:"REDIS_DB" envDefault:"0"`
	}
	JWT struct {
		SigningKey string        `env:"JWT_SIGNING_KEY" envDefault:"your-secret-key"`
		Issuer     string        `env:"JWT_ISSUER" envDefault:"los-auth-service"`
		TTL        time.Duration `env:"JWT_TTL" envDefault:"15m"`
	}
	Logging struct {
		Level  string `env:"LOG_LEVEL" envDefault:"info"`
		Format string `env:"LOG_FORMAT" envDefault:"json"`
	}
}

func main() {
	// Load configuration
	config := loadConfig()

	// Initialize logger
	logger := initLogger(config.Logging.Level, config.Logging.Format)
	defer logger.Sync()

	logger.Info("Starting authentication service",
		zap.String("version", "v1.0.0"),
		zap.String("port", config.Server.Port))

	// Initialize database
	db, err := initDatabase(config, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis
	redisClient := initRedis(config, logger)
	defer redisClient.Close()

	// Initialize i18n
	localizer, err := initI18n(logger)
	if err != nil {
		logger.Fatal("Failed to initialize i18n", zap.Error(err))
	}

	// Initialize services
	authService := initAuthService(db, redisClient, config, logger, localizer)

	// Initialize HTTP server
	server := initServer(config, authService, logger, localizer)

	// Start server
	go func() {
		logger.Info("Server starting", zap.String("address", ":"+config.Server.Port))
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

// loadConfig loads configuration from environment variables
func loadConfig() *Config {
	config := &Config{}

	// Set defaults (in production, use a proper config library like viper)
	config.Server.Port = getEnv("PORT", "8080")
	config.Database.Host = getEnv("DB_HOST", "localhost")
	config.Database.Port = getEnv("DB_PORT", "5432")
	config.Database.Name = getEnv("DB_NAME", "los_auth")
	config.Database.User = getEnv("DB_USER", "postgres")
	config.Database.Password = getEnv("DB_PASSWORD", "password")
	config.Redis.Host = getEnv("REDIS_HOST", "localhost")
	config.Redis.Port = getEnv("REDIS_PORT", "6379")
	config.JWT.SigningKey = getEnv("JWT_SIGNING_KEY", "your-secret-key")
	config.JWT.Issuer = getEnv("JWT_ISSUER", "los-auth-service")
	config.Logging.Level = getEnv("LOG_LEVEL", "info")

	return config
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// initLogger initializes the zap logger
func initLogger(level, format string) *zap.Logger {
	var config zap.Config

	if format == "console" {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	// Set log level
	switch level {
	case "debug":
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn":
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		config.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	}

	logger, err := config.Build()
	if err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}

	return logger
}

// initDatabase initializes the PostgreSQL database connection
func initDatabase(config *Config, logger *zap.Logger) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Database.Host,
		config.Database.Port,
		config.Database.User,
		config.Database.Password,
		config.Database.Name,
		config.Database.SSLMode,
	)
	logger.Info("Database connection string", zap.String("dsn", dsn))
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(config.Database.MaxOpenConns)
	db.SetMaxIdleConns(config.Database.MaxIdleConns)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established",
		zap.String("host", config.Database.Host),
		zap.String("database", config.Database.Name))

	return db, nil
}

// initRedis initializes the Redis client
func initRedis(config *Config, logger *zap.Logger) *redis.Client {
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

// initI18n initializes internationalization with embedded translation files
func initI18n(logger *zap.Logger) (*i18n.Localizer, error) {
	cfg := &i18n.Config{
		DefaultLanguage: "en",
		SupportedLangs:  []string{"en", "vi"},
		FallbackLang:    "en",
	}

	localizer, err := i18n.NewLocalizer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize localizer: %w", err)
	}

	logger.Info("I18n initialized",
		zap.Strings("supported_languages", localizer.SupportedLanguages()),
		zap.String("default_language", cfg.DefaultLanguage),
	)

	return localizer, nil
}

// initAuthService initializes the authentication service with all dependencies
func initAuthService(db *sqlx.DB, redisClient *redis.Client, config *Config, logger *zap.Logger, localizer *i18n.Localizer) *application.AuthService {
	// Initialize repositories
	userRepo := infrastructure.NewPostgresUserRepository(db, logger)
	sessionRepo := infrastructure.NewPostgresSessionRepository(db, logger)

	// Initialize cache service
	cacheService := infrastructure.NewRedisCacheService(redisClient, logger)

	// Initialize token manager
	tokenManager := infrastructure.NewJWTTokenManager(
		[]byte(config.JWT.SigningKey),
		config.JWT.Issuer,
		config.JWT.TTL,
		cacheService,
		logger,
		localizer,
	)

	// Initialize audit logger (placeholder)
	auditLogger := infrastructure.NewAuditLogger(logger)

	// Create auth service
	authService := application.NewAuthService(
		userRepo,
		sessionRepo,
		tokenManager,
		cacheService,
		auditLogger,
		logger,
		localizer,
	)

	logger.Info("Authentication service initialized")
	return authService
}

// initServer initializes the HTTP server with routes and middleware
func initServer(config *Config, authService *application.AuthService, logger *zap.Logger, localizer *i18n.Localizer) *http.Server {
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
	router.Use(requestIDMiddleware())
	router.Use(corsMiddleware())
	router.Use(loggerMiddleware(logger))

	// Add i18n middleware
	i18nMiddleware := middleware.NewI18nMiddleware(localizer, logger)
	router.Use(i18nMiddleware.Handler())

	// Initialize handlers and middleware
	authHandler := interfaces.NewAuthHandler(authService, logger, localizer)
	authMiddleware := interfaces.NewAuthMiddleware(authService, logger, localizer)

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
		ReadTimeout:  config.Server.ReadTimeout,
		WriteTimeout: config.Server.WriteTimeout,
		IdleTimeout:  config.Server.IdleTimeout,
	}
}

// Custom middleware

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func loggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		logger.Info("HTTP request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("duration", duration),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.GetHeader("User-Agent")),
		)
	}
}

func generateRequestID() string {
	// Simple implementation - use proper UUID in production
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
