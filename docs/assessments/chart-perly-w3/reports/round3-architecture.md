# crochet:assess round 3: the architecture and direction lens

**Subject:** revision 3 of `docs/plans/2026-11-23-chart-precedence-in-recognizer.md`, now titled "Chart parser:
the precedence constraint without maps". Read at commit 3e4d0c43b.
**Blob read:** a8960d05a84284fb0f9fcec803d101949fd9c405, checked with `git hash-object` on the live file. It matches
the expected blob.
**Recommendation: modify.** The direction is sound and well evidenced. What remains is small and textual: one
caveat to restore, one bound to qualify, and a few protocol points.

**A note on bias.** I raised A12–A17 and am the only one who may release them, so I lean toward finding my own bar
met. For each release below I name the text it rests on, or the measurement or ruling that retired the finding's
subject.

## What I read

- **Read in full:**
  - the subject;
  - the minute's "Rulings after the round's reports" and "Measurement after round 2" sections;
  - `forest/precedence.go`'s map fields, which I grepped.
- **Read in part:**
  - the recovery plan (`origin/docs/chart-recovery-plan`): its section list, §3 M, R2's design, the §6 exit and
    stops, and the D-list headings;
  - the digest goldens, their headers and `samplePaths`.
- **From round 2:** the README rules I cited then.

## Not examined

- **The measurement itself.** I did not re-run it: not the harness, not the profile, not the context counts. I take
  them as the minute records them.
- **Feasibility of the replacement.** I did not check whether array- or slice-based structures can replace
  `firsts`, `marked` and the goto map at equal or lower cost. That belongs to the codebase lens.
- **The recovery plan's D1–D11 text.** Not read in full.
- **The ratchet's current wall time.** I did not measure what the three new stage keys add to it. My estimate in A21
  is arithmetic, not a run.
- **Nothing run.** No `make test`, no ratchet.

## A12–A17

| id | disposition | ground |
|---|---|---|
| A12 | **moot (approach retired)**; the new text is right | Revision 3 leaves `Build` alone. §3.1 holds both `forest-digest.golden` and `resolved-forest-digest.golden` unchanged, which is now correct by construction. |
| A13 | **moot (approach retired)** | The reference-recognizer contract belonged to the recognizer filter, which §5 puts out of scope. The minute's measurement also recorded the filter as exact ("0 alternatives it removes are kept by Constrain, and 0 verdicts flip"). |
| A14 | **moot (approach retired)** | Revision 3 changes no semantics. `Constrain` "computes exactly what it computes today" (§3), so nothing in v2 §2, §8 or §3's rules 6 and 12 is touched, and nothing needs amending. See A22 for the live README. |
| A15 | **released, answered by measurement** | A15 asked how much B could remove. The minute: the reduce-side pass prunes 5.61% of alternatives, and `Constrain` still takes 0.891 of its time. perigrin decided on that figure. |
| A16 | **released** | §4: "interleaved in one session at the end, with the commit, load, core count and kernel recorded for each … 'Quiet' means a 1-minute load at or below the core count." That is the text A16 asked for. |
| A17 | **released, answered by ruling** | §1 and §4: "Relative and ≤ 5", so the after-reading's median must be below the before-reading's minimum. The half that carries meaning is now part of the exit. The machine-dependence of the "≤ 5" half is handled in A18. |

## Assessment of revision 3

### Consistency with v2 and the recovery plan

- **v2 needs no amendment.** The constraint's semantics, its kept-in-one-context shortcut (§5) and the forest's
  role (v2 §8) are all unchanged. Isolation holds (§5's last line, v2 §11).
- **The recovery plan has a precedent for this work.** Its R2, "No per-node maps or slices", is the same kind of
  change applied to the recognizer.
- **The stop matches the recovery plan's rule.** Stop 2 (the precedence stage key's after-median not below the
  before-minimum) is that rule: "An issue that shows no gain is reported before it merges … its stage key's median
  stays inside the before-reading's min–max range".
- **The stage keys follow the recovery plan's ordering.** §3.4 has them "taken on `pu` before the change", matching
  M ("before any R or F issue"). They are ruled in (§1), and §3.4 calls them never gated and never compared with
  the baseline, which agrees with the README's rules.

### The measurement as a basis for the decision

It is sound for the decision taken: which stage to work on, and that its cost is in map operations, not in forest
size.

- Every figure is internal to one harness: stage against stage, three variants of the forest against each other,
  the profile's split. All of it was interleaved within rounds, over three quiet processes, on the ratchet's own
  population.
