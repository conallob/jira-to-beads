package beads

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	pb "github.com/conallob/jira-beads-sync/gen/beads"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewYAMLRenderer(t *testing.T) {
	renderer := NewYAMLRenderer("/tmp/test")
	if renderer == nil {
		t.Fatal("NewYAMLRenderer returned nil")
	}
	if renderer.outputDir != "/tmp/test" {
		t.Errorf("Expected outputDir /tmp/test, got %s", renderer.outputDir)
	}
}

func TestRenderIssueToString(t *testing.T) {
	renderer := NewYAMLRenderer("/tmp/test")

	issue := &pb.Issue{
		Id:          "issue-1",
		Title:       "Test Issue",
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
	}

	yaml, err := renderer.RenderIssueToString(issue)
	if err != nil {
		t.Fatalf("RenderIssueToString failed: %v", err)
	}

	if yaml == "" {
		t.Error("Generated YAML is empty")
	}

	// Check that key fields are present in the YAML
	expectedFields := []string{"id:", "title:", "status:", "priority:", "jiraKey:"}
	for _, field := range expectedFields {
		if !strings.Contains(yaml, field) {
			t.Errorf("Expected YAML to contain %s, but it doesn't.\nYAML:\n%s", field, yaml)
		}
	}

	// Check that enum values are converted to lowercase
	if !strings.Contains(yaml, "status: open") {
		t.Errorf("Expected 'status: open', got:\n%s", yaml)
	}
	if !strings.Contains(yaml, "priority: p1") {
		t.Errorf("Expected 'priority: p1', got:\n%s", yaml)
	}

	// Check that field names are camelCase
	if !strings.Contains(yaml, "dependsOn:") {
		t.Errorf("Expected 'dependsOn:', got:\n%s", yaml)
	}
	if !strings.Contains(yaml, "jiraKey:") {
		t.Errorf("Expected 'jiraKey:', got:\n%s", yaml)
	}
}

func TestRenderEpicToString(t *testing.T) {
	renderer := NewYAMLRenderer("/tmp/test")

	epic := &pb.Epic{
		Id:          "epic-1",
		Name:        "Test Epic",
		Description: "This is a test epic",
		Status:      pb.Status_STATUS_IN_PROGRESS,
		Created:     timestamppb.Now(),
		Updated:     timestamppb.Now(),
		Metadata: &pb.Metadata{
			JiraKey: "PROJ-1",
			JiraId:  "10001",
		},
	}

	yaml, err := renderer.RenderEpicToString(epic)
	if err != nil {
		t.Fatalf("RenderEpicToString failed: %v", err)
	}

	if yaml == "" {
		t.Error("Generated YAML is empty")
	}

	expectedFields := []string{"id:", "name:", "status:", "jiraKey:"}
	for _, field := range expectedFields {
		if !strings.Contains(yaml, field) {
			t.Errorf("Expected YAML to contain %s, but it doesn't.\nYAML:\n%s", field, yaml)
		}
	}

	// Check status conversion
	if !strings.Contains(yaml, "status: in_progress") {
		t.Errorf("Expected 'status: in_progress', got:\n%s", yaml)
	}
}

