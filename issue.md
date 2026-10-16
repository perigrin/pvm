---
title: "Handle the call-site traps: sort invocant and comma, delete slice shape, condition context, void match"
state: pending
urgency: normal
milestone: core-module-types
blocked_by:
- 01a0fd9b-5700-7514-a143-2caed2d06d8b
created: 2026-10-16T20:59:46.565162192Z
updated: 2026-10-16T21:02:06.927123193Z
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

## Acceptance Criteria

### Positive Scenarios
- [ ] a comma after the sort routine makes it data (`go test ./internal/infer/ -run '^TestInferSortCommaSelectsPlain$' -count=1 -v | grep -q '^--- PASS: TestInferSortCommaSelectsPlain'`)
- [ ] the RFC section on call sites is implemented (`grep -q '^### Call sites (\*Implemented\*' docs/rfcs/0001-type-libraries.md`)

### Negative Scenarios
- [ ] a void-context `$s =~ s/a/b/g;` is untyped and never reported as an error, though no `=~` candidate answers VOID (`go test ./internal/infer/ -run '^TestInferVoidMatchIsUntypedNotError$' -count=1 -v | grep -q '^--- PASS: TestInferVoidMatchIsUntypedNotError'`)
- [ ] a slice delete `delete @h{qw(a b)}` is not typed as the element candidate (Scalar) (`go test ./internal/infer/ -run '^TestInferDeleteSliceNotElement$' -count=1 -v | grep -q '^--- PASS: TestInferDeleteSliceNotElement'`)
- [ ] `sort byname @x` does not select the plain candidate (byname is the invocant, not data) (`go test ./internal/infer/ -run '^TestInferSortInvocantNotPlain$' -count=1 -v | grep -q '^--- PASS: TestInferSortInvocantNotPlain'`)
- [ ] a match in a condition, `if ($s =~ /b/)`, is not typed List (`go test ./internal/infer/ -run '^TestInferConditionMatchNotList$' -count=1 -v | grep -q '^--- PASS: TestInferConditionMatchNotList'`)
