package beads

import "time"

// Issue represents a beads issue stored as YAML in .beads/issues/
type Issue struct {
	ID          string    `yaml:"id"`
	Title       string    `yaml:"title"`
	Description string    `yaml:"description,omitempty"`
	Status      Status    `yaml:"status"`
	Priority    Priority  `yaml:"priority"`
	Epic        string    `yaml:"epic,omitempty"`
	Assignee    string    `yaml:"assignee,omitempty"`
	Labels      []string  `yaml:"labels,omitempty"`
	DependsOn   []string  `yaml:"dependsOn,omitempty"`
	Created     time.Time `yaml:"created"`
	Updated     time.Time `yaml:"updated"`
	Metadata    Metadata  `yaml:"metadata,omitempty"`
}

// Status represents the status of a beads issue
type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusBlocked    Status = "blocked"
	StatusClosed     Status = "closed"
)

// Priority represents the priority level of a beads issue
type Priority string

const (
	PriorityP0 Priority = "p0" // Critical
	PriorityP1 Priority = "p1" // High
	PriorityP2 Priority = "p2" // Medium
	PriorityP3 Priority = "p3" // Low
	PriorityP4 Priority = "p4" // Very Low
)

// Metadata stores additional information about the issue
type Metadata struct {
	JiraKey       string            `yaml:"jiraKey,omitempty"`
	JiraID        string            `yaml:"jiraId,omitempty"`
	JiraIssueType string            `yaml:"jiraIssueType,omitempty"`
	Custom        map[string]string `yaml:"custom,omitempty"`
}

// Epic represents a beads epic
type Epic struct {
	ID          string    `yaml:"id"`
	Name        string    `yaml:"name"`
	Description string    `yaml:"description,omitempty"`
	Status      Status    `yaml:"status"`
	Created     time.Time `yaml:"created"`
	Updated     time.Time `yaml:"updated"`
	Metadata    Metadata  `yaml:"metadata,omitempty"`
}

// Export represents a collection of beads issues and epics for export
type Export struct {
	Issues []Issue
	Epics  []Epic
}
