package converter

import (
	"testing"
	"time"

	"github.com/conallob/jira-to-beads/internal/beads"
	"github.com/conallob/jira-to-beads/internal/jira"
)

func TestNewConverter(t *testing.T) {
	conv := NewConverter()
	if conv == nil {
		t.Fatal("NewConverter returned nil")
	}
	if conv.jiraParser == nil {
		t.Error("jiraParser is nil")
	}
}

func TestMapStatus(t *testing.T) {
	conv := NewConverter()

	tests := []struct {
		name       string
		jiraStatus jira.Status
		wantStatus beads.Status
	}{
		{
			name: "new status",
			jiraStatus: jira.Status{
				Name: "To Do",
				StatusCategory: jira.StatusCategory{
					Key:  "new",
					Name: "To Do",
				},
			},
			wantStatus: beads.StatusOpen,
		},
		{
			name: "in progress status",
			jiraStatus: jira.Status{
				Name: "In Progress",
				StatusCategory: jira.StatusCategory{
					Key:  "indeterminate",
					Name: "In Progress",
				},
			},
			wantStatus: beads.StatusInProgress,
		},
		{
			name: "done status",
			jiraStatus: jira.Status{
				Name: "Done",
				StatusCategory: jira.StatusCategory{
					Key:  "done",
					Name: "Done",
				},
			},
			wantStatus: beads.StatusClosed,
		},
		{
			name: "blocked status by name",
			jiraStatus: jira.Status{
				Name: "Blocked",
				StatusCategory: jira.StatusCategory{
					Key:  "other",
					Name: "Other",
				},
			},
			wantStatus: beads.StatusBlocked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := conv.mapStatus(tt.jiraStatus)
			if got != tt.wantStatus {
				t.Errorf("mapStatus() = %v, want %v", got, tt.wantStatus)
			}
		})
	}
}

func TestMapPriority(t *testing.T) {
	conv := NewConverter()

	tests := []struct {
		name         string
		jiraPriority jira.Priority
		wantPriority beads.Priority
	}{
		{
			name:         "critical priority",
			jiraPriority: jira.Priority{Name: "Critical", ID: "1"},
			wantPriority: beads.PriorityP0,
		},
		{
			name:         "highest priority",
			jiraPriority: jira.Priority{Name: "Highest", ID: "1"},
			wantPriority: beads.PriorityP0,
		},
		{
			name:         "high priority",
			jiraPriority: jira.Priority{Name: "High", ID: "2"},
			wantPriority: beads.PriorityP1,
		},
		{
			name:         "medium priority",
			jiraPriority: jira.Priority{Name: "Medium", ID: "3"},
			wantPriority: beads.PriorityP2,
		},
		{
			name:         "low priority",
			jiraPriority: jira.Priority{Name: "Low", ID: "4"},
			wantPriority: beads.PriorityP3,
		},
		{
			name:         "lowest priority",
			jiraPriority: jira.Priority{Name: "Lowest", ID: "5"},
			wantPriority: beads.PriorityP4,
		},
		{
			name:         "unknown priority defaults to medium",
			jiraPriority: jira.Priority{Name: "Unknown", ID: "99"},
			wantPriority: beads.PriorityP2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := conv.mapPriority(tt.jiraPriority)
			if got != tt.wantPriority {
				t.Errorf("mapPriority() = %v, want %v", got, tt.wantPriority)
			}
		})
	}
}

func TestGenerateBeadsID(t *testing.T) {
	conv := NewConverter()

	tests := []struct {
		jiraKey string
		wantID  string
	}{
		{"PROJ-1", "proj-1"},
		{"MYPROJECT-123", "myproject-123"},
		{"ABC-456", "abc-456"},
	}

	for _, tt := range tests {
		t.Run(tt.jiraKey, func(t *testing.T) {
			got := conv.generateBeadsID(tt.jiraKey)
			if got != tt.wantID {
				t.Errorf("generateBeadsID() = %v, want %v", got, tt.wantID)
			}
		})
	}
}

func TestConvertEpic(t *testing.T) {
	conv := NewConverter()

	jiraIssue := &jira.Issue{
		ID:  "10001",
		Key: "PROJ-1",
		Fields: jira.Fields{
			Summary:     "Test Epic",
			Description: "Epic description",
			IssueType: jira.IssueType{
				Name:    "Epic",
				Subtask: false,
			},
			Status: jira.Status{
				Name: "In Progress",
				StatusCategory: jira.StatusCategory{
					Key:  "indeterminate",
					Name: "In Progress",
				},
			},
			Created: jira.JiraTime{Time: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)},
			Updated: jira.JiraTime{Time: time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)},
		},
	}

	epic, err := conv.convertEpic(jiraIssue)
	if err != nil {
		t.Fatalf("convertEpic failed: %v", err)
	}

	if epic.ID != "proj-1" {
		t.Errorf("Expected ID proj-1, got %s", epic.ID)
	}
	if epic.Name != "Test Epic" {
		t.Errorf("Expected name 'Test Epic', got %s", epic.Name)
	}
	if epic.Status != beads.StatusInProgress {
		t.Errorf("Expected status in_progress, got %s", epic.Status)
	}
	if epic.Metadata.JiraKey != "PROJ-1" {
		t.Errorf("Expected JiraKey PROJ-1, got %s", epic.Metadata.JiraKey)
	}
}

