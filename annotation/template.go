package annotation

import (
	"embed"
	"html/template"
	"io"
	"maps"
)

var (
	//go:embed templates/*
	templateFS embed.FS

	//go:embed assets/css/output.css
	cssContent string

	// Template manager with mold for layout support
	templateManager *TemplateManager = nil

	// TemplateFuncMap contains custom template functions available globally
	TemplateFuncMap = template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
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
