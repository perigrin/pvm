---
title: "Handle the call-site traps: sort invocant and comma, delete slice shape, condition context, void match"
state: done
urgency: normal
milestone: core-module-types
blocked_by:
- 01a0fd9b-5700-7514-a143-2caed2d06d8b
blocks:
- 01a1464e-de34-7403-95d2-25f4855404b2
created: 2026-10-16T20:59:46.565162192Z
updated: 2026-10-17T03:33:13.221707064Z
sessions:
- start_sha: a35012438663295404da21ff5a6a7ec965d1d93a
  end_sha: 3d5883ad3461132cbafb527d2cd0af0b49134811
  commits: 6
  started_at: 2026-10-17T00:07:43.001775262Z
  ended_at: 2026-10-17T01:04:22.914493217Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-17T00:07:43.001775262Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-17T01:04:22.914493217Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
- internal/infer/calls.go
- internal/infer/declared_test.go
- internal/infer/export_test.go
- internal/infer/flow.go
- internal/infer/flow_test.go
- internal/infer/infer.go
- internal/infer/ops.go
- internal/infer/returntypes_test.go
- internal/infer/special.go
- internal/infer/testdata/lowered/special/sort_block.golden
- internal/infer/testdata/lowered/special/sort_list.golden
- internal/infer/testdata/lowered/special/sort_plain.golden
- internal/infer/testdata/return_oracle.pl
- internal/parse/decl.go
- internal/parse/declarations.go
- internal/parse/exportattr_test.go
- internal/parse/heldtoperl_test.go
- internal/parse/qualified_test.go
- internal/parse/testdata/held_to_perl.pl
- internal/parse/use.go
- internal/perltest/perltest.go
- internal/types/declared.go
- internal/types/select.go
- test/e2e/psc_check_test.go
- test/e2e/testdata/check/undecided_declared.pl
---

## Prerequisites

- [ ] Type calls to imported library subs from their .tpm declarations and select multis at call sites (01a0fd9b-5700)

## Context

- paths: `internal/infer/infer.go`, `internal/infer/calls.go`, `internal/infer/declared_test.go`, `internal/types/select.go`, `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/infer/ -run '^TestInferSort|^TestInferDeleteSliceNotElement$|^TestInferConditionMatchNotList$|^TestInferVoidMatchIsUntypedNotError$' -count=1 -v`
- entrypoints: call-site selection as 01a0fd9b-5700 leaves it (infer reads `Call.Signatures` and selects among them), `types.Select`/`types.SelectShaped`, `assignedContext`

Split out of 01a0fd9b-5700 (2026-10-16), which builds call-site selection; this issue fixes the four places where a call site must hand selection something other than the plain argument list -- sort's invocant and comma, delete's slice shape, scalar context for conditions, and a void `=~` -- so it depends on 5700, not the reverse. Each was found by an alignment gate; their findings, moved here verbatim:

