---
title: "Parse aliased parameters (\\$x, \\@a, \\%h, \\&c, \\(@list)) and derive \\$, \\@, \\% and \\& from them"
state: done
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-43ea-716e-9bb8-dd93ce3bbfcb
- 01a0fd9b-46c4-7a8c-b8d6-498c5182a47f
blocks:
- 01a0fd9b-4f8d-71ca-bac4-24eceb1dc4a4
- 01a0fd9b-5698-70cf-b6e4-8899a83b0fe0
- 01a10dc1-3499-7e2b-9e6f-d935084437fa
- 01a10dc1-3530-762f-9a03-177445629c6b
created: 2026-10-02T17:13:27.376660947Z
updated: 2026-10-16T18:48:31.291118101Z
sessions:
- start_sha: 3c0c517fbcb6634d7c96c94b50430cebf214de5d
  end_sha: 3c0c517fbcb6634d7c96c94b50430cebf214de5d
  commits: 0
  started_at: 2026-10-06T08:56:18.108079595Z
  ended_at: 2026-10-06T08:56:18.282478383Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-06T08:56:18.108079595Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-06T08:56:18.282478383Z
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

RFC 0001 "The scalar container" (perigrin, 2026-10-02): a parameter that aliases the caller's container is written with a backslash, as Perl writes aliasing (perlref, "Assigning to References"). `Scalar \$x` derives `\$`, `Array \@a` derives `\@`, `Hash \%h` derives `\%`, `Code \&c` derives `\&` (a code slot, perlref's `\&sub` refaliasing form: `\&foo = \&bar` makes `foo()` mean `bar()`, measured), so `sub any (Code \&block, List @list) Boolean;` derives perl's `\&@` (measured: `prototype("CORE::any")` and `prototype("CORE::all")` are `\&@`); `sub push (Array \@a, List @list) Int;` derives `\@@`, and in it only the final parameter is slurpy (the caller's view; perl's own signatures would reject two aggregates). `\(@args)` is perlref's list form, a reference to each element: `sub chomp (List[Str] \(@args = ($_))) Int;` says chomp writes through every argument (measured: `chomp(@l, $x)` chomps both, bare `chomp` chomps `$_`). Record the alias on `types.Param` (a flag, plus the list form). There is no `:lvalue` parameter attribute. Measured on 5.42, `\$` accepts any scalar lvalue (`$x`, `$h{k}`, `$a[0]`, `f()->[0]`, `$x = 7`, `substr($x, 0, 1)`); refusing what perl refuses at call sites is not this issue's: constants, sub results and aggregates in a `\$` slot are the `\$`-slot refusal issue's, and wrong arity the arity refusal issue's. A glob slot (`\*`) has no spelling (RFC open question 11): `Glob \*g` is not read. What an aliased list is as a type stays the paper's question. This completes the section, so it re-marks "The scalar container". Under the RFC's convention (perigrin, 2026-10-05, commit 72e01c14), a section is Implemented when its rule is built and tested: the re-mark here covers the rule, and the section's example declarations (`sref`, `push`, `chomp`, `any`) are illustrations that belong to the typing batches -- `push` to the measured-rows issue, `chomp` to the no-prototype builtins issue, `any` to the array, list and hash batch -- which write them into CORE.pmt and test them against perl; the tests here use literal declarations. The section's one directive about specific lines, that the six glob-slot builtins (`lock`, `undef`, `tie`, `tied`, `untie`, `pos`) keep prototype-only lines in CORE.pmt, untyped, holds today (those lines are prototype-only now) and is enforced by the last typing batch's every-line check (`TestCoreEveryPrototypedLineIsTyped` and `TestCoreEveryLineCheckExemptsOnlyGlobSlots`), which fails if any of the six is typed.

