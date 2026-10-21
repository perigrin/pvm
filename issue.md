---
title: "Over-engineering cuts from the review: shared perl-driver helpers, Signature.Arity, one scope-chain update, maps.Clone"
state: in-progress
urgency: normal
milestone: core-module-types
created: 2026-10-20T19:18:42.724502357Z
updated: 2026-10-21T03:39:32.787168188Z
sessions:
- start_sha: 22583a402ae797ff8a0a591d3d5c4f0763744131
  end_sha: ""
  commits: 0
  started_at: 2026-10-21T03:39:32.787168188Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-21T03:39:32.787168188Z
---

crochet:review of PR #488, round 1 (2026-10-21). ponytail lens (about -150 lines): return_oracle.pl:149-154 and held_to_perl.pl:145-150 duplicate callable; both drivers carry the same four-place sub lookup (one find_sub in a shared held_lib.pl, one rule for found); held_to_perl.pl installs/skeleton modes build the same exportable list and use-string; derive.go:151-165 arityRange re-implements Signature.accepts (export Signature.Arity); symbols.go UpdatePossibleUndef is a third scope-chain walk with an unread bool (one update(name, f)); calls.go markPossibleUndef/markedPossibleUndef have one caller each; flattenedSelect and selectAnyContext share a tail (types.JoinSelections); flow.go elsifPre copies a map by hand (maps.Clone); isElementVarName is a one-caller wrapper; DeclaredSignatures/DeclaredVariables are two test-only wrappers; Signature.ReturnElement is read only by tests (keep only if a consumer is planned). Do this after the other fix-round issues land, since it touches their files.

Parameter rule (RFC 0001 "What a parameter type means", perigrin 2026-10-02): a parameter type is the type the operation coerces its argument to. A parameter is wrong when its type is not what the operation coerces to or preserves (a lone reference carp/warn/croak pass through; NaN/Inf a C double keeps), or it admits a value perl has no coercion for and dies on. A finding that only shows perl coercing a value (sum("abc"), encode with a ref) is withdrawn, not fixed. Every fix is measured with /home/perigrin/.local/bin/perl; never bend a type to quiet the lib baseline.

## Steps
- [ ] One cut per commit; every hold test and gate unchanged

## Acceptance criteria

### Positive Scenarios
- [ ] One shared perl-driver helper (`test -f internal/parse/testdata/held_lib.pl`)
- [ ] arityRange is gone (`! grep -q 'func arityRange' internal/parse/derive.go`)
- [ ] TestXSUtilitiesPrototypes (`go test ./internal/parse/ -run '^TestXSUtilitiesPrototypes$' -count=1 -v | grep -q '^--- PASS: TestXSUtilitiesPrototypes'`)
- [ ] TestXSUtilitiesReturnTypes (`go test ./internal/infer/ -run '^TestXSUtilitiesReturnTypes$' -count=1 -v | grep -q '^--- PASS: TestXSUtilitiesReturnTypes'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] isElementVarName is inlined (`! grep -q 'func isElementVarName' internal/infer/flow.go`)
