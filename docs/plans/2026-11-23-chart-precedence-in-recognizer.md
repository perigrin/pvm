# Chart parser: bison's reduce-side precedence in the recognizer

Status: proposed, for crochet:assess (revision 2). Date: 2026-11-23.
Assessment: `docs/assessments/chart-precedence-in-recognizer.md`.

The governing design is the accepted v2 design,
`docs/plans/2026-10-30-chart-parser-v2-design.md` (branch
`docs/chart-parser-v2-design`), as amended by the Chart Parser Recovery Plan
(perigrin's D2, 2026-11-13). This decision follows waypoint 2
(`chart-perly-w2`, completed 2026-11-23; postmortem
`docs/postmortems/chart-perly-w2.md`) and the chart-recovery spike C1
(`chart-recovery` Records, 2026-11-13).

## 1. Rulings this decision rests on

All perigrin's, recorded here and, at refinement, in the milestone's Records.

- 2026-11-14, D7: "I think the successor design is something we can revisit
  after §8 as written". §8 as written (precedence as a constraint on the
  forest) is now built.
- 2026-11-14, on the v2 spec's standing: "The spec isn't in-volatile … it was
  designed with a less perfect understanding of the problem".
- 2026-11-23, after round 1 of this assessment: "Do just B first and let's
  re-measure". B is §4 below; the value on items (C) is not in this decision.
- 2026-11-23: the three stage keys of §5 are ruled in ("Rule them in").
- 2026-11-23, the goal: exit `speed-ratio-aggregate-longlived` at or below
  5; budget 7 days.
- 2026-11-23: Chalk is a reference for the mechanism only ("it's
  implemented in pure Perl and its grammar isn't as comprehensive as ours").

## 2. What this decision amends

v2 §8 says precedence leaves recognition (perigrin, 2026-10-30) and v2 §2
says conflicts are kept as ambiguity and resolved in the forest. This
decision amends both for one thing only: recognition and the forest build
refuse a completion that bison's precedence would never make (§4). The
forest's precedence constraint stays the authority, and is unchanged. v2 §3
rule 12 is untouched: items stay (state, distance) and carry nothing new.

## 3. Problem and evidence

From `internal/chart/testdata/ratchet/baseline.txt` on `pu`, chart time over
recursive descent's: `speed-ratio-aggregate-longlived` 4.7915 and
`speed-ratio-std5k-longlived` 7.2212, one run each; the timed stages lex
0.6551, recognize 0.7903 and forest 0.6620. Precedence, resolve and the
readings count are untimed. The postmortem's probe put about 115 ms of
`std5k` in the precedence constraint against about 25 ms for lex plus
recognize.

The chart recognizes perly.y with its precedence stripped, so the forest
holds every grouping of an operator chain and the constraint removes all but
bison's. C1 measured what a value on items would buy: items per token 4.302
to 4.242, while operator groupings were 47.9% of the forest's packed
alternatives. The cost of ambiguity is in the forest, not in the items.

`internal/chart/offline/precedence/count.sh` classifies bison 3.8.2's 1,347
precedence resolutions of the pinned perly.y. 670 resolve as shift: bison
never completes the rule before the token. 669 hold in every state where the
rule completes; the one exception is rule 370, `indirob: scalar`, before
`[`, which bison refuses after `print` and allows after `@`.

## 4. The decision

**The reduce side of bison's precedence is applied where a completion is
made.** A completion of a rule is refused when every terminal that can
follow it there is one before which bison's precedence never reduces that
rule. The forest build refuses the same completions, so it never packs a
derivation through one.

What must become true:

1. **Exactness.** The pipeline's verdict and its resolved forest are
   unchanged on every corpus: what the filter removes is a subset of what
   the precedence constraint removed. The forest digests (`forest/testdata`)
   are unchanged.
2. **Independence of the reference.** The plain-Earley reference recognizer
   stays without precedence. Its contract becomes: the recognizer accepts
   exactly what the reference accepts and the precedence constraint leaves a
   derivation for. The fuzz oracle holds that.
3. **Lookahead.** The filter needs the next token. With lookahead off
   (`Options.NoLookahead`, and interleaved lexing later) it does not run, and
   acceptance with lookahead off is unchanged: the constraint still refuses.
   The documented promise becomes "lookahead changes recognition's
   acceptance only by bison's reduce-side precedence".
4. **Every resolution is accounted for.** The table is generated from the
   grammar, by the same reading of levels and `%prec` that
   `forest.NewPrecedence` makes. A resolution the chart's grammar cannot map
   (for example one naming `POSTJOIN`, which the chart omits) is either on a
   reviewed list or fails generation. The table is committed and
   drift-checked, as the grammar tables are.
5. **Leo and ambiguous tokens.** A Leo memo never completes a link the filter
   refuses. A token admitting several terminals is refused only when every
   terminal it admits, among those that can follow the rule, refuses it.

## 5. Measurement

The three stage keys are taken on `pu` first, before any recognizer change:
`speed-stage-precedence-aggregate-longlived`,
`speed-stage-resolve-aggregate-longlived` and
`speed-stage-readings-aggregate-longlived`. They are reported timing keys,
like every `speed-` key: never gated and never compared with the baseline.

Speed is read as chart-recovery's verdict was: three quiet measurements in
one session, each key's median, minimum and maximum, before (on `pu` with the
stage keys) and after (with this decision built), recorded in Records.

Held at their baseline throughout: `oracle-wrong`,
`rd-accepted-chart-refused`, `chart-accepted-both-refused`,
`precedence-disagree`, `precedence-uncovered-unlisted`, `leo-disagree`,
`tables-drift`, `ledger-gaps` and `collapse-over-limit` at 0;
`lex-unmapped` not above 9; `oracle-exact-chart` reported. `items-per-token`
and `set-reuse` are counted over the files recognition accepts, so a change
in them partly reflects which files are counted, and is reported with that
caveat.

## 6. Exit, stops and what the result decides

Exit: decision §4 built, every held key above unchanged, and the
after-reading's median `speed-ratio-aggregate-longlived` at or below 5.

Stops, reported to perigrin and never tuned toward:

- a held key moves, or a forest digest changes;
- a reduce-side resolution the filter cannot apply without the parser state,
  beyond `indirob` before `[`;
- the after-reading's median `speed-stage-recognize-aggregate-longlived`
  above the before-reading's maximum, or its median
  `speed-ratio-aggregate-longlived` not below the before-reading's minimum.

The result decides one thing: perigrin's ruling on whether the shift side,
the value on items (C), is designed next. Round 1's findings that concern
only the value (C1, C4, C5, C8, C9, P1, P2, P3, P10 of the minute) carry to
that design.

## 7. Out of scope

The value on items; removing or changing the forest's precedence
constraint; grammar rewriting, which grows the grammar, Earley's cost
(perigrin, 2026-11-23); speed work on resolve and the readings count;
interleaved lexing and M2. Only `internal/chart` changes, plus this
document and its assessment.
