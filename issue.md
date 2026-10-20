---
title: Declare the XS constant exporters Fcntl and Socket
state: done
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-dadb-7796-a704-4007150a1bf7
- 01a1464e-d912-7419-8b20-a8e24029b3ec
- 01a1464e-d980-7382-970e-067a5ea3db84
- 01a1464e-db1c-73eb-a3a9-e897d7a56b92
- 01a14683-8375-7810-8a75-b772680603c6
blocks:
- 01a1464e-de34-7403-95d2-25f4855404b2
created: 2026-10-16T20:02:15.752726002Z
updated: 2026-10-20T02:39:32.494021128Z
sessions:
- start_sha: 5b91125f4014127e11963fc610c805ba205b10ff
  end_sha: f8a82e032cf9fa043deec9d5fb866d21bac80c75
  commits: 8
  started_at: 2026-10-17T04:25:19.893120122Z
  ended_at: 2026-10-20T02:28:27.986790751Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-17T04:25:19.893120122Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-20T02:28:27.986790751Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
- internal/cli/symlinks_test.go
- internal/infer/precision_test.go
- internal/infer/returntypes_test.go
- internal/infer/testdata/observe_types.pl
- internal/infer/testdata/oracle/Class-Struct.pl
- internal/infer/testdata/oracle/Errno.pl
- internal/infer/testdata/oracle/Fcntl.pl
- internal/infer/testdata/oracle/File-Basename.pl
- internal/infer/testdata/oracle/File-Copy.pl
- internal/infer/testdata/oracle/File-Find.pl
- internal/infer/testdata/oracle/File-Path.pl
- internal/infer/testdata/oracle/Getopt-Std.pl
- internal/infer/testdata/oracle/POSIX.pl
- internal/infer/testdata/oracle/Socket.pl
- internal/infer/testdata/oracle/Text-ParseWords.pl
- internal/infer/testdata/oracle/Time-Local.pl
- internal/infer/testdata/return_oracle.pl
- internal/parse/declarations/Class/Struct.tpm
- internal/parse/declarations/Errno.tpm
- internal/parse/declarations/Fcntl.tpm
- internal/parse/declarations/File/Basename.tpm
- internal/parse/declarations/File/Copy.tpm
- internal/parse/declarations/File/Find.tpm
- internal/parse/declarations/File/Path.tpm
- internal/parse/declarations/Getopt/Std.tpm
- internal/parse/declarations/POSIX.tpm
- internal/parse/declarations/Socket.tpm
- internal/parse/declarations/Text/ParseWords.tpm
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
- [ ] Read typed variables with :export and :export_ok in declaration files

## Context

- paths: `internal/parse/declarations/` (one `.tpm` per module below), `internal/parse/heldtoperl_test.go`, `internal/infer/returntypes_test.go`, `internal/infer/testdata/oracle/`
- commands: `go test ./internal/parse/ -run '^TestConstantExportersPrototypes$|^TestConstantExportersExports$' -count=1 -v`, `go test ./internal/infer/ -run '^TestConstantExportersReturnTypes$' -count=1 -v`
- entrypoints: `internal/parse/declarations/Carp.tpm` (the model: measured facts in comments), the held-to-perl helpers from the Carp issue (01a1464e-db1c)

Modules (from the ranking): Fcntl, Socket (2).

Write one `.tpm` per module declaring every name a `use` of it installs in a caller, with prototype and typed signature, plus the subs and methods its callers use; a declaration is read in place of the module's source, so a name it leaves out is a name the import loses. A `.tpm` has no `@EXPORT`, `@EXPORT_OK` or `%EXPORT_TAGS` (the parser refuses them): each exported sub or variable carries `:export` (a bare `use` installs it) or `:export_ok` (it must be asked for), with its tags as attribute arguments, e.g. `sub LOCK_EX :export_ok(flock) () Int;`. What is marked is what perl's `use` installs in a caller, measured for a bare `use M`, for `use M qw(name)` and for `use M ':tag'`, not the module's raw lists. A module with its own `import` is declared the same way: mark what that `import` installs. Each file starts with two `ABOUTME:` lines naming the perl (5.42.0) and module version measured. A sub that never returns is declared `None`. XS constant exporters: Fcntl 68 `@EXPORT` / 93 `@EXPORT_OK`; Socket 170 / 140, including the variables `$CR $LF $CRLF`, declared as typed variables. Prototypes from perl's `prototype()`. The oracle calls no Socket sub that opens a socket.

