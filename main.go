package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jerusj/jira-summarizer/pkg/constants"
	"github.com/jerusj/jira-summarizer/pkg/jira"
)

const (
	EnvJiraUsers = "JIRA_USERS"
)

var (
	getIdFlag = flag.String("get-id", "", "If email defined, get the user's ID")

	usersFlag    = flag.String("users", "", "Usernames (separated by ',') to search for")
	emailFlag    = flag.String("email", "", "Email username for Jira Server")
	urlFlag      = flag.String("url", "", "URL to the Jira server")
	startFlag    = flag.String("start", "", "Start date range (MM/DD/YYYY)")
	endFlag      = flag.String("end", "", "End date range (MM/DD/YYYY)")
	templateFlag = flag.String("template", "slack", "Name of Go Text Template (.tmpl) in this repo to use for output.")
)

func main() {
	flag.Parse()

	ctx := context.Background()
	client := jira.NewJiraClient(*urlFlag, *emailFlag)

	if *getIdFlag != "" {
		user, err := client.FetchUser(ctx, *getIdFlag)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("User: '%s' has Jira account ID: '%s'\n", *getIdFlag, user.AccountID)
		return
	}

	if *startFlag == "" {
		log.Fatal("-start cannot be empty")
	}
	if *endFlag == "" {
		log.Fatal("-start cannot be empty")
	}
	if *templateFlag == "" {
		log.Fatal("-output-type cannot be empty")
	}

	startDate, err := time.ParseInLocation(constants.DateLayoutInput, *startFlag, constants.DateLayoutLocation)
	if err != nil {
		log.Fatalf("Invalid start date: %v", err)
	}
	endDate, err := time.ParseInLocation(constants.DateLayoutInput, *endFlag, constants.DateLayoutLocation)
	if err != nil {
		log.Fatalf("Invalid end date: %v", err)
	}

	var usersInput string
	if *usersFlag == "" {
		usersInput = os.Getenv(EnvJiraUsers)
	} else {
		usersInput = *usersFlag
	}
	users := getUsersFromStr(usersInput)
	if len(users) == 0 {
		log.Fatal("no users were passed in input")
	}

	// Set times to cover the entire days
	startDate = startDate.Truncate(24 * time.Hour)
	endDate = endDate.AddDate(0, 0, 1).Truncate(24 * time.Hour).Add(-time.Nanosecond)

	allSummariesByDate, err := client.FetchAssignedIssueSummariesByDate(ctx, startDate, endDate, users)
	if err != nil {
		log.Fatalf("error fetching issues by date: %v", err)
	}

	err = jira.RenderTemplateFromSummariesByDate(os.Stdout, *templateFlag+".tmpl", allSummariesByDate)
	if err != nil {
		panic(err)
	}
}

func getUsersFromStr(s string) []string {
	var users []string
	s = strings.TrimSpace(s)
	if s == "" {
		return users
	}

	if strings.Contains(s, ",") {
		userParts := strings.Split(s, ",")
		for _, userPart := range userParts {
			userPart = strings.TrimSpace(userPart)
			if userPart != "" {
				users = append(users, userPart)
			}
		}
	} else {
		users = []string{s}
	}
	return users
}
