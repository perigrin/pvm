---
title: Audit every .tpm signature against its implementation and documentation, and hold parameter types to perl
state: in-progress
urgency: normal
milestone: core-module-types
created: 2026-10-20T18:55:00.070093546Z
updated: 2026-10-20T19:19:04.364837287Z
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

**Correction, 2026-10-21 (orchestrator):** the carp example below misreads the rule. RFC 0001 "What a parameter type means" (perigrin, 2026-10-02): a parameter type is the type the operation coerces its argument to. carp stringifies its message, so `List[Str]` is right and carp($ref) is a lossy coercion (silent by default, reported under --warn-lossy-coercion); croak passes a lone ref through as the exception, so its Ref is right. The audit holds each parameter to the type the operation coerces to or preserves, and flags a type admitting a value perl has no coercion for (dies). The carp criterion and the "carp takes List[Str|Boolean|DualVar|Ref]" step are withdrawn; the parameter oracle checks that each case argument fits its candidate or is a coercion the operation performs.

perigrin, 2026-10-21: "We should run an audit of all the signatures against the implementation and the documentation to make sure that the tpm files are accurate." Prompted by Carp.tpm: croak and confess take `List[Str|Boolean|DualVar|Ref]` but carp, cluck, longmess and shortmess take `List[Str]`, though all six join `@_` the same way in Carp.pm (~588) and accept the same arguments (perl 5.42: undef warns for both; a Boolean, a DualVar and a reference are all legal; croak($obj) throws the object). Only their returns differ. No gate caught it: the oracle holds returns, the hold tests hold prototypes and installs, and nothing holds parameter types.

**Measured after the correction, 2026-10-21:** carp({c=>42}) and warn({c=>42}) deliver the reference itself to $SIG{__WARN__}, as croak does to $@; so carp, cluck and warn do take croak`s parameter type after all, under the same rule. Fixed in the group E issue.

## Context
- Every .tpm under internal/parse/declarations/ except CORE.tpm (CORE is held by its own tests).
- The implementation is the module's source under perl 5.42's privlib/archlib; the documentation is its POD (`perldoc -T Module`). perl 5.42 is /home/perigrin/.local/bin/perl and is the judge where source and POD disagree.

## Steps
- [ ] Audit each .tpm's every sub and variable against implementation and documentation: parameter types (what perl accepts without dying), optional vs required, arity, slurpy, context-dependent returns, Undef in returns, and documented behaviour the declaration misses. Report-only audit first, findings grouped by module.
- [ ] Add a parameter oracle: every return-oracle call case's arguments must be accepted by the declared parameters of the candidate that case selects (so a declaration that rejects a call perl accepts fails). Cases that exercise a documented argument shape the .tpm omits are added.
- [x] Audit run 2026-10-21 by five fresh Sonnet auditors (groups A-E): A POSIX 102, B sockets 20, C IO/OO 33, D XS 35 (pre-correction), E pure-Perl 20.
- [ ] Fixes are their own issues: 01a15ac0-6c4a (A), 01a15ac0-6c78 (B), 01a15ac0-6ca2 (C), 01a15ac0-6ccc (D), 01a15ac0-6cfa (E), 01a15ac0-6bfd (import semantics).
- [ ] Add DualVar to the observer (observe_types.pl never reports it; getaddrinfo errors are dualvars).
- [ ] psc check over perl 5.42's lib adds no new line; precision holds.

## Acceptance criteria

### Positive Scenarios
- [ ] Every oracle case's arguments fit its declared candidate (`go test ./internal/infer/ -run '^TestOracleCasesFitDeclaredParameters$' -count=1 -v | grep -q '^--- PASS: TestOracleCasesFitDeclaredParameters'`)
- [ ] psc check over perl's lib adds no unlisted line (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] Precision holds with WRONG 0 (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] A declaration that rejects an argument perl accepts fails the parameter oracle (`go test ./internal/infer/ -run '^TestParameterOracleRejectsANarrowParam$' -count=1 -v | grep -q '^--- PASS: TestParameterOracleRejectsANarrowParam'`)
