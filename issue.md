---
title: "Round-3 call-check regressions: only list-returning forms flatten, every list form flattens, and a Cwd::chdir reference argument is no error"
state: done
urgency: normal
milestone: core-module-types
created: 2026-10-21T20:10:34.75896565Z
updated: 2026-10-21T23:44:44.167378157Z
sessions:
- start_sha: 692a560c81fe680ed09244599f0e1bd216c5a075
  end_sha: 7b17f53d429ecda0b5ad2849427d728a982359d4
  commits: 5
  started_at: 2026-10-21T20:10:59.916997421Z
  ended_at: 2026-10-21T23:44:44.167378157Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-21T20:10:59.916997421Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-21T23:44:44.167378157Z
observed_paths:
- docs/rfcs/0001-type-libraries.md
- internal/infer/calls.go
- internal/infer/calls_test.go
- internal/infer/declared_test.go
- internal/infer/dualvar_test.go
- internal/infer/flow.go
- internal/infer/ops.go
- internal/infer/precision_test.go
- internal/infer/returntypes_test.go
- internal/infer/testdata/observe_types.pl
- internal/infer/testdata/oracle/Scalar-Util.pl
- internal/infer/warnpossibleundef_test.go
- internal/parse/declarations.go
- internal/parse/declarations/Cwd.tpm
- internal/parse/declarations/DynaLoader.tpm
- internal/parse/declarations/ExtUtils/Manifest.tpm
- internal/parse/declarations/File/Basename.tpm
- internal/parse/declarations/File/stat.tpm
- internal/parse/declarations/FileHandle.tpm
- internal/parse/declarations/IO/Socket.tpm
- internal/parse/declarations/IO/Socket/IP.tpm
- internal/parse/declarations/Moose.tpm
- internal/parse/declarations/POSIX.tpm
- internal/parse/declarations/Scalar/Util.tpm
- internal/parse/declarations/Socket.tpm
- internal/parse/declarations/Symbol.tpm
- internal/parse/declarations/Time/HiRes.tpm
- internal/parse/exportattr_test.go
- internal/parse/heldtoperl_test.go
- internal/parse/issue1959_test.go
- internal/parse/pin_test.go
- internal/parse/use.go
- internal/perltest/fixture.go
- internal/perltest/perl542_test.go
- internal/perltest/perltest.go
- internal/perltest/staleexempt.go
- internal/perltest/staleexempt_test.go
- internal/types/declared.go
- internal/types/select.go
- internal/types/select_test.go
- test/e2e/psc_check_test.go
- test/e2e/testdata/check/dualvar.pl
---

crochet:review round 3 of PR #488 at 692a560c (2026-10-21). Quality report: scratchpad/review3-core-module-types-paad.md; probes scratchpad/r3probes/ (F-<lens>.md per specialist); rebuild head/r2(2db182e6)/cf(cf08c0ac)/pu psc as the report says. perigrin, 2026-10-21: round 4 runs with Opus workers, and every fix gets a mandatory independent review that runs pu, round 2 and head against perl 5.42 before it merges. A finding pu or round 2 gets right is a regression: fix it without trading it for a new one -- every probe in r2probes/ and r3probes/ that pu or r2 gets right must still be right.

Quality V3 (calls.go:167-179 flattensArgument counts every call as may-flatten, switching off the arity check and every argument check for the whole call: `fmod(length("abc"))`, `fmod(undef, abs(2))` silent on head, r2 and cf report, perl dies or warns), V4 (the predicate misses list forms -- new: a ternary of aggregates, `(0) x 6`, `7..8`; older vs pu: method calls, `(7,3)`, qw, list assignment -- default-on false arity errors), V19 (S17's checkArg path silent under --warn-possible-undef), V13 (Cwd.tpm:32 makes a reference argument a default error -- a new false positive), ponytail round 3 (calls.go:270,:273 two dead checks in possibleMismatch and their stale comment).

## Steps
- [ ] Decide flattening per expression: a call flattens only when its declared or inferred return is a list in list context; aggregates, slices, ranges, repetition of a list, qw and list assignment flatten
- [ ] Run every r2probes/ and r3probes/ call probe on pu, r2, head and this branch against perl

## Acceptance criteria

### Positive Scenarios
- [ ] TestScalarCallArgumentKeepsArity (`go test ./internal/infer/ -run '^TestScalarCallArgumentKeepsArity$' -count=1 -v | grep -q '^--- PASS: TestScalarCallArgumentKeepsArity'`)
- [ ] TestListFormsFlatten (`go test ./internal/infer/ -run '^TestListFormsFlatten$' -count=1 -v | grep -q '^--- PASS: TestListFormsFlatten'`)
- [ ] TestCheckArgPossibleUndefUnderFlag (`go test ./internal/infer/ -run '^TestCheckArgPossibleUndefUnderFlag$' -count=1 -v | grep -q '^--- PASS: TestCheckArgPossibleUndefUnderFlag'`)
- [ ] TestCwdChdirRefArgumentNoError (`go test ./internal/infer/ -run '^TestCwdChdirRefArgumentNoError$' -count=1 -v | grep -q '^--- PASS: TestCwdChdirRefArgumentNoError'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestScalarCallArgumentStillChecked (`go test ./internal/infer/ -run '^TestScalarCallArgumentStillChecked$' -count=1 -v | grep -q '^--- PASS: TestScalarCallArgumentStillChecked'`)
