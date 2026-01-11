# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based CLI tool to convert Jira task trees into beads issues. It handles the hierarchical structure of Jira tasks (epics, stories, subtasks) and maps them to the beads issue tracking system while preserving dependencies and relationships.

The tool supports two modes:
1. **Quickstart mode**: Fetch issues directly from Jira API and convert them
2. **Convert mode**: Convert previously exported Jira JSON files

## Language & Tooling

**Language**: Go (Golang)

### Commands

**Build**:
```bash
go build -o jira-to-beads ./cmd/jira-to-beads
```

**Run**:
```bash
go run ./cmd/jira-to-beads [args]
```

**Test**:
```bash
go test ./...                    # Run all tests
go test -v ./...                 # Run with verbose output
go test -run TestFunctionName    # Run specific test
```

**Lint & Format**:
```bash
go fmt ./...                     # Format all code
golangci-lint run                # Run linter
```

**Generate Protobuf Code**:
```bash
protoc --go_out=. --go_opt=module=github.com/conallob/jira-to-beads --proto_path=proto proto/jira.proto proto/beads.proto
```

### Go Project Structure

- `cmd/jira-to-beads/` - Main application entry point with CLI commands
- `internal/jira/` - Jira integration
  - `adapter.go` - JSON to protobuf adapter for Jira exports
  - `client.go` - Jira REST API v2 client for fetching issues
- `internal/beads/` - YAML rendering layer on top of protobuf
- `internal/converter/` - Conversion logic between Jira and beads protobuf
- `internal/config/` - Configuration management (credentials, base URL)
- `proto/` - Protocol Buffer definitions (source of truth for data structures)
- `gen/jira/` - Generated Go code from jira.proto
- `gen/beads/` - Generated Go code from beads.proto
- `testdata/` - Sample Jira exports and expected beads output
- `go.mod` - Go module definition

## Beads Integration

This tool creates issues for the beads system (git-backed issue tracker). Key beads concepts:

- **Issues**: Work items stored in `.beads/issues/` as YAML files
- **Dependencies**: Issues can depend on other issues using `dependsOn` field
- **Status**: `open`, `in_progress`, `blocked`, or `closed`
- **Priority**: `p0` (critical) through `p4` (low)
- **Epics**: Large features that group related issues using `epic` field

Relevant beads commands for testing:
- `bd list` - List all issues
- `bd show <id>` - Show issue details
- `bd create` - Create new issue interactively
- `bd dep add <issue> <dependency>` - Add dependency between issues
- `bd epic create <name>` - Create a new epic

## Architecture

This tool uses Protocol Buffers as the internal data structure format with rendering layers for external formats:

### Data Flow

```
JSON (Jira) → Protobuf (Jira) → Protobuf (Beads) → YAML (Beads)
     ↓              ↓                  ↓               ↓
Adapter    Generated Types    Converter      Renderer
```

1. **Jira Adapter** (`internal/jira/adapter.go`): Parses Jira JSON exports into protobuf messages defined in `proto/jira.proto`
2. **Converter** (`internal/converter/proto_converter.go`): Transforms Jira protobuf to beads protobuf with mappings for status, priority, and dependencies
3. **YAML Renderer** (`internal/beads/yaml.go`): Renders beads protobuf to human-readable YAML files

### Why Protocol Buffers?

- **Single Source of Truth**: Data structures defined once in `.proto` files
- **Type Safety**: Strong typing across all layers
- **Versioning**: Built-in support for schema evolution
- **Multiple Formats**: Easy to add new rendering formats (JSON, TOML) without changing core logic
- **Performance**: Efficient serialization when needed

### Configuration Format

Users interact with YAML, which is just a rendering layer on top of protobuf. The tool:
- Reads JSON (Jira exports)
- Processes data as protobuf internally
- Writes YAML (beads issues)

This architecture separates concerns: protobuf handles data structure and validation, while format-specific code handles I/O.

## Development Approach

When modifying data structures:

1. **Update `.proto` files first** in `proto/` directory
2. **Regenerate Go code** using the protoc command above
3. **Update rendering layers** if field names or types change
4. **Update tests** to reflect schema changes

When implementing new features:

1. **Preserve Hierarchy**: Jira epics → beads epics, stories → issues, subtasks → issues with dependencies
2. **Map Dependencies**: Convert Jira issue links (blocks, depends on) to beads dependencies
3. **Status Mapping**: Map Jira status categories to beads status enum
4. **Priority Mapping**: Convert Jira priorities to beads p0-p4 scale
5. **Handle Metadata**: Preserve Jira metadata (key, ID, type) for traceability

## Expected Workflow

### Quickstart Mode (Recommended)
Users will:
1. Configure Jira credentials once: `jira-to-beads configure`
2. Fetch and convert issues: `jira-to-beads quickstart PROJ-123`
3. The tool will recursively fetch all dependencies and convert to beads format
4. Validate that hierarchy and dependencies are correct

### Convert Mode
Users will:
1. Export Jira tasks to JSON file
2. Run this tool to convert the Jira data structure
3. Import the converted issues into beads system
4. Validate that hierarchy and dependencies are correct

## Jira API Integration

The tool uses Jira REST API v2 for fetching issues:

- **Authentication**: Basic Auth with username and API token
- **Configuration**: Supports config file, environment variables, or interactive setup
- **Recursive Fetching**: Walks dependency graph including:
  - Subtasks (via `fields.subtasks`)
  - Linked issues (via `fields.issuelinks`, both inward and outward)
  - Parent issues (via `fields.parent`, excluding epics)
- **Duplicate Prevention**: Uses visited map to avoid infinite loops

Key files:
- `internal/jira/client.go`: Jira API client with recursive dependency walking
- `internal/config/config.go`: Configuration management

## Claude Code Plugin

This repository includes a Claude Code plugin that enables importing Jira issues through natural language commands or slash commands.

### Plugin Commands

- `/import-jira <jira-url-or-key>` - Import a Jira issue and its dependency tree
- `/configure-jira` - Configure Jira API credentials
- `/convert-jira-export <file>` - Convert a Jira export JSON file

### Natural Language Usage

Claude can understand requests like:
- "Import PROJ-123 from Jira"
- "Fetch the Jira issue TEAM-456 and all its dependencies"
- "Configure my Jira credentials"
- "Convert jira-export.json to beads format"

### Using in Your Project

Add to your project's `.claude/CLAUDE.md`:

```markdown
# Jira Integration

When working on Jira issues, import them into beads:
- Import: Ask Claude to "import <jira-key> from Jira"
- View: Use `bd list` and `bd show <id>` to see imported issues
- The tool automatically fetches all dependencies and related issues
```

### Plugin Installation

For users:
```bash
# Install the CLI tool first
go install github.com/conallob/jira-to-beads/cmd/jira-to-beads@latest

# Start Claude with plugin enabled
claude --plugin-dir /path/to/jira-to-beads
```

For development:
```bash
# From this repository
claude --plugin-dir .
```

See [PLUGIN.md](PLUGIN.md) for complete plugin documentation.
