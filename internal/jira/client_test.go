package jira

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("https://jira.example.com", "user@example.com", "token123")

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	if client.baseURL != "https://jira.example.com" {
		t.Errorf("Expected baseURL to be 'https://jira.example.com', got '%s'", client.baseURL)
	}

	if client.username != "user@example.com" {
		t.Errorf("Expected username to be 'user@example.com', got '%s'", client.username)
	}

	if client.apiToken != "token123" {
		t.Errorf("Expected apiToken to be 'token123', got '%s'", client.apiToken)
	}

	if client.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}

	if client.adapter == nil {
		t.Error("Expected adapter to be initialized")
	}
}

func TestNewClientTrimsTrailingSlash(t *testing.T) {
	client := NewClient("https://jira.example.com/", "user@example.com", "token123")

	if client.baseURL != "https://jira.example.com" {
		t.Errorf("Expected baseURL to trim trailing slash, got '%s'", client.baseURL)
	}
}

func TestFetchIssue(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request
		if r.URL.Path != "/rest/api/2/issue/PROJ-123" {
			t.Errorf("Expected path '/rest/api/2/issue/PROJ-123', got '%s'", r.URL.Path)
		}

		if r.Method != "GET" {
			t.Errorf("Expected GET method, got '%s'", r.Method)
		}

		// Check authentication
		username, password, ok := r.BasicAuth()
		if !ok {
			t.Error("Expected Basic Auth to be present")
		}
		if username != "user@example.com" {
			t.Errorf("Expected username 'user@example.com', got '%s'", username)
		}
		if password != "token123" {
			t.Errorf("Expected password 'token123', got '%s'", password)
		}

		// Return a mock Jira issue
		response := map[string]interface{}{
			"key": "PROJ-123",
			"id":  "12345",
			"fields": map[string]interface{}{
				"summary":     "Test Issue",
				"description": "Test Description",
				"issuetype": map[string]interface{}{
					"name": "Story",
				},
				"status": map[string]interface{}{
					"name": "In Progress",
					"statusCategory": map[string]interface{}{
						"key": "indeterminate",
					},
				},
				"priority": map[string]interface{}{
					"name": "High",
				},
				"created": "2024-01-01T10:00:00.000+0000",
				"updated": "2024-01-15T14:30:00.000+0000",
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with test server URL
	client := NewClient(server.URL, "user@example.com", "token123")

	// Fetch the issue
	issue, err := client.FetchIssue("PROJ-123")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if issue == nil {
		t.Fatal("Expected issue to be returned, got nil")
	}

	if issue.Key != "PROJ-123" {
		t.Errorf("Expected key 'PROJ-123', got '%s'", issue.Key)
	}

	if issue.Fields.Summary != "Test Issue" {
		t.Errorf("Expected summary 'Test Issue', got '%s'", issue.Fields.Summary)
	}
}

func TestFetchIssueNotFound(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(`{"errorMessages":["Issue does not exist"]}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "token123")

	_, err := client.FetchIssue("NOTFOUND-999")
	if err == nil {
		t.Error("Expected error for non-existent issue, got nil")
	}
}

func TestFetchIssueUnauthorized(t *testing.T) {
	// Create a test server that returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		if _, err := w.Write([]byte(`{"errorMessages":["Invalid credentials"]}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "badtoken")

	_, err := client.FetchIssue("PROJ-123")
	if err == nil {
		t.Error("Expected error for unauthorized request, got nil")
	}
}

func TestFetchIssueWithDependencies(t *testing.T) {
	// Track which issues were fetched
	fetchedIssues := make(map[string]bool)

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issueKey := r.URL.Path[len("/rest/api/2/issue/"):]
		fetchedIssues[issueKey] = true

		var response map[string]interface{}

		switch issueKey {
		case "PROJ-123":
			// Main issue with subtasks and linked issues
			response = map[string]interface{}{
				"key": "PROJ-123",
				"id":  "12345",
				"fields": map[string]interface{}{
					"summary": "Main Issue",
					"issuetype": map[string]interface{}{
						"name": "Story",
					},
					"status": map[string]interface{}{
						"name": "In Progress",
						"statusCategory": map[string]interface{}{
							"key": "indeterminate",
						},
					},
					"priority": map[string]interface{}{
						"name": "High",
					},
					"created": "2024-01-01T10:00:00.000+0000",
					"updated": "2024-01-15T14:30:00.000+0000",
					"subtasks": []map[string]interface{}{
						{"key": "PROJ-124"},
					},
					"issuelinks": []map[string]interface{}{
						{
							"type": map[string]interface{}{
								"name": "Blocks",
							},
							"outwardIssue": map[string]interface{}{
								"key": "PROJ-125",
							},
						},
					},
				},
			}
		case "PROJ-124":
			// Subtask
			response = map[string]interface{}{
				"key": "PROJ-124",
				"id":  "12346",
				"fields": map[string]interface{}{
					"summary": "Subtask",
					"issuetype": map[string]interface{}{
						"name": "Subtask",
					},
					"status": map[string]interface{}{
						"name": "To Do",
						"statusCategory": map[string]interface{}{
							"key": "new",
						},
					},
					"priority": map[string]interface{}{
						"name": "Medium",
					},
					"created": "2024-01-02T10:00:00.000+0000",
					"updated": "2024-01-16T14:30:00.000+0000",
				},
			}
		case "PROJ-125":
			// Linked issue
			response = map[string]interface{}{
				"key": "PROJ-125",
				"id":  "12347",
				"fields": map[string]interface{}{
					"summary": "Linked Issue",
					"issuetype": map[string]interface{}{
						"name": "Task",
					},
					"status": map[string]interface{}{
						"name": "Done",
						"statusCategory": map[string]interface{}{
							"key": "done",
						},
					},
					"priority": map[string]interface{}{
						"name": "Low",
					},
					"created": "2024-01-03T10:00:00.000+0000",
					"updated": "2024-01-17T14:30:00.000+0000",
				},
			}
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "token123")

	// Fetch issue with dependencies
	export, err := client.FetchIssueWithDependencies("PROJ-123")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if export == nil {
		t.Fatal("Expected export to be returned, got nil")
	}

	// Should have fetched 3 issues: main, subtask, linked
	if len(export.Issues) != 3 {
		t.Errorf("Expected 3 issues, got %d", len(export.Issues))
	}

	// Verify all expected issues were fetched
	expectedIssues := []string{"PROJ-123", "PROJ-124", "PROJ-125"}
	for _, key := range expectedIssues {
		if !fetchedIssues[key] {
			t.Errorf("Expected issue %s to be fetched", key)
		}
	}

	// Verify issues are in the export
	issueKeys := make(map[string]bool)
	for _, issue := range export.Issues {
		issueKeys[issue.Key] = true
	}

	for _, key := range expectedIssues {
		if !issueKeys[key] {
			t.Errorf("Expected issue %s in export", key)
		}
	}
}

func TestFetchIssueWithDependenciesCircular(t *testing.T) {
	// Test that circular dependencies don't cause infinite loops
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issueKey := r.URL.Path[len("/rest/api/2/issue/"):]

		var response map[string]interface{}

		switch issueKey {
		case "PROJ-1":
			response = map[string]interface{}{
				"key": "PROJ-1",
				"id":  "1",
				"fields": map[string]interface{}{
					"summary": "Issue 1",
					"issuetype": map[string]interface{}{
						"name": "Story",
					},
					"status": map[string]interface{}{
						"name": "Open",
						"statusCategory": map[string]interface{}{
							"key": "new",
						},
					},
					"priority": map[string]interface{}{
						"name": "Medium",
					},
					"created": "2024-01-01T10:00:00.000+0000",
					"updated": "2024-01-01T10:00:00.000+0000",
					"issuelinks": []map[string]interface{}{
						{
							"type": map[string]interface{}{
								"name": "Blocks",
							},
							"outwardIssue": map[string]interface{}{
								"key": "PROJ-2",
							},
						},
					},
				},
			}
		case "PROJ-2":
			// Links back to PROJ-1, creating a cycle
			response = map[string]interface{}{
				"key": "PROJ-2",
				"id":  "2",
				"fields": map[string]interface{}{
					"summary": "Issue 2",
					"issuetype": map[string]interface{}{
						"name": "Story",
					},
					"status": map[string]interface{}{
						"name": "Open",
						"statusCategory": map[string]interface{}{
							"key": "new",
						},
					},
					"priority": map[string]interface{}{
						"name": "Medium",
					},
					"created": "2024-01-01T10:00:00.000+0000",
					"updated": "2024-01-01T10:00:00.000+0000",
					"issuelinks": []map[string]interface{}{
						{
							"type": map[string]interface{}{
								"name": "Blocks",
							},
							"outwardIssue": map[string]interface{}{
								"key": "PROJ-1", // Circular reference
							},
						},
					},
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "token123")

	// Should handle circular dependency without infinite loop
	export, err := client.FetchIssueWithDependencies("PROJ-1")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should have exactly 2 issues (not infinite)
	if len(export.Issues) != 2 {
		t.Errorf("Expected 2 issues (no duplicates), got %d", len(export.Issues))
	}
}

func TestParseIssueKeyFromURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectedKey string
		expectError bool
	}{
		{
			name:        "browse URL",
			url:         "https://jira.example.com/browse/PROJ-123",
			expectedKey: "PROJ-123",
			expectError: false,
		},
		{
			name:        "projects URL",
			url:         "https://jira.example.com/projects/PROJ/issues/PROJ-123",
			expectedKey: "PROJ-123",
			expectError: false,
		},
		{
			name:        "URL with query parameters",
			url:         "https://jira.example.com/browse/PROJ-456?filter=12345",
			expectedKey: "PROJ-456",
			expectError: false,
		},
		{
			name:        "invalid URL",
			url:         "not a valid url",
			expectedKey: "",
			expectError: true,
		},
		{
			name:        "URL without issue key",
			url:         "https://jira.example.com/dashboard",
			expectedKey: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key, err := ParseIssueKeyFromURL(tt.url)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if key != tt.expectedKey {
					t.Errorf("Expected key '%s', got '%s'", tt.expectedKey, key)
				}
			}
		})
	}
}

func TestGetBaseURLFromIssueURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectedURL string
		expectError bool
	}{
		{
			name:        "https URL",
			url:         "https://jira.example.com/browse/PROJ-123",
			expectedURL: "https://jira.example.com",
			expectError: false,
		},
		{
			name:        "http URL",
			url:         "http://jira.example.com/browse/PROJ-123",
			expectedURL: "http://jira.example.com",
			expectError: false,
		},
		{
			name:        "URL with port",
			url:         "https://jira.example.com:8080/browse/PROJ-123",
			expectedURL: "https://jira.example.com:8080",
			expectError: false,
		},
		{
			name:        "URL without scheme",
			url:         "://jira.example.com/browse/PROJ-123",
			expectedURL: "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseURL, err := GetBaseURLFromIssueURL(tt.url)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if baseURL != tt.expectedURL {
					t.Errorf("Expected base URL '%s', got '%s'", tt.expectedURL, baseURL)
				}
			}
		})
	}
}

func TestFetchRecursiveSkipsEpicParents(t *testing.T) {
	// Test that parent issues that are epics are not fetched
	fetchedIssues := make(map[string]bool)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issueKey := r.URL.Path[len("/rest/api/2/issue/"):]
		fetchedIssues[issueKey] = true

		var response map[string]interface{}

		switch issueKey {
		case "PROJ-123":
			response = map[string]interface{}{
				"key": "PROJ-123",
				"id":  "123",
				"fields": map[string]interface{}{
					"summary": "Story with Epic parent",
					"issuetype": map[string]interface{}{
						"name": "Story",
					},
					"status": map[string]interface{}{
						"name": "Open",
						"statusCategory": map[string]interface{}{
							"key": "new",
						},
					},
					"priority": map[string]interface{}{
						"name": "Medium",
					},
					"created": "2024-01-01T10:00:00.000+0000",
					"updated": "2024-01-01T10:00:00.000+0000",
					"parent": map[string]interface{}{
						"key": "EPIC-100",
						"fields": map[string]interface{}{
							"issuetype": map[string]interface{}{
								"name": "Epic", // This should NOT be fetched
							},
						},
					},
				},
			}
		case "EPIC-100":
			// This should never be requested
			t.Error("Epic parent should not be fetched")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "token123")

	export, err := client.FetchIssueWithDependencies("PROJ-123")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should only have PROJ-123, not EPIC-100
	if len(export.Issues) != 1 {
		t.Errorf("Expected 1 issue (epic parent excluded), got %d", len(export.Issues))
	}

	if fetchedIssues["EPIC-100"] {
		t.Error("Epic parent should not have been fetched")
	}
}

