---
title: "A file verdict names no work: keep the per-statement findings the comparison already computes"
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-19T16:22:28.75915195Z
updated: 2026-09-19T17:06:55.626455084Z
sessions:
- start_sha: 461ea788c4b3a17b7c9771e80fc6e45de3bbbf18
  end_sha: afdd06c9562e8c8efaa4fb31bad3824739ffb0ba
  commits: 1
  started_at: 2026-09-19T16:54:17.669777988Z
  ended_at: 2026-09-19T17:06:55.626455084Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T16:54:17.669777988Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-19T17:06:55.626455084Z
observed_paths:
- internal/parseoracle/compare.go
- internal/parseoracle/compare_facts.go
- internal/parseoracle/findings.go
- internal/parseoracle/findings_test.go
- internal/parseoracle/goparser_test.go
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

- [x] A verdict carries its per-statement findings, not only prose (`go test ./internal/parseoracle/ -run '^TestVerdictCarriesFindings$' -count=1 -v | grep -q '^--- PASS: TestVerdictCarriesFindings'`)
- [x] Findings agree with the Detail text they replace, on the same input (`go test ./internal/parseoracle/ -run '^TestFindingsAgreeWithDetail$' -count=1 -v | grep -q '^--- PASS: TestFindingsAgreeWithDetail'`)
- [x] The buckets and the gate are unchanged, asserted (`go test ./internal/parseoracle/ -run '^TestFindingsDoNotMoveBuckets$' -count=1 -v | grep -q '^--- PASS: TestFindingsDoNotMoveBuckets'`)
- [x] The sweep prints a ranked table of where WRONG comes from (`go test ./internal/parseoracle/ -run '^TestSweepReportsWhereWrongComesFrom$' -count=1 -v | grep -q '^--- PASS: TestSweepReportsWhereWrongComesFrom'`)

Plan: M1. Spec: §7.2.


## Done, 2026-09-19 (afdd06c9)

    full suite green, 0 FAIL across the repo

All four acceptance criteria verified by running each command as written.

## The scope held

Findings are DATA on the verdict; Bucket is untouched and the gate still
reads Bucket alone. TestFindingsDoNotMoveBuckets asserts every bucket across
all five shapes -- exact, wider, wrong, no-answer, declined.

## Two things the implementation forced

**Findings must span every marker.** compareMarkers keeps the WORST verdict
and discarded the rest, so a file wrong on srefgen and also wrong on rv2hv
reported only one of them. With five markers holding 249 sites between them,
that understated the work by most of it. The bucket still comes from the
worst marker; the findings accumulate across all of them.

**A wider finding is not work to do.** WrongSites counts BucketWrong only: a
hedge is the subject saying "I cannot settle this", which is the behaviour
the four-bucket design exists to reward, and counting it would invert the
metric. Both kinds are still REPORTED on the verdict -- a wider site is a
site the subject did not account for -- and only the aggregation filters.

## Where I was wrong

My first TestFindingsAgreeWithDetail asserted that every finding's line
appears in the Detail text. It does not, and the code is right: a WRONG
verdict's sentence names the lines it COMMITTED to and gives only a count for
the wider ones, because one sentence cannot carry two line lists and stay
readable. The test now asserts what the prose actually carries -- the lines
of its own bucket, and a count covering both.

## Verified on a real failure

By emptying the M4 pin and reading the gate's output:

    18 WRONG site(s) in 1 file(s)
    by marker:
      srefgen       18
    by file:
         18  comp/proto.t

which is the pinned prototype work, correctly attributed.

## Incidental

Fourteen positional Verdict literals broke on the new field. Keying them is
what stops the next field from doing it again.
