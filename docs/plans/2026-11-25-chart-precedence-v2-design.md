# Chart precedence V2: judge precedence in the recognizer

**Status:** proposed, for crochet:assess. Revision 9, 2026-11-25. perigrin accepted revision 7 during round 6. Round 6's last report then raised three blocking findings, and perigrin ruled "Revise to rev 8, run round 7". The acceptance is held until the assessment closes.

**Follows:** `2026-11-23-chart-stateless-precedence.md` (V1). V1 was accepted at revision 9 and delivered as milestone chart-perly-w3. It deferred "the recognizer's reduce-side filter (V2)" to this decision.

**Governed by:** `2026-10-30-chart-parser-v2-design.md`, the governing design. perigrin ruled "Supersede, bounded". For precedence, this decision supersedes these parts of the governing design:

- §8, "precedence leaves recognition (perigrin, 2026-10-30)";
- rule 6's "after the chart". The governing design carries rules 1 to 6 over from `2026-10-29-chart-parser-design.md` §3, among them rule 2, "context … never inside Earley items";
- rule 12's "items stay (state, distance)".

It also supersedes V1's "What this does not amend" on the same two points: items, and precedence after the chart. The §10 kill criterion "Facts in items" is replaced by a stop in §7. `TestPrecedenceIgnoredByRecognition`, which encodes §8, is retired.

**Rulings:** perigrin's, recorded verbatim in `docs/assessments/chart-precedence-v2/rulings.md`. The assessment minute is `docs/assessments/chart-precedence-v2.md`.

## 1. Intent

V1 made the precedence judgement state-free and cheap. On a quiet B A B A B A reading, it moved two keys:

- `speed-ratio-aggregate-longlived` fell from 5.17 to 3.53;
- the precedence stage fell from 2.34 to 0.68.

But the judgement still runs as a separate pass over a fully built forest. That forest still holds every derivation precedence will discard.

V2 moves the judgement into the recognizer. A derivation that bison's precedence would never build is never completed, so it never reaches the forest. The forest arrives already judged, and production stops running the pass.

The purpose is speed, with the same answers as V1. V2 changes how precedence is applied, not what it decides.

## 2. Goals, non-goals, anti-goals

### Goals

- Prune forbidden reductions during recognition. This is the reduce side, "B".
- Prune forbidden shifts during recognition. This is the shift side, "C".
- Production's `ParseWith` no longer calls `(*Forest).Constrain`. V1's `Constrain` and `ConstrainReference` become test-only referees.
- On every input, the built forest equals the forest V1 produces after its pass. This holds node for node and alternative for alternative, over every in-tree corpus file and perl.git `t/`.
- `speed-ratio-aggregate-longlived` improves on V1 (§7).

### Non-goals

- Changing what precedence decides. Decision §6's correctness items stay out. This is the orchestrator's working assumption, and perigrin did not object to it. Those items are:
  - the referee's `shifts` defect, pinned by `TestPrecedenceReferenceKnownDivergence`;
  - the nullable-shift limit, pinned by `TestPrecedenceNullableShiftKnownLimit`.

  Both keep their current answers and their pins.
- Allocation work, and the lex, forest, resolve and readings stages. perigrin ruled that these wait for V2's verdict reading.
- Retaking `baseline.txt` inside the V2 milestone. It happens before V2, as its own change (§8).

### Anti-goals

V2 fails, however fast it is, if any of these happens:

- **An answer key moves.** The answer keys are the 27 of the 29 non-timing keys other than `items-per-token` and `set-reuse`. They include:
  - `tables-drift`, `leo-disagree` and `precedence-disagree`;
  - `oracle-wrong` and `oracle-exact-chart`;
  - `rd-accepted-chart-refused` and `chart-accepted-both-refused`;
  - `precedence-covered` and `precedence-uncovered-unlisted`;
  - the tdt keys, the superset keys and `readings-collapsed`.

  `items-per-token` and `set-reuse` describe the chart's shape and are reported, not gated. Both are taken over the files recognition accepts, a population V2 changes.
- **A file's verdict changes.** perigrin ruled "Accept/refuse only". A verdict is accepted or refused. A `precedence` refusal's byte differs from V1's (§4.4).
- **A differential finds a forest that differs from V1's.**

`speed-stage-precedence-aggregate-longlived` is retired by perigrin's ruling.

## 3. Evidence

### The cheap half was tried

The combined prototype (`docs/assessments/chart-perly-w3/prototypes/combined.diff.gz`) added only reduce-side filtering to recognition and Build, and it still ran the pass.

