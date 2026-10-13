---
title: "Port the call rules onto the lowered tree: one Call arm for five builtin sites, Handle apart from Args"
state: done
urgency: normal
milestone: m2-lower-the-tree
blocked_by:
- 01a12a85-3233-7ec8-9428-747be4ccef1c
- 01a0c112-a8bb-7d7f-b156-d4ad99474e87
blocks:
- 01a12a85-3386-7e72-8d74-38bf69a9c7d7
created: 2026-10-11T10:32:14.798479017Z
updated: 2026-10-13T21:46:33.442804362Z
sessions:
- start_sha: e61467531761799cbc9c52bf845a9f73a7ba2854
  end_sha: a45799887c9831e70d88981c993e14c2f1de3741
  commits: 1
  started_at: 2026-10-13T16:07:37.064563388Z
  ended_at: 2026-10-13T21:46:33.442804362Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-13T16:07:37.064563388Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-13T21:46:33.442804362Z
observed_paths:
- internal/infer/infer.go
- internal/infer/lowered.go
- internal/infer/lowered_calls.go
- internal/infer/lowered_flow.go
- internal/infer/lowered_ops.go
- internal/infer/lowered_test.go
- internal/infer/lowered_vars.go
- internal/infer/testdata/lowered/calls/abs_int.golden
- internal/infer/testdata/lowered/calls/abs_int.pl
- internal/infer/testdata/lowered/calls/ampersand.golden
- internal/infer/testdata/lowered/calls/ampersand.pl
- internal/infer/testdata/lowered/calls/bare_decl_call.golden
- internal/infer/testdata/lowered/calls/bare_decl_call.pl
- internal/infer/testdata/lowered/calls/chr_num.golden
- internal/infer/testdata/lowered/calls/chr_num.pl
- internal/infer/testdata/lowered/calls/coderef.golden
- internal/infer/testdata/lowered/calls/coderef.pl
- internal/infer/testdata/lowered/calls/delete.golden
- internal/infer/testdata/lowered/calls/delete.pl
- internal/infer/testdata/lowered/calls/die_msg.golden
- internal/infer/testdata/lowered/calls/die_msg.pl
- internal/infer/testdata/lowered/calls/fq_call.golden
- internal/infer/testdata/lowered/calls/fq_call.pl
- internal/infer/testdata/lowered/calls/keys_values.golden
- internal/infer/testdata/lowered/calls/keys_values.pl
- internal/infer/testdata/lowered/calls/length_array.golden
- internal/infer/testdata/lowered/calls/length_array.pl
- internal/infer/testdata/lowered/calls/length_defined.golden
- internal/infer/testdata/lowered/calls/length_defined.pl
- internal/infer/testdata/lowered/calls/list_op_args.golden
- internal/infer/testdata/lowered/calls/list_op_args.pl
- internal/infer/testdata/lowered/calls/lvalue_substr.golden
- internal/infer/testdata/lowered/calls/lvalue_substr.pl
- internal/infer/testdata/lowered/calls/method_instance.golden
- internal/infer/testdata/lowered/calls/method_instance.pl
- internal/infer/testdata/lowered/calls/method_new.golden
- internal/infer/testdata/lowered/calls/method_new.pl
- internal/infer/testdata/lowered/calls/method_unresolved.golden
- internal/infer/testdata/lowered/calls/method_unresolved.pl
- internal/infer/testdata/lowered/calls/nested.golden
- internal/infer/testdata/lowered/calls/nested.pl
- internal/infer/testdata/lowered/calls/param_usage.golden
- internal/infer/testdata/lowered/calls/param_usage.pl
- internal/infer/testdata/lowered/calls/pop_shift.golden
- internal/infer/testdata/lowered/calls/pop_shift.pl
- internal/infer/testdata/lowered/calls/print_handle.golden
- internal/infer/testdata/lowered/calls/print_handle.pl
- internal/infer/testdata/lowered/calls/print_ref_arg.golden
- internal/infer/testdata/lowered/calls/print_ref_arg.pl
- internal/infer/testdata/lowered/calls/push_bare.golden
- internal/infer/testdata/lowered/calls/push_bare.pl
- internal/infer/testdata/lowered/calls/push_paren.golden
- internal/infer/testdata/lowered/calls/push_paren.pl
- internal/infer/testdata/lowered/calls/scalar.golden
- internal/infer/testdata/lowered/calls/scalar.pl
- internal/infer/testdata/lowered/calls/splice_reverse.golden
- internal/infer/testdata/lowered/calls/splice_reverse.pl
- internal/infer/testdata/lowered/calls/sprintf.golden
- internal/infer/testdata/lowered/calls/sprintf.pl
- internal/infer/testdata/lowered/calls/swallowed_compare.golden
- internal/infer/testdata/lowered/calls/swallowed_compare.pl
- internal/infer/testdata/lowered/calls/ternary_arg.golden
- internal/infer/testdata/lowered/calls/ternary_arg.pl
- internal/infer/testdata/lowered/calls/time.golden
- internal/infer/testdata/lowered/calls/time.pl
- internal/infer/testdata/lowered/calls/undef.golden
- internal/infer/testdata/lowered/calls/undef.pl
- internal/infer/testdata/lowered/calls/user_forward.golden
- internal/infer/testdata/lowered/calls/user_forward.pl
- internal/infer/testdata/lowered/calls/user_return.golden
- internal/infer/testdata/lowered/calls/user_return.pl
- internal/infer/testdata/lowered/matching.txt
---

