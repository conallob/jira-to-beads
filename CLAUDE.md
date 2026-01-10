# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based CLI tool to convert Jira task trees into beads issues. It handles the hierarchical structure of Jira tasks (epics, stories, subtasks) and maps them to the beads issue tracking system while preserving dependencies and relationships.

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

- `cmd/jira-to-beads/` - Main application entry point
- `internal/jira/` - Jira JSON to protobuf adapter
- `internal/beads/` - YAML rendering layer on top of protobuf
- `internal/converter/` - Conversion logic between Jira and beads protobuf
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

Users will likely:
1. Export or fetch Jira tasks via API
2. Run this tool to convert the Jira data structure
3. Import the converted issues into beads system
4. Validate that hierarchy and dependencies are correct
