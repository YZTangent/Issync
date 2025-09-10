package githubapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const getIssuesQuery = `
query($searchQuery: String!, $cursor: String) {
  search(query: $searchQuery, type: ISSUE, first: 100, after: $cursor) {
    nodes {
      ... on Issue {
        number
        title
        body
        updatedAt
      }
    }
    pageInfo {
      endCursor
      hasNextPage
    }
  }
}
`

// Structs for parsing the GraphQL response
type PageInfo struct {
	EndCursor   string `json:"endCursor"`
	HasNextPage bool   `json:"hasNextPage"`
}

type IssueNode struct {
	Number    int       `json:"number"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type SearchResult struct {
	Nodes    []IssueNode `json:"nodes"`
	PageInfo PageInfo    `json:"pageInfo"`
}

type GraphQLResponse struct {
	Data struct {
		Search SearchResult `json:"search"`
	} `json:"data"`
}

type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// Client for GitHub API
type Client struct {
	httpClient *http.Client
	token      string
}

func NewClient(httpClient *http.Client, token string) *Client {
	return &Client{
		httpClient: httpClient,
		token:      token,
	}
}

func (c *Client) GetIssues(ctx context.Context, owner string, projectNumber int, since time.Time) ([]IssueNode, error) {
	var allIssues []IssueNode
	var cursor *string

	for {
		searchQuery := fmt.Sprintf("is:issue is:open project:%s/%d updated:>=%s", owner, projectNumber, since.Format("2006-01-02T15:04:05Z"))

		reqBody := GraphQLRequest{
			Query: getIssuesQuery,
			Variables: map[string]interface{}{
				"searchQuery": searchQuery,
				"cursor":      cursor,
			},
		}

		reqBytes, err := json.Marshal(reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal graphQL request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "https://api.github.com/graphql", bytes.NewBuffer(reqBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create http request: %w", err)
		}

		req.Header.Set("Authorization", "bearer "+c.token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to execute http request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("github api returned non-200 status: %s", resp.Status)
		}

		var graphqlResponse GraphQLResponse
		if err := json.NewDecoder(resp.Body).Decode(&graphqlResponse); err != nil {
			return nil, fmt.Errorf("failed to decode graphQL response: %w", err)
		}

		allIssues = append(allIssues, graphqlResponse.Data.Search.Nodes...)

		if !graphqlResponse.Data.Search.PageInfo.HasNextPage {
			break
		}
		cursor = &graphqlResponse.Data.Search.PageInfo.EndCursor
	}

	return allIssues, nil
}
