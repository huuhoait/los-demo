package i18n

import (
	"context"
	"embed"
	"fmt"
	"path/filepath"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v3"
)

//go:embed ../../i18n/*.toml
var localeFS embed.FS

// Localizer provides internationalization support
type Localizer struct {
	bundle *i18n.Bundle
}

// Config holds i18n configuration
type Config struct {
	DefaultLanguage string   `yaml:"default_language" json:"default_language"`
	SupportedLangs  []string `yaml:"supported_languages" json:"supported_languages"`
	FallbackLang    string   `yaml:"fallback_language" json:"fallback_language"`
}

// NewLocalizer creates a new localizer instance
func NewLocalizer(cfg *Config) (*Localizer, error) {
	if cfg == nil {
		cfg = &Config{
			DefaultLanguage: "en",
			SupportedLangs:  []string{"en", "vi"},
			FallbackLang:    "en",
		}
	}

	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", yaml.Unmarshal)

	// Load all locale files from embedded filesystem
	for _, lang := range cfg.SupportedLangs {
		filename := fmt.Sprintf("../../i18n/%s.toml", lang)
		data, err := localeFS.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read locale file %s: %w", filename, err)
		}

		_, err = bundle.ParseMessageFileBytes(data, filepath.Base(filename))
		if err != nil {
			return nil, fmt.Errorf("failed to parse locale file %s: %w", filename, err)
		}
	}

	return &Localizer{
		bundle: bundle,
	}, nil
}

// GetLocalizer returns a go-i18n localizer for the given language
func (l *Localizer) GetLocalizer(lang string) *i18n.Localizer {
	return i18n.NewLocalizer(l.bundle, lang, "en") // English as fallback
}

// GetLocalizerFromContext extracts language from context and returns localizer
func (l *Localizer) GetLocalizerFromContext(ctx context.Context) *i18n.Localizer {
	lang := GetLanguageFromContext(ctx)
	return l.GetLocalizer(lang)
}

// Localize translates a message with the given parameters
func (l *Localizer) Localize(lang, messageID string, templateData map[string]interface{}) string {
	localizer := l.GetLocalizer(lang)

	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil {
		// Return the message ID if translation fails
		return messageID
	}

	return msg
}

// LocalizeError translates an error message
func (l *Localizer) LocalizeError(lang, errorCode string, templateData map[string]interface{}) string {
	messageID := fmt.Sprintf("errors.%s", errorCode)
	return l.Localize(lang, messageID, templateData)
}

// LocalizeMessage translates a regular message
func (l *Localizer) LocalizeMessage(lang, messageKey string, templateData map[string]interface{}) string {
	messageID := fmt.Sprintf("messages.%s", messageKey)
	return l.Localize(lang, messageID, templateData)
}

// LocalizeValidation translates a validation message
func (l *Localizer) LocalizeValidation(lang, validationKey string, templateData map[string]interface{}) string {
	messageID := fmt.Sprintf("validation.%s", validationKey)
	return l.Localize(lang, messageID, templateData)
}

// LocalizeUI translates a UI element
func (l *Localizer) LocalizeUI(lang, uiKey string, templateData map[string]interface{}) string {
	messageID := fmt.Sprintf("ui.%s", uiKey)
	return l.Localize(lang, messageID, templateData)
}

// LocalizeNotification translates a notification message
func (l *Localizer) LocalizeNotification(lang, notificationKey string, templateData map[string]interface{}) string {
	messageID := fmt.Sprintf("notifications.%s", notificationKey)
	return l.Localize(lang, messageID, templateData)
}

// LocalizeSecurity translates a security message
func (l *Localizer) LocalizeSecurity(lang, securityKey string, templateData map[string]interface{}) string {
	messageID := fmt.Sprintf("security.%s", securityKey)
	return l.Localize(lang, messageID, templateData)
}

// LocalizePermission translates a permission message
func (l *Localizer) LocalizePermission(lang, permissionKey string, templateData map[string]interface{}) string {
	messageID := fmt.Sprintf("permissions.%s", permissionKey)
	return l.Localize(lang, messageID, templateData)
}

// Context keys for language
type contextKey string

const (
	LanguageContextKey  contextKey = "language"
	LocalizerContextKey contextKey = "localizer"
)