## Context

- paths: `internal/infer/lowered.go`, `internal/infer/testdata/lowered/calls/` (new)
- commands: `go test ./internal/infer/ -run "^TestLoweredGoldenRatchet$" -count=1 -v`
- entrypoints: `inferFunctionCallType` (`infer.go:678`), `inferMethodCallType` (`:780`), `contextualReturnType` (`:939`), `collectCallArgs` (`:1022`), `callReturns` (`:2018`), `collectCallConstraints` (`:4177`), the `indirect_object` workaround (`:1144-1184`)

The fifth of seven port issues. It ports builtin and user-sub call typing,
the arity and argument-type diagnostics, method calls and the class-name
tracking, and context-dependent returns.

What gets simpler, and what review should check:

- the five builtin lookup sites (`parse.CoreBuiltin` at `infer.go:701`,
  `:890`, `:1011`, `:4198`, `:4242`) collapse to one `Call` arm;
- the `indirect_object` workaround (`infer.go:1144-1184`) has no successor:
  `Call.Handle` is a separate field from `Call.Args`;
- `classTypes` (`infer.go:131`, a `map[uint32]string` keyed by start byte)
  is keyed by node id like the annotations.

This slice emits most of PSC's diagnostics (`arity-mismatch`,
`type-mismatch`, `coercion-mismatch`, `unknown-builtin`), so it is where a
golden is most likely to change. Each changed golden is its own commit with
a `Diagnostics-changed:` trailer.

The method is the one "AnalyzeLowered: the entry point, the golden
ratchet, and the declaration and literal rules" set up: fixtures
under `testdata/lowered/calls/`, goldens captured from the legacy path, and
`TestLoweredGoldenRatchet` holding `matching.txt` in both directions. Every
rule is REWRITTEN against the lowered vocabulary, not shimmed. No dataflow
IR (§6.1.6): if a rule wants flow, it keeps the approximation it has today.

Rollback: this issue changes only `AnalyzeLowered` and test data. Nothing in `internal/psc` calls it yet, so reverting it changes nothing PSC reports.

## Steps

- [ ] Write fixtures under `testdata/lowered/calls/`, drawn from the
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
      ratchet to green. Every `calls/` fixture must match. One that cannot
      is a legacy defect or a regression: triage it before moving on. If it
      is a legacy defect, edit its golden in its own commit with a
      `Diagnostics-changed:` trailer naming the cause.
- [ ] Write `TestLoweredCallsOverBrokenInput`: `length(;` and
      `Foo->new(;` (perl rejects both) analyse without a panic and draw no
      arity or argument-type diagnostic, because one argument is a `Hole`.
      Run to verify failure, implement, run to green, commit.
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
- [ ] Every calls fixture matches its legacy golden on the lowered path (`test -n "$(ls internal/infer/testdata/lowered/calls/*.pl)" && ! ls internal/infer/testdata/lowered/calls/*.pl | sed 's|.*/lowered/||' | grep -vxFf internal/infer/testdata/lowered/matching.txt`)

### Negative Scenarios
- [ ] A call with an argument perl rejects analyses without a panic and draws no arity or argument diagnostic (`go test ./internal/infer/ -run '^TestLoweredCallsOverBrokenInput$' -count=1 -v | grep -q '^--- PASS: TestLoweredCallsOverBrokenInput'`)
