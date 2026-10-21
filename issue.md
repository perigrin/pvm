---
title: "Call-check regressions from the fix round: flattening is read from the expression, sum's result is a number in arithmetic, a fitting member never hides Undef"
state: in-progress
urgency: normal
milestone: core-module-types
created: 2026-10-21T05:19:06.031297507Z
updated: 2026-10-21T05:19:31.912014048Z
sessions:
- start_sha: 2db182e610c0a9eebf6a623550bd13f550d219a4
  end_sha: ""
  commits: 0
  started_at: 2026-10-21T05:19:31.912014048Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-21T05:19:31.912014048Z
---

crochet:review round 2 of PR #488 at 2db182e6 (2026-10-21). Quality report: scratchpad/review2-core-module-types-paad.md, probes scratchpad/r2probes/, psc builds of head, cf08c0ac and pu in scratchpad/r2bin/. A finding pu gets right is a regression this milestone introduced: fix it. A gap pu shares is out of scope here and goes to the follow-up named in the issue.

Quality I6 (calls.go:242-260 possibleMismatch silences `@ARGV ? \*STDOUT : @ARGV>1 ? [1] : undef; tell($h)` even under --warn-possible-undef; cf and pu report it, perl dies; the RFC says Undef stays possible-undef's business), I7 (calls.go:99-107 decides flattening from the argument's type, not its expression: `strftime("%Y",(localtime)[0..5])`, `@lt[0..5]`, map, `mktime(@lt[0..5])` now false arity errors), I8 (List/Util.tpm:67-68 vs ops.go:364, calls.go:173/:242/:412: `sum(@n)+1`, `abs(sum(@n))`, `sum(@n)/@n` errors by default; each exemption clears one problem bit and the union is reported), S17 (`max(@empty)+1` silent even under --warn-possible-undef), S13 (derive.go:158-170: the Signature.Arity cut refuses a mix it used to accept).

## Steps
- [ ] Failing test per finding from its probe
- [ ] Flattening: a slice, list-returning call, map/grep, array or hash flattens by its expression form
- [ ] Exemptions compose: an argument is silent only if every non-fitting bit is excused by some rule

## Acceptance criteria

### Positive Scenarios
- [ ] TestSliceArgumentFlattens (`go test ./internal/infer/ -run '^TestSliceArgumentFlattens$' -count=1 -v | grep -q '^--- PASS: TestSliceArgumentFlattens'`)
- [ ] TestSumResultInArithmeticIsSilent (`go test ./internal/infer/ -run '^TestSumResultInArithmeticIsSilent$' -count=1 -v | grep -q '^--- PASS: TestSumResultInArithmeticIsSilent'`)
- [ ] TestEmptyMaxPlusOneReportedUnderFlag (`go test ./internal/infer/ -run '^TestEmptyMaxPlusOneReportedUnderFlag$' -count=1 -v | grep -q '^--- PASS: TestEmptyMaxPlusOneReportedUnderFlag'`)
- [ ] TestArityMixAcceptedAsBefore (`go test ./internal/parse/ -run '^TestArityMixAcceptedAsBefore$' -count=1 -v | grep -q '^--- PASS: TestArityMixAcceptedAsBefore'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestFittingMemberNeverHidesUndef (`go test ./internal/infer/ -run '^TestFittingMemberNeverHidesUndef$' -count=1 -v | grep -q '^--- PASS: TestFittingMemberNeverHidesUndef'`)
