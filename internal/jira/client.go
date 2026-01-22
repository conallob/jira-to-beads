package jira

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	pb "github.com/conallob/jira-beads-sync/gen/jira"
)

// Client handles communication with Jira API
type Client struct {
	baseURL    string
	httpClient *http.Client
	username   string
	apiToken   string
	adapter    *Adapter
}

// NewClient creates a new Jira API client
func NewClient(baseURL, username, apiToken string) *Client {
	return &Client{
		baseURL:    strings.TrimSuffix(baseURL, "/"),
		httpClient: &http.Client{},
		username:   username,
		apiToken:   apiToken,
		adapter:    NewAdapter(),
	}
}

// FetchIssue fetches a single issue by key (e.g., "PROJ-123")
func (c *Client) FetchIssue(issueKey string) (*pb.Issue, error) {
	apiURL := fmt.Sprintf("%s/rest/api/2/issue/%s", c.baseURL, issueKey)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.apiToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch issue: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jira API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse single issue into an export with one issue
	var jsonIssue jsonIssue
	if err := json.Unmarshal(body, &jsonIssue); err != nil {
		return nil, fmt.Errorf("failed to parse issue: %w", err)
	}

	issue, err := c.adapter.convertIssue(&jsonIssue)
	if err != nil {
		return nil, fmt.Errorf("failed to convert issue: %w", err)
	}

	return issue, nil
}

// UserInfo represents basic information about a Jira user
type UserInfo struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	Active       bool   `json:"active"`
}

