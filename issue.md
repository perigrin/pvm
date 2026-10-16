---
title: "Hold Carp.tpm to perl: prototypes, export lists and a return-type oracle"
state: pending
urgency: normal
milestone: core-module-types
blocked_by:
- 01a1464e-d912-7419-8b20-a8e24029b3ec
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
updated: 2026-10-16T20:11:54.011591937Z
---

## Prerequisites

- [ ] Commit psc check's diagnostics over perl 5.42's lib as a baseline, and fail on any new one

## Context

- paths: `internal/parse/heldtoperl_test.go` (new), `internal/parse/testdata/held_to_perl.pl` (new), `internal/infer/returntypes_test.go` (new), `internal/infer/testdata/oracle/` (new: per-module call cases), `internal/infer/testdata/observe_types.pl` (reused), `internal/parse/declarations/Carp.tpm`
- commands: `go test ./internal/parse/ -run '^TestCarpPrototypes$|^TestCarpExports$' -count=1 -v`, `go test ./internal/infer/ -run '^TestCarpReturnTypes$|^TestReturnOracleRejectsAnExcludedValue$' -count=1 -v`
- entrypoints: `internal/infer/precision_test.go` (`observedToLattice`, how it drives `observe_types.pl`), `internal/perltest` (from the baseline issue), RFC 0001 "Core-module libraries", "Held to perl."

Build the three perl-held checks every declaration batch uses, with Carp.tpm (already shipped) as their first subject. Each batch issue then adds three named tests, `Test<Batch>Prototypes`, `Test<Batch>Exports`, `Test<Batch>ReturnTypes`, that call these helpers with its module list.

1. **Prototypes.** For every sub the `.tpm` declares, perl's `prototype(\&M::f)` equals the declared prototype, or both are undef.
2. **Export lists.** The module's `@EXPORT`, `@EXPORT_OK` and `%EXPORT_TAGS` equal the declaration's, read AFTER `require M; M->import` in a throwaway package: POSIX builds its lists in `import` (measured 2026-10-16 on 5.42.0: 0/0 after `require`, 587 `@EXPORT` / 179 `@EXPORT_OK` / 35 tags after `import`).
3. **Return types.** An oracle has perl call each callable sub with the argument cases in `testdata/oracle/<Module>.pl`, maps each returned value to a lattice type with `observe_types.pl`'s membership tests (reuse it; do not write a second oracle), and fails when an observed value is not a member of the declared return type (WRONG = 0). The declared return type is read from the parsed `.tpm`'s signatures, so this needs no change to infer. It runs perl with its working directory, `HOME` and `TMPDIR` in `t.TempDir()`. A sub unsafe there -- network, fork/exec, signals, `_exit`, file mutation outside scratch, blocking on a terminal -- is not called; it is held to checks 1 and 2 only, and is listed by name, with the reason, in a map literal in `returntypes_test.go` so the exemption is visible in review. A never-returning sub (`None`) is called inside `eval` and must die.

All three skip where perl 5.42.0 is absent and assume no POSIX paths.

## Steps

- [ ] Write failing test `TestReturnOracleRejectsAnExcludedValue`: a test-declarations `.tpm` declaring a sub `Int` whose perl implementation returns a string fails the oracle helper
- [ ] Run to verify failure: `go test ./internal/infer/ -run '^TestReturnOracleRejectsAnExcludedValue$' -count=1 -v`
- [ ] Implement the oracle helper; run to verify pass; commit
- [ ] Write `TestCarpPrototypes` and `TestCarpExports`; run; fix Carp.tpm where perl disagrees, recording the measurement in its comments; commit
- [ ] Write `TestCarpReturnTypes` with call cases for carp, cluck, longmess, shortmess, croak, confess; run; fix Carp.tpm where perl disagrees; commit
- [ ] Run the m1 gates named in the final-gate issue; `make test`; commit

## Acceptance Criteria

### Positive Scenarios
- [ ] the oracle fails on a value the declared return type excludes (`go test ./internal/infer/ -run '^TestReturnOracleRejectsAnExcludedValue$' -count=1 -v | grep -q '^--- PASS: TestReturnOracleRejectsAnExcludedValue'`)
- [ ] Carp's declared prototypes equal perl's (`go test ./internal/parse/ -run '^TestCarpPrototypes$' -count=1 -v | grep -q '^--- PASS: TestCarpPrototypes'`)
- [ ] Carp's declared export lists equal perl's (`go test ./internal/parse/ -run '^TestCarpExports$' -count=1 -v | grep -q '^--- PASS: TestCarpExports'`)
- [ ] Carp's declared return types hold for every value perl returns (`go test ./internal/infer/ -run '^TestCarpReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestCarpReturnTypes'`)
