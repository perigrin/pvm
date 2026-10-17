---
title: "Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle"
state: reopened
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
- 01a14683-8375-7810-8a75-b772680603c6
blocks:
- 01a1464e-de34-7403-95d2-25f4855404b2
- 01a1464e-db4c-78ae-9bd5-84d71ccf5432
- 01a1464e-db76-704f-bd85-023a54d2cf6e
- 01a1464e-dba9-740b-b8b6-8da484195ce3
- 01a1464e-dbd3-7b16-801c-b9a5893d687b
- 01a1464e-dc07-754e-9ef4-637df15d51a9
- 01a1464e-dc48-720b-9ab9-fcc7fe47f483
- 01a1464e-dca3-7284-85c3-a06107c16646
- 01a1464e-dd27-7b82-ae71-3e6173375fa7
- 01a1464e-dd6c-7b5f-b95b-76286ec110a1
- 01a1464e-ddb0-720a-93f6-450c767e6e2e
- 01a1464e-dde3-73fe-ab1f-ced7f18c4f9d
- 01a14657-682e-7432-8ea9-828f56cd2f69
created: 2026-10-16T20:02:15.452693569Z
updated: 2026-10-17T04:09:51.713319914Z
sessions:
- start_sha: dcc066cc0fb89509812adfaa862f2096d7e4c67f
  end_sha: c707b8354a29d997aed698fe26d8f94558304d31
  commits: 3
  started_at: 2026-10-16T23:04:29.759612367Z
  ended_at: 2026-10-17T00:03:13.570061661Z
- start_sha: 6c727e93d5c2761bfd2d0e4f8eb56a53088e130d
  end_sha: ad6541f641ce73b412934fe86afbae6d6ed5b022
  commits: 4
  started_at: 2026-10-17T00:11:54.242481314Z
  ended_at: 2026-10-17T00:58:20.045749942Z
- start_sha: 8d17e0b3c88c2e801f39107186c29dd64c8491c8
  end_sha: ae36616442d354352dbe47f0a6e4b9bab924e145
  commits: 3
  started_at: 2026-10-17T03:36:55.518360748Z
  ended_at: 2026-10-17T04:01:49.465622649Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-16T23:04:29.759612367Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-17T00:03:13.570061661Z
- state: reopened
  actor: human:git-zhi
  timestamp: 2026-10-17T00:11:54.160103423Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-17T00:11:54.242481314Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-17T00:58:20.045749942Z
- state: reopened
  actor: human:git-zhi
  timestamp: 2026-10-17T03:36:55.325543136Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-17T03:36:55.518360748Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-17T04:01:49.465622649Z
- state: reopened
  actor: human:git-zhi
  timestamp: 2026-10-17T04:09:51.713319914Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
- internal/data/help/types.md
- internal/infer/calls.go
- internal/infer/declared_test.go
- internal/infer/export_test.go
- internal/infer/flow.go
- internal/infer/flow_test.go
- internal/infer/infer.go
- internal/infer/ops.go
- internal/infer/returntypes_test.go
- internal/infer/special.go
- internal/infer/testdata/lowered/special/sort_block.golden
- internal/infer/testdata/lowered/special/sort_list.golden
- internal/infer/testdata/lowered/special/sort_plain.golden
- internal/infer/testdata/oracle/Carp.pl
- internal/infer/testdata/return_oracle.pl
- internal/parse/decl.go
- internal/parse/declarations.go
- internal/parse/declarations/Carp.tpm
- internal/parse/exportattr_test.go
- internal/parse/heldtoperl_test.go
- internal/parse/library_tpm_test.go
- internal/parse/qualified_test.go
- internal/parse/testdata/declarations/My/Multi.tpm
- internal/parse/testdata/declarations/My/OkOnly.tpm
- internal/parse/testdata/declarations/My/UntypedVars.tpm
- internal/parse/testdata/held_to_perl.pl
- internal/parse/typedour_test.go
- internal/parse/use.go
- internal/perltest/perltest.go
- internal/types/declared.go
- internal/types/select.go
- internal/types/select_test.go
- test/e2e/psc_check_test.go
- test/e2e/testdata/check/declared_fit.pl
- test/e2e/testdata/check/declared_mismatch.pl
- test/e2e/testdata/check/declared_single.pl
- test/e2e/testdata/check/undecided_call.pl
- test/e2e/testdata/check/undecided_declared.pl
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one
- [ ] Convert every .tpm to :export and :export_ok, then refuse export lists in a .tpm (01a14683-8375-7810-8a75-b772680603c6)

