---
description: Convert a Jira export JSON file to beads format
argument-hint: <jira-export-file>
---

Convert a previously exported Jira JSON file to beads format. This is useful when you have a Jira export file from the Jira UI or API export, rather than fetching directly from the API.

## Usage

```bash
jira-to-beads convert <jira-export-file>
```

The command will:
1. Parse the Jira JSON export file
2. Convert issues to beads format
3. Generate YAML files in `.beads/` directory

## Jira Export Format

The tool expects a JSON file in Jira's export format:
```json
{
  "issues": [
    {
      "key": "PROJ-123",
      "fields": {
        "summary": "Issue title",
        "description": "Issue description",
        "issuetype": { "name": "Story" },
        "status": { "name": "In Progress", "statusCategory": { "key": "indeterminate" } },
        "priority": { "name": "High" },
        "created": "2024-01-01T10:00:00.000+0000",
        "updated": "2024-01-15T14:30:00.000+0000"
      }
    }
  ]
}
```

## Getting a Jira Export

### Option 1: Jira UI Export
1. Go to your Jira project
2. Click "Issues" → "Search for issues"
3. Use JQL to find issues (e.g., `project = PROJ AND status != Done`)
4. Click "Export" → "Export JSON"
5. Save the file

### Option 2: REST API Export
```bash
curl -u email@example.com:api-token \
  "https://jira.example.com/rest/api/2/search?jql=project=PROJ" \
  > jira-export.json
```

## Example Interaction

```
User: Convert my Jira export file jira-export.json to beads

Claude: I'll convert your Jira export file to beads format.

[Runs: jira-to-beads convert jira-export.json]

Converting jira-export.json to beads format...
✓ Conversion complete!
  3 epic(s) written to .beads/epics/
  15 issue(s) written to .beads/issues/

The issues have been converted and saved to .beads/ directory.

To see the imported issues, run: bd list
```

## Differences from import-jira

- **No API calls**: Works offline with exported files
- **No dependency walking**: Only converts issues in the export file (doesn't fetch linked issues)
- **No credentials needed**: Doesn't require Jira API access

For a complete import with dependency graph walking, use `import-jira` instead.

## Viewing Results

After conversion:
```bash
bd list              # List all imported issues
bd show <issue-id>   # Show details of a specific issue
```

## Notes

- The export file must be valid JSON in Jira's format
- Issue keys are used to generate beads issue IDs (e.g., PROJ-123 → proj-123)
- Jira metadata is preserved in the beads issue metadata
- Epic relationships are converted to beads epics
- Issue links in the export are converted to dependencies
