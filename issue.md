---
title: "[I19] succeeded strips Undef from every builtin's result, including getc and each"
state: done
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.844920751Z
updated: 2026-10-16T03:49:29.496694839Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: 133837eb651f8dcffb4203d072b3701753486869
  commits: 4
  started_at: 2026-10-16T01:17:18.530947391Z
  ended_at: 2026-10-16T03:49:29.496694839Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:18.530947391Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T03:49:29.496694839Z
observed_paths:
- docs/plans/2026-09-05-parser-conformance-plan.md
- docs/plans/2026-09-21-graded-conformance-corpus.md
- docs/specs/perl-parser/04-expressions.md
- docs/specs/perl-parser/05-statements.md
- docs/specs/perl-parser/06-incremental-lsp.md
- internal/infer/calls.go
- internal/infer/calls_test.go
- internal/infer/flow.go
- internal/infer/flow_test.go
- internal/infer/infer.go
- internal/infer/lowered_test.go
- internal/infer/ops.go
- internal/infer/ops_test.go
- internal/infer/testdata/lowered/ops/compound_assign.golden
- internal/infer/testdata/lowered/ops/defined_or.golden
- internal/infer/testdata/lowered/special/ref_arith.golden
- internal/infer/testdata/lowered/special/ref_paren.golden
- internal/infer/testdata/lowered/special/wantarray_top.golden
- internal/infer/testdata/lowered/vars/arraylen.golden
- internal/infer/testdata/lowered/vars/elem_from_var.golden
- internal/infer/testdata/lowered/vars/our_element.golden
- internal/infer/vars.go
- internal/infer/vars_test.go
- internal/lower/call.go
- internal/lower/node.go
- internal/lower/positions.go
- internal/parse/arity.go
- internal/parse/declarations/Carp.pmt
- internal/parse/loader_regular_test.go
- internal/parse/module.go
- internal/parse/parse.go
- internal/parse/staysinroots_test.go
- internal/parse/term.go
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
