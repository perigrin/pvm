---
title: "[R2] psc check ignores no warnings scopes, reporting what perl has been told not to warn about"
state: done
urgency: normal
milestone: pr486-review
created: 2026-10-16T15:49:27.346112746Z
updated: 2026-10-16T15:57:57.286134303Z
sessions:
- start_sha: 77107dfbeb77e7b3f1d94437583eae6d28eb41c3
  end_sha: 3cfb717a6e8ccea9bc35fe40250bf479e8ccf148
  commits: 1
  started_at: 2026-10-16T15:49:28.307037927Z
  ended_at: 2026-10-16T15:57:57.286134303Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T15:49:28.307037927Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T15:57:57.286134303Z
observed_paths:
- internal/infer/calls.go
- internal/infer/flow.go
- internal/infer/infer.go
- internal/infer/nowarnings_test.go
- internal/infer/ops.go
- internal/infer/special.go
- internal/infer/symbols.go
- internal/infer/testdata/lowered/calls/bare_decl_call.golden
- internal/infer/testdata/lowered/flow/modifiers.golden
- internal/lower/literals.go
- internal/lower/literals_test.go
- internal/lower/node.go
- internal/psc/review_regressions_test.go
- internal/psc/testdata/review_regressions/aggregate.pl
- internal/psc/testdata/review_regressions/aggregate_control.pl
- internal/psc/testdata/review_regressions/assign_guard.pl
- internal/psc/testdata/review_regressions/assign_guard_control.pl
- internal/psc/testdata/review_regressions/closure_scope.pl
- internal/psc/testdata/review_regressions/closure_scope_control.pl
- internal/psc/testdata/review_regressions/first_element.pl
- internal/psc/testdata/review_regressions/first_element_control.pl
- internal/psc/testdata/review_regressions/glob_value.pl
- internal/psc/testdata/review_regressions/glob_value_control.pl
- internal/psc/testdata/review_regressions/handle_name.pl
- internal/psc/testdata/review_regressions/handle_name_control.pl
- internal/psc/testdata/review_regressions/list_assign.pl
- internal/psc/testdata/review_regressions/list_assign_control.pl
- internal/psc/testdata/review_regressions/loop_exit.pl
- internal/psc/testdata/review_regressions/loop_exit_control.pl
- internal/psc/testdata/review_regressions/modifier.pl
- internal/psc/testdata/review_regressions/modifier_control.pl
- internal/psc/testdata/review_regressions/ref_eq.pl
- internal/psc/testdata/review_regressions/ref_eq_control.pl
- internal/psc/testdata/review_regressions/ternary_guard.pl
- internal/psc/testdata/review_regressions/ternary_guard_control.pl
- internal/psc/testdata/review_regressions/truthy.pl
- internal/psc/testdata/review_regressions/truthy_control.pl
- internal/psc/testdata/review_regressions/vstring.pl
- internal/psc/testdata/review_regressions/vstring_control.pl
- internal/types/narrowing.go
- internal/types/types.go
---

## Context

- found by: R1's re-sweep of perl 5.42's lib (2026-10-16): Test/Builder.pm:683 is under `no warnings qw/uninitialized numeric/`, and psc reports Undef in a Str slot there
- paths: `internal/parse/` (pragmas: `use warnings` / `no warnings` with category lists), `internal/lower/` (UseStmt), `internal/infer/` (operand and argument checks)
- commands: `go test ./internal/infer/`; perl 5.42 `-w`

perigrin ruled 2026-10-16: honour `no warnings` now, on PR #486. A lexical `no warnings` (all categories, or a list) turns off, for the rest of its enclosing block, the warnings perl would emit there; psc must skip the diagnostics that correspond to those categories:
- `uninitialized`: an Undef-including operand or argument (coercion/type mismatch whose only failing part is Undef)
- `numeric`: a Str (non-numeric) in a numeric slot
- `no warnings;` with no list: all of them
`use warnings` (or a nested block) restores them. perl 5.42: `my $u; { no warnings "uninitialized"; my $s = "x" . $u; } my $t = "y" . $u;` warns only on the second concatenation. A diagnostic perl reports as an error regardless of warnings (e.g. a ref used as a filehandle) is not affected.

## Steps

- [ ] Measure the category for each psc diagnostic kind against perl; failing test first; scope by lexical block; `make test`; commit

## Acceptance criteria

- [ ] a no warnings scope silences the diagnostics of its categories (`go test ./internal/infer/ -run '^TestNoWarningsScopeSilencesItsCategories$' -count=1 -v | grep -q '^--- PASS: TestNoWarningsScopeSilencesItsCategories'`)
