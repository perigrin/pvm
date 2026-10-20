---
title: "Declare the XS library modules: Encode, Time::HiRes, Compress::Raw::Zlib, SDBM_File"
state: reopened
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
- 01a1464e-d980-7382-970e-067a5ea3db84
- 01a1464e-db1c-73eb-a3a9-e897d7a56b92
- 01a14683-8375-7810-8a75-b772680603c6
blocks:
- 01a1464e-de34-7403-95d2-25f4855404b2
created: 2026-10-16T20:02:15.635645821Z
updated: 2026-10-20T03:00:12.394504062Z
sessions:
- start_sha: 5b91125f4014127e11963fc610c805ba205b10ff
  end_sha: 3b04b0218d2848955d5c7b1a6ccb330cf59a1fd1
  commits: 9
  started_at: 2026-10-17T04:25:19.319037318Z
  ended_at: 2026-10-20T02:42:52.815355718Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-17T04:25:19.319037318Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-20T02:42:52.815355718Z
- state: reopened
  actor: human:git-zhi
  timestamp: 2026-10-20T03:00:12.394504062Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
- internal/cli/symlinks_test.go
- internal/infer/precision_test.go
- internal/infer/returntypes_test.go
- internal/infer/testdata/observe_types.pl
- internal/infer/testdata/oracle/Class-Struct.pl
- internal/infer/testdata/oracle/Compress-Raw-Zlib.pl
- internal/infer/testdata/oracle/Encode.pl
- internal/infer/testdata/oracle/Errno.pl
- internal/infer/testdata/oracle/Fcntl.pl
- internal/infer/testdata/oracle/File-Basename.pl
- internal/infer/testdata/oracle/File-Copy.pl
- internal/infer/testdata/oracle/File-Find.pl
- internal/infer/testdata/oracle/File-Path.pl
- internal/infer/testdata/oracle/Getopt-Std.pl
- internal/infer/testdata/oracle/POSIX.pl
- internal/infer/testdata/oracle/SDBM_File.pl
- internal/infer/testdata/oracle/Socket.pl
- internal/infer/testdata/oracle/Text-ParseWords.pl
- internal/infer/testdata/oracle/Time-HiRes.pl
- internal/infer/testdata/oracle/Time-Local.pl
- internal/infer/testdata/return_oracle.pl
- internal/parse/declarations/Class/Struct.tpm
- internal/parse/declarations/Compress/Raw/Zlib.tpm
- internal/parse/declarations/Encode.tpm
- internal/parse/declarations/Errno.tpm
- internal/parse/declarations/Fcntl.tpm
- internal/parse/declarations/File/Basename.tpm
- internal/parse/declarations/File/Copy.tpm
- internal/parse/declarations/File/Find.tpm
- internal/parse/declarations/File/Path.tpm
- internal/parse/declarations/Getopt/Std.tpm
- internal/parse/declarations/POSIX.tpm
- internal/parse/declarations/SDBM_File.tpm
- internal/parse/declarations/Socket.tpm
- internal/parse/declarations/Text/ParseWords.tpm
- internal/parse/declarations/Time/HiRes.tpm
- internal/parse/declarations/Time/Local.tpm
- internal/parse/heldtoperl_test.go
- internal/parse/testdata/held_to_perl.pl
- internal/psc/testdata/lib_true_positives.txt
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
- [ ] Convert every .tpm to :export and :export_ok, then refuse export lists in a .tpm (01a14683-8375-7810-8a75-b772680603c6)
- [ ] Pin the core-module ranking in a test and record the rule in the RFC
- [ ] Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle

## Context

- paths: `internal/parse/declarations/` (one `.tpm` per module below), `internal/parse/heldtoperl_test.go`, `internal/infer/returntypes_test.go`, `internal/infer/testdata/oracle/`
- commands: `go test ./internal/parse/ -run '^TestXSLibrariesPrototypes$|^TestXSLibrariesExports$' -count=1 -v`, `go test ./internal/infer/ -run '^TestXSLibrariesReturnTypes$' -count=1 -v`
- entrypoints: `internal/parse/declarations/Carp.tpm` (the model: measured facts in comments), the held-to-perl helpers from the Carp issue (01a1464e-db1c)

Modules (from the ranking): Encode, Time::HiRes, Compress::Raw::Zlib, SDBM_File (4).

