---
title: "Port the statement and flow rules onto the lowered tree: conditionals, guards, loops, returns, use"
state: done
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a12a85-304b-77f1-9bda-e6b73c317305
blocks:
- 01a12a85-31a5-769d-88c4-e2add3f8fefc
created: 2026-10-11T10:32:14.336031794Z
updated: 2026-10-13T01:56:39.308388094Z
sessions:
- start_sha: 0908d613150f31c001ca72445c97207e13a34b47
  end_sha: de501a52304e0902ef40141ecc5d555442b0866b
  commits: 1
  started_at: 2026-10-13T00:37:12.207504849Z
  ended_at: 2026-10-13T01:56:39.308388094Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-13T00:37:12.207504849Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-13T01:56:39.308388094Z
observed_paths:
- internal/infer/golden_test.go
- internal/infer/infer.go
- internal/infer/lowered.go
- internal/infer/lowered_flow.go
- internal/infer/lowered_test.go
- internal/infer/testdata/lowered/flow/and.golden
- internal/infer/testdata/lowered/flow/and.pl
- internal/infer/testdata/lowered/flow/and_early_return.golden
- internal/infer/testdata/lowered/flow/and_early_return.pl
- internal/infer/testdata/lowered/flow/and_keyword.golden
- internal/infer/testdata/lowered/flow/and_keyword.pl
- internal/infer/testdata/lowered/flow/and_partial.golden
- internal/infer/testdata/lowered/flow/and_partial.pl
- internal/infer/testdata/lowered/flow/and_two_vars.golden
- internal/infer/testdata/lowered/flow/and_two_vars.pl
- internal/infer/testdata/lowered/flow/blessed.golden
- internal/infer/testdata/lowered/flow/blessed.pl
- internal/infer/testdata/lowered/flow/blessed_and_defined.golden
- internal/infer/testdata/lowered/flow/blessed_and_defined.pl
- internal/infer/testdata/lowered/flow/blessed_bare.golden
- internal/infer/testdata/lowered/flow/blessed_bare.pl
- internal/infer/testdata/lowered/flow/blessed_negated.golden
- internal/infer/testdata/lowered/flow/blessed_negated.pl
- internal/infer/testdata/lowered/flow/block_decl_in_branch.golden
- internal/infer/testdata/lowered/flow/block_decl_in_branch.pl
- internal/infer/testdata/lowered/flow/cfor.golden
- internal/infer/testdata/lowered/flow/cfor.pl
- internal/infer/testdata/lowered/flow/early_die.golden
- internal/infer/testdata/lowered/flow/early_die.pl
- internal/infer/testdata/lowered/flow/early_die_msg.golden
- internal/infer/testdata/lowered/flow/early_die_msg.pl
- internal/infer/testdata/lowered/flow/early_die_paren.golden
- internal/infer/testdata/lowered/flow/early_die_paren.pl
- internal/infer/testdata/lowered/flow/early_exit.golden
- internal/infer/testdata/lowered/flow/early_exit.pl
- internal/infer/testdata/lowered/flow/early_exit_code.golden
- internal/infer/testdata/lowered/flow/early_exit_code.pl
- internal/infer/testdata/lowered/flow/early_exit_with_else.golden
- internal/infer/testdata/lowered/flow/early_exit_with_else.pl
- internal/infer/testdata/lowered/flow/early_last.golden
- internal/infer/testdata/lowered/flow/early_last.pl
- internal/infer/testdata/lowered/flow/early_return.golden
- internal/infer/testdata/lowered/flow/early_return.pl
- internal/infer/testdata/lowered/flow/early_return_value.golden
- internal/infer/testdata/lowered/flow/early_return_value.pl
- internal/infer/testdata/lowered/flow/element_guard.golden
- internal/infer/testdata/lowered/flow/element_guard.pl
- internal/infer/testdata/lowered/flow/elsif.golden
- internal/infer/testdata/lowered/flow/elsif.pl
- internal/infer/testdata/lowered/flow/elsif_else.golden
- internal/infer/testdata/lowered/flow/elsif_else.pl
- internal/infer/testdata/lowered/flow/elsif_exit.golden
- internal/infer/testdata/lowered/flow/elsif_exit.pl
- internal/infer/testdata/lowered/flow/elsif_exit_no_else.golden
- internal/infer/testdata/lowered/flow/elsif_exit_no_else.pl
- internal/infer/testdata/lowered/flow/elsif_negated.golden
- internal/infer/testdata/lowered/flow/elsif_negated.pl
- internal/infer/testdata/lowered/flow/elsif_three.golden
- internal/infer/testdata/lowered/flow/elsif_three.pl
- internal/infer/testdata/lowered/flow/for_alias_redeclare.golden
- internal/infer/testdata/lowered/flow/for_alias_redeclare.pl
- internal/infer/testdata/lowered/flow/for_default.golden
- internal/infer/testdata/lowered/flow/for_default.pl
- internal/infer/testdata/lowered/flow/for_list.golden
- internal/infer/testdata/lowered/flow/for_list.pl
- internal/infer/testdata/lowered/flow/for_my.golden
- internal/infer/testdata/lowered/flow/for_my.pl
- internal/infer/testdata/lowered/flow/for_my_scalar.golden
- internal/infer/testdata/lowered/flow/for_my_scalar.pl
- internal/infer/testdata/lowered/flow/for_nested_guard.golden
- internal/infer/testdata/lowered/flow/for_nested_guard.pl
- internal/infer/testdata/lowered/flow/for_no_my.golden
- internal/infer/testdata/lowered/flow/for_no_my.pl
- internal/infer/testdata/lowered/flow/foreach_scalar.golden
- internal/infer/testdata/lowered/flow/foreach_scalar.pl
- internal/infer/testdata/lowered/flow/if_defined_else.golden
- internal/infer/testdata/lowered/flow/if_defined_else.pl
- internal/infer/testdata/lowered/flow/if_not_defined.golden
- internal/infer/testdata/lowered/flow/if_not_defined.pl
- internal/infer/testdata/lowered/flow/if_not_ref_else.golden
- internal/infer/testdata/lowered/flow/if_not_ref_else.pl
- internal/infer/testdata/lowered/flow/if_ref.golden
- internal/infer/testdata/lowered/flow/if_ref.pl
- internal/infer/testdata/lowered/flow/is_bool.golden
- internal/infer/testdata/lowered/flow/is_bool.pl
- internal/infer/testdata/lowered/flow/isa.golden
- internal/infer/testdata/lowered/flow/isa.pl
- internal/infer/testdata/lowered/flow/isa_else.golden
- internal/infer/testdata/lowered/flow/isa_else.pl
- internal/infer/testdata/lowered/flow/modifiers.golden
- internal/infer/testdata/lowered/flow/modifiers.pl
- internal/infer/testdata/lowered/flow/nested_if.golden
- internal/infer/testdata/lowered/flow/nested_if.pl
- internal/infer/testdata/lowered/flow/no_early_exit.golden
- internal/infer/testdata/lowered/flow/no_early_exit.pl
- internal/infer/testdata/lowered/flow/no_guard.golden
- internal/infer/testdata/lowered/flow/no_guard.pl
- internal/infer/testdata/lowered/flow/not_and.golden
- internal/infer/testdata/lowered/flow/not_and.pl
- internal/infer/testdata/lowered/flow/or.golden
- internal/infer/testdata/lowered/flow/or.pl
- internal/infer/testdata/lowered/flow/or_else.golden
- internal/infer/testdata/lowered/flow/or_else.pl
- internal/infer/testdata/lowered/flow/positive_early_return.golden
- internal/infer/testdata/lowered/flow/positive_early_return.pl
- internal/infer/testdata/lowered/flow/reftype.golden
- internal/infer/testdata/lowered/flow/reftype.pl
- internal/infer/testdata/lowered/flow/sub_guard_body.golden
- internal/infer/testdata/lowered/flow/sub_guard_body.pl
- internal/infer/testdata/lowered/flow/sub_returns.golden
- internal/infer/testdata/lowered/flow/sub_returns.pl
- internal/infer/testdata/lowered/flow/undeclared.golden
- internal/infer/testdata/lowered/flow/undeclared.pl
- internal/infer/testdata/lowered/flow/unless_defined.golden
- internal/infer/testdata/lowered/flow/unless_defined.pl
- internal/infer/testdata/lowered/flow/unless_early_return.golden
- internal/infer/testdata/lowered/flow/unless_early_return.pl
- internal/infer/testdata/lowered/flow/unless_elsif.golden
- internal/infer/testdata/lowered/flow/unless_elsif.pl
- internal/infer/testdata/lowered/flow/unless_ref_else.golden
- internal/infer/testdata/lowered/flow/unless_ref_else.pl
- internal/infer/testdata/lowered/flow/until_defined.golden
- internal/infer/testdata/lowered/flow/until_defined.pl
- internal/infer/testdata/lowered/flow/use.golden
- internal/infer/testdata/lowered/flow/use.pl
- internal/infer/testdata/lowered/flow/user_arg0.golden
- internal/infer/testdata/lowered/flow/user_arg0.pl
- internal/infer/testdata/lowered/flow/user_complex_arg.golden
- internal/infer/testdata/lowered/flow/user_complex_arg.pl
- internal/infer/testdata/lowered/flow/user_compound.golden
- internal/infer/testdata/lowered/flow/user_compound.pl
- internal/infer/testdata/lowered/flow/user_isa.golden
- internal/infer/testdata/lowered/flow/user_isa.pl
- internal/infer/testdata/lowered/flow/user_later.golden
- internal/infer/testdata/lowered/flow/user_later.pl
- internal/infer/testdata/lowered/flow/user_negated.golden
- internal/infer/testdata/lowered/flow/user_negated.pl
- internal/infer/testdata/lowered/flow/user_nonguard.golden
- internal/infer/testdata/lowered/flow/user_nonguard.pl
- internal/infer/testdata/lowered/flow/user_shift.golden
- internal/infer/testdata/lowered/flow/user_shift.pl
- internal/infer/testdata/lowered/flow/user_signature.golden
- internal/infer/testdata/lowered/flow/user_signature.pl
- internal/infer/testdata/lowered/flow/user_two_returns.golden
- internal/infer/testdata/lowered/flow/user_two_returns.pl
- internal/infer/testdata/lowered/flow/while_defined.golden
- internal/infer/testdata/lowered/flow/while_defined.pl
- internal/infer/testdata/lowered/matching.txt
---

