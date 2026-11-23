# crochet:assess round 4: the architecture and direction lens

**Subject.** Revision 4 of the design at commit 6dc1d9558. The design is now
`docs/plans/2026-11-23-chart-stateless-precedence.md`, renamed from
`…-chart-precedence-in-recognizer.md`.

**Blob read.** 123012c1a5765883ded36de157c8d0e8ac764f0a, checked with `git hash-object`. It matches the expected
blob.

**Order of reading.** I read the subject first, then the minute (renamed to
`docs/assessments/chart-stateless-precedence.md`): its header and the section "Experiments after round 3". Next
came the v2 design on `origin/docs/chart-parser-v2-design`, where I searched for "state" and re-read §3 and §8
(from round 1). I read `git diff -M --stat 3e4d0c43b 6dc1d9558` last.

**Recommendation: modify.** The direction is sound, and the evidence is the strongest this assessment has had. The
blocking issue is that revision 4 puts precedence back into recognition without the amendment text revision 2
had (A26).

**Bias.** I raised A18–A25 and only I may release them. I am biased toward finding my own bar met, so each release
below names the text it rests on.

## Not examined

- I did not re-run either prototype or read `stateless.diff` or `combined.diff`.
- I did not check:
  - the key counts (257,293 and 460,291);
  - the 25,490 and 30,588 toy-forest figures (the minute gives both);
  - whether "conflict check" in the minute's ablation table is decided by §3.1's four inputs.

  These belong to the codebase lens.
- I did not re-read the README since round 2.
- I did not read the recovery plan's D-list in full.
- I did not check whether `recognize`'s fuzz package can reach `forest` without an import cycle.
- I ran no tests and no ratchet.

## A18–A25

| id | disposition | ground |
|---|---|---|
| A18 | **released** | §2: "measured on one machine, where today's whole pipeline reads 5.62 against the committed baseline's 4.79, so compare them with each other, not with `baseline.txt`." The ≤ 5 half is read in §4's one-session reading on that session's machine. That is implicit, and I accept it. |
| A19 | **moot (claim removed)** | Revision 4 drops the "about 5%" bound. The recognizer's reduce side is now in scope, measured (V2 against V1: 0.69 against 0.73). |
| A20 | **released** | Renamed, and the design's header names the old path. The minute still carries the old name: see A29. |
| A21 | **released** | §4: "Their cost to a ratchet run is reported with the commit that adds them." The fresh-copy-per-pass detail adds memory and time outside the span. That is reported, which is what A21 asked. |
| A22 | **released** | §3: "The live documents change first … the chart README's description of the constraint; the recognizer's `Options` contract; the ratchet key table". §4: "The last issue takes three quiet measurements … and records them in Records". |
| A23 | **released** | §5: "The before-and-after stage table goes to perigrin. Whether there is a next stage of speed work, and which stage it targets, is his ruling." |
| A24 | **released, answered by ruling and measurement** | §3.2 settles it: no hash table, built-in or hand-written, per derivation, and per-node lists may grow. A test is required on the multi-entry path. The once-per-grammar maps stay. That is perigrin's "V2: both", and the prototype bounds contexts per node at 10 and 24. |
| A25 | **released, optional and not adopted** | The interleaved before/after reading at the end is the better verdict reading. With the stage keys first, an issue on the path can still report its key. I do not hold this. |

## Assessment

### Is empirical bison-exactness an acceptable standard?

Yes. It is v2's own standard, and the guard makes it stronger than empirical.

- **v2 §8 defines correctness by outcome, not by mechanism.** "These shapes are examples; the definition is bison's
  resolution of each conflict, and the generated tests below are what pin it." Its referee is "bison … offline",
  per rule 11's outside reference.
- **Tracking bison's LR(0) states is the README's mechanism, not v2's requirement.** The README section "The
  precedence constraint" says "The parser states are bison's … never the chart's split states". v2 §8 itself never
  requires states.
- **The guard is exhaustive at generation.** §3.1 says that if any key's decision "differs between bison's states,
  it fails". That checks every state in bison's automaton where a key occurs, not a sample. If the decision is
  constant for every key, then which bison state a derivation sits in cannot matter.
- **What the differential test adds.** It checks the remaining step: that the forest computes the key bison's
  state would hold. That part is empirical, over the referee, 25,490 toy forests, 4,395 toy sentences against bison
  and the whole aggregate, all with zero differences.
- **Recursive descent is a fair comparison.** perigrin's comparison holds: recursive descent is Pratt binding
  powers from CORE.tpm, graded against the oracle, and nobody asks it to be exact by construction.
