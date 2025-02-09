# About

PingRequest is a Golang CLI tool that pings Pull Request reviewers to remind them to review the PR.

The project is located at https://github.com/Djiit/pingrequest.

It is built using the [Cobra](https://github.com/spf13/cobra) library, which provides a simple interface for creating command-line applications in Go.

It leverages the GitHub API to fetch the list of reviewers for a given PR and sends them a reminder message via GitHub comments and multiple channels (Slack, Email, etc.).

It is designed to be run as a GitHub Action, but can also be run locally. It can also be invoked as a GH CLI extension.

It is tested using unit tests and integration tests.
