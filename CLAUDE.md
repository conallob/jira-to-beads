# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based CLI tool to sync Jira task trees with beads issues. It handles the hierarchical structure of Jira tasks (epics, stories, subtasks) and provides bidirectional synchronization with the beads issue tracking system while preserving dependencies and relationships.

The tool supports multiple modes:
1. **Quickstart mode**: Fetch issues directly from Jira API and sync them to beads
2. **Sync mode**: Sync beads state changes back to Jira (bidirectional)
3. **Convert mode**: One-way conversion of previously exported Jira JSON files

## Language & Tooling

**Language**: Go (Golang)

### Commands

**Build**:
```bash
go build -o jira-beads-sync ./cmd/jira-beads-sync
```

**Run**:
```bash
go run ./cmd/jira-beads-sync [args]
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
protoc --go_out=. --go_opt=module=github.com/conallob/jira-beads-sync --proto_path=proto proto/jira.proto proto/beads.proto
```

### Go Project Structure

- `cmd/jira-beads-sync/` - Main application entry point with CLI commands
- `internal/jira/` - Jira integration
  - `adapter.go` - JSON to protobuf adapter for Jira exports
  - `client.go` - Jira REST API v2 client for fetching issues
- `internal/beads/` - JSONL rendering layer on top of protobuf
  - `jsonl.go` - Renders beads protobuf to JSONL format
- `internal/converter/` - Conversion logic between Jira and beads protobuf
- `internal/config/` - Configuration management (credentials, base URL)
- `proto/` - Protocol Buffer definitions (source of truth for data structures)
- `gen/jira/` - Generated Go code from jira.proto
- `gen/beads/` - Generated Go code from beads.proto
- `testdata/` - Sample Jira exports and expected beads output
- `go.mod` - Go module definition

## Beads Integration

This tool creates issues for the beads system (git-backed issue tracker).

**Official Beads Repository**: https://github.com/steveyegge/beads

Key beads concepts:

- **Issues**: Work items stored in `.beads/issues.jsonl` as JSONL (JSON Lines) format
- **Dependencies**: Issues can depend on other issues using `dependsOn` field
- **Status**: `open`, `in_progress`, `blocked`, or `closed`
- **Priority**: `p0` (critical) through `p4` (low)
- **Epics**: Large features that group related issues using `epic` field, stored in `.beads/epics.jsonl`

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
JSON (Jira) → Protobuf (Jira) → Protobuf (Beads) → JSONL (Beads)
     ↓              ↓                  ↓               ↓
