# Assess round 2: pushback lens (assess-pushback)

- **Subject:** `docs/plans/2026-11-23-chart-precedence-in-recognizer.md` at revision 2, commit c270bbfa3. I read the live text first (`git show c270bbfa3:…`) and the diff from 7b264fa9c last.
- **Rulings taken as given:**
  - "Do just B first and let's re-measure";
  - "Rule them in" (the stage keys);
  - the goal: `speed-ratio-aggregate-longlived` at or below 5, with a budget of 7 days.
- **My role:** I raised P1 to P12 and I am the drafter of the minute. I was told I am biased toward finding my own bar met, so where a release rests on reading the text one way, I say which way.
- **Not done:** I edited nothing in the worktree and did not run `make test`.

## Recommendation: MODIFY

I release eight of the round-1 P findings and agree to carry four (P1, P2, P3, P10).

Revision 2 raises one new high finding:

- **P14:** the exactness clause and the first stop protect the forest digest. That digest hashes the forest `Build` makes, so it is forced to move by the very thing §4 decides.

It also raises two smaller ones:

- **P15:** the contract for the reference recognizer says "exactly" where the reduce side alone gives only containment.
- **P13:** the exit is already met by the baseline.

## The round-1 findings

### Carried

| id | carried? | grounds |
|---|---|---|
| P1 | Carried; I agree | It concerns how the value attaches (per item or per dotted rule). §7 puts the value on items out of scope, and §6 carries P1 to that design by name. |
| P2 | Carried; I agree | It concerns re-propagating a widened value. That applies only to the value, which is out of scope, and §6 carries it. |
| P3 | Carried; I agree | Its consequence (`$a << ($b < $c)` surviving) arises only if a `%prec` unit rule *sets* a value. The reduce side keys on the completing rule's own precedence, read by "the same reading of levels and `%prec` that `forest.NewPrecedence` makes" (§4.4), so revision 2 does not hit the defect. The definition must still be settled when C is designed. |
| P10 | Carried; I agree | It is about the width of the value. There is no value in revision 2. |

### Released

**P4 (the error path for the tables). Released against §4.4.**

§4.4 says:

> "A resolution the chart's grammar cannot map (for example one naming `POSTJOIN`…) is either on a reviewed list or fails generation. The table is committed and drift-checked."

§5 holds `tables-drift` at 0.

The 43 resolutions no referee sentence covers no longer leave an unchecked hole. §4.1 makes the filter a subset of what the precedence constraint removes, and the constraint stays the authority (§2). A table that over-blocks would refuse a derivation bison keeps, and that would move the verdict keys held at 0. A table that under-blocks is harmless, because the constraint still removes the rest.

**P5 (the exit had no derivation). Released against §1 ("2026-11-23, the goal") and §6.**

The exit is now perigrin's ruled goal, not a figure the author proposed. It does not depend on removing `Constrain`, which §7 puts out of scope. Whether the exit means anything is a separate question, raised as P13.

**P6 (stop 4 and the reading protocol). Released against §5, paragraph 2, and §6's third stop.**

- The protocol is named: "three quiet measurements in one session, each key's median, minimum and maximum, before … and after".
- The stops compare the after-reading's median with the before-reading's maximum or minimum, not with a single baseline run.
- The causal clause is gone.
- `Constrain` stays and is not removed partway through, so no stop fires on a state the design expects.

**P7 (the gate for removing `Constrain`). Released: the finding is dissolved by scope (§7: "removing or changing the forest's precedence constraint" is out).**

There is no removal gate left to measure. §4.1's exactness is measurable through the held verdict keys and the resolved-forest digest, though see P14 on which digest.

**P8 (governing decision, amendment, ruling, C1). Released against the header (lines 6-12), §1 and §2.**

- The header cites v2 by path and branch, the recovery plan's D2, and C1.
- §1 records each ruling with its date and its text.
- §2 names the clauses amended (v2 §8 and v2 §2), limits the amendment to one thing, and states that rule 12 is untouched.
- §1 defines B as §4 and C as the value on items.
- §3 cites C1's figures.

Because rule 12 stands, the README's "No fact is added to an Earley item" stays true.

**P9 (Leo skips the reduce filter on intermediate links). Released against §4.5:**

> "A Leo memo never completes a link the filter refuses."

That is the property I asked for, and `leo-disagree` is held at 0. Testing that property, rather than acceptance alone, is the implementing issue's job.

**P11 (where ambiguous lexemes go). Released against §4.5 and §2.**

- §4.5 chooses one reading: refuse "only when every terminal it admits, among those that can follow the rule, refuses it". That is codebase C2's option (a), the lookahead intersected with FOLLOW.
- What remains goes to the precedence constraint, which §2 makes "the authority, and is unchanged".
- So the two candidate homes are reduced to one.

