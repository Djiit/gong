# About

gong is a Golang CLI tool that pings Pull Request reviewers to remind them to review the PR.

It leverages a ruleset to determine when to ping reviewers and which reviewers to ping.

The project is located at https://github.com/Djiit/gong.

It is built using the [Cobra](https://github.com/spf13/cobra) library, which provides a simple interface for creating command-line applications in Go.

It leverages the GitHub API to fetch the list of reviewers for a given PR and sends them a reminder message via GitHub comments and multiple channels (Slack, Email, etc.).

It is designed to be run as a GitHub Action, but can also be run locally. It can also be invoked as a GH CLI extension.

It is tested using unit tests and integration tests. Whenever you are tasked to generate a new feature, please make sure to write tests for it.

When writing comments or documentation, use plain English only.

## Pushing your changes

- We use conventional commits to manage our commit messages. Please follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification when writing your commit messages.
