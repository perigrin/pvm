---
title: "Declare the XS utility modules: Scalar::Util, List::Util, Cwd, Data::Dumper, MIME::Base64, Storable"
state: in-progress
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
- 01a1464e-d980-7382-970e-067a5ea3db84
- 01a1464e-db1c-73eb-a3a9-e897d7a56b92
- 01a14683-8375-7810-8a75-b772680603c6
blocks:
- 01a1464e-de34-7403-95d2-25f4855404b2
created: 2026-10-16T20:02:15.593259461Z
updated: 2026-10-20T01:19:53.667854833Z
sessions:
- start_sha: 5b91125f4014127e11963fc610c805ba205b10ff
  end_sha: ""
  commits: 0
  started_at: 2026-10-17T04:25:17.529290165Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-17T04:25:17.529290165Z
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
- [ ] Convert every .tpm to :export and :export_ok, then refuse export lists in a .tpm (01a14683-8375-7810-8a75-b772680603c6)
- [ ] Pin the core-module ranking in a test and record the rule in the RFC
- [ ] Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle

## Context

- paths: `internal/parse/declarations/` (one `.tpm` per module below), `internal/parse/heldtoperl_test.go`, `internal/infer/returntypes_test.go`, `internal/infer/testdata/oracle/`
- commands: `go test ./internal/parse/ -run '^TestXSUtilitiesPrototypes$|^TestXSUtilitiesExports$' -count=1 -v`, `go test ./internal/infer/ -run '^TestXSUtilitiesReturnTypes$' -count=1 -v`
- entrypoints: `internal/parse/declarations/Carp.tpm` (the model: measured facts in comments), the held-to-perl helpers from the Carp issue (01a1464e-db1c)

Modules (from the ranking): Scalar::Util, List::Util, Cwd, Data::Dumper, MIME::Base64, Storable (6).

Write one `.tpm` per module declaring every name a `use` of it installs in a caller, with prototype and typed signature, plus the subs and methods its callers use; a declaration is read in place of the module's source, so a name it leaves out is a name the import loses. A `.tpm` has no `@EXPORT`, `@EXPORT_OK` or `%EXPORT_TAGS` (the parser refuses them): each exported sub or variable carries `:export` (a bare `use` installs it) or `:export_ok` (it must be asked for), with its tags as attribute arguments, e.g. `sub LOCK_EX :export_ok(flock) () Int;`. What is marked is what perl's `use` installs in a caller, measured for a bare `use M`, for `use M qw(name)` and for `use M ':tag'`, not the module's raw lists. A module with its own `import` is declared the same way: mark what that `import` installs. Each file starts with two `ABOUTME:` lines naming the perl (5.42.0) and module version measured. A sub that never returns is declared `None`. These subs are XS: take prototypes from perl's `prototype()`, never from reading `.xs` by hand. List::Util's block-taking subs (`first`, `any`, `reduce`, ...) carry `&@`.

A `.tpm` changes what the parser reads for every file that uses the module, so run the m1 gates after this batch, not only at the end: `TestParseRatchet`, `TestPerlGitTRatchet`, `TestCanonRatchet`, `TestCanonFidelityRatchet`, `TestT2HoldsAll56`, `TestT2CoreParses`, `TestImportResolutionIsWorthT1Files` (internal/parse), `TestRatchet`, `TestRatchetCorpus`, `TestGoParserExactFloor`, `TestGoParserWrongIsZero` (internal/parseoracle), `TestPSCCheckLibBaseline` (internal/psc). Run each with the command the final-gate issue (01a1464e-de34) gives it: `TestRatchetCorpus` skips without `-parseoracle.corpus`, and the two Go-parser gates skip unless `PARSEORACLE_SHIM` names a built shim, and a skip proves nothing. A ratchet may only be raised. A drop is fixed in the .tpm, or the batch stops and asks perigrin.

**Ruling 1 (perigrin, 2026-10-16): no export lists in a `.tpm`.** "We don't need the @EXPORT in the .tpm since we have :export and :export_ok." On names a module's own `import` installs: "we should figure out as best we can what is exported (loc, loc_lang, etc.) and mark them as :export or :export_ok accordingly."

## Steps

