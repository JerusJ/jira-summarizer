package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/jerusj/jira-summarizer/pkg/constants"
)

// Issue represents a Jira issue.
type Issue struct {
	Key       string      `json:"key"`
	Self      string      `json:"self"`
	Fields    IssueFields `json:"fields"`
	Changelog Changelog   `json:"changelog"`
}

// IssueFields contains fields of an issue.
type IssueFields struct {
	Summary string `json:"summary"`
}

// Changelog contains the history of changes to an issue.
type Changelog struct {
	Histories []History `json:"histories"`
}

// History represents a single change in the changelog.
type History struct {
	Created string        `json:"created"`
	Author  User          `json:"author"`
	Items   []HistoryItem `json:"items"`
}

// User represents a Jira user.
type User struct {
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

// HistoryItem represents a single change item in a history record.
type HistoryItem struct {
	Field      string `json:"field"`
	Fieldtype  string `json:"fieldtype"`
	From       string `json:"from"`
	FromString string `json:"fromString"`
	To         string `json:"to"`
	ToString   string `json:"toString"`
}

// Comment represents a Jira comment.
type Comment struct {
	Body    string `json:"body"`
	Created string `json:"created"`
	Author  User   `json:"author"`
}

// StatusTransition represents a status change in an issue.
type StatusTransition struct {
	From      string
	To        string
	Timestamp time.Time
}

// IssueSummary contains the summarized data for an issue.
type IssueSummary struct {
	Key               string
	Link              string
	Comments          []Comment
	StatusTransitions []StatusTransition

	LastComment          Comment
	LastStatusTransition StatusTransition
}

// fetchAssignedIssues retrieves all issues assigned to the current user.
func (client *JiraClient) FetchAssignedIssues(ctx context.Context) ([]Issue, error) {
	apiURL := fmt.Sprintf("%s/rest/api/2/search", client.BaseURL)
	jql := url.QueryEscape("assignee = currentUser()")
	reqURL := fmt.Sprintf("%s?jql=%s&expand=changelog", apiURL, jql)

	bodyBytes, err := client.makeRequest(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	var searchResult struct {
		Issues []Issue `json:"issues"`
	}
	err = json.Unmarshal(bodyBytes, &searchResult)
	if err != nil {
		return nil, err
	}

	return searchResult.Issues, nil
}

// Helper function to filter status transitions within a date range and by author email
func filterStatusTransitions(issue Issue, startDate, endDate time.Time, email string) ([]StatusTransition, error) {
	var statusTransitions []StatusTransition
	for _, history := range issue.Changelog.Histories {
		historyTime, err := time.Parse(constants.DateLayout, history.Created)
		if err != nil {
			return nil, err
		}
		if historyTime.Before(startDate) || historyTime.After(endDate) {
			continue
		}
		if history.Author.EmailAddress != email {
			continue
		}
		for _, item := range history.Items {
			if item.Field == "status" {
				statusTransitions = append(statusTransitions, StatusTransition{
					From:      item.FromString,
					To:        item.ToString,
					Timestamp: historyTime,
				})
			}
		}
	}
	return statusTransitions, nil
}

// Helper function to fetch and filter comments within a date range and by author email
func (client *JiraClient) fetchIssueCommentsWithinDateRange(ctx context.Context, issueKey string, startDate, endDate time.Time, email string) ([]Comment, error) {
	issueComments, err := client.FetchIssueComments(ctx, issueKey)
	if err != nil {
		return nil, err
	}
	var comments []Comment
	for _, comment := range issueComments {
		commentTime, err := time.Parse(constants.DateLayout, comment.Created)
		if err != nil {
			return nil, err
		}
		if commentTime.Before(startDate) || commentTime.After(endDate) {
			continue
		}
		if comment.Author.EmailAddress != email {
			continue
		}
		comments = append(comments, comment)
	}
	return comments, nil
}

// FetchAssignedIssueSummariesByDate fetches issue summaries with changes grouped by date
func (client *JiraClient) FetchAssignedIssueSummariesByDate(ctx context.Context, startDate, endDate time.Time) (map[string][]IssueSummary, error) {
	summariesByDate := make(map[string]map[string]*IssueSummary)

	// Fetch the assigned issues
	issues, err := client.FetchAssignedIssues(ctx)
	if err != nil {
		return nil, err
	}

	// For each issue
	for _, issue := range issues {
		// Get status transitions
		statusTransitions, err := filterStatusTransitions(issue, startDate, endDate, client.Email)
		if err != nil {
			return nil, err
		}

		// Get comments
		comments, err := client.fetchIssueCommentsWithinDateRange(ctx, issue.Key, startDate, endDate, client.Email)
		if err != nil {
			return nil, err
		}

		// Process status transitions
		for _, st := range statusTransitions {
			dateStr := st.Timestamp.Format(constants.DateLayoutInput)
			if _, ok := summariesByDate[dateStr]; !ok {
				summariesByDate[dateStr] = make(map[string]*IssueSummary)
			}
			issueSummary, ok := summariesByDate[dateStr][issue.Key]
			if !ok {
				issueSummary = &IssueSummary{
					Key:  issue.Key,
					Link: issue.Self,
				}
				summariesByDate[dateStr][issue.Key] = issueSummary
			}
			issueSummary.StatusTransitions = append(issueSummary.StatusTransitions, st)
			issueSummary.LastStatusTransition = st
		}

		// Process comments
		for _, c := range comments {
			commentTime, err := time.Parse(constants.DateLayout, c.Created)
			if err != nil {
				return nil, err
			}
			dateStr := commentTime.Format(constants.DateLayoutInput)
			if _, ok := summariesByDate[dateStr]; !ok {
				summariesByDate[dateStr] = make(map[string]*IssueSummary)
			}
			issueSummary, ok := summariesByDate[dateStr][issue.Key]
			if !ok {
				issueSummary = &IssueSummary{
					Key:  issue.Key,
					Link: issue.Self,
				}
				summariesByDate[dateStr][issue.Key] = issueSummary
			}
			issueSummary.Comments = append(issueSummary.Comments, c)
			issueSummary.LastComment = c
		}
	}

	// Convert summariesByDate to map[string][]IssueSummary
	finalSummariesByDate := make(map[string][]IssueSummary)
	for dateStr, issueMap := range summariesByDate {
		for _, summaryPtr := range issueMap {
			finalSummariesByDate[dateStr] = append(finalSummariesByDate[dateStr], *summaryPtr)
		}
	}

	return finalSummariesByDate, nil
}

// FetchIssueComments retrieves comments for a specific issue.
func (client *JiraClient) FetchIssueComments(ctx context.Context, issueKey string) ([]Comment, error) {
	reqURL := fmt.Sprintf("%s/rest/api/2/issue/%s/comment", client.BaseURL, issueKey)

	bodyBytes, err := client.makeRequest(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	var commentResult struct {
		Comments []Comment `json:"comments"`
	}
	err = json.Unmarshal(bodyBytes, &commentResult)
	if err != nil {
		return nil, err
	}

	return commentResult.Comments, nil
}