- **The README records the risk.** "An earlier form keyed the state on the one symbol before the span, which
  over-refused where two states share it (the `twostates` and `indirobsort` toys)." The guard and the parent-item
  and spine-top keys answer exactly that failure. The minute's ablation (parent item removed: 25 files flip,
  `T1/array.t` among them) shows the context is load-bearing.
- **Conclusion.** The forest-side change (§3.1, §3.2) needs no amendment of v2 §8. It changes the README's
  mechanism, and §3 already lists that README section to change first.

### §3.3 does need an amendment, and revision 4 dropped it: A26

See A26 below.

### Recognition's contract against the recovery plan

- §3's contract is the sandwich A13 asked for: a subset of the reference with lookahead on, and exactly the
  reference with lookahead off. `Expected` stays computed with lookahead off, and the reference stays without
  precedence.
- The measured acceptance is unchanged on all 2,211 aggregate files.
- The recovery plan's D7 ("the successor design is something we can revisit after §8 as written") is met: §8 as
  written is built.
- Nothing in the recovery plan sections I read forbids precedence in recognition once §8 is built.
- What is missing is the holder of the contract's lower bound: see A27.

### §4–§5: stage keys first, verdict reading, stops, and the next decision

- The stage keys come first, in their own commit ✓. The fresh forest copy per pass is right, because `Constrain`
  and resolve mutate their input.
- The verdict-style reading is in one session, records the commit, load, cores and kernel, and defines "quiet" ✓.
- The new stop, "the precedence stage is faster but `speed-ratio-aggregate-longlived` stays above 5", is the case
  where the relative half is met and the absolute half is not. Reporting rather than tuning is right. "Faster"
  should be defined as stop 3's complement: the after-median below the before-minimum. Otherwise it reads as any
  fall. This is minor and part of A28.
- What the result decides ✓ (A23).

### Is the record legible after the rename?

The design's header is fine. The minute is not yet, and the evidence lives only in the scratchpad: see A29 and A30.

## New findings

### A26. §3.3 and the missing amendment of v2 §8, §2 and rules 6 and 12 (must fix)

- **The conflict.** §3.3 puts bison's reduce side into recognition. That is what v2 §8 decided against: "Imitating
  that resolution inside an Earley recognizer would be faithful, not fitted. So precedence leaves recognition
  (perigrin, 2026-10-30): recognition uses `perly.y`'s rules as written, is Boolean, and is ambiguous on operator
  chains".
- **The other clauses it touches:**
  - v2 §2: "conflicts are kept as ambiguity and resolved in the forest (§8), not inherited as an LALR table";
  - rule 6, a constraint may "refuse a span where none does, after the chart (§8)";
  - rule 12, scope collapses "after the chart in waypoint 2, as precedence does (§8)";
  - v2 §8's own first line, "This section is the only statement of how precedence works".
- **What revision 2 had.** Its §2, "What this decision amends", named v2 §8 and §2. Round 2's A14 added rules 6
  and 12.
- **What revision 4 has.** No amendment section at all. Only §1's ruling list implies one ("V2: both", "the
  reduce-side filter in recognition").
- **Consequence if it ships.** The same silent reversal of an accepted ruling that A1 raised in round 1. Refinement
  and review read v2 §8 as the only statement of precedence and find §3.3 outside it.
- **Modify.** Restore an "Amends" section that names:
  - v2 §8's 2026-10-30 ruling (recognition stays precedence-free), amended for the reduce side only, with perigrin's
    "V2: both" as its authority;
  - v2 §2's sentence;
  - rule 6's "after the chart" and rule 12's "as precedence does", each amended to "partly before the chart, for the
    reduce side, with lookahead on";
  - which clauses stay untouched: rule 12's (state, distance) clause (items carry nothing new) and §8's definition
    of correctness by bison's resolution.

  Also state where the amended-by relation is recorded, since v2 is on an unmerged branch.

### A27. §3, recognition's contract: nothing named holds its lower bound (should fix)

- **The bound.** "At least what the reference accepts and precedence leaves a derivation for."
- **The gap.** Revision 2 said the fuzz oracle holds the contract. Revision 4 names no holder. The fuzz oracle
  (`oracleDisagreement` against `refRun`) is in package `recognize`, which cannot call the forest's judgement
  without `forest` importing `recognize`. This is an open question for the codebase lens.
- **"Lookahead off".** "With lookahead off, the filter does not run" should be stated as "with no follow set" (or
  equivalent), because `Expected` re-closes with `complete(nil)` while lookahead is on (`expected.go`). If the
  filter keys on `Options.NoLookahead`, `Expected` changes.
