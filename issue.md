---
title: "Call-site selection fixes: an empty use-list imports nothing, the coercion retry keeps arity, an Any argument does not hide arity"
state: done
urgency: normal
milestone: core-module-types
created: 2026-10-20T19:18:42.346874019Z
updated: 2026-10-20T21:20:50.837679962Z
sessions:
- start_sha: dee5725d764cf502cf58ded7dcdf4778cb5781d1
  end_sha: c09fd9d5ebb069ebd7b7eef335e030ff4675cd59
  commits: 3
  started_at: 2026-10-20T20:42:14.60089025Z
  ended_at: 2026-10-20T21:20:50.837679962Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-20T20:42:14.60089025Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-20T21:20:50.837679962Z
observed_paths:
- internal/infer/calls.go
- internal/infer/declared_test.go
- internal/infer/returntypes_test.go
- internal/infer/testdata/return_oracle.pl
- internal/parse/decl.go
- internal/parse/declarations.go
- internal/parse/derive.go
- internal/parse/derive_test.go
- internal/parse/exportattr_test.go
- internal/parse/heldtoperl_test.go
- internal/parse/ranking_test.go
- internal/parse/typedour_test.go
- internal/parse/use.go
- internal/perltest/perl542_test.go
- internal/perltest/referencebuild.go
- internal/perltest/referencebuild_test.go
- internal/types/select.go
- internal/types/select_test.go
- internal/types/types.go
- internal/types/types_test.go
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
