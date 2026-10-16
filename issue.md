---
title: "[I27] Go-to-definition on $x[0] or $h{k} looks up $x or $h"
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:08.242462873Z
updated: 2026-10-16T01:16:08.242462873Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I27**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/psc/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/psc/lsp.go:185-191`. The base of an Index or Key lowers to `Var Sigil="$"`, as the lowered trees above show.
- **Impact:** with `my $x = 'a'; my @x = (1); print $x[0];`, definition jumps to `my $x` (errors-B verified this), and `$h{a}` finds nothing.
- **Fix:** when the Var is the base of a non-arrow Index, Key or Slice, use `@` or `%`. vars.go:111-146's containerName already does this.
- **Confidence:** High. **NEW** (the code). **Issue:** none. **Found by:** logic-B #13, errors-B #8.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] Go-to-definition on `$x[0]` or `$h{k}` looks up `$x` or `$h` is fixed (`go test ./internal/psc/ -run '^TestLSPDefinitionOfElement$' -count=1 -v | grep -q '^--- PASS: TestLSPDefinitionOfElement'`)
