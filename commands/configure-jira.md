---
description: Configure Jira API credentials for importing issues
argument-hint:
---

Configure Jira credentials to enable importing issues from Jira. This is a one-time setup required before using the `import-jira` command.

## Configuration Methods

Jira credentials can be configured in three ways (in order of precedence):

### 1. Interactive Configuration (Recommended)

Run the interactive configuration command:
```bash
jira-beads-sync configure
```

This will prompt for:
- **Jira Base URL**: The root URL of your Jira instance (e.g., `https://jira.example.com`)
- **Username/Email**: Your Jira account email
- **API Token**: A Jira API token (not your password)

The configuration is saved to `~/.config/jira-beads-sync/config.yml`

### 2. Environment Variables

Set these environment variables:
```bash
export JIRA_BASE_URL=https://jira.example.com
export JIRA_USERNAME=your-email@example.com
export JIRA_API_TOKEN=your-api-token
```

Environment variables override the config file.

### 3. Config File

Manually create `~/.config/jira-beads-sync/config.yml`:
```yaml
jira:
  base_url: https://jira.example.com
  username: your-email@example.com
  api_token: your-api-token
```

## Getting a Jira API Token

1. Go to: https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Give it a descriptive name (e.g., "Claude Code jira-beads-sync")
4. Copy the token (you won't be able to see it again)
5. Use this token in the configuration

**Security Note**: API tokens should be treated like passwords. Never commit them to version control.

## Verification

After configuration, test the connection by importing a known issue:
```bash
jira-beads-sync quickstart <test-issue-key>
```

If you see "invalid configuration" or authentication errors, run `jira-beads-sync configure` again.

## Example Interaction

```
User: Configure Jira credentials

Claude: I'll help you configure Jira credentials. This is stored securely in your home directory.

[Runs: jira-beads-sync configure]

Jira Configuration
==================

Jira Base URL (e.g., https://jira.example.com): https://company.atlassian.net
Jira Username/Email: user@example.com
Jira API Token: ************************

✓ Configuration saved successfully

Configuration is stored at: ~/.config/jira-beads-sync/config.yml

To get a Jira API token, visit:
https://id.atlassian.com/manage-profile/security/api-tokens

You can now use: import-jira <jira-url-or-key>
```

## Troubleshooting

- **Authentication failed**: Verify your API token is correct and hasn't expired
- **Base URL wrong**: Ensure the base URL doesn't include `/browse/` or issue keys
- **Permission denied**: Check that your Jira account has read access to the issues you want to import
- **Network error**: Verify connectivity to your Jira instance

## Security Considerations

- The config file at `~/.config/jira-beads-sync/config.yml` has restricted permissions (0600)
- API tokens are more secure than passwords and can be revoked at any time
- Consider using environment variables in CI/CD environments
- Never commit credentials to git repositories
