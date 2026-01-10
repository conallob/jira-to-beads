package jira

import (
	"testing"
)

func TestParseFile(t *testing.T) {
	parser := NewParser()
	export, err := parser.ParseFile("../../testdata/sample-jira-export.json")
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	if export == nil {
		t.Fatal("Export is nil")
	}

	if len(export.Issues) != 4 {
		t.Errorf("Expected 4 issues, got %d", len(export.Issues))
	}
}

func TestParseInvalidFile(t *testing.T) {
	parser := NewParser()
	_, err := parser.ParseFile("nonexistent.json")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestParseInvalidJSON(t *testing.T) {
	parser := NewParser()
	_, err := parser.Parse([]byte("invalid json"))
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestParseEmptyIssues(t *testing.T) {
	parser := NewParser()
	_, err := parser.Parse([]byte(`{"issues": []}`))
	if err == nil {
		t.Error("Expected error for empty issues, got nil")
	}
}

func TestBuildIssueMap(t *testing.T) {
	parser := NewParser()
	export, err := parser.ParseFile("../../testdata/sample-jira-export.json")
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	issueMap := parser.BuildIssueMap(export)
	if len(issueMap) != 4 {
		t.Errorf("Expected 4 issues in map, got %d", len(issueMap))
	}

	if _, exists := issueMap["PROJ-1"]; !exists {
		t.Error("Expected PROJ-1 to exist in map")
	}
	if _, exists := issueMap["PROJ-2"]; !exists {
		t.Error("Expected PROJ-2 to exist in map")
	}
	if _, exists := issueMap["PROJ-3"]; !exists {
		t.Error("Expected PROJ-3 to exist in map")
	}
	if _, exists := issueMap["PROJ-4"]; !exists {
		t.Error("Expected PROJ-4 to exist in map")
	}
}

func TestGetEpics(t *testing.T) {
	parser := NewParser()
	export, err := parser.ParseFile("../../testdata/sample-jira-export.json")
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	epics := parser.GetEpics(export)
	if len(epics) != 1 {
		t.Errorf("Expected 1 epic, got %d", len(epics))
	}

	if len(epics) > 0 && epics[0].Key != "PROJ-1" {
		t.Errorf("Expected epic key PROJ-1, got %s", epics[0].Key)
	}
}

func TestGetStories(t *testing.T) {
	parser := NewParser()
	export, err := parser.ParseFile("../../testdata/sample-jira-export.json")
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	stories := parser.GetStories(export)
	if len(stories) != 1 {
		t.Errorf("Expected 1 story/task, got %d", len(stories))
	}

	if len(stories) > 0 && stories[0].Key != "PROJ-4" {
		t.Errorf("Expected story key PROJ-4, got %s", stories[0].Key)
	}
}

func TestGetSubtasks(t *testing.T) {
	parser := NewParser()
	export, err := parser.ParseFile("../../testdata/sample-jira-export.json")
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	subtasks := parser.GetSubtasks(export)
	if len(subtasks) != 2 {
		t.Errorf("Expected 2 subtasks, got %d", len(subtasks))
	}
}

func TestGetDependencies(t *testing.T) {
	parser := NewParser()
	export, err := parser.ParseFile("../../testdata/sample-jira-export.json")
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	deps := parser.GetDependencies(export)

	// PROJ-2 is blocked by PROJ-4
	proj2Deps, exists := deps["PROJ-2"]
	if !exists {
		t.Error("Expected PROJ-2 to have dependencies")
	}
	if len(proj2Deps) != 1 || proj2Deps[0] != "PROJ-4" {
		t.Errorf("Expected PROJ-2 to depend on PROJ-4, got %v", proj2Deps)
	}
}

func TestValidation(t *testing.T) {
	parser := NewParser()

	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name:    "valid minimal issue",
			json:    `{"issues": [{"key": "PROJ-1", "fields": {"summary": "Test", "issuetype": {"name": "Task"}}}]}`,
			wantErr: false,
		},
		{
			name:    "missing key",
			json:    `{"issues": [{"fields": {"summary": "Test", "issuetype": {"name": "Task"}}}]}`,
			wantErr: true,
		},
		{
			name:    "missing summary",
			json:    `{"issues": [{"key": "PROJ-1", "fields": {"issuetype": {"name": "Task"}}}]}`,
			wantErr: true,
		},
		{
			name:    "missing issue type",
			json:    `{"issues": [{"key": "PROJ-1", "fields": {"summary": "Test"}}]}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parser.Parse([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