Adapter    Generated Types    Converter      Renderer
```

1. **Jira Adapter** (`internal/jira/adapter.go`): Parses Jira JSON exports into protobuf messages defined in `proto/jira.proto`
2. **Converter** (`internal/converter/proto_converter.go`): Transforms Jira protobuf to beads protobuf with mappings for status, priority, and dependencies
3. **JSONL Renderer** (`internal/beads/jsonl.go`): Renders beads protobuf to JSONL files compatible with the beads CLI

### Why Protocol Buffers?

- **Single Source of Truth**: Data structures defined once in `.proto` files
- **Type Safety**: Strong typing across all layers
- **Versioning**: Built-in support for schema evolution
- **Multiple Formats**: Easy to add new rendering formats without changing core logic
- **Performance**: Efficient serialization when needed

### Output Format

The tool outputs JSONL (JSON Lines) format, which is the standard format used by beads:
- Reads JSON (Jira exports)
- Processes data as protobuf internally
- Writes JSONL (beads issues) to `.beads/issues.jsonl` and `.beads/epics.jsonl`

This architecture separates concerns: protobuf handles data structure and validation, while format-specific code handles I/O.

## Development Approach

### Git Workflow

**IMPORTANT**: All new development must be done in feature branches, not directly on `main`.

1. **Create a feature branch** before starting any new work:
   ```bash
   git checkout -b feature/your-feature-name
   # or for bug fixes:
   git checkout -b fix/bug-description
   ```

2. **Make your changes** in the feature branch

3. **Commit your changes** following conventional commit format:
   - `feat:` for new features
   - `fix:` for bug fixes
   - `docs:` for documentation changes
   - `test:` for test changes
   - `refactor:` for code refactoring
   - `ci:` for CI/CD changes

4. **Push the feature branch** to remote:
   ```bash
   git push origin feature/your-feature-name
   ```

5. **Create a Pull Request** against `main` branch:
   - Include a clear description of changes
   - Reference any related issues
   - Ensure all CI checks pass
   - Wait for code review before merging

6. **Never commit directly to `main`** - always use the PR workflow

### When Modifying Data Structures

1. **Update `.proto` files first** in `proto/` directory
2. **Regenerate Go code** using the protoc command above
3. **Update rendering layers** if field names or types change
4. **Update tests** to reflect schema changes

### When Implementing New Features

1. **Preserve Hierarchy**: Jira epics → beads epics, stories → issues, subtasks → issues with dependencies
2. **Map Dependencies**: Convert Jira issue links (blocks, depends on) to beads dependencies
3. **Status Mapping**: Map Jira status categories to beads status enum
4. **Priority Mapping**: Convert Jira priorities to beads p0-p4 scale
5. **Handle Metadata**: Preserve Jira metadata (key, ID, type) for traceability

## Expected Workflow

### Quickstart Mode (Recommended)
Users will:
1. Configure Jira credentials once: `jira-beads-sync configure`
2. Fetch and sync issues: `jira-beads-sync quickstart PROJ-123`
3. The tool will recursively fetch all dependencies and sync to beads format
4. Validate that hierarchy and dependencies are correct
5. Work in beads, then sync changes back to Jira: `jira-beads-sync sync`

### Sync Mode (Bidirectional)
Users will:
1. Import issues from Jira using quickstart mode
2. Make changes in beads (update status, assignee, etc.)
3. Run sync to push changes back to Jira: `jira-beads-sync sync`
4. The tool will detect changes and update Jira via API
5. Maintain consistency between beads and Jira

### Convert Mode (One-Way Only)
Users will:
1. Export Jira tasks to JSON file
2. Run this tool to convert the Jira data structure
3. Import the converted issues into beads system
4. Validate that hierarchy and dependencies are correct
Note: Convert mode does not support syncing back to Jira

## Jira API Integration

The tool uses Jira REST API v2 for bidirectional synchronization:

### Fetching from Jira (Jira → Beads)
- **Authentication**: Basic Auth with username and API token
- **Configuration**: Supports config file, environment variables, or interactive setup
- **Recursive Fetching**: Walks dependency graph including:
  - Subtasks (via `fields.subtasks`)
  - Linked issues (via `fields.issuelinks`, both inward and outward)
  - Parent issues (via `fields.parent`, excluding epics)
- **Duplicate Prevention**: Uses visited map to avoid infinite loops

### Syncing to Jira (Beads → Jira)
- **Change Detection**: Compares beads state with cached Jira state
- **Field Updates**: Updates status, assignee, priority, description
- **API Operations**: Uses PUT/POST endpoints to update Jira issues
- **Conflict Resolution**: Handles concurrent modifications gracefully

Key files:
- `internal/jira/client.go`: Jira API client with recursive dependency walking and update operations
- `internal/config/config.go`: Configuration management

## Claude Code Plugin

This repository includes a Claude Code plugin that enables syncing Jira issues through natural language commands or slash commands.

### Plugin Commands

- `/import-jira <jira-url-or-key>` - Import a Jira issue and its dependency tree
- `/sync-jira` - Sync beads changes back to Jira
- `/configure-jira` - Configure Jira API credentials
- `/convert-jira-export <file>` - Convert a Jira export JSON file

### Natural Language Usage

Claude can understand requests like:
- "Import PROJ-123 from Jira"
- "Fetch the Jira issue TEAM-456 and all its dependencies"
- "Sync beads changes back to Jira"
- "Configure my Jira credentials"
- "Convert jira-export.json to beads format"

### Using in Your Project

Add to your project's `.claude/CLAUDE.md`:

```markdown
# Jira Integration

When working on Jira issues, sync them with beads:
- Import: Ask Claude to "import <jira-key> from Jira"
- View: Use `bd list` and `bd show <id>` to see imported issues
- Sync back: Ask Claude to "sync beads changes to Jira"
- The tool automatically fetches all dependencies and maintains bidirectional sync
```

### Plugin Installation

For users:
```bash
# Install the CLI tool first
go install github.com/conallob/jira-beads-sync/cmd/jira-beads-sync@latest

# Start Claude with plugin enabled
claude --plugin-dir /path/to/jira-beads-sync
```

For development:
```bash
# From this repository
claude --plugin-dir .
```

See [PLUGIN.md](PLUGIN.md) for complete plugin documentation.
