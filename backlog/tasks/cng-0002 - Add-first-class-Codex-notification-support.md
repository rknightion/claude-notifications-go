---
id: CNG-0002
title: Add first-class Codex notification support
status: Done
assignee:
  - '@codex'
created_date: '2026-08-23 17:30'
updated_date: '2026-08-23 18:48'
labels: []
dependencies:
  - CNG-0001
references:
  - 'https://developers.openai.com/codex/hooks'
  - 'https://developers.openai.com/codex/config-advanced#notifications'
ordinal: 2000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Run the notification engine from Codex lifecycle hooks while retaining native Claude Code behavior. Codex installations must use CODEX_HOME for profile-local writable state and consume documented hook payloads without depending on Claude transcript formats.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 Codex plugin hooks register only supported Codex lifecycle events and use Codex plugin compatibility variables safely
- [x] #2 Codex hook payloads produce task-complete, review-complete, question, plan-ready, session-limit, and API-error notifications where the available event data proves each status
- [x] #3 Codex writable configuration and state resolve under CODEX_HOME without reading or creating ~/.codex
- [x] #4 Installation, uninstallation, and documentation cover both Claude Code and Codex without overwriting unrelated configuration
- [x] #5 Unit and integration tests cover Claude compatibility and representative Codex hook payloads
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [x] #1 go test -race ./...
- [x] #2 go vet ./...
- [x] #3 make build
- [x] #4 bash bin/install_test.sh
- [x] #5 bash bin/install_e2e_test.sh
<!-- DOD:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Add focused contract tests for the Claude and Codex hook registries, documented Codex payloads, runtime argument parsing, and CODEX_HOME isolation; confirm the new tests fail for the intended missing behavior.

2. Add an explicit runtime boundary to configuration, writable state, logging, and hook handling. Preserve Claude transcript analysis unchanged; classify Codex UserPromptSubmit, PermissionRequest, PostToolUse, and Stop payloads only from documented fields and profile-local state under CODEX_HOME.

3. Package the repository for both runtimes with separate hook registries, a valid Codex manifest/marketplace surface, a compatibility-safe wrapper using PLUGIN_ROOT/PLUGIN_DATA aliases, and installation/uninstallation/configuration documentation that never overwrites unrelated Codex config.

4. Run targeted tests and manifest/script validation, review the final diff, run the required CodeRabbit gate, repair any Critical/Warning findings with fresh checks/review, then run the complete repository gate.

5. Read the Backlog finalization guide, attach objective evidence to every criterion and Definition of Done item, finalize CNG-0002 only, explicitly stage the owned files, commit directly to main, and push origin/main.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Official Codex docs checked on 2026-08-23: lifecycle hooks receive JSON on stdin; plugin hooks expose PLUGIN_ROOT/PLUGIN_DATA plus Claude compatibility aliases; Stop includes permission_mode and last_assistant_message; notify is a separate single-JSON-argument surface; CODEX_HOME owns config/state. Selected solo topology because hooks, config, wrapper, manifests, and installer docs are serialized integration surfaces.

Focused verification passed: representative Codex UserPromptSubmit, PermissionRequest, PostToolUse, and Stop payloads produce task_complete, review_complete, question, plan_ready, session_limit_reached, api_error, and api_error_overloaded without reading transcript_path; Claude hook tests remain green. CODEX_HOME isolation tests prove config and state are created only below the explicit Codex profile and ~/.codex is not created. The current Codex CLI successfully completed marketplace add, plugin add, and plugin remove in a throwaway CODEX_HOME; the plugin validator and JSON/shell syntax checks passed. CodeRabbit used the rknightion organization plan; its valid API-error false-positive finding was fixed and regression-tested. Its hooks-manifest finding was rejected because current default discovery loads hooks/hooks.json and the current validator rejects a manifest hooks field; its minor ~/.codex documentation suggestion conflicts with CNG-0002's explicit CODEX_HOME isolation contract. Final gate: go test -race ./... PASS; go vet ./... PASS; make build PASS; bash bin/install_test.sh 16 passed/0 failed; bash bin/install_e2e_test.sh 91 passed/0 failed/12 skipped. Skips were the named stale-lock timing, non-macOS platform, and opt-in real-network legs and were not counted as passes.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added first-class Codex notifications through a separate supported-event registry and documented payload adapter while preserving the original Claude Code hook registry and transcript analyzer. Codex configuration, logs, deduplication, and prompt state are rooted under explicit CODEX_HOME, and marketplace install/remove leaves unrelated config.toml untouched. Verified with focused contract/isolation tests, a throwaway Codex marketplace install/remove, current plugin validation, two CodeRabbit reviews, and the complete repository gate; optional real-network and unavailable-platform E2E legs remain explicitly skipped.
<!-- SECTION:FINAL_SUMMARY:END -->
