package jira

import (
	"encoding/json"
	"fmt"
	"os"
)

// Parser handles parsing Jira export files
type Parser struct{}

// NewParser creates a new Jira parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile reads and parses a Jira export JSON file
func (p *Parser) ParseFile(filename string) (*Export, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	return p.Parse(data)
}

// Parse parses Jira export JSON data
func (p *Parser) Parse(data []byte) (*Export, error) {
	var export Export
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, fmt.Errorf("failed to parse Jira export: %w", err)
	}

	if err := p.validate(&export); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return &export, nil
}

// validate checks if the parsed export is valid
func (p *Parser) validate(export *Export) error {
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
		if issue.Fields.Summary == "" {
			return fmt.Errorf("issue %s has no summary", issue.Key)
		}
		if issue.Fields.IssueType.Name == "" {
			return fmt.Errorf("issue %s has no issue type", issue.Key)
		}
	}

	return nil
}

// BuildIssueMap creates a map of issue keys to issues for quick lookup
func (p *Parser) BuildIssueMap(export *Export) map[string]*Issue {
	issueMap := make(map[string]*Issue)
	for i := range export.Issues {
		issueMap[export.Issues[i].Key] = &export.Issues[i]
	}
	return issueMap
}

// GetEpics returns all issues that are epics
func (p *Parser) GetEpics(export *Export) []*Issue {
	var epics []*Issue
	for i := range export.Issues {
		if export.Issues[i].Fields.IssueType.Name == "Epic" {
			epics = append(epics, &export.Issues[i])
		}
	}
	return epics
}

// GetStories returns all issues that are stories
func (p *Parser) GetStories(export *Export) []*Issue {
	var stories []*Issue
	for i := range export.Issues {
		issueType := export.Issues[i].Fields.IssueType.Name
		if issueType == "Story" || issueType == "Task" {
			stories = append(stories, &export.Issues[i])
		}
	}
	return stories
}

// GetSubtasks returns all issues that are subtasks
func (p *Parser) GetSubtasks(export *Export) []*Issue {
	var subtasks []*Issue
	for i := range export.Issues {
		if export.Issues[i].Fields.IssueType.Subtask {
			subtasks = append(subtasks, &export.Issues[i])
		}
	}
	return subtasks
}

// GetDependencies extracts dependency relationships from issue links
// Returns a map where the key is the issue that depends on other issues,
// and the value is a list of issue keys it depends on
func (p *Parser) GetDependencies(export *Export) map[string][]string {
	dependencies := make(map[string][]string)

	for _, issue := range export.Issues {
		var deps []string
		for _, link := range issue.Fields.IssueLinks {
			// Check if this issue is blocked by another
			if link.Type.Inward == "is blocked by" && link.InwardIssue != nil {
				deps = append(deps, link.InwardIssue.Key)
			}
			// Check if this issue depends on another
			if link.Type.Outward == "depends on" && link.OutwardIssue != nil {
				deps = append(deps, link.OutwardIssue.Key)
			}
		}
		if len(deps) > 0 {
			dependencies[issue.Key] = deps
		}
	}

	return dependencies
}
