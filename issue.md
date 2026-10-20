---
title: Declare POSIX's functions and hold them to the return-type oracle
state: in-progress
urgency: normal
milestone: core-module-types
blocked_by:
- 01a14683-8375-7810-8a75-b772680603c6
- 01a1464e-dc07-754e-9ef4-637df15d51a9
blocks:
- 01a1464e-de34-7403-95d2-25f4855404b2
created: 2026-10-16T20:59:46.730685644Z
updated: 2026-10-20T02:15:10.157420047Z
sessions:
- start_sha: 236f66e72899b4f1ef19ba3d8b8613b74ce46140
  end_sha: 49b2a9a42871b31c8f4d6ae70850d11f3b1d844c
  commits: 3
  started_at: 2026-10-20T00:43:36.52718261Z
  ended_at: 2026-10-20T01:48:32.556825795Z
- start_sha: 49b2a9a42871b31c8f4d6ae70850d11f3b1d844c
  end_sha: ""
  commits: 0
  started_at: 2026-10-20T02:15:10.157420047Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T00:43:36.52718261Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-20T01:48:32.556825795Z
- state: reopened
  actor: human:git-zhi
  timestamp: 2026-10-20T02:15:09.725949294Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T02:15:10.157420047Z
observed_paths:
- internal/infer/precision_test.go
- internal/infer/returntypes_test.go
- internal/infer/testdata/observe_types.pl
- internal/infer/testdata/oracle/Class-Struct.pl
- internal/infer/testdata/oracle/Errno.pl
- internal/infer/testdata/oracle/File-Basename.pl
- internal/infer/testdata/oracle/File-Copy.pl
- internal/infer/testdata/oracle/File-Find.pl
- internal/infer/testdata/oracle/File-Path.pl
- internal/infer/testdata/oracle/Getopt-Std.pl
- internal/infer/testdata/oracle/POSIX.pl
- internal/infer/testdata/oracle/Text-ParseWords.pl
- internal/infer/testdata/oracle/Time-Local.pl
- internal/parse/declarations/Class/Struct.tpm
- internal/parse/declarations/Errno.tpm
- internal/parse/declarations/File/Basename.tpm
- internal/parse/declarations/File/Copy.tpm
- internal/parse/declarations/File/Find.tpm
- internal/parse/declarations/File/Path.tpm
- internal/parse/declarations/Getopt/Std.tpm
- internal/parse/declarations/POSIX.tpm
- internal/parse/declarations/Text/ParseWords.tpm
- internal/parse/declarations/Time/Local.tpm
- internal/parse/heldtoperl_test.go
- internal/parse/testdata/held_to_perl.pl
- internal/psc/testdata/lib_true_positives.txt
---

## Prerequisites

- [ ] Declare POSIX's constants, exports and prototypes (01a1464e-dc07)

## Context

- paths: `internal/parse/declarations/POSIX.tpm`, `internal/infer/returntypes_test.go`, `internal/infer/testdata/oracle/POSIX.pl` (new)
- commands: `go test ./internal/infer/ -run '^TestPOSIXReturnTypes$' -count=1 -v`
- entrypoints: the return-type oracle from the Carp issue (01a1464e-db1c), `POSIX.tpm` as 01a1464e-dc07 left it

The second half of POSIX: give every POSIX function its typed signature and return type, and hold them to perl with the return-type oracle. Call cases go in `internal/infer/testdata/oracle/POSIX.pl`; every sub the oracle cannot call is in the exemption map in `returntypes_test.go`, by name, with its reason. `POSIX::_exit` is declared `None` and is never called.

**Ruling 5 (perigrin, 2026-10-16):** "POSIX's 169 AUTOLOAD stubs are declared `None`. perl dies when they're called, so the oracle confirms each dies."


