# crochet:assess round 6: the architecture and direction lens

**Subject.** Revision 6 of `docs/plans/2026-11-23-chart-stateless-precedence.md`, at 788778ed0. The minute is at
0d2b31610, which is HEAD.

**Blob.** 60327eb7d7e379443281e0623b5d5f586a382f9c, checked with `git hash-object`. It matches the expected blob.

**Recommendation: accept, with edits.** The decision is ready. Its exit, stops, keys and scope are settled, and the
evidence directory is now an adequate audit record. Two points about the V2 target (A36, A37) are about the
*next* decision; they should be written down now, because §1 and §5 state that target as this decision's
direction. A38 is a question that needs an answer before refinement, because it touches CLAUDE.md's
`make test` rule.

**Bias.** I raised A33 to A35, and I am the only one who may release them. That biases me toward finding my own bar
met, so each release below quotes its text, and I checked the evidence files rather than trusting the index.

## What I read and checked

I read the subject first and ran these checks; the diff came last.

- **The subject and the index.** I read the subject and `INDEX.md` in full, and listed every file in the evidence
  directory (89 files).
- **The timing runs.** In `measurement/combined-runs/result-run{1,2,3}.txt` I read the `RESULT` lines and
  recomputed V1 and V2 over V0 for each process. I also checked `referee.txt` (1,296 sentences) and `toys.txt`
  (25,490 forests and 4,395 sentences).
- **Round 5's reproduction.** I read `harness/fresh5/result-t1.txt` and `reports/round5-fresh.md`.
- **Round 4's harness.**
  - `harness/fresh4/`: I found the seeds (`rand.NewSource(4)` and `rand.NewSource(44)`) and the 15,000 loop.
  - `list.txt.gz` has 11,588 lines.
- **`make test` and CI.** I read the `Makefile`'s `test` and `chart-absence-reference` targets and CI's step for
  `chart-absence-reference`.
- **The minute.** I grepped it for scratchpad pointers.

## Not examined

- I did not re-run any harness or prototype.
- I did not check whether `chart-absence-reference`'s `corpus()` reads perl.git. That bears on A38, and I leave it
  to the codebase lens or the author.
- I did not read `constrain-runs/`, `codebase-ov4/` or the round-5 reports other than fresh5's.
- I did not trace the 2,566,149 against 2,565,144 difference that the index leaves unexplained.
- I did not read the minute's round-5 section beyond its ledger rows.
- I ran no tests and no ratchet.

## A33–A35

| id | disposition | ground |
|---|---|---|
| A33 | **released, answered by ruling and text** | §1: "Target for V1 then V2 … 3.92 is where this decision and the later V2 decision should land together. It is not this decision's exit. V1 measured 4.12–4.16 on the machine where V2 measured 3.92." §5: "the V2 decision comes next, toward 3.92 for V1 and V2 together." The two readings are reconciled in the text. What the figure *means* is A36. |
| A34 | **released for the subject; the minute's residue goes to the drafter** | `INDEX.md` maps every file to its scratchpad origin. I checked that it holds what A34 named as missing: the fresh3 harness (`harness/fresh3/exp_test.go.txt`), round 4's codebase probes (`harness/codebase-ov4/`), the C1 report (`c1/c1-report.md`), and the generators behind §2's 11,574, 20,083 and 15,000 (`harness/fresh4/`, with fixed seeds 4 and 44 and the 15,000 loop). The minute still points at the scratchpad at lines 50, 74, 498 and 664. I agree that is the drafter's housekeeping, not a defect in the subject. |
| A35 | **released, with a question raised as A38** | §3.4: "it runs as `make chart-precedence-reference` behind its own build tag … `make test` runs it after the suite, and CI's corpus-backed Ratchets job runs it. When its corpora are missing, the target fails; it never skips. The other inputs run in the ordinary suite." |

## Assessment of revision 6

