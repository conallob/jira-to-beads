package beads

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Generator handles creating beads issue files
type Generator struct {
	outputDir string
}

// NewGenerator creates a new beads generator
func NewGenerator(outputDir string) *Generator {
	return &Generator{
		outputDir: outputDir,
	}
}

// Generate creates beads issue files from the export
func (g *Generator) Generate(export *Export) error {
	if err := g.ensureDirectories(); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Generate epics first
	if err := g.generateEpics(export.Epics); err != nil {
		return fmt.Errorf("failed to generate epics: %w", err)
	}

	// Generate issues
	if err := g.generateIssues(export.Issues); err != nil {
		return fmt.Errorf("failed to generate issues: %w", err)
	}

	return nil
}

// ensureDirectories creates the necessary beads directories
func (g *Generator) ensureDirectories() error {
	issuesDir := filepath.Join(g.outputDir, ".beads", "issues")
	epicsDir := filepath.Join(g.outputDir, ".beads", "epics")

	if err := os.MkdirAll(issuesDir, 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(epicsDir, 0755); err != nil {
		return err
	}

	return nil
}

// generateEpics creates YAML files for epics
func (g *Generator) generateEpics(epics []Epic) error {
	for _, epic := range epics {
		if err := g.writeEpic(&epic); err != nil {
			return fmt.Errorf("failed to write epic %s: %w", epic.ID, err)
		}
	}
	return nil
}

// generateIssues creates YAML files for issues
func (g *Generator) generateIssues(issues []Issue) error {
	for _, issue := range issues {
		if err := g.writeIssue(&issue); err != nil {
			return fmt.Errorf("failed to write issue %s: %w", issue.ID, err)
		}
	}
	return nil
}

// writeEpic writes a single epic to a YAML file
func (g *Generator) writeEpic(epic *Epic) error {
	filename := filepath.Join(g.outputDir, ".beads", "epics", fmt.Sprintf("%s.yaml", epic.ID))
	return g.writeYAML(filename, epic)
}

// writeIssue writes a single issue to a YAML file
func (g *Generator) writeIssue(issue *Issue) error {
	filename := filepath.Join(g.outputDir, ".beads", "issues", fmt.Sprintf("%s.yaml", issue.ID))
	return g.writeYAML(filename, issue)
}

// writeYAML writes data to a YAML file
func (g *Generator) writeYAML(filename string, data any) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)

	if err := encoder.Encode(data); err != nil {
		return err
	}

	if err := encoder.Close(); err != nil {
		return err
	}

	return nil
}

// GenerateToString converts an issue to YAML string (useful for testing)
func (g *Generator) GenerateToString(issue *Issue) (string, error) {
	data, err := yaml.Marshal(issue)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// GenerateEpicToString converts an epic to YAML string (useful for testing)
func (g *Generator) GenerateEpicToString(epic *Epic) (string, error) {
	data, err := yaml.Marshal(epic)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ParseIssueFile reads a beads issue from a YAML file
func (g *Generator) ParseIssueFile(filename string) (*Issue, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var issue Issue
	if err := yaml.Unmarshal(data, &issue); err != nil {
		return nil, fmt.Errorf("failed to parse issue: %w", err)
	}

	return &issue, nil
}

// ParseEpicFile reads a beads epic from a YAML file
func (g *Generator) ParseEpicFile(filename string) (*Epic, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	var epic Epic
	if err := yaml.Unmarshal(data, &epic); err != nil {
		return nil, fmt.Errorf("failed to parse epic: %w", err)
	}

	return &epic, nil
}
