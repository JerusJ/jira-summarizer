package jira

import (
	"embed"
	"io"
	"log"
	"strings"
	"text/template"

	"github.com/jerusj/jira-summarizer/pkg/util"
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

func toDayOfWeek(s string) string {
	return util.GetDayOfWeekOrDie(s)
}

func init() {
	var err error
	templates = template.New("").Funcs(template.FuncMap{
		"toSlice":     toSlice,
		"toDayOfWeek": toDayOfWeek,
	})
	templates, err = templates.ParseFS(templateFiles, "templates/*.tmpl")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
}

// RenderTemplate renders the template to a specified io.Writer
func RenderTemplateFromSummariesByDate(writer io.Writer, tmplName string, data map[string][]IssueSummary) error {
	return templates.ExecuteTemplate(writer, tmplName, data)
}
