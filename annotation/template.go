package annotation

import (
	"embed"
	"html/template"
	"io"

	"github.com/russross/blackfriday"
)

var (
	//go:embed tmpl/*
	templateFiles embed.FS
	Template      *template.Template = nil
)

func init() {
	Template = template.Must(template.ParseFS(templateFiles, "tmpl/*.html"))
}

type TemplateContent struct {
	Title   string
	Content string
}

type templateRuntimeContent struct {
	Title   string
	Content template.HTML
}

func ExecTemplate(w io.Writer, content TemplateContent) error {
	htmlized := blackfriday.MarkdownCommon([]byte(content.Content))
	templateContent := templateRuntimeContent{
		Title:   content.Title,
		Content: template.HTML(string(htmlized)),
	}
	return Template.Execute(w, templateContent)

}
