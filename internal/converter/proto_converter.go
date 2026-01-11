package converter

import (
	"fmt"
	"strings"

	beadspb "github.com/conallob/jira-beads-sync/gen/beads"
	jirapb "github.com/conallob/jira-beads-sync/gen/jira"
)

// ProtoConverter handles converting Jira protobuf to beads protobuf
type ProtoConverter struct {
	issueMap map[string]*jirapb.Issue // Map of Jira keys to issues
	epicMap  map[string]string        // Map of Jira epic keys to beads epic IDs
}

// NewProtoConverter creates a new protobuf-based converter
func NewProtoConverter() *ProtoConverter {
	return &ProtoConverter{
		issueMap: make(map[string]*jirapb.Issue),
		epicMap:  make(map[string]string),
	}
}

// Convert converts a Jira export to beads format
func (c *ProtoConverter) Convert(jiraExport *jirapb.Export) (*beadspb.Export, error) {
	if jiraExport == nil {
		return nil, fmt.Errorf("jira export is nil")
	}

	// Build issue map for quick lookups
	c.issueMap = c.buildIssueMap(jiraExport)

	beadsExport := &beadspb.Export{
		Issues: []*beadspb.Issue{},
		Epics:  []*beadspb.Epic{},
	}

	// Convert epics first so we can reference them in issues
	epics := c.getEpics(jiraExport)
	for _, jiraIssue := range epics {
		beadsEpic, err := c.convertEpic(jiraIssue)
		if err != nil {
			return nil, fmt.Errorf("failed to convert epic %s: %w", jiraIssue.Key, err)
		}
		beadsExport.Epics = append(beadsExport.Epics, beadsEpic)
		c.epicMap[jiraIssue.Key] = beadsEpic.Id
	}

	// Convert all issues (stories, tasks, subtasks)
	for _, jiraIssue := range jiraExport.Issues {
		// Skip epics as they've already been converted
		if jiraIssue.Fields.IssueType.Name == "Epic" {
			continue
		}

		beadsIssue, err := c.convertIssue(jiraIssue)
		if err != nil {
			return nil, fmt.Errorf("failed to convert issue %s: %w", jiraIssue.Key, err)
		}
		beadsExport.Issues = append(beadsExport.Issues, beadsIssue)
	}

	// Add dependencies after all issues are converted
	if err := c.addDependencies(jiraExport, beadsExport); err != nil {
		return nil, fmt.Errorf("failed to add dependencies: %w", err)
	}

	return beadsExport, nil
}

// convertEpic converts a Jira epic to a beads epic
func (c *ProtoConverter) convertEpic(jiraIssue *jirapb.Issue) (*beadspb.Epic, error) {
	epic := &beadspb.Epic{
		Id:          c.generateBeadsID(jiraIssue.Key),
		Name:        jiraIssue.Fields.Summary,
		Description: jiraIssue.Fields.Description,
		Status:      c.mapStatus(jiraIssue.Fields.Status),
		Created:     jiraIssue.Fields.Created,
		Updated:     jiraIssue.Fields.Updated,
		Metadata: &beadspb.Metadata{
			JiraKey:       jiraIssue.Key,
			JiraId:        jiraIssue.Id,
			JiraIssueType: jiraIssue.Fields.IssueType.Name,
		},
	}

	return epic, nil
}

// convertIssue converts a Jira issue to a beads issue
func (c *ProtoConverter) convertIssue(jiraIssue *jirapb.Issue) (*beadspb.Issue, error) {
	issue := &beadspb.Issue{
		Id:          c.generateBeadsID(jiraIssue.Key),
		Title:       jiraIssue.Fields.Summary,
		Description: jiraIssue.Fields.Description,
		Status:      c.mapStatus(jiraIssue.Fields.Status),
		Priority:    c.mapPriority(jiraIssue.Fields.Priority),
		Labels:      jiraIssue.Fields.Labels,
		DependsOn:   []string{},
		Created:     jiraIssue.Fields.Created,
		Updated:     jiraIssue.Fields.Updated,
		Metadata: &beadspb.Metadata{
			JiraKey:       jiraIssue.Key,
			JiraId:        jiraIssue.Id,
			JiraIssueType: jiraIssue.Fields.IssueType.Name,
		},
	}

	// Set assignee if present
	if jiraIssue.Fields.Assignee != nil {
		issue.Assignee = jiraIssue.Fields.Assignee.EmailAddress
		if issue.Assignee == "" {
			issue.Assignee = jiraIssue.Fields.Assignee.DisplayName
		}
	}

	// Link to epic if this issue belongs to one
	if jiraIssue.Fields.Parent != nil {
		// Check if parent is an epic
		if jiraIssue.Fields.Parent.Fields.IssueType.Name == "Epic" {
			if epicID, exists := c.epicMap[jiraIssue.Fields.Parent.Key]; exists {
				issue.Epic = epicID
			}
		}
	}

	// Handle dependencies from parent-child relationships
	if jiraIssue.Fields.Parent != nil && jiraIssue.Fields.IssueType.Subtask {
		// Subtasks depend on their parent (unless parent is an epic)
		if jiraIssue.Fields.Parent.Fields.IssueType.Name != "Epic" {
			parentBeadsID := c.generateBeadsID(jiraIssue.Fields.Parent.Key)
			issue.DependsOn = append(issue.DependsOn, parentBeadsID)
		}
	}

	return issue, nil
}

