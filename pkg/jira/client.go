package jira

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jerusj/jira-summarizer/pkg/constants"
	"github.com/jerusj/jira-summarizer/pkg/util"
)

const (
	ENV_JIRA_API_TOKEN = "JIRA_API_TOKEN"
)

var (
	DefaultTimeout = time.Second * 60
)

// JiraClient represents a client for Jira API.
type JiraClient struct {
	BaseURL    string
	AuthHeader string
	HTTPClient *http.Client
	Email      string
	headers    map[string]string
}

// NewJiraClient initializes and returns a new JiraClient.
func NewJiraClient(baseURL, email string) *JiraClient {
	baseURL = strings.TrimSpace(baseURL)
	email = strings.TrimSpace(email)
	if baseURL == "" {
		panic("Jira BASE URL cannot be empty")
	}
	if email == "" {
		panic("Jira user email cannot be empty")
	}

	apiToken := util.GetEnvOrDie(ENV_JIRA_API_TOKEN)
	auth := email + ":" + apiToken
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))

	client := http.DefaultClient
	client.Timeout = DefaultTimeout

	headers := map[string]string{
		"Authorization": authHeader,
		"Content-Type":  "application/json",
	}

	return &JiraClient{
		BaseURL:    baseURL,
		HTTPClient: client,
		Email:      email,
		headers:    headers,
	}
}

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
}

// fetchAssignedIssues retrieves all issues assigned to the current user.
func (client *JiraClient) FetchAssignedIssues() ([]Issue, error) {
	apiURL := fmt.Sprintf("%s/rest/api/2/search", client.BaseURL)
	jql := url.QueryEscape("assignee = currentUser()")
	reqURL := fmt.Sprintf("%s?jql=%s&expand=changelog", apiURL, jql)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range client.headers {
		req.Header.Add(k, v)
	}

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
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

func (client *JiraClient) FilterChanges(issue Issue, startDate, endDate time.Time) (IssueSummary, error) {
	var comments []Comment
	var statusTransitions []StatusTransition

	// Process changelog histories
	for _, history := range issue.Changelog.Histories {
		historyTime, err := time.Parse(constants.DateLayout, history.Created)
		if err != nil {
			return IssueSummary{}, err
		}

		if historyTime.Before(startDate) || historyTime.After(endDate) {
			continue
		}

		if history.Author.EmailAddress != client.Email {
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

	// Fetch and process comments
	issueComments, err := client.FetchIssueComments(issue.Key)
	if err != nil {
		return IssueSummary{}, err
	}

	for _, comment := range issueComments {
		commentTime, err := time.Parse(constants.DateLayout, comment.Created)
		if err != nil {
			return IssueSummary{}, err
		}
		if commentTime.Before(startDate) || commentTime.After(endDate) {
			continue
		}
		if comment.Author.EmailAddress != client.Email {
			continue
		}
		comments = append(comments, comment)
	}

	summary := IssueSummary{
		Key:               issue.Key,
		Link:              issue.Self,
		Comments:          comments,
		StatusTransitions: statusTransitions,
	}

	return summary, nil
}

// fetchIssueComments retrieves comments for a specific issue.
func (client *JiraClient) FetchIssueComments(issueKey string) ([]Comment, error) {
	reqURL := fmt.Sprintf("%s/rest/api/2/issue/%s/comment", client.BaseURL, issueKey)

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range client.headers {
		req.Header.Add(k, v)
	}

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
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
