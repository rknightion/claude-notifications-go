---
id: CNG-0006
title: Migrate the repo task surface to just and retire Makefiles and ad-hoc scripts
status: Parked
assignee: []
created_date: '2026-08-28 19:17'
updated_date: '2026-08-29 13:54'
labels:
  - 'wave:2-fleet'
dependencies: []
priority: medium
type: chore
ordinal: 6000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Migrate `rknightion/claude-notifications-go` from `Makefile` + ad-hoc `scripts/*.sh` invocations to a
single top-level `justfile`, per the frozen fleet `just` standard (mandatory seven-recipe vocabulary,
six groups, `check` as the gate).

Stack: Go 1.21 module (`github.com/777genius/claude-notifications`, CGO via `malgo`/`beep`), a Swift
package at `swift-notifier/` (SwiftPM, macOS-only), and a large set of shipped installer/bootstrap
shell scripts under `bin/`. Every recipe body below was taken from this repo's real commands and the
whole justfile was parsed and format-checked against `just 1.58.0`.

## 1. Outcome

The repo has one top-level `justfile` and no `Makefile`. `just --list` is the only answer to "what
can I do here": it shows `default`, `setup`, and 21 grouped recipes across `check` / `build` / `dev`
/ `release`. `just check` is the full local gate — `fmt-check`, `vet`, `golangci-lint`, `go test
-race`, the four host binaries plus a `help` smoke test, both `install.sh` shell test suites, and
(on macOS) the Swift notifier tests. The three CI workflows install a pinned `just` and call
one-line `just <recipe>` steps instead of inline shell. `setup.sh` is gone (absorbed into `just
setup`); every shipped runtime script under `bin/` and `scripts/` survives untouched but is reached
through a recipe. `AGENTS.md`, `CONTRIBUTING.md`, `docs/LOCAL_DEVELOPMENT.md`, `docs/RELEASE.md` and
`backlog/config.yml` name `just` recipes and never `make`.

## 2. The complete justfile

Drop this in at the repo root as `justfile` (lowercase, no extension). It parses clean, `just --list`
exits 0, `just --dump --dump-format json` exits 0, and `just --fmt --check` exits 0 under
`just 1.58.0` — all four verified before this task was filed. Do not reorder the attributes: `just
--fmt` enforces alphabetical attribute order, so `[confirm(...)]` must precede `[group(...)]`, and
`[linux, windows]` on one line is rewritten to two lines.

```just
set shell := ["bash", "-euo", "pipefail", "-c"]

goos := if os() == "macos" { "darwin" } else { os() }
goarch := if arch() == "aarch64" { "arm64" } else { "amd64" }
binext := if os() == "windows" { ".exe" } else { "" }
main_bin := "bin/claude-notifications-" + goos + "-" + goarch + binext

# show the task surface
default:
    @just --list

# install Go + Swift dependencies and mark repo scripts executable (idempotent)
setup: _setup-swift
    go mod download
    go mod verify
    chmod +x bin/*.sh bin/*.py scripts/*.sh scripts/*.py

[macos]
_setup-swift:
    swift package resolve --package-path swift-notifier

[linux]
[windows]
_setup-swift:
    @echo "swift package resolve skipped (macOS only)"

# format Go sources and this justfile in place
[group('check')]
fmt:
    gofmt -s -w .
    just --fmt

# verify Go and justfile formatting; never writes
[group('check')]
[no-exit-message]
fmt-check:
    just --fmt --check
    unformatted="$(gofmt -s -l .)"; if [ -n "$unformatted" ]; then echo "gofmt -s -w needed for:"; echo "$unformatted"; exit 1; fi

# run go vet across every package for the host GOOS
[group('check')]
[no-exit-message]
vet:
    go vet ./...

# run go vet plus golangci-lint across every package
[group('check')]
[no-exit-message]
lint: vet
    {{ require("golangci-lint") }} run --timeout=5m ./...

# run the Go test suite with race detection and coverage; filter is a -run regex
[group('check')]
[no-exit-message]
test filter="":
    go test -race -covermode=atomic -coverprofile=coverage.txt -run '{{ filter }}' ./...

# regenerate coverage.html from a full test run
[group('check')]
coverage: test
    go tool cover -html=coverage.txt -o coverage.html

# run the install.sh unit tests
[group('check')]
[no-exit-message]
test-install:
    bash bin/install_test.sh

# run the install.sh end-to-end tests; pass --real-network for the live-network diagnostics
[group('check')]
[no-exit-message]
test-install-e2e *flags:
    bash bin/install_e2e_test.sh {{ flags }}

# run the Swift notifier test suite
[group('check')]
[macos]
[no-exit-message]
test-swift:
    swift test --package-path swift-notifier

# swift toolchain is macOS-only
[group('check')]
[linux]
[windows]
test-swift:
    @echo "swift notifier tests skipped (macOS only)"

# build the host-platform development binaries into bin/
[group('build')]
build:
    go build -o bin/claude-notifications-{{ goos }}-{{ goarch }}{{ binext }} ./cmd/claude-notifications
    go build -o bin/sound-preview-{{ goos }}-{{ goarch }}{{ binext }} ./cmd/sound-preview
    go build -o bin/list-sounds-{{ goos }}-{{ goarch }}{{ binext }} ./cmd/list-sounds
    go build -o bin/list-devices-{{ goos }}-{{ goarch }}{{ binext }} ./cmd/list-devices

# smoke-test the freshly built host binary
[group('check')]
smoke: build
    ./{{ main_bin }} help

# build the four optimized release binaries for one target into dist/
[group('release')]
[script('bash')]
build-release platform arch:
    set -euo pipefail
    mkdir -p dist
    suffix=""
    if [ "{{ platform }}" = "windows" ]; then suffix=".exe"; fi
    for cmd in claude-notifications sound-preview list-devices list-sounds; do
        echo "Building $cmd..."
        go build -ldflags="-s -w" -trimpath -o "dist/${cmd}-{{ platform }}-{{ arch }}${suffix}" "./cmd/${cmd}"
    done

# build, sign and optionally notarize ClaudeNotifier.app; pass --ci and/or --skip-notarize
[group('build')]
[macos]
[working-directory('swift-notifier')]
build-notifier *flags:
    bash scripts/build-app.sh {{ flags }}

# verify a release tag matches every committed version source
[group('release')]
[script('bash')]
verify-version tag:
    set -euo pipefail
    release_tag="{{ tag }}"
    release_version="${release_tag#v}"
    if [[ ! "$release_version" =~ ^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
        echo "Release tag '$release_tag' must use the stable vX.Y.Z format"
        exit 1
    fi
    check_version() {
        if [ "$2" != "$release_version" ]; then
            echo "$1 version '$2' does not match tag '$release_version'"
            exit 1
        fi
    }
    check_version "Go binary" "$(sed -n 's/^const version = "\([^"]*\)"$/\1/p' cmd/claude-notifications/main.go)"
    check_version "Plugin manifest" "$(jq -er '.version' .claude-plugin/plugin.json)"
    check_version "Marketplace metadata" "$(jq -er '.metadata.version' .claude-plugin/marketplace.json)"
    check_version "Marketplace plugin" "$(jq -er '.plugins[] | select(.name == "claude-notifications-go") | .version' .claude-plugin/marketplace.json)"
    if ! awk -v heading="## [$release_version]" '$0 == heading || index($0, heading " - ") == 1 { found = 1 } END { exit(found ? 0 : 1) }' CHANGELOG.md; then
        echo "CHANGELOG.md has no entry for $release_version"
        exit 1
    fi

# remove everything setup and build can reproduce
[group('build')]
clean:
    rm -rf dist swift-notifier/.build bin/test_fixtures
    rm -f coverage.txt coverage.html
    rm -f bin/claude-notifications-* bin/sound-preview* bin/list-sounds* bin/list-devices* bin/*.exe

# install the plugin into an isolated Claude config
[group('dev')]
dev-local-install:
    bash scripts/dev-local-plugin.sh install

# update the plugin in the isolated Claude config
[group('dev')]
dev-local-update:
    bash scripts/dev-local-plugin.sh update

# run bootstrap.sh against the isolated Claude config
[group('dev')]
dev-local-bootstrap:
    bash scripts/dev-local-plugin.sh bootstrap

# show plugin status in the isolated Claude config
[group('dev')]
dev-local-status:
    bash scripts/dev-local-plugin.sh status

# delete the isolated Claude dev config
[confirm('Delete the isolated Claude dev config?')]
[group('dev')]
dev-local-reset:
    bash scripts/dev-local-plugin.sh reset

# point your REAL Claude marketplace at this checkout
[confirm('This rewrites your real Claude plugin config. Continue?')]
[group('dev')]
dev-real-local:
    bash scripts/dev-real-plugin.sh local

# point your REAL Claude marketplace back at the published source
[confirm('This rewrites your real Claude plugin config. Continue?')]
[group('dev')]
dev-real-remote:
    bash scripts/dev-real-plugin.sh remote

# toggle your REAL Claude marketplace between local and published
[confirm('This rewrites your real Claude plugin config. Continue?')]
[group('dev')]
dev-real-toggle:
    bash scripts/dev-real-plugin.sh toggle

# show marketplace and plugin status in your real Claude config
[group('dev')]
dev-real-status:
    bash scripts/dev-real-plugin.sh status

# show real-Claude E2E support and target status
[group('dev')]
e2e-status:
    bash scripts/e2e-real-claude.sh status

# run the real-Claude smoke test via --plugin-dir (long-running)
[group('dev')]
e2e-smoke:
    bash scripts/e2e-real-claude.sh smoke-plugin-dir

# run the real-Claude smoke test against the installed plugin (long-running)
[group('dev')]
e2e-smoke-installed:
    bash scripts/e2e-real-claude.sh smoke-installed

# run manual click-to-focus validation via --plugin-dir (long-running, interactive)
[group('dev')]
e2e-manual:
    bash scripts/e2e-real-claude.sh manual-click-plugin-dir

# run manual click-to-focus validation against the installed plugin (long-running, interactive)
[group('dev')]
e2e-manual-installed:
    bash scripts/e2e-real-claude.sh manual-click-installed

# collect a Linux click-to-focus diagnostics report
[group('dev')]
linux-focus-debug *flags:
    bash scripts/linux-focus-debug.sh {{ flags }}

# THE GATE - everything CI enforces
[group('check')]
check: fmt-check lint build smoke test test-install test-install-e2e test-swift
```

