---
title: "[I8] $x // d and $x || d keep the left operand's Undef"
state: done
urgency: normal
milestone: pr486-review
created: 2026-10-16T01:16:07.419094347Z
updated: 2026-10-16T03:49:11.48036401Z
sessions:
- start_sha: 8a0c85c60a2bae5d04dbbb2b27eb4310312d362f
  end_sha: 133837eb651f8dcffb4203d072b3701753486869
  commits: 4
  started_at: 2026-10-16T01:17:18.179086063Z
  ended_at: 2026-10-16T03:49:11.48036401Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T01:17:18.179086063Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T03:49:11.48036401Z
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

- found by: the agentic review of PR #486 (feat/pgt-m2 at 8a0c85c6), finding **I8**; report `paad/code-reviews/feat-pgt-m2-2026-10-16-01-10-33-8a0c85c6.md`, specialist evidence and repros in the session scratchpad `review/`
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

perigrin ruled 2026-10-16 that every verified finding is fixed on PR #486 before it merges. If the finding names an issue that already covers part of it, close that issue with this one.

### The verified finding

- **Where:** `internal/infer/ops.go:57-58`: `Join(valueOf(l), valueOf(r))` for every logical op.
- **Repro:**
  - `logicB/o1.pl`: `my $v = $x // 0; $v + 1` and `my $w = $x || 0; $w + 1` give new psc errors `got Undef|Int` on lines 4 and 6. origin/pu is silent on both, and perl prints 4 and 4.
  - `v/o2.pl`: `my $v = $o || 0; length($v)` gives `error: call to length: argument 1 expects Str, got Undef|Int` plus a "hint: Add guard" on the most common default idiom in Perl.
- **Fix:** for `//`, `||` and `or`, use `Join(valueOf(l) &^ Undef, valueOf(r))`.
- **Confidence:** High. **NEW** in the variable case. **Issue:** none. **Found by:** logic-B #7.

## Steps

- [ ] Failing test first from the repro; fix so perl 5.42's verdict holds, keeping a control perl warns on or refuses; golden edits in their own commit under a `Diagnostics-changed:` trailer; precision must stay >= 64/76 with WRONG 0 and the m1 gates must hold; `make test`; commit

## Acceptance criteria

- [ ] `$x // d` and `$x || d` keep the left operand's Undef is fixed (`go test ./internal/infer/ -run '^TestDefinedOrDropsLeftUndef$' -count=1 -v | grep -q '^--- PASS: TestDefinedOrDropsLeftUndef'`)