// addDependencies adds dependency relationships from Jira issue links
func (c *ProtoConverter) addDependencies(jiraExport *jirapb.Export, beadsExport *beadspb.Export) error {
	// Get dependencies from Jira
	jiraDeps := c.getDependencies(jiraExport)

	// Create a map of Jira keys to beads issue indices
	beadsIssueMap := make(map[string]int)
	for i, issue := range beadsExport.Issues {
		beadsIssueMap[issue.Metadata.JiraKey] = i
	}

	// Add dependencies to beads issues
	for jiraKey, depKeys := range jiraDeps {
		beadsIdx, exists := beadsIssueMap[jiraKey]
		if !exists {
			continue // Issue might be an epic
		}

		for _, depKey := range depKeys {
			depBeadsID := c.generateBeadsID(depKey)
			// Avoid duplicates
			if !contains(beadsExport.Issues[beadsIdx].DependsOn, depBeadsID) {
				beadsExport.Issues[beadsIdx].DependsOn = append(
					beadsExport.Issues[beadsIdx].DependsOn,
					depBeadsID,
				)
			}
		}
	}

	return nil
}

// mapStatus maps Jira status to beads status
func (c *ProtoConverter) mapStatus(jiraStatus *jirapb.Status) beadspb.Status {
	if jiraStatus == nil || jiraStatus.StatusCategory == nil {
		return beadspb.Status_STATUS_OPEN
	}

	switch jiraStatus.StatusCategory.Key {
	case "new":
		return beadspb.Status_STATUS_OPEN
	case "indeterminate":
		return beadspb.Status_STATUS_IN_PROGRESS
	case "done":
		return beadspb.Status_STATUS_CLOSED
	default:
		// Check specific status names
		statusName := strings.ToLower(jiraStatus.Name)
		if strings.Contains(statusName, "block") {
			return beadspb.Status_STATUS_BLOCKED
		}
		if strings.Contains(statusName, "progress") || strings.Contains(statusName, "doing") {
			return beadspb.Status_STATUS_IN_PROGRESS
		}
		if strings.Contains(statusName, "done") || strings.Contains(statusName, "closed") {
			return beadspb.Status_STATUS_CLOSED
		}
		return beadspb.Status_STATUS_OPEN
	}
}

// mapPriority maps Jira priority to beads priority
func (c *ProtoConverter) mapPriority(jiraPriority *jirapb.Priority) beadspb.Priority {
	if jiraPriority == nil {
		return beadspb.Priority_PRIORITY_P2
	}

	priorityName := strings.ToLower(jiraPriority.Name)

	switch {
	case strings.Contains(priorityName, "critical") || strings.Contains(priorityName, "highest"):
		return beadspb.Priority_PRIORITY_P0
	case strings.Contains(priorityName, "high"):
		return beadspb.Priority_PRIORITY_P1
	case strings.Contains(priorityName, "medium"):
		return beadspb.Priority_PRIORITY_P2
	case strings.Contains(priorityName, "lowest"):
		return beadspb.Priority_PRIORITY_P4
	case strings.Contains(priorityName, "low"):
		return beadspb.Priority_PRIORITY_P3
	default:
		// Default to medium priority
		return beadspb.Priority_PRIORITY_P2
	}
}

// generateBeadsID generates a beads-friendly ID from a Jira key
// Converts "PROJ-123" to "proj-123"
func (c *ProtoConverter) generateBeadsID(jiraKey string) string {
	return strings.ToLower(jiraKey)
}

// buildIssueMap creates a map of issue keys to issues for quick lookup
func (c *ProtoConverter) buildIssueMap(export *jirapb.Export) map[string]*jirapb.Issue {
	issueMap := make(map[string]*jirapb.Issue)
	for _, issue := range export.Issues {
		issueMap[issue.Key] = issue
	}
	return issueMap
}

// getEpics returns all issues that are epics
func (c *ProtoConverter) getEpics(export *jirapb.Export) []*jirapb.Issue {
	var epics []*jirapb.Issue
	for _, issue := range export.Issues {
		if issue.Fields.IssueType.Name == "Epic" {
			epics = append(epics, issue)
		}
	}
	return epics
}

// getDependencies extracts dependency relationships from issue links
func (c *ProtoConverter) getDependencies(export *jirapb.Export) map[string][]string {
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
