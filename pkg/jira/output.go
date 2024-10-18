package jira

import (
	"embed"
	"io"
	"log"
	"text/template"
)

type TemplateType string

const (
	_ = iota
	TemplateSlack
)

var (
	//go:embed templates/*.tmpl
	templateFiles embed.FS
	templates     *template.Template
)

func init() {
	var err error
	// Parse all .tmpl files in the "templates" directory
	templates, err = template.ParseFS(templateFiles, "templates/*.tmpl")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
}

// RenderTemplate renders the template to a specified io.Writer
func RenderTemplateFromSummaries(writer io.Writer, tmplName string, data []IssueSummary) error {
	return templates.ExecuteTemplate(writer, tmplName, data)
}
