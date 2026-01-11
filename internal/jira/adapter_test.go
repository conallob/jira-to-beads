package jira

import (
	"testing"

	pb "github.com/conallob/jira-beads-sync/gen/jira"
)

func TestAdapterParseFile(t *testing.T) {
	adapter := NewAdapter()
	export, err := adapter.ParseFile("../../testdata/sample-jira-export.json")
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

func TestAdapterParseInvalidFile(t *testing.T) {
	adapter := NewAdapter()
	_, err := adapter.ParseFile("nonexistent.json")
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
}

func TestAdapterParseInvalidJSON(t *testing.T) {
	adapter := NewAdapter()
	_, err := adapter.Parse([]byte("invalid json"))
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestAdapterParseEmptyIssues(t *testing.T) {
	adapter := NewAdapter()
	_, err := adapter.Parse([]byte(`{"issues": []}`))
	if err == nil {
		t.Error("Expected error for empty issues, got nil")
	}
}

func TestAdapterValidation(t *testing.T) {
	adapter := NewAdapter()

	tests := []struct {
		name    string
		json    string
		wantErr bool
	}{
		{
			name: "valid minimal issue",
			json: `{"issues": [{"id": "1", "key": "PROJ-1", "fields": {
				"summary": "Test",
				"issuetype": {"name": "Task"},
				"status": {"name": "Open", "statusCategory": {"key": "new", "name": "New"}},
				"priority": {"name": "Medium", "id": "3"},
				"labels": [],
				"issuelinks": [],
				"subtasks": []
			}}]}`,
			wantErr: false,
		},
		{
			name:    "missing key",
			json:    `{"issues": [{"id": "1", "fields": {"summary": "Test", "issuetype": {"name": "Task"}}}]}`,
			wantErr: true,
		},
		{
			name:    "missing summary",
			json:    `{"issues": [{"id": "1", "key": "PROJ-1", "fields": {"issuetype": {"name": "Task"}}}]}`,
			wantErr: true,
		},
		{
			name:    "missing issue type",
			json:    `{"issues": [{"id": "1", "key": "PROJ-1", "fields": {"summary": "Test"}}]}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := adapter.Parse([]byte(tt.json))
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAdapterConvertIssue(t *testing.T) {
	adapter := NewAdapter()
	export, err := adapter.ParseFile("../../testdata/sample-jira-export.json")
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	// Check first issue (epic)
	epic := export.Issues[0]
	if epic.Key != "PROJ-1" {
		t.Errorf("Expected key PROJ-1, got %s", epic.Key)
	}
	if epic.Fields.Summary != "Implement User Authentication" {
		t.Errorf("Expected summary 'Implement User Authentication', got %s", epic.Fields.Summary)
	}
	if epic.Fields.IssueType.Name != "Epic" {
		t.Errorf("Expected issue type Epic, got %s", epic.Fields.IssueType.Name)
	}

	// Check timestamps
	if epic.Fields.Created == nil {
		t.Error("Created timestamp is nil")
	}
	if epic.Fields.Updated == nil {
		t.Error("Updated timestamp is nil")
	}

	// Check assignee
	if epic.Fields.Assignee == nil {
		t.Error("Assignee is nil")
	} else {
		if epic.Fields.Assignee.DisplayName != "John Doe" {
			t.Errorf("Expected assignee 'John Doe', got %s", epic.Fields.Assignee.DisplayName)
		}
	}

	// Check labels
	if len(epic.Fields.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(epic.Fields.Labels))
	}

	// Check subtasks
	if len(epic.Fields.Subtasks) != 2 {
		t.Errorf("Expected 2 subtasks, got %d", len(epic.Fields.Subtasks))
	}
}

func TestAdapterConvertIssueLinks(t *testing.T) {
	adapter := NewAdapter()
	export, err := adapter.ParseFile("../../testdata/sample-jira-export.json")
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	// Find PROJ-2 which has an issue link
	var proj2 *pb.Issue
	for _, issue := range export.Issues {
		if issue.Key == "PROJ-2" {
			proj2 = issue
			break
		}
	}

	if proj2 == nil {
		t.Fatal("Could not find PROJ-2")
	}

	if len(proj2.Fields.IssueLinks) == 0 {
		t.Error("Expected PROJ-2 to have issue links")
	} else {
		link := proj2.Fields.IssueLinks[0]
		if link.Type.Inward != "is blocked by" {
			t.Errorf("Expected inward link 'is blocked by', got %s", link.Type.Inward)
		}
		if link.InwardIssue == nil {
			t.Error("InwardIssue is nil")
		} else if link.InwardIssue.Key != "PROJ-4" {
			t.Errorf("Expected inward issue PROJ-4, got %s", link.InwardIssue.Key)
		}
	}
}

func TestAdapterConvertParent(t *testing.T) {
	adapter := NewAdapter()
	export, err := adapter.ParseFile("../../testdata/sample-jira-export.json")
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	// Find PROJ-2 which has a parent
	var proj2 *pb.Issue
	for _, issue := range export.Issues {
		if issue.Key == "PROJ-2" {
			proj2 = issue
			break
		}
	}

	if proj2 == nil {
		t.Fatal("Could not find PROJ-2")
	}

	if proj2.Fields.Parent == nil {
		t.Error("Expected PROJ-2 to have a parent")
	} else {
		if proj2.Fields.Parent.Key != "PROJ-1" {
			t.Errorf("Expected parent PROJ-1, got %s", proj2.Fields.Parent.Key)
		}
		if proj2.Fields.Parent.Fields.IssueType.Name != "Epic" {
			t.Errorf("Expected parent to be Epic, got %s", proj2.Fields.Parent.Fields.IssueType.Name)
		}
	}
}
