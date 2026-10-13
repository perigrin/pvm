---
title: Module installs leave a build directory behind on every failure, and tests fill the real cache
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-13T20:17:30.813449537Z
updated: 2026-10-13T20:17:30.813449537Z
---

## Context

- paths: `internal/pm/modules/installer.go:416-430` (creates `<cache>/build/modules/<Module>-<timestamp>`), `internal/pm/modules/builder.go:500-513` (the only removal, gated on `options.Cleanup`), `test/e2e/` (installs against the real XDG cache)
- commands: `make test`; `ls ~/.cache/pvm/build/modules | wc -l` before and after

Found 2026-10-13 when the shared disk reached 89%: `~/.cache/pvm/build` had grown to 37G, almost all of it 19,356 per-install directories under `build/modules` dating back to May 2025 (perigrin had the ones older than a day deleted). Two defects:

1. The installer makes a timestamped build directory for every install, but only the builder removes it, and only when `Cleanup` is set and the build step is reached. An install that fails or returns early -- dependency resolution, download, extraction (e.g. the tests' `Nonexistent-Module-That-Does-Not-Exist-*`) -- leaves it behind, and nothing prunes old ones.
2. The tests run installs against the user's real `~/.cache/pvm` rather than a per-test temp directory, so every `make test` adds on the order of a thousand directories (1,370 in two days).

## Steps

- [ ] Failing test first: an install that fails after creating its build directory leaves nothing in `build/modules`; remove the directory on every exit path (defer), unless a keep-build option is set
- [ ] Tests that install point XDG_CACHE_HOME (or the pvm cache override) at `t.TempDir()`; a test asserts `make test` leaves `~/.cache/pvm/build/modules` unchanged
- [ ] `make test`; commit

## Acceptance criteria

- [ ] a failed install removes its build directory (`go test ./internal/pm/modules/ -run '^TestFailedInstallRemovesBuildDir$' -count=1 -v | grep -q '^--- PASS: TestFailedInstallRemovesBuildDir'`)
- [ ] a successful install without keep-build removes its build directory (`go test ./internal/pm/modules/ -run '^TestInstallRemovesBuildDir$' -count=1 -v | grep -q '^--- PASS: TestInstallRemovesBuildDir'`)
