package converter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewPipeline(t *testing.T) {
	pipeline := NewPipeline("/tmp/test")
	if pipeline == nil {
		t.Fatal("NewPipeline returned nil")
	}
	if pipeline.jiraAdapter == nil {
		t.Error("jiraAdapter is nil")
	}
	if pipeline.converter == nil {
		t.Error("converter is nil")
	}
	if pipeline.yamlRenderer == nil {
		t.Error("yamlRenderer is nil")
	}
}

func TestPipelineConvertFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pipeline-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	pipeline := NewPipeline(tmpDir)

	err = pipeline.ConvertFile("../../testdata/sample-jira-export.json")
	if err != nil {
		t.Fatalf("ConvertFile failed: %v", err)
	}

	// Verify output directories exist
	issuesDir := filepath.Join(tmpDir, ".beads", "issues")
	epicsDir := filepath.Join(tmpDir, ".beads", "epics")

	if _, err := os.Stat(issuesDir); os.IsNotExist(err) {
		t.Error("Issues directory was not created")
	}

	if _, err := os.Stat(epicsDir); os.IsNotExist(err) {
		t.Error("Epics directory was not created")
	}

	// Verify epic file exists
	epicFile := filepath.Join(epicsDir, "proj-1.yaml")
	if _, err := os.Stat(epicFile); os.IsNotExist(err) {
		t.Error("Epic file proj-1.yaml was not created")
	}

	// Verify issue files exist
	expectedIssues := []string{"proj-2.yaml", "proj-3.yaml", "proj-4.yaml"}
	for _, issueFile := range expectedIssues {
		path := filepath.Join(issuesDir, issueFile)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Issue file %s was not created", issueFile)
		}
	}

	// Read and verify content of one issue
	proj2File := filepath.Join(issuesDir, "proj-2.yaml")
	content, err := os.ReadFile(proj2File)
	if err != nil {
		t.Fatalf("Failed to read proj-2.yaml: %v", err)
	}

	contentStr := string(content)

	// Verify key fields are present
	expectedFields := []string{
		"id: proj-2",
		"title: Create login API endpoint",
		"status: open",
		"priority: p1",
		"epic: proj-1",
		"dependsOn:",
		"- proj-4",
		"jiraKey: PROJ-2",
	}

	for _, field := range expectedFields {
		if !contains([]string{contentStr}, field) {
			// Check if field exists in content
			found := false
			for _, line := range []string{contentStr} {
				if containsSubstring(line, field) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected field '%s' not found in proj-2.yaml.\nContent:\n%s", field, contentStr)
			}
		}
	}
}

func TestPipelineConvertFileInvalid(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pipeline-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	pipeline := NewPipeline(tmpDir)

	// Test with non-existent file
	err = pipeline.ConvertFile("nonexistent.json")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestPipelineEndToEnd(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pipeline-e2e-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	pipeline := NewPipeline(tmpDir)

	// Convert the sample file
	err = pipeline.ConvertFile("../../testdata/sample-jira-export.json")
	if err != nil {
		t.Fatalf("ConvertFile failed: %v", err)
	}

	// Verify the conversion produced correct structure:
	// - 1 epic (PROJ-1)
	// - 3 issues (PROJ-2, PROJ-3, PROJ-4)
	// - PROJ-2 depends on PROJ-4
	// - PROJ-2 and PROJ-3 are linked to epic PROJ-1

	epicsDir := filepath.Join(tmpDir, ".beads", "epics")
	issuesDir := filepath.Join(tmpDir, ".beads", "issues")

	// Count epics
	epics, err := os.ReadDir(epicsDir)
	if err != nil {
		t.Fatalf("Failed to read epics dir: %v", err)
	}
	if len(epics) != 1 {
		t.Errorf("Expected 1 epic, got %d", len(epics))
	}

	// Count issues
	issues, err := os.ReadDir(issuesDir)
	if err != nil {
		t.Fatalf("Failed to read issues dir: %v", err)
	}
	if len(issues) != 3 {
		t.Errorf("Expected 3 issues, got %d", len(issues))
	}

	// Verify PROJ-2 has correct dependencies and epic link
	proj2Content, err := os.ReadFile(filepath.Join(issuesDir, "proj-2.yaml"))
	if err != nil {
		t.Fatalf("Failed to read proj-2.yaml: %v", err)
	}

	proj2Str := string(proj2Content)
	if !containsSubstring(proj2Str, "epic: proj-1") {
		t.Error("PROJ-2 should be linked to epic proj-1")
	}
	if !containsSubstring(proj2Str, "proj-4") {
		t.Error("PROJ-2 should depend on proj-4")
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