func TestFetchRecursiveFetchesNonEpicParents(t *testing.T) {
	// Test that parent issues that are NOT epics ARE fetched
	fetchedIssues := make(map[string]bool)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issueKey := r.URL.Path[len("/rest/api/2/issue/"):]
		fetchedIssues[issueKey] = true

		var response map[string]interface{}

		switch issueKey {
		case "PROJ-124":
			response = map[string]interface{}{
				"key": "PROJ-124",
				"id":  "124",
				"fields": map[string]interface{}{
					"summary": "Subtask with Story parent",
					"issuetype": map[string]interface{}{
						"name": "Subtask",
					},
					"status": map[string]interface{}{
						"name": "Open",
						"statusCategory": map[string]interface{}{
							"key": "new",
						},
					},
					"priority": map[string]interface{}{
						"name": "Medium",
					},
					"created": "2024-01-01T10:00:00.000+0000",
					"updated": "2024-01-01T10:00:00.000+0000",
					"parent": map[string]interface{}{
						"key": "PROJ-123",
						"fields": map[string]interface{}{
							"issuetype": map[string]interface{}{
								"name": "Story", // This SHOULD be fetched
							},
						},
					},
				},
			}
		case "PROJ-123":
			response = map[string]interface{}{
				"key": "PROJ-123",
				"id":  "123",
				"fields": map[string]interface{}{
					"summary": "Parent Story",
					"issuetype": map[string]interface{}{
						"name": "Story",
					},
					"status": map[string]interface{}{
						"name": "Open",
						"statusCategory": map[string]interface{}{
							"key": "new",
						},
					},
					"priority": map[string]interface{}{
						"name": "High",
					},
					"created": "2024-01-01T10:00:00.000+0000",
					"updated": "2024-01-01T10:00:00.000+0000",
				},
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "token123")

	export, err := client.FetchIssueWithDependencies("PROJ-124")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should have both PROJ-124 and its parent PROJ-123
	if len(export.Issues) != 2 {
		t.Errorf("Expected 2 issues (subtask + parent), got %d", len(export.Issues))
	}

	if !fetchedIssues["PROJ-123"] {
		t.Error("Non-epic parent should have been fetched")
	}
}

func TestFetchIssueInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{invalid json`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "token123")

	_, err := client.FetchIssue("PROJ-123")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestFetchIssueWithBothInwardAndOutwardLinks(t *testing.T) {
	fetchedIssues := make(map[string]bool)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		issueKey := r.URL.Path[len("/rest/api/2/issue/"):]
		fetchedIssues[issueKey] = true

		var response map[string]interface{}

		switch issueKey {
		case "PROJ-100":
			response = map[string]interface{}{
				"key": "PROJ-100",
				"id":  "100",
				"fields": map[string]interface{}{
					"summary": "Issue with inward and outward links",
					"issuetype": map[string]interface{}{
						"name": "Story",
					},
					"status": map[string]interface{}{
						"name": "Open",
						"statusCategory": map[string]interface{}{
							"key": "new",
						},
					},
					"priority": map[string]interface{}{
						"name": "Medium",
					},
					"created": "2024-01-01T10:00:00.000+0000",
					"updated": "2024-01-01T10:00:00.000+0000",
					"issuelinks": []map[string]interface{}{
						{
							"type": map[string]interface{}{
								"name": "Blocks",
							},
							"outwardIssue": map[string]interface{}{
								"key": "PROJ-101",
							},
						},
						{
							"type": map[string]interface{}{
								"name": "Depends",
							},
							"inwardIssue": map[string]interface{}{
								"key": "PROJ-102",
							},
						},
					},
				},
			}
		case "PROJ-101":
			response = createMinimalIssue("PROJ-101", "Outward linked")
		case "PROJ-102":
			response = createMinimalIssue("PROJ-102", "Inward linked")
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "token123")

	export, err := client.FetchIssueWithDependencies("PROJ-100")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should have all 3 issues
	if len(export.Issues) != 3 {
		t.Errorf("Expected 3 issues, got %d", len(export.Issues))
	}

	// Verify both inward and outward linked issues were fetched
	if !fetchedIssues["PROJ-101"] {
		t.Error("Outward linked issue should have been fetched")
	}
	if !fetchedIssues["PROJ-102"] {
		t.Error("Inward linked issue should have been fetched")
	}
}

// Helper function to create minimal issue response
func createMinimalIssue(key, summary string) map[string]interface{} {
	return map[string]interface{}{
		"key": key,
		"id":  fmt.Sprintf("%s-id", key),
		"fields": map[string]interface{}{
			"summary": summary,
			"issuetype": map[string]interface{}{
				"name": "Task",
			},
			"status": map[string]interface{}{
				"name": "Open",
				"statusCategory": map[string]interface{}{
					"key": "new",
				},
			},
			"priority": map[string]interface{}{
				"name": "Medium",
			},
			"created": "2024-01-01T10:00:00.000+0000",
			"updated": "2024-01-01T10:00:00.000+0000",
		},
	}
}

func TestSearchIssues(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the request path
		if r.URL.Path != "/rest/api/2/search" {
			t.Errorf("Expected path '/rest/api/2/search', got '%s'", r.URL.Path)
		}

		// Verify authentication
		username, password, ok := r.BasicAuth()
		if !ok {
			t.Error("Expected Basic Auth to be present")
		}
		if username != "user@example.com" {
			t.Errorf("Expected username 'user@example.com', got '%s'", username)
		}
		if password != "token123" {
			t.Errorf("Expected password 'token123', got '%s'", password)
		}

		// Return mock search results
		response := map[string]interface{}{
			"issues": []map[string]interface{}{
				{"key": "PROJ-100"},
				{"key": "PROJ-101"},
				{"key": "PROJ-102"},
			},
			"total": 3,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "token123")

	issueKeys, err := client.SearchIssues("project = PROJ")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(issueKeys) != 3 {
		t.Errorf("Expected 3 issue keys, got %d", len(issueKeys))
	}

	expectedKeys := []string{"PROJ-100", "PROJ-101", "PROJ-102"}
	for i, key := range expectedKeys {
		if issueKeys[i] != key {
			t.Errorf("Expected key '%s' at index %d, got '%s'", key, i, issueKeys[i])
		}
	}
}

func TestSearchIssuesWithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate pagination warning
		response := map[string]interface{}{
			"issues": []map[string]interface{}{
				{"key": "PROJ-1"},
				{"key": "PROJ-2"},
			},
			"total": 1500, // More than maxResults
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "token123")

	issueKeys, err := client.SearchIssues("project = PROJ")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should still return the issues, just with a warning
	if len(issueKeys) != 2 {
		t.Errorf("Expected 2 issue keys, got %d", len(issueKeys))
	}
}

func TestSearchIssuesByLabel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that the JQL query is properly quoted
		jql := r.URL.Query().Get("jql")
		if jql != `labels = "sprint-23"` {
			t.Errorf("Expected JQL 'labels = \"sprint-23\"', got '%s'", jql)
		}

		response := map[string]interface{}{
			"issues": []map[string]interface{}{
				{"key": "PROJ-100"},
				{"key": "PROJ-101"},
			},
			"total": 2,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "token123")

	issueKeys, err := client.SearchIssuesByLabel("sprint-23")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(issueKeys) != 2 {
		t.Errorf("Expected 2 issue keys, got %d", len(issueKeys))
	}
}

func TestSearchIssuesByLabelWithSpaces(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that labels with spaces are properly quoted
		jql := r.URL.Query().Get("jql")
		if jql != `labels = "my feature"` {
			t.Errorf("Expected JQL 'labels = \"my feature\"', got '%s'", jql)
		}

		response := map[string]interface{}{
			"issues": []map[string]interface{}{
				{"key": "PROJ-200"},
			},
			"total": 1,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "token123")

	issueKeys, err := client.SearchIssuesByLabel("my feature")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(issueKeys) != 1 {
		t.Errorf("Expected 1 issue key, got %d", len(issueKeys))
	}
}

func TestSearchIssuesByLabelWithQuotes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that quotes in labels are escaped
		jql := r.URL.Query().Get("jql")
		expectedJQL := `labels = "fix \"bug\" here"`
		if jql != expectedJQL {
			t.Errorf("Expected JQL '%s', got '%s'", expectedJQL, jql)
		}

		response := map[string]interface{}{
			"issues": []map[string]interface{}{
				{"key": "PROJ-300"},
			},
			"total": 1,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "token123")

	issueKeys, err := client.SearchIssuesByLabel(`fix "bug" here`)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(issueKeys) != 1 {
		t.Errorf("Expected 1 issue key, got %d", len(issueKeys))
	}
}

func TestFetchIssuesByLabel(t *testing.T) {
	fetchedIssues := make(map[string]bool)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rest/api/2/search":
			// Return search results
			response := map[string]interface{}{
				"issues": []map[string]interface{}{
					{"key": "PROJ-100"},
					{"key": "PROJ-101"},
				},
				"total": 2,
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Errorf("Failed to encode response: %v", err)
			}
		case "/rest/api/2/issue/PROJ-100", "/rest/api/2/issue/PROJ-101":
			// Fetch individual issues
			issueKey := r.URL.Path[len("/rest/api/2/issue/"):]
			fetchedIssues[issueKey] = true

			response := createMinimalIssue(issueKey, fmt.Sprintf("Issue %s", issueKey))

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Errorf("Failed to encode response: %v", err)
			}
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "token123")

	export, err := client.FetchIssuesByLabel("sprint-23")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if export == nil {
		t.Fatal("Expected export to be returned, got nil")
	}

	if len(export.Issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(export.Issues))
	}

	// Verify both issues were fetched
	if !fetchedIssues["PROJ-100"] {
		t.Error("Expected PROJ-100 to be fetched")
	}
	if !fetchedIssues["PROJ-101"] {
		t.Error("Expected PROJ-101 to be fetched")
	}
}

func TestFetchIssuesByLabelWithDependencies(t *testing.T) {
	fetchedIssues := make(map[string]bool)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rest/api/2/search":
			// Return search results
			response := map[string]interface{}{
				"issues": []map[string]interface{}{
					{"key": "PROJ-100"},
				},
				"total": 1,
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Errorf("Failed to encode response: %v", err)
			}
		default:
			// Fetch individual issues
			issueKey := r.URL.Path[len("/rest/api/2/issue/"):]
			fetchedIssues[issueKey] = true

			var response map[string]interface{}

			switch issueKey {
			case "PROJ-100":
				response = map[string]interface{}{
					"key": "PROJ-100",
					"id":  "100",
					"fields": map[string]interface{}{
						"summary": "Main issue",
						"issuetype": map[string]interface{}{
							"name": "Story",
						},
						"status": map[string]interface{}{
							"name": "Open",
							"statusCategory": map[string]interface{}{
								"key": "new",
							},
						},
						"priority": map[string]interface{}{
							"name": "Medium",
						},
						"created": "2024-01-01T10:00:00.000+0000",
						"updated": "2024-01-01T10:00:00.000+0000",
						"subtasks": []map[string]interface{}{
							{"key": "PROJ-101"},
						},
					},
				}
			case "PROJ-101":
				response = createMinimalIssue("PROJ-101", "Subtask")
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Errorf("Failed to encode response: %v", err)
			}
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "token123")

	export, err := client.FetchIssuesByLabel("sprint-23")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Should have fetched both the labeled issue and its subtask
	if len(export.Issues) != 2 {
		t.Errorf("Expected 2 issues (main + subtask), got %d", len(export.Issues))
	}

	// Verify both were fetched
	if !fetchedIssues["PROJ-100"] {
		t.Error("Expected PROJ-100 to be fetched")
	}
	if !fetchedIssues["PROJ-101"] {
		t.Error("Expected PROJ-101 (subtask) to be fetched")
	}
}

func TestFetchIssuesByLabelNoResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty search results
		response := map[string]interface{}{
			"issues": []map[string]interface{}{},
			"total":  0,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "token123")

	_, err := client.FetchIssuesByLabel("nonexistent-label")
	if err == nil {
		t.Error("Expected error for label with no results, got nil")
	}

	expectedError := "no issues found with label: nonexistent-label"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestSearchIssuesUnauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		if _, err := w.Write([]byte(`{"errorMessages":["Invalid credentials"]}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "badtoken")

	_, err := client.SearchIssues("project = PROJ")
	if err == nil {
		t.Error("Expected error for unauthorized request, got nil")
	}
}

func TestSearchIssuesInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{invalid json`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "user@example.com", "token123")

	_, err := client.SearchIssues("project = PROJ")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}
