package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists in header
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// Generate new UUID
			requestID = uuid.New().String()
		}

		// Set request ID in context and response header
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}

// CORSMiddleware returns a CORS middleware with common settings
func CORSMiddleware() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Authorization",
		"Accept",
		"Accept-Language",
		"X-Request-ID",
		"X-Language",
		"X-Requested-With",
	}
	config.ExposeHeaders = []string{
		"X-Request-ID",
		"X-Total-Count",
		"X-Page",
		"X-Per-Page",
	}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	return cors.New(config)
}

// SecurityHeadersMiddleware adds security headers
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")

		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")

		// Strict transport security (HTTPS only)
		if c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// Content security policy
		c.Header("Content-Security-Policy", "default-src 'self'")

		// Referrer policy
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

// RateLimitMiddleware provides basic rate limiting
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

// RecoveryMiddleware handles panics gracefully
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Internal server error",
				"error":   err,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Internal server error",
			})
		}
		c.Abort()
	})
}

// TimeoutMiddleware adds request timeout
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set a timeout for the request context
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace the request context
		c.Request = c.Request.WithContext(ctx)

		// Channel to capture if request completed
		finished := make(chan struct{})

		go func() {
			c.Next()
			finished <- struct{}{}
		}()

		select {
		case <-finished:
			// Request completed successfully
		case <-ctx.Done():
			// Request timed out
			c.JSON(http.StatusRequestTimeout, gin.H{
				"success": false,
				"message": "Request timeout",
			})
			c.Abort()
		}
	}
}

// AuthorizationMiddleware validates authorization tokens
type AuthConfig struct {
	TokenHeader     string
	ValidateTokenFn func(token string) (userID string, err error)
	RequiredScopes  []string
}

func AuthorizationMiddleware(config AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header
		token := c.GetHeader(config.TokenHeader)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Authorization token required",
			})
			c.Abort()
			return
		}

		// Remove "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// Validate token
		userID, err := config.ValidateTokenFn(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": "Invalid authorization token",
			})
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("user_id", userID)

		c.Next()
	}
}

// ValidationErrorMiddleware handles validation errors
func ValidationErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check for validation errors
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Validation error",
				"errors":  c.Errors.Errors(),
				"detail":  err.Error(),
			})
		}
	}
}

// PaginationMiddleware handles pagination parameters
type PaginationParams struct {
	Page    int `form:"page,default=1" binding:"min=1"`
	PerPage int `form:"per_page,default=10" binding:"min=1,max=100"`
	Offset  int `json:"-"`
}

func PaginationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var params PaginationParams

		// Bind query parameters
		if err := c.ShouldBindQuery(&params); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid pagination parameters",
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// Calculate offset
		params.Offset = (params.Page - 1) * params.PerPage

		// Set in context
		c.Set("pagination", params)

		// Set headers for pagination info
		c.Header("X-Page", fmt.Sprintf("%d", params.Page))
		c.Header("X-Per-Page", fmt.Sprintf("%d", params.PerPage))

		c.Next()
	}
}

// GetPagination extracts pagination from context
func GetPagination(c *gin.Context) PaginationParams {
	if pagination, exists := c.Get("pagination"); exists {
		return pagination.(PaginationParams)
	}
	return PaginationParams{Page: 1, PerPage: 10, Offset: 0}
}

// ContentTypeMiddleware validates content type for POST/PUT requests
func ContentTypeMiddleware(allowedTypes ...string) gin.HandlerFunc {
	allowed := make(map[string]bool)
	for _, contentType := range allowedTypes {
		allowed[contentType] = true
	}

	return func(c *gin.Context) {
		method := c.Request.Method
		if method == "POST" || method == "PUT" || method == "PATCH" {
			contentType := c.GetHeader("Content-Type")

			// Extract main content type (remove charset, boundary, etc.)
			if idx := strings.Index(contentType, ";"); idx != -1 {
				contentType = contentType[:idx]
			}
			contentType = strings.TrimSpace(contentType)

			if !allowed[contentType] {
				c.JSON(http.StatusUnsupportedMediaType, gin.H{
					"success": false,
					"message": "Unsupported content type",
					"allowed": allowedTypes,
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// MetricsMiddleware collects basic metrics
type Metrics struct {
	RequestCount    int64
	ErrorCount      int64
	TotalDuration   time.Duration
	AverageDuration time.Duration
}

var globalMetrics Metrics

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		// Update metrics
		globalMetrics.RequestCount++
		globalMetrics.TotalDuration += duration
		globalMetrics.AverageDuration = time.Duration(int64(globalMetrics.TotalDuration) / globalMetrics.RequestCount)

		if status >= 400 {
			globalMetrics.ErrorCount++
		}

		// Set metrics headers
		c.Header("X-Response-Time", duration.String())
	}
}

// GetMetrics returns current metrics
func GetMetrics() Metrics {
	return globalMetrics
}

// HealthCheckMiddleware provides health check endpoint
func HealthCheckMiddleware(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == path {
			c.JSON(http.StatusOK, gin.H{
				"status":    "healthy",
				"timestamp": time.Now().UTC(),
				"uptime":    time.Since(startTime),
				"service":   c.GetString("service_name"),
				"version":   c.GetString("service_version"),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

var startTime = time.Now()

// APIVersionMiddleware handles API versioning
func APIVersionMiddleware(supportedVersions []string, defaultVersion string) gin.HandlerFunc {
	supported := make(map[string]bool)
	for _, version := range supportedVersions {
		supported[version] = true
	}

	return func(c *gin.Context) {
		// Get version from header or query parameter
		version := c.GetHeader("API-Version")
		if version == "" {
			version = c.Query("version")
		}
		if version == "" {
			version = defaultVersion
		}

		// Validate version
		if !supported[version] {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":            false,
				"message":            "Unsupported API version",
				"supported_versions": supportedVersions,
			})
			c.Abort()
			return
		}

		// Set version in context
		c.Set("api_version", version)
		c.Header("API-Version", version)

		c.Next()
	}
}
