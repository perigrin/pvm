---
title: Type calls to imported library subs from their .tpm declarations and select multis at call sites
state: pending
urgency: normal
milestone: core-module-types
blocked_by:
- 01a0fd9b-5670-78d0-84be-9b9618e9b081
- 01a0fd9b-5698-70cf-b6e4-8899a83b0fe0
- 01a0c115-df67-7d69-8883-73d40b0ae15d
- 01a10177-547c-73a9-a302-f25f579a4cb3
- 01a1464e-d912-7419-8b20-a8e24029b3ec
created: 2026-10-02T17:13:31.136589197Z
updated: 2026-10-16T20:02:59.377765804Z
---

## Prerequisites

- [ ] Have internal/infer read builtin and operator types from CORE.tpm and remove internal/types' builtin table
- [ ] Let a library .tpm use every construct CORE.tpm uses
- [ ] Refuse at parse time the call arities perl refuses at compile time
- [ ] Retire the tree-sitter vocabulary from internal/infer (m2-lower-the-tree, 01a0c115-df67-7d69-8883-73d40b0ae15d)
- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one (core-module-types)
- [ ] Select a single fitting-arity candidate despite argument types, and join an unknown context (core-module-types)

## Context

- paths: `internal/infer/infer.go`, `docs/rfcs/0001-type-libraries.md`, `internal/infer/declared_test.go` (new), `test/e2e/psc_check_test.go`, `test/e2e/testdata/`, `internal/data/help/types.md`
- commands: `go test ./internal/infer/ -count=1`, `go test ./test/e2e/ -run '^TestPSCCheck' -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `internal/infer/infer.go` call-site typing (after m2 it walks `internal/parse` nodes), `internal/types/select.go`, `parse.Imports` (`internal/parse/exports.go:49`)

The milestone body's "internal/infer reads the declarations": once m2 has moved infer onto `internal/parse`, a call to a sub imported from a module with a `.tpm` is typed from that declaration, and a call to a multi -- builtin or library -- selects its candidate at the call site by arity, argument types, context or operand shape through `internal/types/select.go`, joining when undecided. This changes what `psc check` reports wherever a declaration is more precise, which is why it waits for m2's behaviour-preserving cutover. This is RFC 0001 "Call sites" (perigrin, 2026-10-02): `infer` applies the selection rules of "Multi declarations" and "`:context(...)`" at a real call -- the arguments' inferred types, the call's arity and its context -- and takes the selected candidate's return type, joining only when the call site cannot decide. When the declarations rule a call out, it fails: where perl would compile the call (a candidate set with no prototype, or a library dispatching at run time) the parse stays as perl reads it and `psc check` reports an error; where perl refuses at compile time the parser has already refused (the parse-time refusal issue). This completes "Call sites", so it re-marks it. Use a shipped declaration (e.g. `Moose.tpm`) or extend one with measured types for the e2e case.

Open question from the alignment gate on 01a0fd9b-4a5b-7e94-ad87-756dfcfe1f4c (2026-10-06): `types.Select` returns Failed when a call's arity fits but its argument types do not, so a single-candidate `sqrt(Str)` comes back Failed with type Unknown. perigrin's ruling that a parameter type is its coercion target (RFC "What a parameter type means") implies a single fitting-arity candidate is still selected and `psc check` reports `coercion-mismatch`; argument types should decide only BETWEEN candidates. Settle this before call sites consume `Select`.

From the alignment gate on 01a0fd9b-4ca3-7d55-82ac-ba125816ee71: `types.Select(cands, args, ctx)` returns Failed when `ctx` is `UnknownCtx`, because no candidate answers for an unknown context. RFC 0001 "Multi declarations" says the CONSUMER joins when the call site cannot decide, and Failed means the call is ruled out -- which `psc check` would report as an error. Add a step: a call whose context cannot be determined joins the candidates (or Select reports Undecided), with a failing test first.

From the alignment gate on 01a0fd9b-5468-7355-a22e-1b16fa5ee354: context-only multis (`..`, `...`, `=~`, `\`) have no candidate answering VOID, so `SelectShaped` returns Failed for `$s =~ s/a/b/g;` in void context -- ordinary perl, which the old plain `=~` declaration typed. For `..` and `\` perl warns "Useless use ... in void context", but `=~` in void is idiomatic. A call site must treat a void-context Failed as untyped, never as an error: add a void `=~` case beside the scalar-sort negative scenario, failing first.

From 01a11146-c4da-79e4-8232-6ad2e731f3e7: `delete` is now two candidates, `(List @slice = die) List` and `(Scalar $element) Scalar`, chosen by operand SHAPE through `SelectShaped`. A call site must pass a slice operand (`@h{...}`, `@a[...]`, `%h{...}`, `%a[...]`) as `@`-shaped and an element (`$h{a}`, `$a[1]`) as `$`-shaped, or every delete selects the element candidate. Add a test that `delete @h{qw(a b)}` is typed List.

From 01a0fd9b-519d-7b42-886e-2222a3076986: sort's invocant candidate `(Code|Str $by: List @list = die)` is never selected, because `types.Select` only handles calls that write no invocant (it ignores candidates whose invocant is required). A call site that writes one -- `sort byname @x`, `sort $by @x`, `sort { ... } @x` -- must pass the invocant to selection. Add a test that `sort byname @x` selects that candidate.

From the alignment gate on 01a0fd9b-5670-78d0-84be-9b9618e9b081: infer takes calling context only from assignments (`assignedContext`), so `=~` in a condition (`if ($s =~ /b/)`, a ternary's condition, `!`'s operand) is typed List where it was Boolean -- sound, but wider hints. RFC 0001 ":context(...)" says boolean context is scalar: give conditions scalar context when call sites are built, with a test that `if ($s =~ /b/)`'s match is Boolean.

**Ruling (perigrin, 2026-10-16): inference takes call resolution from the parser.** `parse.(*parser).markResolution` (`internal/parse/arity.go:~139`) sets `Node.Override` and `Node.Signatures`, and `internal/lower/call.go:21` copies both onto `lower.Call`. infer reads `Call.Signatures` and selects among them at the call site; it never re-resolves a name itself (so `parse.Imports` above is not the route). Today `internal/infer/calls.go:26` `call()` reads only `parse.CoreBuiltin(name)` and `userCall` never reads `Call.Signatures`: that is this issue's core. The two `types.Select` rulings above (single-candidate coercion, `UnknownCtx` joins) are settled by their own issue in core-module-types before this one; a qualified call (`Carp::confess(...)`) gets its signatures from the parser in a follow-on issue there.

## Steps

- [ ] Write failing test `TestInferTypesCallFromDeclaration` in `internal/infer/declared_test.go`: a call to an imported sub declared with a return type in a `.tpm` is typed with it
- [ ] Run test to verify failure: `go test ./internal/infer/ -run '^TestInferTypesCallFromDeclaration$' -count=1 -v`
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestInferSelectsMultiAtCallSite`: `localtime(0)` in scalar and list context, `each %h` against `each @a`, and an undecided call joining
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing e2e test `TestPSCCheckUsesDeclarations` in `test/e2e/psc_check_test.go`: `psc check` on a fixture reports a coercion mismatch that only the declaration's types reveal
- [ ] Run test to verify failure: `go test ./test/e2e/ -run '^TestPSCCheckUsesDeclarations$' -count=1 -v`
- [ ] Make it pass; `make test`; commit
- [ ] Write failing test `TestInferNoCandidateArityIsError` in `internal/infer/declared_test.go`: a call to a library multi with no prototype whose arity no candidate accepts is reported as an error, and gets no joined type
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Re-mark the RFC heading "Call sites" `(*Implemented*, <sha>)`; commit
- [ ] Doc: in the "Declaration Files" section of `internal/data/help/types.md`, say that `psc check` types a call to a declared sub from its declaration, using the e2e fixture's diagnostic as the example; keep "Typed Perl is valid only in `.tpm` files, and you cannot yet add your own" (a user search path is RFC open question 6); `go test ./internal/cli/ -run '^TestShowTypesHelp$' -count=1`; commit