Round 5 measured the result (`reports/round5-fresh.md`):

| Version | `speed-ratio-aggregate-longlived` |
|---|---|
| V0 | 5.67 |
| V1 | 4.12 |
| V2 (prototype) | 4.10 |

That is a gain of about 0.5% over V1.

### The ceiling

`measurement/constrain-measurement.md` shows two numbers for packed alternatives, starting from 2,565,144:

- pruning only the reduce side removes 5.6%;
- removing everything the pass drops removes 16.2%, down to 2,150,267.

### Contexts rarely meet

`docs/assessments/chart-precedence-v2/context-split.md` measures this over 1,839 corpus files and 1,410,324 judged nodes:

- 0.24% of nodes are reached under two or more canonical spine tops;
- 0.010% are reached under contexts that keep different alternatives;
- at most 3 contexts meet at one node.

### Feasibility on the LR(0)-state chart

The recognizer's items are split-automaton states plus an origin, not single dotted rules. `docs/assessments/chart-precedence-v2/feasibility.md` measures the committed perly.y tables and the corpus. Every counter was shown to count a case built to trigger it.

| Check | Result |
|---|---|
| A state reducing two rules of one LHS with different verdicts | 0. The only candidate, state 439, is the same rule written twice. |
| A state holding parent items whose `parentForbid` verdicts disagree | 0 |
| A state item needing two contexts that give different shift verdicts | 0. Core items with different tops meet in states 563 and 660, after `my \ (`, and every top gives the same verdict there. All 674 forbidden shifts are scans made by the item that holds `at`. |
| Terminal sets whose terminals get different reduce verdicts after `libraryOperators`' narrowing | 1 in the tables: set 1626, `{HASHBRACK, PERLY_BRACE_OPEN}`, after `indirob ::= scalar`. 0 on the corpus. Before narrowing, 117 tokens had such sets, and 1,110 tokens carried `PLUGIN_*` readings. |

Set 1626 was never reached after that completion, in the corpus or in 14 constructed inputs. §4.1 says what happens if it is reached.

The tables do not cover three things:

- Leo memos skipping a refused completion;
- branched lattices;
- perl.git `t/`.

The result also holds only for perly.y's current shape. §5 turns the static counters into tests.

## 4. The design

### 4.1 Where each side is judged

- **Reduce side (B), at completion.** An item completes a rule before a lookahead terminal set. The completion is refused if V1's reduce verdict forbids the rule before every terminal in the set that can follow the rule's LHS. Terminals that cannot follow the LHS do not continue this completion. If the verdict forbids the rule before some of those terminals and not others, the completion is kept.
  - With narrowed input, the only set where this can happen is set 1626 (§3). Reaching it is a stop (§7).
  - The lookahead set always exists. `Options.NoLookahead` withholds it only from the FOLLOW filter, so the reduce side is judged under every option, and the options still do not change acceptance.
  - A recognizer given an un-narrowed lattice applies the same rule. At the 117 tokens of §3 it keeps completions V1 drops. Forest equality is claimed for narrowed input, which is what production gives it.
  - For the six pairs whose verdict depends on the parent item (indirob and myterm before `[` and `{`), the parent item is part of the verdict.
  - Per §3, the verdict is uniform within a (state, LHS) and within a parent state. So it can be decided per completion, without splitting states.
- **Shift side (C), at the scan.** When an item scans terminal `t`, the scan is refused if V1's shift verdict for (spine-top item, rule-position item, `t`) forbids it. Per §3, the scanning item holds the rule-position item.

Both verdicts come from V1's generated tables, read as they are. V2 adds no precedence logic and does not change the tables. V1's no-map rule carries over to the hot path in recognition.

The recognizer judges precedence only for a grammar that passes §3's checks. Generating its tables refuses any other grammar, as `stateFree` already refuses a state-dependent key. perly.y passes. A bison toy that fails any of the checks is judged as §5 says. Examples:
- `emptyred` and `nullableShift` fail because their forbidden shifts are scanned by an item that does not hold `at`.
- `empty` fails because an empty rule has a non-allow reduce verdict. The recognizer has no completion event for a rule whose right-hand side is all nullable, when it spans no tokens: the nullable closure completes it inside the state.

### 4.2 The context an item carries

The shift side needs each item's **spine-top item**. This is V1's definition: the item that the item's node's left spine hangs from, or the start item at the root.

