# crochet:assess round 4: codebase lens

**Subject.** `docs/plans/2026-11-23-chart-stateless-precedence.md` at 6dc1d9558, blob **123012c1a5765883ded36de157c8d0e8ac764f0a** (`git hash-object`). This is revision 4, the "V2: both" design.

**Order of reading.** The subject first, then both prototypes: `stateless/stateless.diff` and `combined/combined.diff`. The minute's "Experiments after round 3" and `combined/result-compare.txt` next. The plans diff `3e4d0c43b..6dc1d9558` last.

**What I ran.** I applied `combined.diff` to scratch copies of the files it touches. Those copies are loaded with `go test -overlay`, so the worktree is untouched. The internal tree is unchanged from 8c3317079, so the patch applies cleanly. I ran two things on it:
- a probe test, comparing V0 and V2 on recognition and on the pipeline's refusal reason;
- the chart package's Port, Precedence, Waypoint and Refusal tests, with `recognizeSource` switched to the filter.

Files are in `.../assess-w3/codebase/ov4/`.

## 1. Disposition of C17 to C22

| Id | Disposition | Against |
|---|---|---|
| C17 | **Released** | §3: "Today's `Constrain` is kept as a test-only oracle. A differential test compares verdict and exact kept set per file over the aggregate, the referee and the toy referees." |
| C18 | **Released** | §3.2: "Per-node lists may grow. A test forces several entries onto one node". The prototype's V1 keeps per-node slices (`ev`, `mk`). The multi-entry path already runs on the corpus (97,811 nodes with two eval contexts, `result-compare.txt`), so the forcing test is cheap insurance. |
| C19 | **Released** | §4: "Each timed pass works on a fresh copy of the forest made outside the span". The prototype does this (`Inputs`, `CloneF`). |
| C20 | **Released** | §3.2: "The maps `NewPrecedence` builds once per grammar stay." |
| C21 | **Released** | §5's exit lists all parts. |
| C22 | **Released** | §3.2: "Per-call state lives on the call, never on the shared, cached tables." |

I hold the bar for my own findings, so I may be biased toward finding it met. Each release above is against text I can quote, and the prototype does what the text says.

## 2. §3.1: the context, and the guard

**The context.** This is what each prototype keys on:

- **Reduce side** (V1 `reduces`): the rule r and the terminal t after the span. Six mixed (r, t) pairs are also decided by the parent item: `parentForbid` (V1), `byParent` (stateless).
- **Shift side** (`shiftForbid(pend, rpi, f)`): three things.
  - the spine-top item `pend`;
  - **the item at the rule position**, `item(r, m-1)`;
  - **the shifted terminal f**, which is the first terminal of the right sibling, not the terminal after the span.

§3.1's list ("the rule; the terminal after the span; the parent item; the spine-top item") leaves out the rule position and the shifted terminal. Constrain uses both too. This is C25, low: §3.1 should state the key separately for each side.

**The guard.** It is sound relative to `NewPrecedence`'s automaton. In `NewStateless`:
- the reduce side is gathered over every state that reduces r;
- mixed pairs are gathered per parent item over every state that holds it before r's left-hand side;
- the shift side is gathered per spine top, over every state holding it as a kernel item, each rule on its closure's left-corner chain, and each state after that rule's prefix.

A key that is uniform over all of them gives Constrain's verdict on every derivation that can present it. A key no state presents never arises. The prototype panics inside `grammar.Derived`, and §3.1 promises "refused when the tables are built". The issue should return an error through the existing `prec.err` path, which `parseStream` already turns into a refusal (C25).

`ReduceSideOf`, recognition's table, does not fail on a parent item that stays mixed. It leaves the completion allowed. That is conservative and sound, and it is the right behaviour for recognition.

## 3. §3.3 against what the prototype did

