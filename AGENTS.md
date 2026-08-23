
<!-- BACKLOG.MD GUIDELINES START -->
<!-- backlog.md-instructions-version: 1.50.1 -->
<CRITICAL_INSTRUCTION>

## Backlog.md Workflow

This project uses Backlog.md for task and project management.

**For every user request in this project, run `backlog instructions overview` before answering or taking action.**

Use the overview to decide whether to search, read, create, or update Backlog tasks.

Before task lifecycle actions, read the matching detailed guide:
- `backlog instructions task-creation` before creating or splitting tasks
- `backlog instructions task-execution` before planning, changing status or assignee, adding a plan or implementation notes, or implementing task work
- `backlog instructions task-finalization` before checking acceptance criteria, writing final summaries, or moving tasks to terminal statuses

Use `backlog <command> --help` before running unfamiliar commands. Help shows options, fields, and examples.

Do not edit Backlog task, draft, document, decision, or milestone markdown files directly. Use the `backlog` CLI so metadata, relationships, and history stay consistent.

</CRITICAL_INSTRUCTION>
<!-- BACKLOG.MD GUIDELINES END -->

# Repository-specific operating rules

- This is a public maintained fork of `777genius/claude-notifications-go`. Keep `upstream` configured and preserve changes as reviewable commits that can be rebased conceptually onto future upstream releases without rewriting published history.
- Work directly on `main`; do not create pull requests for `rknightion/claude-notifications-go`.
- Never put personal identifiers, private profile paths, credentials, internal hostnames, account IDs, or fleet-specific secrets in `backlog/`, documentation, tests, fixtures, or logs. Use generic temporary paths and profile names.
- Drive tasks and durable documents through the Backlog CLI. Never hand-edit task, draft, document, decision, or milestone Markdown.
- Use `go test -race ./...`, `go vet ./...`, `make build`, `bash bin/install_test.sh`, and `bash bin/install_e2e_test.sh` as the repository gate. Run narrower tests while iterating.
- Treat Claude Code and Codex hook payloads as external contracts. Pin behavior with focused tests before changing adapters, configuration resolution, installation, or hook registration.
