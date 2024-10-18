package jira

import (
	"embed"
	"io"
	"log"
	"strings"
	"text/template"
)

var (
	//go:embed templates/*.tmpl
	templateFiles embed.FS
	templates     *template.Template
)

func toSlice(s string) []string {
	return strings.Split(s, "\n")
}

func init() {
	var err error
	templates = template.New("").Funcs(template.FuncMap{
		"toSlice": toSlice,
	})
	templates, err = template.ParseFS(templateFiles, "templates/*.tmpl")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
}

// RenderTemplate renders the template to a specified io.Writer
func RenderTemplateFromSummaries(writer io.Writer, tmplName string, data []IssueSummary) error {
	return templates.ExecuteTemplate(writer, tmplName, data)
}
