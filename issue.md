---
title: Type CORE.pmt's array, list and hash builtins
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4f8d-71ca-bac4-24eceb1dc4a4
- 01a0fd9b-5017-78d0-ae38-331cbe6a9866
created: 2026-10-03T11:12:38.523502989Z
updated: 2026-10-03T11:16:05.37221223Z
---

## Prerequisites

- [ ] Check CORE.pmt's derived prototypes against perl, with signatures.go's measured rows
- [ ] Declare the no-prototype builtins and each's multi in CORE.pmt, and read :unary

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/coretypes_test.go`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `TestCoreDerivedPrototypesArePerls` (`internal/parse/coretypes_test.go`, the harness); `perl -MPod::Functions -e 'print "@{$Kinds{ARRAY}} @{$Kinds{LIST}} @{$Kinds{HASH}}"'`; `internal/types/signatures.go:42-43` (the measured `keys` and `values` rows); the multi-candidate prototype rule the no-prototype builtins issue adds

Second batch, by perlfunc's categories "Functions for real @ARRAYs", "Functions for list data" and "Functions for real %HASHes" (Pod::Functions kinds `ARRAY`, `LIST`, `HASH`), intersected with CORE.pmt's prototyped lines, excluding `each`, which the no-prototype builtins issue declares as a multi, and `all` and `any`, whose `\&@` holds an aliased `&` and goes to the last batch with the other lines that need one: `keys`, `pop`, `push`, `shift`, `splice`, `unshift`, `values`, `join`, `unpack` (measured on 5.42; `push` to `splice` and `join` were typed with the measured rows). `keys` and `values` are `\[%@]`, so, like `each`, each is declared as two candidates, `(Hash \%h)` and `(Array \@a)`, whose derived prototype is perl's by the multi-candidate rule; copy their two signatures.go rows into the test as golden values (never compare against signatures.go, which a later issue deletes). Measured: `keys` in scalar context is a count (`my $n = keys %h` gives 2), while the row types it `List`; keep the measured row and note the scalar result in the line's comment -- no RFC section asks for a `:context` split here, so this issue does not add one. Every type is measured on perl 5.42 and recorded as a comment. Hold the batch's names in the test. Re-marks no RFC section.

## Steps

- [ ] Measure the batch's names; write failing test `TestCoreArrayListHashBuiltinsTyped`: every name in the batch has a typed CORE.pmt line, `keys` and `values` as two-candidate multis
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreArrayListHashBuiltinsTyped$' -count=1 -v`
- [ ] Type the lines with their measurement comments; run the harness and the test to verify pass; `make test`; commit
- [ ] Write failing test `TestCoreKeysValuesMatchMeasuredSignatures`: the golden `keys` and `values` rows equal their CORE.pmt candidates, and each derives `\[%@]`
- [ ] Implement any fix, measuring which side is right; run test to verify pass; `make test`; commit

### Negative Scenarios

- `keys` and `values` on a scalar (`keys $x`), which perl 5.42 forbids, match neither candidate (`go test ./internal/parse/ -run '^TestCoreKeysValuesRefuseScalar$' -count=1 -v | grep -q '^--- PASS: TestCoreKeysValuesRefuseScalar'`)
- `each` is not typed twice: this batch leaves its line to the multi issue, and the file has exactly one declaration set for `each` (`go test ./internal/parse/ -run '^TestCoreEachDeclaredOnce$' -count=1 -v | grep -q '^--- PASS: TestCoreEachDeclaredOnce'`)

## Acceptance criteria

- [ ] CORE.pmt types every array, list and hash builtin (`go test ./internal/parse/ -run '^TestCoreArrayListHashBuiltinsTyped$' -count=1 -v | grep -q '^--- PASS: TestCoreArrayListHashBuiltinsTyped'`)
- [ ] keys and values carry the measured signatures and derive perl prototype (`go test ./internal/parse/ -run '^TestCoreKeysValuesMatchMeasuredSignatures$' -count=1 -v | grep -q '^--- PASS: TestCoreKeysValuesMatchMeasuredSignatures'`)
- [ ] every CORE.pmt prototype, derived from its types where typed, is still perl's (`go test ./internal/parse/ -run '^TestCoreDerivedPrototypesArePerls$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedPrototypesArePerls'`)
- [ ] keys and values on a scalar match no candidate (`go test ./internal/parse/ -run '^TestCoreKeysValuesRefuseScalar$' -count=1 -v | grep -q '^--- PASS: TestCoreKeysValuesRefuseScalar'`)
- [ ] each is declared once (`go test ./internal/parse/ -run '^TestCoreEachDeclaredOnce$' -count=1 -v | grep -q '^--- PASS: TestCoreEachDeclaredOnce'`)