## 3. Makefile disposition

`Makefile` is 6494 bytes, 18 targets. Every one is accounted for. Delete the file at the end of the
migration with `git rm Makefile`.

| Make target | Replacement | Notes |
|---|---|---|
| `build` | `just build` | Now writes `bin/<cmd>-<goos>-<goarch>[.exe]`, **not** `bin/claude-notifications`. See Traps 1 and 2 — the old path is a tracked symlink. Also builds `list-devices`, which the Makefile omitted but `release.yml` ships. |
| `build-all` | **DROP** — use `just build-release <platform> <arch>` | The Makefile cross-compiled all five targets from one host. This module needs CGO (`malgo`, `beep`); Go silently disables CGO when cross-compiling, so `make build-all` produced binaries that are not release-equivalent. `release.yml` already builds natively per runner and is authoritative. |
| `test` | `just test` | Was `go test -v -cover ./...`. Now always `-race` with an atomic coverprofile, matching what CI actually enforces. |
| `test-race` | `just test` | Merged — there is no non-race test recipe any more. |
| `test-coverage` | `just coverage` | `coverage: test` then `go tool cover -html=coverage.txt -o coverage.html`. |
| `lint` | `just lint` | The make target ran `go vet ./...` **and `go fmt ./...`**, i.e. it mutated the tree. `just lint` is `vet` + `golangci-lint run --timeout=5m ./...` and never writes. Formatting moved to `just fmt` / `just fmt-check`. |
| `install` | **DROP** | It was `cp bin/claude-notifications /usr/local/bin/` — a global install of a dev binary, forbidden by the standard and superseded by `bin/install.sh` / `bin/bootstrap.sh`, which is how users actually install. |
| `dev-local-install` | `just dev-local-install` | Wraps `scripts/dev-local-plugin.sh install`. |
| `dev-local-update` | `just dev-local-update` | |
| `dev-local-bootstrap` | `just dev-local-bootstrap` | |
| `dev-local-status` | `just dev-local-status` | |
| `dev-local-reset` | `just dev-local-reset` | Gains `[confirm]` — it deletes the isolated dev Claude home. |
| `dev-real-local` | `just dev-real-local` | Gains `[confirm]` — it rewrites the developer's real `~/.claude` plugin config. |
| `dev-real-remote` | `just dev-real-remote` | `[confirm]`, same reason. |
| `dev-real-toggle` | `just dev-real-toggle` | `[confirm]`, same reason. |
| `dev-real-status` | `just dev-real-status` | Read-only, no confirm. |
| `e2e-status` | `just e2e-status` | |
| `e2e-smoke` | `just e2e-smoke` | `scripts/e2e-real-claude.sh smoke-plugin-dir`. |
| `e2e-smoke-installed` | `just e2e-smoke-installed` | |
| `e2e-manual` | `just e2e-manual` | |
| `e2e-manual-installed` | `just e2e-manual-installed` | |
| `linux-focus-debug` | `just linux-focus-debug` | Variadic `*flags` so `--output` / `--stdout` still work. |
| `clean` | `just clean` | **Behaviour deliberately narrowed.** `make clean` ran `rm -rf bin/`, which deletes the tracked `bin/install.sh`, `bin/bootstrap.sh`, `bin/hook-wrapper.sh`, `bin/install_test.sh`, `bin/install_e2e_test.sh`, `bin/mock_server.py` and the `bin/claude-notifications` symlink. `just clean` removes only named generated artifacts. |
| `rebuild-and-commit` | **DROP** | It copied `dist/*` into `bin/` and printed `git add bin/claude-notifications-*`, which cannot work — `.gitignore:3` ignores `bin/claude-notifications-*`. Release binaries come from `release.yml` only. |
| `build-notifier` | `just build-notifier` | `[macos]` + `[working-directory('swift-notifier')]`, variadic `*flags` so CI can pass `--ci` / `--ci --skip-notarize`. |
| `help` | `just` / `just --list` | `default: @just --list`. |
| `.PHONY` block (lines 1-5) | delete | Meaningless in `just`. |
| `BINARY` / `SOUND_PREVIEW` / `LIST_SOUNDS` / `*_PATH` vars | folded into recipe bodies | Only `goos`/`goarch`/`binext`/`main_bin` survive as justfile variables. |
| `RELEASE_FLAGS=-ldflags="-s -w" -trimpath` | inline in `build-release` | Kept byte-identical. |

