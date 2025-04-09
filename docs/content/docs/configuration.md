+++
date = '2023-03-16T19:49:16+01:00'
draft = false
title = 'Configuration'
weight = 10
+++

Gong uses a YAML configuration file to define its behavior, including notification rules and which integrations to use.

## Configuration File Location

Gong will look for a configuration file in the following locations, in order:

1. The path specified with the `--config` flag
2. `./.gong.yml` in the current directory
3. `~/.gong.yml` in the user's home directory

## Configuration Structure

A Gong configuration file consists of the following main sections:

- **delay**: The default delay (in seconds) before pinging reviewers
- **enabled**: Whether pinging is enabled by default
- **integrations**: A list of global integrations to use for notifications
- **rules**: A set of rules to customize behavior for specific reviewers or PRs

### Rules Configuration

Rules allow you to customize Gong's behavior based on different conditions. Each rule can match one or more of the following criteria:

- **matchname**: Match reviewers by their GitHub username (supports glob patterns)
- **matchtitle**: Match PRs by their title (supports glob patterns)
- **matchauthor**: Match PRs by their author's GitHub username (supports glob patterns)

When multiple match criteria are provided in a rule, all must match for the rule to apply.

For each rule, you can specify:

- **delay**: Custom delay before pinging (in seconds)
- **enabled**: Whether pinging is enabled for matches
- **integrations**: Custom integrations to use for notifications

### Example Configuration

```yaml
# Global settings
delay: 3600  # Default delay of 1 hour
enabled: true

# Global integrations (applied by default)
integrations:
  - type: stdout  # Print to terminal
  - type: slack
    params:
      channel: "#code-reviews"
      
# Custom rules
rules:
  # Rule for PRs authored by specific users
  - matchauthor: "critical-team-*"
    delay: 1800  # 30 minutes
    enabled: true
    integrations:
      - type: slack
        params:
          channel: "#urgent-reviews"
      
  # Rule for specific reviewers and authors
  - matchname: "lead-*"
    matchauthor: "junior-*"
    delay: 1200  # 20 minutes
    enabled: true
    
  # Rule based on PR titles
  - matchtitle: "fix: critical-*"
    delay: 900  # 15 minutes
    enabled: true

  # Disable pinging for specific reviewer-author combinations
  - matchname: "busy-user"
    matchauthor: "frequent-contributor"
    enabled: false
```

In this example configuration:

1. The default behavior is to ping after 1 hour (3600 seconds)
2. PRs from authors matching "critical-team-*" will be pinged after 30 minutes via Slack
3. When team leads review PRs from junior developers, they'll be pinged after 20 minutes
4. PRs with titles matching "fix: critical-*" will trigger pings after 15 minutes
5. The user "busy-user" won't be pinged when reviewing PRs from "frequent-contributor"
