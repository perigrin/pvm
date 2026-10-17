---
title: Type calls to imported library subs from their .tpm declarations and select multis at call sites
state: in-progress
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
- 01a14683-8405-743e-9c90-7bb5ff9cec88
created: 2026-10-02T17:13:31.136589197Z
updated: 2026-10-17T03:33:37.034700449Z
sessions:
- start_sha: cc1416816b43a75d9d4de73e9269452b26f61f27
  end_sha: a35012438663295404da21ff5a6a7ec965d1d93a
  commits: 3
  started_at: 2026-10-16T23:12:50.914054892Z
  ended_at: 2026-10-17T00:07:03.288585122Z
- start_sha: 6c727e93d5c2761bfd2d0e4f8eb56a53088e130d
  end_sha: 51c907f24c2c8cb98e47b5e665473a52405aa966
  commits: 2
  started_at: 2026-10-17T00:21:50.805673827Z
  ended_at: 2026-10-17T00:45:17.558269979Z
- start_sha: 3bd5ce632c0e695e117434bf3358a8e53a15f8bd
  end_sha: 8d17e0b3c88c2e801f39107186c29dd64c8491c8
  commits: 3
  started_at: 2026-10-17T00:52:11.547877905Z
  ended_at: 2026-10-17T01:07:27.493080025Z
- start_sha: 8d17e0b3c88c2e801f39107186c29dd64c8491c8
  end_sha: ""
  commits: 0
  started_at: 2026-10-17T03:33:37.034700449Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T23:12:50.914054892Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-17T00:07:03.288585122Z
- state: reopened
  actor: human:git-zhi
  timestamp: 2026-10-17T00:21:50.587984361Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-17T00:21:50.805673827Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-17T00:45:17.558269979Z
- state: reopened
  actor: human:git-zhi
  timestamp: 2026-10-17T00:52:11.30950209Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-17T00:52:11.547877905Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-17T01:07:27.493080025Z
- state: reopened
  actor: human:git-zhi
  timestamp: 2026-10-17T03:33:36.983293033Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-17T03:33:37.034700449Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
- internal/data/help/types.md
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
- internal/infer/testdata/oracle/Carp.pl
- internal/infer/testdata/return_oracle.pl
- internal/parse/decl.go
- internal/parse/declarations.go
- internal/parse/declarations/Carp.tpm
- internal/parse/exportattr_test.go
- internal/parse/heldtoperl_test.go
- internal/parse/library_tpm_test.go
- internal/parse/qualified_test.go
- internal/parse/testdata/declarations/My/Multi.tpm
- internal/parse/testdata/declarations/My/OkOnly.tpm
- internal/parse/testdata/declarations/My/UntypedVars.tpm
- internal/parse/testdata/held_to_perl.pl
- internal/parse/typedour_test.go
- internal/parse/use.go
- internal/perltest/perltest.go
- internal/types/declared.go
- internal/types/select.go
- internal/types/select_test.go
- test/e2e/psc_check_test.go
- test/e2e/testdata/check/declared_fit.pl
- test/e2e/testdata/check/declared_mismatch.pl
- test/e2e/testdata/check/declared_single.pl
- test/e2e/testdata/check/undecided_call.pl
- test/e2e/testdata/check/undecided_declared.pl
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

### Review Findings from da0a's gate (forward precision note)

(confidence 70) `internal/types/select.go:77-79`: when no candidate fits the argument types, the fallback replaces every argument with Unknown, discarding arguments that did discriminate. Candidates `(Int a, Str b) Int` and `(Code a, Str b) ArrayRef` with arguments `(Int, FileHandle)` give Undecided `Int|ArrayRef`, though `Int` rules out the Code candidate. Sound but imprecise. When this issue feeds concrete argument types from library multis, set to Unknown only the arguments that fit no candidate, then run the normal fit and mostSpecific, if a criterion here depends on it.

### Review Findings

