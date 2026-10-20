---
title: "Declare the pure-Perl exporters: File::Basename, Symbol, File::Find, File::Path, Errno and five more"
state: done
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
- 01a1464e-d980-7382-970e-067a5ea3db84
- 01a1464e-db1c-73eb-a3a9-e897d7a56b92
- 01a14683-8375-7810-8a75-b772680603c6
blocks:
- 01a1464e-de34-7403-95d2-25f4855404b2
created: 2026-10-16T20:02:15.500338342Z
updated: 2026-10-20T16:02:35.311002663Z
sessions:
- start_sha: 5b91125f4014127e11963fc610c805ba205b10ff
  end_sha: d7e9819e189d4066bdeda28c06b1f2611f1cccb7
  commits: 21
  started_at: 2026-10-17T04:25:18.126940249Z
  ended_at: 2026-10-20T16:02:35.311002663Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-17T04:25:18.126940249Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-20T16:02:35.311002663Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
- internal/cli/symlinks_test.go
- internal/infer/calls.go
- internal/infer/declared_test.go
- internal/infer/flow.go
- internal/infer/infer.go
- internal/infer/narrow_lib_test.go
- internal/infer/ops.go
- internal/infer/precision_test.go
- internal/infer/returntypes_test.go
- internal/infer/symbols.go
- internal/infer/testdata/lowered/flow/element_guard.golden
- internal/infer/testdata/lowered/flow/elsif_exit.golden
- internal/infer/testdata/lowered/flow/unless_elsif.golden
- internal/infer/testdata/observe_types.pl
- internal/infer/testdata/oracle/Archive-Tar.pl
- internal/infer/testdata/oracle/Class-Struct.pl
- internal/infer/testdata/oracle/Compress-Raw-Zlib.pl
- internal/infer/testdata/oracle/Config.pl
- internal/infer/testdata/oracle/Cwd.pl
- internal/infer/testdata/oracle/Data-Dumper.pl
- internal/infer/testdata/oracle/DynaLoader.pl
- internal/infer/testdata/oracle/Encode.pl
- internal/infer/testdata/oracle/Errno.pl
- internal/infer/testdata/oracle/Exporter.pl
- internal/infer/testdata/oracle/ExtUtils-MakeMaker.pl
- internal/infer/testdata/oracle/ExtUtils-Manifest.pl
- internal/infer/testdata/oracle/Fcntl.pl
- internal/infer/testdata/oracle/File-Basename.pl
- internal/infer/testdata/oracle/File-Copy.pl
- internal/infer/testdata/oracle/File-Find.pl
- internal/infer/testdata/oracle/File-Path.pl
- internal/infer/testdata/oracle/File-Spec-Functions.pl
- internal/infer/testdata/oracle/File-Spec-Unix.pl
- internal/infer/testdata/oracle/File-Spec.pl
- internal/infer/testdata/oracle/File-Temp.pl
- internal/infer/testdata/oracle/File-stat.pl
- internal/infer/testdata/oracle/FileHandle.pl
- internal/infer/testdata/oracle/Getopt-Long.pl
- internal/infer/testdata/oracle/Getopt-Std.pl
- internal/infer/testdata/oracle/HTTP-Tiny.pl
- internal/infer/testdata/oracle/IO-File.pl
- internal/infer/testdata/oracle/IO-Handle.pl
- internal/infer/testdata/oracle/IO-Select.pl
- internal/infer/testdata/oracle/IO-Socket-IP.pl
- internal/infer/testdata/oracle/IO-Socket.pl
- internal/infer/testdata/oracle/List-Util.pl
- internal/infer/testdata/oracle/Locale-Maketext-Simple.pl
- internal/infer/testdata/oracle/MIME-Base64.pl
- internal/infer/testdata/oracle/Net-Ping.pl
- internal/infer/testdata/oracle/POSIX.pl
- internal/infer/testdata/oracle/SDBM_File.pl
- internal/infer/testdata/oracle/Scalar-Util.pl
- internal/infer/testdata/oracle/Socket.pl
- internal/infer/testdata/oracle/Storable.pl
- internal/infer/testdata/oracle/Symbol.pl
- internal/infer/testdata/oracle/Term-ReadLine.pl
- internal/infer/testdata/oracle/Test-More-plan.pl
- internal/infer/testdata/oracle/Test-More.pl
- internal/infer/testdata/oracle/Text-ParseWords.pl
- internal/infer/testdata/oracle/Text-Wrap.pl
- internal/infer/testdata/oracle/Tie-Hash.pl
- internal/infer/testdata/oracle/Time-HiRes.pl
- internal/infer/testdata/oracle/Time-Local.pl
- internal/infer/testdata/oracle/XSLoader.pl
- internal/infer/testdata/return_oracle.pl
- internal/infer/vars.go
- internal/infer/vars_test.go
- internal/infer/warnpossibleundef_test.go
- internal/parse/declarations.go
- internal/parse/declarations/Archive/Tar.tpm
- internal/parse/declarations/Class/Struct.tpm
- internal/parse/declarations/Compress/Raw/Zlib.tpm
- internal/parse/declarations/Config.tpm
- internal/parse/declarations/Cwd.tpm
- internal/parse/declarations/Data/Dumper.tpm
- internal/parse/declarations/DynaLoader.tpm
- internal/parse/declarations/Encode.tpm
- internal/parse/declarations/Errno.tpm
- internal/parse/declarations/Exporter.tpm
- internal/parse/declarations/ExtUtils/MakeMaker.tpm
- internal/parse/declarations/ExtUtils/Manifest.tpm
- internal/parse/declarations/Fcntl.tpm
- internal/parse/declarations/File/Basename.tpm
- internal/parse/declarations/File/Copy.tpm
- internal/parse/declarations/File/Find.tpm
- internal/parse/declarations/File/Path.tpm
- internal/parse/declarations/File/Spec.tpm
- internal/parse/declarations/File/Spec/Functions.tpm
- internal/parse/declarations/File/Spec/Unix.tpm
- internal/parse/declarations/File/Temp.tpm
- internal/parse/declarations/File/stat.tpm
- internal/parse/declarations/FileHandle.tpm
- internal/parse/declarations/Getopt/Long.tpm
- internal/parse/declarations/Getopt/Std.tpm
- internal/parse/declarations/HTTP/Tiny.tpm
- internal/parse/declarations/IO/File.tpm
- internal/parse/declarations/IO/Handle.tpm
- internal/parse/declarations/IO/Select.tpm
- internal/parse/declarations/IO/Socket.tpm
- internal/parse/declarations/IO/Socket/IP.tpm
- internal/parse/declarations/List/Util.tpm
- internal/parse/declarations/Locale/Maketext/Simple.tpm
- internal/parse/declarations/MIME/Base64.tpm
- internal/parse/declarations/Net/Ping.tpm
- internal/parse/declarations/POSIX.tpm
- internal/parse/declarations/SDBM_File.tpm
- internal/parse/declarations/Scalar/Util.tpm
- internal/parse/declarations/Socket.tpm
- internal/parse/declarations/Storable.tpm
- internal/parse/declarations/Symbol.tpm
- internal/parse/declarations/Term/ReadLine.tpm
- internal/parse/declarations/Test/More.tpm
- internal/parse/declarations/Text/ParseWords.tpm
- internal/parse/declarations/Text/Wrap.tpm
- internal/parse/declarations/Tie/Hash.tpm
- internal/parse/declarations/Time/HiRes.tpm
- internal/parse/declarations/Time/Local.tpm
- internal/parse/declarations/XSLoader.tpm
- internal/parse/derive.go
- internal/parse/derive_test.go
- internal/parse/exportattr_test.go
- internal/parse/heldtoperl_test.go
- internal/parse/infrastructure_test.go
- internal/parse/testdata/declarations/My/MarkedImport.tpm
- internal/parse/testdata/held_to_perl.pl
- internal/parse/testdata/t1.ratchet
- internal/parse/use.go
- internal/parse/xsutilities_test.go
- internal/psc/check_command.go
- internal/psc/check_command_test.go
- internal/psc/lsp_command.go
- internal/psc/lsp_test.go
- internal/psc/testdata/lib_true_positives.txt
- internal/types/select.go
- internal/types/select_test.go
- internal/types/types.go
- internal/types/types_test.go
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
- [ ] Convert every .tpm to :export and :export_ok, then refuse export lists in a .tpm (01a14683-8375-7810-8a75-b772680603c6)
- [ ] Pin the core-module ranking in a test and record the rule in the RFC
- [ ] Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle

## Context

- paths: `internal/parse/declarations/` (one `.tpm` per module below), `internal/parse/heldtoperl_test.go`, `internal/infer/returntypes_test.go`, `internal/infer/testdata/oracle/`
- commands: `go test ./internal/parse/ -run '^TestPurePerlExportersPrototypes$|^TestPurePerlExportersExports$' -count=1 -v`, `go test ./internal/infer/ -run '^TestPurePerlExportersReturnTypes$' -count=1 -v`
- entrypoints: `internal/parse/declarations/Carp.tpm` (the model: measured facts in comments), the held-to-perl helpers from the Carp issue (01a1464e-db1c)

Modules (from the ranking): File::Basename, Symbol, File::Find, File::Path, Errno, Text::ParseWords, Class::Struct, File::Copy, Time::Local, Getopt::Std (10; Carp is held to perl by its own issue).

Write one `.tpm` per module declaring every name a `use` of it installs in a caller, with prototype and typed signature, plus the subs and methods its callers use; a declaration is read in place of the module's source, so a name it leaves out is a name the import loses. A `.tpm` has no `@EXPORT`, `@EXPORT_OK` or `%EXPORT_TAGS` (the parser refuses them): each exported sub or variable carries `:export` (a bare `use` installs it) or `:export_ok` (it must be asked for), with its tags as attribute arguments, e.g. `sub LOCK_EX :export_ok(flock) () Int;`. What is marked is what perl's `use` installs in a caller, measured for a bare `use M`, for `use M qw(name)` and for `use M ':tag'`, not the module's raw lists. A module with its own `import` is declared the same way: mark what that `import` installs. Each file starts with two `ABOUTME:` lines naming the perl (5.42.0) and module version measured. A sub that never returns is declared `None`. Errno's names are constants generated at build (`ENOENT` and the rest, 134 in `@EXPORT_OK`, tag `:POSIX`); measured 2026-10-16, Errno exports no variable -- `%!` is not in its export lists -- so it is here, not with the variable exporters.