## Context

- paths: `internal/parse/heldtoperl_test.go` (new), `internal/parse/testdata/held_to_perl.pl` (new), `internal/infer/returntypes_test.go` (new), `internal/infer/testdata/oracle/` (new: per-module call cases), `internal/infer/testdata/observe_types.pl` (reused), `internal/parse/declarations/Carp.tpm`
- commands: `go test ./internal/parse/ -run '^TestCarpPrototypes$|^TestCarpExports$' -count=1 -v`, `go test ./internal/infer/ -run '^TestCarpReturnTypes$|^TestReturnOracleRejectsAnExcludedValue$' -count=1 -v`
- entrypoints: `internal/infer/precision_test.go` (`observedToLattice`, how it drives `observe_types.pl`), `internal/perltest` (from the baseline issue), RFC 0001 "Core-module libraries", "Held to perl."

Build the three perl-held checks every declaration batch uses, with Carp.tpm (already shipped) as their first subject. Each batch issue then adds three named tests, `Test<Batch>Prototypes`, `Test<Batch>Exports`, `Test<Batch>ReturnTypes`, that call these helpers with its module list.

1. **Prototypes.** For every sub the `.tpm` declares, perl's `prototype(\&M::f)` equals the declared prototype, or both are undef.
2. **Exports, measured as installs (ruling 1).** The test compares the declaration's `:export`/`:export_ok` marks against what perl's `use` actually installs in a caller, never against the module's raw `@EXPORT` lists. In a fresh package per case, perl runs a bare `use M`, then `use M qw(name)` for each `:export_ok` name, then `use M ':tag'` for each tag a declaration names; a sub counts as installed when `exists &Caller::name` (not `defined`: POSIX's AUTOLOAD subs are not defined until called; measured 2026-10-16, a bare `use POSIX` installs 586 subs that `exists` sees), and a variable when the caller's glob slot is the module's (`\%Caller::Config == \%Config::Config`). The bare install must equal the `:export` names; each `:export_ok` name must be installed when asked for and not by a bare `use`; each tag must install exactly the names declared with it. A module whose own `import` installs names its lists lack (Locale::Maketext::Simple installs `loc` and `loc_lang` with empty lists, measured) is held the same way. Measured 2026-10-16: `use Carp qw(verbose)` installs nothing (`verbose` is an `@EXPORT_FAIL` flag), so Carp.tpm marks no `verbose`.
   - **`&name` entries.** An export list may spell a sub with a leading `&` (Getopt::Long has `&GetOptions` and five more, measured 2026-10-16). It is the sub `name`: perl installs `name`, the `.tpm` declares `sub name :export ...`, and no declared name carries `&`.
   - **Locating modules.** The helper finds a module through perl itself (`require M; $INC{'M.pm'}`), never by joining privlib with a path: measured 2026-10-16, 23 of the top 50 load from archlib, not privlib.
3. **Return types.** An oracle has perl call each callable sub with the argument cases in `testdata/oracle/<Module>.pl`, maps each returned value to a lattice type with `observe_types.pl`'s membership tests (reuse it; do not write a second oracle), and fails when an observed value is not a member of the declared return type (WRONG = 0). The declared return type is read from the parsed `.tpm`'s signatures, so this needs no change to infer. It runs perl with its working directory, `HOME` and `TMPDIR` in `t.TempDir()`. A sub unsafe there -- network, fork/exec, signals, `_exit`, file mutation outside scratch, blocking on a terminal -- is not called; it is held to checks 1 and 2 only, and is listed by name, with the reason, in a map literal in `returntypes_test.go` so the exemption is visible in review. A never-returning sub (`None`) is called inside `eval` and must die.

All three skip where perl 5.42.0 is absent and assume no POSIX paths.

**Ruling 1 (perigrin, 2026-10-16): no export lists in a `.tpm`.** "We don't need the @EXPORT in the .tpm since we have :export and :export_ok." On names a module's own `import` installs: "we should figure out as best we can what is exported (loc, loc_lang, etc.) and mark them as :export or :export_ok accordingly."

## Steps

