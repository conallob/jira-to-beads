package beads

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	pb "github.com/conallob/jira-beads-sync/gen/beads"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v3"
)

// YAMLRenderer handles rendering protobuf beads to YAML files
type YAMLRenderer struct {
	outputDir string
}

// NewYAMLRenderer creates a new YAML renderer
func NewYAMLRenderer(outputDir string) *YAMLRenderer {
	return &YAMLRenderer{
		outputDir: outputDir,
	}
}

// RenderExport renders a beads export to YAML files
func (r *YAMLRenderer) RenderExport(export *pb.Export) error {
	if err := r.ensureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Render epics
	for _, epic := range export.Epics {
		if err := r.RenderEpic(epic); err != nil {
			return fmt.Errorf("failed to render epic %s: %w", epic.Id, err)
		}
	}

	// Render issues
	for _, issue := range export.Issues {
		if err := r.RenderIssue(issue); err != nil {
			return fmt.Errorf("failed to render issue %s: %w", issue.Id, err)
		}
	}

	return nil
}

// RenderIssue renders a single issue to YAML
func (r *YAMLRenderer) RenderIssue(issue *pb.Issue) error {
	if err := r.ensureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}
	filename := filepath.Join(r.outputDir, ".beads", "issues", fmt.Sprintf("%s.yaml", issue.Id))
	return r.renderToYAML(filename, issue)
}

// RenderEpic renders a single epic to YAML
func (r *YAMLRenderer) RenderEpic(epic *pb.Epic) error {
	if err := r.ensureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}
	filename := filepath.Join(r.outputDir, ".beads", "epics", fmt.Sprintf("%s.yaml", epic.Id))
	return r.renderToYAML(filename, epic)
}

// RenderIssueToString renders an issue to YAML string
func (r *YAMLRenderer) RenderIssueToString(issue *pb.Issue) (string, error) {
	return r.protoToYAMLString(issue)
}

// RenderEpicToString renders an epic to YAML string
func (r *YAMLRenderer) RenderEpicToString(epic *pb.Epic) (string, error) {
	return r.protoToYAMLString(epic)
}

// ParseIssueFile reads a beads issue from a YAML file into protobuf
func (r *YAMLRenderer) ParseIssueFile(filename string) (*pb.Issue, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Convert YAML to JSON (intermediate format)
	jsonData, err := r.yamlToJSON(yamlData)
	if err != nil {
		return nil, err
	}

	// Convert JSON to protobuf
	issue := &pb.Issue{}
	if err := protojson.Unmarshal(jsonData, issue); err != nil {
		return nil, fmt.Errorf("failed to parse issue: %w", err)
	}

	return issue, nil
}

// ParseEpicFile reads a beads epic from a YAML file into protobuf
func (r *YAMLRenderer) ParseEpicFile(filename string) (*pb.Epic, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var yamlData map[string]interface{}
	if err := yaml.Unmarshal(data, &yamlData); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Convert YAML to JSON (intermediate format)
	jsonData, err := r.yamlToJSON(yamlData)
	if err != nil {
		return nil, err
	}

	// Convert JSON to protobuf
	epic := &pb.Epic{}
	if err := protojson.Unmarshal(jsonData, epic); err != nil {
		return nil, fmt.Errorf("failed to parse epic: %w", err)
	}

	return epic, nil
}

// ensureDirectories creates the necessary beads directories
func (r *YAMLRenderer) ensureDirectories() error {
	issuesDir := filepath.Join(r.outputDir, ".beads", "issues")
	epicsDir := filepath.Join(r.outputDir, ".beads", "epics")

	if err := os.MkdirAll(issuesDir, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(epicsDir, 0755); err != nil {
		return err
	}

	return nil
}

// renderToYAML renders a protobuf message to a YAML file
func (r *YAMLRenderer) renderToYAML(filename string, msg interface{}) error {
	yamlStr, err := r.protoToYAMLString(msg)
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	_, err = file.WriteString(yamlStr)
	return err
}

// protoToYAMLString converts a protobuf message to YAML string
func (r *YAMLRenderer) protoToYAMLString(msg interface{}) (string, error) {
	// First convert protobuf to JSON
	marshaler := protojson.MarshalOptions{
		UseProtoNames:   true,
		EmitUnpopulated: false,
		Indent:          "  ",
	}

	var jsonBytes []byte
	var err error

	switch v := msg.(type) {
	case *pb.Issue:
		jsonBytes, err = marshaler.Marshal(v)
	case *pb.Epic:
		jsonBytes, err = marshaler.Marshal(v)
	default:
		return "", fmt.Errorf("unsupported message type: %T", msg)
	}

	if err != nil {
		return "", fmt.Errorf("failed to marshal protobuf to JSON: %w", err)
	}

	// Parse JSON into a map
	var jsonData map[string]interface{}
	if err := yaml.Unmarshal(jsonBytes, &jsonData); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert to YAML-friendly format
	yamlData := r.convertToYAMLFormat(jsonData)

	// Marshal to YAML
	yamlBytes, err := yaml.Marshal(yamlData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal to YAML: %w", err)
	}

	return string(yamlBytes), nil
}

// convertToYAMLFormat converts JSON format to more human-readable YAML format
func (r *YAMLRenderer) convertToYAMLFormat(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		// Convert snake_case keys to camelCase for YAML readability
		yamlKey := r.toYAMLKey(key)

		switch v := value.(type) {
		case map[string]interface{}:
			result[yamlKey] = r.convertToYAMLFormat(v)
		case []interface{}:
			result[yamlKey] = v
		case string:
			// Convert enum values to lowercase strings without prefix
			switch key {
			case "status":
				result[yamlKey] = r.statusToYAML(v)
			case "priority":
				result[yamlKey] = r.priorityToYAML(v)
			default:
				result[yamlKey] = v
			}
		default:
			result[yamlKey] = v
		}
	}

	return result
}

