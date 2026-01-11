# Quick Start: Using jira-beads-sync with Claude Code

This guide helps you get started using the jira-beads-sync Claude Code plugin.

## Installation

1. **Install the CLI tool:**
   ```bash
   go install github.com/conallob/jira-beads-sync/cmd/jira-beads-sync@latest
   ```

2. **Clone the repository:**
   ```bash
   git clone https://github.com/conallob/jira-beads-sync.git
   ```

3. **Start Claude Code with the plugin:**
   ```bash
   cd your-project
   claude --plugin-dir ~/path/to/jira-beads-sync
   ```

## First-Time Setup

Configure your Jira credentials:

```
You: Configure my Jira credentials

Claude will run: jira-beads-sync configure
[Follow prompts to enter Jira URL, email, and API token]
```

Get a Jira API token from: https://id.atlassian.com/manage-profile/security/api-tokens

## Using the Plugin

### Import a Jira Issue

Simply ask Claude in natural language:

```
You: Import PROJ-123 from Jira

Claude: I'll import PROJ-123 and its dependencies from Jira.
[Runs: jira-beads-sync quickstart PROJ-123]

✓ Fetched 5 issue(s)
✓ Conversion complete!

Issues imported:
- proj-123: "Implement new authentication system"
- proj-124: "Add OAuth2 support" (subtask)
- proj-125: "Update user model" (subtask)

You can view them with: bd list
```

### Other Examples

```
You: Fetch the Jira issue TEAM-456 and all its dependencies
You: Pull AUTH-789 from Jira into beads
You: Import https://jira.example.com/browse/PROJ-100
```

### Using Slash Commands

You can also use explicit slash commands:

```
/import-jira PROJ-123
/import-jira https://jira.example.com/browse/PROJ-123
/configure-jira
/convert-jira-export jira-export.json
```

## Integration with Your Project

Add to your project's `.claude/CLAUDE.md`:

```markdown
# Jira Integration

This project uses Jira for issue tracking. To work on a Jira issue:

1. Import the issue: Ask Claude to "import <jira-key> from Jira"
2. View imported issues: `bd list`
3. Show details: `bd show <issue-id>`

The tool automatically fetches all dependencies and related issues.

## Common Workflows

**Starting a new feature:**
```
User: I'm working on PROJ-123
Claude: [Imports PROJ-123 and shows dependencies]
```

**Planning work:**
```
User: Import the authentication epic AUTH-100
Claude: [Imports epic and all related stories/tasks]
User: What should I work on first?
Claude: [Analyzes dependencies and suggests starting point]
```
```

## What Gets Imported

When you import a Jira issue, the plugin recursively fetches:

- **The main issue**
- **All subtasks** of the issue
- **All linked issues** (blocks, depends on, relates to)
- **Parent issues** (stories, tasks)
- **Epic information** (converted to beads epics)

## Working with Imported Issues

After import, use beads commands:

```bash
bd list                          # List all issues
bd show proj-123                 # Show issue details
bd update proj-123 --status in_progress
bd dep add proj-124 proj-125    # Add dependency
bd close proj-123               # Close issue
```

Or ask Claude:

```
You: Show me all open issues
You: What are the dependencies for proj-123?
You: Mark proj-124 as in progress
```

## Status and Priority Mapping

The plugin automatically maps Jira values to beads:

### Status
- Jira "To Do" / "New" → beads "open"
- Jira "In Progress" → beads "in_progress"
- Jira "Blocked" → beads "blocked"
- Jira "Done" / "Closed" → beads "closed"

### Priority
- Jira "Blocker" / "Critical" → beads "p0" (critical)
- Jira "Highest" → beads "p1"
- Jira "High" → beads "p2"
- Jira "Medium" → beads "p2"
- Jira "Low" → beads "p3"
- Jira "Lowest" → beads "p4"

## Troubleshooting

### "Command not found: jira-beads-sync"

Install the CLI tool:
```bash
go install github.com/conallob/jira-beads-sync/cmd/jira-beads-sync@latest
```

### "Invalid configuration"

Configure Jira credentials:
```
You: Configure Jira credentials
```

### "Authentication failed"

- Check your API token is correct
- Verify the token hasn't expired
- Get a new token from: https://id.atlassian.com/manage-profile/security/api-tokens

### Plugin not loaded

Make sure you're starting Claude with the plugin directory:
```bash
claude --plugin-dir /path/to/jira-beads-sync
```

## Advanced Usage

### Environment Variables

Instead of running configure, you can set environment variables:

```bash
export JIRA_BASE_URL=https://jira.example.com
export JIRA_USERNAME=your-email@example.com
export JIRA_API_TOKEN=your-api-token

claude --plugin-dir /path/to/jira-beads-sync
```

### Converting Export Files

If you have a Jira export JSON file:

```
You: Convert jira-export.json to beads format

Claude: I'll convert your Jira export file.
[Runs: jira-beads-sync convert jira-export.json]
```

### Project-Specific Plugin

To always use the plugin in a specific project, add to `.claude/settings.json`:

```json
{
  "pluginDirs": ["/path/to/jira-beads-sync"]
}
```

## Next Steps

- See [PLUGIN.md](PLUGIN.md) for complete plugin documentation
- See [README.md](README.md) for CLI usage
- See [CLAUDE.md](CLAUDE.md) for development guide

## Support

For issues or questions:
- GitHub Issues: https://github.com/conallob/jira-beads-sync/issues
- Repository: https://github.com/conallob/jira-beads-sync
