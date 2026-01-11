package jira

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	pb "github.com/conallob/jira-beads-sync/gen/jira"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Adapter handles converting JSON Jira exports to protobuf format
type Adapter struct{}

// NewAdapter creates a new Jira JSON to protobuf adapter
func NewAdapter() *Adapter {
	return &Adapter{}
}

// ParseFile reads and parses a Jira export JSON file into protobuf
func (a *Adapter) ParseFile(filename string) (*pb.Export, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	return a.Parse(data)
}

// Parse parses Jira export JSON data into protobuf
func (a *Adapter) Parse(data []byte) (*pb.Export, error) {
	var jsonExport jsonExport
	if err := json.Unmarshal(data, &jsonExport); err != nil {
		return nil, fmt.Errorf("failed to parse Jira export: %w", err)
	}

	export := &pb.Export{
		Issues: make([]*pb.Issue, len(jsonExport.Issues)),
	}

	for i, jsonIssue := range jsonExport.Issues {
		issue, err := a.convertIssue(&jsonIssue)
		if err != nil {
			return nil, fmt.Errorf("failed to convert issue %s: %w", jsonIssue.Key, err)
		}
		export.Issues[i] = issue
	}

	if err := a.validate(export); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return export, nil
}

// validate checks if the parsed export is valid
func (a *Adapter) validate(export *pb.Export) error {
	if export == nil {
		return fmt.Errorf("export is nil")
	}

	if len(export.Issues) == 0 {
		return fmt.Errorf("export contains no issues")
	}

	for i, issue := range export.Issues {
		if issue.Key == "" {
			return fmt.Errorf("issue at index %d has no key", i)
		}
		if issue.Fields == nil || issue.Fields.Summary == "" {
			return fmt.Errorf("issue %s has no summary", issue.Key)
		}
		if issue.Fields.IssueType == nil || issue.Fields.IssueType.Name == "" {
			return fmt.Errorf("issue %s has no issue type", issue.Key)
		}
	}

	return nil
}

// convertIssue converts a JSON issue to protobuf
func (a *Adapter) convertIssue(jsonIssue *jsonIssue) (*pb.Issue, error) {
	issue := &pb.Issue{
		Id:   jsonIssue.ID,
		Key:  jsonIssue.Key,
		Self: jsonIssue.Self,
		Fields: &pb.Fields{
			Summary:     jsonIssue.Fields.Summary,
			Description: jsonIssue.Fields.Description,
			IssueType: &pb.IssueType{
				Name:        jsonIssue.Fields.IssueType.Name,
				Description: jsonIssue.Fields.IssueType.Description,
				Subtask:     jsonIssue.Fields.IssueType.Subtask,
			},
			Status: &pb.Status{
				Name: jsonIssue.Fields.Status.Name,
				StatusCategory: &pb.StatusCategory{
					Key:  jsonIssue.Fields.Status.StatusCategory.Key,
					Name: jsonIssue.Fields.Status.StatusCategory.Name,
				},
			},
			Priority: &pb.Priority{
				Name: jsonIssue.Fields.Priority.Name,
				Id:   jsonIssue.Fields.Priority.ID,
			},
			Labels:     jsonIssue.Fields.Labels,
			IssueLinks: make([]*pb.IssueLink, len(jsonIssue.Fields.IssueLinks)),
			Subtasks:   make([]*pb.Subtask, len(jsonIssue.Fields.Subtasks)),
		},
	}

	// Convert timestamps
	if !jsonIssue.Fields.Created.IsZero() {
		issue.Fields.Created = timestamppb.New(jsonIssue.Fields.Created)
	}
	if !jsonIssue.Fields.Updated.IsZero() {
		issue.Fields.Updated = timestamppb.New(jsonIssue.Fields.Updated)
	}

	// Convert assignee
	if jsonIssue.Fields.Assignee != nil {
		issue.Fields.Assignee = &pb.User{
			AccountId:    jsonIssue.Fields.Assignee.AccountID,
			DisplayName:  jsonIssue.Fields.Assignee.DisplayName,
			EmailAddress: jsonIssue.Fields.Assignee.EmailAddress,
		}
	}

	// Convert reporter
	if jsonIssue.Fields.Reporter != nil {
		issue.Fields.Reporter = &pb.User{
			AccountId:    jsonIssue.Fields.Reporter.AccountID,
			DisplayName:  jsonIssue.Fields.Reporter.DisplayName,
			EmailAddress: jsonIssue.Fields.Reporter.EmailAddress,
		}
	}

	// Convert issue links
	for i, link := range jsonIssue.Fields.IssueLinks {
		issue.Fields.IssueLinks[i] = a.convertIssueLink(&link)
	}

	// Convert parent
	if jsonIssue.Fields.Parent != nil {
		issue.Fields.Parent = a.convertParent(jsonIssue.Fields.Parent)
	}

	// Convert epic
	if jsonIssue.Fields.Epic != nil {
		issue.Fields.Epic = &pb.Epic{
			Id:      jsonIssue.Fields.Epic.ID,
			Key:     jsonIssue.Fields.Epic.Key,
			Self:    jsonIssue.Fields.Epic.Self,
			Name:    jsonIssue.Fields.Epic.Name,
			Summary: jsonIssue.Fields.Epic.Summary,
			Done:    jsonIssue.Fields.Epic.Done,
		}
	}

	// Convert subtasks
	for i, subtask := range jsonIssue.Fields.Subtasks {
		issue.Fields.Subtasks[i] = &pb.Subtask{
			Id:   subtask.ID,
			Key:  subtask.Key,
			Self: subtask.Self,
			Fields: &pb.LinkedFields{
				Summary: subtask.Fields.Summary,
				Status: &pb.Status{
					Name: subtask.Fields.Status.Name,
					StatusCategory: &pb.StatusCategory{
						Key:  subtask.Fields.Status.StatusCategory.Key,
						Name: subtask.Fields.Status.StatusCategory.Name,
					},
				},
				IssueType: &pb.IssueType{
					Name:        subtask.Fields.IssueType.Name,
					Description: subtask.Fields.IssueType.Description,
					Subtask:     subtask.Fields.IssueType.Subtask,
				},
			},
		}
	}

	return issue, nil
}

