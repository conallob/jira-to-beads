package converter

import (
	"os"
	"path/filepath"
	"strings"
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
	if pipeline.jsonlRenderer == nil {
		t.Error("jsonlRenderer is nil")
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

	// Verify output directory exists
	beadsDir := filepath.Join(tmpDir, ".beads")
	if _, err := os.Stat(beadsDir); os.IsNotExist(err) {
		t.Error("Beads directory was not created")
	}

	// Verify JSONL files exist
	issuesFile := filepath.Join(beadsDir, "issues.jsonl")
	if _, err := os.Stat(issuesFile); os.IsNotExist(err) {
		t.Error("issues.jsonl file was not created")
	}

	epicsFile := filepath.Join(beadsDir, "epics.jsonl")
	if _, err := os.Stat(epicsFile); os.IsNotExist(err) {
		t.Error("epics.jsonl file was not created")
	}

	// Read and verify issues.jsonl content
	content, err := os.ReadFile(issuesFile)
	if err != nil {
		t.Fatalf("Failed to read issues.jsonl: %v", err)
	}

	contentStr := string(content)

	// Verify key fields are present in JSONL format
	expectedFields := []string{
		`"id":"proj-2"`,
		`"title":"Create login API endpoint"`,
		`"status":"open"`,
		`"priority":"p1"`,
		`"epic":"proj-1"`,
		`"dependsOn":["proj-4"]`,
		`"jiraKey":"PROJ-2"`,
	}

	for _, field := range expectedFields {
		if !containsSubstring(contentStr, field) {
			t.Errorf("Expected field '%s' not found in issues.jsonl.\nContent:\n%s", field, contentStr)
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

	beadsDir := filepath.Join(tmpDir, ".beads")
	issuesFile := filepath.Join(beadsDir, "issues.jsonl")
	epicsFile := filepath.Join(beadsDir, "epics.jsonl")

	// Read and check epics JSONL
	epicsContent, err := os.ReadFile(epicsFile)
	if err != nil {
		t.Fatalf("Failed to read epics.jsonl: %v", err)
	}
	epicsStr := string(epicsContent)
	// Count lines (each line is one epic)
	epicLines := 0
	for _, line := range splitLines(epicsStr) {
		if len(line) > 0 {
			epicLines++
		}
	}
	if epicLines != 1 {
		t.Errorf("Expected 1 epic, got %d", epicLines)
	}

	// Read and check issues JSONL
	issuesContent, err := os.ReadFile(issuesFile)
	if err != nil {
		t.Fatalf("Failed to read issues.jsonl: %v", err)
	}
	issuesStr := string(issuesContent)
	// Count lines (each line is one issue)
	issueLines := 0
	for _, line := range splitLines(issuesStr) {
		if len(line) > 0 {
			issueLines++
		}
	}
	if issueLines != 3 {
		t.Errorf("Expected 3 issues, got %d", issueLines)
	}

	// Verify PROJ-2 has correct dependencies and epic link
	if !containsSubstring(issuesStr, `"id":"proj-2"`) {
		t.Error("PROJ-2 should exist in issues")
	}
	if !containsSubstring(issuesStr, `"epic":"proj-1"`) {
		t.Error("PROJ-2 should be linked to epic proj-1")
	}
	if !containsSubstring(issuesStr, `"dependsOn":["proj-4"]`) {
		t.Error("PROJ-2 should depend on proj-4")
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return strings.Contains(s, substr)
}

// Helper function to split string into lines
func splitLines(s string) []string {
	return strings.Split(strings.TrimSpace(s), "\n")
}
