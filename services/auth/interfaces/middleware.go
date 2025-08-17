package interfaces

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/huuhoait/los-demo/services/auth/domain"
	customI18n "// github.com/huuhoait/los-demo/services/shared/pkg/i18n // DISABLED"
)

// AuthMiddleware handles JWT authentication for protected routes
type AuthMiddleware struct {
	authService domain.AuthService
	logger      *zap.Logger
	localizer   *customI18n.Localizer // Use custom i18n Localizer
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService domain.AuthService, logger *zap.Logger, localizer *customI18n.Localizer) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
		localizer:   localizer,
	}
}

// RequireAuth middleware validates JWT tokens and sets user context
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := m.logger.With(
			zap.String("operation", "require_auth"),
			zap.String("path", c.Request.URL.Path),
			zap.String("method", c.Request.Method),
		)

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn("Missing authorization header")
			m.respondWithError(c, http.StatusUnauthorized, domain.AUTH_004,
				"Authorization header is required")
			return
		}

		// Check Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("Invalid authorization header format")
			m.respondWithError(c, http.StatusUnauthorized, domain.AUTH_004,
				"Invalid authorization header format")
			return
		}

		token := parts[1]
		if token == "" {
			logger.Warn("Empty token in authorization header")
			m.respondWithError(c, http.StatusUnauthorized, domain.AUTH_004,
				"Token is required")
			return
		}

		// Validate token
		authContext, err := m.authService.ValidateAccessToken(c.Request.Context(), token)
		if err != nil {
			if authErr, ok := err.(*domain.AuthError); ok {
				logger.Warn("Token validation failed",
					zap.String("error_code", authErr.Code),
					zap.String("error_message", authErr.Message))

				statusCode := http.StatusUnauthorized
				if authErr.Code == domain.AUTH_005 { // Token expired
					statusCode = http.StatusUnauthorized
				}

				m.respondWithError(c, statusCode, authErr.Code,
					m.localizer.LocalizeError(customI18n.GetLanguageFromContext(c.Request.Context()), authErr.Code, nil))
				return
			}

			logger.Error("Unexpected error during token validation", zap.Error(err))
			m.respondWithError(c, http.StatusInternalServerError, domain.AUTH_017,
				m.localizer.LocalizeError(customI18n.GetLanguageFromContext(c.Request.Context()), domain.AUTH_017, nil))
			return
		}

		// Set user context
		c.Set("user_id", authContext.UserID)
		c.Set("user_email", authContext.Email)
		c.Set("user_role", authContext.Role)
		c.Set("session_id", authContext.SessionID)
		c.Set("auth_context", authContext)

		logger.Debug("Authentication successful",
			zap.String("user_id", authContext.UserID),
			zap.String("role", authContext.Role))

		c.Next()
	}
}

// RequireRole middleware checks if the authenticated user has the required role
func (m *AuthMiddleware) RequireRole(requiredRole domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := m.logger.With(
			zap.String("operation", "require_role"),
			zap.String("required_role", string(requiredRole)),
		)

		userRole, exists := c.Get("user_role")
		if !exists {
			logger.Error("User role not found in context")
			m.respondWithError(c, http.StatusInternalServerError, domain.AUTH_017,
				m.localizer.LocalizeError(customI18n.GetLanguageFromContext(c.Request.Context()), domain.AUTH_017, nil))
			return
		}

		role := domain.UserRole(userRole.(string))
		if role != requiredRole {
			logger.Warn("Insufficient role",
				zap.String("user_role", string(role)),
				zap.String("required_role", string(requiredRole)))
			m.respondWithError(c, http.StatusForbidden, domain.AUTH_015,
				m.localizer.LocalizePermission(customI18n.GetLanguageFromContext(c.Request.Context()), "insufficient_role", map[string]interface{}{
					"required_role": string(requiredRole),
					"user_role":     string(role),
				}))
			return
		}

		logger.Debug("Role check passed")
		c.Next()
	}
}

