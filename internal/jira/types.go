package jira

import (
	"encoding/json"
	"time"
)

// JiraTime is a custom time type that handles Jira's timestamp format
type JiraTime struct {
	time.Time
}

// UnmarshalJSON implements custom JSON unmarshaling for Jira timestamps
func (jt *JiraTime) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	if s == "" {
		return nil
	}

	// Jira uses format like "2024-01-01T10:00:00.000+0000"
	t, err := time.Parse("2006-01-02T15:04:05.000-0700", s)
	if err != nil {
		return err
	}

	jt.Time = t
	return nil
}

// Export represents a Jira export file containing multiple issues
type Export struct {
	Issues []Issue `json:"issues"`
}

// Issue represents a Jira issue (story, epic, subtask, etc.)
type Issue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Self   string `json:"self"`
	Fields Fields `json:"fields"`
}

// Fields contains the detailed information about a Jira issue
type Fields struct {
	Summary     string      `json:"summary"`
	Description string      `json:"description"`
	IssueType   IssueType   `json:"issuetype"`
	Status      Status      `json:"status"`
	Priority    Priority    `json:"priority"`
	Assignee    *User       `json:"assignee,omitempty"`
	Reporter    *User       `json:"reporter,omitempty"`
	Created     JiraTime    `json:"created"`
	Updated     JiraTime    `json:"updated"`
	Labels      []string    `json:"labels"`
	IssueLinks  []IssueLink `json:"issuelinks"`
	Parent      *Parent     `json:"parent,omitempty"`
	Epic        *Epic       `json:"epic,omitempty"`
	Subtasks    []Subtask   `json:"subtasks"`
}

// IssueType represents the type of a Jira issue
type IssueType struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Subtask     bool   `json:"subtask"`
}

// Status represents the current status of a Jira issue
type Status struct {
	Name           string         `json:"name"`
	StatusCategory StatusCategory `json:"statusCategory"`
}

// StatusCategory represents the high-level category of a status
type StatusCategory struct {
	Key  string `json:"key"` // "new", "indeterminate", "done"
	Name string `json:"name"`
}

// Priority represents the priority of a Jira issue
type Priority struct {
	Name string `json:"name"`
	ID   string `json:"id"`
}

// User represents a Jira user
type User struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress,omitempty"`
}

// IssueLink represents a link between two Jira issues
type IssueLink struct {
	ID           string        `json:"id"`
	Type         IssueLinkType `json:"type"`
	InwardIssue  *LinkedIssue  `json:"inwardIssue,omitempty"`
	OutwardIssue *LinkedIssue  `json:"outwardIssue,omitempty"`
}

// IssueLinkType describes the type of relationship between issues
type IssueLinkType struct {
	Name    string `json:"name"`
	Inward  string `json:"inward"`  // e.g., "is blocked by"
	Outward string `json:"outward"` // e.g., "blocks"
}

// LinkedIssue represents a reference to another issue in a link
type LinkedIssue struct {
	ID     string       `json:"id"`
	Key    string       `json:"key"`
	Self   string       `json:"self"`
	Fields LinkedFields `json:"fields"`
}

// LinkedFields contains minimal fields for a linked issue
type LinkedFields struct {
	Summary   string    `json:"summary"`
	Status    Status    `json:"status"`
	IssueType IssueType `json:"issuetype"`
}

// Parent represents a parent issue reference
type Parent struct {
	ID     string       `json:"id"`
	Key    string       `json:"key"`
	Self   string       `json:"self"`
	Fields LinkedFields `json:"fields"`
}

// Epic represents an epic reference
type Epic struct {
	ID      string `json:"id"`
	Key     string `json:"key"`
	Self    string `json:"self"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Done    bool   `json:"done"`
}

// Subtask represents a subtask reference
type Subtask struct {
	ID     string       `json:"id"`
	Key    string       `json:"key"`
	Self   string       `json:"self"`
	Fields LinkedFields `json:"fields"`
}
