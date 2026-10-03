---
title: Check CORE.pmt's derived prototypes against perl, with signatures.go's measured rows
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fd9b-4850-704d-8c14-c15add6194f2
blocks:
- 01a0fd9b-5670-78d0-84be-9b9618e9b081
created: 2026-10-02T17:13:29.229586875Z
updated: 2026-10-03T11:14:17.363797428Z
---

## Prerequisites

- [ ] Parse aliased parameters (\$x, \@a, \%h, \(@list)) and derive \$, \@ and \% from them

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `internal/parse/declarations.go`, `internal/parse/coretypes_test.go` (new), `internal/parse/export_test.go`, `docs/rfcs/0001-type-libraries.md`, `internal/data/help/types.md`
- commands: `go test ./internal/parse/ -count=1`
- commands: full gate `make test` (100% green before every commit; never `--no-verify`)
- perl 5.42 for measuring behaviour: `/home/perigrin/.local/bin/perl` (measure, never assume)
- entrypoints: `internal/parse/declarations/CORE.pmt` (188 `sub NAME :prototype(...);` lines), `coreTable` (`internal/parse/declarations.go:39`), `TestCoreDeclarationsArePerls` (`internal/parse/coretable_test.go:20`) for asking perl via `conformance.PerlPath()`, `internal/types/signatures.go:30-112` (the `builtins` map: 40 typed builtins and their measurement comments)

The harness every CORE.pmt typing batch is held to, and the first lines typed. RFC "Typed Perl, in `.pmt` only": `sub bless :prototype($;$) (Ref $ref, Str $class = __PACKAGE__) Object;`. `TestCoreDerivedPrototypesArePerls` reads every one of CORE.pmt's 188 prototyped lines and holds its prototype -- derived from its types where the line is typed, as written where it is not yet -- to `prototype("CORE::name")`; a derived `@` (no prototype) matches perl's `@` (`reverse`, `die`, `warn`), as the RFC's "No prototype is the default prototype" says the two parse identically. `internal/types/signatures.go` types 40 builtins; 12 of them have no prototype (`delete`, `exists`, `chomp`, `chop`, `split`, `defined`, `print`, `say`, `return`, `map`, `grep`, `sort`), and their rows go to the issues that declare them. Of the 28 with a prototype, `each`, `keys` and `values` are `\[%@]`, which takes the multi-candidate rule of the no-prototype builtins issue, so `each`'s row goes there and `keys`' and `values`' to the array, list and hash batch. This issue types the other 25 (`push`, `pop`, `shift`, `unshift`, `splice`, `length`, `chr`, `ord`, `join`, `sprintf`, `substr`, `ref`, `scalar`, `die`, `warn`, `bless`, `abs`, `int`, `uc`, `lc`, `ucfirst`, `lcfirst`, `index`, `rindex`, `reverse`; measured with `prototype("CORE::name")`) from signatures.go's argument and return types, moving its measurement comments above their CORE.pmt lines. Copy those 25 rows as golden values into the new test -- never compare against signatures.go, which a later issue deletes, and `git-zhi verify` re-runs this issue's criteria on the final tree. Element types the paper has not defined are written as plain `List`. Keep `coreTable()` and `TestCoreDeclarationsArePerls` green. Re-marks no RFC section: the last batch completes the 188.

## Steps

- [ ] Write failing test `TestCoreDerivedPrototypesArePerls` in `internal/parse/coretypes_test.go`: every CORE.pmt line's prototype, derived from its types where it has them, equals perl's (skip only when no perl, as `TestCoreDeclarationsArePerls` does), with one typed line to make it fail first
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreDerivedPrototypesArePerls$' -count=1 -v`
- [ ] Implement the harness; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestCoreDerivedPrototypeCheckCatchesMismatch`: a deliberately wrong line run through the same check (`sub push (List @a, List @l) Int;`) is reported as disagreeing with perl, naming `push`
- [ ] Implement; run test to verify pass; `make test`; commit
- [ ] Write failing test `TestCoreTypesMatchMeasuredSignatures`: the 25 golden rows (arity, argument and return types) equal their CORE.pmt declarations
- [ ] Run test to verify failure: `go test ./internal/parse/ -run '^TestCoreTypesMatchMeasuredSignatures$' -count=1 -v`
- [ ] Type the 25 lines and move signatures.go's measurement comments into CORE.pmt; fix any line that disagrees, measuring which side is right; run test to verify pass; `make test`; commit
- [ ] Doc: `internal/data/help/types.md` says the shipped files "use only the plain-Perl part"; replace that with a sentence saying `CORE.pmt` carries typed declarations, and show its typed `bless` line; in the RFC's "CORE.pmt: the builtin table" section, show typed lines in place of the bare-prototype ones; `go test ./internal/cli/ -run '^TestShowTypesHelp$' -count=1`; commit

### Negative Scenarios

- the derived-prototype comparison is not tautological: a deliberately wrong line run through the same check is reported as disagreeing with perl, naming the builtin (`go test ./internal/parse/ -run '^TestCoreDerivedPrototypeCheckCatchesMismatch$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedPrototypeCheckCatchesMismatch'`)
- a typed line whose types derive a different prototype from the `:prototype(...)` it still carries is reported as a declaration error, not silently resolved in favour of either (`go test ./internal/parse/ -run '^TestCoreTypedLineDisagreeingWithItsPrototype$' -count=1 -v | grep -q '^--- PASS: TestCoreTypedLineDisagreeingWithItsPrototype'`)

## Acceptance criteria

- [ ] every CORE.pmt prototype, derived from its types where typed, is perl's (`go test ./internal/parse/ -run '^TestCoreDerivedPrototypesArePerls$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedPrototypesArePerls'`)
- [ ] CORE.pmt carries the measured types of the prototyped builtins internal/infer types today (`go test ./internal/parse/ -run '^TestCoreTypesMatchMeasuredSignatures$' -count=1 -v | grep -q '^--- PASS: TestCoreTypesMatchMeasuredSignatures'`)
- [ ] the CORE prototype check reports a wrong line (`go test ./internal/parse/ -run '^TestCoreDerivedPrototypeCheckCatchesMismatch$' -count=1 -v | grep -q '^--- PASS: TestCoreDerivedPrototypeCheckCatchesMismatch'`)
- [ ] a typed CORE line that disagrees with its own prototype is an error (`go test ./internal/parse/ -run '^TestCoreTypedLineDisagreeingWithItsPrototype$' -count=1 -v | grep -q '^--- PASS: TestCoreTypedLineDisagreeingWithItsPrototype'`)
- [ ] the help page shows a typed CORE.pmt declaration and no longer calls the shipped files untyped (`sh -c '! grep -q "plain-Perl part" internal/data/help/types.md && grep -qF "sub bless" internal/data/help/types.md'`)
