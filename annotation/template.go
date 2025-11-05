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

	//go:embed templates/*
	templateFiles embed.FS

	//go:embed assets/css/output.css
	cssContent string

	Template *template.Template = nil
)

func init() {
	Template = template.Must(template.ParseFS(templateFiles, "templates/*.html"))
}

type TemplateContent struct {
	Title   string
	Content string
}

type templateRuntimeContent struct {
	Title   string
	Content template.HTML
	CSS     template.CSS
}

func ExecTemplate(w io.Writer, content TemplateContent) error {
	htmlized := blackfriday.MarkdownCommon([]byte(content.Content))
	templateContent := templateRuntimeContent{
		Title:   content.Title,
		Content: template.HTML(string(htmlized)),
		CSS:     template.CSS(cssContent),
	}
	return Template.ExecuteTemplate(w, "base.html", templateContent)
}

// ExecNamedTemplate executes a named template with custom data
func ExecNamedTemplate(w io.Writer, templateName string, data interface{}) error {
	// Create a wrapper that includes CSS
	type wrapper struct {
		Data interface{}
		CSS  template.CSS
	}

	wrapped := wrapper{
		Data: data,
		CSS:  template.CSS(cssContent),
	}

	return Template.ExecuteTemplate(w, templateName, wrapped.Data)
}

// RenderWithLayout renders content within the base layout
func RenderWithLayout(w io.Writer, title string, data interface{}) error {
	type layoutData struct {
		Title   string
		Content interface{}
		CSS     template.CSS
	}

	ld := layoutData{
		Title:   title,
		Content: data,
		CSS:     template.CSS(cssContent),
	}

	return Template.ExecuteTemplate(w, "base.html", ld)
}
