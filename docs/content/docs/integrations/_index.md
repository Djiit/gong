+++
date = '2025-03-16T19:49:16+01:00'
draft = false
title = 'Integrations'
weight = 20
+++

# Integrations

Gong supports multiple integration channels to notify reviewers about pending pull requests. These integrations can be configured globally or per rule in your configuration file.

## Available Integrations

Gong currently supports the following integrations:

### Standard Output (`stdout`)

The simplest integration that outputs review request information to the console. This is useful for local development or when running Gong manually.

**Configuration:**
```yaml
integrations:
  - type: stdout
```

No additional parameters are required for this integration.

### GitHub Comment (`comment`)

This integration adds a comment to the Pull Request that mentions all reviewers who need to review the PR.

**Configuration:**
```yaml
integrations:
  - type: comment
```

The GitHub token must be provided via the `--github-token` flag or the `GONG_GITHUB_TOKEN` environment variable.

### Slack (`slack`)

Sends notifications to a specified Slack channel about pending review requests.

**Configuration:**
```yaml
integrations:
  - type: slack
    params:
      channel: "#reviews" # The Slack channel to send notifications to
```

**Parameters:**
- `channel`: The Slack channel to send the notifications to (required)

### GitHub Actions (`actions`)

When used within GitHub Actions workflows, this integration outputs results as GitHub Actions environment variables and workflow outputs for further processing.

**Configuration:**
```yaml
integrations:
  - type: actions
```

**Outputs:**

To `GITHUB_OUTPUT`:
- `reviewers`: Comma-separated list of reviewers to ping
- `reviewersCount`: Number of reviewers to ping
- `reviewersDetails`: Multiline list of all reviewers with their status

To `GITHUB_ENV`:
- `GONG_REVIEWERS`: Comma-separated list of reviewers to ping
- `GONG_REVIEWERS_COUNT`: Number of reviewers to ping
- `GONG_REVIEWERS_DETAILS`: Multiline list of all reviewers with their status

## Using Multiple Integrations

You can configure multiple integrations to run simultaneously:

```yaml
integrations:
  - type: stdout
  - type: slack
    params:
      channel: "#reviews"
  - type: comment
```

## Per-Rule Integrations

In addition to global integrations, you can specify different integrations for specific rules:

```yaml
rules:
  - matchName: "urgent-team"
    delay: 0
    integrations:
      - type: slack
        params:
          channel: "#urgent"
      - type: comment
  
  - matchName: "external-*"
    delay: 172800
    integrations:
      - type: comment
```

When a rule doesn't specify any integrations, the global integrations are used.

## Example Configuration

Here's a complete example of a configuration using multiple integrations with rule-specific overrides:

```yaml
repository: Djiit/gong
delay: 0
enabled: true

# Global integrations configuration
integrations:
  - type: stdout
  - type: slack
    params:
      channel: "#reviews" 

# Rules with custom integrations
rules:
  - matchName: "@org/*"
    delay: 86400  # 24 hours delay for organization members
    integrations:
      - type: slack
        params:
          channel: "#org-reviews"
      - type: comment
  
  - matchName: "external-*"
    delay: 172800  # 48 hours delay for external reviewers
    integrations: 
      - type: comment

  - matchName: "urgent-team" 
    delay: 0  # No delay for urgent team members
    integrations:
      - type: slack
        params:
          channel: "#urgent"
      - type: comment
