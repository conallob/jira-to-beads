# jira-to-beads Claude Code Plugin

This directory contains a Claude Code plugin that enables Claude to import Jira issues directly into beads issue tracker through natural language commands.

## Installation

### From Local Directory

To use this plugin in development or from a local clone:

```bash
# Clone the repository
git clone https://github.com/conallob/jira-to-beads.git
cd jira-to-beads

# Build the CLI tool
make build
sudo mv jira-to-beads /usr/local/bin/

# Start Claude Code with plugin enabled
claude --plugin-dir /path/to/jira-to-beads
```

### From Plugin Marketplace

Once published to a marketplace:

```bash
# Install the plugin
claude plugin install jira-to-beads

# Or install from a specific marketplace
claude plugin install jira-to-beads@your-marketplace
```

## Available Commands

The plugin provides three slash commands for Claude Code:

### `/import-jira <jira-url-or-key>`

Import a Jira issue and its entire dependency tree into beads.

**Examples:**
```
/import-jira https://jira.example.com/browse/PROJ-123
/import-jira PROJ-123
```

**Natural Language:**
```
Import PROJ-123 from Jira
Fetch the Jira issue TEAM-456 and all its dependencies
Pull PROJ-789 from Jira into beads
```

### `/configure-jira`

Configure Jira API credentials for importing issues.

**Examples:**
```
/configure-jira
```

**Natural Language:**
```
Configure Jira credentials
Set up my Jira API token
Help me connect to Jira
```

### `/convert-jira-export <file>`

Convert a Jira export JSON file to beads format.

**Examples:**
```
/convert-jira-export jira-export.json
```

**Natural Language:**
```
Convert my Jira export file to beads
Import issues from jira-export.json
```

## Usage in CLAUDE.md

You can add instructions to your project's `.claude/CLAUDE.md` file to make Claude automatically aware of Jira integration:

```markdown
# Jira Integration

This project uses Jira for issue tracking. Issues can be imported into beads for local development:

## Importing Jira Issues

When starting work on a Jira issue:
1. Import the issue: `/import-jira <jira-key>`
2. View imported issues: `bd list`
3. Show details: `bd show <issue-id>`

## Jira Issue Mapping

- Jira epics → beads epics
- Jira stories/tasks → beads issues
- Jira subtasks → beads issues with dependencies
- Issue links → beads dependencies

## Example Workflow

```
User: I'm working on PROJ-123
Claude: [Imports PROJ-123 and shows dependencies]
User: Show me what needs to be done first
Claude: [Lists dependencies and suggests starting point]
```
```

## Direct Prompts

Claude will understand natural language requests without explicit slash commands:

```
User: Import PROJ-123 from Jira
Claude: I'll fetch PROJ-123 and its dependencies from Jira...
```

```
User: I need to work on the authentication epic from Jira
Claude: Which Jira issue key would you like me to import? Or can you provide the Jira URL?
User: AUTH-100
Claude: Importing AUTH-100 and walking the dependency tree...
```

## Features

### Automatic Dependency Walking

The plugin automatically fetches:
- **Subtasks**: All subtasks of the main issue
- **Linked Issues**: Issues linked via "blocks", "depends on", "relates to"
- **Parent Issues**: Parent stories/tasks (epics become beads epics)
- **Transitive Dependencies**: Recursively walks the entire graph

### Status and Priority Mapping

- **Status**: Jira status categories → beads status (open, in_progress, blocked, closed)
- **Priority**: Jira priorities → beads p0-p4 scale
- **Metadata**: Preserves Jira key, ID, and type for traceability

### Integration with Beads

After import, issues are fully integrated with beads:
- View: `bd list`, `bd show <id>`
- Update: `bd update <id> --status in_progress`
- Dependencies: `bd dep add <issue> <dependency>`
- Close: `bd close <id>`

## Requirements

- **jira-to-beads CLI**: Must be installed and in PATH
- **beads plugin**: Required for beads integration
- **Jira credentials**: Configured via `/configure-jira` or environment variables

## Configuration

### First-Time Setup

1. Install the CLI tool:
   ```bash
   go install github.com/conallob/jira-to-beads/cmd/jira-to-beads@latest
   ```

2. Configure Jira credentials:
   ```
   Claude: /configure-jira
   [Follow prompts to enter Jira URL, email, and API token]
   ```

3. Get a Jira API token:
   - Visit: https://id.atlassian.com/manage-profile/security/api-tokens
   - Create a new token
   - Copy and use in configuration

### Environment Variables

Alternatively, set environment variables:
```bash
export JIRA_BASE_URL=https://jira.example.com
export JIRA_USERNAME=your-email@example.com
export JIRA_API_TOKEN=your-api-token
```

## Plugin Structure

```
jira-to-beads/
├── .claude-plugin/
│   └── plugin.json          # Plugin metadata
├── commands/
│   ├── import-jira.md       # Import from Jira API
│   ├── configure-jira.md    # Configure credentials
│   └── convert-jira-export.md  # Convert export files
├── cmd/jira-to-beads/       # CLI implementation
└── internal/                # Core functionality
```

## Troubleshooting

### "Command not found: jira-to-beads"

The CLI tool is not in your PATH. Install it:
```bash
go install github.com/conallob/jira-to-beads/cmd/jira-to-beads@latest
```

Or build from source:
```bash
git clone https://github.com/conallob/jira-to-beads.git
cd jira-to-beads
make build
sudo mv jira-to-beads /usr/local/bin/
```

### "Invalid configuration"

Jira credentials are not configured. Run:
```
/configure-jira
```

Or set environment variables.

### "Authentication failed"

- Verify your API token is correct
- Check that the token hasn't expired
- Ensure your Jira account has access to the issues

### "Issue not found"

- Verify the issue key is correct (case-sensitive)
- Check that you have read permissions on the issue
- Ensure the issue exists in Jira

## Development

To modify the plugin:

1. Edit command files in `commands/` directory
2. Update plugin metadata in `.claude-plugin/plugin.json`
3. Test with: `claude --plugin-dir .`

## License

Apache 2.0 - See LICENSE file for details
