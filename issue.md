---
title: "Over-engineering cuts from the review: shared perl-driver helpers, Signature.Arity, one scope-chain update, maps.Clone"
state: done
urgency: normal
milestone: core-module-types
created: 2026-10-20T19:18:42.724502357Z
updated: 2026-10-21T16:13:02.447690784Z
sessions:
- start_sha: 22583a402ae797ff8a0a591d3d5c4f0763744131
  end_sha: 2db182e610c0a9eebf6a623550bd13f550d219a4
  commits: 1
  started_at: 2026-10-21T03:39:32.787168188Z
  ended_at: 2026-10-21T04:35:35.94184157Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-21T03:39:32.787168188Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-21T04:35:35.94184157Z
observed_paths:
- internal/infer/flow.go
- internal/infer/infer.go
- internal/infer/narrow_lib_test.go
- internal/infer/symbols.go
- internal/infer/testdata/return_oracle.pl
- internal/infer/vars.go
- internal/parse/derive.go
- internal/parse/testdata/held_lib.pl
- internal/parse/testdata/held_to_perl.pl
- internal/types/select.go
- internal/types/select_test.go
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

## Skipped cuts

Four of this issue's own findings were not cut. 01a15ce6-19ab's round-2 review (2026-10-21) asked that each be recorded here, with its reason, rather than left as silent drift:

- **flattenedSelect and selectAnyContext's shared tail.** The two do not actually share one: one returns the agreed `Selected` candidate, the other always returns `Undecided`. They answer different questions, so unifying their tail would have to throw one answer away.
- **markPossibleUndef/markedPossibleUndef's "one caller each."** markedPossibleUndef has three callers (calls.go's own recursive call, and two in ops.go), not one -- the finding's premise does not hold.
- **DeclaredSignatures/DeclaredVariables as "two test-only wrappers."** They are the only way internal/infer's tests reach internal/parse's declaration data at all -- the package boundary itself, not a wrapper around something already reachable. Cutting them would cut that boundary.
- **Signature.ReturnElement as "read only by tests, keep only if a consumer is planned."** It already has a consumer: the corpus-wide oracle checks (returnTypeFailures' own list-context handling) read it on every run, not a planned one.