A `.tpm` changes what the parser reads for every file that uses the module, so run the m1 gates after this batch, not only at the end: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor`, `TestGoParserWrongIsZero` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc). Run each with the command the final-gate issue (01a1464e-de34) gives it: `TestRatchetCorpus` skips without `-parseoracle.corpus`, and the two Go-parser gates skip unless `PARSEORACLE_SHIM` names a built shim, and a skip proves nothing. A ratchet may only be raised. A drop is fixed in the .tpm, or the batch stops and asks perigrin.

**Ruling 1 (perigrin, 2026-10-16): no export lists in a `.tpm`.** "We don't need the @EXPORT in the .tpm since we have :export and :export_ok." On names a module's own `import` installs: "we should figure out as best we can what is exported (loc, loc_lang, etc.) and mark them as :export or :export_ok accordingly."

## Steps

- [ ] Write failing tests `TestConstantExportersPrototypes` and `TestConstantExportersExports` in `internal/parse/heldtoperl_test.go` over the module list above
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestConstantExportersPrototypes$|^TestConstantExportersExports$' -count=1 -v`
- [ ] Write the `.tpm` files; run to verify pass; commit
- [ ] Write failing test `TestConstantExportersReturnTypes` in `internal/infer/returntypes_test.go`, with call cases in `internal/infer/testdata/oracle/` and every uncallable sub listed by name with its reason
- [ ] Run to verify failure: `go test ./internal/infer/ -run '^TestConstantExportersReturnTypes$' -count=1 -v`
- [ ] Fix declared return types where perl disagrees; run to verify pass; commit
- [ ] Run the m1 gates above; `make test`; commit

### Review Findings

Gate, 2026-10-20, on f8a82e03 (fresh Sonnet reviewer; six criteria pass, one proven non-vacuous by mutation; the by-name retry cannot hide a missing sub; 13 spot checks match perl). VERDICT: REOPEN.

1. (CODE, 90) Socket.tpm:347 `multi sub sockaddr_in ... (Int $port, Str|Undef $addr) Str;` widens addr to keep Net::Ping's call quiet, but perl 5.42 dies on `Socket::sockaddr_in(80, undef)` ("Undefined address for Socket::pack_sockaddr_in"), and sockaddr_in6 was left plain Str though it dies the same way. Net::Ping.pm:1797-1798 (`wakeonlan`: `my $ip_addr = inet_aton($host); sockaddr_in($port, $ip_addr)`) really can pass undef for an unresolvable host. Fix: declare addr `Str` as perl's contract states; the resulting Net::Ping diagnostic is a true positive, so list it in lib_true_positives.txt with perl -w (or die) output produced by calling Net::Ping's wakeonlan with an unresolvable host, from the real lib path, not a minimal repro.

Out of scope, filed separately: the oracle calls every case in every context, so it cannot hold a sub that dispatches on arity and wantarray together (sockaddr_in, sockaddr_in6 stay exempt).

## Acceptance Criteria

### Positive Scenarios
- [ ] declared prototypes equal perl's prototype() (`go test ./internal/parse/ -run '^TestConstantExportersPrototypes$' -count=1 -v | grep -q '^--- PASS: TestConstantExportersPrototypes'`)
- [ ] declared `:export`/`:export_ok` marks equal what perl's `use` installs in a caller, bare, by name and by tag (`go test ./internal/parse/ -run '^TestConstantExportersExports$' -count=1 -v | grep -q '^--- PASS: TestConstantExportersExports'`)
- [ ] declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestConstantExportersReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestConstantExportersReturnTypes'`)

### Negative Scenarios
- [ ] a bare `use Socket;` does not import `$CRLF`, which is in `@EXPORT_OK` only (measured) (`go test ./internal/parse/ -run '^TestSocketBareUseOmitsCRLFVariable$' -count=1 -v | grep -q '^--- PASS: TestSocketBareUseOmitsCRLFVariable'`)
- [ ] a bare `use Fcntl;` does not import `LOCK_EX`, which is in `@EXPORT_OK` only (measured) (`go test ./internal/parse/ -run '^TestFcntlBareUseOmitsFlockConstants$' -count=1 -v | grep -q '^--- PASS: TestFcntlBareUseOmitsFlockConstants'`)
- [ ] the oracle never resolves a host name: Socket's `inet_aton`, `getaddrinfo` and `getnameinfo` get numeric addresses only, or are exempt with a reason (`go test ./internal/infer/ -run '^TestSocketOracleNeverResolvesAName$' -count=1 -v | grep -q '^--- PASS: TestSocketOracleNeverResolvesAName'`)
