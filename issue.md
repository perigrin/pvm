---
title: "Call-site selection fixes: an empty use-list imports nothing, the coercion retry keeps arity, an Any argument does not hide arity"
state: pending
urgency: normal
milestone: core-module-types
created: 2026-10-20T19:18:42.346874019Z
updated: 2026-10-20T19:18:42.346874019Z
---

crochet:review of PR #488, round 1 (2026-10-21). Quality I2 internal/parse/use.go:378 (use Test::More qw(); installs defaults; perl imports nothing), I5 internal/types/select.go:121-139 + infer/calls.go:111 (coercion retry consults wrong-arity candidates; sockaddr_in("abc","a") fails silently), I8 calls.go:98-103 (an Any- or List-typed argument counts as flattening; measure which flatten in perl), S1 internal/parse/derive.go:172-182 (candidatesDifferInArity compares prototype string length), S7 internal/types/types.go:368 (a Code union rejects CodeRef).

Parameter rule (RFC 0001 "What a parameter type means", perigrin 2026-10-02): a parameter type is the type the operation coerces its argument to. A parameter is wrong when its type is not what the operation coerces to or preserves (a lone reference carp/warn/croak pass through; NaN/Inf a C double keeps), or it admits a value perl has no coercion for and dies on. A finding that only shows perl coercing a value (sum("abc"), encode with a ref) is withdrawn, not fixed. Every fix is measured with /home/perigrin/.local/bin/perl; never bend a type to quiet the lib baseline.

## Steps
- [ ] Failing test per finding from the verifier's probe
- [ ] Measure each against perl

## Acceptance criteria

### Positive Scenarios
- [ ] TestEmptyUseListImportsNothing (`go test ./internal/parse/ -run '^TestEmptyUseListImportsNothing$' -count=1 -v | grep -q '^--- PASS: TestEmptyUseListImportsNothing'`)
- [ ] TestCoercionRetryKeepsArity (`go test ./internal/types/ -run '^TestCoercionRetryKeepsArity$' -count=1 -v | grep -q '^--- PASS: TestCoercionRetryKeepsArity'`)
- [ ] TestCandidatesDifferInArityComparesArity (`go test ./internal/parse/ -run '^TestCandidatesDifferInArityComparesArity$' -count=1 -v | grep -q '^--- PASS: TestCandidatesDifferInArityComparesArity'`)
- [ ] TestCodeUnionAcceptsCodeRef (`go test ./internal/types/ -run '^TestCodeUnionAcceptsCodeRef$' -count=1 -v | grep -q '^--- PASS: TestCodeUnionAcceptsCodeRef'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestAnyScalarArgumentKeepsArity (`go test ./internal/infer/ -run '^TestAnyScalarArgumentKeepsArity$' -count=1 -v | grep -q '^--- PASS: TestAnyScalarArgumentKeepsArity'`)
