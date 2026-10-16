---
title: "[R1] The review fixes add 75 lines to psc check over perl's lib"
state: done
urgency: normal
milestone: pr486-review
created: 2026-10-16T12:09:13.711338306Z
updated: 2026-10-16T15:57:55.812232067Z
sessions:
- start_sha: 77107dfbeb77e7b3f1d94437583eae6d28eb41c3
  end_sha: 3cfb717a6e8ccea9bc35fe40250bf479e8ccf148
  commits: 1
  started_at: 2026-10-16T12:09:17.185041672Z
  ended_at: 2026-10-16T15:57:55.812232067Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T12:09:17.185041672Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-16T15:57:55.812232067Z
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

- found by: re-measuring psc check over perl 5.42's lib after the PR #486 review fixes merged (77107dfb), 2026-10-16
- evidence: `scratchpad/sweep/{old,new,added,removed}.diag` -- old = 8a0c85c6 (before the review fixes), new = 77107dfb
- commands: `PVM_PERL_VERSION=5.42.0 psc check <perl 5.42 privlib>`; perl 5.42 `-w`

perigrin's bar for this PR is that psc check output only improves. Before the review fixes, perl 5.42's lib drew 8 lines; after them, 79: 75 added, 4 removed. The added lines cluster: `ref` argument 1 given a near-total union (7), `waitpid`/`read`/`syswrite` given Undef|Int, `eq`/`==`/`+`/`>` operands, `chomp` given List, `open`/`fileno`/`close` given Str, `chdir`, `lc`, `substr`. One removed line, JSON/PP.pm:918, was a true positive (perl warns on truncated input).

## Steps

- [ ] Bucket every added line and the removed JSON/PP.pm:918 by root cause, with a minimal repro and perl -w evidence each; attribute each to the review merge that introduced it
- [ ] Fix each false-positive bucket test-first, keeping a control; restore the JSON/PP true positive or show perl does not warn there
- [ ] Re-sweep: no added line remains that is not a true positive perl warns on

## Acceptance criteria

- [ ] psc check over perl 5.42's lib adds nothing over the pre-review baseline but true positives (`go test ./internal/psc/ -run '^TestReviewFixesAddNoFalsePositives$' -count=1 -v | grep -q '^--- PASS: TestReviewFixesAddNoFalsePositives'`)
