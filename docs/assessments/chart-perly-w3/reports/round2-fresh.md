# Round 2, fresh participant: chart precedence in the recognizer

Subject: `docs/plans/2026-11-23-chart-precedence-in-recognizer.md` at c270bbfa3
(branch docs/chart-precedence-in-recognizer), blob
**68371487a68e3f8c1661a80ea830356228a3a163** (`git hash-object`, live file, tree clean).

I wrote this before reading `docs/assessments/chart-precedence-in-recognizer.md`, and I have still not read it.

## Recommendation: modify

The mechanism is sound, and it fits the code. The exactness claim (§4.1) held on every one of 1.56M
packed alternatives I tested, and §3's bison count reproduces. But one of §4's five promises cannot hold
for this mechanism (§4.2, F3). The speed premise is contradicted by a cheap measurement (F2). Because of
that, the exit and the "what the result decides" clause cannot answer the question they are there to
answer (F1, F2). These are design-level problems, and I would fix them before refinement, not in issues.

## What I ran

- `count.sh /opt/corpora/perl5` with bison 3.8.2. The output matches §3: 1,347 pairs and 57 states.
  Shift: 669 global and 1 state-dependent (rule 370 before `PERLY_BRACKET_OPEN`, reduce@153). Reduce:
  667 global. Error: 10 global.
- A scratch probe: the repo exported to
  `/tmp/claude-0/-home-user/f57559af-f6d7-50b1-b4d1-2c3218e70a25/scratchpad/assess-w3/fresh/repo`, plus a
  test file `internal/chart/forest/zz_reduceside_test.go` in that copy only. It does the following:
  - It builds the "never reduces" table on `forest.NewPrecedence`'s own automaton, as §4.4 requires. A
    pair (r,t) counts when at least one state resolves the conflict on r and t as shift or error, and
    every state with t in r's LALR(1) lookahead is such a state.
  - It builds unfiltered forests and marks each symbol-node alternative whose rule is refused before
    every terminal of the next token. That is §4 and §4.5. It then marks dead the alternatives and nodes
    that depend on the refused ones.
  - It compares that set with what `Constrain` drops, checks the verdict, and times `Constrain` on three
    forests: the original, the B-pruned one, and Constrain's own output.
  - Sample: 1,676 files, of which 1,662 were accepted. They are std5k.pl, the guacamole `parses/*.pl`,
    and every 4th `.pm/.t/.pl` file under 150 KB in /opt/corpora/PerlOnJava. This is **not** the
    ratchet's aggregate. The machine was not quiet (load 0.7–1.1) and each figure is a single run.
  - Run it with `PROBE_CORPUS=/opt/corpora/PerlOnJava go test ./internal/chart/forest -run TestZZReduceSideProbe -v`.

Probe output (stable across runs; timings are given as ranges over runs):

| quantity | value |
|---|---|
| conflict pairs on the chart's automaton, resolved shift / error | 629 / 8 (bison: 670 / 10) |
| global never-reduce pairs, shift / error | 628 / 8; the one state-dependent pair is chart rule 347 (indirob) before `[` |
| packed alternatives | 1,558,495 |
| refused directly by B / with dependents | 21,822 / **108,427 (7.0%)** |
| dropped by `Constrain` | 164,085 (10.5%) |
| alternatives B refuses that `Constrain` keeps | **0** |
| verdict flips | **0** |
| `Constrain` time, B-pruned vs original | ×0.87, ×0.90, ×0.94, ×0.97 (4 runs) |
| `Constrain` on its own output (every alternative it drops removed beforehand) vs original | ×0.94, ×1.00 |
| share of `ParseWith` time in `Constrain` | about 46% (about 0.95 s of 2.08 s) |

## Findings

### F1. §6 Exit: the threshold of 5 cannot tell B's effect from noise (direction)

The baseline reading of `speed-ratio-aggregate-longlived` is already 4.7915 (one run, §3). The two
checkpoint-3 readings in `testdata/ratchet/checkpoint3.txt` are 5.1765 and 5.1051 (c4eb1e46). Records
show the host drifting between sessions on one commit: recognize read about 0.77 and about 0.90 on
a5498fa4, and the quiet stage sum fell from 3.638 to 3.13 after a host move. So the before-reading could
come out on either side of 5 without any code change.

