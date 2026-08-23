---
id: CNG-0005
title: Onboard the fork to canonical document drift checks
status: To Do
assignee: []
created_date: '2026-08-23 17:40'
updated_date: '2026-08-23 18:11'
labels: []
dependencies: []
ordinal: 5000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Register the fork canonical fan-out document with m7kni/agent-docs and add the public repository to the existing rknightion drift permission set. The local document already exists; central onboarding was deferred because no current OpenBao admin token was available.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 agent-docs registry maps rknightion/claude-notifications-go to doc-0001
- [ ] #2 The agent-docs-drift-rknightion permission set contains claude-notifications-go with contents read only
- [ ] #3 bin/doctor and bin/doctor-remote both verify the fork document
- [ ] #4 Changes are committed and pushed in both repositories with explicit staging
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 go test -race ./...
- [ ] #2 go vet ./...
- [ ] #3 make build
- [ ] #4 bash bin/install_test.sh
- [ ] #5 bash bin/install_e2e_test.sh
<!-- DOD:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
The initial public-fork import was locally sanitized through Backlog to remove machine-specific Claude profile paths from Appendix B. Canonical onboarding must reconcile that generic wording with agent-docs before registering the consumer, so the first rendered copy does not reintroduce the private path.
<!-- SECTION:NOTES:END -->
