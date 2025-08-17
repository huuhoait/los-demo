package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// Config holds cache configuration
type Config struct {
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Password string `yaml:"password" json:"password"`
	Database int    `yaml:"database" json:"database"`
	PoolSize int    `yaml:"pool_size" json:"pool_size"`
}

// Client wraps redis client with additional functionality
type Client struct {
	*redis.Client
	config Config
}

// NewClient creates a new cache client
func NewClient(config Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.Database,
		PoolSize: config.PoolSize,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{
		Client: rdb,
		config: config,
	}, nil
}

// Health returns Redis health information
func (c *Client) Health() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := c.Ping(ctx).Err(); err != nil {
		return map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	}

	return map[string]interface{}{
		"status": "healthy",
	}
}

// SetJSON stores a value as JSON
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return c.Set(ctx, key, data, expiration).Err()
}

// GetJSON retrieves a JSON value
func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

// SetString stores a string value
func (c *Client) SetString(ctx context.Context, key, value string, expiration time.Duration) error {
	return c.Set(ctx, key, value, expiration).Err()
}

// GetString retrieves a string value
func (c *Client) GetString(ctx context.Context, key string) (string, error) {
	return c.Get(ctx, key).Result()
}

// Exists checks if a key exists
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.Client.Exists(ctx, key).Result()
	return result > 0, err
}

// Delete deletes a key
func (c *Client) Delete(ctx context.Context, keys ...string) error {
	return c.Del(ctx, keys...).Err()
}

// Increment increments a counter
func (c *Client) Increment(ctx context.Context, key string) (int64, error) {
	return c.Incr(ctx, key).Result()
}

// IncrementBy increments a counter by a specific value
func (c *Client) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return c.IncrBy(ctx, key, value).Result()
}

// Expire sets expiration for a key
func (c *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return c.Client.Expire(ctx, key, expiration).Err()
}

// TTL returns time to live for a key
func (c *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.Client.TTL(ctx, key).Result()
}

// Repository provides cache repository functionality
type Repository struct {
	client *Client
	prefix string
}

// NewRepository creates a new cache repository
func NewRepository(client *Client, prefix string) *Repository {
	return &Repository{
		client: client,
		prefix: prefix,
	}
}

// buildKey builds a cache key with prefix
func (r *Repository) buildKey(key string) string {
	if r.prefix != "" {
		return fmt.Sprintf("%s:%s", r.prefix, key)
	}
	return key
}

// Set stores a value
func (r *Repository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.SetJSON(ctx, r.buildKey(key), value, expiration)
}

// Get retrieves a value
func (r *Repository) Get(ctx context.Context, key string, dest interface{}) error {
	return r.client.GetJSON(ctx, r.buildKey(key), dest)
}

// Exists checks if a key exists
func (r *Repository) Exists(ctx context.Context, key string) (bool, error) {
	return r.client.Exists(ctx, r.buildKey(key))
}

// Delete deletes a key
func (r *Repository) Delete(ctx context.Context, key string) error {
	return r.client.Delete(ctx, r.buildKey(key))
}

// DeleteByPattern deletes keys matching a pattern
func (r *Repository) DeleteByPattern(ctx context.Context, pattern string) error {
	fullPattern := r.buildKey(pattern)
	keys, err := r.client.Keys(ctx, fullPattern).Result()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	return r.client.Delete(ctx, keys...)
}

// SessionStore provides session storage functionality
type SessionStore struct {
	repository *Repository
	prefix     string
	expiration time.Duration
}

// NewSessionStore creates a new session store
func NewSessionStore(client *Client, expiration time.Duration) *SessionStore {
	return &SessionStore{
		repository: NewRepository(client, "session"),
		expiration: expiration,
	}
}

// Set stores session data
func (s *SessionStore) Set(ctx context.Context, sessionID string, data interface{}) error {
	return s.repository.Set(ctx, sessionID, data, s.expiration)
}

// Get retrieves session data
func (s *SessionStore) Get(ctx context.Context, sessionID string, dest interface{}) error {
	return s.repository.Get(ctx, sessionID, dest)
}

// Delete deletes a session
func (s *SessionStore) Delete(ctx context.Context, sessionID string) error {
	return s.repository.Delete(ctx, sessionID)
}

