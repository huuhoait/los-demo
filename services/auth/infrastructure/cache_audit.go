package infrastructure

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/huuhoait/los-demo/services/auth/domain"
)

// RedisCacheService implements CacheService using Redis
type RedisCacheService struct {
	client *redis.Client
	logger *zap.Logger
}

// NewRedisCacheService creates a new Redis cache service
func NewRedisCacheService(client *redis.Client, logger *zap.Logger) *RedisCacheService {
	return &RedisCacheService{
		client: client,
		logger: logger,
	}
}

// Set stores a value in the cache with expiration
func (r *RedisCacheService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	logger := r.logger.With(
		zap.String("operation", "cache_set"),
		zap.String("key", key),
		zap.Duration("expiration", expiration),
	)

	// Serialize value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		logger.Error("Failed to marshal cache value", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_018, "Cache serialization error", "Failed to serialize cache value")
	}

	// Set value in Redis
	err = r.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		logger.Error("Failed to set cache value", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_018, "Cache storage error", "Failed to store value in cache")
	}

	logger.Debug("Cache value set successfully")
	return nil
}

// Get retrieves a value from the cache
func (r *RedisCacheService) Get(ctx context.Context, key string) (interface{}, error) {
	logger := r.logger.With(
		zap.String("operation", "cache_get"),
		zap.String("key", key),
	)

	// Get value from Redis
	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			logger.Debug("Cache key not found")
			return nil, domain.NewAuthError("CACHE_NOT_FOUND", "Key not found", "Cache key does not exist")
		}
		logger.Error("Failed to get cache value", zap.Error(err))
		return nil, domain.NewAuthError(domain.AUTH_018, "Cache retrieval error", "Failed to retrieve value from cache")
	}

	// Deserialize JSON
	var value interface{}
	err = json.Unmarshal([]byte(data), &value)
	if err != nil {
		logger.Error("Failed to unmarshal cache value", zap.Error(err))
		return nil, domain.NewAuthError(domain.AUTH_018, "Cache deserialization error", "Failed to deserialize cache value")
	}

	logger.Debug("Cache value retrieved successfully")
	return value, nil
}

// Delete removes a value from the cache
func (r *RedisCacheService) Delete(ctx context.Context, key string) error {
	logger := r.logger.With(
		zap.String("operation", "cache_delete"),
		zap.String("key", key),
	)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		logger.Error("Failed to delete cache value", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_018, "Cache deletion error", "Failed to delete value from cache")
	}

	logger.Debug("Cache value deleted successfully")
	return nil
}

// Exists checks if a key exists in the cache
func (r *RedisCacheService) Exists(ctx context.Context, key string) (bool, error) {
	logger := r.logger.With(
		zap.String("operation", "cache_exists"),
		zap.String("key", key),
	)

	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		logger.Error("Failed to check cache key existence", zap.Error(err))
		return false, domain.NewAuthError(domain.AUTH_018, "Cache existence check error", "Failed to check key existence")
	}

	exists := count > 0
	logger.Debug("Cache key existence checked", zap.Bool("exists", exists))
	return exists, nil
}

// Increment atomically increments a numeric value
func (r *RedisCacheService) Increment(ctx context.Context, key string) (int64, error) {
	logger := r.logger.With(
		zap.String("operation", "cache_increment"),
		zap.String("key", key),
	)

	value, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		logger.Error("Failed to increment cache value", zap.Error(err))
		return 0, domain.NewAuthError(domain.AUTH_018, "Cache increment error", "Failed to increment cache value")
	}

	logger.Debug("Cache value incremented", zap.Int64("new_value", value))
	return value, nil
}

// SetExpiration sets expiration for an existing key
func (r *RedisCacheService) SetExpiration(ctx context.Context, key string, expiration time.Duration) error {
	logger := r.logger.With(
		zap.String("operation", "cache_set_expiration"),
		zap.String("key", key),
		zap.Duration("expiration", expiration),
	)

	success, err := r.client.Expire(ctx, key, expiration).Result()
	if err != nil {
		logger.Error("Failed to set cache expiration", zap.Error(err))
		return domain.NewAuthError(domain.AUTH_018, "Cache expiration error", "Failed to set key expiration")
	}

	if !success {
		logger.Warn("Key does not exist for expiration setting")
		return domain.NewAuthError("CACHE_NOT_FOUND", "Key not found", "Cannot set expiration for non-existent key")
	}

	logger.Debug("Cache expiration set successfully")
	return nil
}

// AuditLogger implements audit logging functionality
type AuditLogger struct {
	logger *zap.Logger
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(logger *zap.Logger) *AuditLogger {
	return &AuditLogger{
		logger: logger,
	}
}

// LogAuthEvent logs an authentication event
func (a *AuditLogger) LogAuthEvent(ctx context.Context, event *domain.AuthEvent) error {
	a.logger.Info("Authentication event",
		zap.String("event_id", event.ID),
		zap.String("user_id", event.UserID),
		zap.String("event_type", event.EventType),
		zap.String("session_id", event.SessionID),
		zap.String("ip_address", event.IPAddress),
		zap.String("user_agent", event.UserAgent),
		zap.Bool("success", event.Success),
		zap.String("error_code", event.ErrorCode),
		zap.String("error_message", event.ErrorMessage),
		zap.Any("metadata", event.Metadata),
		zap.Time("timestamp", event.Timestamp),
	)

	// In production, this would also send to audit storage (database, kafka, etc.)
	return nil
}

// LogSecurityEvent logs a security event
func (a *AuditLogger) LogSecurityEvent(ctx context.Context, event *domain.SecurityEvent) error {
	a.logger.Warn("Security event",
		zap.String("event_id", event.ID),
		zap.String("event_type", event.EventType),
		zap.String("user_id", event.UserID),
		zap.String("ip_address", event.IPAddress),
		zap.String("user_agent", event.UserAgent),
		zap.String("severity", event.Severity),
		zap.String("description", event.Description),
		zap.Any("metadata", event.Metadata),
		zap.Time("timestamp", event.Timestamp),
	)

	// In production, this would also send to security monitoring systems
	return nil
}
