---
title: Type CORE.pmt's string, numeric and regex builtins
state: in-progress
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4f8d-71ca-bac4-24eceb1dc4a4
blocks:
- 01a10c6e-9a95-719a-80e2-2e33037d0096
created: 2026-10-03T11:12:37.957978283Z
updated: 2026-10-06T11:15:05.095552868Z
sessions:
- start_sha: bc832959484d0baaa61b2e4c0358264fb3f1aa62
  end_sha: ""
  commits: 0
  started_at: 2026-10-06T11:15:05.095552868Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-06T11:15:05.095552868Z
---

## Prerequisites

- [ ] Check CORE.pmt's derived prototypes against perl, with signatures.go's measured rows

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/coretypes_test.go`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `TestCoreDerivedPrototypesArePerls` (`internal/parse/coretypes_test.go`, the harness); perlfunc's own categories, `perl -MPod::Functions -e 'print "@{$Kinds{String}} @{$Kinds{Regexp}} @{$Kinds{Math}}"'` ("Perl Functions by Category" in `perldoc perlfunc`)

First of five batches typing CORE.pmt's prototyped builtins, batched by perlfunc's own categories: this one is "Functions for SCALARs or strings", "Regular expressions and pattern matching" and "Numeric functions" (Pod::Functions kinds `String`, `Regexp`, `Math`), intersected with CORE.pmt's prototyped lines and without `pos` -- 29 names measured on 5.42, of which the measured-rows issue already typed 14. `pos` (`;\[$*]`) is in `Regexp` but holds a glob slot, which has no spelling yet (RFC open question 11, perigrin 2026-10-03): its line stays prototype-only, untyped, and is left out of this batch's names; the last batch's every-line check exempts it with the other five glob-slot lines. Every type is measured on perl 5.42 and the measurement recorded as a comment above its line, as signatures.go's were; element types the paper has not defined are written as plain `List`. Each line is held to the harness, so its derived prototype must be perl's. Hold the batch's names in the test, computed once with the command above, never against a hand-kept table. Re-marks no RFC section.

## Steps

- [ ] Measure the batch's names with Pod::Functions and perl; write failing test `TestCoreStringNumericRegexBuiltinsTyped` in `internal/parse/coretypes_test.go`: every name in the batch has a typed CORE.pmt line
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreStringNumericRegexBuiltinsTyped$' -count=1 -v`
- [ ] Type the lines with their measurement comments, measuring with perl where the type is not obvious; run the harness and the test to verify pass; `make test`; commit

### Negative Scenarios

- the batch check is not tautological: the same check over a copy of CORE.pmt with one batch line left untyped (`sub crypt :prototype($$);`) reports it, naming `crypt` (`go test ./internal/parse/ -run '^TestCoreBatchCheckCatchesUntypedLine$' -count=1 -v | grep -q '^--- PASS: TestCoreBatchCheckCatchesUntypedLine'`)
- feature-gated keywords in the batch (`fc`, available only under the `fc` feature or as `CORE::fc`) are typed like the rest, and typing them leaves their parse unchanged in source without the feature (`go test ./internal/parse/ -run '^TestCoreFeatureGatedBuiltinTypedParseUnchanged$' -count=1 -v | grep -q '^--- PASS: TestCoreFeatureGatedBuiltinTypedParseUnchanged'`)

## Acceptance criteria

- [ ] CORE.pmt types every string, numeric and regex builtin (`go test ./internal/parse/ -run '^TestCoreStringNumericRegexBuiltinsTyped$' -count=1 -v | grep -q '^--- PASS: TestCoreStringNumericRegexBuiltinsTyped'`)
- [ ] every CORE.pmt prototype, derived from its types where typed, is still perl's (`go test ./internal/parse/ -run '^TestCoreDerivedPrototypesArePerls$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedPrototypesArePerls'`)
- [ ] the batch check reports an untyped line (`go test ./internal/parse/ -run '^TestCoreBatchCheckCatchesUntypedLine$' -count=1 -v | grep -q '^--- PASS: TestCoreBatchCheckCatchesUntypedLine'`)
- [ ] typing a feature-gated builtin leaves its parse unchanged (`go test ./internal/parse/ -run '^TestCoreFeatureGatedBuiltinTypedParseUnchanged$' -count=1 -v | grep -q '^--- PASS: TestCoreFeatureGatedBuiltinTypedParseUnchanged'`)
