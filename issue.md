---
title: dualvar's DualVar return fits the string and numeric operations perl applies to it
state: pending
urgency: normal
milestone: core-module-types
created: 2026-10-21T20:10:35.060487693Z
updated: 2026-10-21T20:10:35.060487693Z
---

crochet:review round 3 of PR #488 at 692a560c (2026-10-21). Quality report: scratchpad/review3-core-module-types-paad.md; probes scratchpad/r3probes/ (F-<lens>.md per specialist); rebuild head/r2(2db182e6)/cf(cf08c0ac)/pu psc as the report says. perigrin, 2026-10-21: round 4 runs with Opus workers, and every fix gets a mandatory independent review that runs pu, round 2 and head against perl 5.42 before it merges. A finding pu or round 2 gets right is a regression: fix it without trading it for a new one -- every probe in r2probes/ and r3probes/ that pu or r2 gets right must still be right.

Quality V7 (Scalar/Util.tpm:22, 91fb51e3): dualvar now returns bare DualVar, which the lattice fits to neither Str nor Num, so `$x+1`, lc, eq, length and substr on a dualvar give five default diagnostics and exit 1 -- new false positives. A dualvar is both a string and a number in perl. Measure each operation, then make the lattice or the declaration say so; if that needs a lattice ruling (what DualVar satisfies), STOP and bring perigrin the options with a recommendation.

## Steps
- [ ] Measure the five operations on a dualvar in perl 5.42 with -w
- [ ] Fix where perl says, keeping the return-type oracle's DualVar observation

## Acceptance criteria

### Positive Scenarios
- [ ] TestDualvarInStringAndNumericOps (`go test ./internal/infer/ -run '^TestDualvarInStringAndNumericOps$' -count=1 -v | grep -q '^--- PASS: TestDualvarInStringAndNumericOps'`)
- [ ] TestDualvarReturnsDualVar (`go test ./internal/infer/ -run '^TestDualvarReturnsDualVar$' -count=1 -v | grep -q '^--- PASS: TestDualvarReturnsDualVar'`)
- [ ] TestPSCCheckLibBaseline (`go test ./internal/psc/ -run '^TestPSCCheckLibBaseline$' -count=1 -v | grep -q '^--- PASS: TestPSCCheckLibBaseline'`)
- [ ] TestInferencePrecision (`go test ./internal/infer/ -run '^TestInferencePrecision$' -count=1 -v | grep -q '^--- PASS: TestInferencePrecision'`)

### Negative Scenarios
- [ ] TestDualvarStillDistinct (`go test ./internal/infer/ -run '^TestDualvarStillDistinct$' -count=1 -v | grep -q '^--- PASS: TestDualvarStillDistinct'`)
