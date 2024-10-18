package jira

import (
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"github.com/jerusj/jira-summarizer/pkg/util"
)

const (
	ENV_JIRA_API_TOKEN = "JIRA_API_TOKEN"
	ENV_JIRA_URL       = "JIRA_URL"
	ENV_JIRA_EMAIL     = "JIRA_EMAIL"
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
	if baseURL == "" {
		baseURL = util.GetEnvOrDie(ENV_JIRA_URL)
		if baseURL == "" {
			panic("Jira BASE URL cannot be empty")
		}
	}
	email = strings.TrimSpace(email)
	if email == "" {
		email = util.GetEnvOrDie(ENV_JIRA_EMAIL)
		if email == "" {
			panic("Jira user email cannot be empty")
		}
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