Tier-2 gate, 2026-10-17, on a3501243 (fresh reviewer subagent; lenses applied inline; psc check over perl 5.42's privlib identical to the parent; exact 47 of 56 under perl 5.42). Severity of Undef->Str as error matches builtins' existing policy; OverlayDeclarations justified. VERDICT: REOPEN.

1. (in scope, 85) `internal/infer/calls.go:91-113` declaredCall counts a flattening argument as one argument when selecting: `use My::Multi; my @a=(1,2); my $r = two(@a);` picks `two (Int $a) Int` and types the call Int; `two(%h)` likewise. Perl gives `two(@a)` two arguments and returns the two-argument candidate's value (perl -wle prints `two` for both). The guard skips only the arity error; selection still uses the wrong count, breaking "selects by arity". Fix: when an argument flattens in a call to a multi with no prototype, do not let Select choose by arity: join the candidates whose arity can take at least the non-flattening arguments (or Undecided). Test: `two(@a)` must not be Int.
2. (in scope, 75) `internal/infer/calls.go:98`: no test reaches declaredCall's Undecided path. Mutation A (`case sel.Outcome != types.Selected && !flattens:`, reporting every Undecided library call as an arity error) survives all 13 criteria; the "Undecided multi joins and psc check reports nothing" criterion is met only through builtin localtime (callReturns, not this code) and passes at the parent. Fix: add a library-multi undecided case, e.g. `use My::Lib; sub g { return when_() }` expecting `Str|List` and no diagnostics, in infer and in the e2e fixture; show mutation A fails it.

### Review Findings (pass 2)

Tier-2 gate pass 2, 2026-10-17, on 51c907f2 (fresh reviewer subagent; overlay mutations prove both pass-1 findings fixed; perl 5.42 agrees on two(@a), two(%h), two(1,@e), two(1,2,3,@e)). OverlayDeclarations in e2e does not leak (shuffle 1-3). VERDICT: REOPEN.

A. (in scope, 90; introduced by b935fd11) `internal/infer/calls.go:138-142` flattenedSelect's loop `for count := least; count <= most+1; count++` never runs when the fixed arguments exceed most+1, so a slurpy call becomes Failed: `use My::Lib; my @a=(1); my $s = join_all(1, 2, 3, 4, 5, @a);` against `sub join_all (Str $sep, List[Str] @parts) Str` gives Unknown and "no declared candidate takes at least 5 argument(s) [arity-mismatch]"; the parent gives Str with no diagnostic; perl 5.42 runs it. Fix: loop to max(least, most+1); test `join_all(1, 2, 3, 4, 5, @a)` expecting Str and no diagnostics, failing first.
B. (folded in by the orchestrator, same function; 65) `internal/infer/calls.go:96-97` `flattens` looks only at the argument's type, so an `@a` passed to a backslashed `Array \@a` parameter, which perl passes as one container, is counted as zero or more: `multi f (Array \@a) Int; multi f (Array \@a, Array \@b) Str;` makes `f(@x)` Int|Str instead of Int, and skips coercion checks. Fix: count an Array or Hash argument as flattening only where no candidate's parameter in that position is backslashed; a test.

### Review Findings (pass 3)

Tier-2 gate pass 3, 2026-10-17, on 8d17e0b3 (fresh reviewer subagent; overlays prove pass-2 A and B fixed; 8405 interplay probed clean; precision 65/76 WRONG 0; lib baseline holds). VERDICT: REOPEN (third and final reopen allowed).

1. (in scope, 75; introduced by b8b5f2c1) `internal/infer/calls.go:99-103,157-163` aliasedAt counts an array as one fixed argument when SOME candidate backslashes that position, so disagreeing candidates give a type that is wrong, not merely wide: `multi sub d (Array \@a) Int; multi sub d (List @l) Str;` makes `d(@x)` Int; `(Array \@a) Int` beside `(Int $a, Int $b, Int $c) Str` makes `e(@x)` Int. The loader (`prototypeFromCandidates`, internal/parse/derive.go:154; declarations.go:198-200) gives such a multi NO prototype, and without a prototype perl flattens: `sub h { scalar @_ } h(@x)` prints 3. The fixture `box` ((\@) and (\@\@)) also derives no prototype, so its Int test rests on a `\@;\@` prototype the loader does not give it.
   Resolution (orchestrator, by perl and perigrin's ruling that call resolution comes from the parser): an argument is passed as one container only when the prototype perl applies to this call (the one the parser resolved for the sub, as `lower.Call` carries it) backslashes that position; with no resolved prototype, an array or hash argument flattens. Tests: `d(@x)` and `e(@x)` are the join, failing first; `box`'s cases follow from box's resolved prototype (measure with perl what a sub with no prototype does with `box(@x)`), and if the fixture's intent needs `\@;\@`, give it that prototype explicitly so the test measures what perl does.

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
