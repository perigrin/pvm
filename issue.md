---
title: Type CORE.pmt's remaining builtins, with localtime and gmtime as :context multis
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4ca3-7d55-82ac-ba125816ee71
- 01a0fd9b-4f8d-71ca-bac4-24eceb1dc4a4
created: 2026-10-03T11:12:38.999799272Z
updated: 2026-10-03T11:16:01.541400333Z
---

## Prerequisites

- [ ] Check CORE.pmt's derived prototypes against perl, with signatures.go's measured rows
- [ ] Type CORE.pmt's string, numeric and regex builtins
- [ ] Type CORE.pmt's array, list and hash builtins
- [ ] Read :context(...) and select a multi candidate by calling context

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/coretypes_test.go`, `docs/rfcs/0001-type-libraries.md`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `TestCoreDerivedPrototypesArePerls` (`internal/parse/coretypes_test.go`, the harness); `perl -MPod::Functions -e 'print join " ", sort keys %Kinds'` for the remaining categories; RFC 0001 "`:context(...)`" for `localtime`

Last batch: every prototyped CORE.pmt line the earlier batches did not type -- perlfunc's "Input and output functions", "Functions for fixed-length data or records", "Functions for filehandles, files, or directories", "Keywords related to the control flow", processes, modules, classes, sockets, System V IPC, user and network information, time, and the rest (Pod::Functions kinds `I/O`, `Binary`, `File`, `Flow`, `Process`, `Modules`, `Objects`, `Socket`, `SysV`, `User`, `Network`, `Time`, `Namespace`, `Misc`), plus `pos`, `all` and `any`: 149 lines measured on 5.42, most with `()`, `$` or `;$` prototypes, some already typed with the measured rows (`bless`, `die`, `warn`, `ref`, `scalar`). `localtime` and `gmtime` are `:context` multis, as the RFC's ":context(...)" section declares them: `multi sub localtime :prototype(;$) :context($) (Int $time = time) Str;` and `multi sub localtime :prototype(;$) :context(@) (Int $time = time) List[Int];` (measured: `my $t = localtime(0)` is a date string, the list's first element `0`), and the same for `gmtime`. Every type is measured and recorded as a comment. Open point: eight lines alias a code or glob container -- `all` and `any` (`\&@`), `lock` and `undef` (`\[$@%&*]`), `tie`, `tied` and `untie` (`\[$@%*]`), `pos` (`;\[$*]`) -- and the RFC spells aliasing only for `\$`, `\@`, `\%` and `\(@list)`; `Code \&c` and `Glob \*g`, typed by the multi-candidate rule for the `\[...]` forms, are the natural extension but need perigrin's ruling before those lines are typed. This finishes the 188, so it completes RFC "A typed signature and a prototype say the same thing" (the derivation, the aliased rows, the final-`List` rule and the multi-candidate rule came first) and re-marks it. If this overruns one session, split by category (I/O and files; processes, sockets and IPC; the rest) before starting.

## Steps

- [ ] Write failing test `TestCoreLocaltimeGmtimeAreContextMultis`: `localtime` and `gmtime` each have a `:context($)` candidate returning `Str` and a `:context(@)` candidate returning `List[Int]`, both deriving `;$`
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreLocaltimeGmtimeAreContextMultis$' -count=1 -v`
- [ ] Add the declarations with their measurement comments; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestCoreEveryPrototypedLineIsTyped`: every one of the 188 prototyped lines carries a typed signature, and the file reads with no declaration errors (unknown names, bare slurpies)
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreEveryPrototypedLineIsTyped$' -count=1 -v`
- [ ] Type the remaining lines, category by category, with measurement comments; run the harness after each category; `make test`; commit per category
- [ ] Re-mark the RFC heading "A typed signature and a prototype say the same thing" `(*Implemented*, <sha>)`; commit

### Negative Scenarios

- no CORE.pmt line is left half-done: every one of the 188 prototyped lines carries a typed signature, and the file reads with no declaration errors (`go test ./internal/parse/ -run '^TestCoreEveryPrototypedLineIsTyped$' -count=1 -v | grep -q '^--- PASS: TestCoreEveryPrototypedLineIsTyped'`)
- a call to `localtime` or `gmtime` in void context has no type, since neither candidate answers for void (`go test ./internal/parse/ -run '^TestCoreLocaltimeVoidHasNoType$' -count=1 -v | grep -q '^--- PASS: TestCoreLocaltimeVoidHasNoType'`)
- a niladic builtin (`time`, `wait`, `fork`: prototype `()`) is typed with zero parameters, not with an empty list parameter, so its derived prototype stays `()` and not absent (`go test ./internal/parse/ -run '^TestCoreNiladicBuiltinsTypedEmpty$' -count=1 -v | grep -q '^--- PASS: TestCoreNiladicBuiltinsTypedEmpty'`)

## Acceptance criteria

- [ ] localtime and gmtime are declared as context multis (`go test ./internal/parse/ -run '^TestCoreLocaltimeGmtimeAreContextMultis$' -count=1 -v | grep -q '^--- PASS: TestCoreLocaltimeGmtimeAreContextMultis'`)
- [ ] every prototyped CORE.pmt line is typed and the file has no declaration errors (`go test ./internal/parse/ -run '^TestCoreEveryPrototypedLineIsTyped$' -count=1 -v | grep -q '^--- PASS: TestCoreEveryPrototypedLineIsTyped'`)
- [ ] every prototype derived from CORE.pmt types is perl's (`go test ./internal/parse/ -run '^TestCoreDerivedPrototypesArePerls$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedPrototypesArePerls'`)
- [ ] the RFC section on typed signatures and prototypes is implemented (`grep -q '^### A typed signature and a prototype.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
- [ ] a void call to localtime has no type (`go test ./internal/parse/ -run '^TestCoreLocaltimeVoidHasNoType$' -count=1 -v | grep -q '^--- PASS: TestCoreLocaltimeVoidHasNoType'`)
- [ ] niladic builtins derive an empty prototype, not an absent one (`go test ./internal/parse/ -run '^TestCoreNiladicBuiltinsTypedEmpty$' -count=1 -v | grep -q '^--- PASS: TestCoreNiladicBuiltinsTypedEmpty'`)
