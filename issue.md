---
title: Declare POSIX's constants, exports and prototypes
state: done
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
- 01a1464e-d980-7382-970e-067a5ea3db84
- 01a1464e-db1c-73eb-a3a9-e897d7a56b92
- 01a14683-8375-7810-8a75-b772680603c6
- 01a1464e-dadb-7796-a704-4007150a1bf7
blocks:
- 01a1464e-de34-7403-95d2-25f4855404b2
- 01a14683-84ab-7efb-a0c5-b2db8deb8e28
created: 2026-10-16T20:02:15.687306804Z
updated: 2026-10-20T00:43:19.19168408Z
sessions:
- start_sha: 5b91125f4014127e11963fc610c805ba205b10ff
  end_sha: 236f66e72899b4f1ef19ba3d8b8613b74ce46140
  commits: 3
  started_at: 2026-10-17T04:25:16.940428706Z
  ended_at: 2026-10-20T00:43:19.19168408Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-17T04:25:16.940428706Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-20T00:43:19.19168408Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
- internal/parse/declarations/POSIX.tpm
- internal/parse/heldtoperl_test.go
- internal/parse/testdata/held_to_perl.pl
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
- [ ] Pin the core-module ranking in a test and record the rule in the RFC
- [ ] Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle
- [ ] Convert every .tpm to :export and :export_ok, then refuse export lists in a .tpm (01a14683-8375-7810-8a75-b772680603c6)
- [ ] Read typed variables with :export and :export_ok in declaration files (01a1464e-dadb)

## Context

- paths: `internal/parse/declarations/POSIX.tpm` (new), `internal/parse/heldtoperl_test.go`
- commands: `go test ./internal/parse/ -run '^TestPOSIXPrototypes$|^TestPOSIXExports$' -count=1 -v`
- entrypoints: `internal/parse/declarations/Carp.tpm` (the model: measured facts in comments), the held-to-perl helpers from the Carp issue (01a1464e-db1c)

Modules (from the ranking): POSIX (1). POSIX is split in two: this issue writes `POSIX.tpm` with every name a `use POSIX` installs, its `:export`/`:export_ok` marks and tags, and its prototypes, and types the constants; the follow-on issue (01a14683-84ab-7efb-a0c5-b2db8deb8e28, Declare POSIX's functions and hold them to the return-type oracle) types the functions and runs the oracle.

Declare every name a `use` of POSIX installs in a caller, with its prototype from perl's `prototype()`. A `.tpm` has no `@EXPORT`, `@EXPORT_OK` or `%EXPORT_TAGS` (the parser refuses them): each exported name carries `:export` (a bare `use` installs it) or `:export_ok` (it must be asked for), with its tags as attribute arguments. What is marked is what perl's `use` installs, measured for a bare `use POSIX`, for `use POSIX qw(name)` and for `use POSIX ':tag'`, not POSIX's raw lists. Measured 2026-10-16 on 5.42.0: POSIX builds its lists in `import` (587 `@EXPORT`, 179 `@EXPORT_OK`, 35 tags after `import`, 0/0 after `require`), and a bare `use POSIX` installs 586 subs and the variable `%SIGRT`. 169 of the exported subs are not defined until called (AUTOLOAD), so the install check uses `exists &X::name`, not `defined`. `%SIGRT` is a typed variable, so this issue depends on the typed-variable reader (01a1464e-dadb). Constants (the names whose prototype is `()`) get their measured type here; a function's signature here may be untyped (`sub f;` with its prototype) until the follow-on issue types it. Each file starts with two `ABOUTME:` lines naming the perl (5.42.0) and module version measured.

A `.tpm` changes what the parser reads for every file that uses the module, so run the m1 gates after this batch, not only at the end: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor`, `TestGoParserWrongIsZero` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc). Run each with the command the final-gate issue (01a1464e-de34) gives it: `TestRatchetCorpus` skips without `-parseoracle.corpus`, and the two Go-parser gates skip unless `PARSEORACLE_SHIM` names a built shim, and a skip proves nothing. A ratchet may only be raised. A drop is fixed in the .tpm, or the batch stops and asks perigrin.

**Ruling 1 (perigrin, 2026-10-16): no export lists in a `.tpm`.** "We don't need the @EXPORT in the .tpm since we have :export and :export_ok." On names a module's own `import` installs: "we should figure out as best we can what is exported (loc, loc_lang, etc.) and mark them as :export or :export_ok accordingly."

## Steps

- [ ] Write failing tests `TestPOSIXPrototypes` and `TestPOSIXExports` in `internal/parse/heldtoperl_test.go`
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestPOSIXPrototypes$|^TestPOSIXExports$' -count=1 -v`
- [ ] Write `POSIX.tpm` (names, marks, tags, prototypes, constant types, `our ... %SIGRT :export;`); run to verify pass; commit
- [ ] Write the negative tests below; run to verify pass; commit
- [ ] Run the m1 gates above; `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] declared prototypes equal perl's prototype() (`go test ./internal/parse/ -run '^TestPOSIXPrototypes$' -count=1 -v | grep -q '^--- PASS: TestPOSIXPrototypes'`)
- [ ] declared `:export`/`:export_ok` marks equal what perl's `use POSIX` installs in a caller, bare, by name and by tag (`go test ./internal/parse/ -run '^TestPOSIXExports$' -count=1 -v | grep -q '^--- PASS: TestPOSIXExports'`)
- [ ] a bare `use POSIX;` imports `%SIGRT` with its declared type (measured: it is in the bare install) (`go test ./internal/parse/ -run '^TestPOSIXBareUseImportsSIGRT$' -count=1 -v | grep -q '^--- PASS: TestPOSIXBareUseImportsSIGRT'`)

### Negative Scenarios
- [ ] the POSIX export comparison is not vacuous: a bare `use POSIX` is measured installing 586 subs and `%SIGRT`, counting the AUTOLOAD names that `defined` would miss (`go test ./internal/parse/ -run '^TestPOSIXExportsAreNotVacuous$' -count=1 -v | grep -q '^--- PASS: TestPOSIXExportsAreNotVacuous'`)
- [ ] `use POSIX ();` imports nothing, so a bare `floor(1.5)` does not resolve to the declaration (`go test ./internal/parse/ -run '^TestPOSIXEmptyImportListImportsNothing$' -count=1 -v | grep -q '^--- PASS: TestPOSIXEmptyImportListImportsNothing'`)
- [ ] `printf STDOUT "x"` after `use POSIX;` keeps perl's filehandle parse (CORE printf has no prototype and keeps its own parse) (`go test ./internal/parse/ -run '^TestPOSIXImportKeepsPrintfFilehandleParse$' -count=1 -v | grep -q '^--- PASS: TestPOSIXImportKeepsPrintfFilehandleParse'`)