// RequirePermission middleware checks if the authenticated user has the required permission
func (m *AuthMiddleware) RequirePermission(permission domain.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := m.logger.With(
			zap.String("operation", "require_permission"),
			zap.String("required_permission", string(permission)),
		)

		userRole, exists := c.Get("user_role")
		if !exists {
			logger.Error("User role not found in context")
			m.respondWithError(c, http.StatusInternalServerError, domain.AUTH_017,
				m.localizer.LocalizeError(customI18n.GetLanguageFromContext(c.Request.Context()), domain.AUTH_017, nil))
			return
		}

		role := domain.UserRole(userRole.(string))
		if !role.HasPermission(permission) {
			logger.Warn("Insufficient permissions",
				zap.String("user_role", string(role)),
				zap.String("required_permission", string(permission)))
			m.respondWithError(c, http.StatusForbidden, domain.AUTH_015,
				m.localizer.LocalizePermission(customI18n.GetLanguageFromContext(c.Request.Context()), "insufficient_permission", map[string]interface{}{
					"required_permission": string(permission),
					"user_role":           string(role),
				}))
			return
		}

		logger.Debug("Permission check passed")
		c.Next()
	}
}

// OptionalAuth middleware validates tokens if present but doesn't require them
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Try to authenticate if token is present
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" && parts[1] != "" {
			authContext, err := m.authService.ValidateAccessToken(c.Request.Context(), parts[1])
			if err == nil {
				// Set context if authentication succeeds
				c.Set("user_id", authContext.UserID)
				c.Set("user_email", authContext.Email)
				c.Set("user_role", authContext.Role)
				c.Set("session_id", authContext.SessionID)
				c.Set("auth_context", authContext)
			}
		}

		c.Next()
	}
}

// HTTPSignatureAuth middleware validates HTTP signatures
func (m *AuthMiddleware) HTTPSignatureAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := m.logger.With(
			zap.String("operation", "http_signature_auth"),
			zap.String("path", c.Request.URL.Path),
		)

		signatureHeader := c.GetHeader("Signature")

		if signatureHeader == "" {
			logger.Warn("Missing signature header")
			m.respondWithError(c, http.StatusUnauthorized, domain.AUTH_012,
				m.localizer.LocalizeSecurity(customI18n.GetLanguageFromContext(c.Request.Context()), "missing_signature_header", nil))
			return
		}

		// Read request body for signature validation
		body, err := c.GetRawData()
		if err != nil {
			logger.Error("Failed to read request body", zap.Error(err))
			m.respondWithError(c, http.StatusBadRequest, domain.AUTH_020,
				m.localizer.LocalizeError(customI18n.GetLanguageFromContext(c.Request.Context()), domain.AUTH_020, nil))
			return
		}

		// Validate HTTP signature
		err = m.authService.ValidateHTTPSignature(c.Request.Context(),
			signatureHeader, "dummy_key_id", body)
		if err != nil {
			if authErr, ok := err.(*domain.AuthError); ok {
				logger.Warn("HTTP signature validation failed",
					zap.String("error_code", authErr.Code))
				m.respondWithError(c, http.StatusUnauthorized, authErr.Code,
					m.localizer.LocalizeSecurity(customI18n.GetLanguageFromContext(c.Request.Context()), "signature_validation_failed", nil))
				return
			}

			logger.Error("Unexpected error during signature validation", zap.Error(err))
			m.respondWithError(c, http.StatusInternalServerError, domain.AUTH_017,
				m.localizer.LocalizeError(customI18n.GetLanguageFromContext(c.Request.Context()), domain.AUTH_017, nil))
			return
		}

		logger.Debug("HTTP signature validation successful")
		c.Next()
	}
}

