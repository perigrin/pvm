---
title: T2's target excludes what M1 does not own, and says which milestone does
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0c13f-6816-79d2-945a-b84a12d39e6c
blocks:
- 01a0c10b-7148-77e5-8923-16fbce93c64b
created: 2026-09-20T23:56:24.16253937Z
updated: 2026-09-20T23:57:07.958042262Z
---

M1's gate says "100% of T2 (56 files)". Some of those 56 files cannot be M1's
to fix, and a target that includes them is unreachable by construction. The
plan itself warns against exactly this: "Do not carry an unreachable 100% into
a gate: that is how a gate becomes advisory."

This issue makes T2's denominator honest, in the test, with each exclusion
named and attributed to the milestone that owns it.

## Context

- paths: `internal/parse/t2_test.go`,
  `docs/plans/2026-09-05-parser-conformance-plan.md` (§"M1 -- Parses the core
  without error")
- commands: `go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v`
- entrypoints: `internal/parse/t2_test.go`, `TestT2CoreParses` and its
  `shortfall` map

## What is out of M1's scope, measured at b6e7c351

    comp/proto.t          41 Unknowns   prototype recognition -- M4
    t/class/*.t (9 files) 52 Unknowns   class syntax; ADJUST and class
                                        phasers are recorded in the shortfall
                                        map as owned by NOTHING on the chain
    block-form filehandle  1 per site   `print {$fh} "x"` -- blocked per the
                                        m2 refinement

`comp/proto.t` is the unambiguous one: M4 is titled "Prototypes, context, and
the hard ambiguities", and 41 nodes of prototype work counted against M1's
100% is 41 nodes M1 cannot move without doing M4's job early.

The class files need a judgement, not an assumption -- some class syntax is
already landed (`class/class.t` left the map) and some is not. The step below
says to decide each with a measurement rather than by category.

## What this issue does NOT do

It does not lower the bar to whatever today's number happens to be. A file is
excluded only when the construct blocking it is owned by a named later
milestone or a named blocked issue. Everything else stays in, at 100%.

## Steps

- [ ] Write a failing test asserting T2's M1-scoped denominator: the file
      list minus the exclusions, each exclusion carrying a comment naming the
      milestone or issue that owns it
- [ ] Run it and see it fail on today's numbers (`go test ./internal/parse/ -run '^TestT2ScopeIsM1s$' -count=1 -v`)
- [ ] Implement the split in `TestT2CoreParses`: report BOTH numbers -- clean
      files over all 56, and clean files over the M1-scoped denominator -- so
      the plan's metric and the reachable one are both visible and neither is
      quietly replaced
- [ ] For each excluded file, confirm by measurement that its Unknowns are
      ACTUALLY the excluded construct, not a mixture. A file whose Unknowns
      are half prototypes and half something M1 owns is not excluded; it is
      partially fixable and stays in.
- [ ] Update the plan's M1 table to state the scoped target beside the raw
      one, with the exclusion list and its reason
- [ ] Run to green and commit

## Acceptance Criteria

### Positive Scenarios

- [ ] T2's M1-scoped denominator is asserted in the test, with each exclusion
      naming its owning milestone (`go test ./internal/parse/ -run '^TestT2ScopeIsM1s$' -count=1 -v | grep -q '^--- PASS: TestT2ScopeIsM1s'`)
- [ ] `TestT2CoreParses` reports both the raw 56-file rate and the M1-scoped
      rate, and is green (`go test ./internal/parse/ -run '^TestT2CoreParses$' -count=1 -v | grep -q '^--- PASS: TestT2CoreParses'`)
- [ ] `comp/proto.t` is excluded with M4 named as its owner (`grep -q 'comp/proto.t' internal/parse/t2_test.go`)
- [ ] The plan records the scoped target beside the raw one (`grep -qi 'scoped' docs/plans/2026-09-05-parser-conformance-plan.md`)

Plan: M1, M4. Survey: 01a0c10b's closing note.
