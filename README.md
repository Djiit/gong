# pingrequest

Tiny, work-in-progress, opinionated CLI for Asana, written in Go.

## Features

### Rules-Based Delay System

PingRequest allows you to configure custom delays based on reviewer names using glob patterns. This lets you define different waiting periods before pinging different reviewers or teams.

#### Configuration Example

In your config file, you can specify a global delay and override it with specific rules:

```yaml
# Global delay (in seconds) before pinging reviewers
delay: 0

# Rules for custom delays based on reviewer name patterns
rules:
  - matchName: "@org/*"
    delay: 86400  # 24 hours delay for organization members
  - matchName: "external-*"
    delay: 172800  # 48 hours delay for external reviewers
  - matchName: "urgent-team" 
    delay: 0  # No delay for urgent team
```

Rules are evaluated in order, and the first matching rule's delay is applied. If no rules match a reviewer, the global delay is used. Setting a delay to 0 means reviewers will be pinged immediately.
