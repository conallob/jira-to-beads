# jira-beads-sync

[![Test and Lint](https://github.com/conallob/jira-beads-sync/actions/workflows/test.yml/badge.svg)](https://github.com/conallob/jira-beads-sync/actions/workflows/test.yml)
[![License](https://img.shields.io/github/license/conallob/jira-beads-sync)](LICENSE)
[![codecov](https://codecov.io/gh/conallob/jira-beads-sync/branch/main/graph/badge.svg)](https://codecov.io/gh/conallob/jira-beads-sync)
[![Release](https://img.shields.io/github/v/release/conallob/jira-beads-sync)](https://github.com/conallob/jira-beads-sync/releases/latest)
[![BuyMeACoffee](https://raw.githubusercontent.com/pachadotdev/buymeacoffee-badges/main/bmc-yellow.svg)](https://www.buymeacoffee.com/conallob)

> Bridge Jira and beads: Import Jira issues locally, work with git-backed issue tracking, sync changes back to Jira.

**jira-beads-sync** synchronizes Jira issues with [beads](https://github.com/steveyegge/beads), a git-backed issue tracker. Work with Jira issues as YAML files in your repository, manage them with beads commands or Claude Code, then sync your changes back to Jira.

Perfect for developers who want to:
- Track Jira issues alongside code in version control
- Work with issues offline using beads
- Use natural language with Claude Code to manage issues
- Maintain bidirectional sync between Jira and local git repos

## Quick Links

📚 **New to jira-beads-sync?** → [Getting Started Guide](GETTING_STARTED.md)

🎯 **Use Cases:**
- **CLI User?** → [CLI Guide](docs/CLI_GUIDE.md) - Complete command reference
- **Claude Code User?** → [Plugin Guide](docs/PLUGIN_GUIDE.md) - Natural language workflows
- **Opencode User?** → [Opencode Guide](docs/OPENCODE_GUIDE.md) - Opencode integration
- **Need Examples?** → [Real-World Examples](docs/EXAMPLES.md) - Practical scenarios

👩‍💻 **For Developers:**
- [CLAUDE.md](CLAUDE.md) - Architecture and development guide
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines

## What Does It Do?

```bash
# Import a Jira issue with all its dependencies
jira-beads-sync quickstart PROJ-123

# Work locally with beads
bd list
bd show proj-123
bd update proj-123 --status in_progress

# Sync changes back to Jira
jira-beads-sync sync
```

Or use natural language with Claude Code:

```
You: Import PROJ-123 from Jira

Claude: [Imports issue and dependencies]
        ✓ Fetched 5 issue(s)
        Ready to work!
```

## Key Features

- ✅ **Bidirectional Sync**: Import from Jira ↔ Work locally ↔ Push back to Jira
- ✅ **Automatic Dependencies**: Recursively fetches epics, stories, subtasks, and links
- ✅ **Two Interfaces**: CLI commands or natural language with Claude Code
- ✅ **Preserves Structure**: Maintains hierarchies, relationships, and metadata
- ✅ **Git-Backed**: Issues stored as YAML files in `.beads/` directory
- ✅ **Type-Safe**: Built on Protocol Buffers for reliability

## Installation

### Homebrew (Recommended for macOS/Linux)

```bash
# Direct install (recommended)
brew install conallob/tap/jira-beads-sync

# Or tap first, then install
brew tap conallob/tap
brew install jira-beads-sync
```

### From Binary

Download from the [releases page](https://github.com/conallob/jira-beads-sync/releases/latest) or:

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/conallob/jira-beads-sync/releases/latest/download/jira-beads-sync_Darwin_arm64.tar.gz
tar xzf jira-beads-sync_Darwin_arm64.tar.gz
sudo mv jira-beads-sync /usr/local/bin/

# Linux (x86_64)
curl -LO https://github.com/conallob/jira-beads-sync/releases/latest/download/jira-beads-sync_Linux_x86_64.tar.gz
tar xzf jira-beads-sync_Linux_x86_64.tar.gz
sudo mv jira-beads-sync /usr/local/bin/
```

### From Source

```bash
go install github.com/conallob/jira-beads-sync/cmd/jira-beads-sync@latest
```

**More options:** See [Getting Started Guide](GETTING_STARTED.md#installation) for all installation methods including Docker, DEB, and RPM packages.

## Quick Start

### 1. Configure Credentials

```bash
jira-beads-sync configure
```

You'll need a Jira API token from https://id.atlassian.com/manage-profile/security/api-tokens

### 2. Import a Jira Issue

```bash
jira-beads-sync quickstart PROJ-123
```

This fetches the issue and all its dependencies (subtasks, linked issues, parent issues).

### 3. Work Locally

```bash
bd list                              # List all issues
bd show proj-123                     # Show details
bd update proj-123 --status in_progress
```

### 4. Sync Back to Jira

```bash
jira-beads-sync sync
```

**📖 Detailed Usage:** See [CLI Guide](docs/CLI_GUIDE.md) for all commands and options.

## Using with AI Coding Tools

jira-beads-sync is a standard CLI binary that works with any AI coding assistant. Install it and configure Jira credentials once, then ask your AI tool to run commands on your behalf.

### Claude Code

Enable the built-in plugin for natural language issue management:

```bash
claude --plugin-dir /path/to/jira-beads-sync
```

Then simply ask Claude:

```
You: Import PROJ-123 from Jira
You: Show me all open issues
You: Mark PROJ-124 as in progress
You: Sync changes back to Jira
```

**📖 Plugin Guide:** See [Plugin Guide](docs/PLUGIN_GUIDE.md) for complete plugin documentation.

### Opencode

Add project instructions so Opencode knows how to invoke the CLI, then use natural language:

```
You: Import PROJ-123 from Jira
You: Fetch all issues for sprint-47
You: Sync my changes back to Jira
```

**📖 Opencode Guide:** See [Opencode Guide](docs/OPENCODE_GUIDE.md) for setup and workflow examples.

### Other AI Tools

Any AI coding tool that can run shell commands can use jira-beads-sync. Add these commands to your project's instructions file:

```
Import a Jira issue:   jira-beads-sync quickstart <JIRA-KEY>
Fetch by label:        jira-beads-sync fetch-by-label <label>
Fetch by JQL:          jira-beads-sync fetch-jql '<jql>'
List issues:           bd list
Sync back to Jira:     jira-beads-sync sync
```

## How It Works

This tool uses Protocol Buffers internally for type-safe data handling:

```
Jira (JSON) → Protobuf → beads (YAML) → Protobuf → Jira (JSON)
     ↓           ↓             ↓           ↓           ↓
  Fetch     Convert      Render      Detect      Update
```

**Learn more:** See [CLAUDE.md](CLAUDE.md) for detailed architecture and [CONTRIBUTING.md](CONTRIBUTING.md) for development setup.

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Development setup
- Coding standards
- Testing guidelines
- Pull request process

## Resources

- 📚 [Getting Started Guide](GETTING_STARTED.md) - New user walkthrough
- 🖥️ [CLI Guide](docs/CLI_GUIDE.md) - Complete CLI reference
- 🤖 [Plugin Guide](docs/PLUGIN_GUIDE.md) - Claude Code plugin usage
- 💡 [Examples](docs/EXAMPLES.md) - Real-world scenarios
- 🏗️ [CLAUDE.md](CLAUDE.md) - Architecture for developers
- 🤝 [Contributing](CONTRIBUTING.md) - Development guidelines

## Support

- **Issues:** https://github.com/conallob/jira-beads-sync/issues
- **Discussions:** https://github.com/conallob/jira-beads-sync/discussions
- **Sponsor:** [![BuyMeACoffee](https://raw.githubusercontent.com/pachadotdev/buymeacoffee-badges/main/bmc-yellow.svg)](https://www.buymeacoffee.com/conallob)

## License

BSD-3-Clause - See [LICENSE](LICENSE) file for details.
