# About

gong is a Golang CLI tool that pings Pull Request reviewers to remind them to review the PR.

It leverages a ruleset to determine when to ping reviewers and which reviewers to ping.

The project is located at https://github.com/Djiit/gong.

It is built using the [Cobra](https://github.com/spf13/cobra) library, which provides a simple interface for creating command-line applications in Go.

It leverages the GitHub API to fetch the list of reviewers for a given PR and sends them a reminder message via GitHub comments and multiple channels (Slack, Email, etc.).

It is configurable via a configuration file that can be stored in the repository or in the user's home directory. Configuration contains global settings and an advanced rules system to determine when to ping reviewers.

It is designed to be run as a GitHub Action, but can also be run locally. It can also be invoked as a GH CLI extension.

It is tested using unit tests and integration tests. Whenever you are tasked to generate a new feature, please make sure to write tests for it.

When writing comments or documentation, use plain English only.

## Testing your work

You can validate your work by running the following command:

Running the linter:

```bash
mise r ling
```

Running the tests:

```bash
mise r test
```