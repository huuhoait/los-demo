package logger

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger provides structured logging capabilities
type Logger struct {
	*zap.Logger
}

// Config holds logger configuration
type Config struct {
	Level       string `yaml:"level" json:"level"`
	Format      string `yaml:"format" json:"format"`
	Output      string `yaml:"output" json:"output"`
	Environment string `yaml:"environment" json:"environment"`
}

// New creates a new logger with the specified configuration
func New(config Config) (*Logger, error) {
	var zapConfig zap.Config

	// Set config based on environment
	if config.Environment == "production" {
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set log level
	switch config.Level {
	case "debug":
		zapConfig.Level.SetLevel(zapcore.DebugLevel)
	case "info":
		zapConfig.Level.SetLevel(zapcore.InfoLevel)
	case "warn":
		zapConfig.Level.SetLevel(zapcore.WarnLevel)
	case "error":
		zapConfig.Level.SetLevel(zapcore.ErrorLevel)
	default:
		zapConfig.Level.SetLevel(zapcore.InfoLevel)
	}

	// Set output format
	if config.Format == "console" {
		zapConfig.Encoding = "console"
	} else {
		zapConfig.Encoding = "json"
	}

	logger, err := zapConfig.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &Logger{Logger: logger}, nil
}

// WithContext adds context fields to the logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	fields := []zap.Field{}

	// Extract common context values
	if requestID := ctx.Value("request_id"); requestID != nil {
		fields = append(fields, zap.String("request_id", requestID.(string)))
	}

	if userID := ctx.Value("user_id"); userID != nil {
		fields = append(fields, zap.String("user_id", userID.(string)))
	}

	if traceID := ctx.Value("trace_id"); traceID != nil {
		fields = append(fields, zap.String("trace_id", traceID.(string)))
	}

	return &Logger{Logger: l.Logger.With(fields...)}
}

// WithRequest adds request-specific fields to the logger
func (l *Logger) WithRequest(c *gin.Context) *Logger {
	fields := []zap.Field{
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("remote_addr", c.ClientIP()),
	}

	if requestID := c.GetString("request_id"); requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	if userID := c.GetString("user_id"); userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	return &Logger{Logger: l.Logger.With(fields...)}
}

// LoggerMiddleware provides a Gin middleware for request logging
type LoggerMiddleware struct {
	logger *Logger
}

// NewLoggerMiddleware creates a new logger middleware
func NewLoggerMiddleware(logger *Logger) *LoggerMiddleware {
	return &LoggerMiddleware{logger: logger}
}

// Handler returns a Gin middleware function for logging requests
func (m *LoggerMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		status := c.Writer.Status()

		// Create log fields
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
		}

		// Add request ID if available
		if requestID := c.GetString("request_id"); requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}

		// Add user ID if available
		if userID := c.GetString("user_id"); userID != "" {
			fields = append(fields, zap.String("user_id", userID))
		}

		// Log based on status code
		if status >= 500 {
			m.logger.Error("HTTP request failed", fields...)
		} else if status >= 400 {
			m.logger.Warn("HTTP request failed", fields...)
		} else {
			m.logger.Info("HTTP request", fields...)
		}
	}
}

// Structured logging helpers

// LogUserAction logs user actions for audit purposes
func (l *Logger) LogUserAction(userID, action, resource string, metadata map[string]interface{}) {
	fields := []zap.Field{
		zap.String("user_id", userID),
		zap.String("action", action),
		zap.String("resource", resource),
		zap.String("type", "user_action"),
	}

	for key, value := range metadata {
		fields = append(fields, zap.Any(key, value))
	}

	l.Info("User action performed", fields...)
}

// LogSecurityEvent logs security-related events
func (l *Logger) LogSecurityEvent(eventType, userID, description string, metadata map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", eventType),
		zap.String("user_id", userID),
		zap.String("description", description),
		zap.String("type", "security_event"),
		zap.Time("timestamp", time.Now()),
	}

	for key, value := range metadata {
		fields = append(fields, zap.Any(key, value))
	}

	l.Warn("Security event", fields...)
}

// LogAPIRequest logs API requests with details
func (l *Logger) LogAPIRequest(method, path string, statusCode int, duration time.Duration, userID string) {
	fields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
		zap.String("type", "api_request"),
	}

	if userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	if statusCode >= 400 {
		l.Warn("API request completed with error", fields...)
	} else {
		l.Info("API request completed", fields...)
	}
}

// LogDatabaseOperation logs database operations
func (l *Logger) LogDatabaseOperation(operation, table string, duration time.Duration, err error) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Duration("duration", duration),
		zap.String("type", "database_operation"),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Error("Database operation failed", fields...)
	} else {
		l.Debug("Database operation completed", fields...)
	}
}

// LogCacheOperation logs cache operations
func (l *Logger) LogCacheOperation(operation, key string, hit bool, duration time.Duration) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("key", key),
		zap.Bool("hit", hit),
		zap.Duration("duration", duration),
		zap.String("type", "cache_operation"),
	}

	l.Debug("Cache operation", fields...)
}

// LogExternalServiceCall logs external service calls
func (l *Logger) LogExternalServiceCall(service, method, endpoint string, statusCode int, duration time.Duration, err error) {
	fields := []zap.Field{
		zap.String("service", service),
		zap.String("method", method),
		zap.String("endpoint", endpoint),
		zap.Int("status_code", statusCode),
		zap.Duration("duration", duration),
		zap.String("type", "external_service_call"),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.Error("External service call failed", fields...)
	} else {
		l.Info("External service call completed", fields...)
	}
}

// LogBusinessEvent logs business events
func (l *Logger) LogBusinessEvent(eventType, description string, metadata map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", eventType),
		zap.String("description", description),
		zap.String("type", "business_event"),
		zap.Time("timestamp", time.Now()),
	}

	for key, value := range metadata {
		fields = append(fields, zap.Any(key, value))
	}

	l.Info("Business event", fields...)
}

// LogError logs errors with context
func (l *Logger) LogError(err error, message string, metadata map[string]interface{}) {
	fields := []zap.Field{
		zap.Error(err),
		zap.String("type", "error"),
		zap.Time("timestamp", time.Now()),
	}

	for key, value := range metadata {
		fields = append(fields, zap.Any(key, value))
	}

	l.Error(message, fields...)
}
