package converter

import (
	"testing"

	beadspb "github.com/conallob/jira-to-beads/gen/beads"
	jirapb "github.com/conallob/jira-to-beads/gen/jira"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNewProtoConverter(t *testing.T) {
	conv := NewProtoConverter()
	if conv == nil {
		t.Fatal("NewProtoConverter returned nil")
	}
	if conv.issueMap == nil {
		t.Error("issueMap is nil")
	}
	if conv.epicMap == nil {
		t.Error("epicMap is nil")
	}
}

func TestProtoMapStatus(t *testing.T) {
	conv := NewProtoConverter()

	tests := []struct {
		name       string
		jiraStatus *jirapb.Status
		wantStatus beadspb.Status
	}{
		{
			name: "new status",
			jiraStatus: &jirapb.Status{
				Name: "To Do",
				StatusCategory: &jirapb.StatusCategory{
					Key:  "new",
					Name: "To Do",
				},
			},
			wantStatus: beadspb.Status_STATUS_OPEN,
		},
		{
			name: "in progress status",
			jiraStatus: &jirapb.Status{
				Name: "In Progress",
				StatusCategory: &jirapb.StatusCategory{
					Key:  "indeterminate",
					Name: "In Progress",
				},
			},
			wantStatus: beadspb.Status_STATUS_IN_PROGRESS,
		},
		{
			name: "done status",
			jiraStatus: &jirapb.Status{
				Name: "Done",
				StatusCategory: &jirapb.StatusCategory{
					Key:  "done",
					Name: "Done",
				},
			},
			wantStatus: beadspb.Status_STATUS_CLOSED,
		},
		{
			name: "blocked status by name",
			jiraStatus: &jirapb.Status{
				Name: "Blocked",
				StatusCategory: &jirapb.StatusCategory{
					Key:  "other",
					Name: "Other",
				},
			},
			wantStatus: beadspb.Status_STATUS_BLOCKED,
		},
		{
			name:       "nil status",
			jiraStatus: nil,
			wantStatus: beadspb.Status_STATUS_OPEN,
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

func TestProtoMapPriority(t *testing.T) {
	conv := NewProtoConverter()

	tests := []struct {
		name         string
		jiraPriority *jirapb.Priority
		wantPriority beadspb.Priority
	}{
		{
			name:         "critical priority",
			jiraPriority: &jirapb.Priority{Name: "Critical", Id: "1"},
			wantPriority: beadspb.Priority_PRIORITY_P0,
		},
		{
			name:         "highest priority",
			jiraPriority: &jirapb.Priority{Name: "Highest", Id: "1"},
			wantPriority: beadspb.Priority_PRIORITY_P0,
		},
		{
			name:         "high priority",
			jiraPriority: &jirapb.Priority{Name: "High", Id: "2"},
			wantPriority: beadspb.Priority_PRIORITY_P1,
		},
		{
			name:         "medium priority",
			jiraPriority: &jirapb.Priority{Name: "Medium", Id: "3"},
			wantPriority: beadspb.Priority_PRIORITY_P2,
		},
		{
			name:         "low priority",
			jiraPriority: &jirapb.Priority{Name: "Low", Id: "4"},
			wantPriority: beadspb.Priority_PRIORITY_P3,
		},
		{
			name:         "lowest priority",
			jiraPriority: &jirapb.Priority{Name: "Lowest", Id: "5"},
			wantPriority: beadspb.Priority_PRIORITY_P4,
		},
		{
			name:         "unknown priority defaults to medium",
			jiraPriority: &jirapb.Priority{Name: "Unknown", Id: "99"},
			wantPriority: beadspb.Priority_PRIORITY_P2,
		},
		{
			name:         "nil priority defaults to medium",
			jiraPriority: nil,
			wantPriority: beadspb.Priority_PRIORITY_P2,
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

func TestProtoGenerateBeadsID(t *testing.T) {
	conv := NewProtoConverter()

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

func TestProtoConvertEpic(t *testing.T) {
	conv := NewProtoConverter()

	now := timestamppb.Now()
	jiraIssue := &jirapb.Issue{
		Id:  "10001",
		Key: "PROJ-1",
		Fields: &jirapb.Fields{
			Summary:     "Test Epic",
			Description: "Epic description",
			IssueType: &jirapb.IssueType{
				Name:    "Epic",
				Subtask: false,
			},
			Status: &jirapb.Status{
				Name: "In Progress",
				StatusCategory: &jirapb.StatusCategory{
					Key:  "indeterminate",
					Name: "In Progress",
				},
			},
			Created: now,
			Updated: now,
		},
	}

	epic, err := conv.convertEpic(jiraIssue)
	if err != nil {
		t.Fatalf("convertEpic failed: %v", err)
	}

	if epic.Id != "proj-1" {
		t.Errorf("Expected ID proj-1, got %s", epic.Id)
	}
	if epic.Name != "Test Epic" {
		t.Errorf("Expected name 'Test Epic', got %s", epic.Name)
	}
	if epic.Status != beadspb.Status_STATUS_IN_PROGRESS {
		t.Errorf("Expected status STATUS_IN_PROGRESS, got %v", epic.Status)
	}
	if epic.Metadata.JiraKey != "PROJ-1" {
		t.Errorf("Expected JiraKey PROJ-1, got %s", epic.Metadata.JiraKey)
	}
}

func TestProtoConvertIssue(t *testing.T) {
	conv := NewProtoConverter()

	now := timestamppb.Now()
	jiraIssue := &jirapb.Issue{
		Id:  "10002",
		Key: "PROJ-2",
		Fields: &jirapb.Fields{
			Summary:     "Test Issue",
			Description: "Issue description",
			IssueType: &jirapb.IssueType{
				Name:    "Story",
				Subtask: false,
			},
			Status: &jirapb.Status{
				Name: "To Do",
				StatusCategory: &jirapb.StatusCategory{
					Key:  "new",
					Name: "To Do",
				},
			},
			Priority: &jirapb.Priority{
				Name: "High",
				Id:   "2",
			},
			Assignee: &jirapb.User{
				DisplayName:  "John Doe",
				EmailAddress: "john@example.com",
			},
			Labels:  []string{"test", "api"},
			Created: now,
			Updated: now,
		},
	}

	issue, err := conv.convertIssue(jiraIssue)
	if err != nil {
		t.Fatalf("convertIssue failed: %v", err)
	}

	if issue.Id != "proj-2" {
		t.Errorf("Expected ID proj-2, got %s", issue.Id)
	}
	if issue.Title != "Test Issue" {
		t.Errorf("Expected title 'Test Issue', got %s", issue.Title)
	}
	if issue.Status != beadspb.Status_STATUS_OPEN {
		t.Errorf("Expected status STATUS_OPEN, got %v", issue.Status)
	}
	if issue.Priority != beadspb.Priority_PRIORITY_P1 {
		t.Errorf("Expected priority PRIORITY_P1, got %v", issue.Priority)
	}
	if issue.Assignee != "john@example.com" {
		t.Errorf("Expected assignee john@example.com, got %s", issue.Assignee)
	}
	if len(issue.Labels) != 2 {
		t.Errorf("Expected 2 labels, got %d", len(issue.Labels))
	}
}

func TestProtoConvertNilExport(t *testing.T) {
	conv := NewProtoConverter()
	_, err := conv.Convert(nil)
	if err == nil {
		t.Error("Expected error for nil export, got nil")
	}
}

func TestProtoBuildIssueMap(t *testing.T) {
	conv := NewProtoConverter()

	export := &jirapb.Export{
		Issues: []*jirapb.Issue{
			{Key: "PROJ-1", Id: "1"},
			{Key: "PROJ-2", Id: "2"},
			{Key: "PROJ-3", Id: "3"},
		},
	}

	issueMap := conv.buildIssueMap(export)
	if len(issueMap) != 3 {
		t.Errorf("Expected 3 issues in map, got %d", len(issueMap))
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
}

func TestProtoGetEpics(t *testing.T) {
	conv := NewProtoConverter()

	export := &jirapb.Export{
		Issues: []*jirapb.Issue{
			{
				Key: "PROJ-1",
				Fields: &jirapb.Fields{
					IssueType: &jirapb.IssueType{Name: "Epic"},
				},
			},
			{
				Key: "PROJ-2",
				Fields: &jirapb.Fields{
					IssueType: &jirapb.IssueType{Name: "Story"},
				},
			},
		},
	}

	epics := conv.getEpics(export)
	if len(epics) != 1 {
		t.Errorf("Expected 1 epic, got %d", len(epics))
	}

	if len(epics) > 0 && epics[0].Key != "PROJ-1" {
		t.Errorf("Expected epic key PROJ-1, got %s", epics[0].Key)
	}
}

func TestProtoGetDependencies(t *testing.T) {
	conv := NewProtoConverter()

	export := &jirapb.Export{
		Issues: []*jirapb.Issue{
			{
				Key: "PROJ-1",
				Fields: &jirapb.Fields{
					IssueLinks: []*jirapb.IssueLink{
						{
							Type: &jirapb.IssueLinkType{
								Inward: "is blocked by",
							},
							InwardIssue: &jirapb.LinkedIssue{
								Key: "PROJ-2",
							},
						},
					},
				},
			},
		},
	}

	deps := conv.getDependencies(export)

	proj1Deps, exists := deps["PROJ-1"]
	if !exists {
		t.Error("Expected PROJ-1 to have dependencies")
	}
	if len(proj1Deps) != 1 || proj1Deps[0] != "PROJ-2" {
		t.Errorf("Expected PROJ-1 to depend on PROJ-2, got %v", proj1Deps)
	}
}
