+++
date = '2023-03-16T19:49:16+01:00'
draft = false
title = 'Documentation'
+++

> [!WARNING]
> This documentation is incomplete.

This section contains the complete technical documentation for Gong, including installation guides, configuration options, and usage examples.

## Installation

Gong can be installed in several ways:

### Using Homebrew (preferred)

```bash
brew install Djiit/homebrew-gong/gong
```

### As a GitHub CLI extension

```bash
gh extension install Djiit/gh-gong
```

### Using Go

```bash
go install github.com/Djiit/gong@latest
```

### Using Docker

```bash
docker pull ghcr.io/djiit/gong
```

You can then run gong using:

```bash
docker run ghcr.io/djiit/gong ping --help
```

### Using Binaries

Download the appropriate binary for your platform from the [GitHub releases page](https://github.com/Djiit/gong/releases).

### As a GitHub Action

See the [gong Action](https://github.com/marketplace/actions/gong-action).
