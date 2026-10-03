---
title: "Parse aliased parameters (\\$x, \\@a, \\%h, \\(@list)) and derive \\$, \\@ and \\% from them"
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-43ea-716e-9bb8-dd93ce3bbfcb
- 01a0fd9b-46c4-7a8c-b8d6-498c5182a47f
blocks:
- 01a0fd9b-4f8d-71ca-bac4-24eceb1dc4a4
- 01a0fd9b-5017-78d0-ae38-331cbe6a9866
created: 2026-10-02T17:13:27.376660947Z
updated: 2026-10-03T11:14:17.322780072Z
---

## Prerequisites

- [ ] Enforce the slurpy rule and a final List parameter in .pmt signatures
- [ ] Derive prototypes from typed signatures and typed signatures from prototypes

## Context

- paths: `internal/parse/decl.go`, `internal/parse/derive.go`, `internal/parse/derive_test.go`, `internal/parse/pmt_test.go`, `internal/types/declared.go`, `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `internal/parse/derive.go`; the param reader in `internal/parse/decl.go`; `types.Param` (`internal/types/declared.go`)

RFC 0001 "The scalar container" (perigrin, 2026-10-02): a parameter that aliases the caller's container is written with a backslash, as Perl writes aliasing (perlref, "Assigning to References"). `Scalar \$x` derives `\$`, `Array \@a` derives `\@`, `Hash \%h` derives `\%`; `sub push (Array \@a, List @list) Int;` derives `\@@`, and in it only the final parameter is slurpy (the caller's view; perl's own signatures would reject two aggregates). `\(@args)` is perlref's list form, a reference to each element: `sub chomp (List[Str] \(@args = ($_))) Int;` says chomp writes through every argument (measured: `chomp(@l, $x)` chomps both, bare `chomp` chomps `$_`). Record the alias on `types.Param` (a flag, plus the list form). There is no `:lvalue` parameter attribute. Measured on 5.42, `\$` accepts any scalar lvalue (`$x`, `$h{k}`, `$a[0]`, `f()->[0]`, `$x = 7`, `substr($x, 0, 1)`); refusing what perl refuses at call sites (constants, sub results, aggregates, wrong arity) is the parse-time refusal issue's, not this one's. What an aliased list is as a type stays the paper's question. This completes the section, so it re-marks "The scalar container".

## Steps

- [ ] Write failing test `TestPmtAliasedParamsParse` in `internal/parse/pmt_test.go`: `Scalar \$x`, `Array \@a`, `Hash \%h` parse as aliased params, and `List[Str] \(@args = ($_))` parses as an aliased list with its default
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtAliasedParamsParse$' -count=1 -v`
- [ ] Implement the backslash in the param reader; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestAliasedParamsDerivePrototypes` in `internal/parse/derive_test.go`: `Scalar \$x` derives `\$`, `Array \@a` derives `\@`, `Hash \%h` derives `\%`; `:prototype(\$)`, `:prototype(\@)` and `:prototype(\%)` derive the backslashed params; each round trip holds
- [ ] Implement the three rows; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestPushRoundTripsRefArrayAt`: `sub push (Array \@a, List @list) Int;` derives `\@@` with only the final param slurpy, and `:prototype(\@@)` derives `(Array \@a, List @list)`
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Re-mark the RFC heading "The scalar container" `(*Implemented*, <sha>)`; commit

### Negative Scenarios

- the retired attribute is not read: `sub f (Scalar $x :lvalue);` is an error, and so is any parameter attribute (`Scalar $x :bogus`), deriving no `\$` (`go test ./internal/parse/ -run '^TestParamAttributeRefused$' -count=1 -v | grep -q '^--- PASS: TestParamAttributeRefused'`)
- a backslash the RFC does not define is an error, not a guess: `Str \@a` (a non-container type on an aliased array), `\(@args` (unclosed list form), `\(Str $x)` (a list form holding a typed scalar) (`go test ./internal/parse/ -run '^TestPmtMalformedAliasedParamRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtMalformedAliasedParamRefused'`)
- Array against List is a disagreement: `:prototype(\@) (List @l)` and `:prototype(@) (Array \@a)` are errors in the declaration file (`go test ./internal/parse/ -run '^TestAliasedPrototypeDisagreementIsError$' -count=1 -v | grep -q '^--- PASS: TestAliasedPrototypeDisagreementIsError'`)
- the backslash is typed Perl and `.pmt`-only: ordinary source `sub f (\$x) { 1 }` without signatures keeps `(\$x)` as a prototype and records no aliased param (`go test ./internal/parse/ -run '^TestAliasedParamOnlyInPmt$' -count=1 -v | grep -q '^--- PASS: TestAliasedParamOnlyInPmt'`)

## Acceptance criteria

- [ ] \$x, \@a, \%h and \(@list) parse as aliased parameters (`go test ./internal/parse/ -run '^TestPmtAliasedParamsParse$' -count=1 -v | grep -q '^--- PASS: TestPmtAliasedParamsParse'`)
- [ ] aliased parameters derive \$, \@ and \% and back (`go test ./internal/parse/ -run '^TestAliasedParamsDerivePrototypes$' -count=1 -v | grep -q '^--- PASS: TestAliasedParamsDerivePrototypes'`)
- [ ] push round-trips \@@ (`go test ./internal/parse/ -run '^TestPushRoundTripsRefArrayAt$' -count=1 -v | grep -q '^--- PASS: TestPushRoundTripsRefArrayAt'`)
- [ ] the RFC section on the scalar container is implemented (`grep -q '^### The scalar container.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] a parameter attribute such as :lvalue is refused (`go test ./internal/parse/ -run '^TestParamAttributeRefused$' -count=1 -v | grep -q '^--- PASS: TestParamAttributeRefused'`)
- [ ] a malformed aliased parameter is an error (`go test ./internal/parse/ -run '^TestPmtMalformedAliasedParamRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtMalformedAliasedParamRefused'`)
- [ ] an Array/List disagreement between prototype and types is an error (`go test ./internal/parse/ -run '^TestAliasedPrototypeDisagreementIsError$' -count=1 -v | grep -q '^--- PASS: TestAliasedPrototypeDisagreementIsError'`)
- [ ] aliased parameters are not read in ordinary source (`go test ./internal/parse/ -run '^TestAliasedParamOnlyInPmt$' -count=1 -v | grep -q '^--- PASS: TestAliasedParamOnlyInPmt'`)