Also covers a container type before an aliased hash or array: `Hash[Str] \%h` and `Array[Str] \@a` must parse (today both fail with "type Hash[Str] names no variable", as `Array \@a` does). The flattening-sigil error added by 01a0fd9b-43ea-716e-9bb8-dd93ce3bbfcb already points users to that spelling. The lexer's `containerTypeSigil` (internal/lexer/proto.go) reads `%` after a container's `]` as a sigil, but nothing reads it after `\`.

From the alignment gate on 01a0fd9b-4a5b-7e94-ad87-756dfcfe1f4c: `types.Select`'s `accepts` treats every sigil other than `$` as slurpy, including `&`, so a Go-built `(Code &cb)` called with three arguments is Selected -- an arity failure rescued. Latent until a .pmt can write `Code &cb`; when this issue makes `&` and backslashed parameters readable, make `accepts` count each as exactly one argument, with a test.

## Steps

- [ ] Write failing test `TestPmtAliasedParamsParse` in `internal/parse/pmt_test.go`: `Scalar \$x`, `Array \@a`, `Hash \%h` parse as aliased params, and `List[Str] \(@args = ($_))` parses as an aliased list with its default
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestPmtAliasedParamsParse$' -count=1 -v`
- [ ] Implement the backslash in the param reader; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestAliasedParamsDerivePrototypes` in `internal/parse/derive_test.go`: `Scalar \$x` derives `\$`, `Array \@a` derives `\@`, `Hash \%h` derives `\%`; `:prototype(\$)`, `:prototype(\@)` and `:prototype(\%)` derive the backslashed params; each round trip holds
- [ ] Implement the three rows; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestPushRoundTripsRefArrayAt`: `sub push (Array \@a, List @list) Int;` derives `\@@` with only the final param slurpy, and `:prototype(\@@)` derives `(Array \@a, List @list)`
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestCodeAliasDerivesRefAmp` in `internal/parse/derive_test.go`: `Code \&c` parses as an aliased code param and derives `\&`; `sub any (Code \&block, List @list) Boolean;` derives `\&@`; `:prototype(\&@)` derives `(Code \&block, List @list)`; each round trip holds
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCodeAliasDerivesRefAmp$' -count=1 -v`
- [ ] Implement the row; run test to verify pass; `make test`; commit
- [ ] Re-mark the RFC heading "The scalar container" `(*Implemented*, <sha>)`; commit

### Negative Scenarios

- the retired attribute is not read: `sub f (Scalar $x :lvalue);` is an error, and so is any parameter attribute (`Scalar $x :bogus`), deriving no `\$` (`go test ./internal/parse/ -run '^TestParamAttributeRefused$' -count=1 -v | grep -q '^--- PASS: TestParamAttributeRefused'`)
- a backslash the RFC does not define is an error, not a guess: `Str \@a` (a non-container type on an aliased array), `\(@args` (unclosed list form), `\(Str $x)` (a list form holding a typed scalar), `Glob \*g` (a glob slot, which has no spelling yet) (`go test ./internal/parse/ -run '^TestPmtMalformedAliasedParamRefused$' -count=1 -v | grep -q '^--- PASS: TestPmtMalformedAliasedParamRefused'`)
- Array against List is a disagreement: `:prototype(\@) (List @l)` and `:prototype(@) (Array \@a)` are errors in the declaration file (`go test ./internal/parse/ -run '^TestAliasedPrototypeDisagreementIsError$' -count=1 -v | grep -q '^--- PASS: TestAliasedPrototypeDisagreementIsError'`)
- the backslash is typed Perl and `.pmt`-only: ordinary source `sub f (\$x) { 1 }` without signatures keeps `(\$x)` as a prototype and records no aliased param (`go test ./internal/parse/ -run '^TestAliasedParamOnlyInPmt$' -count=1 -v | grep -q '^--- PASS: TestAliasedParamOnlyInPmt'`)

## Acceptance criteria

- [ ] \$x, \@a, \%h and \(@list) parse as aliased parameters (`go test ./internal/parse/ -run '^TestTpmAliasedParamsParse$' -count=1 -v | grep -q '^--- PASS: TestTpmAliasedParamsParse'`)
- [ ] aliased parameters derive \$, \@ and \% and back (`go test ./internal/parse/ -run '^TestAliasedParamsDerivePrototypes$' -count=1 -v | grep -q '^--- PASS: TestAliasedParamsDerivePrototypes'`)
- [ ] Code \&c derives \& and any round-trips \&@ (`go test ./internal/parse/ -run '^TestCodeAliasDerivesRefAmp$' -count=1 -v | grep -q '^--- PASS: TestCodeAliasDerivesRefAmp'`)
- [ ] push round-trips \@@ (`go test ./internal/parse/ -run '^TestPushRoundTripsRefArrayAt$' -count=1 -v | grep -q '^--- PASS: TestPushRoundTripsRefArrayAt'`)
- [ ] the RFC section on the scalar container is implemented (`grep -q '^### The scalar container.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] a parameter attribute such as :lvalue is refused (`go test ./internal/parse/ -run '^TestParamAttributeRefused$' -count=1 -v | grep -q '^--- PASS: TestParamAttributeRefused'`)
- [ ] a malformed aliased parameter is an error (`go test ./internal/parse/ -run '^TestTpmMalformedAliasedParamRefused$' -count=1 -v | grep -q '^--- PASS: TestTpmMalformedAliasedParamRefused'`)
- [ ] an Array/List disagreement between prototype and types is an error (`go test ./internal/parse/ -run '^TestAliasedPrototypeDisagreementIsError$' -count=1 -v | grep -q '^--- PASS: TestAliasedPrototypeDisagreementIsError'`)
- [ ] aliased parameters are not read in ordinary source (`go test ./internal/parse/ -run '^TestAliasedParamOnlyInTpm$' -count=1 -v | grep -q '^--- PASS: TestAliasedParamOnlyInTpm'`)