func TestRenderExport(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "beads-yaml-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	renderer := NewYAMLRenderer(tmpDir)

	export := &pb.Export{
		Issues: []*pb.Issue{
			{
				Id:       "issue-1",
				Title:    "Test Issue 1",
				Status:   pb.Status_STATUS_OPEN,
				Priority: pb.Priority_PRIORITY_P1,
				Created:  timestamppb.Now(),
				Updated:  timestamppb.Now(),
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

	err = renderer.RenderExport(export)
	if err != nil {
		t.Fatalf("RenderExport failed: %v", err)
	}

	// Verify directories were created
	issuesDir := filepath.Join(tmpDir, ".beads", "issues")
	epicsDir := filepath.Join(tmpDir, ".beads", "epics")

	if _, err := os.Stat(issuesDir); os.IsNotExist(err) {
		t.Error("Issues directory was not created")
	}

	if _, err := os.Stat(epicsDir); os.IsNotExist(err) {
		t.Error("Epics directory was not created")
	}

	// Verify issue files were created
	issue1File := filepath.Join(issuesDir, "issue-1.yaml")
	if _, err := os.Stat(issue1File); os.IsNotExist(err) {
		t.Error("Issue 1 file was not created")
	}

	issue2File := filepath.Join(issuesDir, "issue-2.yaml")
	if _, err := os.Stat(issue2File); os.IsNotExist(err) {
		t.Error("Issue 2 file was not created")
	}

	// Verify epic file was created
	epicFile := filepath.Join(epicsDir, "epic-1.yaml")
	if _, err := os.Stat(epicFile); os.IsNotExist(err) {
		t.Error("Epic file was not created")
	}
}

// Note: YAML round-trip parsing tests are skipped because the human-readable YAML format
// doesn't directly map to protobuf JSON format. The ParseIssueFile/ParseEpicFile methods
// would need additional conversion logic to handle the camelCase field names and enum formats.
// For this tool, we only need to write YAML files, not parse them back.

func TestStatusConversion(t *testing.T) {
	renderer := NewYAMLRenderer("/tmp/test")

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
			issue := &pb.Issue{
				Id:       "test",
				Title:    "Test",
				Status:   tt.status,
				Priority: pb.Priority_PRIORITY_P2,
				Created:  timestamppb.Now(),
				Updated:  timestamppb.Now(),
			}

			yaml, err := renderer.RenderIssueToString(issue)
			if err != nil {
				t.Fatalf("RenderIssueToString failed: %v", err)
			}

			expected := "status: " + tt.want
			if !strings.Contains(yaml, expected) {
				t.Errorf("Expected '%s' in YAML, got:\n%s", expected, yaml)
			}
		})
	}
}

func TestPriorityConversion(t *testing.T) {
	renderer := NewYAMLRenderer("/tmp/test")

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
			issue := &pb.Issue{
				Id:       "test",
				Title:    "Test",
				Status:   pb.Status_STATUS_OPEN,
				Priority: tt.priority,
				Created:  timestamppb.Now(),
				Updated:  timestamppb.Now(),
			}

			yaml, err := renderer.RenderIssueToString(issue)
			if err != nil {
				t.Fatalf("RenderIssueToString failed: %v", err)
			}

			expected := "priority: " + tt.want
			if !strings.Contains(yaml, expected) {
				t.Errorf("Expected '%s' in YAML, got:\n%s", expected, yaml)
			}
		})
	}
}

func TestAddRepositoryAnnotation(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	renderer := NewYAMLRenderer(tempDir)

	// Create a test issue
	issue := &pb.Issue{
		Id:          "test-issue",
		Title:       "Test Issue",
		Description: "Test Description",
		Status:      pb.Status_STATUS_OPEN,
		Priority:    pb.Priority_PRIORITY_P1,
		Created:     timestamppb.Now(),
		Updated:     timestamppb.Now(),
		Metadata: &pb.Metadata{
			JiraKey:       "PROJ-123",
			JiraId:        "10001",
			JiraIssueType: "Story",
		},
	}

	// Render the issue to create the file
	if err := renderer.RenderIssue(issue); err != nil {
		t.Fatalf("Failed to render issue: %v", err)
	}

	// Add a repository annotation
	repo1 := "https://github.com/org/repo1"
	if err := renderer.AddRepositoryAnnotation("test-issue", repo1); err != nil {
		t.Fatalf("Failed to add repository annotation: %v", err)
	}

	// Parse the issue back
	filename := filepath.Join(tempDir, ".beads", "issues", "test-issue.yaml")
	updatedIssue, err := renderer.ParseIssueFile(filename)
	if err != nil {
		t.Fatalf("Failed to parse updated issue: %v", err)
	}

	// Verify the repository was added
	if len(updatedIssue.Metadata.Repositories) != 1 {
		t.Fatalf("Expected 1 repository, got %d", len(updatedIssue.Metadata.Repositories))
	}

	if updatedIssue.Metadata.Repositories[0] != repo1 {
		t.Errorf("Expected repository '%s', got '%s'", repo1, updatedIssue.Metadata.Repositories[0])
	}
}

