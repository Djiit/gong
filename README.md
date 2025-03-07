# pingrequest

Tiny, work-in-progress, opinionated CLI for Asana, written in Go.

## Features

### Rules-Based Delay System

PingRequest allows you to configure custom delays based on reviewer names using glob patterns. This lets you define different waiting periods before pinging different reviewers or teams.

#### Configuration Example

In your config file, you can specify a global delay and override it with specific rules:

```yaml
# Global settings
enabled: true # Enable or disable pinging functionality globally
delay: 0      # Global delay (in seconds) before pinging reviewers

# Rules for custom delays based on reviewer name patterns
rules:
  - matchName: "@org/*"
    delay: 86400  # 24 hours delay for organization members
    enabled: true # Enable or disable this specific rule (defaults to true if not specified)
  - matchName: "external-*"
    delay: 172800  # 48 hours delay for external reviewers
    enabled: true
  - matchName: "urgent-team" 
    delay: 0  # No delay for urgent team
  - matchName: "do-not-ping-*"
    delay: 0
    enabled: false # Disable pinging for reviewers matching this pattern
```

Rules are evaluated in order, and the first matching rule's delay is applied. If no rules match a reviewer, the global delay is used. Setting a delay to 0 means reviewers will be pinged immediately.

The `enabled` setting can be specified:
- At the global level to enable/disable all pinging functionality
- At the rule level to enable/disable specific rules

By default, `enabled` is set to `true` if not specified.
