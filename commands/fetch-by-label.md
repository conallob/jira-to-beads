# Fetch Jira Issues by Label

Fetch all Jira issues with a specific label and their dependencies, then convert to beads format.

## Usage

This command is invoked when the user wants to import multiple Jira issues that share a common label. This is useful for importing sprints, epics, or any group of related issues.

## Command Patterns

The command should be invoked when users say things like:
- "Import all issues with label <label-name> from Jira"
- "Fetch Jira issues labeled <label-name>"
- "Get all issues tagged with <label-name> from Jira"
- "Import sprint <sprint-label>"
- "Fetch all issues for label <label-name>"

## CLI Invocation

```bash
jira-beads-sync fetch-by-label <label-name>
```

## Examples

### Example 1: Import Sprint
```
User: Import all issues with label sprint-23 from Jira

Claude: I'll fetch all issues labeled "sprint-23" from Jira and their dependencies.

[Runs: jira-beads-sync fetch-by-label sprint-23]

Searching for issues with label: sprint-23
Found 12 issue(s) with label sprint-23

Fetching PROJ-100...
Fetching PROJ-101...
Fetching PROJ-102...
...
✓ Fetched 25 issue(s) total (including dependencies)

Converting to beads format...
✓ Conversion complete!
  2 epic(s) written to .beads/epics/
  23 issue(s) written to .beads/issues/
```

### Example 2: Import Team Label
```
User: Fetch all issues tagged with team-frontend from Jira

Claude: I'll fetch all issues with the "team-frontend" label.

[Runs: jira-beads-sync fetch-by-label team-frontend]

Found 8 issues with label team-frontend
✓ Imported 15 issues total (including dependencies)
```

### Example 3: Import Release Label
```
User: Get all issues for label v2.0-release

Claude: Importing all issues labeled "v2.0-release"...

[Runs: jira-beads-sync fetch-by-label v2.0-release]

✓ Imported 42 issues
All issues are now available in .beads/issues/
```

## What Gets Fetched

1. **All issues with the specified label** - Uses Jira JQL search
2. **All subtasks** of labeled issues
3. **All linked issues** (blocks, depends on, relates to)
4. **Parent issues** (stories, tasks)
5. **Epic information** (converted to beads epics)

The command recursively walks the entire dependency graph, just like `quickstart`, but starts from multiple root issues instead of one.

## Use Cases

- **Sprint Import**: Import all issues for a sprint using sprint labels
- **Team Work**: Import all issues assigned to a specific team
- **Release Planning**: Import all issues for a release milestone
- **Project Area**: Import issues for a specific component or area
- **Priority Groups**: Import all critical or high-priority issues

## Configuration

Requires Jira configuration (same as quickstart):
- Jira base URL
- Username (email)
- API token

Run `jira-beads-sync configure` if not already set up.

## Limitations

- Maximum 1000 issues per label (Jira API pagination limit)
- Label must match exactly (case-sensitive)
- Requires appropriate Jira permissions to view labeled issues