func TestAddRepositoryAnnotationMultiple(t *testing.T) {
	tempDir := t.TempDir()
	renderer := NewYAMLRenderer(tempDir)

	issue := &pb.Issue{
		Id:       "test-issue",
		Title:    "Test Issue",
		Status:   pb.Status_STATUS_OPEN,
		Priority: pb.Priority_PRIORITY_P2,
		Created:  timestamppb.Now(),
		Updated:  timestamppb.Now(),
		Metadata: &pb.Metadata{
			JiraKey: "PROJ-456",
		},
	}

	if err := renderer.RenderIssue(issue); err != nil {
		t.Fatalf("Failed to render issue: %v", err)
	}

	// Add multiple repositories
	repos := []string{
		"https://github.com/org/frontend",
		"https://github.com/org/backend",
		"https://github.com/org/shared",
	}

	for _, repo := range repos {
		if err := renderer.AddRepositoryAnnotation("test-issue", repo); err != nil {
			t.Fatalf("Failed to add repository %s: %v", repo, err)
		}
	}

	// Parse the issue back
	filename := filepath.Join(tempDir, ".beads", "issues", "test-issue.yaml")
	updatedIssue, err := renderer.ParseIssueFile(filename)
	if err != nil {
		t.Fatalf("Failed to parse updated issue: %v", err)
	}

	// Verify all repositories were added
	if len(updatedIssue.Metadata.Repositories) != len(repos) {
		t.Fatalf("Expected %d repositories, got %d", len(repos), len(updatedIssue.Metadata.Repositories))
	}

	for i, expectedRepo := range repos {
		if updatedIssue.Metadata.Repositories[i] != expectedRepo {
			t.Errorf("Expected repository '%s' at index %d, got '%s'", expectedRepo, i, updatedIssue.Metadata.Repositories[i])
		}
	}
}

func TestAddRepositoryAnnotationDuplicate(t *testing.T) {
	tempDir := t.TempDir()
	renderer := NewYAMLRenderer(tempDir)

	issue := &pb.Issue{
		Id:       "test-issue",
		Title:    "Test Issue",
		Status:   pb.Status_STATUS_OPEN,
		Priority: pb.Priority_PRIORITY_P2,
		Created:  timestamppb.Now(),
		Updated:  timestamppb.Now(),
		Metadata: &pb.Metadata{
			JiraKey: "PROJ-789",
		},
	}

	if err := renderer.RenderIssue(issue); err != nil {
		t.Fatalf("Failed to render issue: %v", err)
	}

	repo := "https://github.com/org/repo"

	// Add repository first time
	if err := renderer.AddRepositoryAnnotation("test-issue", repo); err != nil {
		t.Fatalf("Failed to add repository: %v", err)
	}

	// Try to add the same repository again
	err := renderer.AddRepositoryAnnotation("test-issue", repo)
	if err == nil {
		t.Error("Expected error when adding duplicate repository, got nil")
	}

	expectedError := "already associated with issue"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestAddRepositoryAnnotationNonExistentIssue(t *testing.T) {
	tempDir := t.TempDir()
	renderer := NewYAMLRenderer(tempDir)

	// Try to annotate an issue that doesn't exist
	err := renderer.AddRepositoryAnnotation("nonexistent-issue", "https://github.com/org/repo")
	if err == nil {
		t.Error("Expected error for non-existent issue, got nil")
	}

	expectedError := "issue file not found"
	if !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected error to contain '%s', got '%s'", expectedError, err.Error())
	}
}