// GetCurrentUser fetches information about the currently authenticated user
// This is useful for validating credentials and testing connectivity
func (c *Client) GetCurrentUser() (*UserInfo, error) {
	apiURL := fmt.Sprintf("%s/rest/api/2/myself", c.baseURL)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.apiToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Jira: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusUnauthorized {
			return nil, fmt.Errorf("authentication failed: invalid username or API token")
		}
		return nil, fmt.Errorf("jira API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var userInfo UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}

	return &userInfo, nil
}

// FetchIssueWithDependencies fetches an issue and all its dependencies recursively
func (c *Client) FetchIssueWithDependencies(issueKey string) (*pb.Export, error) {
	visited := make(map[string]bool)
	issues := make([]*pb.Issue, 0)

	if err := c.fetchRecursive(issueKey, visited, &issues); err != nil {
		return nil, err
	}

	return &pb.Export{Issues: issues}, nil
}

// fetchRecursive recursively fetches an issue and all its related issues
func (c *Client) fetchRecursive(issueKey string, visited map[string]bool, issues *[]*pb.Issue) error {
	if visited[issueKey] {
		return nil
	}

	fmt.Printf("Fetching %s...\n", issueKey)
	visited[issueKey] = true

	issue, err := c.FetchIssue(issueKey)
	if err != nil {
		return fmt.Errorf("failed to fetch %s: %w", issueKey, err)
	}

	*issues = append(*issues, issue)

	// Fetch subtasks
	for _, subtask := range issue.Fields.Subtasks {
		if err := c.fetchRecursive(subtask.Key, visited, issues); err != nil {
			return err
		}
	}

	// Fetch linked issues (dependencies)
	for _, link := range issue.Fields.IssueLinks {
		if link.InwardIssue != nil {
			if err := c.fetchRecursive(link.InwardIssue.Key, visited, issues); err != nil {
				return err
			}
		}
		if link.OutwardIssue != nil {
			if err := c.fetchRecursive(link.OutwardIssue.Key, visited, issues); err != nil {
				return err
			}
		}
	}

	// Fetch parent if it exists and isn't an epic
	if issue.Fields.Parent != nil && issue.Fields.Parent.Fields.IssueType.Name != "Epic" {
		if err := c.fetchRecursive(issue.Fields.Parent.Key, visited, issues); err != nil {
			return err
		}
	}

	return nil
}

// ParseIssueKeyFromURL extracts the issue key from a Jira URL
// Handles URLs like:
// - https://jira.example.com/browse/PROJ-123
// - https://jira.example.com/projects/PROJ/issues/PROJ-123
func ParseIssueKeyFromURL(jiraURL string) (string, error) {
	u, err := url.Parse(jiraURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	path := strings.TrimPrefix(u.Path, "/")
	parts := strings.Split(path, "/")

	// Handle /browse/PROJ-123 format
	if len(parts) >= 2 && parts[0] == "browse" {
		return parts[1], nil
	}

	// Handle /projects/PROJ/issues/PROJ-123 format
	if len(parts) >= 4 && parts[0] == "projects" && parts[2] == "issues" {
		return parts[3], nil
	}

	// Try to find any part that looks like an issue key (PROJECT-123)
	for _, part := range parts {
		if strings.Contains(part, "-") && len(part) > 3 {
			// Basic validation: must have letters followed by dash and numbers
			dashIdx := strings.Index(part, "-")
			if dashIdx > 0 && dashIdx < len(part)-1 {
				return part, nil
			}
		}
	}

	return "", fmt.Errorf("could not extract issue key from URL: %s", jiraURL)
}

// GetBaseURLFromIssueURL extracts the base Jira URL from an issue URL
func GetBaseURLFromIssueURL(jiraURL string) (string, error) {
	u, err := url.Parse(jiraURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	return fmt.Sprintf("%s://%s", u.Scheme, u.Host), nil
}

// SearchIssuesByLabel fetches all issues with a given label using JQL
func (c *Client) SearchIssuesByLabel(label string) ([]string, error) {
	// Build JQL query for label with proper quoting
	// Escape any quotes in the label value
	escapedLabel := strings.ReplaceAll(label, `"`, `\"`)
	jql := fmt.Sprintf(`labels = "%s"`, escapedLabel)
	return c.SearchIssues(jql)
}

// SearchIssues performs a JQL search and returns issue keys
func (c *Client) SearchIssues(jql string) ([]string, error) {
	// URL encode the JQL query
	encodedJQL := url.QueryEscape(jql)
	apiURL := fmt.Sprintf("%s/rest/api/2/search?jql=%s&fields=key&maxResults=1000", c.baseURL, encodedJQL)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.apiToken)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to search issues: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jira API returned status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse search results
	var searchResult struct {
		Issues []struct {
			Key string `json:"key"`
		} `json:"issues"`
		Total int `json:"total"`
	}

	if err := json.Unmarshal(body, &searchResult); err != nil {
		return nil, fmt.Errorf("failed to parse search results: %w", err)
	}

	// Extract issue keys
	issueKeys := make([]string, 0, len(searchResult.Issues))
	for _, issue := range searchResult.Issues {
		issueKeys = append(issueKeys, issue.Key)
	}

	if len(issueKeys) < searchResult.Total {
		fmt.Printf("⚠ Warning: Retrieved %d of %d total issues (pagination limit)\n", len(issueKeys), searchResult.Total)
	}

	return issueKeys, nil
}

// FetchIssuesByLabel fetches all issues with a given label and their dependencies
func (c *Client) FetchIssuesByLabel(label string) (*pb.Export, error) {
	fmt.Printf("Searching for issues with label: %s\n", label)

	issueKeys, err := c.SearchIssuesByLabel(label)
	if err != nil {
		return nil, fmt.Errorf("failed to search by label: %w", err)
	}

	if len(issueKeys) == 0 {
		return nil, fmt.Errorf("no issues found with label: %s", label)
	}

	fmt.Printf("Found %d issue(s) with label %s\n", len(issueKeys), label)
	fmt.Println()

	// Fetch all issues and their dependencies
	visited := make(map[string]bool)
	issues := make([]*pb.Issue, 0)

	for _, key := range issueKeys {
		if err := c.fetchRecursive(key, visited, &issues); err != nil {
			return nil, err
		}
	}

	return &pb.Export{Issues: issues}, nil
}