Consequence if it ships: the exit can pass with B doing nothing, or fail because of drift. Only the stop
"after-median not below the before-minimum" tests the change.

Fix: make the exit relative. Either take before and after interleaved in one session, or keep 5 as
perigrin's goal but say plainly that it is not the test.

### F2. §3, §6: the speed premise does not follow, and the result cannot decide C (direction; the central finding)

§3's own evidence points away from the premise:
- C1 found items barely moved (4.302 to 4.242 per token) with both sides of precedence applied.
- The 47.9% figure is the operator share of packed alternatives, not the share precedence removes.

The probe:
- `Constrain` removes 10.5% of alternatives, and B removes 7.0% (two-thirds of that).
- `Constrain` on a B-pruned forest takes ×0.87–0.97 of its original time.
- `Constrain` on its own output takes ×0.94–1.00. That run removes every alternative precedence will
  drop before the constraint starts, so it is the best case for B and C together, the value on items
  included.

The constraint's cost is in judging the forest it **keeps**, in every edge context (`eval`/`mark`
memoized per (node, state, terminal)). §4 keeps the constraint running unchanged over that forest as
the authority, and §7 puts changing it out of scope. So no amount of earlier filtering can remove more
than about 5–10% of the constraint's cost, which is about 3–5% of the pipeline.

On top of that, B can at best shrink the forest build by a few percent (it removes 7% of alternatives),
and it adds a check per completion to recognition. Resolve and the readings count read the constrained
forest, which §4.1 keeps identical, so they cannot change.

Expected effect on the target: 4.79 falls to roughly 4.6. That is the same size as the drift in F1.

Consequence if it ships: seven days buys a few percent. §6's question ("is the shift side, C, designed
next?") gets an answer only through noise, while the same bound already limits C's speed value as long
as `Constrain` runs.

Fix, one of:
- (a) Give perigrin this bound now, and let him rule on B and C from it.
- (b) Aim the speed work at the constraint's own evaluation cost. Examples: skip subforests that reach
  no conflict state, or cut the contexts `eval` visits. This changes the constraint, so §7's scope moves.
- (c) Keep B for what it does buy (a 7% smaller forest), and state the speed expectation as a few
  percent, without an exit that depends on it.

My timings are from a non-quiet machine on a non-aggregate sample. The direction should be confirmed
with a quiet reading, but a ×0.94–1.00 ideal is far from what the exit needs.

### F3. §4.2: the contract is false for this mechanism, and the named oracle cannot test it (codebase, architecture)

