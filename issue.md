---
title: "[I19] succeeded strips Undef from every builtin's result, including getc and each"
state: pending
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.844920751Z
updated: 2026-10-16T01:16:07.844920751Z
---

## Context

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I19**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/infer/calls.go:246-262`.
- **Repro:** `logicB/s2.pl` (stdin from /dev/null): `getc(STDIN) . "x"` and `each(%h) . "y"`. perl warns `uninitialized value` on both. New psc is silent; origin/pu reported getc.
- **Also:** ponytail-B noted that `user := c.Ampersand || len(CoreBuiltin)==0` treats a builtin overridden by `use subs` as a builtin, so it strips Undef from a user sub's result. That follows from reading the code; I did not reproduce it.
- **Fix:** allow-list the failure-only builtins (fileno, readlink, read, ...), and treat an overridden name as a user sub.
- **Confidence:** High for getc/each, Medium for `use subs`. **NEW.** **Issue:** none. **Found by:** logic-B #12, ponytail-B ("left alone" note).

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `succeeded` strips Undef from every builtin's result, including `getc` and `each` is fixed (`go test ./internal/infer/ -run '^TestSucceededKeepsNormalUndef$' -count=1 -v | grep -q '^--- PASS: TestSucceededKeepsNormalUndef'`)
