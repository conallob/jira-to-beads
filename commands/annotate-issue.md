# Annotate Issue with Repository

Add repository information to a beads issue to support polyrepo tracking.

## Usage

This command is invoked when the user wants to associate a Jira issue with a specific git repository. This is useful in polyrepo setups where different issues are worked on in different repositories.

## Command Patterns

The command should be invoked when users say things like:
- "Annotate issue <issue-id> with repository <repo>"
- "Add repository <repo> to issue <issue-id>"
- "Associate <issue-id> with repository <repo>"
- "Mark <issue-id> as belonging to <repo> repository"
- "Link issue <issue-id> to repository <repo>"

## CLI Invocation

```bash
jira-beads-sync annotate <issue-id> <repository>
```

## Examples

### Example 1: Add GitHub Repository
```
User: Annotate issue proj-123 with repository https://github.com/org/frontend

Claude: I'll add the repository to proj-123's metadata.

[Runs: jira-beads-sync annotate proj-123 https://github.com/org/frontend]

✓ Added repository 'https://github.com/org/frontend' to issue proj-123
  Updated: .beads/issues/proj-123.yaml
```

### Example 2: Simple Repository Name
```
User: Add repository backend-api to issue proj-456

Claude: Adding repository annotation...

[Runs: jira-beads-sync annotate proj-456 backend-api]

✓ Added repository 'backend-api' to issue proj-456
```

### Example 3: Multiple Repositories for One Issue
```
User: Annotate proj-789 with repository https://github.com/org/shared-lib

Claude: [Runs: jira-beads-sync annotate proj-789 https://github.com/org/shared-lib]

✓ Added repository 'https://github.com/org/shared-lib' to issue proj-789

Note: This issue now has 2 repositories associated:
- https://github.com/org/frontend
- https://github.com/org/shared-lib
```

### Example 4: Workflow - Import and Annotate
```
User: Import PROJ-100 from Jira and annotate it with our mobile repo

Claude: I'll import PROJ-100 first, then annotate it.

[Runs: jira-beads-sync quickstart PROJ-100]
✓ Imported PROJ-100

[Runs: jira-beads-sync annotate proj-100 https://github.com/org/mobile-app]
✓ Annotated proj-100 with repository

Done! PROJ-100 is now associated with the mobile-app repository.
```

## What Gets Annotated

The repository information is added to the issue's metadata under the `repositories` field:

```yaml
# .beads/issues/proj-123.yaml
id: proj-123
title: Implement user authentication
status: in_progress
priority: p1
metadata:
  jiraKey: PROJ-123
  jiraId: "10001"
  jiraIssueType: Story
  repositories:
    - https://github.com/org/frontend
    - https://github.com/org/auth-service
```

## Use Cases

### Polyrepo Tracking
Track which repository(ies) an issue is being worked on:
```bash
# Frontend issues
jira-beads-sync annotate proj-123 https://github.com/org/frontend
jira-beads-sync annotate proj-124 https://github.com/org/frontend

# Backend issues
jira-beads-sync annotate proj-125 https://github.com/org/backend
jira-beads-sync annotate proj-126 https://github.com/org/backend

# Cross-repo issues
jira-beads-sync annotate proj-127 https://github.com/org/frontend
jira-beads-sync annotate proj-127 https://github.com/org/backend
```

### Repository Filtering
Later, you can filter issues by repository:
```bash
# Find all issues for a specific repo
bd list --format json | jq '.[] | select(.metadata.repositories[] | contains("frontend"))'

# Count issues per repository
bd list --format json | jq -r '.[] | .metadata.repositories[]?' | sort | uniq -c
```

### Team Organization
In a polyrepo setup, different teams own different repositories:
```
Team Frontend: github.com/org/web-app
Team Backend: github.com/org/api-server
Team Mobile: github.com/org/mobile-app
Team Platform: github.com/org/platform-services
```

Annotate issues with their respective repositories to:
- Track which team is responsible
- Filter issues by repository/team
- Generate team-specific reports
- Plan work across repositories

## Repository Format

The repository can be:
- **Full URL**: `https://github.com/org/repo`
- **SSH URL**: `git@github.com:org/repo.git`
- **Short name**: `frontend-app`, `backend-api`
- **GitLab/Bitbucket**: Any git repository URL

The tool doesn't validate the repository format - it stores whatever you provide. This allows flexibility for different git hosting platforms and naming conventions.

## Multiple Repositories

An issue can be associated with multiple repositories:
```bash
# Initial annotation
jira-beads-sync annotate proj-100 https://github.com/org/frontend

# Add second repository
jira-beads-sync annotate proj-100 https://github.com/org/shared-lib

# Add third repository
jira-beads-sync annotate proj-100 https://github.com/org/docs
```

Each repository is added to the issue's `repositories` array. Duplicate repositories are prevented.

## Error Handling

### Issue Not Found
```
Error: issue file not found: .beads/issues/proj-999.yaml

Make sure the issue exists (run quickstart to import from Jira first)
```

### Duplicate Repository
```
Error: repository 'https://github.com/org/frontend' is already associated with issue proj-123

The repository is already linked to this issue.
```

## Integration with Git

While the tool doesn't enforce any git operations, you can integrate annotations with your git workflow:

```bash
#!/bin/bash
# Script: annotate-current-issue.sh
# Annotates an issue with the current git repository

ISSUE_ID=$1
REPO_URL=$(git remote get-url origin)

jira-beads-sync annotate "$ISSUE_ID" "$REPO_URL"
```

## Future Enhancements

Potential future features for repository annotations:
- Sync repository information back to Jira custom fields
- Auto-detect repository from git context
- Generate cross-repo dependency reports
- Filter/query issues by repository in beads CLI

## See Also

- `fetch-by-label` - Import multiple issues that can then be annotated
- `quickstart` - Import single issues before annotating
- beads issue format documentation
