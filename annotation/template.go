package annotation

import (
	"bytes"
	"embed"
	"fmt"
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

// RenderPage renders a page using a base layout
func RenderPage(w io.Writer, pageName string, data map[string]interface{}) error {
	// Ensure data map exists
	if data == nil {
		data = make(map[string]interface{})
	}

	// Inject CSS automatically
	data["CSS"] = template.CSS(cssContent)

	// Create a new map for the layout data to avoid polluting the page data
	layoutData := make(map[string]interface{})
	for k, v := range data {
		layoutData[k] = v
	}

	// First, render the page content itself to a buffer
	var pageContent bytes.Buffer
	// We render the page without a layout first.
	// The "extends" directive is not working, so we manually implement the layout.
	// The mold engine seems to ignore rendering a template that only has blocks,
	// so we render the page directly.
	err := templateManager.Render(&pageContent, "pages/"+pageName, data)
	if err != nil {
		return fmt.Errorf("failed to render page content '%s': %w", pageName, err)
	}

	// Now, add the rendered page content to the layout data under the "content" block name
	layoutData["content"] = template.HTML(pageContent.String())

	// Finally, render the base layout, which will use the "content" data
	return templateManager.Render(w, "layouts/base.html", layoutData)
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