Final step of the migration: `git rm Makefile`.

## 4. Script disposition

`git ls-files` finds 13 script files. Nine are KEEP, one is ABSORB, three are KEEP with no recipe
(they are not developer tasks). Do not open `bin/install.sh` (64 KB), `bin/bootstrap.sh` (34 KB) or
`bin/install_e2e_test.sh` (81 KB) expecting to inline them — they are shipped programs.

| Script | Verdict | Recipe | Reason / exact lines |
|---|---|---|---|
| `setup.sh` (62 lines) | **ABSORB** | `just setup` | Thin: it `chmod +x`es two files and echoes install instructions that already live in `README.md`. Nothing in the repo, plugin manifests or docs references it (`grep -rn 'setup\.sh'` matches only its own header comment). Its only real effect becomes the `chmod +x bin/*.sh bin/*.py scripts/*.sh scripts/*.py` line in `setup`. Then `git rm setup.sh`. |
| `bin/install.sh` | KEEP | reached via `bin/bootstrap.sh` and the plugin `/init` command; no dev recipe | Shipped end-user installer fetched over `curl` on machines with no `just`. |
| `bin/bootstrap.sh` | KEEP | `just dev-local-bootstrap` (via `scripts/dev-local-plugin.sh bootstrap`) | Shipped: `README.md:71` and `README.md:131` tell users to `curl … /main/bin/bootstrap.sh | bash`. The raw URL is a public contract. |
| `bin/hook-wrapper.sh` | KEEP | none | Shipped runtime artifact executed by Claude Code / Codex hooks on the user's machine. |
| `bin/claude-notifications` | KEEP | none | Tracked **symlink** (git mode 120000) to `claude-notifications-darwin-arm64`. Not a script. See Trap 2. |
| `bin/install_test.sh` (7.5 KB) | KEEP | `just test-install` | Shell test suite. The recipe runs it; it is not task-shaped. |
| `bin/install_e2e_test.sh` (81 KB) | KEEP | `just test-install-e2e` | Shell test suite with a `--real-network` mode; spawns `bin/mock_server.py`. |
| `bin/mock_server.py` (180 lines) | KEEP | none — invoked by `install_e2e_test.sh` | A real program (an HTTP fault-injection server). |
| `scripts/dev-local-plugin.sh` (332 lines) | KEEP | `just dev-local-install` / `-update` / `-bootstrap` / `-status` / `-reset` | 20+ shell functions, a `case` dispatcher and JSON probing. Non-trivial control flow. |
| `scripts/dev-real-plugin.sh` (275 lines) | KEEP | `just dev-real-local` / `-remote` / `-toggle` / `-status` | Same: functions, backup/restore state machine, `case` dispatch. |
| `scripts/e2e-real-claude.sh` (550 lines) | KEEP | `just e2e-status` / `e2e-smoke` / `e2e-smoke-installed` / `e2e-manual` / `e2e-manual-installed` | Platform-capability matrix, timeouts, output slicing. A harness, not a task. |
| `scripts/linux-focus-debug.sh` (476 lines) | KEEP | `just linux-focus-debug` | Diagnostics report builder **and** a shipped artifact: `docs/troubleshooting.md:75` and `docs/CLICK_TO_FOCUS.md:76` instruct users to `curl … /main/scripts/linux-focus-debug.sh | bash`. Path is a public contract — never move it. |
| `scripts/iterm2-select-tab.py` (295 lines) | KEEP | none | Resolved at runtime by `internal/notifier/tmux_iterm2.go:29` as `<pluginRoot>/scripts/iterm2-select-tab.py` and asserted by `internal/notifier/notifier_test.go:1446` and `internal/notifier/tmux_iterm2_test.go:176`. Moving or deleting it breaks click-to-focus and four tests. |
| `swift-notifier/scripts/build-app.sh` (201 lines) | KEEP | `just build-notifier [flags]` | Codesign / entitlement / notarization program with `--ci` and `--skip-notarize` argument parsing. |

Only two files are deleted by this task: `Makefile` and `setup.sh`.

## 5. CI changes

There are seven workflow files. Three change substantially, two change in one or two steps, two must
not be touched at all.

**This repo has NO `ci-success` aggregator job.** Do not invent one — it is out of scope here. The
required-check names are the job `name:` strings themselves: `Test on Ubuntu`, `Lint`,
`Test on macOS (${{ matrix.go }})`, `Swift notifier tests`, `Test on Windows`. **Do not rename a
single job or step-bearing `name:`.**

Also preserve everywhere: `strategy.matrix` blocks and the `matrix.go` / `matrix.platform` /
`matrix.arch` / `matrix.binary` values, `timeout-minutes`, `permissions:` blocks, `if:` conditions,
`continue-on-error: true`, every `uses:` action, and the `on:` triggers.

### 5.1 The setup-just step

Insert this immediately after `actions/setup-go` (or after `actions/checkout` in jobs with no Go
step) in every job that calls `just`:

```yaml
      - name: Set up just
        uses: extractions/setup-just@RESOLVE_SHA # v4
        with:
          just-version: '1.58.0'
```

Resolve `RESOLVE_SHA` before committing:

```bash
gh api repos/extractions/setup-just/git/ref/tags/v4 --jq .object.sha
```

`just-version` is pinned exactly because `just --fmt` output carries no backwards-compatibility
guarantee and an unpinned bump would turn `fmt-check` red with no repo change. Note this repo's
other actions use tag pins (`actions/checkout@v7`); that pre-existing inconsistency is out of scope,
but the new action still gets a SHA pin per the fleet standard.

### 5.2 `.github/workflows/ci-ubuntu.yml`

`test` job — insert setup-just after `Set up Go` (after line 27), then:

| Current lines | Step | Becomes |
|---|---|---|
| 29-30 | Display Go version | unchanged |
| 32-33 | `sudo apt-get install -y libasound2-dev` | **unchanged** — system CGO dep, `just setup` must not sudo |
| 35-40 | Download + Verify dependencies (2 steps) | one step, `run: just setup` |
| 41-42 | `go vet ./...` | `run: just vet` |
| 44-51 | `go fmt` output-capture block | `run: just fmt-check` |
| 53-54 | `go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...` | `run: just test` |
| 56-62 | `codecov/codecov-action@v7` | **unchanged** — it is a `uses:`, and `just test` still writes `./coverage.txt` |
| 64-68 | Build binary + Test binary execution (2 steps) | one step, `run: just smoke` (`smoke` depends on `build`) |
| 70-71 | `bash bin/install_test.sh` | `run: just test-install` |
| 73-74 | `bash bin/install_e2e_test.sh` | `run: just test-install-e2e` |
| 76-80 | `bash bin/install_e2e_test.sh --real-network` | `run: just test-install-e2e --real-network` — keep the `if:`, `timeout-minutes: 5` and `continue-on-error: true` |

`lint` job (lines 82-106) — **unchanged**. `golangci/golangci-lint-action@v9` pinned to `v2.12.2`
stays a `uses:`; never convert it to `run: just lint`.

### 5.3 `.github/workflows/ci-macos.yml`

`test` job — insert setup-just after `Set up Go` (after line 31), then the same mapping as ubuntu:
35-40 → `just setup`; 42-43 → `just vet`; 45-52 → `just fmt-check`; 54-55 → `just test`;
**57-67 (Build binary, Test binary execution, Build sound preview binary, Build list-sounds binary —
four steps) collapse into one `run: just smoke`**; 69-76 "Check for system sounds" stays unchanged
(pure diagnostics, no build logic); 78-79 → `just test-install`; 81-82 → `just test-install-e2e`;
84-88 → `just test-install-e2e --real-network` with its `if:`/`timeout-minutes`/`continue-on-error`
intact.

`swift-test` job (lines 90-100) — insert setup-just after `Checkout code` (after line 97), then
replace `run: swift test --package-path swift-notifier` with `run: just test-swift`.

### 5.4 `.github/workflows/ci-windows.yml`

Insert setup-just after `Set up Go` (after line 27), then: 32-36 → `just setup`; 38-39 →
`just vet`; **41-49 (the `shell: pwsh` `$fmt_output` block) → `run: just fmt-check`, and delete the
`shell: pwsh` line** — just runs its own bash; 51-53 → `run: just test` (drop the now-redundant
`shell: bash`); **55-66 (Build binary, Test binary execution, Build sound preview binary, Build
list-sounds binary) collapse into one `run: just smoke`**; 68-70 → `just test-install`; 72-74 →
`just test-install-e2e`.

Windows runners carry Git Bash, which is what `set shell := ["bash", "-euo", "pipefail", "-c"]`
resolves to. There is no `set windows-shell` in the justfile, deliberately.

### 5.5 `.github/workflows/release.yml`

- `verify-version` job: insert setup-just after `Checkout code` (after line 19). Replace the entire
  `run: |` body at lines 24-52 with `run: just verify-version "$RELEASE_TAG"`. Keep the
  `env: RELEASE_TAG: ${{ github.ref_name }}` block. The recipe is a line-for-line port including the
  `sed`/`jq`/`awk` checks; `jq` is present on `ubuntu-latest`.
- `build-matrix` job: insert setup-just after `Set up Go` (after line 100). Replace the `run: |`
  body at lines 114-129 with `run: just build-release ${{ matrix.platform }} ${{ matrix.arch }}`.
  **Keep** `shell: bash` and the `env:` block (`CGO_ENABLED: '1'`,
  `MACOSX_DEPLOYMENT_TARGET: ${{ matrix.macos_deployment_target }}`) — the recipe inherits them. The
  recipe emits exactly the same four filenames the matrix's `matrix.binary` values expect.
- `build-matrix` "Verify macOS deployment target" (lines 131-143): **unchanged.**
- `build-notifier` job: insert setup-just after `Checkout code` (after line 159). Replace lines
  199-201 (`cd swift-notifier` + `bash scripts/build-app.sh --ci`) with `run: just build-notifier
  --ci`. The `[working-directory('swift-notifier')]` attribute replaces the `cd`. Keep the Apple
  certificate import step (162-192) exactly as it is — keychain plumbing, not a task.
- `create-release` and `test-binaries` jobs: **unchanged.** Artifact download, checksums and
  `softprops/action-gh-release@v3` stay as they are.

### 5.6 `.github/workflows/notifier-signing-smoke.yml`

Insert setup-just after `Checkout code` (after line 26). Replace the `run: |` body at lines 65-71
with:

```yaml
        run: |
          if [ "${{ inputs.skip_notarize }}" = "true" ]; then
            just build-notifier --ci --skip-notarize
          else
            just build-notifier --ci
          fi
```

Keep `permissions: contents: read`, the certificate import step, the signature-verification step and
the artifact upload unchanged.

### 5.7 Do not touch

`.github/workflows/reviewrouter.yml` and `.github/workflows/reviewrouter-interaction.yml` are pure
`uses:` calls into `777genius/review-router/.github/workflows/*-reusable.yml@v1` with secrets and
`id-token: write`. No `run:` steps exist in them. Leave both files byte-identical.

## 6. Docs and agent-contract changes

`grep -rn --include='*.md' -E 'make [a-z-]+|scripts/[a-z-]+\.sh' . --exclude-dir=.git` finds every
site. `CHANGELOG.md` is historical — never rewrite it.

### `AGENTS.md`

Line 32 currently reads:

> `- Use `go test -race ./...`, `go vet ./...`, `make build`, `bash bin/install_test.sh`, and `bash bin/install_e2e_test.sh` as the repository gate. Run narrower tests while iterating.`

