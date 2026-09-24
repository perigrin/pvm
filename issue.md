---
title: Three .t headers state measured claims that are false
state: pending
urgency: normal
milestone: corpus-complete
created: 2026-09-24T05:53:42.371355094Z
updated: 2026-09-24T05:53:42.371355094Z
---

Three `.t` headers state measured claims that are false. Found by the
verify phase of the mdtest port, which measured every prose claim it
carried across rather than trusting it.

The mdtest copies are repaired in `6105cc3d`. These are the originals,
still wrong.

## The three

**`conformance/03_context/04_reverse.t:12`** gives targ numbers that do
not match:

    claimed   reverse[t6] lK/1   reverse[t4] sK/1
    measured  reverse[t4] lK/1   reverse[t7] sK/1

The CONTEXT FLAGS are right and are the claim. The targ numbers are pad
allocation, are not stable, and were never assertable. Two more in the
same tier's `05_sort.t` region: `aassign[t6]` where perl emits `t4`,
and `aassign[t8] lKPS` where perl emits `t5 vKS/COM_AGG`.

**`conformance/10_io/09_glob_assign.t:11-12`**: "Measured across 1,141
lines of source, `*` occurred three times and every one was a `sprintf`
width specifier."

Measured now: 34 occurrences across 1,081 lines excluding tier 10. The
claim was true when written; the corpus grew, and the count included
multiplication `*` as well. The durable claim is the KIND -- no `*` was
the sigil -- not the tally.

**`conformance/13_opaque/*`**, the quote-like terms region: "Three of
the five cases emit an op". Four do. `<*.glob>` emits `glob[t3] lK/1`
and was missed.

## Why this is filed rather than fixed

Those files are due for deletion once the harness migrates off them, so
a fix there is discarded work. Filed so the defect is recorded rather
than silently resolved by deletion -- and because the pattern matters
more than the three instances.

## The pattern

All three are the same shape as the defects this corpus has found
repeatedly: a claim that was true once, or plausible, and that nothing
re-measured. Two of them survived because they cite a NUMBER, which
reads as evidence.

`TestNegativeFactsNameTextTheSourceContains` made token facts
mechanically checkable. Header prose is not, and these are what that
gap costs. A check that re-runs every `MEASURED` block in a header
would have caught all three, and is the obvious follow-on -- the mdtest
format makes it easier, since a case's program is already a fenced
block a tool can extract.

## Acceptance criteria

- [ ] No corpus prose asserts an op targ number (`! grep -rlqE '\[t[0-9]+\]' conformance/mdtest/`)
- [ ] The glob-count claim states a kind rather than a stale tally (`! grep -rq '1,141 lines' conformance/`)
- [ ] The whole conformance suite is green (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
