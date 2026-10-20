---
title: "Declare the variable exporters: Config, Getopt::Long, Test::More, Text::Wrap, ExtUtils::MakeMaker, File::stat"
state: reopened
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
created: 2026-10-16T20:02:15.843622709Z
updated: 2026-10-20T15:30:07.85239748Z
sessions:
- start_sha: 34826451ff25d9be31c553948ea6de4ec295f77d
  end_sha: 9e5ae578baacecaaf348477c8982d9720ed83cf0
  commits: 2
  started_at: 2026-10-20T02:59:44.677331931Z
  ended_at: 2026-10-20T03:43:26.128274693Z
- start_sha: ac90b82c584f38414304222586f301892e559272
  end_sha: 495c3e916a4a7469c1dd381b63c41207206e73d2
  commits: 3
  started_at: 2026-10-20T14:18:47.80930801Z
  ended_at: 2026-10-20T15:07:31.104133424Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T02:59:44.677331931Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-20T03:43:26.128274693Z
- state: reopened
  actor: human:git-zhi
  timestamp: 2026-10-20T13:38:14.800054743Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T14:18:47.80930801Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-20T15:07:31.104133424Z
- state: reopened
  actor: human:git-zhi
  timestamp: 2026-10-20T15:30:07.85239748Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
- internal/infer/declared_test.go
- internal/infer/returntypes_test.go
- internal/infer/testdata/oracle/Archive-Tar.pl
- internal/infer/testdata/oracle/Config.pl
- internal/infer/testdata/oracle/DynaLoader.pl
- internal/infer/testdata/oracle/Exporter.pl
- internal/infer/testdata/oracle/ExtUtils-MakeMaker.pl
- internal/infer/testdata/oracle/ExtUtils-Manifest.pl
- internal/infer/testdata/oracle/File-Spec-Functions.pl
- internal/infer/testdata/oracle/File-Spec-Unix.pl
- internal/infer/testdata/oracle/File-Spec.pl
- internal/infer/testdata/oracle/File-Temp.pl
- internal/infer/testdata/oracle/File-stat.pl
- internal/infer/testdata/oracle/FileHandle.pl
- internal/infer/testdata/oracle/Getopt-Long.pl
- internal/infer/testdata/oracle/HTTP-Tiny.pl
- internal/infer/testdata/oracle/IO-File.pl
- internal/infer/testdata/oracle/IO-Handle.pl
- internal/infer/testdata/oracle/IO-Select.pl
- internal/infer/testdata/oracle/IO-Socket-IP.pl
- internal/infer/testdata/oracle/IO-Socket.pl
- internal/infer/testdata/oracle/Locale-Maketext-Simple.pl
- internal/infer/testdata/oracle/Net-Ping.pl
- internal/infer/testdata/oracle/Term-ReadLine.pl
- internal/infer/testdata/oracle/Test-More-plan.pl
- internal/infer/testdata/oracle/Test-More.pl
- internal/infer/testdata/oracle/Text-Wrap.pl
- internal/infer/testdata/oracle/Tie-Hash.pl
- internal/infer/testdata/oracle/XSLoader.pl
- internal/infer/testdata/return_oracle.pl
- internal/parse/declarations.go
- internal/parse/declarations/Archive/Tar.tpm
- internal/parse/declarations/Config.tpm
- internal/parse/declarations/DynaLoader.tpm
- internal/parse/declarations/Encode.tpm
- internal/parse/declarations/Exporter.tpm
- internal/parse/declarations/ExtUtils/MakeMaker.tpm
- internal/parse/declarations/ExtUtils/Manifest.tpm
- internal/parse/declarations/File/Spec.tpm
- internal/parse/declarations/File/Spec/Functions.tpm
- internal/parse/declarations/File/Spec/Unix.tpm
- internal/parse/declarations/File/Temp.tpm
- internal/parse/declarations/File/stat.tpm
- internal/parse/declarations/FileHandle.tpm
- internal/parse/declarations/Getopt/Long.tpm
- internal/parse/declarations/HTTP/Tiny.tpm
- internal/parse/declarations/IO/File.tpm
- internal/parse/declarations/IO/Handle.tpm
- internal/parse/declarations/IO/Select.tpm
- internal/parse/declarations/IO/Socket.tpm
- internal/parse/declarations/IO/Socket/IP.tpm
- internal/parse/declarations/Locale/Maketext/Simple.tpm
- internal/parse/declarations/Net/Ping.tpm
- internal/parse/declarations/Term/ReadLine.tpm
- internal/parse/declarations/Test/More.tpm
- internal/parse/declarations/Text/Wrap.tpm
- internal/parse/declarations/Tie/Hash.tpm
- internal/parse/declarations/XSLoader.tpm
- internal/parse/exportattr_test.go
- internal/parse/heldtoperl_test.go
- internal/parse/infrastructure_test.go
- internal/parse/testdata/declarations/My/MarkedImport.tpm
- internal/parse/testdata/held_to_perl.pl
- internal/parse/testdata/t1.ratchet
- internal/parse/use.go
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
- [ ] Convert every .tpm to :export and :export_ok, then refuse export lists in a .tpm (01a14683-8375-7810-8a75-b772680603c6)
- [ ] Pin the core-module ranking in a test and record the rule in the RFC
- [ ] Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle
- [ ] Read typed variables with :export and :export_ok in declaration files