// toYAMLKey converts protobuf field names to YAML-friendly names
func (r *YAMLRenderer) toYAMLKey(key string) string {
	switch key {
	case "depends_on":
		return "dependsOn"
	case "jira_key":
		return "jiraKey"
	case "jira_id":
		return "jiraId"
	case "jira_issue_type":
		return "jiraIssueType"
	case "repositories":
		return "repositories"
	default:
		return key
	}
}

// statusToYAML converts status enum to YAML string
func (r *YAMLRenderer) statusToYAML(status string) string {
	switch status {
	case "STATUS_OPEN":
		return "open"
	case "STATUS_IN_PROGRESS":
		return "in_progress"
	case "STATUS_BLOCKED":
		return "blocked"
	case "STATUS_CLOSED":
		return "closed"
	default:
		return "open"
	}
}

// priorityToYAML converts priority enum to YAML string
func (r *YAMLRenderer) priorityToYAML(priority string) string {
	switch priority {
	case "PRIORITY_P0":
		return "p0"
	case "PRIORITY_P1":
		return "p1"
	case "PRIORITY_P2":
		return "p2"
	case "PRIORITY_P3":
		return "p3"
	case "PRIORITY_P4":
		return "p4"
	default:
		return "p2"
	}
}

// yamlToJSON converts YAML data to JSON format for protobuf unmarshaling
func (r *YAMLRenderer) yamlToJSON(yamlData map[string]interface{}) ([]byte, error) {
	// Convert YAML format back to protobuf JSON format
	jsonData := r.convertFromYAMLFormat(yamlData)

	// Marshal to JSON bytes (not YAML!)
	return json.Marshal(jsonData)
}

// convertFromYAMLFormat converts YAML format back to protobuf JSON format
func (r *YAMLRenderer) convertFromYAMLFormat(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		// Convert camelCase back to snake_case
		protoKey := r.toProtoKey(key)

		switch v := value.(type) {
		case map[string]interface{}:
			result[protoKey] = r.convertFromYAMLFormat(v)
		case []interface{}:
			result[protoKey] = v
		case string:
			// Convert enum values back to protobuf format
			switch key {
			case "status":
				result[protoKey] = r.statusFromYAML(v)
			case "priority":
				result[protoKey] = r.priorityFromYAML(v)
			case "created", "updated":
				// Timestamp fields - keep as RFC3339 string, protojson will handle it
				result[protoKey] = v
			default:
				result[protoKey] = v
			}
		default:
			result[protoKey] = v
		}
	}

	return result
}

// toProtoKey converts YAML keys to protobuf field names
func (r *YAMLRenderer) toProtoKey(key string) string {
	switch key {
	case "dependsOn":
		return "depends_on"
	case "jiraKey":
		return "jira_key"
	case "jiraId":
		return "jira_id"
	case "jiraIssueType":
		return "jira_issue_type"
	case "repositories":
		return "repositories"
	default:
		return key
	}
}

// statusFromYAML converts YAML status string to protobuf enum
func (r *YAMLRenderer) statusFromYAML(status string) string {
	switch status {
	case "open":
		return "STATUS_OPEN"
	case "in_progress":
		return "STATUS_IN_PROGRESS"
	case "blocked":
		return "STATUS_BLOCKED"
	case "closed":
		return "STATUS_CLOSED"
	default:
		return "STATUS_OPEN"
	}
}

// priorityFromYAML converts YAML priority string to protobuf enum
func (r *YAMLRenderer) priorityFromYAML(priority string) string {
	switch priority {
	case "p0":
		return "PRIORITY_P0"
	case "p1":
		return "PRIORITY_P1"
	case "p2":
		return "PRIORITY_P2"
	case "p3":
		return "PRIORITY_P3"
	case "p4":
		return "PRIORITY_P4"
	default:
		return "PRIORITY_P2"
	}
}

// AddRepositoryAnnotation adds a repository to an issue's metadata
func (r *YAMLRenderer) AddRepositoryAnnotation(issueID, repository string) error {
	// Construct issue file path
	filename := filepath.Join(r.outputDir, ".beads", "issues", fmt.Sprintf("%s.yaml", issueID))

	// Check if file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return fmt.Errorf("issue file not found: %s", filename)
	}

	// Parse existing issue
	issue, err := r.ParseIssueFile(filename)
	if err != nil {
		return fmt.Errorf("failed to parse issue: %w", err)
	}

	// Initialize metadata if needed
	if issue.Metadata == nil {
		issue.Metadata = &pb.Metadata{}
	}

	// Check if repository already exists
	for _, repo := range issue.Metadata.Repositories {
		if repo == repository {
			return fmt.Errorf("repository '%s' is already associated with issue %s", repository, issueID)
		}
	}

	// Add repository
	issue.Metadata.Repositories = append(issue.Metadata.Repositories, repository)

	// Render updated issue back to file
	if err := r.RenderIssue(issue); err != nil {
		return fmt.Errorf("failed to save updated issue: %w", err)
	}

	return nil
}
