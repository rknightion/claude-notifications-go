---
id: CNG-0001
title: Respect profile-specific Claude configuration roots
status: Done
assignee:
  - '@codex'
created_date: '2026-08-23 17:30'
updated_date: '2026-08-23 18:13'
labels: []
dependencies: []
references:
  - 'https://github.com/777genius/claude-notifications-go/issues/43'
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Resolve durable notification configuration and Claude-owned runtime paths from the active Claude Code configuration root so isolated Claude profiles never read from or write to a shared default home. Preserve the default ~/.claude behavior when no supported override is set.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 CLAUDE_CONFIG_DIR takes precedence for the stable notification config directory
- [x] #2 The compatibility CLAUDE_HOME override remains supported after CLAUDE_CONFIG_DIR and before the default home
- [x] #3 The settings wizard, runtime config loading, team discovery, and iTerm helper paths use the same resolved Claude root
- [x] #4 Focused tests fail on the previous shared-home behavior and pass for override and default cases
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
1. Add failing tests that pin CLAUDE_CONFIG_DIR, legacy CLAUDE_HOME, and default ~/.claude precedence for stable config, team discovery, and iTerm marker paths.

2. Centralize Claude configuration-root resolution in the Go config package and route all runtime consumers through it.

3. Update bootstrap/install helpers and plugin command instructions so setup, settings, sounds, and iTerm assets use the same active profile root.

4. Update user documentation, run focused checks, then run the repository gate and review the final diff.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Added regression tests that failed against the previous shared-home behavior, centralized active Claude root resolution, and routed runtime config, team discovery, iTerm markers, installers, commands, and documentation through the profile-aware precedence.

Verification: go test -race ./... passed; go vet ./... passed; make build passed; bash bin/install_test.sh passed 16/16 after the first pre-build invocation correctly failed on the missing ignored wrapper; bash bin/install_e2e_test.sh passed 91/91 with 12 explicitly skipped platform or real-network checks. git diff --check passed.

Resume boundary: coderabbit review --agent --include-untracked failed before analysis with a recoverable WebSocket closed error on three attempts. Policy requires that review before commit. Retry when the CodeRabbit service is reachable, address any Critical or Warning findings, then read the task-finalization guide, finalize CNG-0001, commit, and push main.

CodeRabbit retry on 2026-08-23 connected successfully, confirmed the repository used the rknightion organization plan, completed setup/tool analysis, and entered reviewing, then failed before emitting findings with recoverable error: WebSocket subscription completed unexpectedly. This was the fourth failed review attempt overall. No review findings were available to address.

CodeRabbit completed on the rknightion organization plan with six findings. Addressed all six: exported the resolved installer root, quoted paths containing profile expansions, made README path guidance platform-neutral, added make build to the declared gate and affected tasks, and sanitized the locally imported canonical document pending CNG-0005 source reconciliation. No findings were dismissed.

Final gate after review fixes: go test -race ./... passed; go vet ./... passed; make build passed; bash bin/install_test.sh passed 16/16; bash bin/install_e2e_test.sh passed 91/91 with 12 explicit platform or real-network skips; bash -n and git diff --check passed.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Implemented profile-aware Claude configuration-root resolution across runtime configuration, team discovery, iTerm helper state, installers, plugin commands, and documentation. Regression tests cover CLAUDE_CONFIG_DIR precedence, legacy CLAUDE_HOME compatibility, default behavior, team discovery, and iTerm marker paths. CodeRabbit completed with six findings and all were addressed. The full repository gate passed; 12 optional platform or real-network E2E legs were skipped and remain unproven.

Implementation commit: 1e7098010525d501f7ec92edad310c0c56ee226c.
<!-- SECTION:FINAL_SUMMARY:END -->
