---
title: Enforce the slurpy rule and a final List parameter in .pmt signatures
state: done
urgency: normal
milestone: type-libraries
blocked_by:
- 01a10177-49d9-7915-975e-2a8e3da943bf
blocks:
- 01a0fd9b-4850-704d-8c14-c15add6194f2
- 01a0fd9b-49b2-7ea9-bfa1-ba398baadc3f
- 01a0fd9b-46c4-7a8c-b8d6-498c5182a47f
created: 2026-10-02T17:13:26.249797267Z
updated: 2026-10-16T18:48:31.051349656Z
sessions:
- start_sha: 836d0d5516dd6cb86c7549e943180666e0e70813
  end_sha: 836d0d5516dd6cb86c7549e943180666e0e70813
  commits: 0
  started_at: 2026-10-06T01:28:27.799948383Z
  ended_at: 2026-10-06T01:28:28.223800193Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-06T01:28:27.799948383Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-06T01:28:28.223800193Z
---

## Prerequisites

- [ ] Parse return types and type expressions in .pmt declarations

## Context

- paths: `internal/parse/decl.go`, `internal/types/declared.go`, `internal/parse/pmt_test.go`, `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: the typed-signature and type-expression reader in `internal/parse/decl.go`; `types.Param`

RFC 0001 "A slurpy takes no bare element type", and the final-`List` rule from "A typed signature and a prototype say the same thing" (perigrin, 2026-10-02). `Str @args` is refused (ambiguous container, ambiguous element assertion); a slurpy is untyped or carries an explicit container type (`List[Str] @args`, parsed by the previous issue). A container type with a flattening sigil (`Array @a`, `Hash %h`) is not valid: a parameter that takes the caller's container is backslashed (`Array \@a`, the aliased-parameter issue's). A `List` parameter must be last: measured on 5.42, a signature `(@a, $x)` dies "Slurpy parameter not last" and a prototype `(@$)` warns "Prototype after '@'", since its `$` can never be filled; a declaration with anything after a `List` parameter is an error whose message points to `Array \@a` for a single array followed by more parameters (`\@$`). This completes the slurpy section, so it re-marks it; Under the RFC's convention (perigrin, 2026-10-05, commit 72e01c14), a section is Implemented when its rule is built and tested: the re-mark here covers the rule, and the section's example declaration (`print` with `List[Str] @args = ($_)`) is an illustration that belongs to the invocant-colon issue, which writes it into CORE.pmt and tests it against perl; the final-`List` rule is part of the typed-signature section, which the last CORE.pmt typing batch re-marks.

## Steps

- [ ] Write failing test `TestPmtBareSlurpyTypeRefused`: `sub f (Str @args);` and `sub f (Int %h);` are errors; untyped `sub f (@args);` and `sub f (List[Str] @args);` are not
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtBareSlurpyTypeRefused$' -count=1 -v`
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestPmtContainerWithFlatteningSigilRefused`: `sub f (Array @a);` and `sub f (Hash %h);` are errors naming the backslashed spelling
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Measure with perl, then write failing test `TestPmtNonFinalListParam`: `sub f (List @a, Str $x);` and `sub f (List @a, List @b);` are errors, and the message names `Array \@a`
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtNonFinalListParam$' -count=1 -v`
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Re-mark the RFC heading "A slurpy takes no bare element type" `(*Implemented*, <sha>)`; commit

### Negative Scenarios

- a union is still a bare element type: `sub f (Str|Undef @args);` is refused like `Str @args` (`go test ./internal/parse/ -run '^TestPmtBareUnionSlurpyRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtBareUnionSlurpyRefused'`)
- a `List` parameter that is not last is an error (perigrin, 2026-10-02), including after an untyped slurpy (`sub f (@a, $x);`); the message points to `Array \@a` (`go test ./internal/parse/ -run '^TestPmtNonFinalListParam$' -count=1 -v | grep -q '^--- PASS: TestPmtNonFinalListParam'`)
- `Array @a` is not quietly read as an aliased array: it records no parameter and derives no prototype (`go test ./internal/parse/ -run '^TestPmtContainerWithFlatteningSigilRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtContainerWithFlatteningSigilRefused'`)

## Acceptance criteria

- [ ] a slurpy with a bare element type is refused (`go test ./internal/parse/ -run '^TestTpmBareSlurpyTypeRefused$' -count=1 -v | grep -q '^--- PASS: TestTpmBareSlurpyTypeRefused'`)
- [ ] a container type with a flattening sigil is refused (`go test ./internal/parse/ -run '^TestTpmContainerWithFlatteningSigilRefused$' -count=1 -v | grep -q '^--- PASS: TestTpmContainerWithFlatteningSigilRefused'`)
- [ ] anything after a List parameter is an error pointing to Array \@a (`go test ./internal/parse/ -run '^TestTpmNonFinalListParam$' -count=1 -v | grep -q '^--- PASS: TestTpmNonFinalListParam'`)
- [ ] the RFC section on slurpies is implemented (`grep -q '^### A slurpy takes no bare.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] a slurpy typed with a bare union is refused (`go test ./internal/parse/ -run '^TestTpmBareUnionSlurpyRefused$' -count=1 -v | grep -q '^--- PASS: TestTpmBareUnionSlurpyRefused'`)