- [ ] Write failing tests `TestXSUtilitiesPrototypes` and `TestXSUtilitiesExports` in `internal/parse/heldtoperl_test.go` over the module list above
- [ ] Run to verify failure: `go test ./internal/parse/ -run '^TestXSUtilitiesPrototypes$|^TestXSUtilitiesExports$' -count=1 -v`
- [ ] Write the `.tpm` files; run to verify pass; commit
- [ ] Write failing test `TestXSUtilitiesReturnTypes` in `internal/infer/returntypes_test.go`, with call cases in `internal/infer/testdata/oracle/` and every uncallable sub listed by name with its reason
- [ ] Run to verify failure: `go test ./internal/infer/ -run '^TestXSUtilitiesReturnTypes$' -count=1 -v`
- [ ] Fix declared return types where perl disagrees; run to verify pass; commit
- [ ] Run the m1 gates above; `make test`; commit

### Review Findings

Gate, 2026-10-20, on exec/dba9 before merge (fresh Sonnet reviewer; 26 lib sites read in context). VERDICT: REOPEN (held unmerged). The types.go (CodeRef satisfies Code) and vars.go (deref copy hands on Unknown) fixes are sound and stay.

76 new psc lines over perl's lib were listed as true positives from minimal repros. Classification of the sample:
- REAL: CPAN/Distribution.pm:848; File/Temp.pm:916, 964, 1147; Safe.pm:424; Test2/Compare/Array.pm:20; Test2/Mock.pm:122, 125, 425; Test2/Workflow.pm:111; Test2/IPC/Driver/Files.pm:71; Term/Table.pm:103 left operand (only with zero columns).
- GUARDED (false positives): JSON/PP.pm:398, 411 (blessed + ref guard); Math/BigInt.pm:2400, 3380 (prior method call proves an object); Term/Table.pm:103 right, 106, 124 (sum of three literals); Test2/Compare/Ref.pm:48 and Test2/Tools/Ref.pm:54, 78 (return unless ref); Test2/Mock.pm:54, 57 (falsy branch consumed by if), 356 (ref && reftype), 368 (after elsif !ref); threads/shared.pm:115, 125, 137, 145, 147, 177 (reftype of an item proven a ref at line 100).
- UNKNOWABLE: Test/Builder.pm:207, 259.

perigrin ruled "fix narrowing first": issue 01a156e4 narrows these guards and declares reftype/refaddr/blessed/sum as multis by argument type or arity; dba9 merges after it, and lib_true_positives.txt keeps only lines proven real at the lib site. Out of scope, noted: a bare coderef scalar passed to a `&`-prototyped call (`first($cb, @list)`) is a perl compile error psc does not report.

## Acceptance Criteria

### Positive Scenarios
- [ ] declared prototypes equal perl's prototype() (`go test ./internal/parse/ -run '^TestXSUtilitiesPrototypes$' -count=1 -v | grep -q '^--- PASS: TestXSUtilitiesPrototypes'`)
- [ ] declared `:export`/`:export_ok` marks equal what perl's `use` installs in a caller, bare, by name and by tag (`go test ./internal/parse/ -run '^TestXSUtilitiesExports$' -count=1 -v | grep -q '^--- PASS: TestXSUtilitiesExports'`)
- [ ] declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestXSUtilitiesReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestXSUtilitiesReturnTypes'`)

### Negative Scenarios
- [ ] a bare `use Scalar::Util;` imports nothing (its `@EXPORT` is empty, measured), so `blessed($x)` does not resolve to the declaration (`go test ./internal/parse/ -run '^TestScalarUtilBareUseImportsNothing$' -count=1 -v | grep -q '^--- PASS: TestScalarUtilBareUseImportsNothing'`)
- [ ] `blessed` on an unblessed reference returns undef, which its declared return type admits (`go test ./internal/infer/ -run '^TestScalarUtilBlessedUnblessedReturnIsDeclared$' -count=1 -v | grep -q '^--- PASS: TestScalarUtilBlessedUnblessedReturnIsDeclared'`)
- [ ] `first { $_ } @x` after `use List::Util qw(first);` parses the braces as a block, not a hash constructor (`go test ./internal/parse/ -run '^TestListUtilFirstParsesBlockArgument$' -count=1 -v | grep -q '^--- PASS: TestListUtilFirstParsesBlockArgument'`)
- [ ] MIME::Base64 `decode_base64` given invalid base64 returns a value its declared return type admits (`go test ./internal/infer/ -run '^TestMIMEBase64InvalidInputReturnIsDeclared$' -count=1 -v | grep -q '^--- PASS: TestMIMEBase64InvalidInputReturnIsDeclared'`)
- [ ] Storable oracle cases write files only under the test's temp dir (`go test ./internal/infer/ -run '^TestXSUtilitiesOracleStaysInScratch$' -count=1 -v | grep -q '^--- PASS: TestXSUtilitiesOracleStaysInScratch'`)