func TestConvertIssue(t *testing.T) {
	conv := NewConverter()

	jiraIssue := &jira.Issue{
		ID:  "10002",
		Key: "PROJ-2",
		Fields: jira.Fields{
			Summary:     "Test Issue",
			Description: "Issue description",
			IssueType: jira.IssueType{
				Name:    "Story",
				Subtask: false,
			},
			Status: jira.Status{
				Name: "To Do",
				StatusCategory: jira.StatusCategory{
					Key:  "new",
					Name: "To Do",
				},
			},
			Priority: jira.Priority{
				Name: "High",
				ID:   "2",
			},
			Assignee: &jira.User{
				DisplayName:  "John Doe",
				EmailAddress: "john@example.com",
			},
			Labels:  []string{"test", "api"},
			Created: jira.JiraTime{Time: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC)},
			Updated: jira.JiraTime{Time: time.Date(2024, 1, 2, 10, 0, 0, 0, time.UTC)},
		},
	}

	issue, err := conv.convertIssue(jiraIssue)
	if err != nil {
		t.Fatalf("convertIssue failed: %v", err)
	}

	if issue.ID != "proj-2" {
		t.Errorf("Expected ID proj-2, got %s", issue.ID)
	}
	if issue.Title != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got %s", issue.Title)
	}
	if issue.Status != beads.StatusOpen {
		t.Errorf("Expected status open, got %s", issue.Status)
	}
	if issue.Priority != beads.PriorityP1 {
		t.Errorf("Expected priority p1, got %s", issue.Priority)
	}
	if issue.Assignee != "john@example.com" {
		t.Errorf("Expected assignee john@example.com, got %s", issue.Assignee)
	}
	if len(issue.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(issue.Labels))
	}
}

func TestConvert(t *testing.T) {
	conv := NewConverter()

	// Use the parser to read the sample Jira export
	parser := jira.NewParser()
	jiraExport, err := parser.ParseFile("../../testdata/sample-jira-export.json")
	if err != nil {
		t.Fatalf("Failed to parse sample file: %v", err)
	}

	beadsExport, err := conv.Convert(jiraExport)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Verify epics were converted
	if len(beadsExport.Epics) != 1 {
		t.Errorf("Expected 1 epic, got %d", len(beadsExport.Epics))
	}

	if len(beadsExport.Epics) > 0 {
		epic := beadsExport.Epics[0]
		if epic.ID != "proj-1" {
			t.Errorf("Expected epic ID proj-1, got %s", epic.ID)
		}
		if epic.Name != "Implement User Authentication" {
			t.Errorf("Expected epic name 'Implement User Authentication', got %s", epic.Name)
		}
	}

	// Verify issues were converted (should be 3: 2 subtasks + 1 task, epic excluded)
	if len(beadsExport.Issues) != 3 {
		t.Errorf("Expected 3 issues, got %d", len(beadsExport.Issues))
	}

	// Check that subtasks have the epic linked
	subtaskCount := 0
	for _, issue := range beadsExport.Issues {
		if issue.Metadata.JiraIssueType == "Subtask" {
			subtaskCount++
			if issue.Epic != "proj-1" {
				t.Errorf("Expected subtask %s to have epic proj-1, got %s", issue.ID, issue.Epic)
			}
		}
	}

	if subtaskCount != 2 {
		t.Errorf("Expected 2 subtasks, got %d", subtaskCount)
	}

	// Verify dependencies were added
	// PROJ-2 should depend on PROJ-4
	var proj2Issue *beads.Issue
	for i, issue := range beadsExport.Issues {
		if issue.Metadata.JiraKey == "PROJ-2" {
			proj2Issue = &beadsExport.Issues[i]
			break
		}
	}

	if proj2Issue == nil {
		t.Fatal("Could not find PROJ-2 in converted issues")
	}

	if len(proj2Issue.DependsOn) == 0 {
		t.Error("Expected PROJ-2 to have dependencies")
	} else {
		foundProj4Dep := false
		for _, dep := range proj2Issue.DependsOn {
			if dep == "proj-4" {
				foundProj4Dep = true
				break
			}
		}
		if !foundProj4Dep {
			t.Errorf("Expected PROJ-2 to depend on proj-4, got dependencies: %v", proj2Issue.DependsOn)
		}
	}
}

func TestConvertNilExport(t *testing.T) {
	conv := NewConverter()
	_, err := conv.Convert(nil)
	if err == nil {
		t.Error("Expected error for nil export, got nil")
	}
}
