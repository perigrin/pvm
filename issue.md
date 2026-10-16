---
title: "[I30] incRoots finds pvm.toml from the process cwd, not from the target"
state: in-progress
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:08.402897704Z
updated: 2026-10-16T01:17:15.405837282Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: ""
  commits: 0
  started_at: 2026-10-16T01:17:15.405837282Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:15.405837282Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I30**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/psc/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/perl/resolver.go:553`: `resolveFromProjectConfig` ignores its `projectDir` parameter and calls `config.GetProjectRoot()`, which walks up from `os.Getwd()`.
- **Impact:** an LSP started from `$HOME` and `cd proj && psc check` resolve different perls, so they get a different @INC, a different parse and different diagnostics for the same file. That breaks the check==LSP contract. It also widens C1 and C2, since the cwd's pvm.toml is read even for `psc check /other/dir`.
- **Fix:** have resolveFromProjectConfig look for the project root from `options.ProjectDir`.
- **Confidence:** High, by reading the code; I did not reproduce the cwd split. **NEW in reach.** **Issue:** none. **Found by:** contract-B #7.

---

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] incRoots finds `pvm.toml` from the process cwd, not from the target is fixed (`go test ./internal/psc/ -run '^TestIncRootsFromTargetProject$' -count=1 -v | grep -q '^--- PASS: TestIncRootsFromTargetProject'`)
