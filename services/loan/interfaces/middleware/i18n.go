package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"loan-service/pkg/i18n"
)

// I18nMiddleware handles internationalization for HTTP requests
type I18nMiddleware struct {
	localizer *i18n.Localizer
	logger    *zap.Logger
}

// NewI18nMiddleware creates a new i18n middleware
func NewI18nMiddleware(localizer *i18n.Localizer, logger *zap.Logger) *I18nMiddleware {
	return &I18nMiddleware{
		localizer: localizer,
		logger:    logger,
	}
}

// Handler returns the middleware handler function
func (m *I18nMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Detect language from multiple sources in priority order:
		// 1. Query parameter 'lang'
		// 2. Header 'X-Language'
		// 3. Accept-Language header
		// 4. Default to English

		var lang string

		// Check query parameter first
		if queryLang := c.Query("lang"); queryLang != "" {
			lang = queryLang
		} else if headerLang := c.GetHeader("X-Language"); headerLang != "" {
			lang = headerLang
		} else {
			// Parse Accept-Language header
			acceptLang := c.GetHeader("Accept-Language")
			lang = i18n.DetectLanguage(acceptLang)
		}

		// Validate and normalize language
		if lang != "en" && lang != "vi" {
			lang = "en" // Default to English
		}

		// Set language in context
		ctx := i18n.SetLanguageInContext(c.Request.Context(), lang)
		c.Request = c.Request.WithContext(ctx)

		// Store in Gin context for easy access
		c.Set("language", lang)
		c.Set("localizer", m.localizer)

		m.logger.Debug("Language detected",
			zap.String("language", lang),
			zap.String("request_id", c.GetString("request_id")),
		)

		c.Next()
	}
}

// ErrorResponse represents a standardized error response
// @Description Standardized error response
type ErrorResponse struct {
	Success  bool                   `json:"success" example:"false"`
	Data     interface{}            `json:"data"`
	Error    *ErrorDetail           `json:"error,omitempty"`
	Metadata map[string]interface{} `json:"metadata"`
}

// ErrorDetail contains detailed error information
// @Description Detailed error information with comprehensive error details
type ErrorDetail struct {
	Code        string                 `json:"code" example:"LOAN_020" description:"Error code identifier (e.g., LOAN_020)"`
	Message     string                 `json:"message" example:"Invalid request format - please check your JSON data and field validation" description:"Human-readable error message"`
	Description string                 `json:"description,omitempty" example:"Validation error: parsing time \"1990-01-01\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"\" as \"T\"; date_format: Date must be in ISO 8601 format (e.g., 1990-01-01T00:00:00Z)" description:"Detailed error description with specific validation errors and guidance"`
	Field       string                 `json:"field,omitempty" example:"date_of_birth" description:"Specific field that caused the error (if applicable)"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" description:"Additional error details including validation errors, field errors, and request body"`
}

// SuccessResponse represents a standardized success response
// @Description Standardized success response
type SuccessResponse struct {
	Success  bool                   `json:"success" example:"true"`
	Data     interface{}            `json:"data"`
	Message  string                 `json:"message,omitempty" example:"Application created successfully"`
	Metadata map[string]interface{} `json:"metadata"`
}

// CreateErrorResponse creates a localized error response
func CreateErrorResponse(c *gin.Context, statusCode int, errorCode string, templateData map[string]interface{}) {
	localizer, exists := c.Get("localizer")
	if !exists {
		c.JSON(statusCode, gin.H{"error": "Localizer not found"})
		return
	}

	loc := localizer.(*i18n.Localizer)
	message := loc.LocalizeError(c.Request.Context(), errorCode, templateData)

	// Create a more descriptive error message
	errorMessage := message
	if errorMessage == errorCode {
		// If localization failed, provide a fallback message
		switch errorCode {
		case "LOAN_020":
			errorMessage = "Invalid request format - please check your JSON data and field validation"
		case "LOAN_021":
			errorMessage = "User not found"
		case "LOAN_022":
			errorMessage = "Unauthorized access"
		case "LOAN_023":
			errorMessage = "Database connection error"
		default:
			errorMessage = "An error occurred while processing your request"
		}
	}

	response := ErrorResponse{
		Success: false,
		Data:    nil,
		Error: &ErrorDetail{
			Code:        errorCode,
			Message:     errorMessage,
			Description: getErrorDescription(errorCode, templateData),
			Metadata:    templateData,
		},
		Metadata: map[string]interface{}{
			"request_id": c.GetHeader("X-Request-ID"),
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"version":    "v1",
			"service":    "loan-service",
		},
	}

	c.Header("Content-Type", "application/json")
	c.JSON(statusCode, response)
}