// convertIssueLink converts a JSON issue link to protobuf
func (a *Adapter) convertIssueLink(link *jsonIssueLink) *pb.IssueLink {
	pbLink := &pb.IssueLink{
		Id: link.ID,
		Type: &pb.IssueLinkType{
			Name:    link.Type.Name,
			Inward:  link.Type.Inward,
			Outward: link.Type.Outward,
		},
	}

	if link.InwardIssue != nil {
		pbLink.InwardIssue = &pb.LinkedIssue{
			Id:   link.InwardIssue.ID,
			Key:  link.InwardIssue.Key,
			Self: link.InwardIssue.Self,
			Fields: &pb.LinkedFields{
				Summary: link.InwardIssue.Fields.Summary,
				Status: &pb.Status{
					Name: link.InwardIssue.Fields.Status.Name,
					StatusCategory: &pb.StatusCategory{
						Key:  link.InwardIssue.Fields.Status.StatusCategory.Key,
						Name: link.InwardIssue.Fields.Status.StatusCategory.Name,
					},
				},
				IssueType: &pb.IssueType{
					Name:        link.InwardIssue.Fields.IssueType.Name,
					Description: link.InwardIssue.Fields.IssueType.Description,
					Subtask:     link.InwardIssue.Fields.IssueType.Subtask,
				},
			},
		}
	}

	if link.OutwardIssue != nil {
		pbLink.OutwardIssue = &pb.LinkedIssue{
			Id:   link.OutwardIssue.ID,
			Key:  link.OutwardIssue.Key,
			Self: link.OutwardIssue.Self,
			Fields: &pb.LinkedFields{
				Summary: link.OutwardIssue.Fields.Summary,
				Status: &pb.Status{
					Name: link.OutwardIssue.Fields.Status.Name,
					StatusCategory: &pb.StatusCategory{
						Key:  link.OutwardIssue.Fields.Status.StatusCategory.Key,
						Name: link.OutwardIssue.Fields.Status.StatusCategory.Name,
					},
				},
				IssueType: &pb.IssueType{
					Name:        link.OutwardIssue.Fields.IssueType.Name,
					Description: link.OutwardIssue.Fields.IssueType.Description,
					Subtask:     link.OutwardIssue.Fields.IssueType.Subtask,
				},
			},
		}
	}

	return pbLink
}

