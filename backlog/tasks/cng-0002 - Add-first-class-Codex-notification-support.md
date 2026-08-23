---
id: CNG-0002
title: Add first-class Codex notification support
status: To Do
assignee: []
created_date: '2026-08-23 17:30'
updated_date: '2026-08-23 18:11'
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
- [ ] #1 Codex plugin hooks register only supported Codex lifecycle events and use Codex plugin compatibility variables safely
- [ ] #2 Codex hook payloads produce task-complete, review-complete, question, plan-ready, session-limit, and API-error notifications where the available event data proves each status
- [ ] #3 Codex writable configuration and state resolve under CODEX_HOME without reading or creating ~/.codex
- [ ] #4 Installation, uninstallation, and documentation cover both Claude Code and Codex without overwriting unrelated configuration
- [ ] #5 Unit and integration tests cover Claude compatibility and representative Codex hook payloads
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 go test -race ./...
- [ ] #2 go vet ./...
- [ ] #3 make build
- [ ] #4 bash bin/install_test.sh
- [ ] #5 bash bin/install_e2e_test.sh
<!-- DOD:END -->
