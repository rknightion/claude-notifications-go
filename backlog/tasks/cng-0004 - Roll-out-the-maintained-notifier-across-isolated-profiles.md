---
id: CNG-0004
title: Roll out the maintained notifier across isolated profiles
status: To Do
assignee: []
created_date: '2026-08-23 17:39'
labels: []
dependencies:
  - CNG-0001
  - CNG-0002
  - CNG-0003
ordinal: 4000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Replace the removed upstream notification plugin with the maintained fork across the isolated Claude and Codex profile fleet after the fork supports each runtime. Publication must originate from the neutral profile source and preserve independent profile homes and credentials.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 All three Claude profiles install the maintained fork and keep independent config and helper state under their own CLAUDE_CONFIG_DIR
- [ ] #2 All three Codex profiles install the maintained fork and keep independent config and state under their own CODEX_HOME
- [ ] #3 Neutral profile validation rejects upstream notification artifacts, shared default homes, and cross-profile notifier state
- [ ] #4 A real notification is exercised in each runtime and profile class on at least one supported Mac, with unexercised fleet machines reported separately
- [ ] #5 Profile backup hooks publish the resulting profile and neutral-source state without manual commits to backup repositories
<!-- AC:END -->

## Definition of Done
<!-- DOD:BEGIN -->
- [ ] #1 go test -race ./...
- [ ] #2 go vet ./...
- [ ] #3 make build
- [ ] #4 bash bin/install_test.sh
- [ ] #5 bash bin/install_e2e_test.sh
<!-- DOD:END -->