- The minute's caveat, "compare these stages with each other, not with the ratchet's keys", is the right scope.

It is not a basis for the exit's absolute half, and it does not claim to be. A18 covers what the decision's own §2
drops.

### Retiring revisions 1 and 2

§5 is honest that they are retired. Its *bound* is narrower than it reads; see A19.

### The file name

The file name matters a little; see A20.

### What the result decides next

The decision does not say; see A23.

## New findings

### A18. §2: the decision drops the minute's caveat and the whole-pipeline figure

- **What §2 does.** It presents the harness's stage ratios under the heading "Stage time, as a ratio over
  recursive descent's time" and sends the reader to the minute for "the full report".
- **What it leaves out.** The minute's caveat: "The whole-pipeline ratio differs from the baseline's 4.79: the
  machine is different and the harness holds every forest in memory. Compare these stages with each other, not
  with the ratchet's keys."
- **Why that matters.** A reader of the decision alone compares recognize 0.94 and build 0.73 with the baseline's
  `speed-stage-recognize` 0.7903 and `-forest` 0.6620, and sees a regression that is not there.
- **The absolute half of the exit.** §4's "at or below 5" will be read on whatever machine the session runs on,
  where the harness read 5.49 for the whole pipeline. The decision should say the absolute half is read on the
  session's machine and is not comparable with `baseline.txt`'s 4.79.
- **Consequence if it ships.** Refinement or review misreads the stage figures, or anchors the ≤ 5 half to the wrong
  number.
- **Modify:**
  - copy the caveat and the 5.49 (5.48–5.64) figure into §2;
  - add one sentence on where the ≤ 5 half is read.

### A19. §5: "bounds their gain at about 5% of the pipeline" is narrower than it reads

- **Where 5% comes from.** It is `Constrain`'s saving alone: (1 − 0.877) × 2.38 / 5.49 ≈ 5.3%.
- **What it leaves out.** Earlier filtering also shrinks the forest by up to 16.17% of packed alternatives, for
  `build` (0.73), `resolve` (0.20) and `readings` (0.15). On the other side, recognition would gain work.
- **What it depends on.** The bound was measured on today's `Constrain`, whose cost "does not follow the forest's
  size" because per-context map overhead dominates. Once this decision removes that overhead, `Constrain`'s cost may
  follow the forest's size more closely, and the bound has to be re-measured.
- **What still holds.** The retirement is sound for now. C1's data supports it too: recognition took 1.7× to 2.9×
  as long with values, at items per token 4.302 to 4.242.
- **Consequence if it ships.** A later reader treats "about 5%" as a permanent ceiling and never revisits
  precedence in recognition after the map work changes the cost shape.
- **Modify.** Say "bounds `Constrain`'s share of their gain at about 5%, measured on the map-based implementation",
  and note that the bound is re-measured after this decision.

### A20. Header: the retained file name says the opposite of the decision

- **The problem.** `docs/plans/2026-11-23-chart-precedence-in-recognizer.md` will be the record of a decision *not*
  to put precedence in the recognizer. Line 6 explains this, but only to someone who has already opened the file. A
  search of `docs/plans/` by name for "was precedence in the recognizer decided" finds the wrong answer.
- **Precedent.** The assessments use a name-for-name convention (`docs/assessments/chart-waypoint.md` for its plan),
  so renaming has a cost, and the round-1 and round-2 rows cite the old name.
- **Consequence.** Small: misleading discovery.
- **Modify.** Rename the plan and the minute with `git mv` when merging (`chart-precedence-constraint-maps`, say),
  and have the minute record the old name. Or perigrin rules that the name stays.

### A21. §3.4: the three keys' cost to the ratchet's run time is unstated

- **The budget.** It is already overrun: "budgeted at 300 s … 543 s on pu … 664 and 683 s in KEYS's verifying and
  baseline runs" (README, "Budget"; input 01a1ec3c).
- **What one stage key costs.** Three measurements of 2 warm-up and 9 timed pairs means 33 passes of the stage and
  33 of recursive descent over the aggregate. Recursive descent is 546 ms in the harness, so the precedence key alone
  is about 33 × 0.546 × (1 + 2.38) ≈ 61 s. resolve and readings add about 21–22 s each.
- **Prepared inputs.** The keys also need built forests for 2,211 files, prepared outside the span and held live.
- **Total.** Roughly +100 s (my arithmetic, not a run), plus memory.
- **Consequence.** Every close gate that runs the ratchet gets slower. Verify's 1,800 s per-criterion timeout
  matters at the close (w2 Design Rationale, "The close runs").