### Negative Scenarios

- a call to an imported sub whose module has no `.tpm` gets no declared type (unchanged behaviour) (`go test ./internal/infer/ -run '^TestInferUndeclaredImportUntyped$' -count=1 -v | grep -q '^--- PASS: TestInferUndeclaredImportUntyped'`)
- context with no answer has no type at a real call site: `my $n = sort @x` is untyped, not `List` (`go test ./internal/infer/ -run '^TestInferScalarSortHasNoType$' -count=1 -v | grep -q '^--- PASS: TestInferScalarSortHasNoType'`)
- the comma decides sort: `sort $n @x` selects the invocant candidate, `sort "byname", @x` the plain one with the string as data (measured) (`go test ./internal/infer/ -run '^TestInferSortCommaSelectsPlain$' -count=1 -v | grep -q '^--- PASS: TestInferSortCommaSelectsPlain'`)
- `psc check` raises no false positive: on a fixture whose arguments fit the declaration, it reports no coercion mismatch (`go test ./test/e2e/ -run '^TestPSCCheckDeclarationsNoFalsePositive$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckDeclarationsNoFalsePositive'`)

## Acceptance criteria

- [ ] infer types a call from its .tpm declaration (`go test ./internal/infer/ -run '^TestInferTypesCallFromDeclaration$' -count=1 -v | grep -q '^--- PASS: TestInferTypesCallFromDeclaration'`)
- [ ] infer selects a multi candidate at the call site (`go test ./internal/infer/ -run '^TestInferSelectsMultiAtCallSite$' -count=1 -v | grep -q '^--- PASS: TestInferSelectsMultiAtCallSite'`)
- [ ] psc check reports what a declaration reveals (`go test ./test/e2e/ -run '^TestPSCCheckUsesDeclarations$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckUsesDeclarations'`)
- [ ] a call into a module with no .tpm gets no declared type (`go test ./internal/infer/ -run '^TestInferUndeclaredImportUntyped$' -count=1 -v | grep -q '^--- PASS: TestInferUndeclaredImportUntyped'`)
- [ ] a scalar sort call site has no type (`go test ./internal/infer/ -run '^TestInferScalarSortHasNoType$' -count=1 -v | grep -q '^--- PASS: TestInferScalarSortHasNoType'`)
- [ ] a comma after the sort routine makes it data (`go test ./internal/infer/ -run '^TestInferSortCommaSelectsPlain$' -count=1 -v | grep -q '^--- PASS: TestInferSortCommaSelectsPlain'`)
- [ ] psc check reports no mismatch where arguments fit a declaration (`go test ./test/e2e/ -run '^TestPSCCheckDeclarationsNoFalsePositive$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckDeclarationsNoFalsePositive'`)
- [ ] a call whose arity no candidate accepts is an error at the call site (`go test ./internal/infer/ -run '^TestInferNoCandidateArityIsError$' -count=1 -v | grep -q '^--- PASS: TestInferNoCandidateArityIsError'`)
- [ ] the RFC section on call sites is implemented (`grep -q '^### Call sites (\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
