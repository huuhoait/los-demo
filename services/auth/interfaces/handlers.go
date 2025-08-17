package interfaces

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/huuhoait/los-demo/services/auth/domain"
	"github.com/huuhoait/los-demo/services/auth/interfaces/middleware"
	"github.com/huuhoait/los-demo/services/auth/pkg/i18n"
)

// AuthHandler handles authentication HTTP endpoints
type AuthHandler struct {
	authService domain.AuthService
	logger      *zap.Logger
	localizer   *i18n.Localizer
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService domain.AuthService, logger *zap.Logger, localizer *i18n.Localizer) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
		localizer:   localizer,
	}
}

// Login handles user login requests
// POST /v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "login"),
		zap.String("ip_address", c.ClientIP()),
	)

	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid login request format", zap.Error(err))
		h.respondWithError(c, http.StatusBadRequest, domain.AUTH_020, nil)
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Attempt login
	tokenResponse, err := h.authService.Login(c.Request.Context(), req.Email, req.Password, ipAddress, userAgent)
	if err != nil {
		if authErr, ok := err.(*domain.AuthError); ok {
			logger.Warn("Login failed",
				zap.String("email", req.Email),
				zap.String("error_code", authErr.Code))

			statusCode := http.StatusUnauthorized
			if authErr.Code == domain.AUTH_002 || authErr.Code == domain.AUTH_010 {
				statusCode = http.StatusTooManyRequests
			}

			h.respondWithError(c, statusCode, authErr.Code, nil)
			return
		}

		logger.Error("Unexpected error during login", zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, domain.AUTH_017, nil)
		return
	}

	logger.Info("Login successful",
		zap.String("user_id", tokenResponse.User.ID),
		zap.String("email", req.Email))

	h.respondWithSuccess(c, tokenResponse, "LOGIN_SUCCESS", nil)
}

// RefreshToken handles token refresh requests
// POST /v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "refresh_token"),
		zap.String("ip_address", c.ClientIP()),
	)

	var req domain.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid refresh request format", zap.Error(err))
		h.respondWithError(c, http.StatusBadRequest, domain.AUTH_020, nil)
		return
	}

	// Get client info
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Refresh token
	tokenResponse, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken, ipAddress, userAgent)
	if err != nil {
		if authErr, ok := err.(*domain.AuthError); ok {
			logger.Warn("Token refresh failed",
				zap.String("error_code", authErr.Code))

			statusCode := http.StatusUnauthorized
			if authErr.Code == domain.AUTH_007 || authErr.Code == domain.AUTH_009 {
				statusCode = http.StatusUnauthorized
			}

			h.respondWithError(c, statusCode, authErr.Code, nil)
			return
		}

		logger.Error("Unexpected error during token refresh", zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, domain.AUTH_017, nil)
		return
	}

	logger.Info("Token refresh successful", zap.String("user_id", tokenResponse.User.ID))
	h.respondWithSuccess(c, tokenResponse, "REFRESH_SUCCESS", nil)
}

// Logout handles user logout requests
// POST /v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "logout"),
	)

	userID, exists := GetUserID(c)
	if !exists {
		logger.Error("User ID not found in context")
		h.respondWithError(c, http.StatusInternalServerError, domain.AUTH_017, nil)
		return
	}

	sessionID, exists := c.Get("session_id")
	if !exists {
		logger.Error("Session ID not found in context")
		h.respondWithError(c, http.StatusInternalServerError, domain.AUTH_017, nil)
		return
	}

	// Logout user
	err := h.authService.Logout(c.Request.Context(), userID, sessionID.(string))
	if err != nil {
		if authErr, ok := err.(*domain.AuthError); ok {
			logger.Warn("Logout failed",
				zap.String("user_id", userID),
				zap.String("error_code", authErr.Code))
			h.respondWithError(c, http.StatusInternalServerError, authErr.Code, nil)
			return
		}

		logger.Error("Unexpected error during logout", zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, domain.AUTH_017, nil)
		return
	}

	logger.Info("Logout successful", zap.String("user_id", userID))
	h.respondWithSuccess(c, nil, "LOGOUT_SUCCESS", nil)
}