## Context

- paths: `internal/parse/declarations/` (one `.tpm` per module below), `internal/parse/heldtoperl_test.go`, `internal/infer/returntypes_test.go`, `internal/infer/testdata/oracle/`
- commands: `go test ./internal/parse/ -run '^TestVariableExportersPrototypes$|^TestVariableExportersExports$' -count=1 -v`, `go test ./internal/infer/ -run '^TestVariableExportersReturnTypes$' -count=1 -v`
- entrypoints: `internal/parse/declarations/Carp.tpm` (the model: measured facts in comments), the held-to-perl helpers from the Carp issue (01a1464e-db1c)

Modules (from the ranking): Config, Getopt::Long, Test::More, Text::Wrap, ExtUtils::MakeMaker, File::stat (6).

Write one `.tpm` per module declaring every name a `use` of it installs in a caller, with prototype and typed signature, plus the subs and methods its callers use; a declaration is read in place of the module's source, so a name it leaves out is a name the import loses. A `.tpm` has no `@EXPORT`, `@EXPORT_OK` or `%EXPORT_TAGS` (the parser refuses them): each exported sub or variable carries `:export` (a bare `use` installs it) or `:export_ok` (it must be asked for), with its tags as attribute arguments, e.g. `sub LOCK_EX :export_ok(flock) () Int;`. What is marked is what perl's `use` installs in a caller, measured for a bare `use M`, for `use M qw(name)` and for `use M ':tag'`, not the module's raw lists. A module with its own `import` is declared the same way: mark what that `import` installs. Each file starts with two `ABOUTME:` lines naming the perl (5.42.0) and module version measured. A sub that never returns is declared `None`. Each exports a variable (measured 2026-10-16): Config `%Config` (`:export`) and subs in `@EXPORT_OK`; Getopt::Long `$REQUIRE_ORDER $PERMUTE $RETURN_IN_ORDER`; Test::More `$TODO`; Text::Wrap `$columns $break $huge`; ExtUtils::MakeMaker `$Verbose $VERSION`; File::stat thirteen `$st_*`. Declare each as `our Type <var> :export;` or `:export_ok`. Config and Test::More have their own `import`. The rule for every module with its own `import`: its `:export`/`:export_ok` marks are what that `import` installs in a caller, measured for a bare `use`, `use M qw(name)` and `use M ':tag'`, never its raw lists; an import argument that is neither a name nor a tag (`use Test::More tests => 3`) is measured too, and declares no name. `%Config` is declared `our Hash[Str|Undef] %Config :export;` (ruling 3). Getopt::Long's lists spell six subs with a leading `&` (`&GetOptions`, `&HelpMessage`, `&VersionMessage`, `&Configure`, `&GetOptionsFromArray`, `&GetOptionsFromString`, measured 2026-10-16): each is the sub of that name, declared without the `&` (the Carp issue's `&name` rule). Test::More is used across the T1 corpus, so expect T1 to move.

A `.tpm` changes what the parser reads for every file that uses the module, so run the m1 gates after this batch, not only at the end: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor`, `TestGoParserWrongIsZero` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc). Run each with the command the final-gate issue (01a1464e-de34) gives it: `TestRatchetCorpus` skips without `-parseoracle.corpus`, and the two Go-parser gates skip unless `PARSEORACLE_SHIM` names a built shim, and a skip proves nothing. A ratchet may only be raised. A drop is fixed in the .tpm, or the batch stops and asks perigrin.

**Ruling 1 (perigrin, 2026-10-16): no export lists in a `.tpm`.** "We don't need the @EXPORT in the .tpm since we have :export and :export_ok." On names a module's own `import` installs: "we should figure out as best we can what is exported (loc, loc_lang, etc.) and mark them as :export or :export_ok accordingly."

**Ruling 3 (perigrin, 2026-10-16):** `%Config`'s type is `Hash[Str|Undef]` (172 values are undef on 5.42).

## Steps

- [ ] Write failing tests `TestVariableExportersPrototypes` and `TestVariableExportersExports` in `internal/parse/heldtoperl_test.go` over the module list above
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestVariableExportersPrototypes$|^TestVariableExportersExports$' -count=1 -v`
- [ ] Write the `.tpm` files; run to verify pass; commit
- [ ] Write failing test `TestVariableExportersReturnTypes` in `internal/infer/returntypes_test.go`, with call cases in `internal/infer/testdata/oracle/` and every uncallable sub listed by name with its reason
- [ ] Run to verify failure: `go test ./internal/infer/ -run '^TestVariableExportersReturnTypes$' -count=1 -v`
- [ ] Fix declared return types where perl disagrees; run to verify pass; commit
- [ ] Run the m1 gates above; `make test`; commit

### Review Findings from dadb's gate (forward risk)

(confidence 65) A .tpm has no way to say "this import list is arguments, not names": .tpm facts never set `facts.builder`, so `use Test::More tests => 5;` imports nothing (wanted = `tests`, `5`, both filtered), and likewise `use Locale::Maketext::Simple (Style => 'gettext')` imports neither `loc` nor `loc_lang`. Not a regression (neither worked before dadb), but this issue must decide, by measuring what perl's `use` installs, how a declaration states that its module's import takes arguments, and test the bare and argument forms.

### Review Findings

Gate, 2026-10-20, on 9e5ae578 (fresh Sonnet reviewer; 10 criteria pass, one proven by mutation; 12 types and 8 :listop subs match perl 5.42). VERDICT: REOPEN.

1. (CODE, 85) internal/parse/module.go:376-385 hard-codes `if module == "Test::More" { facts.builder = true }`, a per-module special case that recurs for Locale::Maketext::Simple (`use Locale::Maketext::Simple (Style => 'gettext')`, dd27), and Test/More.tpm and commit 07d6206d cite an "RFC 0001 open question" that does not exist.
   **perigrin's ruling, 2026-10-20: "The parameters are the signature to import()."** A .tpm states what a `use` list takes by declaring the module's `import` with a signature: the use-list is the argument list of `import`, checked against that signature, and a module whose .tpm declares `import` does not treat its use-list as export names (e.g. Test::More: `sub import (Str $class, List @plan) Void;`, measured; Locale::Maketext::Simple: the hash-pair options its import takes, measured, with loc/loc_lang still marked :export as its import installs them). Replace the by-name case with this; record the rule in RFC 0001 (Declaration files / Every exported name), not as an open question; correct the false citation; apply it to Locale::Maketext::Simple (dd27) too.
2. (TEST-ONLY, 65; fix in place) Test::More::plan is exempt for colliding with done_testing in the same oracle run and has no call case: hold it in a separate oracleRun whose cases file holds only `plan => [[ tests => 1 ]]`.

### Review Findings (pass 2)

Gate pass 2, 2026-10-20, on 495c3e91 (fresh Sonnet reviewer; 10 criteria pass; the by-name case is gone; the marked-import guard is load-bearing). VERDICT: REOPEN.

1. (CODE, 90) internal/parse/use.go:376-385 treats any non-empty use-list to a module with a declared import as installing the bare-use names, but perl's import branches on argument values: `use Locale::Maketext::Simple (Export=>"zz")` installs zz (and zz_lang), not loc; `use Test::More import => ["!is"], tests=>1` does not install is. psc would claim names that do not exist.
   **perigrin's ruling, 2026-10-20: the signature names the keys.** A .tpm's import signature names, as measured, the keys whose presence changes what import installs (Test::More: `import`; Locale::Maketext::Simple: `Export`). A use-list without such a key installs the bare-use names; one with such a key claims no names and leaves the symbols open. Choose the smallest signature syntax that states this, record it in RFC 0001 beside the import rule, and test both modules' altering and non-altering forms against perl 5.42.

## Acceptance Criteria

### Positive Scenarios
- [ ] declared prototypes equal perl's prototype() (`go test ./internal/parse/ -run '^TestVariableExportersPrototypes$' -count=1 -v | grep -q '^--- PASS: TestVariableExportersPrototypes'`)
- [ ] declared `:export`/`:export_ok` marks equal what perl's `use` installs in a caller, bare, by name and by tag (`go test ./internal/parse/ -run '^TestVariableExportersExports$' -count=1 -v | grep -q '^--- PASS: TestVariableExportersExports'`)
- [ ] declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestVariableExportersReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestVariableExportersReturnTypes'`)

### Negative Scenarios
- [ ] `%Config` is declared `Hash[Str|Undef]` (ruling 3) and every value is a member; 172 values are undef on 5.42.0 (measured), so `Hash[Str]` would fail (`go test ./internal/infer/ -run '^TestConfigValuesFitDeclaredType$' -count=1 -v | grep -q '^--- PASS: TestConfigValuesFitDeclaredType'`)
- [ ] a bare `use Config;` imports `%Config` and none of Config's `@EXPORT_OK` subs (`go test ./internal/parse/ -run '^TestConfigBareUseImportsOnlyConfigHash$' -count=1 -v | grep -q '^--- PASS: TestConfigBareUseImportsOnlyConfigHash'`)
- [ ] a bare `use Getopt::Long;` does not import `Configure`, which is in `@EXPORT_OK` only (measured) (`go test ./internal/parse/ -run '^TestGetoptLongBareUseOmitsConfigure$' -count=1 -v | grep -q '^--- PASS: TestGetoptLongBareUseOmitsConfigure'`)
- [ ] Test::More `BAIL_OUT` is declared `None` and is exempt from the oracle, since it exits rather than dies (`go test ./internal/infer/ -run '^TestTestMoreBailOutNeverCalled$' -count=1 -v | grep -q '^--- PASS: TestTestMoreBailOutNeverCalled'`)
- [ ] ExtUtils::MakeMaker `prompt` and `WriteMakefile` are exempt with a reason and never called (`go test ./internal/infer/ -run '^TestMakeMakerInteractiveSubsExempt$' -count=1 -v | grep -q '^--- PASS: TestMakeMakerInteractiveSubsExempt'`)
- [ ] `use File::stat ();` leaves `stat($f)` as CORE's stat, not File::stat's (`go test ./internal/parse/ -run '^TestFileStatEmptyImportKeepsCoreStat$' -count=1 -v | grep -q '^--- PASS: TestFileStatEmptyImportKeepsCoreStat'`)
- [ ] File::stat `stat` on a missing file returns the empty list (measured), which its declared return type admits (`go test ./internal/infer/ -run '^TestFileStatMissingFileReturnIsDeclared$' -count=1 -v | grep -q '^--- PASS: TestFileStatMissingFileReturnIsDeclared'`)