- [ ] Write failing test `TestReturnOracleRejectsAnExcludedValue`: a test-declarations `.tpm` declaring a sub `Int` whose perl implementation returns a string fails the oracle helper
- [ ] Run to verify failure: `go test ./internal/infer/ -run '^TestReturnOracleRejectsAnExcludedValue$' -count=1 -v`
- [ ] Implement the oracle helper; run to verify pass; commit
- [ ] Write the export helper's negative tests below (missing name, extra name, `&name`, install-not-lists, `exists` not `defined`); run to verify failure; implement; run to verify pass; commit
- [ ] Write `TestCarpPrototypes` and `TestCarpExports`; run; fix Carp.tpm where perl disagrees, recording the measurement in its comments; commit
- [ ] Write `TestCarpReturnTypes` with call cases for carp, cluck, longmess, shortmess, croak, confess; run; fix Carp.tpm where perl disagrees; commit
- [ ] Run the m1 gates named in the final-gate issue, with the commands it gives them; a ratchet may only be raised. A drop is fixed in the .tpm, or the batch stops and asks perigrin. `make test`; commit

### Review Findings

Tier-2 gate, 2026-10-17, on c707b835 (fresh reviewer subagent; lenses applied inline; all 50 ranked modules measured on perl 5.42). The installed rule (defined, or exists in a package defining AUTOLOAD) is right: confirmed across the top 50, including Encode's and Test::More's AUTOLOAD-less stubs. VERDICT: REOPEN.

