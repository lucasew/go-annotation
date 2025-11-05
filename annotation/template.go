package annotation

import (
	"embed"
	"html/template"
	"io"

	"github.com/russross/blackfriday"
)

var (
	//go:embed tmpl/*
	oldTemplateFiles embed.FS

	//go:embed templates/**/*
	templateFS embed.FS

	//go:embed assets/css/output.css
	cssContent string

	// Legacy template system
	Template *template.Template = nil

	// New template manager with mold for layout support
	TemplateManager *TemplateManager = nil
)

func init() {
	// Initialize legacy template system for backward compatibility
	Template = template.Must(template.ParseFS(templateFS, "templates/*.html"))

	// Initialize new template manager with mold
	TemplateManager = NewTemplateManager()
	if err := TemplateManager.LoadFromFS(templateFS, "templates/layouts/*.html", "templates/pages/*.html"); err != nil {
		panic(err)
	}
}

// TemplateContent represents legacy template content
type TemplateContent struct {
	Title   string
	Content string
}

type templateRuntimeContent struct {
	Title   string
	Content template.HTML
	CSS     template.CSS
}

// ExecTemplate executes a template using the legacy system (for backward compatibility)
func ExecTemplate(w io.Writer, content TemplateContent) error {
	htmlized := blackfriday.MarkdownCommon([]byte(content.Content))
	templateContent := templateRuntimeContent{
		Title:   content.Title,
		Content: template.HTML(string(htmlized)),
		CSS:     template.CSS(cssContent),
	}
	return Template.ExecuteTemplate(w, "base.html", templateContent)
}

// RenderPage renders a page using mold with automatic CSS injection
func RenderPage(w io.Writer, pageName string, data map[string]interface{}) error {
	// Inject CSS automatically
	if data == nil {
		data = make(map[string]interface{})
	}
	data["CSS"] = template.CSS(cssContent)

	return TemplateManager.Render(w, "pages/"+pageName, data)
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