// CreateSuccessResponse creates a localized success response
func CreateSuccessResponse(c *gin.Context, data interface{}, successKey string, templateData map[string]interface{}) {
	localizer, exists := c.Get("localizer")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"data": data})
		return
	}

	loc := localizer.(*i18n.Localizer)
	message := ""
	if successKey != "" {
		message = loc.Localize(c.Request.Context(), successKey, templateData)
	}

	response := SuccessResponse{
		Success: true,
		Data:    data,
		Message: message,
		Metadata: map[string]interface{}{
			"request_id": c.GetHeader("X-Request-ID"),
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"version":    "v1",
			"service":    "loan-service",
		},
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, response)
}

// CreateValidationErrorResponse creates a localized validation error response
func CreateValidationErrorResponse(c *gin.Context, field string, errorCode string, templateData map[string]interface{}) {
	localizer, exists := c.Get("localizer")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed"})
		return
	}

	loc := localizer.(*i18n.Localizer)
	message := loc.LocalizeError(c.Request.Context(), errorCode, templateData)

	response := ErrorResponse{
		Success: false,
		Data:    nil,
		Error: &ErrorDetail{
			Code:     errorCode,
			Message:  message,
			Field:    field,
			Metadata: templateData,
		},
		Metadata: map[string]interface{}{
			"request_id": c.GetHeader("X-Request-ID"),
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
			"version":    "v1",
			"service":    "loan-service",
		},
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusBadRequest, response)
}

// GetLocalizer helper function to get localizer from context
func GetLocalizer(c *gin.Context) *i18n.Localizer {
	localizer, exists := c.Get("localizer")
	if !exists {
		return nil
	}
	return localizer.(*i18n.Localizer)
}

// GetLanguage helper function to get language from context
func GetLanguage(c *gin.Context) string {
	lang, exists := c.Get("language")
	if !exists {
		return "en"
	}
	return lang.(string)
}

// LocalizeMessage helper function to localize a message
func LocalizeMessage(c *gin.Context, messageID string, templateData map[string]interface{}) string {
	localizer := GetLocalizer(c)
	if localizer == nil {
		return messageID
	}
	return localizer.Localize(c.Request.Context(), messageID, templateData)
}

// RequestIDMiddleware adds a request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
		}
		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// getErrorDescription generates a descriptive error message based on error code and template data
func getErrorDescription(errorCode string, templateData map[string]interface{}) string {
	if templateData == nil {
		return ""
	}

	// Build description from available template data
	var descriptions []string

	if validationError, ok := templateData["validation_error"].(string); ok && validationError != "" {
		descriptions = append(descriptions, fmt.Sprintf("Validation error: %s", validationError))
	}

	if fieldErrors, ok := templateData["field_errors"].(map[string]string); ok && len(fieldErrors) > 0 {
		for field, error := range fieldErrors {
			descriptions = append(descriptions, fmt.Sprintf("%s: %s", field, error))
		}
	}

	if requestBody, ok := templateData["request_body"].(string); ok && requestBody != "" {
		descriptions = append(descriptions, fmt.Sprintf("Request body: %s", requestBody))
	}

	if len(descriptions) > 0 {
		return strings.Join(descriptions, "; ")
	}

	return ""
}
