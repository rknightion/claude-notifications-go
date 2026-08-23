---
id: CNG-0003
title: Own the fork release and update channel
status: To Do
assignee: []
created_date: '2026-08-23 17:39'
labels: []
dependencies: []
ordinal: 3000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Make rknightion/claude-notifications-go a self-contained distribution rather than a source fork whose bootstrap, plugin commands, diagnostics, and release metadata still download or link to upstream artifacts. Preserve upstream as the synchronization remote.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Bootstrap, installer, hook-wrapper updates, plugin commands, and user-facing install URLs default to the maintained fork
- [ ] #2 Release automation builds all supported binaries and either signs the macOS notifier with provisioned fork credentials or documents and enforces an explicit unsigned boundary
- [ ] #3 Version sources, checksums, manifests, and update detection agree for a fork release
- [ ] #4 Tests reject an accidental return to upstream release assets in the maintained distribution
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 go test -race ./...
- [ ] #2 go vet ./...
- [ ] #3 make build
- [ ] #4 bash bin/install_test.sh
- [ ] #5 bash bin/install_e2e_test.sh
<!-- DOD:END -->
