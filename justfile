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
[script('bash')]
test-install: build
    placeholder=""
    if [ "{{ goos }}" = "linux" ]; then
        wrapper_target="$(readlink bin/claude-notifications)"
        if [ ! -e "bin/$wrapper_target" ]; then
            placeholder="bin/$wrapper_target"
            : > "$placeholder"
        fi
    fi
    trap 'rm -f "$placeholder"' EXIT
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