- A child predicted at rule position 0 (the left corner) inherits its parent's spine top.
- A child predicted at a later position takes the predicting item as its spine top.

Items become **(state, distance, context)**. The context is the canonical spine top. Any spine top that forbids no shift collapses to `none`, as V1's `canon` already does. So a context is either one of V1's 46 spine tops or `none`.

**One rule from §3.** A predicted item takes its predictor's one non-`none` spine top as its context, even if the predictor starts some rules under `none`. This happens at three predictors: `do`, `UNIOP` and `my \`. Taking the real top is exact. Taking `none` would keep shifts that V1 forbids.

The context is the only fact an item gains. This is the bounded supersession named in the header.

### 4.3 Collapsing contexts

The context also ties a completion to the predictor it was predicted under. A completion advances only the parent items that predicted it. Example: in `foo -$x * $y;`, `$x * $y` completes under binary minus and not under unary minus.

An item carries `none` only where its context neither changes a verdict reachable from it nor selects among the parents it can complete into. That keeps equal sets interning equal and keeps the transition cache hitting. §3 measures how often a context actually matters: 0.01% of nodes.

### 4.4 What else changes

- **Library operators, before recognition.** perigrin ruled "Narrow before recognition".
  - `libraryOperators`' dead (token, terminal) pairs are worked out from the token stream and the scope.
  - They are applied to the lattice before recognition.
- **Refusal records beside the sets.** perigrin ruled "Lift rule 12 for refusals". The README states the rule in its bullet "No fact is added to an Earley item" (spec §3 rule 12): nothing recognition stores beside the sets is shaped for the forest.
  - Recognition records which parent each completion was allowed for.
  - Build reads that record and never pairs a completion with a parent the recognizer refused.
  - The lift covers these records only.
- **Refusal reasons are diagnosed by rerun.** perigrin ruled "Keep 'precedence'", "Accept/refuse only" and "Diagnose by rerun".
  - When the chart refuses a file, recognition is rerun on that file in the no-precedence mode. The rerun runs first without narrowing, then with it. This reproduces V1's order: `grammar` at the first run's byte, then `scope`, then `precedence`.
  - `lex-unmapped`, and `scope` from resolve, are decided as today.
  - Every reason matches V1's.
  - A `precedence` refusal reports the byte where the chart died, not byte 0.
  - A `scope` refusal from narrowing reports where the rerun's chart died, which may differ from V1's byte. perigrin's ruling "Narrow before recognition" says that byte may move.
  - Every other byte matches V1's.
  - `refused-where.txt`, `TestParseRefusalMatchesRecognition` and `pipelineReason` follow this rule.
- **Set interning and the transition cache** (`sets.go`, `cache.go`). Their keys include the context, and their values include the refusal records.
- **Leo** (`leo.go`, `leo_memo.go`). A Leo chain must not skip a completion that precedence refuses, and its memo carries the context. `leo-disagree` stays at 0. §3 does not cover Leo, so this needs its own test (§5).
- **`Expected`** (`expected.go`). An error reports the continuations that precedence allows. A derivation precedence forbids is not a valid continuation: each candidate terminal is judged as the lookahead.
- **The guards.** perigrin ruled "Regenerate everything" and "Hold golden". No caller of `Parse` can turn precedence off. The no-precedence mode has two uses:
  - in production, it classifies refusals;
  - in tests, it builds V1's forest for the referees (§4.5).

  Every guard over perly.y runs with precedence on:
  - the textbook reference recognizer and `FuzzRecognize` gain precedence;
  - the set digest, the forest digest (`forest-digest.golden`) and the brute-force outside reference are regenerated. Each regeneration is its own reviewed commit, so the old golden and the new one can be compared;
  - `resolved-forest-digest.golden` digests `Parse`'s final forest and is held byte-identical. It checks the goal for free.

### 4.5 Production and the referees

