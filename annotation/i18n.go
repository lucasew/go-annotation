package annotation

import (
	"context"
	"embed"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localesFS embed.FS

var (
	bundle        *i18n.Bundle
	defaultLocal  *i18n.Localizer
	currentLocale string = "en"
)

type localizerKey struct{}

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load all locale files
	locales := []string{"en", "pt-BR"}
	for _, locale := range locales {
		data, err := localesFS.ReadFile("locales/" + locale + ".json")
		if err != nil {
			log.Printf("Warning: failed to read locale file %s: %v", locale, err)
			continue
		}

		_, err = bundle.ParseMessageFileBytes(data, locale+".json")
		if err != nil {
			log.Printf("Warning: failed to parse locale file %s: %v", locale, err)
		}
	}

	defaultLocal = i18n.NewLocalizer(bundle, currentLocale)
}

// SetLanguage sets the current language for translations
func SetLanguage(lang string) {
	currentLocale = lang
	defaultLocal = i18n.NewLocalizer(bundle, currentLocale)
}

// GetLocalizerFromContext retrieves the localizer from context, or returns default
func GetLocalizerFromContext(ctx context.Context) *i18n.Localizer {
	if ctx == nil {
		return defaultLocal
	}

	if localizer, ok := ctx.Value(localizerKey{}).(*i18n.Localizer); ok {
		return localizer
	}
	return defaultLocal
}

// WithLocalizer adds a localizer to the context
func WithLocalizer(ctx context.Context, localizer *i18n.Localizer) context.Context {
	return context.WithValue(ctx, localizerKey{}, localizer)
}

// GetLocalizerFromRequest creates a localizer based on the Accept-Language header
func GetLocalizerFromRequest(r *http.Request) *i18n.Localizer {
	acceptLang := r.Header.Get("Accept-Language")

	// Parse Accept-Language header to get preferred languages
	// Format: "en-US,en;q=0.9,pt-BR;q=0.8,pt;q=0.7"
	var langs []string
	if acceptLang != "" {
		parts := strings.Split(acceptLang, ",")
		for _, part := range parts {
			// Remove quality values (;q=0.9)
			lang := strings.TrimSpace(strings.Split(part, ";")[0])
			if lang != "" {
				langs = append(langs, lang)
			}
		}
	}

	// Add default language as fallback
	if len(langs) == 0 {
		langs = []string{currentLocale}
	}

	return i18n.NewLocalizer(bundle, langs...)
}

// i translates a message ID to the current language
// Note: This uses the default localizer. For per-request localization,
// use templates with context or call T() directly
func i(messageID string) string {
	msg, err := defaultLocal.Localize(&i18n.LocalizeConfig{
		MessageID: messageID,
	})
	if err != nil {
		// Return the message ID if translation not found
		return messageID
	}
	return msg
}

// T is an alias for i (for backward compatibility and convenience)
func T(messageID string) string {
	return i(messageID)
}

// LocalizeWithData translates a message with template data
func LocalizeWithData(messageID string, data map[string]interface{}) string {
	msg, err := defaultLocal.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})
	if err != nil {
		return messageID
	}
	return msg
}

// LocalizeWithContext translates a message using the localizer from context
func LocalizeWithContext(ctx context.Context, messageID string) string {
	localizer := GetLocalizerFromContext(ctx)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID: messageID,
	})
	if err != nil {
		return messageID
	}
	return msg
}

// LocalizeWithContextAndData translates a message with template data using context
func LocalizeWithContextAndData(ctx context.Context, messageID string, data map[string]interface{}) string {
	localizer := GetLocalizerFromContext(ctx)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: data,
	})
	if err != nil {
		return messageID
	}
	return msg
}
