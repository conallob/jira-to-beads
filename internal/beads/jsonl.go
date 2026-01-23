package beads

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pb "github.com/conallob/jira-beads-sync/gen/beads"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// JSONLRenderer handles rendering protobuf beads to JSONL files
type JSONLRenderer struct {
	outputDir string
}

// NewJSONLRenderer creates a new JSONL renderer
func NewJSONLRenderer(outputDir string) *JSONLRenderer {
	return &JSONLRenderer{
		outputDir: outputDir,
	}
}

// RenderExport renders a beads export to JSONL files
func (r *JSONLRenderer) RenderExport(export *pb.Export) error {
	if err := r.ensureDirectory(); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Render all issues to a single JSONL file
	issuesFile := filepath.Join(r.outputDir, ".beads", "issues.jsonl")
	if err := r.renderIssuesToJSONL(issuesFile, export.Issues); err != nil {
		return fmt.Errorf("failed to render issues: %w", err)
	}

	// Render all epics to a single JSONL file
	if len(export.Epics) > 0 {
		epicsFile := filepath.Join(r.outputDir, ".beads", "epics.jsonl")
		if err := r.renderEpicsToJSONL(epicsFile, export.Epics); err != nil {
			return fmt.Errorf("failed to render epics: %w", err)
		}
	}

	return nil
}

// ensureDirectory creates the necessary beads directory
func (r *JSONLRenderer) ensureDirectory() error {
	beadsDir := filepath.Join(r.outputDir, ".beads")
	return os.MkdirAll(beadsDir, 0755)
}

// renderIssuesToJSONL renders issues to a JSONL file
func (r *JSONLRenderer) renderIssuesToJSONL(filename string, issues []*pb.Issue) (err error) {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	encoder := json.NewEncoder(file)
	for _, issue := range issues {
		jsonIssue := r.issueToJSON(issue)
		if err := encoder.Encode(jsonIssue); err != nil {
			return fmt.Errorf("failed to encode issue %s: %w", issue.Id, err)
		}
	}

	return nil
}

// renderEpicsToJSONL renders epics to a JSONL file
func (r *JSONLRenderer) renderEpicsToJSONL(filename string, epics []*pb.Epic) (err error) {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	encoder := json.NewEncoder(file)
	for _, epic := range epics {
		jsonEpic := r.epicToJSON(epic)
		if err := encoder.Encode(jsonEpic); err != nil {
			return fmt.Errorf("failed to encode epic %s: %w", epic.Id, err)
		}
	}

	return nil
}

// BeadsIssue represents a beads issue in JSON format
type BeadsIssue struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Status      string            `json:"status"`
	Priority    string            `json:"priority,omitempty"`
	Epic        string            `json:"epic,omitempty"`
	Assignee    string            `json:"assignee,omitempty"`
	Labels      []string          `json:"labels,omitempty"`
	DependsOn   []string          `json:"dependsOn,omitempty"`
	Created     string            `json:"created,omitempty"`
	Updated     string            `json:"updated,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// BeadsEpic represents a beads epic in JSON format
type BeadsEpic struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Status      string            `json:"status"`
	Created     string            `json:"created,omitempty"`
	Updated     string            `json:"updated,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// issueToJSON converts a protobuf issue to JSON format
func (r *JSONLRenderer) issueToJSON(issue *pb.Issue) *BeadsIssue {
	jsonIssue := &BeadsIssue{
		ID:          issue.Id,
		Title:       issue.Title,
		Description: issue.Description,
		Status:      r.statusToString(issue.Status),
		Priority:    r.priorityToString(issue.Priority),
		Epic:        issue.Epic,
		Assignee:    issue.Assignee,
		Labels:      issue.Labels,
		DependsOn:   issue.DependsOn,
	}

	if issue.Created != nil {
		jsonIssue.Created = r.timestampToString(issue.Created)
	}
	if issue.Updated != nil {
		jsonIssue.Updated = r.timestampToString(issue.Updated)
	}

	if issue.Metadata != nil {
		jsonIssue.Metadata = make(map[string]string)
		if issue.Metadata.JiraKey != "" {
			jsonIssue.Metadata["jiraKey"] = issue.Metadata.JiraKey
		}
		if issue.Metadata.JiraId != "" {
			jsonIssue.Metadata["jiraId"] = issue.Metadata.JiraId
		}
		if issue.Metadata.JiraIssueType != "" {
			jsonIssue.Metadata["jiraIssueType"] = issue.Metadata.JiraIssueType
		}
		for k, v := range issue.Metadata.Custom {
			jsonIssue.Metadata[k] = v
		}
	}

	return jsonIssue
}

