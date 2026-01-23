package converter

import (
	"fmt"

	"github.com/conallob/jira-beads-sync/internal/beads"
	"github.com/conallob/jira-beads-sync/internal/jira"
)

// Pipeline orchestrates the full conversion from Jira JSON to beads JSONL
type Pipeline struct {
	jiraAdapter   *jira.Adapter
	converter     *ProtoConverter
	jsonlRenderer *beads.JSONLRenderer
}

// NewPipeline creates a new conversion pipeline
func NewPipeline(outputDir string) *Pipeline {
	return &Pipeline{
		jiraAdapter:   jira.NewAdapter(),
		converter:     NewProtoConverter(),
		jsonlRenderer: beads.NewJSONLRenderer(outputDir),
	}
}

// ConvertFile converts a Jira JSON export file to beads JSONL files
func (p *Pipeline) ConvertFile(jiraFile string) error {
	// Step 1: Parse Jira JSON to protobuf
	jiraExport, err := p.jiraAdapter.ParseFile(jiraFile)
	if err != nil {
		return fmt.Errorf("failed to parse Jira file: %w", err)
	}

	// Step 2: Convert Jira protobuf to beads protobuf
	beadsExport, err := p.converter.Convert(jiraExport)
	if err != nil {
		return fmt.Errorf("failed to convert to beads format: %w", err)
	}

	// Step 3: Render beads protobuf to JSONL files
	if err := p.jsonlRenderer.RenderExport(beadsExport); err != nil {
		return fmt.Errorf("failed to render JSONL files: %w", err)
	}

	return nil
}
