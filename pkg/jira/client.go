package jira

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
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
	headers    map[string]string
	mu         sync.Mutex
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
		headers:    headers,
	}
}

func (c *JiraClient) makeRequest(ctx context.Context, reqMethod string, url string, data []byte) (respBytes []byte, err error) {
	body := bytes.NewReader(data)

	req, err := http.NewRequestWithContext(ctx, reqMethod, url, body)
	if err != nil {
		return
	}

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	respBytes, err = io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("cannot '%s': '%s'", reqMethod, err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("cannot '%s': status code not OK: '%s'", reqMethod, string(respBytes))
		return
	}

	return
}
