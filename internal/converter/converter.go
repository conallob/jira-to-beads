package converter

import (
	"fmt"
	"strings"

	"github.com/conallob/jira-beads-sync/internal/beads"
	"github.com/conallob/jira-beads-sync/internal/jira"
)

// Converter handles converting Jira exports to beads format
type Converter struct {
	jiraParser *jira.Parser
	issueMap   map[string]*jira.Issue // Map of Jira keys to issues
	epicMap    map[string]string      // Map of Jira epic keys to beads epic IDs
}

// NewConverter creates a new Jira to beads converter
func NewConverter() *Converter {
	return &Converter{
		jiraParser: jira.NewParser(),
		issueMap:   make(map[string]*jira.Issue),
		epicMap:    make(map[string]string),
	}
}

// Convert converts a Jira export to beads format
func (c *Converter) Convert(jiraExport *jira.Export) (*beads.Export, error) {
	if jiraExport == nil {
		return nil, fmt.Errorf("jira export is nil")
	}

	// Build issue map for quick lookups
	c.issueMap = c.jiraParser.BuildIssueMap(jiraExport)

	beadsExport := &beads.Export{
		Issues: []beads.Issue{},
		Epics:  []beads.Epic{},
	}

	// Convert epics first so we can reference them in issues
	epics := c.jiraParser.GetEpics(jiraExport)
	for _, jiraIssue := range epics {
		beadsEpic, err := c.convertEpic(jiraIssue)
		if err != nil {
			return nil, fmt.Errorf("failed to convert epic %s: %w", jiraIssue.Key, err)
		}
		beadsExport.Epics = append(beadsExport.Epics, *beadsEpic)
		c.epicMap[jiraIssue.Key] = beadsEpic.ID
	}

	// Convert all issues (stories, tasks, subtasks)
	for i := range jiraExport.Issues {
		jiraIssue := &jiraExport.Issues[i]

		// Skip epics as they've already been converted
		if jiraIssue.Fields.IssueType.Name == "Epic" {
			continue
		}

		beadsIssue, err := c.convertIssue(jiraIssue)
		if err != nil {
			return nil, fmt.Errorf("failed to convert issue %s: %w", jiraIssue.Key, err)
		}
		beadsExport.Issues = append(beadsExport.Issues, *beadsIssue)
	}

	// Add dependencies after all issues are converted
	if err := c.addDependencies(jiraExport, beadsExport); err != nil {
		return nil, fmt.Errorf("failed to add dependencies: %w", err)
	}

	return beadsExport, nil
}

// convertEpic converts a Jira epic to a beads epic
func (c *Converter) convertEpic(jiraIssue *jira.Issue) (*beads.Epic, error) {
	epic := &beads.Epic{
		ID:          c.generateBeadsID(jiraIssue.Key),
		Name:        jiraIssue.Fields.Summary,
		Description: jiraIssue.Fields.Description,
		Status:      c.mapStatus(jiraIssue.Fields.Status),
		Created:     jiraIssue.Fields.Created.Time,
		Updated:     jiraIssue.Fields.Updated.Time,
		Metadata: beads.Metadata{
			JiraKey:       jiraIssue.Key,
			JiraID:        jiraIssue.ID,
			JiraIssueType: jiraIssue.Fields.IssueType.Name,
		},
	}

	return epic, nil
}

// convertIssue converts a Jira issue to a beads issue
func (c *Converter) convertIssue(jiraIssue *jira.Issue) (*beads.Issue, error) {
	issue := &beads.Issue{
		ID:          c.generateBeadsID(jiraIssue.Key),
		Title:       jiraIssue.Fields.Summary,
		Description: jiraIssue.Fields.Description,
		Status:      c.mapStatus(jiraIssue.Fields.Status),
		Priority:    c.mapPriority(jiraIssue.Fields.Priority),
		Labels:      jiraIssue.Fields.Labels,
		Created:     jiraIssue.Fields.Created.Time,
		Updated:     jiraIssue.Fields.Updated.Time,
		Metadata: beads.Metadata{
			JiraKey:       jiraIssue.Key,
			JiraID:        jiraIssue.ID,
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
func (c *Converter) addDependencies(jiraExport *jira.Export, beadsExport *beads.Export) error {
	// Get dependencies from Jira
	jiraDeps := c.jiraParser.GetDependencies(jiraExport)

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
func (c *Converter) mapStatus(jiraStatus jira.Status) beads.Status {
	switch jiraStatus.StatusCategory.Key {
	case "new":
		return beads.StatusOpen
	case "indeterminate":
		return beads.StatusInProgress
	case "done":
		return beads.StatusClosed
	default:
		// Check specific status names
		statusName := strings.ToLower(jiraStatus.Name)
		if strings.Contains(statusName, "block") {
			return beads.StatusBlocked
		}
		if strings.Contains(statusName, "progress") || strings.Contains(statusName, "doing") {
			return beads.StatusInProgress
		}
		if strings.Contains(statusName, "done") || strings.Contains(statusName, "closed") {
			return beads.StatusClosed
		}
		return beads.StatusOpen
	}
}

// mapPriority maps Jira priority to beads priority
func (c *Converter) mapPriority(jiraPriority jira.Priority) beads.Priority {
	priorityName := strings.ToLower(jiraPriority.Name)

	switch {
	case strings.Contains(priorityName, "critical") || strings.Contains(priorityName, "highest"):
		return beads.PriorityP0
	case strings.Contains(priorityName, "high"):
		return beads.PriorityP1
	case strings.Contains(priorityName, "medium"):
		return beads.PriorityP2
	case strings.Contains(priorityName, "lowest"):
		return beads.PriorityP4
	case strings.Contains(priorityName, "low"):
		return beads.PriorityP3
	default:
		// Default to medium priority
		return beads.PriorityP2
	}
}

// generateBeadsID generates a beads-friendly ID from a Jira key
// Converts "PROJ-123" to "proj-123"
func (c *Converter) generateBeadsID(jiraKey string) string {
	return strings.ToLower(jiraKey)
}