A `.tpm` changes what the parser reads for every file that uses the module, so run the m1 gates after this batch, not only at the end: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor`, `TestGoParserWrongIsZero` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc). Run each with the command the final-gate issue (01a1464e-de34) gives it: `TestRatchetCorpus` skips without `-parseoracle.corpus`, and the two Go-parser gates skip unless `PARSEORACLE_SHIM` names a built shim, and a skip proves nothing. A ratchet may only be raised. A drop is fixed in the .tpm, or the batch stops and asks perigrin.

**Ruling 1 (perigrin, 2026-10-16): no export lists in a `.tpm`.** "We don't need the @EXPORT in the .tpm since we have :export and :export_ok." On names a module's own `import` installs: "we should figure out as best we can what is exported (loc, loc_lang, etc.) and mark them as :export or :export_ok accordingly."

## Steps

- [ ] Write failing tests `TestPurePerlExportersPrototypes` and `TestPurePerlExportersExports` in `internal/parse/heldtoperl_test.go` over the module list above
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestPurePerlExportersPrototypes$|^TestPurePerlExportersExports$' -count=1 -v`
- [ ] Write the `.tpm` files; run to verify pass; commit
- [ ] Write failing test `TestPurePerlExportersReturnTypes` in `internal/infer/returntypes_test.go`, with call cases in `internal/infer/testdata/oracle/` and every uncallable sub listed by name with its reason
- [ ] Run to verify failure: `go test ./internal/infer/ -run '^TestPurePerlExportersReturnTypes$' -count=1 -v`
- [ ] Fix declared return types where perl disagrees; run to verify pass; commit
- [ ] Run the m1 gates above; `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] declared prototypes equal perl's prototype() (`go test ./internal/parse/ -run '^TestPurePerlExportersPrototypes$' -count=1 -v | grep -q '^--- PASS: TestPurePerlExportersPrototypes'`)
- [ ] declared `:export`/`:export_ok` marks equal what perl's `use` installs in a caller, bare, by name and by tag (`go test ./internal/parse/ -run '^TestPurePerlExportersExports$' -count=1 -v | grep -q '^--- PASS: TestPurePerlExportersExports'`)
- [ ] declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestPurePerlExportersReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestPurePerlExportersReturnTypes'`)

### Negative Scenarios
- [ ] Errno.tpm declares no exported variable: `%!` is in none of Errno's export lists (measured) (`go test ./internal/parse/ -run '^TestErrnoDeclaresNoExportedVariable$' -count=1 -v | grep -q '^--- PASS: TestErrnoDeclaresNoExportedVariable'`)
- [ ] Getopt::Std `getopts` given an unknown option returns a value its declared return type admits (`go test ./internal/infer/ -run '^TestGetoptStdUnknownOptionReturnIsDeclared$' -count=1 -v | grep -q '^--- PASS: TestGetoptStdUnknownOptionReturnIsDeclared'`)
- [ ] Text::ParseWords `shellwords` given an unbalanced quote returns a value its declared return type admits (`go test ./internal/infer/ -run '^TestTextParseWordsUnbalancedQuoteReturnIsDeclared$' -count=1 -v | grep -q '^--- PASS: TestTextParseWordsUnbalancedQuoteReturnIsDeclared'`)
- [ ] File::Basename `fileparse`, `basename` and `dirname` given a Windows path such as `C:\dir\f.txt` return values their declared types admit (`go test ./internal/infer/ -run '^TestFileBasenameWindowsPathReturnIsDeclared$' -count=1 -v | grep -q '^--- PASS: TestFileBasenameWindowsPathReturnIsDeclared'`)
- [ ] File::Path, File::Copy and File::Find oracle cases touch nothing outside the test's temp dir (`go test ./internal/infer/ -run '^TestPurePerlExportersOracleStaysInScratch$' -count=1 -v | grep -q '^--- PASS: TestPurePerlExportersOracleStaysInScratch'`)