1. (in scope, 90) A list return is not checked against its element type: a sub declared `List[Int]` returning `('a', [1], undef)` passes (`internal/infer/returntypes_test.go:112-114,128-139`; `internal/parse/decl.go:153,282` discard the element type). Also `callContext` calls a sub in list context whenever any candidate returns a list, so a context-selected multi's Scalar candidate is never called. About 50 top-50 exports return lists, so WRONG = 0 would be vacuous for them against goal (4). Fix: add ReturnElement to types.Signature (as Param.Element), set it from typeExpr; in list context check each element; call each candidate in the context it answers; negative test: a `List[Int]` sub returning a Str fails.
2. (in scope, 90) The export helper fails an :export name when `use M qw(name)` dies (Test::More's import reads names as a plan: "dies plan() doesn't understand ok"), and the message wrongly says :export_ok. Only :export_ok names must install when asked for. Fix: add `!slices.Contains(facts.defaults, list)` to that case; fixture test for an import that dies on names.
3. (in scope, 85) A sub that calls `exit` silently ends the oracle run: END still prints the JSON, exit 0, later subs never called. Fix: `BEGIN { *CORE::GLOBAL::exit = sub { die "exit\n" } }` before requiring the module, and a completion sentinel after the loop that Go requires.
4. (in scope, 80) Gaps that pass a sub unchecked: (a) a case entry `a => []` passes; (b) a `None` sub perl does not define passes, because "Undefined subroutine" counts as the expected death. Fix: `@$calls or die "$name has no call case"`; before a None call require the sub be callable by the same rule as `callable`.
5. (in scope, 75) The prototype helper's "missing" branch (`held_to_perl.pl:27`, `heldtoperl_test.go:123-125`) is untested: treating missing as none fails nothing. Fix: TestPrototypeHelperRejectsASubPerlLacks.
6. (in scope, 70) Both helpers pass the caller's PERL5LIB/PERL5OPT to perl (`heldtoperl_test.go:68`, `returntypes_test.go:97`): with PERL5LIB at a fake Carp.pm, TestCarpExports measured the fake. Fix: strip PERL5LIB, PERLLIB, PERL5OPT and PERL_LOCAL_LIB_ROOT from both environments; a test.
7. (folded in by the orchestrator for goal (4), 70) Tags the .tpm omits are never checked: a .tpm leaving out `%EXPORT_TAGS{gh}` passes, so `use POSIX ':sys_wait_h'` could stay undeclared. Goal (3) requires every %EXPORT_TAGS name declared. Fix: the driver also asks `":$_"` for each %EXPORT_TAGS key and reports any tag that installs names but is not declared.
8. (folded in by the orchestrator, 70) A sub printing to STDOUT corrupts the oracle JSON (loud, not silent), so Test::More's ok/is can't be checked. Fix: dup STDOUT for the JSON and point STDOUT at STDERR while calls run; a test.

### Review Findings (pass 2)

Tier-2 gate pass 2, 2026-10-17, on ad6541f6 (fresh reviewer subagent; one mutation per pass-1 finding: F1, F2, F4-F8 each caught; 18/18 criteria; precision WRONG 0; lib baseline holds; List::Util and POSIX run clean through the driver). VERDICT: REOPEN.

A. (in scope, 90) `internal/infer/testdata/return_oracle.pl:28-30`: the STDOUT dup/redirect comes after `require $file`, so a module that dups STDOUT at load (Test::Builder) still writes into the report ("ok 1 - name / # hi / complete / [...]"), and Test::Builder's END sets exit 254, so the oracle reports "did not run". Fix: redirect before `require $file`; an END defined before the observer and module are required runs last and sets `$? = 0 if $complete`; the fixture's module dups STDOUT at load.
B. (in scope, 80; regressions from 652f77ad, both passed at ad6541f6^1) (1) a sub declared `Void` with body `return;` now fails "in scalar context returned Undef" because callContext maps VoidCtx to scalar; (2) `multi sub f (Int $x) Int; multi sub f () None;` runs every case in both none and scalar context, the None-meant case dies outside eval, and the driver aborts. Fix: call a Void candidate in list context (or admit Undef for Void in scalar); for a sub with a None candidate call every case inside eval in every context and admit a death only through the None candidate; a test each.
C. (in scope, 75) TestReturnOracleFailsOnASubThatExits pins neither half of finding 3: removing the exit override, the sentinel check, or both leaves it passing (the STDOUT redirect is what catches it). Fix: a None sub that calls `exit` must pass (the override turns it into a death); assert the sentinel's message for the CORE::exit subtest.
D. (folded in by the orchestrator; 85; blocks the OO-only batch dd27) The oracle cannot call inherited class methods: File::Spec gives "perl defines no sub catfile"; the prototype mode uses `$module->can($name)` but the oracle and installs mode use the exists/defined/AUTOLOAD rule. Fix: for a declaration of a method (invocant), resolve through `can` and call with the class as invocant; a test with an inherited class method.

### Review Findings (pass 3)

Tier-2 gate pass 3, 2026-10-17, on ae366164 (fresh reviewer subagent; one mutation per fix A-D, each caught; 18/18 criteria; precision WRONG 0; lib baseline holds; driver run against Scalar::Util, List::Util, POSIX, File::Spec methods, Cwd, Data::Dumper, Test::More and Getopt::Long with perl 5.42: all complete, valid JSON, exit 0). VERDICT: REOPEN (third and final reopen allowed).

1. (in scope, 75; introduced by f575da65) `internal/infer/testdata/return_oracle.pl:57-67`, `internal/infer/returntypes_test.go:117-142`: a call meant for a None candidate that returns an empty list in list context passes silently, since an empty @values records nothing and there is no death. `multi sub f (Int $x) List[Int]; multi sub f () None;` with `f()` returning `()` gives 0 failures; likewise with a Void candidate, and any list-context call no candidate takes that returns `()`. Before the fix None candidates ran in scalar context, so an empty return was seen as Undef and failed. Fix: after a call returns, record a Returned marker (`push @OBSERVED, [$i, $key, 'Returned']`); in Go fail a Returned marker when no non-None candidate takes that call form and argument count in that context (`takes`); a test with the List[Int]/None multi returning `()`, failing first.

## Acceptance Criteria

### Positive Scenarios
- [ ] the oracle fails on a value the declared return type excludes (`go test ./internal/infer/ -run '^TestReturnOracleRejectsAnExcludedValue$' -count=1 -v | grep -q '^--- PASS: TestReturnOracleRejectsAnExcludedValue'`)
- [ ] Carp's declared prototypes equal perl's (`go test ./internal/parse/ -run '^TestCarpPrototypes$' -count=1 -v | grep -q '^--- PASS: TestCarpPrototypes'`)
- [ ] Carp's `:export`/`:export_ok` marks equal what `use Carp` installs in a caller, bare and by name (`go test ./internal/parse/ -run '^TestCarpExports$' -count=1 -v | grep -q '^--- PASS: TestCarpExports'`)
- [ ] Carp's declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestCarpReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestCarpReturnTypes'`)

### Negative Scenarios
- [ ] the prototype helper fails a declaration whose prototype disagrees with perl's `prototype()` (`go test ./internal/parse/ -run '^TestPrototypeHelperRejectsADisagreement$' -count=1 -v | grep -q '^--- PASS: TestPrototypeHelperRejectsADisagreement'`)
- [ ] the prototype helper distinguishes an undef prototype from an empty `()` one (`go test ./internal/parse/ -run '^TestPrototypeHelperDistinguishesUndefFromEmpty$' -count=1 -v | grep -q '^--- PASS: TestPrototypeHelperDistinguishesUndefFromEmpty'`)
- [ ] the export helper fails a `.tpm` that does not mark a name `use M qw(name)` installs (`go test ./internal/parse/ -run '^TestExportHelperRejectsAMissingName$' -count=1 -v | grep -q '^--- PASS: TestExportHelperRejectsAMissingName'`)
- [ ] the export helper fails a `.tpm` marking `:export` a name a bare `use` does not install (`go test ./internal/parse/ -run '^TestExportHelperRejectsAnExtraName$' -count=1 -v | grep -q '^--- PASS: TestExportHelperRejectsAnExtraName'`)
- [ ] the export helper measures installs, not lists: a fixture module whose own `import` installs a name absent from its (empty) lists fails a `.tpm` that does not mark it `:export` (`go test ./internal/parse/ -run '^TestExportHelperMeasuresInstallsNotLists$' -count=1 -v | grep -q '^--- PASS: TestExportHelperMeasuresInstallsNotLists'`)
- [ ] the export helper counts an installed AUTOLOAD sub (exists, not defined) as installed (`go test ./internal/parse/ -run '^TestExportHelperCountsAutoloadedSubs$' -count=1 -v | grep -q '^--- PASS: TestExportHelperCountsAutoloadedSubs'`)
- [ ] an `&name` entry in a fixture module's export list is matched by `sub name :export` and is not reported as a second, missing name (`go test ./internal/parse/ -run '^TestExportHelperTreatsAmpersandAsSubName$' -count=1 -v | grep -q '^--- PASS: TestExportHelperTreatsAmpersandAsSubName'`)
- [ ] a declared module perl cannot `require` fails by name rather than passing vacuously (`go test ./internal/parse/ -run '^TestHeldToPerlFailsOnAModulePerlCannotLoad$' -count=1 -v | grep -q '^--- PASS: TestHeldToPerlFailsOnAModulePerlCannotLoad'`)
- [ ] a module that loads from archlib, not privlib (e.g. POSIX), is found through `%INC` (`go test ./internal/parse/ -run '^TestHeldToPerlFindsArchlibModules$' -count=1 -v | grep -q '^--- PASS: TestHeldToPerlFindsArchlibModules'`)
- [ ] the oracle never calls a sub in the exemption map: a fixture exempt sub that would write a marker file leaves none (`go test ./internal/infer/ -run '^TestReturnOracleNeverCallsAnExemptSub$' -count=1 -v | grep -q '^--- PASS: TestReturnOracleNeverCallsAnExemptSub'`)
- [ ] every exemption-map entry names a non-empty reason (`go test ./internal/infer/ -run '^TestReturnOracleExemptionsNameAReason$' -count=1 -v | grep -q '^--- PASS: TestReturnOracleExemptionsNameAReason'`)
- [ ] a sub declared `None` that returns under the oracle fails it (`go test ./internal/infer/ -run '^TestReturnOracleRejectsAReturningNoneSub$' -count=1 -v | grep -q '^--- PASS: TestReturnOracleRejectsAReturningNoneSub'`)
- [ ] a declared sub with neither an oracle call case nor an exemption fails the oracle, so no sub goes unchecked silently (`go test ./internal/infer/ -run '^TestReturnOracleRejectsAnUncoveredSub$' -count=1 -v | grep -q '^--- PASS: TestReturnOracleRejectsAnUncoveredSub'`)
- [ ] the oracle runs perl with cwd, `HOME` and `TMPDIR` in the test's temp dir: a fixture sub writing under `$ENV{HOME}` writes there (`go test ./internal/infer/ -run '^TestReturnOracleRunsInScratchDir$' -count=1 -v | grep -q '^--- PASS: TestReturnOracleRunsInScratchDir'`)