// SetLanguageInContext sets the language in context
func SetLanguageInContext(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, LanguageContextKey, lang)
}

// GetLanguageFromContext extracts language from context
func GetLanguageFromContext(ctx context.Context) string {
	if lang, ok := ctx.Value(LanguageContextKey).(string); ok {
		return lang
	}
	return "en" // Default to English
}

// SetLocalizerInContext sets the localizer in context
func SetLocalizerInContext(ctx context.Context, localizer *i18n.Localizer) context.Context {
	return context.WithValue(ctx, LocalizerContextKey, localizer)
}

// GetLocalizerFromContextDirect extracts localizer directly from context
func GetLocalizerFromContextDirect(ctx context.Context) *i18n.Localizer {
	if localizer, ok := ctx.Value(LocalizerContextKey).(*i18n.Localizer); ok {
		return localizer
	}
	return nil
}

// SupportedLanguages returns list of supported languages
func (l *Localizer) SupportedLanguages() []string {
	return []string{"en", "vi"}
}

// IsLanguageSupported checks if a language is supported
func (l *Localizer) IsLanguageSupported(lang string) bool {
	supported := l.SupportedLanguages()
	for _, supportedLang := range supported {
		if supportedLang == lang {
			return true
		}
	}
	return false
}

// DetectLanguageFromHeader detects language from Accept-Language header
func DetectLanguageFromHeader(acceptLang string) string {
	// Simple language detection - you might want to use a more sophisticated library
	if acceptLang == "" {
		return "en"
	}

	// Check for Vietnamese
	if contains(acceptLang, "vi") {
		return "vi"
	}

	// Default to English
	return "en"
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		(len(s) > len(substr) && s[len(s)-len(substr):] == substr) ||
		(len(s) > 2*len(substr) && s[len(substr):len(s)-len(substr)] == substr)
}

// LocalizedError represents an error with localized message
type LocalizedError struct {
	Code         string                 `json:"code"`
	Message      string                 `json:"message"`
	TemplateData map[string]interface{} `json:"template_data,omitempty"`
	Language     string                 `json:"language"`
}

// Error implements the error interface
func (e LocalizedError) Error() string {
	return e.Message
}

// NewLocalizedError creates a new localized error
func NewLocalizedError(code, lang string, templateData map[string]interface{}, localizer *Localizer) *LocalizedError {
	message := localizer.LocalizeError(lang, code, templateData)
	return &LocalizedError{
		Code:         code,
		Message:      message,
		TemplateData: templateData,
		Language:     lang,
	}
}

// LocalizationMiddleware provides common localization data
type LocalizationMiddleware struct {
	localizer *Localizer
}

// NewLocalizationMiddleware creates a new localization middleware
func NewLocalizationMiddleware(localizer *Localizer) *LocalizationMiddleware {
	return &LocalizationMiddleware{
		localizer: localizer,
	}
}

// AddLocalizationToContext adds localization data to context
func (m *LocalizationMiddleware) AddLocalizationToContext(ctx context.Context, acceptLang string) context.Context {
	lang := DetectLanguageFromHeader(acceptLang)
	if !m.localizer.IsLanguageSupported(lang) {
		lang = "en"
	}

	ctx = SetLanguageInContext(ctx, lang)
	localizer := m.localizer.GetLocalizer(lang)
	ctx = SetLocalizerInContext(ctx, localizer)

	return ctx
}

// FormatMessage is a helper function to format localized messages with template data
func FormatMessage(message string, data map[string]interface{}) string {
	if data == nil {
		return message
	}

	result := message
	for key, value := range data {
		placeholder := fmt.Sprintf("{%s}", key)
		replacement := fmt.Sprintf("%v", value)
		result = replaceAll(result, placeholder, replacement)
	}

	return result
}

// Simple string replacement function
func replaceAll(s, old, new string) string {
	// Simple implementation - you might want to use strings.ReplaceAll in Go 1.12+
	result := ""
	for i := 0; i < len(s); {
		if i+len(old) <= len(s) && s[i:i+len(old)] == old {
			result += new
			i += len(old)
		} else {
			result += string(s[i])
			i++
		}
	}
	return result
}
