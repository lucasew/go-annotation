package annotation

import (
    "embed"
    "html/template"
)

var (
    //go:embed tmpl/*
    templateFiles embed.FS
    Template *template.Template = nil
)

func init() {
    Template = template.Must(template.ParseFS(templateFiles, "tmpl/*.html"))
}
