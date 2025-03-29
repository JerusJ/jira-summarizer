package jira

import (
	"context"
	"encoding/json"
	"fmt"
)

// FetchIssueComments retrieves comments for a specific issue.
func (client *JiraClient) FetchUser(ctx context.Context, userEmail string) (User, error) {
	reqURL := fmt.Sprintf("%s/rest/api/3/user/search?query=%s", client.BaseURL, userEmail)
	var result User
	var resultResp []User

	bodyBytes, err := client.makeRequest(ctx, "GET", reqURL, nil)
	if err != nil {
		return result, err
	}

	err = json.Unmarshal(bodyBytes, &resultResp)
	if err != nil {
		return result, err
	}

	if len(resultResp) == 1 {
		result = resultResp[0]
	} else if len(resultResp) > 1 {
		panic(fmt.Sprintf("got more than 1 result for user ID query, got: %d; how is this possible?", len(resultResp)))
	}

	return result, nil
}
