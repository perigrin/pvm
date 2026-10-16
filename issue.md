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
- 01a1464e-da0a-7ea5-9232-1cd68885d4e0
- 01a14683-8375-7810-8a75-b772680603c6
blocks:
- 01a1464e-da9d-7f9e-b2cf-b3d8607907a2
- 01a1464e-de34-7403-95d2-25f4855404b2
created: 2026-10-02T17:13:31.136589197Z
updated: 2026-10-16T21:02:05.963670651Z
---

## Prerequisites

- [ ] Have internal/infer read builtin and operator types from CORE.tpm and remove internal/types' builtin table
- [ ] Let a library .tpm use every construct CORE.tpm uses
- [ ] Refuse at parse time the call arities perl refuses at compile time
- [ ] Retire the tree-sitter vocabulary from internal/infer (m2-lower-the-tree, 01a0c115-df67-7d69-8883-73d40b0ae15d)
- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one (core-module-types)
- [ ] Select a single fitting-arity candidate despite argument types, and join an unknown context (core-module-types)
- [ ] Convert every .tpm to :export and :export_ok, then refuse export lists in a .tpm (01a14683-8375-7810-8a75-b772680603c6)

## Context

- paths: `internal/infer/infer.go`, `docs/rfcs/0001-type-libraries.md`, `internal/infer/declared_test.go` (new), `test/e2e/psc_check_test.go`, `test/e2e/testdata/`, `internal/data/help/types.md`
- commands: `go test ./internal/infer/ -count=1`, `go test ./test/e2e/ -run '^TestPSCCheck' -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `internal/infer/infer.go` call-site typing (after m2 it walks `internal/parse` nodes), `internal/types/select.go`, `parse.Imports` (`internal/parse/exports.go:49`)

The milestone body's "internal/infer reads the declarations": once m2 has moved infer onto `internal/parse`, a call to a sub imported from a module with a `.tpm` is typed from that declaration, and a call to a multi -- builtin or library -- selects its candidate at the call site by arity, argument types, context or operand shape through `internal/types/select.go`, joining when undecided. This changes what `psc check` reports wherever a declaration is more precise, which is why it waits for m2's behaviour-preserving cutover. This is RFC 0001 "Call sites" (perigrin, 2026-10-02): `infer` applies the selection rules of "Multi declarations" and "`:context(...)`" at a real call -- the arguments' inferred types, the call's arity and its context -- and takes the selected candidate's return type, joining only when the call site cannot decide. When the declarations rule a call out, it fails: where perl would compile the call (a candidate set with no prototype, or a library dispatching at run time) the parse stays as perl reads it and `psc check` reports an error; where perl refuses at compile time the parser has already refused (the parse-time refusal issue). The call-site traps issue that follows this one completes "Call sites" and re-marks it. Use a shipped declaration (e.g. `Moose.tpm`) or extend one with measured types for the e2e case.

Open question from the alignment gate on 01a0fd9b-4a5b-7e94-ad87-756dfcfe1f4c (2026-10-06): `types.Select` returns Failed when a call's arity fits but its argument types do not, so a single-candidate `sqrt(Str)` comes back Failed with type Unknown. perigrin's ruling that a parameter type is its coercion target (RFC "What a parameter type means") implies a single fitting-arity candidate is still selected and `psc check` reports `coercion-mismatch`; argument types should decide only BETWEEN candidates. Settle this before call sites consume `Select`.

From the alignment gate on 01a0fd9b-4ca3-7d55-82ac-ba125816ee71: `types.Select(cands, args, ctx)` returns Failed when `ctx` is `UnknownCtx`, because no candidate answers for an unknown context. RFC 0001 "Multi declarations" says the CONSUMER joins when the call site cannot decide, and Failed means the call is ruled out -- which `psc check` would report as an error. Add a step: a call whose context cannot be determined joins the candidates (or Select reports Undecided), with a failing test first.

**Ruling (perigrin, 2026-10-16): inference takes call resolution from the parser.** `parse.(*parser).markResolution` (`internal/parse/arity.go:~139`) sets `Node.Override` and `Node.Signatures`, and `internal/lower/call.go:21` copies both onto `lower.Call`. infer reads `Call.Signatures` and selects among them at the call site; it never re-resolves a name itself (so `parse.Imports` above is not the route). Today `internal/infer/calls.go:26` `call()` reads only `parse.CoreBuiltin(name)` and `userCall` never reads `Call.Signatures`: that is this issue's core. The two `types.Select` rulings above (single-candidate coercion, `UnknownCtx` joins) are settled by their own issue in core-module-types before this one; a qualified call (`Carp::confess(...)`) gets its signatures from the parser in a follow-on issue there.

**Split (2026-10-16).** The call-site traps -- sort's invocant and comma, delete's slice shape, scalar context for conditions, and a void `=~` -- are their own issue (01a14683-8405-743e-9c90-7bb5ff9cec88, Handle the call-site traps: sort invocant and comma, delete slice shape, condition context, void match), which depends on this one: each refines selection at a call site, and call-site selection is what this issue builds.

**Choosing between candidates.** Selection decides by arity first, then by argument types between candidates of the same arity (01a1464e-da0a's ruling: a type mismatch never rules a call out). Both are tested at a real call site here, not only in `types`.

**Fixtures.** The unimported-`:export_ok` test uses a fixture declaration in the test declarations dir (`internal/parse/testdata/declarations/My/OkOnly.tpm`, via the `declarations` fs.FS var), never `Scalar/Util.tpm`: that file ships only with a later batch, and a test on a missing declaration passes vacuously. The shadowing test uses a non-builtin name and puts the call AFTER the file's own `sub` definition, so the parse sees the local sub. Measured 2026-10-16 on 5.42.0 with a fixture module exporting `pick`: a local `sub pick` makes both an earlier and a later `pick(1)` call the local sub at run time ("Subroutine pick redefined"). Scalar::Util is the exception the test avoids: `blessed`, `reftype`, `refaddr` and `weaken` are aliases of `builtin::` functions, and a call compiled before the local sub becomes the builtin op -- measured, `use Scalar::Util qw(blessed); print blessed(bless {}, "Z"); sub blessed { "local" }` prints `Z`, while the same call after the definition prints `local`.

## Steps

- [ ] Write failing test `TestInferTypesCallFromDeclaration` in `internal/infer/declared_test.go`: a call to an imported sub declared with a return type in a `.tpm` is typed with it
- [ ] Run test to verify failure: `go test ./internal/infer/ -run '^TestInferTypesCallFromDeclaration$' -count=1 -v`
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestInferSelectsMultiAtCallSite`: `localtime(0)` in scalar and list context, `each %h` against `each @a`, and an undecided call joining
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing e2e test `TestPSCCheckUsesDeclarations` in `test/e2e/psc_check_test.go`: `psc check` on a fixture reports a coercion mismatch that only the declaration's types reveal
- [ ] Run test to verify failure: `go test ./test/e2e/ -run '^TestPSCCheckUsesDeclarations$' -count=1 -v`
- [ ] Make it pass; `make test`; commit
- [ ] Write failing tests `TestInferArityChoosesCandidate` (CORE's `select`: `my $h = select($fh)` is typed `Str`, `my @r = select(undef, undef, undef, 0.25)` is typed `List`) and `TestInferArgTypesChooseSameArityCandidate` (a fixture multi with two same-arity candidates whose parameter types are disjoint, e.g. `(Int $n) Int` and `(Code $c) Str`: an `Int` argument types the call `Int`, a code-ref argument `Str`)
- [ ] Run to verify failure: `go test ./internal/infer/ -run '^TestInferArityChoosesCandidate$|^TestInferArgTypesChooseSameArityCandidate$' -count=1 -v`
- [ ] Implement; run to verify pass; commit
- [ ] Write failing tests `TestInferUnimportedExportOkUntyped` and `TestInferLocalSubShadowsDeclaration` as "Fixtures" above describes; implement; run to verify pass; commit
- [ ] Write failing test `TestInferNoCandidateArityIsError` in `internal/infer/declared_test.go`: a call to a library multi with no prototype whose arity no candidate accepts is reported as an error, and gets no joined type
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Doc: in the "Declaration Files" section of `internal/data/help/types.md`, say that `psc check` types a call to a declared sub from its declaration, using the e2e fixture's diagnostic as the example; keep "Typed Perl is valid only in `.tpm` files, and you cannot yet add your own" (a user search path is RFC open question 6); `go test ./internal/cli/ -run '^TestShowTypesHelp$' -count=1`; commit

## Acceptance criteria

### Positive Scenarios
- [ ] infer types a call from its .tpm declaration (`go test ./internal/infer/ -run '^TestInferTypesCallFromDeclaration$' -count=1 -v | grep -q '^--- PASS: TestInferTypesCallFromDeclaration'`)
- [ ] infer selects a multi candidate at the call site (`go test ./internal/infer/ -run '^TestInferSelectsMultiAtCallSite$' -count=1 -v | grep -q '^--- PASS: TestInferSelectsMultiAtCallSite'`)
- [ ] psc check reports what a declaration reveals (`go test ./test/e2e/ -run '^TestPSCCheckUsesDeclarations$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckUsesDeclarations'`)
- [ ] a call into a module with no .tpm gets no declared type (`go test ./internal/infer/ -run '^TestInferUndeclaredImportUntyped$' -count=1 -v | grep -q '^--- PASS: TestInferUndeclaredImportUntyped'`)
- [ ] a scalar sort call site has no type (`go test ./internal/infer/ -run '^TestInferScalarSortHasNoType$' -count=1 -v | grep -q '^--- PASS: TestInferScalarSortHasNoType'`)
- [ ] psc check reports no mismatch where arguments fit a declaration (`go test ./test/e2e/ -run '^TestPSCCheckDeclarationsNoFalsePositive$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckDeclarationsNoFalsePositive'`)
- [ ] a call whose arity no candidate accepts is an error at the call site (`go test ./internal/infer/ -run '^TestInferNoCandidateArityIsError$' -count=1 -v | grep -q '^--- PASS: TestInferNoCandidateArityIsError'`)
- [ ] arity chooses between candidates at a call site: `select($fh)` is `Str`, 4-argument `select` is `List` (`go test ./internal/infer/ -run '^TestInferArityChoosesCandidate$' -count=1 -v | grep -q '^--- PASS: TestInferArityChoosesCandidate'`)
- [ ] argument types choose between two same-arity candidates at a call site (`go test ./internal/infer/ -run '^TestInferArgTypesChooseSameArityCandidate$' -count=1 -v | grep -q '^--- PASS: TestInferArgTypesChooseSameArityCandidate'`)

### Negative Scenarios
- [ ] an Undecided multi call (e.g. `return localtime($t)` in a sub whose calling context is unknown) joins its candidates and psc check reports nothing for it (`go test ./test/e2e/ -run '^TestPSCCheckUndecidedCallReportsNothing$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckUndecidedCallReportsNothing'`)
- [ ] a name a declaration marks `:export_ok`, which the `use` line did not ask for, gets no declared type at a call (bare `use My::OkOnly; f($x)`, with `My/OkOnly.tpm` a fixture in the test declarations dir declaring `sub f :export_ok`, not Scalar::Util.tpm) (`go test ./internal/infer/ -run '^TestInferUnimportedExportOkUntyped$' -count=1 -v | grep -q '^--- PASS: TestInferUnimportedExportOkUntyped'`)
- [ ] a sub the file defines itself, with the same non-builtin name as an imported declared sub, called after its definition, is typed as perl calls it (the file's own definition), not from the declaration (`go test ./internal/infer/ -run '^TestInferLocalSubShadowsDeclaration$' -count=1 -v | grep -q '^--- PASS: TestInferLocalSubShadowsDeclaration'`)
- [ ] a single fitting-arity candidate given a mismatched argument is reported as coercion-mismatch, not as a call no candidate accepts (`go test ./test/e2e/ -run '^TestPSCCheckSingleCandidateMismatchIsCoercion$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckSingleCandidateMismatchIsCoercion'`)
