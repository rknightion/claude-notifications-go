# Contributing

Thank you for your interest in contributing to Claude Notifications!

## Prerequisites

- **Go 1.21+** (tested with 1.25)
- **just** (https://just.systems) — `brew install just`
- **Claude Code** (tested on v2.0.15)

## Getting Started

### 1. Clone and build

```bash
git clone https://github.com/777genius/claude-notifications-go
cd claude-notifications-go
just setup
just build
```

### 2. Install as local plugin

```bash
# Add as local marketplace
/plugin marketplace add .

# Install plugin
/plugin install claude-notifications-go@claude-notifications-go

# Restart Claude Code for hooks to take effect

# Download binary and configure settings
/claude-notifications-go:init
/claude-notifications-go:settings
```

`/claude-notifications-go:init` will use your locally built binary from `bin/` if it exists, otherwise it downloads from GitHub Releases.

For repeatable local install/update testing without touching your real Claude setup, use:

```bash
just dev-local-install
just dev-local-bootstrap
just dev-local-status
```

This uses an isolated Claude config dir under `~/.claude-dev/claude-notifications-go` by default.

For the full local-development workflow, including real-`claude` E2E tests and switching your real Claude environment between local and remote sources, see **[docs/LOCAL_DEVELOPMENT.md](docs/LOCAL_DEVELOPMENT.md)**.

## Local Testing Workflows

Use the smallest workflow that matches the change:

- `just dev-local-install` / `just dev-local-update` / `just dev-local-bootstrap` for safe testing in an isolated Claude config
- `just e2e-smoke` / `just e2e-manual` for real-`claude` validation
- `just dev-real-local` (asks for confirmation) only when you must validate inside your real `~/.claude` setup

Start with:

```bash
just e2e-status
```

Then follow **[docs/LOCAL_DEVELOPMENT.md](docs/LOCAL_DEVELOPMENT.md)** for the detailed workflow, platform support, and recommended validation matrix.

## Project Structure

See [Architecture](docs/ARCHITECTURE.md) for a detailed overview. Key directories:

| Directory | Description |
|-----------|-------------|
| `cmd/` | CLI entry points (`claude-notifications`, `sound-preview`, `list-devices`, `list-sounds`) |
| `internal/` | Core logic (analyzer, hooks, notifier, webhook, config, audio, etc.) |
| `pkg/jsonl/` | JSONL streaming parser |
| `commands/` | Plugin skill definitions (`.md` files) |
| `sounds/` | Built-in notification sounds (MP3) |

## Task Surface

```bash
just --list
```

`just check` is the gate; every recipe carries a doc comment and a group.

## Testing

### Run all tests

```bash
just test
```

### Run specific packages

```bash
go test ./internal/analyzer -v
go test ./internal/hooks -v
go test ./internal/config -v
go test ./internal/dedup -v -race
```

### Real-Claude smoke tests

See **[docs/LOCAL_DEVELOPMENT.md](docs/LOCAL_DEVELOPMENT.md)** for supported platforms, command examples, and manual click-to-focus validation notes.

### Run a single test

```bash
just test TestStateMachine
```

### Coverage

```bash
just coverage
open coverage.html
```

## CI/CD

GitHub Actions run on every push:

- **ci-ubuntu.yml** — Tests on Ubuntu
- **ci-macos.yml** — Tests on macOS
- **ci-windows.yml** — Tests on Windows
- **release.yml** — Builds and publishes release binaries

All three platform CIs must pass before merging.

## Submitting Changes

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Make your changes
4. Run the gate: `just check`
5. Commit with a descriptive message following [Conventional Commits](https://www.conventionalcommits.org/):
   - `feat:` — new features
   - `fix:` — bug fixes
   - `docs:` — documentation changes
   - `test:` — adding/updating tests
   - `chore:` — maintenance tasks
6. Open a Pull Request against `main`

## Releasing

See **[Release Checklist](docs/RELEASE.md)** for the full step-by-step guide.

## Code Style

- Standard Go formatting (`go fmt`)
- Use `go vet` for static analysis
- Keep functions focused and small
- Add tests for new functionality
- Use structured logging via `internal/logging` package

## Reporting Issues

Found a bug or have a feature request? [Open an issue](https://github.com/777genius/claude-notifications-go/issues).
