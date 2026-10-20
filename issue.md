---
title: Narrow undef after ref, defined and blessed guards, so typed Scalar::Util and List::Util returns add no false positive
state: done
urgency: normal
milestone: core-module-types
created: 2026-10-20T01:19:39.205726373Z
updated: 2026-10-20T16:34:25.232726306Z
sessions:
- start_sha: e80f4f920667615dc28874844b338224612fffe3
  end_sha: d7e9819e189d4066bdeda28c06b1f2611f1cccb7
  commits: 17
  started_at: 2026-10-20T01:19:55.217000079Z
  ended_at: 2026-10-20T16:02:19.389253128Z
- start_sha: 6304f38e2c9293c66b4aec0e1aee6148414b1b1c
  end_sha: 26e1e411d74a1cc918cb03c0e49393735133d349
  commits: 1
  started_at: 2026-10-20T16:17:50.740421899Z
  ended_at: 2026-10-20T16:34:25.232726306Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T01:19:55.217000079Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-20T16:02:19.389253128Z
- state: reopened
  actor: human:git-zhi
  timestamp: 2026-10-20T16:17:50.682845856Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T16:17:50.740421899Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-20T16:34:25.232726306Z
observed_paths:
- docs/contributing/coding-conventions.md
- docs/contributing/core-module-declarations.md
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

perigrin ruled on 2026-10-17 (issue dba9's gate): fix narrowing first. Declaring Scalar::Util, List::Util and Cwd's returns as `...|Undef` (which perl 5.42 measures) makes psc check report 76 new lib sites. dba9's reviewer read 26 of them: about 11 are real (undef reaches eq/=~ unguarded), about 13 are false positives because the code already rules undef out, 2 are unknowable. Goal (5) allows no added false positive, so dba9 cannot merge until psc sees through those guards.

## Context
- Branch from exec/dba9 (the .tpm files that expose the sites).
- The reviewer's classification, with file:line and reason for each site, is in dba9's issue (01a1464e-dba9) gate record; the sites are under perl 5.42's privlib.
- The guard shapes that must narrow, from that sample:
  - `ref($x)` truthy, including `return ... unless ref $x;`, `if (!ref($x)) {...} elsif ...`, and `ref($x) && reftype($x) eq 'CODE'` (short-circuit): after it, `$x` is a Ref, so a multi like reftype/refaddr/blessed selects its Ref candidate (Str/Int), not Undef.
  - `if (!$t) {...} elsif ($t eq 'HASH')`: the falsy branch consumes Undef.
  - A value assigned from reftype($item) after `$item` was proven a ref (threads/shared.pm:100-177).
- Perl decides each declaration: reftype, refaddr and blessed return a defined value for a reference argument and undef otherwise, so declare them as multis selected by argument type (`(Ref $r) Str` beside `(Any $x) Undef`, measured), rather than one `Str|Undef`. `List::Util::sum` with one or more arguments returns a number and with none returns undef: declare by arity.
- Selection by argument type and arity already exists (types.Select, da0a, 5700).

## Steps
- [ ] For each reviewer-classified site, a test fixture reproducing its guard shape; the GUARDED ones must give no diagnostic, the REAL ones must still be reported.
- [ ] Declare reftype, refaddr, blessed and sum (and any other dba9 sub with the same shape) as multis by argument type or arity, each measured with perl 5.42.
- [ ] Narrow a variable after `ref`, `defined` and `blessed` guards, short-circuit `&&`, `unless ... return`, and an if/elsif chain whose earlier branch consumes the falsy value.
- [ ] Re-run psc check over perl 5.42's lib on the dba9 tree: no line is added that the reviewer classified GUARDED; every remaining added line is in lib_true_positives.txt with perl -w output from the real lib site's path, not a minimal repro of the construct.
- [ ] Precision stays >= 65/76 with WRONG 0.

### Added 2026-10-20: Symbol (from db4c)

db4c's worker parked Symbol.tpm (scratchpad/db4c-symbol-deferred/) because measuring `Symbol::qualify` accurately (Str for a name, the argument unchanged for a ref: Str|GlobRef on perl 5.42) adds `IO/Handle.pm:515`, `tell qualify($_[0], caller) if ref($_[0])`, as a new psc diagnostic ("argument 1 expects Glob"). perl -w is silent there: `$_[0]` is the invocant, proven a ref by the `if ref($_[0])` guard. Same root cause as this issue: narrow `$_[0]` after the `ref` guard, and declare qualify / qualify_to_ref as multis by argument type (a Ref argument returns that Ref; a Str returns Str), measured. Add IO/Handle.pm:515 to TestNarrowGuardedLibSites; db4c ships Symbol.tpm after this lands.

### Added 2026-10-20: Encode::encode (from dbd3)

dbd3 declared `Encode::encode` as plain `Str` (commented, pointing here), though perl 5.42 measures `Str|Undef`, because the accurate type adds one guarded site, Pod/Text.pm's `encode($encoding, $text, $check)` (both arguments defined in practice). After the narrowing lands, restore `encode` to its measured `Str|Undef` (or a multi by argument type, if undef is returned only for an undef input) and confirm the Pod/Text.pm site stays silent. The declared type must match perl before this issue closes.

### Review Findings

Gate, 2026-10-21, on d7e9819e (fresh Sonnet reviewer; psc check over perl 5.42's lib adds exactly Sys/Syslog.pm:562, with real lib-path evidence; possible-undef probed sound flag on and off; precision 65/76 WRONG 0; Select changes sound; 14 declarations match perl). VERDICT: REOPEN.

1. (CODE, 80) internal/infer/infer.go ~656-667 (Options.WarnPossibleUndef godoc) and internal/psc/check_command.go:89 (--warn-possible-undef help) still describe the pre-generalization rule ("only through an Undecided multi call's join ... A Selected call's own declared return ... reports either way"), now false: `blessed($obj) eq 'Foo'` is silent by default. Rewrite both to perigrin's general rule: any declared Undef the call's arguments do not prove is possible-undef; only an outcome the arguments force outright (sum() with no arguments) reports.
2. (TEST-ONLY, 85; fix in place) TestNarrowGuardedLibSites passes with flow.go and vars.go reverted to before the merge, because possible-undef silences every subtest regardless of narrowing. Add a case whose narrowed return carries no Undef at all, or assert the resolved type or selected candidate, so the test fails without narrowing.
3. (TEST-ONLY, 90; fix in place, issue db4c) internal/infer/returntypes_test.go's pureExporters never gained "Symbol", so Symbol's oracle cases are never run; its comment still says Symbol is held back. Add it and fix the comment.

## Acceptance criteria

### Positive Scenarios
- [ ] Guarded sites give no diagnostic `go test ./internal/infer/ -run '^TestNarrowGuardedLibSites$' -count=1 -v | grep -q '^--- PASS: TestNarrowGuardedLibSites'`
- [ ] Real sites are still reported `go test ./internal/infer/ -run '^TestNarrowKeepsRealLibSites$' -count=1 -v | grep -q '^--- PASS: TestNarrowKeepsRealLibSites'`
- [ ] reftype of a proven reference is a Str `go test ./internal/infer/ -run '^TestReftypeOfRefIsStr$' -count=1 -v | grep -q '^--- PASS: TestReftypeOfRefIsStr'`
- [ ] sum with arguments is a number `go test ./internal/infer/ -run '^TestSumWithArgsIsNum$' -count=1 -v | grep -q '^--- PASS: TestSumWithArgsIsNum'`

### Negative Scenarios
- [ ] psc check over perl's lib adds no unlisted line `go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`
- [ ] Precision holds with WRONG 0 `go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`