Replace that single bullet with:

```markdown
- `just check` is the repository gate. Run narrower recipes (`just test <regex>`, `just vet`) while iterating.
```

Then append a new section at the end of `AGENTS.md` (after the existing repository-specific rules,
outside the `<!-- BACKLOG.MD GUIDELINES -->` block):

```markdown
## Task interface

This repo's task surface is a `justfile`. Discover it, don't guess it:

    just --list                        # human-readable
    just --dump --dump-format json     # machine-readable
    just --show <recipe>               # what a recipe actually runs

- `just check` is the full gate and is exactly what CI enforces. It must pass before you commit.
- Prefer `just <recipe>` over the underlying tool. If you are typing `go test`, you want `just test`.
- Run `just` with stdin from /dev/null. Recipes marked `[confirm]` are destructive — stop and ask
  before running one; never pass `--yes` or `JUST_YES=1`.
- If a task you need does not exist, add a recipe with a `#` doc comment and a `[group(...)]`
  rather than running a bare command.
```

Do **not** paste the recipe list into `AGENTS.md`. `CLAUDE.md` needs no change — it is four lines
and `@AGENTS.md`-includes the above.

### `CONTRIBUTING.md`

| Line | Current | New |
|---|---|---|
| 7 | `- **Make** (for build commands)` | `- **just** (https://just.systems) — `brew install just`` |
| 18 | `make build` | `just setup` then `just build` |
| 42-44 | `scripts/dev-local-plugin.sh install` / `bootstrap` / `status` | `just dev-local-install` / `just dev-local-bootstrap` / `just dev-local-status` |
| 55 | ``- `scripts/dev-local-plugin.sh` for safe install/update/bootstrap testing…`` | ``- `just dev-local-install` / `just dev-local-update` / `just dev-local-bootstrap` for safe testing in an isolated Claude config`` |
| 56 | ``- `scripts/e2e-real-claude.sh` for real-`claude` smoke/manual validation`` | ``- `just e2e-smoke` / `just e2e-manual` for real-`claude` validation`` |
| 57 | ``- `scripts/dev-real-plugin.sh` only when you must validate inside your real `~/.claude` setup`` | ``- `just dev-real-local` (asks for confirmation) only when you must validate inside your real `~/.claude` setup`` |
| 60 | `scripts/e2e-real-claude.sh status` | `just e2e-status` |
| 80-90 | The `## Make Targets` heading and its nine-line `make …` block | Replace the whole section with: heading `## Task Surface`, body ```bash\njust --list\n``` and one sentence: "`just check` is the gate; every recipe carries a doc comment and a group." Do not enumerate recipes — the list rots. |
| 98 | `make test` | `just test` |
| 104-108 | `go test ./internal/analyzer -v` etc. | keep raw `go test ./internal/<pkg>` for package scoping; add `just test TestStateMachine` as the filtered form |
| 111-115 | `### Integration tests` / `go test ./test -v` | **delete** — there is no `test/` directory in this repo; the instruction is stale |
| 125-127 | `go test -run TestStateMachine ./internal/analyzer -v` | `just test TestStateMachine` |
| 129 | `make test-coverage` | `just coverage` |
| 149 | `4. Run tests: `make test-race`` | `4. Run the gate: `just check`` |
| 150 | `5. Run linter: `make lint`` | **delete** — `just check` covers lint |

### `docs/LOCAL_DEVELOPMENT.md`

| Line | Current | New |
|---|---|---|
| 7 | ``Use `scripts/dev-local-plugin.sh` first…`` | ``Use `just dev-local-install` first…`` |
| 9 | ``Use `scripts/dev-real-plugin.sh` only when…`` | ``Use `just dev-real-local` only when…`` |
| 17 | `make build` | `just build` |
| 25-27 | `scripts/dev-local-plugin.sh install` / `bootstrap` / `status` | `just dev-local-install` / `just dev-local-bootstrap` / `just dev-local-status` |
| 33-34 | `scripts/dev-local-plugin.sh update` / `reset` | `just dev-local-update` / `just dev-local-reset` |
| 168-170 | `scripts/dev-real-plugin.sh local` / `status` / `remote` | `just dev-real-local` / `just dev-real-status` / `just dev-real-remote` |
| 176 | `scripts/dev-real-plugin.sh toggle` | `just dev-real-toggle` |
| 186 | ``Prefer `scripts/dev-local-plugin.sh` before touching…`` | ``Prefer `just dev-local-install` before touching…`` |
| 187 | ``Prefer `scripts/e2e-real-claude.sh smoke-plugin-dir`…`` | ``Prefer `just e2e-smoke`…`` |
| 188 | ``Use `scripts/dev-real-plugin.sh local` only for end-to-end checks…`` | ``Use `just dev-real-local` only for end-to-end checks…`` |
| 189 | ``Keep `make build` up to date…`` | ``Keep `just build` up to date…`` |
| 197-199 | `scripts/dev-local-plugin.sh install` / `bootstrap` / `status` | `just dev-local-install` / `just dev-local-bootstrap` / `just dev-local-status` |
| ~202-203 | `scripts/e2e-real-claude.sh smoke-plugin-dir` / `smoke-installed` | `just e2e-smoke` / `just e2e-smoke-installed` |

Sweep the rest of the file for any remaining `scripts/e2e-real-claude.sh <mode>` occurrence and map
it: `status`→`e2e-status`, `smoke-plugin-dir`→`e2e-smoke`, `smoke-installed`→`e2e-smoke-installed`,
`manual-click-plugin-dir`→`e2e-manual`, `manual-click-installed`→`e2e-manual-installed`. The
`REAL_CLAUDE_HOME=… scripts/e2e-real-claude.sh status` form at line 160 becomes
`REAL_CLAUDE_HOME=… just e2e-status` — recipes inherit the step environment.

### `docs/RELEASE.md`

| Line | Current | New |
|---|---|---|
| 41-42 | `make test-race` / `make lint` | a single `just check` |
| 99 | `make build-notifier` | `just build-notifier` |
| 100 | `cd swift-notifier && bash scripts/build-app.sh --ci` | `just build-notifier --ci` |

### Leave alone

- `README.md` — contains no `make` reference. Its two `curl … /main/bin/bootstrap.sh | bash` lines
  (71, 131) are end-user instructions for machines with no checkout and no `just`.
- `docs/troubleshooting.md:75` and `docs/CLICK_TO_FOCUS.md:76` — `curl … /main/scripts/linux-focus-debug.sh | bash`,
  same reason. The raw URL is a public support contract.
- `CHANGELOG.md` — historical record.
- `commands/*.md` — the `make` hits there are the English word "make", not the tool.

## 7. backlog/config.yml

Current (`backlog/config.yml:4-9`):

```yaml
definition_of_done:
  - "go test -race ./..."
  - "go vet ./..."
  - "make build"
  - "bash bin/install_test.sh"
  - "bash bin/install_e2e_test.sh"
```

Replace with:

```yaml
definition_of_done:
  - "just check"
```

All five old entries are dependencies of `check` (`go test -race` → `test`, `go vet` → `vet` via
`lint`, `make build` → `build`, and both install suites → `test-install` / `test-install-e2e`), so
nothing is lost. `backlog/config.yml` is repository configuration, not a task/doc/decision markdown
file, so editing it directly is correct — the "never hand-edit" rule covers `backlog/tasks/`,
`backlog/docs/`, drafts, decisions and milestones only.

## 8. Order of work

The repo stays green at every step. Nothing is deleted until nothing references it.

1. **Install the toolchain.** `brew install just` (or verify `just --version` reports `1.58.0`).
   Confirm `golangci-lint --version` reports `2.12.2` — the same version `ci-ubuntu.yml:105` pins.
2. **Add `justfile`** at the repo root, verbatim from §2. Do not touch `Makefile` yet; the two can
   coexist.
3. **Prove the justfile parses and formats.** All four must pass before going further:
   `just --list`, `just --groups`, `just --dump --dump-format json > /dev/null`,
   `just --fmt --check`.
4. **Prove each recipe locally, on macOS** (the only host that exercises `test-swift` and
   `build-notifier`): `just setup`, `just fmt-check`, `just vet`, `just lint`, `just build`,
   `just smoke`, `just test`, `just test-install`, `just test-install-e2e`, `just test-swift`.
   Then `just check` end to end. Run with stdin from `/dev/null`.
5. **Check the tree is still clean:** `git status --porcelain` must be empty after `just check`.
   If `bin/claude-notifications` shows as modified, the build recipe wrote through the tracked
   symlink — fix the recipe, do not commit the change (Trap 2).
6. **Commit the justfile alone.** CI is untouched and still green.
7. **Switch CI**, one workflow per commit, in this order: `ci-ubuntu.yml`, `ci-macos.yml`,
   `ci-windows.yml`, then `release.yml` and `notifier-signing-smoke.yml`. Push and watch each run go
   green (`gh run watch`) before starting the next. `release.yml` and
   `notifier-signing-smoke.yml` are tag/dispatch-triggered — validate `notifier-signing-smoke.yml`
   with a manual `workflow_dispatch` (`skip_notarize: true`) rather than waiting for a release.
8. **Update the docs and the agent contract:** `AGENTS.md`, `CONTRIBUTING.md`,
   `docs/LOCAL_DEVELOPMENT.md`, `docs/RELEASE.md`, `backlog/config.yml`. Re-run the grep from §6 and
   confirm the only surviving `make`/`scripts/*.sh` hits are the deliberate exclusions listed there.
9. **Only now, delete:** `git rm Makefile setup.sh`. Before doing it, prove nothing references them:
   `grep -rn --include='*.md' --include='*.yml' --include='*.go' -E 'Makefile|make [a-z-]|setup\.sh' . --exclude-dir=.git --exclude=CHANGELOG.md`
   must return nothing but the `swift-notifier/scripts/build-app.sh` hits (unrelated) and CHANGELOG
   history.
10. **Final gate:** `just check` green locally, `git status --porcelain` empty, all three CI
    workflows green on the pushed commit.

## 9. Traps specific to this repo

1. **`make clean` deletes tracked source.** `Makefile:127` is `rm -rf bin/ dist/ coverage.* *.log`.
   `bin/` holds six tracked files (`install.sh`, `bootstrap.sh`, `hook-wrapper.sh`,
   `install_test.sh`, `install_e2e_test.sh`, `mock_server.py`) plus the `claude-notifications`
   symlink. `just clean` must enumerate generated artifacts by name and **never** `rm -rf bin/`. Do
   not "simplify" it back.
2. **`bin/claude-notifications` is a tracked symlink, not a script.** `git cat-file -p HEAD:bin/claude-notifications`
   yields `claude-notifications-darwin-arm64` at mode 120000, and `.gitignore:3` ignores
   `bin/claude-notifications-*`. `Makefile:22` (`go build -o bin/claude-notifications`) writes
   through or over it, which on any host that is not darwin/arm64 replaces a tracked symlink with a
   platform binary and dirties the tree. That is why `just build` writes
   `bin/<cmd>-<goos>-<goarch>[.exe]` and `just smoke` invokes `{{ main_bin }}`. Consequence, stated
   for the record: on a non-darwin/arm64 host the tracked symlink dangles after `just build` — invoke
   the suffixed binary or `just smoke`. That was already true of a fresh clone before this change.
3. **`just --fmt` enforces attribute order and one attribute per line.** Verified on 1.58.0:
   `[group('dev')]` before `[confirm(...)]` fails `--fmt --check`, and `[linux, windows]` on one line
   is rewritten to two. The justfile in §2 is already in canonical form — if you edit it, re-run
   `just --fmt --check`.
4. **Platform-split duplicate recipes are legal and are used twice** (`_setup-swift`, `test-swift`)
   with mutually exclusive `[macos]` / `[linux]` `[windows]` attributes. No `set allow-duplicate-recipes`
   is needed, and `just --list` shows one entry. Verified on 1.58.0.
5. **`go vet` must keep running per-OS.** This repo is full of GOOS-split files
   (`internal/notifier/focus_{darwin,linux,windows,other}.go`, `terminal_*.go`, `bell_*.go`,
   `ax_focus_*.go`, `internal/platform/proc_*.go`, `internal/teamstate/lock_*.go`). `golangci-lint`
   runs only in the ubuntu `lint` job and therefore only sees `GOOS=linux` build tags. Keep the
   `just vet` step in all three test jobs or darwin/windows vet findings go unnoticed.
6. **Linux CGO system dependency.** `malgo` needs `libasound2-dev`. `just setup` must not install it
   (no `sudo`, per the standard); `ci-ubuntu.yml:33` and `release.yml:104` keep their `apt-get`
   steps, and a Linux developer installs it by hand.
7. **golangci-lint version parity.** `just lint` runs whatever `golangci-lint` is on `PATH`;
   `ci-ubuntu.yml:105` pins `v2.12.2` through the action. If they drift, `just check` and CI
   disagree. `.golangci.yml` is v2 schema with `exclusions` only and no `linters.default` override,
   so the default set (`errcheck`, `govet`, `ineffassign`, `staticcheck`, `unused`) applies.
8. **`fmt-check` cannot be `go fmt`.** The current CI trick — run `go fmt ./...` and fail if it
   printed anything — *mutates the tree* before deciding. `fmt-check` uses `gofmt -s -l .` (read-only)
   and `fmt` uses `gofmt -s -w .`. `gofmt -s -l .` is clean on the tree as of this task, so `-s`
   introduces no churn.
9. **`test-install-e2e` is slow and needs `python3`.** `bin/install_e2e_test.sh` is 81 KB, spawns
   `bin/mock_server.py`, and writes `bin/test_fixtures/` (gitignored). It is inside `check` because
   CI runs it. `--real-network` is diagnostics only and stays `continue-on-error` in CI — never add
   it to `check`.
10. **`verify-version` and `build-release` need a persistent shell**, so both use `[script('bash')]`:
    the first defines a `check_version()` function and uses `[[ =~ ]]`; the second uses a `for` loop
    over four commands. A line-based recipe would fail with "extra leading whitespace" on the
    multi-line constructs. `[script('bash')]` is stable in 1.58 — no `--unstable`.
11. **`{{ }}` inside `verify-version`'s `awk` program.** The awk body uses single braces only
    (`{ found = 1 }`), which `just` does not interpolate. If you edit it, never let two braces become
    adjacent.
12. **`require("golangci-lint")` is evaluated when the recipe runs**, not on `just --list` or
    `just --dump`, so a machine without golangci-lint can still enumerate the task surface. It halts
    `just lint` with a clear message instead of a bare exit 127.
13. **`go test -run ''` runs every test** — verified. That is what makes the optional `filter`
    parameter work without a conditional expression.
14. **Windows CI runs the recipes under Git Bash.** `set shell` applies on Windows because no
    `set windows-shell` is declared. Deleting the `shell: pwsh` line from `ci-windows.yml:49` is
    required, not optional — leaving it would run `just fmt-check` under PowerShell's argument
    parsing.
15. **`scripts/linux-focus-debug.sh` and `bin/bootstrap.sh` are fetched by URL.** `README.md:71`,
    `README.md:131`, `docs/troubleshooting.md:75` and `docs/CLICK_TO_FOCUS.md:76` point at
    `raw.githubusercontent.com/.../main/<path>`. Moving either file breaks live user instructions.
16. **`scripts/iterm2-select-tab.py` is resolved by Go at runtime** (`internal/notifier/tmux_iterm2.go:29`)
    and asserted by tests at `internal/notifier/notifier_test.go:1446,1470,1629` and
    `internal/notifier/tmux_iterm2_test.go:176,260`. It gets no recipe and must not move.
17. **`.gitattributes` has no `justfile` entry.** It forces `eol=lf` on `*.sh`/`*.yml`/`*.md` but an
    extensionless `justfile` falls through to `* text=auto`. Optionally add `justfile text eol=lf`
    for Windows contributors; harmless either way.
18. **No `ci-success` aggregator and no CodeQL/zizmor/actionlint/scorecard/dependency-review/
    release-please workflows exist in this repo.** Do not add any as part of this task, and do not
    rename existing jobs — the required-check names are those job names.

## 10. Out of scope

Do not touch, in this task:

- **KEEP scripts, contents unchanged:** `bin/install.sh`, `bin/bootstrap.sh`, `bin/hook-wrapper.sh`,
  `bin/install_test.sh`, `bin/install_e2e_test.sh`, `bin/mock_server.py`, `bin/claude-notifications`
  (symlink), `scripts/dev-local-plugin.sh`, `scripts/dev-real-plugin.sh`,
  `scripts/e2e-real-claude.sh`, `scripts/linux-focus-debug.sh`, `scripts/iterm2-select-tab.py`,
  `swift-notifier/scripts/build-app.sh`. Recipes wrap them; the files are not edited.
- **Workflows:** `.github/workflows/reviewrouter.yml` and
  `.github/workflows/reviewrouter-interaction.yml` — reusable `uses:` calls only, byte-identical.
  Never convert any `uses:` to `run: just`: that covers `actions/checkout@v7`,
  `actions/setup-go@v7`, `codecov/codecov-action@v7`, `golangci/golangci-lint-action@v9`,
  `actions/upload-artifact@v7`, `actions/download-artifact@v8`, `softprops/action-gh-release@v3`.
- **`release.yml`'s** `verify macOS deployment target`, `create-release` and `test-binaries` jobs, and
  its Apple certificate import steps.
- **Go, Swift or shell source behaviour.** No `internal/`, `cmd/`, `pkg/` or `swift-notifier/Sources`
  changes. No `.golangci.yml` changes.
- **New gates.** No `shellcheck`, no `govulncheck`/`audit`, no `swift-format`, no `gen`/`gen-check`,
  no `infra` group — this repo has no generated committed artifacts and no infrastructure code, and
  adding a gate CI does not run is a separate decision.
- **Version bumps, upstream sync, release mechanics.** No touching
  `.claude-plugin/plugin.json`, `.claude-plugin/marketplace.json`, `CHANGELOG.md` or `go.mod`.
- **End-user documentation URLs.** `README.md:71,131`, `docs/troubleshooting.md:75`,
  `docs/CLICK_TO_FOCUS.md:76`.
- **Adding a `ci-success` aggregator job**, SHA-pinning the pre-existing tag-pinned actions, or
  renaming any workflow or job.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A top-level justfile exists and just --list shows default, setup, fmt, fmt-check, lint, test and check, with a # doc comment and a group of check/build/dev/release on every public recipe; just --fmt --check and just --dump --dump-format json both exit 0 under just 1.58.0
- [ ] #2 just check passes on a macOS host with git status --porcelain empty afterwards, and its dependency list covers fmt-check, vet, golangci-lint, go test -race, build, smoke, bin/install_test.sh, bin/install_e2e_test.sh and swift test --package-path swift-notifier
- [ ] #3 Makefile and setup.sh are removed with git rm, and no 'make <target>' or 'bash setup.sh' reference survives anywhere outside CHANGELOG.md
- [ ] #4 just build writes bin/<cmd>-<goos>-<goarch>[.exe] and never modifies the tracked bin/claude-notifications symlink, and just clean never runs rm -rf bin/ so install.sh, bootstrap.sh, hook-wrapper.sh, install_test.sh, install_e2e_test.sh and mock_server.py all survive it
- [ ] #5 Every KEEP script still exists byte-identical and is reachable through a recipe: scripts/dev-local-plugin.sh via just dev-local-*, scripts/dev-real-plugin.sh via just dev-real-*, scripts/e2e-real-claude.sh via just e2e-*, scripts/linux-focus-debug.sh via just linux-focus-debug, bin/install_test.sh via just test-install, bin/install_e2e_test.sh via just test-install-e2e, swift-notifier/scripts/build-app.sh via just build-notifier; scripts/iterm2-select-tab.py and bin/mock_server.py remain in place with no recipe
- [ ] #6 ci-ubuntu.yml, ci-macos.yml and ci-windows.yml each carry a SHA-pinned extractions/setup-just step with just-version '1.58.0' and call just setup, just vet, just fmt-check, just test, just smoke, just test-install and just test-install-e2e; the ci-macos swift-test job calls just test-swift; all three workflows are green on the pushed commit
- [ ] #7 release.yml calls just verify-version "$RELEASE_TAG", just build-release <platform> <arch> and just build-notifier --ci, and notifier-signing-smoke.yml calls just build-notifier --ci with the optional --skip-notarize; no job or workflow was renamed and golangci-lint-action, codecov-action and both reviewrouter reusable uses: calls are unchanged
- [ ] #8 AGENTS.md names just check as the repository gate and carries a '## Task interface' section that does not enumerate recipes; CONTRIBUTING.md, docs/LOCAL_DEVELOPMENT.md and docs/RELEASE.md contain no make target or scripts/*.sh invocation, while the end-user curl URLs in README.md, docs/troubleshooting.md and docs/CLICK_TO_FOCUS.md are unchanged
- [ ] #9 backlog/config.yml definition_of_done is exactly one entry, 'just check'
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 go test -race ./...
- [ ] #2 go vet ./...
- [ ] #3 make build
- [ ] #4 bash bin/install_test.sh
- [ ] #5 bash bin/install_e2e_test.sh
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Establish the exact Makefile/script/workflow/doc mapping and preserve the task-defined immutable files.
2. Add the canonical justfile, validate its parse/format and the macOS recipe surface, then move CI build/test/lint/generate commands to one-line just recipes while preserving actions, job names, and triggers.
3. Update the task contract, contributor docs, and Backlog gate; remove Makefile and setup.sh only after repository-wide reference and hook sweeps.
4. Run the required local gates and workflow/static validation, commit named paths to main, push, and capture CI evidence at the final SHA.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented the just surface and workflow migration in commits ba0702d, 1eb82cf, 2fb7fc1, 6dda2a1, acca3a1, and 4209b93.

Final local validation on macOS passed: just --fmt --check, just --dump --dump-format json, and just check (Go race suite, installer unit tests 16/0, installer E2E 91/0 with 12 explicit skips, and Swift tests 30/0). actionlint passed for every workflow.

A retained Unix wrapper assertion would fail on Linux after the safe suffixed build output, so test-install now provides a temporary ignored wrapper target only during that recipe; retained runtime scripts remain byte-identical.

The full zizmor scan reports pre-existing tag-pinned action and permission findings; task scope explicitly preserves those actions and permission blocks. The new setup-just action is SHA-pinned.

GitHub REST is currently rate-limited for the shared unauthenticated address, so exact final CI run IDs remain pending.

Parked at the mandatory CodeRabbit pre-commit gate. The final staged migration diff is locally validated but must not be committed without review.

CodeRabbit final-diff attempts were all organisation-plan rate-limited: first reported a 15-minute wait; evidence-based retries reported 1 minute, 18 seconds, then 7 minutes. No login, installation, or quota bypass was attempted.

Resume boundary: wait for the review allowance, run `coderabbit review --agent` against the existing staged paths, triage any findings, then commit the reviewed named paths, push main, and capture green exact-SHA CI before finalizing the task.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: campaign-ordering
created: 2026-08-29 09:18
---
## Fleet ordering — WAVE 2. Starts after the Wave 0 pilot (`sf2loki` / SFL-0073) and the Wave 1 hubs land.

Within Wave 2 the order is free — these repos do not depend on each other. Batching by language is worthwhile so one lane reuses its Makefile-to-recipe mapping across similar repos.

Do not start before the pilot reports. The standard may be amended off the back of it, and picking this up early risks coding against a superseded seam.

**Provisioning `just` in CI.** Which mechanism depends on the runner, and the two must not be mixed:

| Runner | Mechanism |
| --- | --- |
| `arc-arm64` (m7kni self-hosted) | `just` is **baked into the runner image** by `m7kni/ci-tools` (`runner-image/Dockerfile`, `ARG JUST_VERSION`). Do **not** add `extractions/setup-just`, and delete the step if this repo already has one — it installs a second `just` earlier on `PATH` and turns the image pin into a lie. |
| GitHub-hosted (all `rknightion` repos) | `extractions/setup-just`, SHA-pinned, with an explicit `just-version:`. |

Both sides currently sit on **1.58.0** and are Renovate-managed. `ci-tools`' `Tool version drift` workflow fails if the Dockerfile `ARG` and the published image ever disagree, and lists any repo still carrying a second pin.

**While you are in the workflow files, check the hub pin.** On 2026-08-29 Renovate was unfrozen for `rknightion/.github` in `m7kni/renovate-config` — it had been `enabled: false` on the mistaken belief that callers tracked `@main`, which froze the fleet across 19 different hub SHAs (v1.3.1 June → v1.9.7 August) so that no hub fix ever propagated. Bumps now arrive as one grouped, CI-gated, automerged PR per repo. **A `uses:` whose comment is not a real `# vX.Y.Z` still cannot be bumped** (it resolves to a digest-only update, which the fleet rules disable) — if you find one, repair the comment as part of this task.
---

author: campaign-ordering
created: 2026-08-29 10:42
---
## Standard amendment — `ci` is the sanctioned superset of `check` (RATIFIED)

This supersedes the frozen wording *"`check` is the complete local gate and reproduces every CI job that can run off a GitHub runner"*, which several lanes could not honour without making the pre-commit gate depend on a Docker daemon.

**The definitions now are:**

- **`check`** — everything that runs with **only the language toolchain installed**. This is the pre-commit gate. A leg that runs on a bare toolchain belongs here *however long it takes*.
- **`ci`** — `check` plus the legs CI gates that need a **Docker daemon, a service container, or cross-compilation**, and nothing else. Written as `ci: check <heavy legs>`.

**Every leg you put in `ci` must carry a comment naming which of those three it needs.** That comment is the guard: without it `ci` becomes the bin for anything slow or awkward, `check` quietly stops meaning much, and the fleet is back to a per-repo gate.

Eleven of the 42 lanes arrived at this shape independently before it was ratified, which is why it won.

**If this repo has no such legs, it has no `ci` recipe at all** and `check` is the whole gate. Do not add an empty one.
---
<!-- COMMENTS:END -->
