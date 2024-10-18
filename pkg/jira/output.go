package jira

import (
	"embed"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"
	"text/template"

	"github.com/jerusj/jira-summarizer/pkg/util"
)

var (
	//go:embed templates/*.tmpl
	templateFiles embed.FS
	templates     *template.Template

	regexpJiraSmartlink = regexp.MustCompile(`(\[.*\|.*\])`)
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

func cleanJiraLinks(s string) string {
	sNew := s
	smartlinks := regexpJiraSmartlink.FindAllString(s, -1)
	for _, smartlink := range smartlinks {
		actualLink := strings.Split(smartlink, "|")[0]
		actualLink = strings.ReplaceAll(actualLink, "[", "")
		actualLink = fmt.Sprintf("[%s](%s)", actualLink, actualLink)
		sNew = strings.ReplaceAll(sNew, smartlink, actualLink)
	}
	return sNew
}

func toDayOfWeek(s string) string {
	return util.GetDayOfWeekOrDie(s)
}

func init() {
	var err error
	templates = template.New("").Funcs(template.FuncMap{
		"toSlice":        toSlice,
		"cleanJiraLinks": cleanJiraLinks,
		"toDayOfWeek":    toDayOfWeek,
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