// RateLimit middleware implements rate limiting per IP or user
func (m *AuthMiddleware) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := m.logger.With(
			zap.String("operation", "rate_limit"),
			zap.String("client_ip", c.ClientIP()),
		)

		// Use user ID if authenticated, otherwise use IP
		identifier := c.ClientIP()
		if userID, exists := c.Get("user_id"); exists {
			identifier = userID.(string)
		}

		err := m.authService.CheckRateLimit(c.Request.Context(), identifier)
		if err != nil {
			if authErr, ok := err.(*domain.AuthError); ok {
				logger.Warn("Rate limit exceeded",
					zap.String("identifier", identifier),
					zap.String("error_code", authErr.Code))
				m.respondWithError(c, http.StatusTooManyRequests, authErr.Code,
					m.localizer.LocalizeSecurity(customI18n.GetLanguageFromContext(c.Request.Context()), "rate_limit_exceeded", nil))
				return
			}

			logger.Error("Unexpected error during rate limiting", zap.Error(err))
			// Continue on rate limiting errors to avoid blocking legitimate requests
		}

		logger.Debug("Rate limit check passed")
		c.Next()
	}
}

// GetAuthContext helper function to extract auth context from Gin context
func GetAuthContext(c *gin.Context) (*domain.AuthContext, bool) {
	if authContext, exists := c.Get("auth_context"); exists {
		if ctx, ok := authContext.(*domain.AuthContext); ok {
			return ctx, true
		}
	}
	return nil, false
}

// GetUserID helper function to extract user ID from Gin context
func GetUserID(c *gin.Context) (string, bool) {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id, true
		}
	}
	return "", false
}

// GetUserRole helper function to extract user role from Gin context
func GetUserRole(c *gin.Context) (domain.UserRole, bool) {
	if userRole, exists := c.Get("user_role"); exists {
		if role, ok := userRole.(string); ok {
			return domain.UserRole(role), true
		}
	}
	return "", false
}

// respondWithError sends a standardized error response
func (m *AuthMiddleware) respondWithError(c *gin.Context, statusCode int, errorCode string, message string) {
	// Use the localizer to get the message if it's an error code
	localizedMessage := m.localizer.LocalizeError(customI18n.GetLanguageFromContext(c.Request.Context()), errorCode, nil)
	if localizedMessage == errorCode {
		// Fallback to provided message if localization fails
		localizedMessage = message
	}

	response := gin.H{
		"success": false,
		"data":    nil,
		"metadata": gin.H{
			"request_id": c.GetHeader("X-Request-ID"),
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"version":    "v1",
			"service":    "auth-service",
		},
		"error": gin.H{
			"code":        errorCode,
			"message":     localizedMessage,
			"description": m.getErrorDescription(errorCode), // Still use the old description for now
		},
	}

	c.Header("Content-Type", "application/json")
	c.JSON(statusCode, response)
	c.Abort()
}

// getErrorDescription returns localized error description
func (m *AuthMiddleware) getErrorDescription(errorCode string) string {
	// This function can be removed if all error descriptions are handled by i18n
	// For now, it provides a fallback or additional context
	descriptions := map[string]string{
		domain.AUTH_001: "The provided credentials are invalid",
		domain.AUTH_002: "Account is temporarily locked",
		domain.AUTH_003: "Account is disabled",
		domain.AUTH_004: "Token is invalid or malformed",
		domain.AUTH_005: "Token has expired",
		domain.AUTH_006: "Token has been revoked",
		domain.AUTH_007: "Refresh token is invalid",
		domain.AUTH_008: "Session not found",
		domain.AUTH_009: "Session has expired",
		domain.AUTH_010: "Too many requests - rate limit exceeded",
		domain.AUTH_011: "HTTP signature validation failed",
		domain.AUTH_012: "Required signature header is missing",
		domain.AUTH_015: "User does not have required permissions",
		domain.AUTH_017: "Internal authentication service error",
		domain.AUTH_020: "Request format is invalid",
	}

	if desc, exists := descriptions[errorCode]; exists {
		return desc
	}
	return "An authentication error occurred"
}
