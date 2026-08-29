---
id: doc-0002
title: Wave operating model
type: specification
created_date: '2026-08-23 17:30'
updated_date: '2026-08-29 13:32'
---
# Wave operating model — claude-notifications-go

This document carries only repository-specific additions to the **Agent fan-out protocol (canonical)** document. Read the canonical document first. Nothing here changes its ownership, evidence, or blocker contracts.

## 1. Repository boundaries

### Preserve the upstream seam

This repository is a maintained public fork of `777genius/claude-notifications-go`. The `upstream` remote is the source of vendor updates; `origin` is the maintained distribution. Fork-specific changes stay in focused commits so an upstream merge can distinguish imported work from local product direction. Published history is never rewritten to make later synchronization look cleaner.

A compatibility fix that applies to ordinary upstream users may be offered upstream. Fleet integration, private topology, credentials, and profile names never belong in an upstream issue or pull request.

### Runtime roots are a contract

Claude Code owns `CLAUDE_CONFIG_DIR`; the legacy `CLAUDE_HOME` override is a compatibility layer; `~/.claude` is only the final default. All durable settings, helper assets, team discovery, installer paths, and plugin-command fallbacks must resolve that precedence through the same contract. No component may independently reconstruct `~/.claude`.

Codex owns `CODEX_HOME`. Codex support must not reuse a Claude root merely because the hook shapes are compatible. Cross-runtime code may share notification analysis and delivery, but it must receive an explicit runtime and runtime root at the boundary.

### Hook payloads are external interfaces

Claude Code and Codex hook JSON are vendor contracts. Keep representative fixtures and focused adapter tests. Unknown fields must remain harmless; missing optional fields must not become proof of a status. Transcript files are useful evidence but are not automatically stable interfaces, especially for Codex.

## 2. Recurring defects and likely traps

**Bootstrap and installer drift.** `bin/bootstrap.sh` and `bin/install.sh` both contain setup behavior. A path or release-channel change applied to only one passes some installation routes and fails others. Search both files and the hook wrapper before calling installer work complete.

**Plugin commands are executable documentation.** Files under `commands/` contain shell fragments that users and agents execute. A runtime fix without the matching settings, init, and sounds command updates recreates the old behavior during the next configuration run.

**Stable data must stay outside replaceable plugin caches.** Plugin updates may delete and recreate cache directories. Configuration and writable helper state belong under the active runtime data root; bundled sounds, icons, scripts, and binaries belong under the plugin root.

**The installer unit test runs through its recipe.** `just test-install` builds the host binaries first and, on Linux, supplies a temporary ignored target so the retained Unix wrapper assertion remains valid. Use it (or `just check`) instead of calling the script directly.

**Platform evidence does not transfer.** A macOS pass proves neither Windows toast activation nor Linux compositor focus. The E2E suite reports platform and real-network skips separately; preserve those as skipped, not passed.

## 3. Ownership and integration files

Natural lanes are Go packages under `internal/`, platform-specific notifier files, documentation, and installer tests when their seams are frozen.

Serialize edits to these integration surfaces:

- `internal/hooks/`, which composes analyzers, state, notification delivery, and runtime payloads;
- `internal/config/`, which owns configuration schema and runtime-root resolution;
- `hooks/hooks.json` and `.claude-plugin/`, which define installed lifecycle behavior;
- `bin/bootstrap.sh`, `bin/install.sh`, and `bin/hook-wrapper.sh`, which form one installation/update channel;
- `commands/`, whose repeated shell fragments must move together;
- `go.mod` and `go.sum`;
- release version sources and `CHANGELOG.md`.

A lane that needs one of these but does not own it returns a diff-shaped handoff. It does not add a second resolver, hook registry, version source, or installer path.

## 4. Verification and delivery

Use package-level tests while iterating. The repository gate is:

```bash
just check
```

Run platform-specific Swift, Windows, Linux, real-network, and manual click-to-focus checks when the changed surface requires them. State every unexercised platform or optional network leg explicitly.

Completed changes go directly to `main` on this fork after the required review gate. Upstream contributions, when justified, are separate and never block the maintained fork from serving its own release channel.

## 5. Tracker run-end

A completed task records the exact commit and the checks actually observed. A parked task names the vendor contract, platform, credential, or decision needed to resume. Work discovered in another runtime or installation surface becomes a separate task rather than expanding a focused compatibility fix mid-review.
