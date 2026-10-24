# Milestone m3-conformance-corpus Postmortem

## What Worked Well

- **The five goal issues went through cleanly.** Each landed as one commit and one merge, with no reopens:
  - 01a0de11 (fc0447d7)
  - 01a0df09 (203e7db7)
  - 01a0ded8 (02accc49)
  - 01a0dec7 (da1176d9)
  - 01a0dde3 (d36e267d)

  Each ran in its own worktree (wt-y-*) through a serial merge queue gated on make test, and none of the five merges conflicted.
- **Perl stayed the judge.** The review re-judged every new or extended pinned output under perl 5.42.0, without warnings. That was 18 cases in round 1 and 6 more in round 2, and all matched byte for byte. The given/when tier's claims were confirmed by running perl, not by reading perldoc: `use v5.36` drops `given`, and `use v5.42` rejects `~~`.
- **Restoring a deleted property found real gaps at once.** TestAdjacencyCoversItsTier, from perigrin's ruling, failed 7 of 14 tiers on its first run. Tier 02 was missing `shift`, `pop` and `splice` (the gap 01a1696e had predicted), and tier 04 was missing 18 ops. All were fixed without adding a case, so the corpus stayed at 270.
- **Review by mutation worked.** The coverage lens ran every goal criterion at the base commit, 054e75e6, and found 7 that already passed there. All 7 were repointed, and in round 2 each new criterion was confirmed to fail at base and pass at HEAD.

## What Didn't Work

- **A refactor silently broke 40 done issues' criteria.** e8ee9001 renamed the topics and deleted the per-tier adjacency tests. The criteria on done issues still named the old tests, and nothing ran them until milestone verify, which read 175/219. Getting to 220/220 took four passes:
  1. Remapping the stale criteria took it to 18 regressions.
  2. Restoring adjacency, which needed perigrin's ruling, took it to 2.
  3. The last 2 came from issue 01a1696e, moved in from v0.1, whose body had a `- [ ]` steps checklist that the milestone's meta-criteria rejected.
  4. Rewriting that checklist gave 220/220.
- **The refactor also deleted a property the docs still claimed.** e8ee9001's message said "the claim is now in the source; a reader sees that by reading it". The tier READMEs went on asserting adjacency coverage, and 7 tiers no longer met it. That is the corpus-level form of "a claim with no check".
- **The criteria checked documents, not behaviour.** Before review, 7 of the goal issues' criteria were greps or package-wide runs that passed at base. One example is `grep 'floor|at least'`, which matched before any floor existed.
- **The delivered design inverted its decision without a recorded ruling.** 01a0df09 decided on a floor and the code shipped an exact pin, with a comment headed "A FLOOR, NOT AN EXACT COUNT". Review caught it, and perigrin ruled exact.
- **A merge ran with the wrong branch prefix.** The qmerge-cc.sh run for adj omitted `BR=ccx`, so it tried `exec/adj`, merged nothing, ran make test on an unchanged tree, and stopped at NOT-COMMITTED. The guard worked. The default was the wrong one for this milestone.

## What Puzzles Us

- The milestone telemetry from sanbao is not usable. avg_lead_time_hours reads 309849235102893 and agent_autonomy_rate reads 0, though every transition this session ran under an agent: actor. No claim here relies on it.
- Four commits' `Implements:` lines don't parse as git trailers because a blank line separates them from `Claude-Session:`. The worker brief produced that shape, so trailer-based traceability missed 4 of the 5 goal issues.

## Where A Human Had To Act

- **Four rulings that belonged to perigrin. This is collaboration, not friction.**
  - Restoring the adjacency property instead of declaring it obsolete.
  - Two criteria ruled obsolete (the tier-11 ADJUST failure, and tier numbering).
  - Exact pin over floor.
  - Restoring the Grouped coverage check.
- **The 18-criterion escalation was avoidable in timing.** e8ee9001's author could have seen at refactor time that it removed checks named by done criteria. A verify run at refactor time, not at milestone close, would have surfaced the question months earlier.

## What Will We Change

1. **Run `git zhi verify` on every refactor that deletes or renames tests.** Run it on each milestone whose criteria could name those tests, in the refactor's own commit gate, so a rename that orphans a done issue's criterion fails before the merge.
2. **Make refinement's criterion check run each criterion at the base commit.** A criterion that passes before the work exists is a defect. This review found 7 in one milestone.
3. **When a refactor deletes a check, delete or reword the claim it backed in the same commit.** If the docs still assert a property, the check stays.
4. **Fix the worker brief's trailer block.** Put `Implements:` and `Claude-Session:` in one trailer paragraph with no blank line between them, so `%(trailers)` sees both.
5. **Moving an issue into a milestone means conforming its body.** Check it against the milestone's meta-criteria (no `- [ ]` step lines) before moving it in.
6. **Have qmerge-cc.sh default `BR` to its own milestone's prefix (`ccx`)**, not `exec`.

---
Generated by crochet:postmortem on 2026-10-24.
Data sources:
- the issue list and milestone body: the review minute and perigrin's rulings
- the verify logs cc-verify.log through cc-verify4.log (175/219, then 200/219, then 218/220, then 220/220)
- git log 054e75e6..c139e83f: 13 non-merge commits
- sanbao, which was read but is unusable, as noted above
