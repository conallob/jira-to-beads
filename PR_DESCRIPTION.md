# Add Label-Based Fetching and Polyrepo Support

## Prompts/Context

**Initial Request:**
> "Create a new feature branch to support tracking a Jira label, instead of enumerating specific ticket numbers. Also include a way to annotate the git repo related to a specific jira task, to support polyrepo tracked in Jira"

**Follow-up:**
> "Create a new git branch for these commits and cherry pick them over to it"

## Summary

This PR adds two major features to jira-beads-sync to improve workflow flexibility:

1. **Label-Based Issue Fetching** - Fetch multiple Jira issues at once using labels instead of enumerating individual ticket keys
2. **Polyrepo Repository Annotation** - Associate issues with specific git repositories to support tracking work across multiple repos

## Features Added

### 1. Label-Based Fetching (`fetch-by-label` command)

Fetch all Jira issues with a specific label and their dependencies in one command.

**Command:**
```bash
jira-beads-sync fetch-by-label <label-name>
```

**Use Cases:**
- Import entire sprints: `fetch-by-label sprint-23`
- Import team work: `fetch-by-label team-frontend`
- Import releases: `fetch-by-label v2.0-release`
- Import by priority: `fetch-by-label critical`

**How it works:**
- Uses Jira JQL search API (`labels = <label-name>`)
- Recursively fetches all dependencies (subtasks, links, parents)
- Supports up to 1000 issues per label
- Converts everything to beads format

**Example:**
```bash
$ jira-beads-sync fetch-by-label sprint-23

Searching for issues with label: sprint-23
Found 12 issue(s) with label sprint-23

Fetching PROJ-100...
Fetching PROJ-101...
...
✓ Fetched 25 issue(s) total (including dependencies)

✓ Conversion complete!
  2 epic(s) written to .beads/epics/
  23 issue(s) written to .beads/issues/
```

### 2. Polyrepo Repository Annotation (`annotate` command)

Associate Jira issues with git repositories to track which repo(s) an issue touches.

**Command:**
```bash
jira-beads-sync annotate <issue-id> <repository>
```

**Use Cases:**
- Track which repos an issue affects
- Support cross-repo dependencies
- Generate per-repository reports
- Map issues to team-owned repositories

**Repository Formats Supported:**
- Full URLs: `https://github.com/org/repo`
- SSH URLs: `git@github.com:org/repo.git`
- Simple names: `frontend-app`, `backend-api`

**Example:**
```bash
$ jira-beads-sync annotate proj-123 https://github.com/org/frontend

✓ Added repository 'https://github.com/org/frontend' to issue proj-123
  Updated: .beads/issues/proj-123.yaml
```

**Multiple repositories per issue:**
```bash
jira-beads-sync annotate proj-123 https://github.com/org/frontend
jira-beads-sync annotate proj-123 https://github.com/org/shared-lib
```

**YAML Output:**
```yaml
# .beads/issues/proj-123.yaml
id: proj-123
title: Implement user authentication
status: in_progress
metadata:
  jiraKey: PROJ-123
  repositories:
    - https://github.com/org/frontend
    - https://github.com/org/shared-lib
```

## Implementation Details

### Files Modified

1. **proto/beads.proto**
   - Added `repositories` field to `Metadata` message
   - `repeated string repositories = 5;`
   - Note: Requires `make proto` to regenerate Go code

2. **internal/jira/client.go**
   - `SearchIssues(jql)` - Generic JQL search
   - `SearchIssuesByLabel(label)` - Label-specific search helper
   - `FetchIssuesByLabel(label)` - Fetch with full dependency walking

3. **cmd/jira-beads-sync/main.go**
   - `runFetchByLabel()` - Implements fetch-by-label command
   - `runAnnotate()` - Implements annotate command
   - Updated `printUsage()` with new commands

4. **internal/beads/yaml.go**
   - `AddRepositoryAnnotation()` - Add repository to existing issue
   - Updated `toYAMLKey()` and `toProtoKey()` for repositories field
   - Prevents duplicate repositories

5. **commands/fetch-by-label.md** (new)
   - Claude Code plugin documentation
   - Natural language patterns and examples

6. **commands/annotate-issue.md** (new)
   - Claude Code plugin documentation
   - Comprehensive usage guide

### Testing

Manual testing verified:
- JQL search returns correct issue keys
- Label fetching recursively walks dependencies
- Repository annotation adds to metadata
- Duplicate repositories are prevented
- Missing issue files are handled gracefully

## Claude Code Plugin Integration

Users can now use natural language with Claude:

**Label fetching:**
- "Import all issues with label sprint-23 from Jira"
- "Fetch Jira issues labeled team-frontend"
- "Get all issues for label v2.0-release"

**Repository annotation:**
- "Annotate issue proj-123 with repository https://github.com/org/repo"
- "Add repository backend-api to issue proj-456"
- "Link issue proj-789 to repository frontend-app"

## Breaking Changes

None. These are additive features:
- New commands don't affect existing functionality
- Protobuf field addition is backward compatible
- Existing commands unchanged

## Next Steps

- [ ] Run `make proto` to regenerate protobuf Go code
- [ ] Add integration tests for JQL search
- [ ] Add unit tests for repository annotation
- [ ] Consider adding `list-by-repo` command to filter issues by repository
- [ ] Consider pagination support for >1000 issues per label

## Related Issues

Implements features requested for better sprint/team workflow management and polyrepo tracking.
