---
title: Audit every .tpm signature against its implementation and documentation, and hold parameter types to perl
state: in-progress
urgency: normal
milestone: core-module-types
created: 2026-10-20T18:55:00.070093546Z
updated: 2026-10-20T18:55:34.658232436Z
sessions:
- start_sha: cf08c0acdb986b7a59c463c2ed994e0721307294
  end_sha: ""
  commits: 0
  started_at: 2026-10-20T18:55:34.658232436Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T18:55:34.658232436Z
---

perigrin, 2026-10-21: "We should run an audit of all the signatures against the implementation and the documentation to make sure that the tpm files are accurate." Prompted by Carp.tpm: croak and confess take `List[Str|Boolean|DualVar|Ref]` but carp, cluck, longmess and shortmess take `List[Str]`, though all six join `@_` the same way in Carp.pm (~588) and accept the same arguments (perl 5.42: undef warns for both; a Boolean, a DualVar and a reference are all legal; croak($obj) throws the object). Only their returns differ. No gate caught it: the oracle holds returns, the hold tests hold prototypes and installs, and nothing holds parameter types.

## Context
- Every .tpm under internal/parse/declarations/ except CORE.tpm (CORE is held by its own tests).
- The implementation is the module's source under perl 5.42's privlib/archlib; the documentation is its POD (`perldoc -T Module`). perl 5.42 is /home/perigrin/.local/bin/perl and is the judge where source and POD disagree.

## Steps
- [ ] Audit each .tpm's every sub and variable against implementation and documentation: parameter types (what perl accepts without dying), optional vs required, arity, slurpy, context-dependent returns, Undef in returns, and documented behaviour the declaration misses. Report-only audit first, findings grouped by module.
- [ ] Add a parameter oracle: every return-oracle call case's arguments must be accepted by the declared parameters of the candidate that case selects (so a declaration that rejects a call perl accepts fails). Cases that exercise a documented argument shape the .tpm omits are added.
- [ ] Fix every audit finding, each measured with perl; carp/cluck/longmess/shortmess take `List[Str|Boolean|DualVar|Ref]`.
- [ ] psc check over perl 5.42's lib adds no new line; precision holds.

## Acceptance criteria

### Positive Scenarios
- [ ] Every oracle case's arguments fit its declared candidate (`go test ./internal/infer/ -run '^TestOracleCasesFitDeclaredParameters$' -count=1 -v | grep -q '^--- PASS: TestOracleCasesFitDeclaredParameters'`)
- [ ] carp accepts a reference and a boolean (`go test ./internal/infer/ -run '^TestCarpAcceptsAnyDefinedMessage$' -count=1 -v | grep -q '^--- PASS: TestCarpAcceptsAnyDefinedMessage'`)
- [ ] psc check over perl's lib adds no unlisted line (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] Precision holds with WRONG 0 (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] A declaration that rejects an argument perl accepts fails the parameter oracle (`go test ./internal/infer/ -run '^TestParameterOracleRejectsANarrowParam$' -count=1 -v | grep -q '^--- PASS: TestParameterOracleRejectsANarrowParam'`)