**P12 (the count was not committed). Released against commit 7b264fa9c, which tracks `count.sh` and `count.pl`, and §3's naming of bison 3.8.2 and the pinned perly.y.**

The script refuses any other bison version and any other revision of perly.y. The finding's consequence, "cannot be reproduced from pu", no longer holds once this merges. There is still no committed expected output. I note it as a nit and do not hold on it.

## New findings

### P14. §4.1 and §6, first stop: "the forest digests are unchanged" contradicts §4's own decision (high)

**If it ships:** the first issue that changes the forest build fires the first stop ("a forest digest changes") on the intended behaviour. The alternative is that someone regenerates a golden whose own comment says "Never regenerate it to make a change pass", without a ruling.

**Because:**

- §4 says "The forest build refuses the same completions, so it never packs a derivation through one", and §3's case for doing so is that operator groupings were 47.9% of the packed alternatives.
- `TestForestDigestGolden` (`internal/chart/forest/forestdigest_test.go:101`) hashes "the forest Build makes" (`forestDigest`, line 76, through `built`). That is the forest *before* `Constrain`. Dropping the packed operator alternatives therefore changes it by construction.
- The recognizer's `TestSetDigestGolden` (`recognize/setdigest_test.go:110`, "a change that moves it has changed which sets the recognizer builds") also moves under the default options, since the filter removes items. The decision does not mention it.
- Only `TestResolvedForestDigestGolden` (line 120, `forest.Parse` after resolve) is expected to stay fixed.
- The recovery plan made a moved digest "a stop reported to perigrin", so a moved digest needs a ruling, not a quiet regeneration.

**Modify:**

- §4.1 holds the resolved-forest digest unchanged.
- It states that the built-forest digest and the default-options set digest move by design, each regenerated once under a named ruling. NoLookahead's lines in the set digest do not move.
- The first stop then reads "a held key moves, or the resolved-forest digest changes".

### P15. §4.2: the reference contract says "exactly", and the reduce side alone gives only containment (medium)

**If it ships:** the fuzz oracle written to §4.2's contract reports a disagreement on the first `%nonassoc` chain or shift-side case. Either the oracle is then weakened in an unplanned way, or the issue stalls.

**Because:** §4.2 says "the recognizer accepts exactly what the reference accepts and the precedence constraint leaves a derivation for". The filter refuses only reduce-side completions; §7 puts the shift side out of scope.

Take `1 .. 2 .. 3`, a `%nonassoc` error that the README says the constraint refuses ("Forbid"):

- The reduce side blocks completing `1 .. 2` before `..`.
- The grouping `1 .. (2 .. 3)` is refused only on the shift side.
- So the recognizer still accepts the file, and the constraint refuses it.

The same holds for every input that bison refuses or groups only through a shift-side resolution. What holds is containment: (reference accepts ∧ constraint leaves a derivation) ⊆ recognizer accepts ⊆ reference accepts.

**Modify:** state the two inclusions. The oracle checks both, plus the property that a precedence refusal never comes from recognition with lookahead off.

**What would change my mind:** none. This is unconditional while the shift side is out of scope.

### P13. §6: the exit is met by the baseline, so meeting it carries no information about B (medium-low)

**If it ships:** the waypoint can record "exit met" with zero gain, or with a small regression. That is the one result §6 says decides whether C is designed next.

**Because:**

- The baseline `speed-ratio-aggregate-longlived` is 4.7915 (§3), already at or below 5.
- Repeated quiet readings have straddled 5: the postmortem's checkpoint 3 read 5.18 and 5.11.
- Only the third stop ("median … not below the before-reading's minimum") detects no gain. A stop is "reported", not a failure of the exit, so "exit met" and "stop fired" can both be recorded for one run.

**Modify:** say how the exit relates to the stops. For example: the exit needs the after-median at or below 5 *and* below the before-reading's minimum. Or record that 5 is a ceiling and the gain is read from the stop.

**What would change my mind:** if perigrin meant 5 as a ceiling that the work must not break, and not as a gain target, this reduces to one sentence of wording.

## Checked and not raised

- §4.3's lookahead promise: "lookahead changes recognition's acceptance only by bison's reduce-side precedence". This is consistent with P15's containment.
- The second stop, a resolution needing parser state beyond `indirob`, can be checked now with `count.sh`: there is 1 state-dependent pair.
- I dropped `chart-beats-rd` (reported, not held) as a missing hold. An over-block there would also refuse a file perl accepts, but `rd-accepted-chart-refused` does not see it. Of the population (3 files), all would need a reduce-side resolution to be wrongly tabled for the gap to bite. I judge that too remote to defend as a Y.
