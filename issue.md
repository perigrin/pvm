---
title: Give a qualified call its declaration's signatures, so Carp::confess is an exit and typed
state: pending
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
- 01a0fd9b-5700-7514-a143-2caed2d06d8b
blocks:
- 01a1464e-de34-7403-95d2-25f4855404b2
created: 2026-10-16T20:02:15.325366834Z
updated: 2026-10-16T21:00:21.014794589Z
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
