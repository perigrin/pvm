---
title: "[C1] psc check and the LSP run /usr/bin/perl -v when a version file says system"
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:06.94921152Z
updated: 2026-10-16T01:16:06.94921152Z
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