- **The missing option.** §6's note that `TestOracleCatchesSkewUnderOneOption` "counts the recognizer's `Options`"
  implies a new option, presumably a switch for the filter. §3 does not name it.
- **Consequence.** An over-refusing filter on input outside the aggregate goes unseen by any recognizer-level test.
  And "`Expected` is unchanged" depends on an implementation detail the decision does not fix.
- **Modify.** Name the test that holds the lower bound, say the filter keys on the absence of follows, and name the
  new option if there is one.

### A28. §6 and §5 wording (low)

- **§6 contradicts §3.** "The forest build applies the shift side, where each child's rule is known". But §3.3 has
  the forest build apply the *reduce* side, and §3.1's forest judgement applies the shift side.
- **"Faster" in §5's new stop needs a definition.** Use stop 3's complement.

### A29. The minute, renamed, still describes revisions 1 to 3 under the old title (should fix before the fixed point)

- **What it says.** `docs/assessments/chart-stateless-precedence.md` at 6dc1d9558 still has:
  - an ABOUTME saying "applying precedence inside the chart recognizer … rounds 1 to 3";
  - the title "Assessment: Chart parser: precedence in the recognizer";
  - "**Subject:** `docs/plans/2026-11-23-chart-precedence-in-recognizer.md`", a path that no longer exists on the
    branch.
- **What it lacks.** No line records the rename or revision 4's commit and blob.
- **Partly expected.** The coordinator updates the minute each round.
- **Consequence.** A reader arriving from the design's "Assessment:" link finds a subject path that is gone, under
  a title naming the retired approach.
- **Modify.** Retitle it. Keep the revision list, add revision 4 with the rename (old path, then new path, at
  6dc1d9558), and keep the old path as history. Earlier rows citing it stay as written.

### A30. The evidence for an empirical decision lives only in the scratchpad (should fix)

- **Where it is.** §2 rests the decision on prototypes ("stateless-proto", "combined-proto"). The minute records
  their figures, but their code (`stateless/stateless.diff`, `combined/combined.diff`) and harnesses exist only in
  the session scratchpad. The same goes for every round's report path the minute lists.
- **Why that matters.** The scratchpad can be cleared, and this decision's standard of correctness is empirical.
  The w1 postmortem's sixth change: "Keep evidence in the Records, not the scratchpad." The C1 spike kept its
  commit reachable on a local branch.
- **Consequence.** Once the scratchpad goes, the decision's measured basis cannot be reproduced or audited, and the
  minute's links are dead.
- **Modify.** Commit the two prototype diffs, or push them to a scratch branch whose name the minute records. Or
  record in the minute that they are retained elsewhere. The differential test in §3 becomes the durable form at
  build time, but until then the basis is perishable.

### A31. The ratchet's precedence referee should judge the production path, not the oracle (should fix)

- **The current path.** `precedence-disagree` and `precedence-covered` come from `chart.CountPrecedence`, whose
  `referee.judge` calls `f.Constrain` (`internal/chart/precedence.go`). `TestPrecedenceToyReferees` also runs
  against `Constrain`.
- **The gap.** §3 keeps `Constrain` as a test-only oracle and requires "every ratchet count key unchanged". It does
  not say which path these keys judge afterwards.
- **Consequence.** If `judge` keeps calling the oracle, the keys stay at 0 while measuring nothing about production.
- **Modify.** One sentence: the referee and the toy referees judge the production precedence, and the oracle sits
  beside them only in the differential test.

## Summary

| id | section | status |
|---|---|---|
| A18 | §2 | released |
| A19 | §6 | moot (claim removed) |
| A20 | header | released (the minute's residue is A29) |
| A21 | §4 | released |
| A22 | §3, §4 | released |
| A23 | §5 | released |
| A24 | §3.2 | released (ruling and measurement) |
| A25 | §4 | released (optional, not adopted) |
| A26 | §3.3 | new, must fix: no amendment of v2 §8, v2 §2 and rules 6 and 12 |
| A27 | §3 | new, should fix: lower-bound holder; filter keyed on absent follows; name the option |
| A28 | §5, §6 | new, low: shift side or reduce side in §6; define "faster" |
| A29 | minute | new, should fix: title, ABOUTME and Subject name the retired approach and a gone path |
| A30 | §2 | new, should fix: prototype evidence only in the scratchpad |
| A31 | §3 | new, should fix: referee keys must judge production, not the oracle |

**Recommendation: modify.**
- Empirical exactness backed by an exhaustive generation-time guard meets v2 §8's own standard. The forest side
  needs no amendment.
- §3.3 does need one, and revision 2 had the text (A26).
- A27–A31 are each a sentence, plus one housekeeping step for the evidence (A30).
