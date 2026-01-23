package beads

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pb "github.com/conallob/jira-beads-sync/gen/beads"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewJSONLRenderer(t *testing.T) {
	renderer := NewJSONLRenderer("/tmp/test")
	if renderer == nil {
		t.Fatal("NewJSONLRenderer returned nil")
	}
	if renderer.outputDir != "/tmp/test" {
		t.Errorf("Expected outputDir /tmp/test, got %s", renderer.outputDir)
	}
}

func TestRenderExport(t *testing.T) {
	tmpDir := t.TempDir()
	renderer := NewJSONLRenderer(tmpDir)

	export := &pb.Export{
		Issues: []*pb.Issue{
			{
				Id:          "issue-1",
				Title:       "Test Issue 1",
				Description: "This is a test issue",
				Status:      pb.Status_STATUS_OPEN,
				Priority:    pb.Priority_PRIORITY_P1,
				Labels:      []string{"test", "example"},
				DependsOn:   []string{"issue-2"},
				Created:     timestamppb.Now(),
				Updated:     timestamppb.Now(),
				Metadata: &pb.Metadata{
					JiraKey:       "PROJ-1",
					JiraId:        "10001",
					JiraIssueType: "Story",
				},
			},
			{
				Id:       "issue-2",
				Title:    "Test Issue 2",
				Status:   pb.Status_STATUS_CLOSED,
				Priority: pb.Priority_PRIORITY_P2,
				Created:  timestamppb.Now(),
				Updated:  timestamppb.Now(),
			},
		},
		Epics: []*pb.Epic{
			{
				Id:      "epic-1",
				Name:    "Test Epic",
				Status:  pb.Status_STATUS_OPEN,
				Created: timestamppb.Now(),
				Updated: timestamppb.Now(),
			},
		},
	}

	err := renderer.RenderExport(export)
	if err != nil {
		t.Fatalf("RenderExport failed: %v", err)
	}

	// Verify .beads directory was created
	beadsDir := filepath.Join(tmpDir, ".beads")
	if _, err := os.Stat(beadsDir); os.IsNotExist(err) {
		t.Error("Beads directory was not created")
	}

	// Verify issues.jsonl was created
	issuesFile := filepath.Join(beadsDir, "issues.jsonl")
	if _, err := os.Stat(issuesFile); os.IsNotExist(err) {
		t.Error("Issues JSONL file was not created")
	}

	// Verify epics.jsonl was created
	epicsFile := filepath.Join(beadsDir, "epics.jsonl")
	if _, err := os.Stat(epicsFile); os.IsNotExist(err) {
		t.Error("Epics JSONL file was not created")
	}

	// Parse and verify issues.jsonl content
	issuesData, err := os.Open(issuesFile)
	if err != nil {
		t.Fatalf("Failed to open issues.jsonl: %v", err)
	}
	defer func() {
		_ = issuesData.Close()
	}()

	scanner := bufio.NewScanner(issuesData)
	issueCount := 0
	for scanner.Scan() {
		issueCount++
		var issue BeadsIssue
		if err := json.Unmarshal(scanner.Bytes(), &issue); err != nil {
			t.Errorf("Failed to parse JSON line %d: %v", issueCount, err)
		}

		// Verify required fields
		if issue.ID == "" {
			t.Error("Issue ID is empty")
		}
		if issue.Title == "" {
			t.Error("Issue title is empty")
		}
		if issue.Status == "" {
			t.Error("Issue status is empty")
		}
	}

	if issueCount != 2 {
		t.Errorf("Expected 2 issues, got %d", issueCount)
	}
}

