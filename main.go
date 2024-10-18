package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/jerusj/jira-summarizer/pkg/constants"
	"github.com/jerusj/jira-summarizer/pkg/jira"
)

var (
	emailFlag      = flag.String("email", "", "Email username for Jira Server")
	urlFlag        = flag.String("url", "", "URL to the Jira server")
	startFlag      = flag.String("start", "", "Start date range (MM/DD/YYYY)")
	endFlag        = flag.String("end", "", "End date range (MM/DD/YYYY)")
	outputTypeFlag = flag.String("output-type", "text", "Type of output to present to STDOUT")
)

func main() {
	flag.Parse()

	if *emailFlag == "" {
		panic("-email cannot be empty")
	}
	if *urlFlag == "" {
		panic("-url cannot be empty")
	}
	if *startFlag == "" {
		panic("-start cannot be empty")
	}
	if *endFlag == "" {
		panic("-start cannot be empty")
	}
	if *outputTypeFlag == "" {
		panic("-output-type cannot be empty")
	}

	startDate, err := time.ParseInLocation(constants.DateLayoutInput, *startFlag, constants.DateLayoutLocation)
	if err != nil {
		log.Fatalf("Invalid start date: %v", err)
	}
	endDate, err := time.ParseInLocation(constants.DateLayoutInput, *endFlag, constants.DateLayoutLocation)
	if err != nil {
		log.Fatalf("Invalid end date: %v", err)
	}

	// Set times to cover the entire days
	startDate = startDate.Truncate(24 * time.Hour)
	endDate = endDate.AddDate(0, 0, 1).Truncate(24 * time.Hour).Add(-time.Nanosecond)

	client := jira.NewJiraClient(*urlFlag, *emailFlag)

	issues, err := client.FetchAssignedIssues()
	if err != nil {
		log.Fatalf("Error fetching issues: %v", err)
	}

	var summaries []jira.IssueSummary
	for _, issue := range issues {
		summary, err := client.FilterChanges(issue, startDate, endDate)
		if err != nil {
			log.Printf("Error processing issue %s: %v", issue.Key, err)
			continue
		}
		if len(summary.Comments) > 0 || len(summary.StatusTransitions) > 0 {
			summaries = append(summaries, summary)
		}
	}

	err = jira.RenderTemplateFromSummaries(os.Stdout, "slack.tmpl", summaries)
	if err != nil {
		panic(err)
	}
}