// convertParent converts a JSON parent to protobuf
func (a *Adapter) convertParent(parent *jsonParent) *pb.Parent {
	return &pb.Parent{
		Id:   parent.ID,
		Key:  parent.Key,
		Self: parent.Self,
		Fields: &pb.LinkedFields{
			Summary: parent.Fields.Summary,
			Status: &pb.Status{
				Name: parent.Fields.Status.Name,
				StatusCategory: &pb.StatusCategory{
					Key:  parent.Fields.Status.StatusCategory.Key,
					Name: parent.Fields.Status.StatusCategory.Name,
				},
			},
			IssueType: &pb.IssueType{
				Name:        parent.Fields.IssueType.Name,
				Description: parent.Fields.IssueType.Description,
				Subtask:     parent.Fields.IssueType.Subtask,
			},
		},
	}
}

// JSON types for unmarshaling (kept internal)
type jsonExport struct {
	Issues []jsonIssue `json:"issues"`
}

type jsonIssue struct {
	ID     string     `json:"id"`
	Key    string     `json:"key"`
	Self   string     `json:"self"`
	Fields jsonFields `json:"fields"`
}

type jsonFields struct {
	Summary     string          `json:"summary"`
	Description string          `json:"description"`
	IssueType   jsonIssueType   `json:"issuetype"`
	Status      jsonStatus      `json:"status"`
	Priority    jsonPriority    `json:"priority"`
	Assignee    *jsonUser       `json:"assignee,omitempty"`
	Reporter    *jsonUser       `json:"reporter,omitempty"`
	Created     time.Time       `json:"created"`
	Updated     time.Time       `json:"updated"`
	Labels      []string        `json:"labels"`
	IssueLinks  []jsonIssueLink `json:"issuelinks"`
	Parent      *jsonParent     `json:"parent,omitempty"`
	Epic        *jsonEpic       `json:"epic,omitempty"`
	Subtasks    []jsonSubtask   `json:"subtasks"`
}

type jsonIssueType struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Subtask     bool   `json:"subtask"`
}

type jsonStatus struct {
	Name           string             `json:"name"`
	StatusCategory jsonStatusCategory `json:"statusCategory"`
}

type jsonStatusCategory struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

type jsonPriority struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

type jsonUser struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress,omitempty"`
}

type jsonIssueLink struct {
	ID           string            `json:"id"`
	Type         jsonIssueLinkType `json:"type"`
	InwardIssue  *jsonLinkedIssue  `json:"inwardIssue,omitempty"`
	OutwardIssue *jsonLinkedIssue  `json:"outwardIssue,omitempty"`
}

type jsonIssueLinkType struct {
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
}

type jsonLinkedIssue struct {
	ID     string           `json:"id"`
	Key    string           `json:"key"`
	Self   string           `json:"self"`
	Fields jsonLinkedFields `json:"fields"`
}

type jsonLinkedFields struct {
	Summary   string        `json:"summary"`
	Status    jsonStatus    `json:"status"`
	IssueType jsonIssueType `json:"issuetype"`
}

type jsonParent struct {
	ID     string           `json:"id"`
	Key    string           `json:"key"`
	Self   string           `json:"self"`
	Fields jsonLinkedFields `json:"fields"`
}

type jsonEpic struct {
	ID      string `json:"id"`
	Key     string `json:"key"`
	Self    string `json:"self"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Done    bool   `json:"done"`
}

type jsonSubtask struct {
	ID     string           `json:"id"`
	Key    string           `json:"key"`
	Self   string           `json:"self"`
	Fields jsonLinkedFields `json:"fields"`
}

// UnmarshalJSON implements custom JSON unmarshaling for timestamps
func (jf *jsonFields) UnmarshalJSON(b []byte) error {
	type Alias jsonFields
	aux := &struct {
		Created string `json:"created"`
		Updated string `json:"updated"`
		*Alias
	}{
		Alias: (*Alias)(jf),
	}

	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}

	// Parse Jira timestamp format
	if aux.Created != "" {
		t, err := time.Parse("2006-01-02T15:04:05.000-0700", aux.Created)
		if err != nil {
			return err
		}
		jf.Created = t
	}

	if aux.Updated != "" {
		t, err := time.Parse("2006-01-02T15:04:05.000-0700", aux.Updated)
		if err != nil {
			return err
		}
		jf.Updated = t
	}

	return nil
}
