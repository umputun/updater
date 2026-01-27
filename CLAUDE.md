# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Updater is a webhook-based remote command executor. It receives HTTP requests and executes predefined shell commands - typically used for triggering container updates from CI/CD pipelines without exposing SSH keys or credentials.

## Build and Test Commands

```bash
# run all tests
go test ./...

# run single test
go test -run TestName ./app/...

# run tests with race detector
go test -race ./...

# run linter
golangci-lint run

# build binary locally
make build

# build docker image
make docker

# local release build (goreleaser snapshot)
make dist
```

## Architecture

The codebase is small and focused:

```
app/
├── main.go              # CLI entry, signal handling, wires components
├── server/
│   └── server.go        # HTTP server with two endpoints:
│                        #   GET  /update/{task}/{key}?async=[0|1]
│                        #   POST /update (json body with task, secret, async)
└── task/
    ├── config.go        # YAML config loader for task definitions
    └── shell_runner.go  # Executes shell commands (line-by-line or batch mode)
```

**Request flow**: HTTP request → secret key validation → lookup task command from config → execute via ShellRunner → return result

**Key interfaces** (in `server/server.go`):
- `Config` - retrieves command string for task name
- `Runner` - executes command with context and log writer

## Configuration

Tasks defined in YAML (default: `updater.yml`):
```yaml
tasks:
  - name: my-service
    command: |
      docker pull image:tag
      docker restart my-service
```

## CLI Flags

Key flags: `-f` (config file), `-l` (listen address), `-k` (secret key, required), `-b` (batch mode), `--timeout`, `--update-delay`

## Docker

Container includes docker client and openssh-client for typical update workflows. Supports `APP_UID` and `DOCKER_GID` env vars for permission customization.