func TestIssueToJSON(t *testing.T) {
	renderer := NewJSONLRenderer("/tmp/test")

	issue := &pb.Issue{
		Id:          "test-123",
		Title:       "Test Issue",
		Description: "Test Description",
		Status:      pb.Status_STATUS_IN_PROGRESS,
		Priority:    pb.Priority_PRIORITY_P0,
		Epic:        "epic-1",
		Assignee:    "user@example.com",
		Labels:      []string{"label1", "label2"},
		DependsOn:   []string{"dep-1", "dep-2"},
		Created:     timestamppb.Now(),
		Updated:     timestamppb.Now(),
		Metadata: &pb.Metadata{
			JiraKey:       "PROJ-123",
			JiraId:        "10123",
			JiraIssueType: "Bug",
		},
	}

	jsonIssue := renderer.issueToJSON(issue)

	if jsonIssue.ID != "test-123" {
		t.Errorf("Expected ID 'test-123', got '%s'", jsonIssue.ID)
	}
	if jsonIssue.Title != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got '%s'", jsonIssue.Title)
	}
	if jsonIssue.Description != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", jsonIssue.Description)
	}
	if jsonIssue.Status != "in_progress" {
		t.Errorf("Expected status 'in_progress', got '%s'", jsonIssue.Status)
	}
	if jsonIssue.Priority != "p0" {
		t.Errorf("Expected priority 'p0', got '%s'", jsonIssue.Priority)
	}
	if jsonIssue.Epic != "epic-1" {
		t.Errorf("Expected epic 'epic-1', got '%s'", jsonIssue.Epic)
	}
	if jsonIssue.Assignee != "user@example.com" {
		t.Errorf("Expected assignee 'user@example.com', got '%s'", jsonIssue.Assignee)
	}
	if len(jsonIssue.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(jsonIssue.Labels))
	}
	if len(jsonIssue.DependsOn) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(jsonIssue.DependsOn))
	}
	if jsonIssue.Metadata == nil {
		t.Fatal("Metadata is nil")
	}
	if jsonIssue.Metadata["jiraKey"] != "PROJ-123" {
		t.Errorf("Expected jiraKey 'PROJ-123', got '%s'", jsonIssue.Metadata["jiraKey"])
	}
}

func TestStatusConversion(t *testing.T) {
	renderer := NewJSONLRenderer("/tmp/test")

	tests := []struct {
		status pb.Status
		want   string
	}{
		{pb.Status_STATUS_OPEN, "open"},
		{pb.Status_STATUS_IN_PROGRESS, "in_progress"},
		{pb.Status_STATUS_BLOCKED, "blocked"},
		{pb.Status_STATUS_CLOSED, "closed"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := renderer.statusToString(tt.status)
			if got != tt.want {
				t.Errorf("statusToString(%v) = %s, want %s", tt.status, got, tt.want)
			}
		})
	}
}

func TestPriorityConversion(t *testing.T) {
	renderer := NewJSONLRenderer("/tmp/test")

	tests := []struct {
		priority pb.Priority
		want     string
	}{
		{pb.Priority_PRIORITY_P0, "p0"},
		{pb.Priority_PRIORITY_P1, "p1"},
		{pb.Priority_PRIORITY_P2, "p2"},
		{pb.Priority_PRIORITY_P3, "p3"},
		{pb.Priority_PRIORITY_P4, "p4"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := renderer.priorityToString(tt.priority)
			if got != tt.want {
				t.Errorf("priorityToString(%v) = %s, want %s", tt.priority, got, tt.want)
			}
		})
	}
}

func TestJSONLFormat(t *testing.T) {
	tmpDir := t.TempDir()
	renderer := NewJSONLRenderer(tmpDir)

	export := &pb.Export{
		Issues: []*pb.Issue{
			{
				Id:       "issue-1",
				Title:    "First Issue",
				Status:   pb.Status_STATUS_OPEN,
				Priority: pb.Priority_PRIORITY_P1,
				Created:  timestamppb.Now(),
				Updated:  timestamppb.Now(),
			},
			{
				Id:       "issue-2",
				Title:    "Second Issue",
				Status:   pb.Status_STATUS_CLOSED,
				Priority: pb.Priority_PRIORITY_P2,
				Created:  timestamppb.Now(),
				Updated:  timestamppb.Now(),
			},
		},
	}

	err := renderer.RenderExport(export)
	if err != nil {
		t.Fatalf("RenderExport failed: %v", err)
	}

	// Read the JSONL file
	issuesFile := filepath.Join(tmpDir, ".beads", "issues.jsonl")
	data, err := os.ReadFile(issuesFile)
	if err != nil {
		t.Fatalf("Failed to read issues.jsonl: %v", err)
	}

	// Check that it's valid JSONL (one JSON object per line)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("Expected 2 lines in JSONL, got %d", len(lines))
	}

	for i, line := range lines {
		var issue BeadsIssue
		if err := json.Unmarshal([]byte(line), &issue); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", i+1, err)
		}
	}
}