A `.tpm` changes what the parser reads for every file that uses the module, so run the m1 gates after this batch, not only at the end: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor`, `TestGoParserWrongIsZero` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc). Run each with the command the final-gate issue (01a1464e-de34) gives it: `TestRatchetCorpus` skips without `-parseoracle.corpus`, and the two Go-parser gates skip unless `PARSEORACLE_SHIM` names a built shim, and a skip proves nothing. A ratchet may only be raised. A drop is fixed in the .tpm, or the batch stops and asks perigrin.


**Resolved 2026-10-16 (by measurement; perl is the judge):** of the 169 AUTOLOAD names, the 107 that die when called (80 in `%replacement`, "Unimplemented"; 27 via `constant()`) are declared `None`, as perigrin ruled. The 62 in `%reimpl` return values (`POSIX::getpid()` is `$$`, `POSIX::abs(-1)` is 1), so ruling 5 does not apply to them: each is declared with the return type the oracle measures. No stop is needed for them.

## Steps

- [ ] Ask perigrin about the 62 `%reimpl` names (above); record the answer in this issue
- [ ] Write failing test `TestPOSIXReturnTypes` in `internal/infer/returntypes_test.go`, with call cases in `internal/infer/testdata/oracle/POSIX.pl` and every uncallable sub listed by name with its reason
- [ ] Run to verify failure: `go test ./internal/infer/ -run '^TestPOSIXReturnTypes$' -count=1 -v`
- [ ] Type the functions in `POSIX.tpm`; fix declared return types where perl disagrees; run to verify pass; commit
- [ ] Write `TestPOSIXAutoloadStubsDie`; run to verify pass; commit
- [ ] Run the m1 gates above; `make test`; commit

### Review Findings

Gate, 2026-10-20, on 49b2a9a4 (fresh Sonnet reviewer; criteria pass and TestPOSIXReturnTypes proven non-vacuous; 80/62/27 recomputed from POSIX 2.23 and correct; Sys/Syslog.pm:562 confirmed a true positive; ten spot-checked returns correct; :listop parses as perl does). VERDICT: REOPEN.

1. (CODE, 90) POSIX.tpm:761, 814 and returntypes_test.go:53, 57: `signal` and `tcsetattr` are declared with return types and exempted ("installs a signal handler", "mutates terminal attributes"), but on perl 5.42 both always die "is not a valid POSIX macro" (`POSIX::signal(SIGUSR1,"IGNORE")`, `POSIX::tcsetattr(0,0,undef)`). They belong with the undefined-macro names: declare None, drop the exemptions, add them to posixUndefinedMacroNames.
2. (CODE, 80) POSIX.tpm:1-6, returntypes_test.go:71-84: L_tmpnam, fileno, rand and srand (4 of the 80 %replacement names) are undeclared, and posixReplacementNames holds 76. The reviewer measured that `use POSIX qw(fileno)` dies at BEGIN but a bare `use POSIX;` installs all four as AUTOLOAD stubs that die "Unimplemented" when called. That contradicts dc07's comment and its passing TestPOSIXExports: measure with perl which is true, then declare all four None with the mark perl's bare use implies, and have TestPOSIXAutoloadStubsDie cover all 80.
3. (CODE, 75) POSIX.tpm:764 `sigprocmask` (exempt, so typed from POSIX.pod) is `Int`; POSIX.pod says "Returns undef on failure" and success returns "0 but true". Declare from the documentation, at least `Int|Undef` (or Str|Undef if "0 but true" is not an Int in this lattice: measure).

## Acceptance Criteria

### Positive Scenarios
- [ ] declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestPOSIXReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestPOSIXReturnTypes'`)
- [ ] every POSIX AUTOLOAD name declared `None` dies when the oracle calls it, and the 80 `%replacement` and 27 undefined-macro names are all declared `None` (`go test ./internal/infer/ -run '^TestPOSIXAutoloadStubsDie$' -count=1 -v | grep -q '^--- PASS: TestPOSIXAutoloadStubsDie'`)

### Negative Scenarios
- [ ] POSIX `_exit`, `abort` and `pause` are in the exemption map with a reason and are never called (`go test ./internal/infer/ -run '^TestPOSIXOracleNeverCallsExit$' -count=1 -v | grep -q '^--- PASS: TestPOSIXOracleNeverCallsExit'`)
- [ ] `POSIX::strtol` returns two values in list context and one in scalar (measured), and both fit the declaration (`go test ./internal/infer/ -run '^TestPOSIXStrtolContextReturnIsDeclared$' -count=1 -v | grep -q '^--- PASS: TestPOSIXStrtolContextReturnIsDeclared'`)
- [ ] a `%reimpl` name that returns, such as `getpid`, is not declared `None` (`go test ./internal/infer/ -run '^TestPOSIXReimplementedSubsAreNotNone$' -count=1 -v | grep -q '^--- PASS: TestPOSIXReimplementedSubsAreNotNone'`)
