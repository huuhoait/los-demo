package i18n

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
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

// Config holds i18n configuration
type Config struct {
	DefaultLanguage string   `yaml:"default_language" json:"default_language"`
	Languages       []string `yaml:"languages" json:"languages"`
	BundlePath      string   `yaml:"bundle_path" json:"bundle_path"`
}

// Manager manages internationalization
type Manager struct {
	bundle          *i18n.Bundle
	defaultLanguage string
	supportedLangs  map[string]bool
}

// NewManager creates a new i18n manager
func NewManager(config Config) (*Manager, error) {
	// Create bundle
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)

	// Load message files
	if config.BundlePath != "" {
		for _, lang := range config.Languages {
			messageFile := filepath.Join(config.BundlePath, lang+".toml")
			if _, err := bundle.LoadMessageFile(messageFile); err != nil {
				// Try YAML format
				messageFile = filepath.Join(config.BundlePath, lang+".yaml")
				if _, err := bundle.LoadMessageFile(messageFile); err != nil {
					return nil, fmt.Errorf("failed to load message file for language %s: %w", lang, err)
				}
			}
		}
	}

	// Create supported languages map
	supportedLangs := make(map[string]bool)
	for _, lang := range config.Languages {
		supportedLangs[lang] = true
	}

	return &Manager{
		bundle:          bundle,
		defaultLanguage: config.DefaultLanguage,
		supportedLangs:  supportedLangs,
	}, nil
}

// GetLocalizer returns a localizer for the given language
func (m *Manager) GetLocalizer(lang string) *i18n.Localizer {
	// Use default language if not supported
	if !m.supportedLangs[lang] {
		lang = m.defaultLanguage
	}

	return i18n.NewLocalizer(m.bundle, lang)
}

// GetLocalizerFromAcceptLanguage returns a localizer based on Accept-Language header
func (m *Manager) GetLocalizerFromAcceptLanguage(acceptLanguage string) *i18n.Localizer {
	// Parse Accept-Language header
	lang := m.parseAcceptLanguage(acceptLanguage)
	return m.GetLocalizer(lang)
}

// parseAcceptLanguage parses Accept-Language header and returns the best match
func (m *Manager) parseAcceptLanguage(acceptLanguage string) string {
	if acceptLanguage == "" {
		return m.defaultLanguage
	}

	// Split by comma and check each language
	languages := strings.Split(acceptLanguage, ",")
	for _, lang := range languages {
		// Remove quality value (e.g., "en-US;q=0.9" -> "en-US")
		lang = strings.TrimSpace(strings.Split(lang, ";")[0])

		// Check exact match
		if m.supportedLangs[lang] {
			return lang
		}

		// Check base language (e.g., "en-US" -> "en")
		if parts := strings.Split(lang, "-"); len(parts) > 1 {
			baseLang := parts[0]
			if m.supportedLangs[baseLang] {
				return baseLang
			}
		}
	}

	return m.defaultLanguage
}

// Localize translates a message
func (m *Manager) Localize(localizer *i18n.Localizer, messageID string, templateData map[string]interface{}) string {
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil {
		// Return message ID if translation fails
		return messageID
	}
	return msg
}

// LocalizeWithDefault translates a message with a default value
func (m *Manager) LocalizeWithDefault(localizer *i18n.Localizer, messageID, defaultMessage string, templateData map[string]interface{}) string {
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:      messageID,
		DefaultMessage: &i18n.Message{ID: messageID, Other: defaultMessage},
		TemplateData:   templateData,
	})
	if err != nil {
		return defaultMessage
	}
	return msg
}

// LocalizeError translates an error message
func (m *Manager) LocalizeError(localizer *i18n.Localizer, errorCode string, templateData map[string]interface{}) string {
	return m.Localize(localizer, "error."+errorCode, templateData)
}

// LocalizeSuccess translates a success message
func (m *Manager) LocalizeSuccess(localizer *i18n.Localizer, messageCode string, templateData map[string]interface{}) string {
	return m.Localize(localizer, "success."+messageCode, templateData)
}

// Middleware provides Gin middleware for i18n
type Middleware struct {
	manager *Manager
}

// NewMiddleware creates a new i18n middleware
func NewMiddleware(manager *Manager) *Middleware {
	return &Middleware{manager: manager}
}

// Handler returns a Gin middleware function
func (m *Middleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get language from various sources (query param, header, etc.)
		lang := m.getLanguageFromRequest(c)

		// Create localizer
		localizer := m.manager.GetLocalizer(lang)

		// Store in context
		c.Set("localizer", localizer)
		c.Set("language", lang)

		// Add to request context as well
		ctx := context.WithValue(c.Request.Context(), "localizer", localizer)
		ctx = context.WithValue(ctx, "language", lang)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// getLanguageFromRequest extracts language from request
func (m *Middleware) getLanguageFromRequest(c *gin.Context) string {
	// 1. Check query parameter
	if lang := c.Query("lang"); lang != "" && m.manager.supportedLangs[lang] {
		return lang
	}

	// 2. Check header
	if lang := c.GetHeader("Accept-Language"); lang != "" {
		return m.manager.parseAcceptLanguage(lang)
	}

	// 3. Check custom header
	if lang := c.GetHeader("X-Language"); lang != "" && m.manager.supportedLangs[lang] {
		return lang
	}

	// 4. Return default
	return m.manager.defaultLanguage
}

// Helper functions for Gin context

// GetLocalizer gets localizer from Gin context
func GetLocalizer(c *gin.Context) *i18n.Localizer {
	if localizer, exists := c.Get("localizer"); exists {
		return localizer.(*i18n.Localizer)
	}
	return nil
}

// GetLocalizerFromContext gets localizer from context
func GetLocalizerFromContext(ctx context.Context) *i18n.Localizer {
	if localizer := ctx.Value("localizer"); localizer != nil {
		return localizer.(*i18n.Localizer)
	}
	return nil
}

// GetLanguage gets current language from Gin context
func GetLanguage(c *gin.Context) string {
	if lang, exists := c.Get("language"); exists {
		return lang.(string)
	}
	return "en"
}

// GetLanguageFromContext gets current language from context
func GetLanguageFromContext(ctx context.Context) string {
	if lang := ctx.Value("language"); lang != nil {
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

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})
	if err != nil {
		return messageID
	}
	return msg
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

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:      messageID,
		DefaultMessage: &i18n.Message{ID: messageID, Other: defaultMessage},
		TemplateData:   data,
	})
	if err != nil {
		return defaultMessage
	}
	return msg
}

// TCtx is a helper function for translation from context
func TCtx(ctx context.Context, messageID string, templateData ...map[string]interface{}) string {
	localizer := GetLocalizerFromContext(ctx)
	if localizer == nil {
		return messageID
	}

	var data map[string]interface{}
	if len(templateData) > 0 {
		data = templateData[0]
	}

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})
	if err != nil {
		return messageID
	}
	return msg
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