**§1 lists only the rulings in force.** That is right for a decision document, as long as the history is kept
elsewhere, and it is. The minute records every ruling, and §1 says so ("The minute records every ruling, including
the ones these superseded"). The first entry names what it supersedes ("Do just B first" and "V2: both"). "Measure
first, then decide" is spent rather than superseded, so dropping it is correct.

**§3.4 and §6 handle the path rule the right way.**
- **The rule.** v2 §11 and the chart README's isolation paragraph allow a change outside `internal/chart` only
  under perigrin's ruling naming the path.
- **The precedent.** chart-perly-w2 ruled `Makefile` and `.github/workflows/ci.yml` in for MAPDISPATCH
  ("Add to CI job", "Add both to allow-list"). Those rulings came at execute and forced a criterion amendment.
- **What this decision does instead.** It names the two paths now, for ruling at refinement, so the milestone's
  path criterion can carry them from the start. That avoids the amendment and fits the w2 postmortem's N4 lesson
  ("criterion 12 keyed by issue").
- **One suggestion.** Refinement should key the allowance to the issue that adds the target, not to the milestone.

**The evidence directory as an audit record.** It is adequate. Every figure in §2 that I traced has a file behind it:
- **1,296:** `combined-runs/referee.txt`.
- **25,490 and 4,395:** `combined-runs/toys.txt`.
- **24 memo entries:** the `HIST … max 24` line.
- **5.62:** `RESULT V0-parsewith ratio 5.6207`.
- **5.49:** `constrain-measurement.md`.
- **0.727:** `reports/round5-fresh.md`.
- **11,574, 20,083 and 15,000:** generated by `harness/fresh4/`.

The index is also honest about its two known discrepancies. The small inconsistencies I found are A39.

**§4 and §5 are consistent.**
- The stage keys come first.
- The fourth stop, "§3 is built and its checks pass, but the exit is otherwise not met", is a cleaner form of
  round 5's fourth stop.
- §5 says what the result decides.

## New findings

### A36. §1 and §5: the target "3.92" is a harness figure on one machine, not a value of any key (for perigrin; one sentence)

- **Where 3.92 comes from.** It is the combined harness's V2 ParseWith ratio, with V0 reading 5.62 on the same
  machine (`combined-runs/result-run3.txt`: `RESULT V2-parsewith ratio 3.9205`, a median over three runs of 3.85
  to 4.02).
- **How progress will actually be read.** Through `speed-ratio-aggregate-longlived` in the ratchet. On pu the
  baseline reads 4.79 for the same code that the harness reads at 5.62.
- **What the target means in each instrument.**
  - The same improvement, 0.69 to 0.70 of V0, would read about 3.3 to 3.4 on the ratchet.
  - Read literally, 3.92 on the ratchet would be a much weaker goal than the one perigrin saw.
  - On this machine, the ratchet may never read 5.62 for V0, so "3.92" has no stated instrument.
- **Consequence if it ships.** The V2 decision inherits a target it cannot check without restating it. Or it
  claims success, or failure, against the wrong instrument.
- **Modify.** State it as a ratio, for example: "about 0.70 of today's pipeline, measured as the relative exit
  is, before against after in one session". Or name the key and machine, and keep "3.92 in the combined harness" as
  the provenance.

### A37. §1 and §2: V2's increment was not reproduced, and the target rests on one session (should fix; for the V2 decision)

- **The two sessions.** Round 5's fresh reproduction (`reports/round5-fresh.md`) read V2 at 0.723 of V0 (4.10),
  against V1's 0.727. Its words: "V2's ParseWith came out at 0.723 of V0's, not 0.69". The combined session read
  V2 at 3.85 to 4.02 over three processes, with a median of 3.92.
- **What §2 says.** It cites the fresh5 reproduction only for V1 ("0.727 in round 5's independent reproduction").
  It omits fresh5's V2 figure.
- **What §1 says.** "V1 measured 4.12–4.16" mixes two sources: fresh5's 4.12, and the combined runs' median of
  4.16, whose three runs read 4.03 to 4.17.
- **Consequence if it ships.** The direction in §1 and §5 ("toward 3.92 for V1 and V2 together") looks settled
  by measurement. But the only reproduction found V2 adding almost nothing (0.004). The V2 decision then starts
  from an overstated expectation.
- **Modify.** Add fresh5's V2 figure beside the 0.727 in §2. In §1, write "V1 4.03–4.17 over the combined runs,
  4.12 reproduced; V2 3.85–4.02, not reproduced (4.10)". This changes nothing that this decision builds.

### A38. §3.4: "`make test` runs it … it never skips" may make perl.git a hard dependency of `make test` (question; needs an answer before refinement)

- **The new dependency.** §3.4 says the aggregate "needs the T1 and perl.git corpora", that `make test` runs the
  target after the suite, and that "when its corpora are missing, the target fails; it never skips". perl.git
  means `/opt/corpora/perl5` at 94e50866, clean.
- **Why it matters.** CLAUDE.md requires `make test` to pass 100% before every commit, in every package. If
  `make test` does not already need perl.git, this change makes every pvm contributor and every worktree without
  that corpus unable to commit.
- **The cited precedent.** `chart-absence-reference` is run by `make test` too (Makefile, `test:` lines 21 to 23).
  But I did not check whether its `corpus()` reads perl.git. The README says the tdt keys exclude perl.git, which
  suggests it may not.
- **Consequence if it ships.** The `make test` gate gains a machine-specific corpus dependency that CLAUDE.md's
  100% rule cannot accommodate.
- **Modify.** Either:
  - confirm that `make test` already requires the perl5 corpus and cite the test that does; or
  - restrict what `make test` runs to the T1 and in-tree part, and leave the full aggregate to CI's Ratchets job
    and a target of its own.

### A39. §2 and `INDEX.md`: small figure drifts (low)

- **The precedence stage.** §2 gives "0.40–0.41 of today's". Recomputed per process from `combined-runs/`, V1 over
  V0 is 0.423, 0.406 and 0.392, so 0.39 to 0.42 with a median of 0.41. fresh5 read 0.40. Write "0.39–0.42 (median
  0.41; 0.40 reproduced)".
- **The file list.** `harness/fresh4/list.txt.gz` has 11,588 lines against §2's 11,574 files, and the index does
  not explain the 14-line difference. It is likely files tracked in the repository or duplicates, but nothing says
  so. Add one line to the index, as it already does for the alternatives count.

## Summary

| id | status |
|---|---|
| A33 | released (ruling and text; meaning raised as A36) |
| A34 | released for the subject; the minute's scratchpad pointers go to the drafter |
| A35 | released (question raised as A38) |
| A36 | new, for perigrin: restate 3.92 as a ratio, or name its instrument |
| A37 | new, should fix: V2's 0.69 not reproduced (0.723); say so in §1 and §2 |
| A38 | new, question: does `make test` thereby require perl.git? |
| A39 | new, low: 0.39–0.42; the 11,588 against 11,574 lines |

**Recommendation: accept, with edits.**
- What this decision builds is settled and soundly evidenced.
- A36 and A37 keep the next decision honest, and each needs one sentence.
- A38 needs an answer before refinement, not a redesign.
- A39 is cosmetic.
