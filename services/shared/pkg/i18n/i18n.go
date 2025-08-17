package i18n

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Simple Localizer for loan-api compatibility
type Localizer struct {
	language string
}

// NewLocalizer creates a new simple localizer
func NewLocalizer() (*Localizer, error) {
	return &Localizer{
		language: "en",
	}, nil
}

// DetectLanguage detects language from Accept-Language header
func DetectLanguage(acceptLanguage string) string {
	if acceptLanguage == "" {
		return "en"
	}

	// Simple language detection
	if strings.Contains(acceptLanguage, "vi") {
		return "vi"
	}
	
	return "en"
}

// SetLanguageInContext sets language in context
func SetLanguageInContext(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, "language", lang)
}

// GetLanguageFromContext gets language from context
func GetLanguageFromContext(ctx context.Context) string {
	if lang := ctx.Value("language"); lang != nil {
		return lang.(string)
	}
	return "en"
}

// Localize translates a message
func (l *Localizer) Localize(ctx context.Context, messageID string, templateData map[string]interface{}) string {
	// For now, return the messageID as a fallback
	// In a real implementation, this would load translations from files
	
	// Basic error message translations
	switch messageID {
	case "error.LOAN_020":
		return "Invalid request format - please check your JSON data and field validation"
	case "error.LOAN_021":
		return "User not found"
	case "error.LOAN_022":
		return "Unauthorized access"
	case "error.LOAN_023":
		return "Database connection error"
	case "error.validation_failed":
		return "Validation failed"
	case "error.invalid_request":
		return "Invalid request"
	case "error.internal_server_error":
		return "Internal server error"
	case "success.application_created":
		return "Application created successfully"
	case "success.application_updated":
		return "Application updated successfully"
	default:
		return messageID
	}
}

// LocalizeError translates an error message
func (l *Localizer) LocalizeError(ctx context.Context, errorCode string, templateData map[string]interface{}) string {
	if !strings.HasPrefix(errorCode, "error.") {
		errorCode = "error." + errorCode
	}
	return l.Localize(ctx, errorCode, templateData)
}

// LocalizeSuccess translates a success message
func (l *Localizer) LocalizeSuccess(ctx context.Context, messageCode string, templateData map[string]interface{}) string {
	if !strings.HasPrefix(messageCode, "success.") {
		messageCode = "success." + messageCode
	}
	return l.Localize(ctx, messageCode, templateData)
}

// Helper functions for Gin context

// GetLocalizer gets localizer from Gin context
func GetLocalizer(c *gin.Context) *Localizer {
	if localizer, exists := c.Get("localizer"); exists {
		return localizer.(*Localizer)
	}
	return &Localizer{language: "en"}
}

// GetLanguage gets current language from Gin context
func GetLanguage(c *gin.Context) string {
	if lang, exists := c.Get("language"); exists {
		return lang.(string)
	}
	return "en"
}

// T is a helper function for quick translation
func T(c *gin.Context, messageID string, templateData ...map[string]interface{}) string {
	localizer := GetLocalizer(c)
	if localizer == nil {
		return messageID
	}

	var data map[string]interface{}
	if len(templateData) > 0 {
		data = templateData[0]
	}

	return localizer.Localize(c.Request.Context(), messageID, data)
}

// TWithDefault is a helper function for translation with default
func TWithDefault(c *gin.Context, messageID, defaultMessage string, templateData ...map[string]interface{}) string {
	localizer := GetLocalizer(c)
	if localizer == nil {
		return defaultMessage
	}

	var data map[string]interface{}
	if len(templateData) > 0 {
		data = templateData[0]
	}

	result := localizer.Localize(c.Request.Context(), messageID, data)
	if result == messageID {
		return defaultMessage
	}
	return result
}

// TCtx is a helper function for translation from context
func TCtx(ctx context.Context, messageID string, templateData ...map[string]interface{}) string {
	localizer := &Localizer{language: GetLanguageFromContext(ctx)}

	var data map[string]interface{}
	if len(templateData) > 0 {
		data = templateData[0]
	}

	return localizer.Localize(ctx, messageID, data)
}

// Response helpers

// ErrorResponse creates a localized error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// SuccessResponse creates a localized success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RespondWithError sends a localized error response
func RespondWithError(c *gin.Context, statusCode int, errorCode string, templateData ...map[string]interface{}) {
	var data map[string]interface{}
	if len(templateData) > 0 {
		data = templateData[0]
	}

	message := T(c, "error."+errorCode, data)

	c.JSON(statusCode, ErrorResponse{
		Success: false,
		Message: message,
		Code:    errorCode,
	})
}

// RespondWithSuccess sends a localized success response
func RespondWithSuccess(c *gin.Context, messageCode string, responseData interface{}, templateData ...map[string]interface{}) {
	var data map[string]interface{}
	if len(templateData) > 0 {
		data = templateData[0]
	}

	message := T(c, "success."+messageCode, data)

	c.JSON(http.StatusOK, SuccessResponse{
		Success: true,
		Message: message,
		Data:    responseData,
	})
}

// Common error codes
const (
	ErrInvalidRequest     = "invalid_request"
	ErrUnauthorized       = "unauthorized"
	ErrForbidden          = "forbidden"
	ErrNotFound           = "not_found"
	ErrInternalServer     = "internal_server_error"
	ErrValidationFailed   = "validation_failed"
	ErrDuplicateResource  = "duplicate_resource"
	ErrResourceNotFound   = "resource_not_found"
	ErrInvalidCredentials = "invalid_credentials"
	ErrTokenExpired       = "token_expired"
	ErrRateLimitExceeded  = "rate_limit_exceeded"
)

// Common success codes
const (
	SuccessCreated = "created"
	SuccessUpdated = "updated"
	SuccessDeleted = "deleted"
	SuccessLogin   = "login"
	SuccessLogout  = "logout"
)