// Refresh extends session expiration
func (s *SessionStore) Refresh(ctx context.Context, sessionID string) error {
	key := s.repository.buildKey(sessionID)
	return s.repository.client.Expire(ctx, key, s.expiration)
}

// RateLimiter provides rate limiting functionality
type RateLimiter struct {
	client *Client
	prefix string
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(client *Client) *RateLimiter {
	return &RateLimiter{
		client: client,
		prefix: "rate_limit",
	}
}

// Allow checks if an action is allowed within the rate limit
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit int64, window time.Duration) (bool, error) {
	fullKey := fmt.Sprintf("%s:%s", rl.prefix, key)

	current, err := rl.client.Increment(ctx, fullKey)
	if err != nil {
		return false, err
	}

	if current == 1 {
		// Set expiration for the first request
		if err := rl.client.Expire(ctx, fullKey, window); err != nil {
			return false, err
		}
	}

	return current <= limit, nil
}

// GetLimit returns current usage for a key
func (rl *RateLimiter) GetLimit(ctx context.Context, key string) (int64, error) {
	fullKey := fmt.Sprintf("%s:%s", rl.prefix, key)
	result, err := rl.client.Get(ctx, fullKey).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return result, err
}

// Reset resets the rate limit for a key
func (rl *RateLimiter) Reset(ctx context.Context, key string) error {
	fullKey := fmt.Sprintf("%s:%s", rl.prefix, key)
	return rl.client.Delete(ctx, fullKey)
}

// Cache provides high-level caching functionality
type Cache struct {
	client     *Client
	defaultTTL time.Duration
}

// NewCache creates a new cache
func NewCache(client *Client, defaultTTL time.Duration) *Cache {
	return &Cache{
		client:     client,
		defaultTTL: defaultTTL,
	}
}

// Remember gets value from cache or executes function and caches result
func (c *Cache) Remember(ctx context.Context, key string, fn func() (interface{}, error)) (interface{}, error) {
	return c.RememberFor(ctx, key, c.defaultTTL, fn)
}

// RememberFor gets value from cache or executes function and caches result for specific duration
func (c *Cache) RememberFor(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error)) (interface{}, error) {
	// Try to get from cache first
	var result interface{}
	if err := c.client.GetJSON(ctx, key, &result); err == nil {
		return result, nil
	}

	// Execute function
	value, err := fn()
	if err != nil {
		return nil, err
	}

	// Store in cache
	if err := c.client.SetJSON(ctx, key, value, ttl); err != nil {
		// Log error but don't fail
		// In production, you might want to use a logger here
	}

	return value, nil
}

// Forget removes a key from cache
func (c *Cache) Forget(ctx context.Context, key string) error {
	return c.client.Delete(ctx, key)
}

// Flush clears all cache
func (c *Cache) Flush(ctx context.Context) error {
	return c.client.FlushDB(ctx).Err()
}

// Put stores a value in cache
func (c *Cache) Put(ctx context.Context, key string, value interface{}) error {
	return c.PutFor(ctx, key, value, c.defaultTTL)
}

// PutFor stores a value in cache for specific duration
func (c *Cache) PutFor(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return c.client.SetJSON(ctx, key, value, ttl)
}

// Has checks if a key exists in cache
func (c *Cache) Has(ctx context.Context, key string) (bool, error) {
	return c.client.Exists(ctx, key)
}

// Lock provides distributed locking functionality
type Lock struct {
	client *Client
	prefix string
}

// NewLock creates a new lock manager
func NewLock(client *Client) *Lock {
	return &Lock{
		client: client,
		prefix: "lock",
	}
}

// Acquire acquires a distributed lock
func (l *Lock) Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	lockKey := fmt.Sprintf("%s:%s", l.prefix, key)
	result, err := l.client.SetNX(ctx, lockKey, "locked", ttl).Result()
	return result, err
}

// Release releases a distributed lock
func (l *Lock) Release(ctx context.Context, key string) error {
	lockKey := fmt.Sprintf("%s:%s", l.prefix, key)
	return l.client.Delete(ctx, lockKey)
}

// IsLocked checks if a lock is currently held
func (l *Lock) IsLocked(ctx context.Context, key string) (bool, error) {
	lockKey := fmt.Sprintf("%s:%s", l.prefix, key)
	return l.client.Exists(ctx, lockKey)
}