Write one `.tpm` per module declaring every name a `use` of it installs in a caller, with prototype and typed signature, plus the subs and methods its callers use; a declaration is read in place of the module's source, so a name it leaves out is a name the import loses. A `.tpm` has no `@EXPORT`, `@EXPORT_OK` or `%EXPORT_TAGS` (the parser refuses them): each exported sub or variable carries `:export` (a bare `use` installs it) or `:export_ok` (it must be asked for), with its tags as attribute arguments, e.g. `sub LOCK_EX :export_ok(flock) () Int;`. What is marked is what perl's `use` installs in a caller, measured for a bare `use M`, for `use M qw(name)` and for `use M ':tag'`, not the module's raw lists. A module with its own `import` is declared the same way: mark what that `import` installs. Each file starts with two `ABOUTME:` lines naming the perl (5.42.0) and module version measured. A sub that never returns is declared `None`. XS: prototypes from perl's `prototype()`. Compress::Raw::Zlib exports 47 constants; SDBM_File is a tie class whose `@EXPORT_OK` holds constants.

A `.tpm` changes what the parser reads for every file that uses the module, so run the m1 gates after this batch, not only at the end: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor`, `TestGoParserWrongIsZero` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc). Run each with the command the final-gate issue (01a1464e-de34) gives it: `TestRatchetCorpus` skips without `-parseoracle.corpus`, and the two Go-parser gates skip unless `PARSEORACLE_SHIM` names a built shim, and a skip proves nothing. A ratchet may only be raised. A drop is fixed in the .tpm, or the batch stops and asks perigrin.

**Ruling 1 (perigrin, 2026-10-16): no export lists in a `.tpm`.** "We don't need the @EXPORT in the .tpm since we have :export and :export_ok." On names a module's own `import` installs: "we should figure out as best we can what is exported (loc, loc_lang, etc.) and mark them as :export or :export_ok accordingly."

## Steps

- [ ] Write failing tests `TestXSLibrariesPrototypes` and `TestXSLibrariesExports` in `internal/parse/heldtoperl_test.go` over the module list above
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestXSLibrariesPrototypes$|^TestXSLibrariesExports$' -count=1 -v`
- [ ] Write the `.tpm` files; run to verify pass; commit
- [ ] Write failing test `TestXSLibrariesReturnTypes` in `internal/infer/returntypes_test.go`, with call cases in `internal/infer/testdata/oracle/` and every uncallable sub listed by name with its reason
- [ ] Run to verify failure: `go test ./internal/infer/ -run '^TestXSLibrariesReturnTypes$' -count=1 -v`
- [ ] Fix declared return types where perl disagrees; run to verify pass; commit
- [ ] Run the m1 gates above; `make test`; commit

### Review Findings

Gate, 2026-10-20, on 3b04b021 (fresh Sonnet reviewer; criteria pass and proven non-vacuous by mutation; ~18 subs checked against perl 5.42, no tempering beyond the tracked Encode::encode). VERDICT: REOPEN.

A. (CODE, 70) Encode.tpm:34, 35, 42, 43, 60, 79: CHECK is declared `Int|CodeRef $check = 0` on bytes2str, str2bytes, decode, decode_utf8, encode and from_to, but perl 5.42 accepts a literal undef and treats it as 0 (`Encode::encode("ascii", $s, undef)` returns "caf?"). Declare `Int|CodeRef|Undef`, measured for each sub.
B. (TEST-ONLY, 85; fixed in the same pass) held_to_perl.pl prototypes mode: the new `if (!exists ... && !$module->can($name)) { use $module qw($name) }` block is subsumed by the existing per-name retry below it; removing it changes no test across internal/parse. Delete it and its comment.

## Acceptance Criteria

### Positive Scenarios
- [ ] declared prototypes equal perl's prototype() (`go test ./internal/parse/ -run '^TestXSLibrariesPrototypes$' -count=1 -v | grep -q '^--- PASS: TestXSLibrariesPrototypes'`)
- [ ] declared `:export`/`:export_ok` marks equal what perl's `use` installs in a caller, bare, by name and by tag (`go test ./internal/parse/ -run '^TestXSLibrariesExports$' -count=1 -v | grep -q '^--- PASS: TestXSLibrariesExports'`)
- [ ] declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestXSLibrariesReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestXSLibrariesReturnTypes'`)

### Negative Scenarios
- [ ] a bare `use Time::HiRes;` imports nothing (its `@EXPORT` is empty, measured), so `time()` stays CORE's (`go test ./internal/parse/ -run '^TestTimeHiResBareUseKeepsCoreTime$' -count=1 -v | grep -q '^--- PASS: TestTimeHiResBareUseKeepsCoreTime'`)
- [ ] Encode `encode_utf8(undef)` returns a value its declared return type admits (`go test ./internal/infer/ -run '^TestEncodeUndefInputReturnIsDeclared$' -count=1 -v | grep -q '^--- PASS: TestEncodeUndefInputReturnIsDeclared'`)
- [ ] the oracle ties SDBM_File only under the test's temp dir (`go test ./internal/infer/ -run '^TestSDBMFileOracleStaysInScratch$' -count=1 -v | grep -q '^--- PASS: TestSDBMFileOracleStaysInScratch'`)
