package annotation

import (
	"embed"
	"html/template"
	"io"
)

var (
	//go:embed templates/**/*
	templateFS embed.FS

	//go:embed assets/css/output.css
	cssContent string

	// Template manager with mold for layout support
	templateManager *TemplateManager = nil
)

func init() {
	// Initialize template manager with mold
	// Mold will automatically parse all templates from the embed.FS
	var err error
	templateManager, err = NewTemplateManager(templateFS)
	if err != nil {
		panic(err)
	}
}

// RenderPage renders a page using mold with automatic CSS injection
func RenderPage(w io.Writer, pageName string, data map[string]interface{}) error {
	// Inject CSS automatically
	if data == nil {
		data = make(map[string]any)
	}
	data["CSS"] = template.CSS(cssContent)

	return templateManager.Render(w, "pages/"+pageName, data)
}

// RenderPageWithTitle is a convenience function to render a page with just a title
func RenderPageWithTitle(w io.Writer, pageName, title string, data interface{}) error {
	dataMap := make(map[string]interface{})

	// Set title
	dataMap["Title"] = title

	// If data is already a map, merge it
	if m, ok := data.(map[string]interface{}); ok {
		for k, v := range m {
			dataMap[k] = v
		}
	} else if data != nil {
		dataMap["Data"] = data
	}

	return RenderPage(w, pageName, dataMap)
}