// epicToJSON converts a protobuf epic to JSON format
func (r *JSONLRenderer) epicToJSON(epic *pb.Epic) *BeadsEpic {
	jsonEpic := &BeadsEpic{
		ID:          epic.Id,
		Name:        epic.Name,
		Description: epic.Description,
		Status:      r.statusToString(epic.Status),
	}

	if epic.Created != nil {
		jsonEpic.Created = r.timestampToString(epic.Created)
	}
	if epic.Updated != nil {
		jsonEpic.Updated = r.timestampToString(epic.Updated)
	}

	if epic.Metadata != nil {
		jsonEpic.Metadata = make(map[string]string)
		if epic.Metadata.JiraKey != "" {
			jsonEpic.Metadata["jiraKey"] = epic.Metadata.JiraKey
		}
		if epic.Metadata.JiraId != "" {
			jsonEpic.Metadata["jiraId"] = epic.Metadata.JiraId
		}
		if epic.Metadata.JiraIssueType != "" {
			jsonEpic.Metadata["jiraIssueType"] = epic.Metadata.JiraIssueType
		}
	}

	return jsonEpic
}

// statusToString converts status enum to string
func (r *JSONLRenderer) statusToString(status pb.Status) string {
	switch status {
	case pb.Status_STATUS_OPEN:
		return "open"
	case pb.Status_STATUS_IN_PROGRESS:
		return "in_progress"
	case pb.Status_STATUS_BLOCKED:
		return "blocked"
	case pb.Status_STATUS_CLOSED:
		return "closed"
	default:
		return "open"
	}
}

// priorityToString converts priority enum to string
func (r *JSONLRenderer) priorityToString(priority pb.Priority) string {
	switch priority {
	case pb.Priority_PRIORITY_P0:
		return "p0"
	case pb.Priority_PRIORITY_P1:
		return "p1"
	case pb.Priority_PRIORITY_P2:
		return "p2"
	case pb.Priority_PRIORITY_P3:
		return "p3"
	case pb.Priority_PRIORITY_P4:
		return "p4"
	default:
		return ""
	}
}

// timestampToString converts protobuf timestamp to RFC3339 string
func (r *JSONLRenderer) timestampToString(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().Format("2006-01-02T15:04:05Z07:00")
}

// AddRepositoryAnnotation adds a repository to an issue's metadata in the JSONL file
func (r *JSONLRenderer) AddRepositoryAnnotation(issueID, repository string) (err error) {
	issuesFile := filepath.Join(r.outputDir, ".beads", "issues.jsonl")

	// Read all issues
	file, err := os.Open(issuesFile)
	if err != nil {
		return fmt.Errorf("failed to open issues file: %w", err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var issues []*BeadsIssue
	scanner := bufio.NewScanner(file)
	found := false
	for scanner.Scan() {
		var issue BeadsIssue
		if err := json.Unmarshal(scanner.Bytes(), &issue); err != nil {
			return fmt.Errorf("failed to parse issue: %w", err)
		}

		// If this is the target issue, add the repository
		if issue.ID == issueID {
			found = true
			if issue.Metadata == nil {
				issue.Metadata = make(map[string]string)
			}

			// Check for duplicate (storing as comma-separated in metadata)
			reposKey := "repositories"
			existingRepos := issue.Metadata[reposKey]
			if existingRepos != "" {
				repos := strings.Split(existingRepos, ",")
				for _, r := range repos {
					if strings.TrimSpace(r) == repository {
						return fmt.Errorf("repository '%s' is already associated with issue %s", repository, issueID)
					}
				}
				issue.Metadata[reposKey] = existingRepos + "," + repository
			} else {
				issue.Metadata[reposKey] = repository
			}
		}

		issues = append(issues, &issue)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading issues file: %w", err)
	}

	if !found {
		return fmt.Errorf("issue %s not found in issues file", issueID)
	}

	// Write all issues back to the file
	outFile, err := os.Create(issuesFile)
	if err != nil {
		return fmt.Errorf("failed to create issues file: %w", err)
	}
	defer func() {
		if cerr := outFile.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	encoder := json.NewEncoder(outFile)
	for _, issue := range issues {
		if err := encoder.Encode(issue); err != nil {
			return fmt.Errorf("failed to write issue: %w", err)
		}
	}

	return nil
}
