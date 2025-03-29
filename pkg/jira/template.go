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

	regexpJiraSmartlink = regexp.MustCompile(`\[(.*?\|.*?)\]`)
	regexpJiraMention   = regexp.MustCompile(`\[~accountid:([^\]]+)\]`)
	regexpJiraImage     = regexp.MustCompile(`!image-[^!]+!`)
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

func toMarkdownLinks(s string) string {
	sNew := s
	smartlinks := regexpJiraSmartlink.FindAllString(s, -1)
	for _, smartlink := range smartlinks {
		smartLinkParts := strings.Split(smartlink, "|")
		alias := smartLinkParts[0]
		alias = strings.TrimLeft(alias, "[")
		url := smartLinkParts[1]
		url = strings.TrimRight(url, "]")
		markdownLink := fmt.Sprintf("[%s](%s)", alias, url)
		sNew = strings.ReplaceAll(sNew, smartlink, markdownLink)
	}
	return sNew
}

// TODO(jesse): lookup user ID beforehand to transform comments so we don't need to do this?
func cleanJiraUserMentions(s string) string {
	sNew := s
	userMentions := regexpJiraMention.FindAllString(s, -1)
	for _, userMention := range userMentions {
		sNew = strings.ReplaceAll(sNew, userMention, "<JIRA_USER>")
	}
	return sNew
}

func cleanJiraImages(s string) string {
	sNew := regexpJiraImage.ReplaceAllString(s, "(see image in Jira issue)")
	return sNew
}

func toDayOfWeek(s string) string {
	return util.GetDayOfWeekOrDie(s)
}

func init() {
	var err error
	templates = template.New("").Funcs(template.FuncMap{
		"toSlice":               toSlice,
		"toMarkdownLinks":       toMarkdownLinks,
		"cleanJiraUserMentions": cleanJiraUserMentions,
		"cleanJiraImages":       cleanJiraImages,
		"toDayOfWeek":           toDayOfWeek,
	})
	templates, err = templates.ParseFS(templateFiles, "templates/*.tmpl")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
}

// RenderTemplate renders the template to a specified io.Writer
func RenderTemplateFromSummariesByDate(writer io.Writer, tmplName string, data map[string]map[string][]IssueSummary) error {
	return templates.ExecuteTemplate(writer, tmplName, data)
}