- **Modify.** State the expected cost, or put it in the key issue's criteria, and say that 01a1ec3c still covers
  the overrun.

### A22. The README and the exit's mechanics are unstated

- **The live README must change before the build.** That covers the key table and the rules, the
  `speed-stage sum lex+recognize+forest` stderr line (now six stages), and "There is no per-stage collector key"
  context. The decision does not say so.
- **How the relative exit is read.** It needs a before-state rebuilt in the same session as the after-state. That
  cannot be a `git zhi verify` criterion under the close's timeouts.
- **Precedent.** chart-recovery kept its verdict reading out of the resolution command: "It does not check a speed
  limit … The verdict reading itself … is recorded in the milestone's Records" (recovery plan §6, Exit).
- **Consequence.** Refinement writes an exit criterion that cannot run, or that runs the ratchet once and compares
  it with `baseline.txt`, which the README forbids.
- **Modify:**
  - one sentence: the exit is read by the last issue and recorded in Records, as chart-recovery did;
  - one sentence naming the README sections that change, doc-first.

### A23. §4 and §5: the decision says nothing about what its result decides next

- **The gap.** Revision 2's §6 had a "what the result decides" sentence. Revision 3 dropped it.
- **What is unsettled.** After the map work, the possible next targets are recognize (0.94), build (0.73), the rest
  of `Constrain`, or nothing more at ≤ 5. And does a met exit re-open precedence in recognition, given A19?
- **Consequence.** The next decision starts from no recorded intent.
- **Modify.** One sentence. For example: "The result goes to perigrin with the six stage keys; which stage is next,
  if any, is his ruling."

### A24 (for perigrin). §3.2: taken literally, "no hash map" excludes the recovery plan's proven hybrid

- **The precedent.** Recovery plan R2's design for the recognizer: "find duplicates by scanning, switch to a map at
  24 items".
- **The conflict.** The minute's counts fit a hybrid: 86% of nodes have one context, and none has more than 12 or
  32. But §3.2 forbids any map on the path, while §3.3 asks that a slow node stay correct. A hybrid satisfies §3.3
  and violates §3.2.
- **Consequence.** Small, if the scan bound is fixed and correct. But the ruling "We need to eliminate the map" may
  have meant the profile's maps, not every map in every case.
- **For perigrin.** Confirm the literal reading, or allow a bounded fallback.

### A25. §4: the precedence stop fires only after merge

- **The stop.** It reads before and after "at the end". The recovery plan instead had each issue report its stage
  key before merging (M, "reported before it merges").
- **Why it matters little here.** If this is one or two issues it matters little, and the interleaved end session
  is the better final reading.
- **Why it might matter.** An issue whose precedence stage key does not move would otherwise merge unreported. That
  is the w2 postmortem's lesson: "stopped the waypoint at FOREST's gate rather than after its merge".
- **Modify, optional.** Each issue on the evaluation path reports the precedence stage key before and after, in an
  interleaved pair, before it merges.

## Summary

| id | section | status | what it needs |
|---|---|---|---|
| A12 | §3.1 | moot (retired); the new text is correct | |
| A13 | — | moot (retired) | |
| A14 | — | moot (retired); nothing in v2 needs amending | |
| A15 | — | released, answered by measurement | |
| A16 | §4 | released | |
| A17 | §1, §4 | released, answered by ruling | |
| A18 | §2 | new, should fix | restore the caveat and 5.49; say where ≤ 5 is read |
| A19 | §5 | new, should fix | the 5% is `Constrain`'s share, implementation-relative |
| A20 | header | new, low | rename at merge, or perigrin rules it stays |
| A21 | §3.4 | new, should fix | state the keys' cost to the ratchet run |
| A22 | §3.4, §4 | new, should fix | README sections that change; exit read into Records, not a verify criterion |
| A23 | §4, §5 | new, should fix | what the result decides next |
| A24 | §3.2 | for perigrin | literal "no map" against a bounded hybrid |
| A25 | §4 | optional | per-issue stage-key report before merge |

**Recommendation: modify.**
- Revision 3 is consistent with v2, which it no longer amends, and with the recovery plan's method. Its exit and
  stops follow the README's timing rules and chart-recovery's verdict practice.
- The measurement supports the decision as taken.
- Nothing blocks. A18, A19, A21, A22 and A23 are each a sentence or two. A20 and A24 are perigrin's to settle.
