package logger

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/huuhoait/los-demo/services/user/pkg/config"
)

// NewZapLogger creates a new zap logger based on configuration
func NewZapLogger(cfg *config.Config) (*zap.Logger, error) {
	// Determine log level
	level := zapcore.InfoLevel
	switch cfg.Logging.Level {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "fatal":
		level = zapcore.FatalLevel
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
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
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create encoder
	var encoder zapcore.Encoder
	if cfg.Logging.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Create writer syncer
	var writeSyncer zapcore.WriteSyncer
	switch cfg.Logging.Output {
	case "file":
		// Ensure log directory exists
		if err := os.MkdirAll(filepath.Dir(cfg.Logging.FilePath), 0755); err != nil {
			return nil, err
		}

		// Create rotating file writer
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.Logging.FilePath,
			MaxSize:    cfg.Logging.MaxSize, // MB
			MaxBackups: cfg.Logging.MaxBackups,
			MaxAge:     cfg.Logging.MaxAge, // days
			Compress:   true,
		}
		writeSyncer = zapcore.AddSync(fileWriter)
	case "both":
		// Ensure log directory exists
		if err := os.MkdirAll(filepath.Dir(cfg.Logging.FilePath), 0755); err != nil {
			return nil, err
		}

		// Create rotating file writer
		fileWriter := &lumberjack.Logger{
			Filename:   cfg.Logging.FilePath,
			MaxSize:    cfg.Logging.MaxSize,
			MaxBackups: cfg.Logging.MaxBackups,
			MaxAge:     cfg.Logging.MaxAge,
			Compress:   true,
		}
		writeSyncer = zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
			zapcore.AddSync(fileWriter),
		)
	default: // stdout
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// Create logger with additional options
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	// Add service information to all logs
	logger = logger.With(
		zap.String("service", cfg.Service.Name),
		zap.String("version", cfg.Service.Version),
		zap.String("environment", cfg.Environment),
	)

	return logger, nil
}

// LoggerMiddleware provides logging context for HTTP requests
type LoggerMiddleware struct {
	logger *zap.Logger
}

// NewLoggerMiddleware creates a new logger middleware
func NewLoggerMiddleware(logger *zap.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{
		logger: logger,
	}
}

// WithRequestID adds request ID to logger
func (l *LoggerMiddleware) WithRequestID(requestID string) *zap.Logger {
	return l.logger.With(zap.String("request_id", requestID))
}

// WithUserID adds user ID to logger
func (l *LoggerMiddleware) WithUserID(userID string) *zap.Logger {
	return l.logger.With(zap.String("user_id", userID))
}

// WithContext adds multiple context fields to logger
func (l *LoggerMiddleware) WithContext(fields map[string]interface{}) *zap.Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	return l.logger.With(zapFields...)
}

// Structured logging helpers

// LogUserAction logs user actions for audit purposes
func LogUserAction(logger *zap.Logger, userID, action, resource string, metadata map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", "user_action"),
		zap.String("user_id", userID),
		zap.String("action", action),
		zap.String("resource", resource),
	}

	if metadata != nil {
		for key, value := range metadata {
			fields = append(fields, zap.Any(key, value))
		}
	}

	logger.Info("User action performed", fields...)
}

// LogSecurityEvent logs security-related events
func LogSecurityEvent(logger *zap.Logger, eventType, userID, description string, metadata map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", "security_event"),
		zap.String("security_event_type", eventType),
		zap.String("user_id", userID),
		zap.String("description", description),
	}

	if metadata != nil {
		for key, value := range metadata {
			fields = append(fields, zap.Any(key, value))
		}
	}

	logger.Warn("Security event detected", fields...)
}

// LogAPIRequest logs API requests
func LogAPIRequest(logger *zap.Logger, method, path string, statusCode int, duration float64, userID string) {
	logger.Info("API request",
		zap.String("event_type", "api_request"),
		zap.String("method", method),
		zap.String("path", path),
		zap.Int("status_code", statusCode),
		zap.Float64("duration_ms", duration),
		zap.String("user_id", userID),
	)
}

// LogDatabaseOperation logs database operations
func LogDatabaseOperation(logger *zap.Logger, operation, table string, duration float64, error error) {
	fields := []zap.Field{
		zap.String("event_type", "database_operation"),
		zap.String("operation", operation),
		zap.String("table", table),
		zap.Float64("duration_ms", duration),
	}

	if error != nil {
		fields = append(fields, zap.Error(error))
		logger.Error("Database operation failed", fields...)
	} else {
		logger.Debug("Database operation completed", fields...)
	}
}

// LogCacheOperation logs cache operations
func LogCacheOperation(logger *zap.Logger, operation, key string, hit bool, duration float64) {
	logger.Debug("Cache operation",
		zap.String("event_type", "cache_operation"),
		zap.String("operation", operation),
		zap.String("key", key),
		zap.Bool("hit", hit),
		zap.Float64("duration_ms", duration),
	)
}

// LogExternalServiceCall logs external service calls
func LogExternalServiceCall(logger *zap.Logger, service, operation string, statusCode int, duration float64, error error) {
	fields := []zap.Field{
		zap.String("event_type", "external_service_call"),
		zap.String("service", service),
		zap.String("operation", operation),
		zap.Int("status_code", statusCode),
		zap.Float64("duration_ms", duration),
	}

	if error != nil {
		fields = append(fields, zap.Error(error))
		logger.Error("External service call failed", fields...)
	} else {
		logger.Info("External service call completed", fields...)
	}
}

// LogFileOperation logs file operations
func LogFileOperation(logger *zap.Logger, operation, fileName string, fileSize int64, error error) {
	fields := []zap.Field{
		zap.String("event_type", "file_operation"),
		zap.String("operation", operation),
		zap.String("file_name", fileName),
		zap.Int64("file_size", fileSize),
	}

	if error != nil {
		fields = append(fields, zap.Error(error))
		logger.Error("File operation failed", fields...)
	} else {
		logger.Info("File operation completed", fields...)
	}
}

// LogKYCEvent logs KYC-related events
func LogKYCEvent(logger *zap.Logger, userID, verificationType, status string, metadata map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", "kyc_event"),
		zap.String("user_id", userID),
		zap.String("verification_type", verificationType),
		zap.String("status", status),
	}

	if metadata != nil {
		for key, value := range metadata {
			fields = append(fields, zap.Any(key, value))
		}
	}

	logger.Info("KYC event", fields...)
}

// LogDocumentEvent logs document-related events
func LogDocumentEvent(logger *zap.Logger, userID, documentID, documentType, event string, metadata map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event_type", "document_event"),
		zap.String("user_id", userID),
		zap.String("document_id", documentID),
		zap.String("document_type", documentType),
		zap.String("document_event", event),
	}

	if metadata != nil {
		for key, value := range metadata {
			fields = append(fields, zap.Any(key, value))
		}
	}

	logger.Info("Document event", fields...)
}

// LogValidationError logs validation errors
func LogValidationError(logger *zap.Logger, field, value, rule, message string) {
	logger.Warn("Validation error",
		zap.String("event_type", "validation_error"),
		zap.String("field", field),
		zap.String("value", value),
		zap.String("rule", rule),
		zap.String("message", message),
	)
}