| Claim | Prototype | Verdict |
|---|---|---|
| Refuse when every terminal that can follow is a never-reduce terminal | `refusedBy`: `la & follow[lhs] &^ Never[r]` | Matches. "That can follow" means `la ∩ FOLLOW(lhs)`. |
| Leo memo never completes a refused link | `leo()` falls back to link-by-link when any skipped link is refused or has a parent table (`for l := m; l.next != nil`; the top's own completion runs normally) | Matches. 0 refused links on the aggregate. |
| Lookahead off: the filter does not run | gated on `la != nil`; `Chart.Refuses` returns false under `NoLookahead` | Matches. |
| Forest build "refuses the same completions, except over empty spans" | `build` filters facts with `c.Refuses(r, next)`, which is **`Never` only**. It exempts every fact with origin j, including slipper kernel items that recognition does filter. `leoFacts` stops at a refused link. | **Overstated** (C26, low). Build applies only the rule-level table. It is sound, because Build keeping more is judged by the forest filter, but "the same" is not what was built. |
| `Expected` unchanged | `Expected` reruns `complete(nil)`, and the filter is gated on `la != nil` | Matches. |

## 4. Identical output and the recognition contract

**`forest-digest.golden`:** §3 now says it moves once, under a check against the oracle. That resolves my round-2 C12.

### C23: Blocking. §3 "identical output" claims too much

The text: "For every input, the parse verdict, the refusal reason, … equal today's".

This is false for inputs whose every derivation dies on the reduce side. Measured on the patched prototype (`TestZZAssess4`):

```
"print $fh [];"   recognize V0 true V2 false | pipeline V0 "precedence" V2 "grammar"
```

The ten other probes are unchanged in both. They were `print $fh [1]`, `print $$fh[0]`, `$a <=> $b <=> $c`, `1 .. 2 .. 3`, `$a + $b * $c`, `@$array[1..3]`, `print {$fh} [1]`, `print STDOUT [1]`, `++$x++` and `$a = $b = $c`.

With the filter switched on in `recognizeSource`, the chart's Port, Precedence, Waypoint and Refusal tests fail in exactly one place: `TestPrecedenceIgnoredByRecognition`, at `"print $fh [];" is refused at byte 11 … recognition must ignore precedence`.

The aggregate's "0 reason differences" is true, but no aggregate file has this shape. Fix: state the exception. A file whose every derivation dies on bison's reduce side is refused by recognition, with reason `grammar` and a byte, instead of `precedence` at Byte 0. Name the test that changes.

### C24: Missing. The amendment of the governing design is no longer stated

Revision 2 had "What this decision amends", naming v2 §8 ("precedence leaves recognition") and v2 §2 ("conflicts are kept as ambiguity"). Revision 4 drops that section; `grep -i amend` finds only the recovery-plan citation. But V2 puts bison's reduce side into recognition, and `TestPrecedenceIgnoredByRecognition` encodes v2 §8.

The decision should name both amendments and the test that encodes them. Otherwise refinement will meet the conflict mid-issue, which is §5's "amendment of a started issue" stop.

### C27: Partial. The recognition contract has no named oracle, and §6's test extension checks nothing about it

`TestOracleCatchesSkewUnderOneOption` asserts that `Options` has 3 fields, and checks skews on `toyStmt`, which has no precedence. Adding a bit for the filter there exercises nothing precedence-related. `oracleDisagreement` requires equality with `refRun` under every option combination. That equality holds for the perly corpus sample only because the sample files are all accepted.

What needs saying:
- Name a differential check of the contract over the aggregate and the referee sentences: acceptance with the filter on is a subset of the reference, and a superset of (reference ∧ oracle-Constrain accepts). The prototype measured only "acceptance unchanged on the aggregate".
- Decide whether the filter is an `Options` field whose zero value is off, as in the prototype (a pointer, nil off), or always on from the tables. This decides which tests change (C23) and what "the recognizer's `Options` contract" in §3 says.

## 5. What a careful implementer could get wrong

These are issue-level.

### C28: Low. The Leo check

The check must skip the top link: its completion is made normally. It must also refuse on a parent-tabled link, not only on `Never`. Otherwise the parent-dependent `indirob` completion is made through Leo, where the `indirob ::= scalar` item is a Leo item. Cost is O(chain) per Leo use with the filter on. That is added to the existing walk, and is negligible: the multi-link chains are `optlistexpr`/`else`, and there are 0 refused links on the aggregate.

### C29: Low. The parent-table path

Its key is the **chart** parent state, not an LR item. The prototype refuses only when every item in that chart state awaiting the LHS is uniformly forbidden. An implementer keying by a single item, or by bison's state, gets it wrong.

That path also scans the whole origin set (`src.Items()` with `goTo`), bypassing the postdot index and Leo. Only one rule has a parent table, so cost is negligible, but it should use `awaiting()` for interned sets.

### Also

- `seq` returns eval's cached slice when m is 1. A per-node arena would make that alias dangerous. The prototype's `append` into a fresh `out` is safe, and must be kept.
- `make([][]v1Ent, len(f.Nodes))` per call. Three million nodes means about 72 MB of slice headers on the largest forests. That is fine, but it should be measured against `alloc-per-token`.

## 6. §2's facts

These match `combined/result-compare.txt`:
- items per token 4.7035 → 4.4707;
- packed alternatives 2,565,144 → 2,420,235 (−5.65%);
- at most 10 eval contexts and 24 in all;
- reduce side: 54 rules with `Never`, 1 with `Parent`;
- 6 mixed reduce pairs;
- 0 differences in verdict, kept set, refusal reason and readings, for V1 and V2;
- acceptance differs on 0 files.

Ablation evidence for the parent item: `ablate-mixed-forbid` flips `T1/array.t` and others.

I did not verify the 25,490 toy forests or the timing medians beyond one run's file.

## 7. What I did not examine

- The stateless prototype's toy and referee tests themselves.
- The timing protocol of `run.sh`.
- The other Port fixtures under the filter, beyond the test groups named above. I did not run the whole chart package, recognize's agree/fuzz tests, or forest's tests with the filter on.
- The live README, `Options` and ratchet-table text changes, which are not yet written.

## 8. Recommendation: **modify**

The design matches the prototypes, and the prototypes are exact on everything measured. Two things must change in the text:

- **C23:** restate "identical output" with the reduce-side-refusal exception.
- **C24:** record the v2 §8 and §2 amendments, and the test that encodes them.

Also in the text:
- **C27:** name the contract's oracle, and say whether the filter is an option or always on.

Issue-level: C25, C26, C28, C29.
