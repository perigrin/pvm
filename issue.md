---
title: Type CORE.pmt's process, time, object and miscellaneous builtins, with localtime and gmtime as :context multis
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-05T14:18:57.039335349Z
updated: 2026-10-05T14:18:57.039335349Z
---

## Prerequisites

- [ ] Type CORE.pmt's I/O, file and directory builtins
- [ ] Read :context(...) and select a multi candidate by calling context

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/coretypes_test.go`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `TestCoreDerivedPrototypesArePerls` (`internal/parse/coretypes_test.go`, the harness); `perl -MPod::Functions -e 'print "@{$Kinds{Process}} @{$Kinds{Time}} @{$Kinds{Modules}} @{$Kinds{Objects}} @{$Kinds{Namespace}} @{$Kinds{Misc}}"'`; RFC 0001 "`:context(...)`" for `localtime`

Fourth batch, by perlfunc's categories "Functions for processes and process groups", "Time-related functions", "Keywords related to Perl modules", "Keywords related to classes and object-orientation", "Keywords related to scoping" and "Miscellaneous functions" (Pod::Functions kinds `Process`, `Time`, `Modules`, `Objects`, `Namespace`, `Misc`), plus the three prototyped keywords Pod::Functions lists under no category (`catch`, `isa`, `not`), intersected with CORE.pmt's prototyped lines and without names earlier batches hold. A name perlfunc lists under several categories belongs to the first of the three remaining batches (I/O and files; processes, time and the rest; system and flow) that holds one of its categories. Measured on 5.42 that is 34 lines: `alarm`, `fork`, `getpgrp`, `getppid`, `getpriority`, `kill`, `pipe`, `readpipe`, `setpgrp`, `setpriority`, `sleep`, `wait`, `waitpid`, `times`; `gmtime`, `localtime`, `time`; `__CLASS__`, `bless`, `ref`, `tie`, `tied`, `untie`, `method`; `caller`; `formline`, `lock`, `prototype`, `reset`, `scalar`, `undef`; `catch`, `isa`, `not` (`bless`, `ref` and `scalar` were typed with the measured rows). `localtime` and `gmtime` are `:context` multis, as the RFC's ":context(...)" section declares them: `multi sub localtime :prototype(;$) :context($) (Int $time = time) Str;` and `multi sub localtime :prototype(;$) :context(@) (Int $time = time) List[Int];` (measured: `my $t = localtime(0)` is a date string, the list's first element `0`), and the same for `gmtime`. Five of the batch's lines hold a glob slot -- `lock` (`\[$@%&*]`), `undef` (`;\[$@%&*]`), `tie` (`\[$@%*]$@`), `tied` and `untie` (`\[$@%*]`) -- which has no spelling yet (RFC open question 11, perigrin 2026-10-03): they stay prototype-only lines, untyped, and are left out of this batch's names; the last batch's every-line check exempts them with `pos`. Every type is measured and recorded as a comment. Hold the batch's names in the test. Re-marks no RFC section.

## Steps

- [ ] Write failing test `TestCoreLocaltimeGmtimeAreContextMultis`: `localtime` and `gmtime` each have a `:context($)` candidate returning `Str` and a `:context(@)` candidate returning `List[Int]`, both deriving `;$`
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreLocaltimeGmtimeAreContextMultis$' -count=1 -v`
- [ ] Add the declarations with their measurement comments; run test to verify pass; `make test`; commit
- [ ] Measure the batch's names with Pod::Functions and perl; write failing test `TestCoreProcessTimeMiscBuiltinsTyped`: every name in the batch, less the five glob-slot lines, has a typed CORE.pmt line, and those five are still prototype-only
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreProcessTimeMiscBuiltinsTyped$' -count=1 -v`
- [ ] Type the lines, category by category, with measurement comments; run the harness after each category; `make test`; commit per category

### Negative Scenarios

- a call to `localtime` or `gmtime` in void context has no type, since neither candidate answers for void (`go test ./internal/parse/ -run '^TestCoreLocaltimeVoidHasNoType$' -count=1 -v | grep -q '^--- PASS: TestCoreLocaltimeVoidHasNoType'`)
- a niladic builtin (`time`, `wait`, `fork`: prototype `()`) is typed with zero parameters, not with an empty list parameter, so its derived prototype stays `()` and not absent (`go test ./internal/parse/ -run '^TestCoreNiladicBuiltinsTypedEmpty$' -count=1 -v | grep -q '^--- PASS: TestCoreNiladicBuiltinsTypedEmpty'`)
- the glob-slot lines are not typed by a guess: `lock`, `undef`, `tie`, `tied` and `untie` keep their prototype-only lines, with no typed signature (`go test ./internal/parse/ -run '^TestCoreGlobSlotLinesStayUntyped$' -count=1 -v | grep -q '^--- PASS: TestCoreGlobSlotLinesStayUntyped'`)

## Acceptance criteria

- [ ] localtime and gmtime are declared as context multis (`go test ./internal/parse/ -run '^TestCoreLocaltimeGmtimeAreContextMultis$' -count=1 -v | grep -q '^--- PASS: TestCoreLocaltimeGmtimeAreContextMultis'`)
- [ ] CORE.pmt types every process, time, module, object, scoping and miscellaneous builtin but the glob-slot lines (`go test ./internal/parse/ -run '^TestCoreProcessTimeMiscBuiltinsTyped$' -count=1 -v | grep -q '^--- PASS: TestCoreProcessTimeMiscBuiltinsTyped'`)
- [ ] every CORE.pmt prototype, derived from its types where typed, is still perl's (`go test ./internal/parse/ -run '^TestCoreDerivedPrototypesArePerls$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedPrototypesArePerls'`)
- [ ] a void call to localtime has no type (`go test ./internal/parse/ -run '^TestCoreLocaltimeVoidHasNoType$' -count=1 -v | grep -q '^--- PASS: TestCoreLocaltimeVoidHasNoType'`)
- [ ] niladic builtins derive an empty prototype, not an absent one (`go test ./internal/parse/ -run '^TestCoreNiladicBuiltinsTypedEmpty$' -count=1 -v | grep -q '^--- PASS: TestCoreNiladicBuiltinsTypedEmpty'`)
- [ ] the glob-slot lines stay untyped (`go test ./internal/parse/ -run '^TestCoreGlobSlotLinesStayUntyped$' -count=1 -v | grep -q '^--- PASS: TestCoreGlobSlotLinesStayUntyped'`)
