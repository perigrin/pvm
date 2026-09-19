---
title: "A file verdict names no work: keep the per-statement findings the comparison already computes"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-19T16:22:28.75915195Z
updated: 2026-09-19T16:22:28.75915195Z
---

A file verdict says `op/bop.t: WRONG`. It does not say which statement, which
marker, or which construct -- so the 87-vs-2 WRONG gap 01a0b835 measured is a
number with no work attached to it.

The per-statement data ALREADY EXISTS and is thrown away.
`decideMarker` (compare_facts.go:210-249) attributes every one of perl's
sites to the innermost statement covering it and collects them as lines:

    var wider, wrong []int
    for _, line := range refs {
        owner := innermost(stmts, line)
        switch {
        case owner != nil && owner.decided[m] > 0:  owner.decided[m]--
        case owner != nil && owner.hedged[m] > 0:   wider = append(wider, line)
        default:                                    wrong = append(wrong, line)
        }
    }

Then it formats those slices into prose and returns one Verdict. The
line numbers survive only inside `Detail`, where nothing can aggregate them.

## What this is NOT

Not a change to the buckets or to the gate. `CompareFacts` keeps returning
one verdict per file, and TestGoParserWrongIsZero keeps reading
`Verdict.Bucket`. A per-statement BUCKET would need a threshold -- how does a
file with 3 wrong statements out of 200 score? -- and a threshold is the
thing ratchets exist to avoid.

This is diagnosis, not scoring. The two are separable and should stay so.

## What it settles

01a0b835 is paused on whether five Unresolved sites per Unknown is the right
hedge. That hedge is CORRECT as a scoring answer -- sites.go argues it well,
and staying quiet is a silent wrong answer. What makes it feel wrong is that
it is useless as a DIAGNOSIS: five identical markers on one statement say
nothing about which construct defeated the parser.

Per-statement findings fix the diagnosis without touching the scoring, which
is why this is filed separately rather than as a change to the hedge.

## Precedent

This is what moved canon today. 92 files failing fidelity turned out to be
ONE bug (the fat comma) and 177 another (a missing block separator). A file
count cannot show that; a construct histogram can. The same read of the same
data is what this asks for on the oracle side.

## Shape

Add findings as DATA on the verdict rather than as prose:

    type Finding struct {
        Line   int
        Marker Marker
        Bucket Bucket   // Wrong or Wider only
    }

`decideMarker` populates it from the slices it already builds. The sweep then
groups by (marker, construct) across the corpus and prints a ranked table.

Attributing a finding to a CONSTRUCT rather than a line needs the subject to
say what it found there -- the statement's node kind, or Unknown's text. That
is a subject-contract addition and may be worth its own issue once the
line-level table shows whether it is needed.

## Acceptance

- [ ] A verdict carries its per-statement findings, not only prose (`go test ./internal/parseoracle/ -run '^TestVerdictCarriesFindings$' -count=1 -v | grep -q '^--- PASS: TestVerdictCarriesFindings'`)
- [ ] Findings agree with the Detail text they replace, on the same input (`go test ./internal/parseoracle/ -run '^TestFindingsAgreeWithDetail$' -count=1 -v | grep -q '^--- PASS: TestFindingsAgreeWithDetail'`)
- [ ] The buckets and the gate are unchanged, asserted (`go test ./internal/parseoracle/ -run '^TestFindingsDoNotMoveBuckets$' -count=1 -v | grep -q '^--- PASS: TestFindingsDoNotMoveBuckets'`)
- [ ] The sweep prints a ranked table of where WRONG comes from (`go test ./internal/parseoracle/ -run '^TestSweepReportsWhereWrongComesFrom$' -count=1 -v | grep -q '^--- PASS: TestSweepReportsWhereWrongComesFrom'`)

Plan: M1. Spec: §7.2.
