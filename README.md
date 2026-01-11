# jira-beads-sync

[![Test and Lint](https://github.com/conallob/jira-beads-sync/actions/workflows/test.yml/badge.svg)](https://github.com/conallob/jira-beads-sync/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/conallob/jira-beads-sync)](https://goreportcard.com/report/github.com/conallob/jira-beads-sync)
[![License](https://img.shields.io/github/license/conallob/jira-beads-sync)](LICENSE)
[![Release](https://img.shields.io/github/v/release/conallob/jira-beads-sync)](https://github.com/conallob/jira-beads-sync/releases/latest)

A Go-based CLI tool to convert Jira task trees into beads issues. This tool handles the hierarchical structure of Jira tasks (epics, stories, subtasks) and maps them to the beads issue tracking system while preserving dependencies and relationships.

## Features

- **Claude Code Plugin**: Import Jira issues through natural language with Claude Code
- **Quickstart Mode**: Fetch issues directly from Jira with a single command
- **Dependency Graph Walking**: Automatically fetch and convert entire task hierarchies
- **Protocol Buffers Architecture**: Uses protobuf as internal data format with YAML rendering layer
- **Hierarchical Mapping**: Converts Jira epics, stories, and subtasks to beads format
- **Dependency Preservation**: Maintains issue links and parent-child relationships
- **Type-Safe Conversion**: Strong typing through protobuf definitions
- **Multiple Platforms**: Binaries available for Linux, macOS, and Windows (x86_64 and ARM64)

## Installation

### Download Pre-built Binary

Download the latest release for your platform from the [releases page](https://github.com/conallob/jira-beads-sync/releases/latest).

**macOS (Homebrew):**
```bash
brew tap conallob/tap
brew install jira-beads-sync
```

**macOS (Manual):**
```bash
# For Apple Silicon (M1/M2/M3)
curl -LO https://github.com/conallob/jira-beads-sync/releases/latest/download/jira-beads-sync_Darwin_arm64.tar.gz
tar xzf jira-beads-sync_Darwin_arm64.tar.gz
sudo mv jira-beads-sync /usr/local/bin/

# For Intel
curl -LO https://github.com/conallob/jira-beads-sync/releases/latest/download/jira-beads-sync_Darwin_x86_64.tar.gz
tar xzf jira-beads-sync_Darwin_x86_64.tar.gz
sudo mv jira-beads-sync /usr/local/bin/
```

**Linux:**
```bash
# For x86_64
curl -LO https://github.com/conallob/jira-beads-sync/releases/latest/download/jira-beads-sync_Linux_x86_64.tar.gz
tar xzf jira-beads-sync_Linux_x86_64.tar.gz
sudo mv jira-beads-sync /usr/local/bin/

# For ARM64
curl -LO https://github.com/conallob/jira-beads-sync/releases/latest/download/jira-beads-sync_Linux_arm64.tar.gz
tar xzf jira-beads-sync_Linux_arm64.tar.gz
sudo mv jira-beads-sync /usr/local/bin/
```

**Docker:**
```bash
docker pull ghcr.io/conallob/jira-beads-sync:latest
docker run --rm -v $(pwd):/data ghcr.io/conallob/jira-beads-sync:latest convert /data/jira-export.json
```

### Install from Source

```bash
go install github.com/conallob/jira-beads-sync/cmd/jira-beads-sync@latest
```

Or build from source:

```bash
git clone https://github.com/conallob/jira-beads-sync.git
cd jira-beads-sync
make build
```

## Usage

### Quickstart Mode (Recommended)

Fetch issues directly from Jira and convert to beads format:

```bash
# Configure Jira credentials (one-time setup)
jira-beads-sync configure

# Fetch and convert a Jira issue with its entire dependency graph
jira-beads-sync quickstart https://jira.example.com/browse/PROJ-123

# Or use issue key directly (uses base URL from config)
jira-beads-sync quickstart PROJ-123
```

The quickstart command will:
1. Fetch the specified issue from Jira
2. Recursively walk the dependency graph (subtasks, linked issues, parents)
3. Convert all issues to beads format
4. Generate YAML files in `.beads/` directory

### Configuration

Jira credentials can be configured in three ways (in order of precedence):

1. **Interactive configuration:**
   ```bash
   jira-beads-sync configure
   ```

2. **Environment variables:**
   ```bash
   export JIRA_BASE_URL=https://jira.example.com
   export JIRA_USERNAME=your-email@example.com
   export JIRA_API_TOKEN=your-api-token
   ```

3. **Config file** at `~/.config/jira-beads-sync/config.yml`:
   ```yaml
   jira:
     base_url: https://jira.example.com
     username: your-email@example.com
     api_token: your-api-token
   ```

To generate a Jira API token, visit: https://id.atlassian.com/manage-profile/security/api-tokens

### Convert Mode

Convert a previously exported Jira JSON file:

```bash
# Convert a Jira export file to beads format
jira-beads-sync convert jira-export.json
```

### Other Commands

```bash
# Show version
jira-beads-sync version

# Show help
jira-beads-sync help
```

## Claude Code Plugin

This tool can be used as a Claude Code plugin to import Jira issues through natural language:

```bash
# Install and start Claude with plugin
claude --plugin-dir /path/to/jira-beads-sync
```

Then use natural language commands:
- "Import PROJ-123 from Jira"
- "Fetch the Jira issue TEAM-456 and all its dependencies"
- "Configure my Jira credentials"

See [PLUGIN.md](PLUGIN.md) for complete plugin documentation and usage examples.

## Development

This project uses a Makefile for common development tasks:

```bash
make help        # Show all available targets
make proto       # Generate protobuf files
make build       # Build the binary
make test        # Run tests with coverage
make lint        # Run linter
make fmt         # Format code
make verify      # Run all verification steps (fmt, lint, test)
```

### Requirements

- Go 1.21 or later
- Protocol Buffers compiler (`protoc`)
- `protoc-gen-go` plugin
- `golangci-lint` for linting

### Architecture

This tool uses Protocol Buffers as the internal data structure format:

```
JSON (Jira) → Protobuf (Jira) → Protobuf (Beads) → YAML (Beads)
     ↓              ↓                  ↓               ↓
  Adapter    Generated Types    Converter      Renderer
```

See [CLAUDE.md](CLAUDE.md) for detailed architecture documentation.

## Project Structure

- `cmd/jira-beads-sync/` - Main application entry point
- `internal/jira/` - Jira JSON to protobuf adapter
- `internal/beads/` - YAML rendering layer on top of protobuf
- `internal/converter/` - Conversion logic between Jira and beads protobuf
- `proto/` - Protocol Buffer definitions (source of truth for data structures)
- `gen/` - Generated Go code from protobuf definitions
- `.github/workflows/` - CI/CD workflows

## Releasing

Releases are automated via GitHub Actions and GoReleaser:

1. Create a new tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
2. Push the tag: `git push origin v1.0.0`
3. GitHub Actions will automatically:
   - Run tests
   - Build binaries for all platforms
   - Create GitHub release with artifacts
   - Build and push Docker images
   - Update Homebrew formula

To test the release process locally:
```bash
make release-dry-run  # Test without publishing
```

## License

See LICENSE file for details.
