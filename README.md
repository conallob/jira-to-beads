# jira-to-beads

[![Test and Lint](https://github.com/conallob/jira-to-beads/actions/workflows/test.yml/badge.svg)](https://github.com/conallob/jira-to-beads/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/conallob/jira-to-beads)](https://goreportcard.com/report/github.com/conallob/jira-to-beads)
[![License](https://img.shields.io/github/license/conallob/jira-to-beads)](LICENSE)
[![Release](https://img.shields.io/github/v/release/conallob/jira-to-beads)](https://github.com/conallob/jira-to-beads/releases/latest)

A Go-based CLI tool to convert Jira task trees into beads issues. This tool handles the hierarchical structure of Jira tasks (epics, stories, subtasks) and maps them to the beads issue tracking system while preserving dependencies and relationships.

## Features

- **Protocol Buffers Architecture**: Uses protobuf as internal data format with YAML rendering layer
- **Hierarchical Mapping**: Converts Jira epics, stories, and subtasks to beads format
- **Dependency Preservation**: Maintains issue links and parent-child relationships
- **Type-Safe Conversion**: Strong typing through protobuf definitions
- **Multiple Platforms**: Binaries available for Linux, macOS, and Windows (x86_64 and ARM64)

## Installation

### Download Pre-built Binary

Download the latest release for your platform from the [releases page](https://github.com/conallob/jira-to-beads/releases/latest).

**macOS (Homebrew):**
```bash
brew tap conallob/tap
brew install jira-to-beads
```

**macOS (Manual):**
```bash
# For Apple Silicon (M1/M2/M3)
curl -LO https://github.com/conallob/jira-to-beads/releases/latest/download/jira-to-beads_Darwin_arm64.tar.gz
tar xzf jira-to-beads_Darwin_arm64.tar.gz
sudo mv jira-to-beads /usr/local/bin/

# For Intel
curl -LO https://github.com/conallob/jira-to-beads/releases/latest/download/jira-to-beads_Darwin_x86_64.tar.gz
tar xzf jira-to-beads_Darwin_x86_64.tar.gz
sudo mv jira-to-beads /usr/local/bin/
```

**Linux:**
```bash
# For x86_64
curl -LO https://github.com/conallob/jira-to-beads/releases/latest/download/jira-to-beads_Linux_x86_64.tar.gz
tar xzf jira-to-beads_Linux_x86_64.tar.gz
sudo mv jira-to-beads /usr/local/bin/

# For ARM64
curl -LO https://github.com/conallob/jira-to-beads/releases/latest/download/jira-to-beads_Linux_arm64.tar.gz
tar xzf jira-to-beads_Linux_arm64.tar.gz
sudo mv jira-to-beads /usr/local/bin/
```

**Docker:**
```bash
docker pull ghcr.io/conallob/jira-to-beads:latest
docker run --rm -v $(pwd):/data ghcr.io/conallob/jira-to-beads:latest convert /data/jira-export.json
```

### Install from Source

```bash
go install github.com/conallob/jira-to-beads/cmd/jira-to-beads@latest
```

Or build from source:

```bash
git clone https://github.com/conallob/jira-to-beads.git
cd jira-to-beads
make build
```

## Usage

```bash
# Convert a Jira export file to beads format
jira-to-beads convert jira-export.json

# Show version
jira-to-beads version

# Show help
jira-to-beads help
```

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

- `cmd/jira-to-beads/` - Main application entry point
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
