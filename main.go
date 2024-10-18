package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/jerusj/jira-summarizer/pkg/constants"
	"github.com/jerusj/jira-summarizer/pkg/jira"
)

var (
	emailFlag    = flag.String("email", "", "Email username for Jira Server")
	urlFlag      = flag.String("url", "", "URL to the Jira server")
	startFlag    = flag.String("start", "", "Start date range (MM/DD/YYYY)")
	endFlag      = flag.String("end", "", "End date range (MM/DD/YYYY)")
	templateFlag = flag.String("template", "slack", "Name of Go Text Template (.tmpl) in this repo to use for output.")
)

func main() {
	flag.Parse()

	if *startFlag == "" {
		panic("-start cannot be empty")
	}
	if *endFlag == "" {
		panic("-start cannot be empty")
	}
	if *templateFlag == "" {
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

	ctx := context.Background()

	allSummariesByDate, err := client.FetchAssignedIssueSummariesByDate(ctx, startDate, endDate)
	if err != nil {
		log.Fatalf("error fetching issues by date: %v", err)
	}

	err = jira.RenderTemplateFromSummariesByDate(os.Stdout, *templateFlag+".tmpl", allSummariesByDate)
	if err != nil {
		panic(err)
	}
}
