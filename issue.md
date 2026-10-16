---
title: "[C1] psc check and the LSP run /usr/bin/perl -v when a version file says system"
state: done
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:06.94921152Z
updated: 2026-10-16T03:25:03.526610117Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: 29abad9cb87176fb5e3e8aa8e113ffd4e0fad78b
  commits: 3
  started_at: 2026-10-16T01:17:15.139938436Z
  ended_at: 2026-10-16T03:25:03.526610117Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:15.139938436Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T03:25:03.526610117Z
observed_paths:
- docs/plans/2026-09-05-parser-conformance-plan.md
- docs/plans/2026-09-21-graded-conformance-corpus.md
- docs/specs/perl-parser/04-expressions.md
- docs/specs/perl-parser/05-statements.md
- docs/specs/perl-parser/06-incremental-lsp.md
- internal/infer/calls.go
- internal/infer/flow.go
- internal/infer/flow_test.go
- internal/infer/infer.go
- internal/lower/node.go
- internal/lower/positions.go
- internal/parse/loader_regular_test.go
- internal/parse/module.go
- internal/parse/staysinroots_test.go
- internal/perl/inc.go
- internal/perl/inc_test.go
- internal/perl/static.go
- internal/perl/static_test.go
- internal/psc/check_command.go
- internal/psc/check_command_test.go
- internal/psc/check_escape_test.go
- internal/psc/inc_roots_test.go
- internal/psc/lsp.go
- internal/psc/resolve_guard_test.go
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **C1**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/psc/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/psc/check_command.go:99-105` (incRoots) → `perl.ResolveVersion` → `internal/perl/version.go:285` (`case "system": return GetSystemVersionString()`) → DetectSystemPerl → exec.
- **Bug:** `SkipSystemPerl` gates only resolution step 7. The alias table in ResolveVersionAlias still runs system perl for the literal alias `system`. That alias can come from `.perl-version`, `pvm.toml`, env vars or user config.
- **Repro:**
  ```
  cd v/sec1; echo system > .perl-version; echo '1;' > a.pl
  strace -f -e trace=execve psc check a.pl
  ```
  The trace shows `execve("/usr/bin/perl", ["/usr/bin/perl", "-v"], ...) = 0`.
- **Impact:** this breaks the maintainer's security ruling that psc never executes perl. The LSP calls incRoots on every open and every change.
- **Fix:** add a static/no-exec resolution mode for psc, in which ResolveVersionAlias and getPathForVersion refuse `system` and the dynamic aliases. Alternatively, read `.perl-version` and `pvm.toml` directly and map them through the registry. Add a test with stubs for DetectSystemPerl and MetaCPAN that fail if called.
- **Confidence:** High. **NEW:** incRoots is added by this PR; origin/pu's psc never called ResolveVersion. **Issue:** none. **Found by:** security (S1).

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] psc check and the LSP run `/usr/bin/perl -v` when a version file says `system` is fixed (`go test ./internal/psc/ -run '^TestPscNeverExecutesPerl$' -count=1 -v | grep -q '^--- PASS: TestPscNeverExecutesPerl'`)
