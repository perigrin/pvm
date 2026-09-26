---
title: sanbao's per-issue hotspot count is the repo-wide total, so the gate tier is chosen from the wrong number
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-26T05:37:44.027283816Z
updated: 2026-09-26T05:37:44.027283816Z
---

`git zhi sanbao --format json` reports a per-issue `hotspot_count` that is
the REPO-WIDE hotspot total, not the issue's own. `crochet:execute`'s gate
analyst reads that field to choose a review tier, so the tier is chosen from
a number that does not describe the issue.

## Measured, 2026-09-26, milestone corpus-covers-t2

    $ git zhi sanbao corpus-covers-t2 --format json
    difficulties[]:
      01a0db78  score 0.134   hotspot_count 22     coupling_count 5
      01a0dba9  score 0.138   hotspot_count 8      coupling_count 0
      01a0dbe8  score 0.3324  hotspot_count 1195   coupling_count 9755

    $ git ls-files | wc -l
    908

**1195 hotspots against 908 tracked files.** The number cannot be a count of
files an issue touched, because it exceeds every file in the repository.

    $ python3 -c "...len(d['complexity']['hotspots'])"
    1195

It is the length of the repo-wide `complexity.hotspots` array, whose top
entry is `internal/infer/infer.go` (churn 55, score 234575) -- a file
`01a0dbe8` never touched.

The real per-issue counts, from the issues' own commits:

    01a0db78   9 files   (77cf71aa, 4620d9d3)
    01a0dbe8  10 files   (ab15f0b3, 23805082)

`coupling_count` looks the same: 9755 pairs reported while
`complexity.coupling` is an EMPTY array, so the field is not reading what a
reader would assume either.

## Why it matters even though it changed nothing here

`crochet:execute` step 3.5 hands these numbers to a gate analyst whose rule
is "TIER_2 when hotspot files > 2". Every real count here already exceeds 2,
so all three issues were TIER_2 either way and no review was skipped.

The failure mode is the other direction: **a genuinely small issue would
also report ~1195 and escalate to TIER_2**, which dispatches roughly a dozen
subagents. The skill says so itself -- "Tier 2 is expensive... a gate too
expensive to run is one that gets worked around". A metric that always says
TIER_2 makes the tier decision inert and the gate a cost with no signal.

`difficulty score` appears sound: 0.134 / 0.138 / 0.3324 track the actual
size and blast radius, and `01a0dbe8` is correctly the largest of the three.
So the scoring is not wholly broken; the two count fields are.

## Scope

1. Establish whether `hotspot_count` and `coupling_count` are meant to be
   per-issue and are mis-populated, or are documented as repo-wide and
   `crochet:execute` reads them wrongly. Those have different owners.
2. If the fields are meant to be per-issue, they need the issue's own file
   set -- the commits carrying its `Implements:` trailer.
3. Until then, `crochet:execute`'s gate step should derive the file count
   from the issue's commits rather than the field, and say so.

## Context

- paths: n/a here -- the sanbao implementation is in the git-zhi binary, not this repo
- commands: `git zhi sanbao <milestone> --format json`, `git ls-files | wc -l`, `git show --stat <sha>`
- entrypoints: `crochet:execute` step 3.5, which quotes `hotspot_count` into the gate analyst's prompt

## Acceptance criteria

- [ ] The per-issue count no longer exceeds the repository's file count (`test "$(git zhi sanbao corpus-covers-t2 --format json | python3 -c 'import json,sys; print(max(x["hotspot_count"] for x in json.load(sys.stdin)["difficulties"]))')" -le "$(git ls-files | wc -l)"`)
- [ ] Which of the two causes it is, is written down (`grep -qiE 'mis-populated|repo-wide by design' docs/audits/2026-09-26-sanbao-attribution.md`)
- [ ] The measurement that found it is recorded (`grep -q '1195' docs/audits/2026-09-26-sanbao-attribution.md`)
