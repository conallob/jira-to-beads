---
description: Import Jira issues directly from Jira API into beads
argument-hint: <jira-url-or-key>
---

Import a Jira issue and its entire dependency tree into beads. This command fetches the issue from Jira, walks the dependency graph, converts to beads format, and imports into the local beads repository.

## Usage

The user can provide either:
- A full Jira URL: `https://jira.example.com/browse/PROJ-123`
- An issue key: `PROJ-123` (uses base URL from configuration)

## Steps

1. **Check Configuration**: Verify Jira credentials are configured
   - Run `jira-beads-sync configure` if not configured
   - Configuration is stored at `~/.config/jira-beads-sync/config.yml`
   - Can also use environment variables: `JIRA_BASE_URL`, `JIRA_USERNAME`, `JIRA_API_TOKEN`

2. **Fetch and Convert**: Run the quickstart command
   ```bash
   jira-beads-sync quickstart <jira-url-or-key>
   ```

   This will:
   - Fetch the issue from Jira REST API
   - Recursively walk dependencies (subtasks, linked issues, parents)
   - Convert to beads format
   - Generate YAML files in `.beads/` directory

3. **Verify Import**: List the imported issues
   ```bash
   bd list
   ```

4. **Show Details**: Display issue details to confirm
   ```bash
   bd show <issue-id>
   ```

## What Gets Imported

The command recursively fetches:
- **Subtasks**: All subtasks of the main issue
- **Linked Issues**: Issues linked via "blocks", "depends on", etc.
- **Parent Issues**: Parent stories/tasks (but not epics, which become beads epics)
- **Epics**: Jira epics are converted to beads epics

## Mapping

- **Status**:
  - Jira "new" → beads "open"
  - Jira "in progress" → beads "in_progress"
  - Jira "blocked" → beads "blocked"
  - Jira "done" → beads "closed"

- **Priority**:
  - Jira "critical/blocker" → beads "p0"
  - Jira "highest" → beads "p1"
  - Jira "high" → beads "p2"
  - Jira "medium" → beads "p2"
  - Jira "low" → beads "p3"
  - Jira "lowest" → beads "p4"

- **Dependencies**: Jira issue links are converted to beads `dependsOn` relationships

## Error Handling

If the command fails:
- **Authentication Error**: Check credentials with `jira-beads-sync configure`
- **Issue Not Found**: Verify the issue key or URL is correct
- **Network Error**: Check connectivity to Jira server
- **Permission Error**: Ensure API token has read access to the issue

## Example Interaction

```
User: Import PROJ-123 from Jira

Claude: I'll import PROJ-123 and its dependencies from Jira.
[Runs: jira-beads-sync quickstart PROJ-123]

✓ Fetched 5 issue(s)
✓ Conversion complete!

Issues imported:
- proj-123: "Implement new authentication system"
- proj-124: "Add OAuth2 support" (subtask)
- proj-125: "Update user model" (subtask)
- proj-126: "Add security tests" (depends on PROJ-123)
- proj-120: "Security infrastructure" (epic)

You can view them with: bd list
```

## Integration with Beads

After import, the issues are available in the beads system:
- Use `bd show <id>` to view issue details
- Use `bd update <id>` to change status or priority
- Use `bd dep add <issue> <dependency>` to add more dependencies
- Dependencies from Jira are automatically preserved

## Notes

- Jira metadata (key, ID, type) is preserved in the beads issue metadata
- The import is idempotent - running it again won't create duplicates (based on Jira key)
- Epic relationships are preserved through the `epic` field in beads issues