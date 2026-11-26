# chain-review round 8: closing repair on chart-perly-w4

perigrin ruled at the bound: "Apply fixes, accept (Recommended)". This repair applies exactly the three fixes the raisers proposed and nothing else.

Every edit went through the `git zhi` CLI (git-zhi 0.5.2). The repository, git config and branches were not touched. No issue, edge, title or label changed, so the milestone still has 19 issues.

## The three fixes

### r7-fresh7 FR7-1 (bb9b item 8)

**Issue:** 01a2144c-bb9b.

**Old text.** Item 8's Known-limits bullet said "an alternative kept in one context is kept in all" applies to "`Constrain` only, since production's Build walks by context".

**New text.**
- The union is a property of the shared-node forest.
- It holds for V2's production forest as it does for `Constrain`. d0ba's and d218's union-shortcut equality requires this.
- Walking by context keeps out only the alternatives of a context whose parent chain dies, which V1 never reaches either.

**Criterion.** Unchanged. `TestPrecedenceShortcutStated` requires the phrase "alternative kept in one context is kept in all" in the section, and the reworded item keeps that phrase.

### FR6-6 (d166)

**Issue:** 01a21447-d166.

**New fixture.** `TestNarrowedSetIdsStableAcrossFiles` now uses two files that both keep a library operator in scope. Each narrows to a different terminal set that the committed tables lack, for example two different operators of one library. With per-file id numbering, the two sets would collide on one id.

**Step and criterion.** The step and the criterion description were reworded to say so. The criterion now also requires the log line `distinct narrowed sets outside the committed tables: 2`, so a fixture that does not reach the collision fails. The criterion is still red today.

### r6-fresh6 FR6-3, promoted by r8-fresh3 (d218 and 3282)

**Issues:** 01a21447-d218 and 01a21463-3282.

**The rule.** The candidate recognizes the narrowed lattice production builds: library operators are narrowed by each input's scope, as `parseStream` narrows them before recognition. Design §4.1 claims forest equality only for narrowed input.

**d218.** The rule is stated in the candidate sentence of its repointed production-versus-reference differentials.

**3282.** The rule is stated in its differential prose for the corpus. The `TestForestEqualityCorpusJudgedRecognizer` criterion description now says "over the narrowed lattice production builds". Its command is unchanged.

## Records

The milestone body's Records gained this line verbatim. It sits after the round-7 repair entry, and the rest of the body is preserved:

> - **Chain-review, 2026-11-26.** The bound of eight rounds was reached without a fixed point. Three blocking items remained (r7-fresh7 FR7-1, FR6-6, r6-fresh6 FR6-3), each with a one-line fix the raisers proposed. perigrin ruled "Apply fixes, accept (Recommended)". The three fixes are applied, and chain-review passed by that ruling. It was not backfilled. Participants: review-alignment, review-pushback (drafter from round 4), and fresh readers fresh2 to fresh8. Reports are in the session scratchpad and will be exported with the chain.

No mapping row changed.

## Change log

| Issue | Change | Finding |
|---|---|---|
| 01a2144c-bb9b | Item 8's Known-limits bullet reworded | r7-fresh7 FR7-1 |
| 01a21447-d166 | `TestNarrowedSetIdsStableAcrossFiles` fixture changed in the step and the criterion; log check added | FR6-6 |
| 01a21447-d218 | Candidate sentence now says the candidate recognizes the narrowed lattice | r6-fresh6 FR6-3 |
| 01a21463-3282 | Corpus prose and corpus-criterion description now say the candidate recognizes the narrowed lattice | r6-fresh6 FR6-3 |
| Milestone Records | The ruling line added | coordinator |

## Checks

- **Bodies.** Every other line of each edited body is unchanged. All four stored bodies and the milestone body were re-read and match their source files. Labels and the resolution command are unchanged.
- **Criteria.** Two criteria changed, d166's and 3282's. Both were run at e3432c0f with `PERL5_CORPUS=/nonexistent`, and both are red.
- **Lint.** `internal/chart/offline/criteria-lint.sh` exits 0 over all 19 issues.
