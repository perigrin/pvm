---
title: Declare grep, map and sort as multis in CORE.pmt, with sort's invocant slot
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-49b2-7ea9-bfa1-ba398baadc3f
- 01a0fd9b-4ca3-7d55-82ac-ba125816ee71
- 01a0fd9b-4e65-7397-836f-140fdc7feb53
- 01a0fd9b-5017-78d0-ae38-331cbe6a9866
blocks:
- 01a0fd9b-55cb-70cd-a9da-929190cb581c
created: 2026-10-02T17:13:29.757085237Z
updated: 2026-10-02T17:24:13.584145047Z
---

## Prerequisites

- [ ] Parse the invocant colon in .pmt signatures and declare print, printf, say, exec and system with it
- [ ] Read :context(...) and select a multi candidate by calling context
- [ ] Choose the map and grep variant by the comma after the closing brace
- [ ] Declare the plain no-prototype builtins in CORE.pmt and read :unary

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/coretypes_test.go`, `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: RFC 0001 "Builtins that keep their own parse" (the six multi lines and `multi sub sort (Code|Str $by: List @list) List;`), `parseListOpBlock` (`internal/parse/call.go:492`)

RFC 0001 "Builtins that keep their own parse": perlsub reimplements `grep` as `sub mygrep (&@)`, and the expression form deparses as a `($@)` sub's would, so grep and map are each a `Code &block, List @list` multi and a `Scalar $expr, List @list` multi; the comma after the brace (previous issue) is what selects. `sort` is `(Code &block, List @list)`, `(List @list)`, and `(Code|Str $by: List @list)` for `sort byname @x` and `sort $n @x`; with a comma the string is data (`sort "byname", @x`). Scalar sort is undefined, so sort's candidates carry `:context(@)`. perl reports `undef` for all three prototypes, so none enters `coreTable()`. This finishes both "Builtins with no prototype" and "Builtins that keep their own parse". Open questions 9 and 10 (per-element evaluation, `defined &f`) stay open.

## Steps

- [ ] Measure each form with perl, then write failing test `TestCoreGrepMapSortAreMultis`: grep and map have the block and expression candidates; sort has the block, plain and invocant candidates, all `:context(@)`; none derives a prototype
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreGrepMapSortAreMultis$' -count=1 -v`
- [ ] Add the declarations with their measurement comments; run test to verify pass; `make test`; commit
- [ ] Re-mark the RFC headings "Builtins with no prototype" and "Builtins that keep their own parse" `(*Implemented*, <sha>)`; commit

### Negative Scenarios

- scalar `sort` has no type: every sort candidate carries `:context(@)`, so a scalar call selects none (measured: `my $n = sort 3,1,2` is undef) (`go test ./internal/parse/ -run '^TestCoreSortScalarHasNoType$' -count=1 -v | grep -q '^--- PASS: TestCoreSortScalarHasNoType'`)
- grep, map and sort stay out of `coreTable()` and `TestCoreDeclarationsArePerls` stays green (`go test ./internal/parse/ -run '^TestCoreGrepMapSortOutOfCoreTable$' -count=1 -v | grep -q '^--- PASS: TestCoreGrepMapSortOutOfCoreTable'`)

## Acceptance criteria

- [ ] CORE.pmt declares grep, map and sort as multis (`go test ./internal/parse/ -run '^TestCoreGrepMapSortAreMultis$' -count=1 -v | grep -q '^--- PASS: TestCoreGrepMapSortAreMultis'`)
- [ ] the RFC section on builtins with no prototype is implemented (`grep -q '^### Builtins with no prototype.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] the RFC section on builtins that keep their own parse is implemented (`grep -q '^### Builtins that keep their own parse.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] a scalar sort selects no candidate (`go test ./internal/parse/ -run '^TestCoreSortScalarHasNoType$' -count=1 -v | grep -q '^--- PASS: TestCoreSortScalarHasNoType'`)
- [ ] grep, map and sort stay out of the prototype table (`go test ./internal/parse/ -run '^TestCoreGrepMapSortOutOfCoreTable$' -count=1 -v | grep -q '^--- PASS: TestCoreGrepMapSortOutOfCoreTable'`)