func TestAddRepositoryAnnotationWithNilMetadata(t *testing.T) {
	tempDir := t.TempDir()
	renderer := NewYAMLRenderer(tempDir)

	// Create issue without metadata
	issue := &pb.Issue{
		Id:       "test-issue",
		Title:    "Test Issue",
		Status:   pb.Status_STATUS_OPEN,
		Priority: pb.Priority_PRIORITY_P2,
		Created:  timestamppb.Now(),
		Updated:  timestamppb.Now(),
		// No Metadata field
	}

	if err := renderer.RenderIssue(issue); err != nil {
		t.Fatalf("Failed to render issue: %v", err)
	}

	// Add repository annotation
	repo := "https://github.com/org/repo"
	if err := renderer.AddRepositoryAnnotation("test-issue", repo); err != nil {
		t.Fatalf("Failed to add repository annotation: %v", err)
	}

	// Parse the issue back
	filename := filepath.Join(tempDir, ".beads", "issues", "test-issue.yaml")
	updatedIssue, err := renderer.ParseIssueFile(filename)
	if err != nil {
		t.Fatalf("Failed to parse updated issue: %v", err)
	}

	// Verify metadata was initialized and repository was added
	if updatedIssue.Metadata == nil {
		t.Fatal("Expected metadata to be initialized, got nil")
	}

	if len(updatedIssue.Metadata.Repositories) != 1 {
		t.Fatalf("Expected 1 repository, got %d", len(updatedIssue.Metadata.Repositories))
	}

	if updatedIssue.Metadata.Repositories[0] != repo {
		t.Errorf("Expected repository '%s', got '%s'", repo, updatedIssue.Metadata.Repositories[0])
	}
}

func TestAddRepositoryAnnotationPreservesOtherFields(t *testing.T) {
	tempDir := t.TempDir()
	renderer := NewYAMLRenderer(tempDir)

	originalIssue := &pb.Issue{
		Id:          "test-issue",
		Title:       "Test Issue",
		Description: "Original description",
		Status:      pb.Status_STATUS_IN_PROGRESS,
		Priority:    pb.Priority_PRIORITY_P0,
		Labels:      []string{"label1", "label2"},
		DependsOn:   []string{"other-issue"},
		Created:     timestamppb.Now(),
		Updated:     timestamppb.Now(),
		Metadata: &pb.Metadata{
			JiraKey:       "PROJ-999",
			JiraId:        "99999",
			JiraIssueType: "Bug",
			Custom: map[string]string{
				"customField1": "value1",
			},
		},
	}

	if err := renderer.RenderIssue(originalIssue); err != nil {
		t.Fatalf("Failed to render issue: %v", err)
	}

	// Add repository annotation
	repo := "https://github.com/org/repo"
	if err := renderer.AddRepositoryAnnotation("test-issue", repo); err != nil {
		t.Fatalf("Failed to add repository annotation: %v", err)
	}

	// Parse the issue back
	filename := filepath.Join(tempDir, ".beads", "issues", "test-issue.yaml")
	updatedIssue, err := renderer.ParseIssueFile(filename)
	if err != nil {
		t.Fatalf("Failed to parse updated issue: %v", err)
	}

	// Verify all original fields are preserved
	if updatedIssue.Title != originalIssue.Title {
		t.Errorf("Title was modified: expected '%s', got '%s'", originalIssue.Title, updatedIssue.Title)
	}

	if updatedIssue.Description != originalIssue.Description {
		t.Errorf("Description was modified: expected '%s', got '%s'", originalIssue.Description, updatedIssue.Description)
	}

	if updatedIssue.Status != originalIssue.Status {
		t.Errorf("Status was modified: expected %v, got %v", originalIssue.Status, updatedIssue.Status)
	}

	if updatedIssue.Priority != originalIssue.Priority {
		t.Errorf("Priority was modified: expected %v, got %v", originalIssue.Priority, updatedIssue.Priority)
	}

	if len(updatedIssue.Labels) != len(originalIssue.Labels) {
		t.Errorf("Labels were modified: expected %d, got %d", len(originalIssue.Labels), len(updatedIssue.Labels))
	}

	if len(updatedIssue.DependsOn) != len(originalIssue.DependsOn) {
		t.Errorf("DependsOn was modified: expected %d, got %d", len(originalIssue.DependsOn), len(updatedIssue.DependsOn))
	}

	if updatedIssue.Metadata.JiraKey != originalIssue.Metadata.JiraKey {
		t.Errorf("JiraKey was modified: expected '%s', got '%s'", originalIssue.Metadata.JiraKey, updatedIssue.Metadata.JiraKey)
	}

	// Verify repository was added
	if len(updatedIssue.Metadata.Repositories) != 1 {
		t.Fatalf("Expected 1 repository, got %d", len(updatedIssue.Metadata.Repositories))
	}

	if updatedIssue.Metadata.Repositories[0] != repo {
		t.Errorf("Expected repository '%s', got '%s'", repo, updatedIssue.Metadata.Repositories[0])
	}
}
