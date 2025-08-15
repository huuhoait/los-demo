package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/lendingplatform/los/services/auth/pkg/i18n"
)

// I18nMiddleware provides internationalization support for HTTP requests
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

// Handler returns a Gin middleware function for internationalization
func (m *I18nMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get language from various sources (in order of priority)
		lang := m.detectLanguage(c)

		// Validate language
		if !m.localizer.IsLanguageSupported(lang) {
			lang = "en" // fallback to English
		}

		// Add language to context
		ctx := i18n.SetLanguageInContext(c.Request.Context(), lang)
		
		// Get localizer for this language
		localizer := m.localizer.GetLocalizer(lang)
		ctx = i18n.SetLocalizerInContext(ctx, localizer)

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		// Set language in response header for client reference
		c.Header("Content-Language", lang)

		// Add helper functions to Gin context
		c.Set("lang", lang)
		c.Set("localizer", localizer)
		c.Set("localize", func(messageID string, templateData map[string]interface{}) string {
			return m.localizer.Localize(lang, messageID, templateData)
		})

		m.logger.Debug("Language detected",
			zap.String("language", lang),
			zap.String("path", c.Request.URL.Path),
		)

		c.Next()
	}
}

// detectLanguage detects the preferred language from request
func (m *I18nMiddleware) detectLanguage(c *gin.Context) string {
	// 1. Check query parameter
	if lang := c.Query("lang"); lang != "" {
		return lang
	}

	// 2. Check header parameter
	if lang := c.GetHeader("X-Language"); lang != "" {
		return lang
	}

	// 3. Check Accept-Language header
	if acceptLang := c.GetHeader("Accept-Language"); acceptLang != "" {
		return i18n.DetectLanguageFromHeader(acceptLang)
	}

	// 4. Check user preference from context/session (if available)
	if userID := c.GetString("user_id"); userID != "" {
		// Here you could retrieve user's preferred language from database
		// For now, we'll skip this step
	}

	// 5. Default to English
	return "en"
}

// GetLanguageFromGinContext extracts language from Gin context
func GetLanguageFromGinContext(c *gin.Context) string {
	if lang, exists := c.Get("lang"); exists {
		if langStr, ok := lang.(string); ok {
			return langStr
		}
	}
	return "en"
}

// GetLocalizerFromGinContext extracts localizer from Gin context
func GetLocalizerFromGinContext(c *gin.Context) *i18n.Localizer {
	if localizer, exists := c.Get("localizer"); exists {
		if loc, ok := localizer.(*i18n.Localizer); ok {
			return loc
		}
	}
	return nil
}

// LocalizeError is a helper function to localize error messages in handlers
func LocalizeError(c *gin.Context, localizer *i18n.Localizer, errorCode string, templateData map[string]interface{}) string {
	lang := GetLanguageFromGinContext(c)
	return localizer.LocalizeError(lang, errorCode, templateData)
}

// LocalizeMessage is a helper function to localize messages in handlers
func LocalizeMessage(c *gin.Context, localizer *i18n.Localizer, messageKey string, templateData map[string]interface{}) string {
	lang := GetLanguageFromGinContext(c)
	return localizer.LocalizeMessage(lang, messageKey, templateData)
}

// LocalizeValidation is a helper function to localize validation messages in handlers
func LocalizeValidation(c *gin.Context, localizer *i18n.Localizer, validationKey string, templateData map[string]interface{}) string {
	lang := GetLanguageFromGinContext(c)
	return localizer.LocalizeValidation(lang, validationKey, templateData)
}

// LocalizeUI is a helper function to localize UI elements in handlers
func LocalizeUI(c *gin.Context, localizer *i18n.Localizer, uiKey string, templateData map[string]interface{}) string {
	lang := GetLanguageFromGinContext(c)
	return localizer.LocalizeUI(lang, uiKey, templateData)
}

// ErrorResponse represents a localized error response
type ErrorResponse struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	Language  string      `json:"language"`
	Timestamp string      `json:"timestamp"`
}

// CreateErrorResponse creates a localized error response
func CreateErrorResponse(c *gin.Context, localizer *i18n.Localizer, errorCode string, details interface{}, templateData map[string]interface{}) ErrorResponse {
	lang := GetLanguageFromGinContext(c)
	message := localizer.LocalizeError(lang, errorCode, templateData)
	
	return ErrorResponse{
		Code:      errorCode,
		Message:   message,
		Details:   details,
		Language:  lang,
		Timestamp: c.GetString("timestamp"),
	}
}

// SuccessResponse represents a localized success response
type SuccessResponse struct {
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Language  string      `json:"language"`
	Timestamp string      `json:"timestamp"`
}

// CreateSuccessResponse creates a localized success response
func CreateSuccessResponse(c *gin.Context, localizer *i18n.Localizer, messageKey string, data interface{}, templateData map[string]interface{}) SuccessResponse {
	lang := GetLanguageFromGinContext(c)
	message := localizer.LocalizeMessage(lang, messageKey, templateData)
	
	return SuccessResponse{
		Message:   message,
		Data:      data,
		Language:  lang,
		Timestamp: c.GetString("timestamp"),
	}
}

// ValidationErrorResponse represents localized validation errors
type ValidationErrorResponse struct {
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Errors    map[string]string      `json:"errors"`
	Language  string                 `json:"language"`
	Timestamp string                 `json:"timestamp"`
}

// CreateValidationErrorResponse creates a localized validation error response
func CreateValidationErrorResponse(c *gin.Context, localizer *i18n.Localizer, validationErrors map[string]string) ValidationErrorResponse {
	lang := GetLanguageFromGinContext(c)
	
	// Localize validation errors
	localizedErrors := make(map[string]string)
	for field, errorKey := range validationErrors {
		localizedErrors[field] = localizer.LocalizeValidation(lang, errorKey, nil)
	}
	
	message := localizer.LocalizeMessage(lang, "validation_failed", nil)
	
	return ValidationErrorResponse{
		Code:      "VALIDATION_ERROR",
		Message:   message,
		Errors:    localizedErrors,
		Language:  lang,
		Timestamp: c.GetString("timestamp"),
	}
}
