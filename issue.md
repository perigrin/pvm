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
updated: 2026-10-20T00:43:36.52718261Z
sessions:
- start_sha: 236f66e72899b4f1ef19ba3d8b8613b74ce46140
  end_sha: ""
  commits: 0
  started_at: 2026-10-20T00:43:36.52718261Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T00:43:36.52718261Z
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

## Acceptance Criteria

### Positive Scenarios
- [ ] declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestPOSIXReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestPOSIXReturnTypes'`)
- [ ] every POSIX AUTOLOAD name declared `None` dies when the oracle calls it, and the 80 `%replacement` and 27 undefined-macro names are all declared `None` (`go test ./internal/infer/ -run '^TestPOSIXAutoloadStubsDie$' -count=1 -v | grep -q '^--- PASS: TestPOSIXAutoloadStubsDie'`)

### Negative Scenarios
- [ ] POSIX `_exit`, `abort` and `pause` are in the exemption map with a reason and are never called (`go test ./internal/infer/ -run '^TestPOSIXOracleNeverCallsExit$' -count=1 -v | grep -q '^--- PASS: TestPOSIXOracleNeverCallsExit'`)
- [ ] `POSIX::strtol` returns two values in list context and one in scalar (measured), and both fit the declaration (`go test ./internal/infer/ -run '^TestPOSIXStrtolContextReturnIsDeclared$' -count=1 -v | grep -q '^--- PASS: TestPOSIXStrtolContextReturnIsDeclared'`)
- [ ] a `%reimpl` name that returns, such as `getpid`, is not declared `None` (`go test ./internal/infer/ -run '^TestPOSIXReimplementedSubsAreNotNone$' -count=1 -v | grep -q '^--- PASS: TestPOSIXReimplementedSubsAreNotNone'`)