## Context

- paths: `internal/infer/lowered.go`, `internal/infer/testdata/lowered/flow/` (new)
- commands: `go test ./internal/infer/ -run "^TestLoweredGoldenRatchet$" -count=1 -v`
- entrypoints: `walkConditionalStatement` (`infer.go:2750`), `extractGuardPattern` (`:2283`), `extractNegatedGuard` (`:2469`), `walkLoopStatement` (`:3107`), `collectExplicitReturns` (`:3294`), `collectImplicitReturn` (`:3337`), `blockAlwaysExits` (`:3495`)

The second of seven port issues. It ports the rules that walk statements: guard narrowing in
conditionals (`defined`, `ref`, `isa`, and user-defined guards), the
negated-guard and branch-join logic, loops, explicit and implicit returns,
and `use` statements.

Two things the lowered tree changes here:

- **`unless` is a branch swap, not a `Not`** (§4.14.5, and the statements
  issue's lowering). A guard under `unless` narrows the ELSE branch. The
  legacy rule reaches the same answer by negating; check each guard fixture
  for both spellings.
- **`blockAlwaysExits` detects `return`/`die`/`exit` by text.** `LoopEx.Op`
  makes `return` a field read. `last`/`next` are needed for correct branch
  joins too; adding them changes what PSC reports, so it is NOT part of this
  port. If a fixture shows the gap, record it and file a follow-up.

The method is the one "AnalyzeLowered: the entry point, the golden
ratchet, and the declaration and literal rules" set up: fixtures
under `testdata/lowered/flow/`, goldens captured from the legacy path, and
`TestLoweredGoldenRatchet` holding `matching.txt` in both directions. Every
rule is REWRITTEN against the lowered vocabulary, not shimmed. No dataflow
IR (§6.1.6): if a rule wants flow, it keeps the approximation it has today.

Rollback: this issue changes only `AnalyzeLowered` and test data. Nothing in `internal/psc` calls it yet, so reverting it changes nothing PSC reports.

## Steps

- [ ] Write fixtures under `testdata/lowered/flow/`, drawn from the
      existing cases in `infer_test.go` where they exist, using only
      constructs whose rules are already ported plus this slice's.
- [ ] Capture their goldens from the legacy path:
      `go test ./internal/infer/ -run "^TestCaptureLegacyGolden$" -infer.capture -count=1`.
      Commit the fixtures and goldens.
- [ ] Run the ratchet and confirm the new fixtures do not match yet:
      `go test ./internal/infer/ -run "^TestLoweredGoldenRatchet$" -count=1 -v`
- [ ] Port this slice's rules into `AnalyzeLowered`. Commit as each rule
      group goes green rather than once at the end.
- [ ] Update `matching.txt` with `-infer.ratchet.update` and re-run the
      ratchet to green. Every `flow/` fixture must match. One that cannot
      is a legacy defect or a regression: triage it before moving on. If it
      is a legacy defect, edit its golden in its own commit with a
      `Diagnostics-changed:` trailer naming the cause.
- [ ] Write `TestLoweredFlowOverBrokenInput`: `if (;) { my $y = 1; }` and
      `my $x; if ($x) {` (perl rejects both) analyse without a panic, emit
      no diagnostic inside a `Hole`, and still type `$y` in the intact
      block. Run to verify failure, implement, run to green, commit.
- [ ] Run the infer suite: `go test ./internal/infer/ -count=1`
- [ ] Commit.

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

## Acceptance Criteria

### Positive Scenarios
- [ ] The lowered path holds the golden ratchet in both directions (`go test ./internal/infer/ -run '^TestLoweredGoldenRatchet$' -count=1 -v | grep -q '^--- PASS: TestLoweredGoldenRatchet'`)
- [ ] Every flow fixture matches its legacy golden on the lowered path (`test -n "$(ls internal/infer/testdata/lowered/flow/*.pl)" && ! ls internal/infer/testdata/lowered/flow/*.pl | sed 's|.*/lowered/||' | grep -vxFf internal/infer/testdata/lowered/matching.txt`)

### Negative Scenarios
- [ ] A conditional perl rejects analyses without a panic, says nothing about the Hole, and still types the intact block (`go test ./internal/infer/ -run '^TestLoweredFlowOverBrokenInput$' -count=1 -v | grep -q '^--- PASS: TestLoweredFlowOverBrokenInput'`)
