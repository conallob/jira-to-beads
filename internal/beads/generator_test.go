package beads

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator("/tmp/test")
	if gen == nil {
		t.Fatal("NewGenerator returned nil")
	}
	if gen.outputDir != "/tmp/test" {
		t.Errorf("Expected outputDir /tmp/test, got %s", gen.outputDir)
	}
}

func TestGenerateToString(t *testing.T) {
	gen := NewGenerator("/tmp/test")

	issue := &Issue{
		ID:          "issue-1",
		Title:       "Test Issue",
		Description: "This is a test issue",
		Status:      StatusOpen,
		Priority:    PriorityP1,
		Labels:      []string{"test", "example"},
		DependsOn:   []string{"issue-2"},
		Created:     time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		Updated:     time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
		Metadata: Metadata{
			JiraKey:       "PROJ-1",
			JiraID:        "10001",
			JiraIssueType: "Story",
		},
	}

	yaml, err := gen.GenerateToString(issue)
	if err != nil {
		t.Fatalf("GenerateToString failed: %v", err)
	}

	if yaml == "" {
		t.Error("Generated YAML is empty")
	}

	// Check that key fields are present in the YAML
	expectedFields := []string{"id:", "title:", "status:", "priority:", "jiraKey:"}
	for _, field := range expectedFields {
		if !containsString(yaml, field) {
			t.Errorf("Expected YAML to contain %s, but it doesn't", field)
		}
	}
}

func TestGenerateEpicToString(t *testing.T) {
	gen := NewGenerator("/tmp/test")

	epic := &Epic{
		ID:          "epic-1",
		Name:        "Test Epic",
		Description: "This is a test epic",
		Status:      StatusOpen,
		Created:     time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		Updated:     time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
		Metadata: Metadata{
			JiraKey: "PROJ-1",
			JiraID:  "10001",
		},
	}

	yaml, err := gen.GenerateEpicToString(epic)
	if err != nil {
		t.Fatalf("GenerateEpicToString failed: %v", err)
	}

	if yaml == "" {
		t.Error("Generated YAML is empty")
	}

	expectedFields := []string{"id:", "name:", "status:", "jiraKey:"}
	for _, field := range expectedFields {
		if !containsString(yaml, field) {
			t.Errorf("Expected YAML to contain %s, but it doesn't", field)
		}
	}
}

func TestGenerate(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "beads-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	gen := NewGenerator(tmpDir)

	export := &Export{
		Issues: []Issue{
			{
				ID:       "issue-1",
				Title:    "Test Issue 1",
				Status:   StatusOpen,
				Priority: PriorityP1,
				Created:  time.Now(),
				Updated:  time.Now(),
			},
			{
				ID:       "issue-2",
				Title:    "Test Issue 2",
				Status:   StatusClosed,
				Priority: PriorityP2,
				Created:  time.Now(),
				Updated:  time.Now(),
			},
		},
		Epics: []Epic{
			{
				ID:      "epic-1",
				Name:    "Test Epic",
				Status:  StatusOpen,
				Created: time.Now(),
				Updated: time.Now(),
			},
		},
	}

	err = gen.Generate(export)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
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

func TestParseIssueFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "beads-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	gen := NewGenerator(tmpDir)

	originalIssue := &Issue{
		ID:          "issue-1",
		Title:       "Test Issue",
		Description: "Test Description",
		Status:      StatusOpen,
		Priority:    PriorityP1,
		Labels:      []string{"test"},
		Created:     time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		Updated:     time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
	}

	export := &Export{
		Issues: []Issue{*originalIssue},
	}

	if err := gen.Generate(export); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	issueFile := filepath.Join(tmpDir, ".beads", "issues", "issue-1.yaml")
	parsedIssue, err := gen.ParseIssueFile(issueFile)
	if err != nil {
		t.Fatalf("ParseIssueFile failed: %v", err)
	}

	if parsedIssue.ID != originalIssue.ID {
		t.Errorf("Expected ID %s, got %s", originalIssue.ID, parsedIssue.ID)
	}
	if parsedIssue.Title != originalIssue.Title {
		t.Errorf("Expected title %s, got %s", originalIssue.Title, parsedIssue.Title)
	}
	if parsedIssue.Status != originalIssue.Status {
		t.Errorf("Expected status %s, got %s", originalIssue.Status, parsedIssue.Status)
	}
}

func TestParseEpicFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "beads-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	gen := NewGenerator(tmpDir)

	originalEpic := &Epic{
		ID:      "epic-1",
		Name:    "Test Epic",
		Status:  StatusOpen,
		Created: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		Updated: time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC),
	}

	export := &Export{
		Epics: []Epic{*originalEpic},
	}

	if err := gen.Generate(export); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	epicFile := filepath.Join(tmpDir, ".beads", "epics", "epic-1.yaml")
	parsedEpic, err := gen.ParseEpicFile(epicFile)
	if err != nil {
		t.Fatalf("ParseEpicFile failed: %v", err)
	}

	if parsedEpic.ID != originalEpic.ID {
		t.Errorf("Expected ID %s, got %s", originalEpic.ID, parsedEpic.ID)
	}
	if parsedEpic.Name != originalEpic.Name {
		t.Errorf("Expected name %s, got %s", originalEpic.Name, parsedEpic.Name)
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && contains(s, substr))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
