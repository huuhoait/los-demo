package i18n

import (
	"embed"
	"fmt"
	"path/filepath"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

//go:embed *.toml
var localeFS embed.FS

// Localizer wraps the go-i18n localizer with additional functionality
type Localizer struct {
	bundle    *i18n.Bundle
	localizer *i18n.Localizer
	language  string
}

// Config represents localizer configuration
type Config struct {
	DefaultLanguage string
	AcceptLanguages []string
}

// NewLocalizer creates a new localizer instance
func NewLocalizer(config Config) (*Localizer, error) {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", func(data []byte, v interface{}) error {
		return yaml.Unmarshal(data, v)
	})

	// Load embedded locale files
	localeFiles, err := localeFS.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read locale directory: %w", err)
	}

	for _, file := range localeFiles {
		if filepath.Ext(file.Name()) == ".toml" {
			data, err := localeFS.ReadFile(file.Name())
			if err != nil {
				return nil, fmt.Errorf("failed to read locale file %s: %w", file.Name(), err)
			}

			if _, err := bundle.ParseMessageFileBytes(data, file.Name()); err != nil {
				return nil, fmt.Errorf("failed to parse locale file %s: %w", file.Name(), err)
			}
		}
	}

	defaultLang := config.DefaultLanguage
	if defaultLang == "" {
		defaultLang = "en"
	}

	localizer := i18n.NewLocalizer(bundle, defaultLang)

	return &Localizer{
		bundle:    bundle,
		localizer: localizer,
		language:  defaultLang,
	}, nil
}

// Localize translates a message with the given ID and optional template data
func (l *Localizer) Localize(messageID string, templateData map[string]interface{}) string {
	config := &i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	}

	result, err := l.localizer.Localize(config)
	if err != nil {
		return messageID // Return the ID if translation fails
	}

	return result
}

// LocalizeWithCount translates a message with plural support
func (l *Localizer) LocalizeWithCount(messageID string, count int, templateData map[string]interface{}) string {
	if templateData == nil {
		templateData = make(map[string]interface{})
	}
	templateData["Count"] = count

	config := &i18n.LocalizeConfig{
		MessageID:    messageID,
		PluralCount:  count,
		TemplateData: templateData,
	}

	result, err := l.localizer.Localize(config)
	if err != nil {
		return messageID
	}

	return result
}

// GetLanguage returns the current language
func (l *Localizer) GetLanguage() string {
	return l.language
}

// SetLanguage changes the current language
func (l *Localizer) SetLanguage(lang string) {
	l.localizer = i18n.NewLocalizer(l.bundle, lang)
	l.language = lang
}

// MustLocalize is a compatibility method that returns the localized string
// This method is added for compatibility with existing code
func (l *Localizer) MustLocalize(config *i18n.LocalizeConfig) string {
	result, err := l.localizer.Localize(config)
	if err != nil {
		if config.MessageID != "" {
			return config.MessageID
		}
		return "translation_error"
	}
	return result
}

// GetAvailableLanguages returns a list of available languages
func (l *Localizer) GetAvailableLanguages() []string {
	// This would typically be populated from the loaded locale files
	return []string{"en", "vi"}
}