// LogoutAll handles logout from all devices
// POST /v1/auth/logout-all
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "logout_all"),
	)

	userID, exists := GetUserID(c)
	if !exists {
		logger.Error("User ID not found in context")
		h.respondWithError(c, http.StatusInternalServerError, domain.AUTH_017, nil)
		return
	}

	// Logout from all devices
	err := h.authService.LogoutAll(c.Request.Context(), userID)
	if err != nil {
		if authErr, ok := err.(*domain.AuthError); ok {
			logger.Warn("Logout all failed",
				zap.String("user_id", userID),
				zap.String("error_code", authErr.Code))
			h.respondWithError(c, http.StatusInternalServerError, authErr.Code, nil)
			return
		}

		logger.Error("Unexpected error during logout all", zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, domain.AUTH_017, nil)
		return
	}

	logger.Info("Logout all successful", zap.String("user_id", userID))
	h.respondWithSuccess(c, nil, "LOGOUT_ALL_SUCCESS", nil)
}

// GetProfile handles get current user profile requests
// GET /v1/auth/me
func (h *AuthHandler) GetProfile(c *gin.Context) {
	logger := h.logger.With(
		zap.String("operation", "get_profile"),
	)

	userID, exists := GetUserID(c)
	if !exists {
		logger.Error("User ID not found in context")
		h.respondWithError(c, http.StatusInternalServerError, domain.AUTH_017, nil)
		return
	}

	// Get user profile
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		if authErr, ok := err.(*domain.AuthError); ok {
			logger.Warn("Get profile failed",
				zap.String("user_id", userID),
				zap.String("error_code", authErr.Code))

			statusCode := http.StatusNotFound
			if authErr.Code == domain.AUTH_017 {
				statusCode = http.StatusInternalServerError
			}

			h.respondWithError(c, statusCode, authErr.Code, nil)
			return
		}

		logger.Error("Unexpected error during get profile", zap.Error(err))
		h.respondWithError(c, http.StatusInternalServerError, domain.AUTH_017, nil)
		return
	}

	logger.Debug("Get profile successful", zap.String("user_id", userID))
	h.respondWithSuccess(c, user, "PROFILE_SUCCESS", nil)
}

// Health check endpoint
// GET /v1/auth/health
func (h *AuthHandler) Health(c *gin.Context) {
	h.respondWithSuccess(c, gin.H{
		"status":    "healthy",
		"service":   "auth-service",
		"version":   "v1.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, "HEALTH_SUCCESS", nil)
}

// Helper methods

// respondWithError sends a standardized localized error response
func (h *AuthHandler) respondWithError(c *gin.Context, statusCode int, errorCode string, data map[string]interface{}) {
	middleware.CreateErrorResponse(c, h.localizer, errorCode, data, nil)
}

// respondWithSuccess sends a standardized localized success response
func (h *AuthHandler) respondWithSuccess(c *gin.Context, data interface{}, successKey string, templateData map[string]interface{}) {
	middleware.CreateSuccessResponse(c, h.localizer, successKey, data, templateData)
}

// RegisterRoutes registers all authentication routes
func (h *AuthHandler) RegisterRoutes(router *gin.RouterGroup, authMiddleware *AuthMiddleware) {
	// Public routes (no authentication required)
	router.POST("/login", h.Login)
	router.POST("/refresh", h.RefreshToken)
	router.GET("/health", h.Health)

	// Protected routes (authentication required)
	protected := router.Group("")
	protected.Use(authMiddleware.RequireAuth())
	{
		protected.POST("/logout", h.Logout)
		protected.POST("/logout-all", h.LogoutAll)
		protected.GET("/me", h.GetProfile)
	}
}
