---
title: Give a qualified call its declaration's signatures, so Carp::confess is an exit and typed
state: done
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
- 01a0fd9b-5700-7514-a143-2caed2d06d8b
blocks:
- 01a1464e-de34-7403-95d2-25f4855404b2
created: 2026-10-16T20:02:15.325366834Z
updated: 2026-10-17T03:41:54.843258937Z
sessions:
- start_sha: a35012438663295404da21ff5a6a7ec965d1d93a
  end_sha: 3bd5ce632c0e695e117434bf3358a8e53a15f8bd
  commits: 4
  started_at: 2026-10-17T00:07:41.393710811Z
  ended_at: 2026-10-17T00:52:11.478193727Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-17T00:07:41.393710811Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-17T00:52:11.478193727Z
observed_paths:
- internal/infer/calls.go
- internal/infer/declared_test.go
- internal/infer/flow_test.go
- internal/parse/declarations.go
- internal/parse/exportattr_test.go
- internal/parse/qualified_test.go
- internal/parse/use.go
- test/e2e/psc_check_test.go
- test/e2e/testdata/check/undecided_declared.pl
---

## Prerequisites

- [ ] Type calls to imported library subs from their .tpm declarations and select multis at call sites (01a0fd9b-5700)

## Context

- paths: `internal/parse/use.go`, `internal/parse/arity.go`, `internal/parse/qualified_test.go` (new), `internal/infer/declared_test.go`, `internal/infer/flow_test.go`
- commands: `go test ./internal/parse/ -run '^TestQualifiedCallCarriesSignatures$' -count=1 -v`, `go test ./internal/infer/ -run '^TestInferTypesQualifiedCall$|^TestQualifiedCroakConfessAreExits$' -count=1 -v`
- entrypoints: `use.go` (~line 300-320, `p.moduleSubs` records `Import{Name, Prototype, PrototypeKnown}` with no `Signatures`), `lookupSub` (`shape.go:172`), `markResolution` (`arity.go:~139`), `exits` (`internal/infer/flow.go:~624`)

A qualified call (`Carp::confess(...)`, `Scalar::Util::blessed($x)`) resolves through `lookupSub` to a `moduleSubs` entry that carries no signatures, so it reaches `lower.Call.Signatures` empty: infer cannot type it and `exits` does not see it end its path. Give `moduleSubs` entries the declaration's signatures so `markResolution` sets `Node.Signatures` for a qualified call as it does for an imported one. infer needs no name lookup for this (perigrin, 2026-10-16: inference takes call resolution from the parser); with 01a0fd9b-5700 done it reads `Call.Signatures` and the qualified call is typed. Bare croak/confess are already exits (`TestExitsAndGuardsByCoreAndCarp`); this adds the qualified forms, which the final gate checks.

## Steps

- [ ] Write failing test `TestQualifiedCallCarriesSignatures`: after `use Carp ();`, the node for `Carp::confess("x")` has Carp.tpm's signature, and `Scalar::Util::blessed($x)` without an import has its declaration's once Scalar/Util.tpm exists (use a test declarations dir via the `declarations` fs.FS var until then)
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestQualifiedCallCarriesSignatures$' -count=1 -v`
- [ ] Implement in `use.go`/`arity.go`; run to verify pass; commit
- [ ] Write failing test `TestQualifiedCroakConfessAreExits` in `internal/infer/flow_test.go`: a path ending in `Carp::croak(...)` or `Carp::confess(...)` ends as `die` does, with `use Carp;` and with `use Carp ();`
- [ ] Write failing test `TestInferTypesQualifiedCall` in `internal/infer/declared_test.go`: `my $s = Carp::shortmess("x");` is typed Str from Carp.tpm
- [ ] Run to verify failure, implement anything missing, run to verify pass; commit
- [ ] Run the m1 gates named in the final-gate issue, with the commands it gives them; a ratchet may only be raised. A drop is fixed in the .tpm, or the batch stops and asks perigrin. `make test`; commit

### Review Findings

Tier-2 gate, 2026-10-17, on 3bd5ce63 (fresh reviewer subagent; five overlay mutations; all m1 gates pass, exact 47 of 56; psc check over perl 5.42's privlib identical with and without the fix; perl 5.42 measurements). VERDICT: PASS with two minor in-scope test findings, so reopened per the gate rule.

A. (in scope, minor, 65) `internal/parse/qualified_test.go:46-52` TestQualifiedCallUndeclaredNameHasNoSignatures is near-vacuous: `Carp::nosuch` is in neither protos nor signatures, so no mutation of this fix fails it. The discriminating case is a name in protos but not in the .tpm's signatures (an XS sub added at module.go:385-395 after signatures are derived). Fix: a fixture .tpm plus a `.xs` loader entry; assert the XS name carries no signatures; show it fails when every protos name gets a signature.
B. (in scope, cosmetic, 70) `internal/parse/qualified_test.go:32` comment says "loaded with no import list at all" but the source is `use Scalar::Util ()`, an empty import list. Reword.

Out of scope, filed separately: `require M` then `M::f(...)` carries no signatures; a module loaded only transitively; `&M::f(...)` with a None signature is not an exit; `goto &M::f` is not an exit.

## Acceptance Criteria

### Positive Scenarios
- [ ] a qualified call carries its declaration's signatures (`go test ./internal/parse/ -run '^TestQualifiedCallCarriesSignatures$' -count=1 -v | grep -q '^--- PASS: TestQualifiedCallCarriesSignatures'`)
- [ ] qualified croak and confess end their path (`go test ./internal/infer/ -run '^TestQualifiedCroakConfessAreExits$' -count=1 -v | grep -q '^--- PASS: TestQualifiedCroakConfessAreExits'`)
- [ ] a qualified call is typed from its declaration (`go test ./internal/infer/ -run '^TestInferTypesQualifiedCall$' -count=1 -v | grep -q '^--- PASS: TestInferTypesQualifiedCall'`)

### Negative Scenarios
- [ ] a qualified call to a name the module's `.tpm` does not declare, such as `Carp::nosuch(1)`, carries no signatures (`go test ./internal/parse/ -run '^TestQualifiedCallUndeclaredNameHasNoSignatures$' -count=1 -v | grep -q '^--- PASS: TestQualifiedCallUndeclaredNameHasNoSignatures'`)
- [ ] a qualified call into a module with no `.tpm` carries no signatures (`go test ./internal/parse/ -run '^TestQualifiedCallUndeclaredModuleHasNoSignatures$' -count=1 -v | grep -q '^--- PASS: TestQualifiedCallUndeclaredModuleHasNoSignatures'`)
- [ ] `Foo::confess(...)`, where package Foo defines its own `confess`, does not end its path (`go test ./internal/infer/ -run '^TestQualifiedCallOtherPackageConfessNotExit$' -count=1 -v | grep -q '^--- PASS: TestQualifiedCallOtherPackageConfessNotExit'`)
- [ ] a code reference `\&Carp::confess` is not a call and does not end its path (`go test ./internal/infer/ -run '^TestQualifiedCodeRefIsNotExit$' -count=1 -v | grep -q '^--- PASS: TestQualifiedCodeRefIsNotExit'`)
