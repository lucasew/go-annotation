package annotation

import (
	"context"
	"embed"
	"html/template"
	"io"
	"maps"
	"net/http"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/russross/blackfriday/v2"
)

var (
	//go:embed templates/*
	templateFS embed.FS

	//go:embed assets/css/output.css
	cssContent string

	//go:embed assets/favicon.svg
	faviconContent string

	// Template manager with mold for layout support
	templateManager *TemplateManager = nil

	// TemplateFuncMap contains custom template functions available globally
	TemplateFuncMap = template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"i": i, // Internationalization function (uses default localizer, override via data for context-aware)
		"markdown": func(text string) template.HTML {
			// Convert markdown to HTML using blackfriday v2
			return template.HTML(blackfriday.Run([]byte(text)))
		},
	}
)

func init() {
	// Initialize template manager with mold
	// Mold will automatically parse all templates from the embed.FS
	var err error
	templateManager, err = NewTemplateManagerWithFuncMap(templateFS, TemplateFuncMap)
	if err != nil {
		panic(err)
	}
}

// RenderPage renders a page using mold with automatic CSS injection
func RenderPage(w io.Writer, pageName string, data map[string]any) error {
	// Inject CSS automatically
	if data == nil {
		data = make(map[string]any)
	}
	data["CSS"] = template.CSS(cssContent)

	return templateManager.Render(w, "pages/"+pageName, data)
}

// RenderPageWithContext renders a page with context-aware i18n
func RenderPageWithContext(ctx context.Context, w io.Writer, pageName string, data map[string]any) error {
	// Inject CSS automatically
	if data == nil {
		data = make(map[string]any)
	}
	data["CSS"] = template.CSS(cssContent)

	// Inject context-aware translation function
	localizer := GetLocalizerFromContext(ctx)
	data["i"] = func(messageID string) string {
		msg, err := localizer.Localize(&i18n.LocalizeConfig{
			MessageID: messageID,
		})
		if err != nil {
			return messageID
		}
		return msg
	}

	return templateManager.Render(w, "pages/"+pageName, data)
}

// RenderPageWithRequest renders a page with request-aware i18n
func RenderPageWithRequest(r *http.Request, w io.Writer, pageName string, data map[string]any) error {
	return RenderPageWithContext(r.Context(), w, pageName, data)
}

// RenderPageWithTitle is a convenience function to render a page with just a title
func RenderPageWithTitle(w io.Writer, pageName, title string, data any) error {
	dataMap := make(map[string]any)

	// Set title
	dataMap["Title"] = title

	// If data is already a map, merge it
	if m, ok := data.(map[string]any); ok {
		maps.Copy(dataMap, m)
	} else if data != nil {
		dataMap["Data"] = data
	}

	return RenderPage(w, pageName, dataMap)
}

// GetFavicon returns the embedded favicon content
func GetFavicon() string {
	return faviconContent
}