From the alignment gate on 01a0fd9b-5468-7355-a22e-1b16fa5ee354: context-only multis (`..`, `...`, `=~`, `\`) have no candidate answering VOID, so `SelectShaped` returns Failed for `$s =~ s/a/b/g;` in void context -- ordinary perl, which the old plain `=~` declaration typed. For `..` and `\` perl warns "Useless use ... in void context", but `=~` in void is idiomatic. A call site must treat a void-context Failed as untyped, never as an error: add a void `=~` case beside the scalar-sort negative scenario, failing first.

From 01a11146-c4da-79e4-8232-6ad2e731f3e7: `delete` is now two candidates, `(List @slice = die) List` and `(Scalar $element) Scalar`, chosen by operand SHAPE through `SelectShaped`. A call site must pass a slice operand (`@h{...}`, `@a[...]`, `%h{...}`, `%a[...]`) as `@`-shaped and an element (`$h{a}`, `$a[1]`) as `$`-shaped, or every delete selects the element candidate. Add a test that `delete @h{qw(a b)}` is typed List.

From 01a0fd9b-519d-7b42-886e-2222a3076986: sort's invocant candidate `(Code|Str $by: List @list = die)` is never selected, because `types.Select` only handles calls that write no invocant (it ignores candidates whose invocant is required). A call site that writes one -- `sort byname @x`, `sort $by @x`, `sort { ... } @x` -- must pass the invocant to selection. Add a test that `sort byname @x` selects that candidate.

From the alignment gate on 01a0fd9b-5670-78d0-84be-9b9618e9b081: infer takes calling context only from assignments (`assignedContext`), so `=~` in a condition (`if ($s =~ /b/)`, a ternary's condition, `!`'s operand) is typed List where it was Boolean -- sound, but wider hints. RFC 0001 ":context(...)" says boolean context is scalar: give conditions scalar context when call sites are built, with a test that `if ($s =~ /b/)`'s match is Boolean.

Where perl decides a question, run perl (`/home/perigrin/.local/bin/perl`) rather than escalating it. When all four hold, "Call sites" is complete: re-mark it.

## Steps

- [ ] Write failing tests `TestInferSortCommaSelectsPlain` and `TestInferSortInvocantNotPlain`; run to verify failure; pass the invocant to selection; run to verify pass; commit
- [ ] Write failing test `TestInferDeleteSliceNotElement`; run to verify failure; pass operand shape; run to verify pass; commit
- [ ] Write failing test `TestInferConditionMatchNotList`; run to verify failure; give conditions scalar context; run to verify pass; commit
- [ ] Write failing test `TestInferVoidMatchIsUntypedNotError`; run to verify failure; treat a void-context Failed as untyped; run to verify pass; commit
- [ ] Re-mark the RFC heading "Call sites" `(*Implemented*, <shas>)`; `make test`; commit

### Review Findings

Tier-2 gate, 2026-10-17, on 3d5883ad (fresh reviewer subagent; lenses applied inline; ten overlay mutations; psc check over perl 5.42's privlib identical to the parent; perl 5.42 measurements). Sort's List type is consistent with other CORE.tpm-typed builtins; SortSelection/newAnalysis justified; void arity failure correctly stays an error. VERDICT: REOPEN.

1. (in scope, 95) `internal/infer/flow.go:320-325` elsif walks n.Cond without `a.condition(n.Cond)`: `if (0) {...} elsif ($s =~ /b/) {...}` types the match List, contradicting 73a6178e and the condition doc comment. Fix: call condition at the top of elsif; add an elsif case to TestInferConditionMatchNotList.
2. (in scope, 95) TestInferVoidMatchIsUntypedNotError's last-statement assertion can never fail: `sub f { $_[0] =~ s/b/c/g }` has no semicolon, so findNodeType returns the ExprStmt (Any). Fix: `sub f { $_[0] =~ s/b/c/g; }` or assert types.Str; show it fails when the last statement is also put in void.
3. (in scope, 80) `internal/infer/infer.go:245` voidStatements in block() is untested: removing it passes every test; `sub f { $s =~ s/a/b/g; return 1 }` then gets Str. Fix: a block-level case expecting Unknown.
4. (in scope, 75) `internal/infer/infer.go:211` C-style `for(;COND;)` gives the condition no scalar context; perl 5.42 reports scalar for wantarray there; `for (; $s =~ /b/; )` types the match List. Fix: a CForStmt case calling condition then walking children; a test.
5. (folded in by the orchestrator, 65) `internal/infer/special.go:20` "a sort in scalar context is left untyped" is untested: removing the ListCtx guard passes every test. Fix: `my $c = sort @n` is Unknown.
6. (in scope, Step 5, 70) The RFC "Call sites" paragraph overclaims: after 1 and 4 the condition sentence holds; "a statement whose value is discarded is in void context" is broader than the code (only a block's or file's non-final statements; not a loop body's last, not modifier forms). Fix: say "a statement that is not its block's last".

## Acceptance Criteria

### Positive Scenarios
- [ ] a comma after the sort routine makes it data (`go test ./internal/infer/ -run '^TestInferSortCommaSelectsPlain$' -count=1 -v | grep -q '^--- PASS: TestInferSortCommaSelectsPlain'`)
- [ ] the RFC section on call sites is implemented (`grep -q '^### Call sites (\*Implemented\*' docs/rfcs/0001-type-libraries.md`)

### Negative Scenarios
- [ ] a void-context `$s =~ s/a/b/g;` is untyped and never reported as an error, though no `=~` candidate answers VOID (`go test ./internal/infer/ -run '^TestInferVoidMatchIsUntypedNotError$' -count=1 -v | grep -q '^--- PASS: TestInferVoidMatchIsUntypedNotError'`)
- [ ] a slice delete `delete @h{qw(a b)}` is not typed as the element candidate (Scalar) (`go test ./internal/infer/ -run '^TestInferDeleteSliceNotElement$' -count=1 -v | grep -q '^--- PASS: TestInferDeleteSliceNotElement'`)
- [ ] `sort byname @x` does not select the plain candidate (byname is the invocant, not data) (`go test ./internal/infer/ -run '^TestInferSortInvocantNotPlain$' -count=1 -v | grep -q '^--- PASS: TestInferSortInvocantNotPlain'`)
- [ ] a match in a condition, `if ($s =~ /b/)`, is not typed List (`go test ./internal/infer/ -run '^TestInferConditionMatchNotList$' -count=1 -v | grep -q '^--- PASS: TestInferConditionMatchNotList'`)