"The recognizer accepts exactly what the reference accepts and the precedence constraint leaves a
derivation for" would require recognition to do the whole constraint. B only does the reduce side. For
`$a <=> $b <=> $c;` (`TestParseAppliesPrecedence`'s input), B refuses the left-nested completion, but
the right-nested derivation survives. The probe's root stays live, so recognition accepts and
`Constrain` refuses. `1 .. 2 + 3 .. 4;` behaves the same way.

The provable contract is a sandwich:
- reference-accepted ∩ constraint-accepted ⊆ recognizer-accepted ⊆ reference-accepted, or more strongly,
- the B-pruned chart's forest after the constraint equals the unpruned forest after the constraint.

"The fuzz oracle holds that" does not work as written:
- `recognize/fuzz_test.go`'s toy grammars declare no precedence, so the filter does nothing on them.
- `recognize` cannot import `forest` (`forest` imports `recognize`).
- The constraint runs over a forest built from the chart under test. That falls short of v2 §3 rule
  11's outside reference.

Workable alternative: use a `NoLookahead` chart (which §4.3 makes filter-free) as the unpruned side,
and test the sandwich in `forest` over toy grammars that carry levels. The decision should name this
test.

Consequence if it ships: an implementer writes the equality test, it goes red on the first nonassoc
chain, and the contract gets rewritten during execution.

### F4. §4.1, §4.3, §5: the tests whose contracts move are not named (codebase; partly issue-level)

Recognition acceptance hardly changes: B killed no root among 1,662 accepted files, and the refused
examples keep a derivation. So the refusal reason stays `precedence`, which is good.

The sets do change, though, and these hold sets or acceptance to the reference or a golden:
- `TestReferenceRecognizer`: acceptance, `Inserted` and `Expected` at every node, under every Options
  combination including the default.
- `TestReferenceAgreesOnTruncatedFiles`.
- `TestSetDigestGolden`.
- `FuzzRecognize`.

`Expected` re-closes through `b.complete(nil)`, so if the filter is keyed on `la` it is skipped there.
But the origin sets that closure reads were themselves pruned. §4.1 names only the forest digests.

Consequence: §6's "a held key moves" stop becomes ambiguous, and goldens get regenerated ad hoc.

Fix: name the set digest as moving by design, and state the `Expected` contract. The rest belongs in
issues.

### F5. §4.1 exactness: supported (codebase, positive)

The probe found 0 B-refused alternatives that `Constrain` keeps, and 0 verdict flips.

The argument for why it holds: in a derivation the grammar allows, reducing r before t in LR(0) state s
puts t in LALR(1) la(r, s). A global pair means every such s is a conflict resolved as shift or error,
and `Precedence.reduces` refuses exactly those.

This holds only while the table is computed on `NewPrecedence`'s own automaton, as §4.4 says, and not
copied from bison's report. The decision should state this argument, because the subset claim rests on
it. The probe did not model contexts with an inserted `;`.

### F6. §3, §4.4: the figures are bison's, but the table will be the chart's (codebase, minor)

On the chart's automaton there are 629 shift and 8 error conflict pairs. Bison has 670 and 10. The
difference is about the 44 `POSTJOIN` lines in the committed `testdata/precedence/bison-conflicts.txt`.

§3 also leaves out the 10 error resolutions, although §4's "never reduces" covers them. On the chart's
automaton, 8 of them are live, and the comparison chains depend on them.

The committed `bison-conflicts.txt` lets §4.4's check run without bison at `make` time; say so.

### F7. §6 stops: the recognize-stage stop is likely to trip for reasons unrelated to the value (direction, risk)

By C1, items barely move, so B should save recognition almost nothing, while it adds a lookup per
completion. The stop compares the after-median with the before-maximum. Records show recognize drifting
about 15% between sessions on one commit.

"Three quiet measurements in one session ... before ... and after" does not say whether before and
after are in the same session. They should be, interleaved.

### F8. §4.5: implementation granularity (issue-level)

- `recognize.complete` completes per (state, LHS) (`g.lhs[state]`), not per rule. So the table the
  recognizer uses is keyed by (state, LHS, terminal class), and the forest drops facts per rule. Today
  only one state×LHS has several reducing rules, and none has a mixed refusal. A test should guard that.
- Leo memos are keyed by (node, symbol) and do not depend on `la`. So each use must check every link
  against `la`, or the memo must carry a per-link summary.
- At a slipper node, `la` holds the inserted `;` besides the token's terminals, and "every terminal"
  must include it.

### F9. §5: the precedence and resolve stage keys need fresh inputs (issue-level)

`Constrain` and resolve mutate the forest in place. Each timed pass therefore needs its forests cloned
outside the timed span, the way `measureStages` prepares lattices and charts.

### F10. §2: rule 6 should be on the list of what is amended (architecture, minor)

§2 amends v2 §8 and §2. V2 §3 rule 6 also places the constraint "after the chart", so it should be
listed too. Rule 12 is untouched, as the decision says, and the 2026-11-14 ruling on the spec's
standing (in chart-perly-w2-inputs' Records) permits the amendment.

## Checked and found as cited

- Baseline keys 4.7915, 7.2212, 0.6551, 0.7903, 0.6620.
- C1's figures, 4.302 → 4.242 and 47.9% (chart-recovery Records, 2026-11-13).
- D7's quote (chart-recovery Records, 2026-11-14).
- The "in-volatile" ruling, which is in chart-perly-w2-inputs' Records, not chart-recovery's.
- The postmortem's 115 ms and 25 ms (`docs/postmortems/chart-perly-w2.md:229`).
- count.sh's counts.
- The ratchet's chart side includes the constraint, resolve and the readings count (`pipelineSide` →
  `forest.ParseWith`).

## Not examined

- No B implementation in the recognizer, so no item-count or recognize-time effect was measured.
- Quiet-machine timings, and the ratchet's own aggregate of 1,391 files.
- Inserted-`;` contexts and Leo in the probe.
- Resolve and readings costs, beyond reasoning that they are unchanged.
- Interleaved lexing.
- The round-1 minute and the assessment file.
- The rulings of 2026-11-23 against any record, since they are not in Records yet.
- The v2 design outside §3 and §8.