`ParseWith` stops calling `(*Forest).Constrain`. `Constrain` (V1's evaluator) and `ConstrainReference` (the old judgement) stay as test-only code beside the differentials.

The forest-equality differentials need V1's output to compare against, which comes from the old pipeline: the unpruned recognizer, then Build, then `Constrain`. The no-precedence mode that classifies refusals is that unpruned recognizer, so the referees use it and no separate copy is kept.

The bison referee (`referee.judge`, behind `precedence-disagree`) judges what production runs.
- Today it narrows the forest to the sentence's own terminals after Build, then calls `Constrain`.
- Under V2 it narrows the lattice to those terminals before recognition, the same way production narrows library operators. It then judges the recognizer's forest.

So `precedence-disagree` checks the recognizer, and `Constrain` stays test-only.

## 5. Correctness

The guards are written before the recognizer changes.

- **Forest equality.** V2's forest is compared with V1's on these inputs:
  - every in-tree corpus file;
  - every waypoint2 row;
  - every referee sentence;
  - every generated sentence, seed 4 and seed 44.

  V1's forest comes from the old recognizer, then Build, then `LibraryOperators`, then `Constrain`. The comparison covers:
  - accept or refuse;
  - the reachable node set;
  - each node's kept alternatives, in order.
- **The perl.git reference.** `make chart-precedence-reference` grows the same comparison over perl.git `t/`.
- **Feasibility as tests.** §3's counters become tests over the tables, so V2's assumptions are checked mechanically rather than remembered:
  - six counters must be 0:
    - per-rule reduce conflicts;
    - parent conflicts;
    - context conflicts;
    - forbidden shifts scanned by an item that does not hold `at`, meaning shifts whose terminal is not `at`'s own next symbol;
    - rules whose right-hand side is all nullable, with a non-allow reduce verdict before any terminal;
    - core items of one state, under one context, that scan one terminal with different shift verdicts;
  - the post-narrowing terminal-set counter must list exactly set 1626.

  A grammar change that moves any of them fails a test.
- **Leo.** A test where a Leo chain would cross a completion that precedence refuses. It must agree with plain Earley.
- **The existing differentials stay.** These are production against `ConstrainReference` in forest, chart and cmd/ratchet, plus the bison toys. "Production" now means the recognizer's forest. The exception is every toy that fails §3's checks (§4.1). It stays judged by `Constrain` over the no-precedence mode.
- **The known-limit pins stay green.** V2 does not change their answers.
- **Keys.** The 27 answer keys equal the retaken baseline's (§8).

## 6. Left to the builder

- How the context and the refusal records are laid out in a set, and how the cache keys on them.
- How the Leo memo carries the context.
- Where the checks sit in `builder.complete` and `builder.scan`, and how they read V1's dense tables without maps.
- How the no-precedence mode is selected internally, so that no caller of `Parse` can turn precedence off.

## 7. Exit and Stops

perigrin ruled "Relative only". The exit is:

- a quiet, interleaved B A B A B A verdict reading, comparing two commits:
  - before: pu with V1, after the baseline retake;
  - after: pu with V2;
- `speed-ratio-aggregate-longlived`'s after-median below its before-minimum;
- the 27 answer keys unchanged;
- the differentials green.

These Stops are reported to perigrin and never tuned toward:

- a forest differs from V1's on any input;
- an answer key moves, or a file is accepted or refused differently from V1;
- a §3 feasibility test fails, or any input reaches set 1626 after `indirob ::= scalar`;
- a fix would need any item fact other than the canonical spine top;
- the checks pass but the exit is unmet;
- anything that would need a criterion change, or an amendment of a started issue.

## 8. Before V2

Retake `internal/chart/testdata/ratchet/baseline.txt` on current pu, as its own pull request. It still holds pre-V1 speeds. For example, `speed-ratio-aggregate-longlived` reads 5.0350.

## 9. Process

These carry over from the chart-perly-w3 postmortem:

- Export the chain to `docs/assessments/<milestone>/chain/` three times: after refinement, after chain-review and at close. Each export is a small docs PR.
- Record rulings in `docs/assessments/chart-precedence-v2/rulings.md` on the day they are made.
- Runners execute a prepared driver script.
- Every close sets `T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit`, and `--timeout 1800` where it is accepted.
- A README section touched by several issues has one owner: a final docs issue that rewrites it whole.
- Milestone criteria are keyed on the commit range or the PR set, not on issue ids.

## 10. Approaches considered

1. **Full B + C in the recognizer, with the post-pass retired.** perigrin chose this one: "1".
   - It takes the whole precedence stage, about 0.68, plus about 16% less forest downstream.
   - The risk sits in the recognizer, and the forest-equality differentials answer it.
   - §3's feasibility check showed that the LR(0)-state items need no automaton split.
2. **Full B + C in the recognizer, with V1's pass kept in production.**
   - Lower risk, but most of the gain is left on the table.
   - This is close to the prototype: 4.10 against V1's 4.12.
3. **A set of contexts per item, never split.**
   - No item growth, but it prunes less and still needs the pass.
   - Splitting costs about 0.01% (§3), so a union saves nothing.
