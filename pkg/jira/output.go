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
	var sLines []string
	sSplit := strings.Split(s, "\n")
	for _, line := range sSplit {
		line = strings.TrimSpace(line)
		if line != "" {
			sLines = append(sLines, line)
		}
	}
	return sLines
}

func init() {
	var err error
	templates = template.New("").Funcs(template.FuncMap{
		"toSlice": toSlice,
	})
	templates, err = templates.ParseFS(templateFiles, "templates/*.tmpl")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
}

// RenderTemplate renders the template to a specified io.Writer
func RenderTemplateFromSummaries(writer io.Writer, tmplName string, data []IssueSummary) error {
	return templates.ExecuteTemplate(writer, tmplName, data)
}
