# Chart precedence V2: judge precedence in the recognizer

**Status:** proposed, for crochet:assess. Revision 2, 2026-11-25.

**Follows:** `2026-11-23-chart-stateless-precedence.md` (V1, accepted at revision 9 and delivered as milestone chart-perly-w3). V1 deferred "the recognizer's reduce-side filter (V2)" to this decision.

**Rulings:** perigrin's, recorded verbatim in `docs/assessments/chart-precedence-v2/rulings.md`.

## 1. Intent

V1 made the precedence judgement state-free and cheap. On a quiet B A B A B A reading, `speed-ratio-aggregate-longlived` fell from 5.17 to 3.53, and the precedence stage from 2.34 to 0.68. But the judgement still runs as a separate pass over a fully built forest, and that forest still holds every derivation precedence will discard.

V2 moves the judgement into the recognizer. A derivation bison's precedence would never build is never completed, so it never reaches the forest. The forest arrives already judged, and production stops running the pass.

The purpose is speed while giving V1's answers. V2 changes how precedence is applied, not what it decides.

## 2. Goals, non-goals, anti-goals

### Goals

- Prune forbidden reductions (the reduce side, "B") and forbidden shifts (the shift side, "C") during recognition.
- Production's `ParseWith` no longer calls `(*Forest).Constrain`. V1's `Constrain` and `ConstrainReference` become test-only referees.
- On every input, the built forest equals the forest V1 produces after its pass. This holds node for node and alternative for alternative, over every in-tree corpus file and perl.git `t/`.
- `speed-ratio-aggregate-longlived` improves on V1 (§7).

### Non-goals

- Changing what precedence decides. Decision §6's correctness items stay out; this is the orchestrator's working assumption, and perigrin did not object to it. Those items are the referee's `shifts` defect, pinned by `TestPrecedenceReferenceKnownDivergence`, and the nullable-shift limit, pinned by `TestPrecedenceNullableShiftKnownLimit`. Both keep their current answers and their pins.
- Allocation work and the lex, forest, resolve and readings stages. perigrin ruled that they wait for V2's verdict reading.
- Retaking `baseline.txt` inside the V2 milestone. It happens before V2, as its own change (§8).

### Anti-goals

V2 fails, however fast it is, if any of these happens:

- **An answer key moves.** These are the 27 of the 29 non-timing keys other than `items-per-token` and `set-reuse`. They include `tables-drift`, `leo-disagree`, `oracle-wrong`, `oracle-exact-chart`, `rd-accepted-chart-refused`, `chart-accepted-both-refused`, `precedence-disagree`, `precedence-covered`, `precedence-uncovered-unlisted`, the tdt keys, the superset keys and `readings-collapsed`.
- **A file's verdict changes.** That means a file accepted or refused differently from V1.
- **The two chart-shape keys rise.** `items-per-token` and `set-reuse` describe the chart's shape and move by design, since pruning removes items. They may fall. They may not rise.
- **A differential finds a forest that differs from V1's.**
- **The recognize stage rises above 1.0.**

## 3. Evidence

### The cheap half was tried

The combined prototype (`docs/assessments/chart-perly-w3/prototypes/combined.diff.gz`) added only reduce-side filtering to recognition and Build, and still ran the pass. Round 5 measured V0 at 5.67, V1 at 4.12 and V2 at 4.10, so V2 gained about 0.5% over V1 (`reports/round5-fresh.md`).

### The ceiling

`measurement/constrain-measurement.md` shows that pruning only the reduce side removes 5.6% of packed alternatives. Removing everything the pass drops removes 16.2%, taking 2,565,144 down to 2,150,267.

### Contexts rarely meet

From `docs/assessments/chart-precedence-v2/context-split.md`, over 1,839 corpus files and 1,410,324 judged nodes:

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
| Tokens whose terminals get different verdicts after `libraryOperators`' narrowing | 0. There were 117 before narrowing, and 1,110 tokens carried `PLUGIN_*` readings. |

The one conflicting terminal set left, `{HASHBRACK, {}` after `indirob ::= scalar`, never occurs in the corpus.

The tables do not cover three things:

- Leo memos skipping a refused completion;
- branched lattices;
- perl.git `t/`.

The result also holds only for perly.y's current shape. §5 turns the static counters into tests.

## 4. The design

### 4.1 Where each side is judged

- **Reduce side (B), at completion.** When an item completes a rule before lookahead terminal `t`, the completion is refused if V1's reduce verdict for (rule, `t`) forbids it.
  - For the six pairs whose verdict depends on the parent item (indirob and myterm before `[` and `{`), the parent item is part of the verdict.
  - Per §3, the verdict is uniform within a (state, LHS) and within a parent state. So it can be decided per completion, without splitting states.
- **Shift side (C), at the scan.** When an item scans terminal `t`, the scan is refused if V1's shift verdict for (spine-top item, rule-position item, `t`) forbids it. Per §3, the scanning item holds the rule-position item.

Both verdicts are V1's generated tables, read as they are. V2 adds no precedence logic and does not change the tables. V1's no-map rule carries over to the hot path in recognition.

### 4.2 The context an item carries

The shift side needs each item's **spine-top item**: the item its node's left spine hangs from, or the start item at the root. This is V1's definition.

- A child predicted at rule position 0 (the left corner) inherits its parent's spine top.
- A child predicted at a later position takes the predicting item as its spine top.

Items become **(state, distance, context)**. The context is the canonical spine top, where any spine top that forbids no shift collapses to `none`, as V1's `canon` already does. So a context is one of V1's 46 spine tops, or `none`.

**One rule from §3.** A predicted item takes its predictor's one non-`none` spine top as its context, even if the predictor starts some rules under `none`. This happens at three predictors: `do`, `UNIOP` and `my \`. Taking the real top is exact. Taking `none` would keep shifts that V1 forbids.

This is the departure from recognizer rule 2 ("items hold no context facts"). perigrin ruled that rule 2 is lifted for the precedence value only.

### 4.3 Collapsing contexts

An item carries `none` wherever its context cannot change a verdict reachable from it. That keeps equal sets interning equal and keeps the transition cache hitting. §3 measures how often a context actually matters: 0.01%.

### 4.4 What else changes

- **Library operators, before recognition.** perigrin ruled "Narrow before recognition". `libraryOperators`' dead (token, terminal) pairs are worked out from the token stream and the scope, and applied to the lattice before recognition. A scope refusal keeps its reason `scope`, but its byte may move. A file's verdict must not change.
- **Refusal records beside the sets.** perigrin ruled "Lift rule 12 for refusals". Recognition records which parent each completion was allowed for. Build reads that record and never pairs a completion with a parent the recognizer refused. README rule 12 is lifted for these records only.
- **Precedence refusals.** perigrin ruled "Keep 'precedence'". The recognizer records when a set died only because of a precedence refusal. It reports reason `precedence` at the byte where the chart died, not at V1's byte 0. `refused-where.txt`, `TestParseRefusalMatchesRecognition` and `pipelineReason` follow this rule.
- **Set interning and the transition cache** (`sets.go`, `cache.go`). Their keys include the context, and their values include the refusal records.
- **Leo** (`leo.go`, `leo_memo.go`). A Leo chain must not skip a completion that precedence refuses, and its memo carries the context. `leo-disagree` stays at 0. §3 does not cover Leo, so this needs its own test (§5).
- **`Expected`** (`expected.go`). An error reports the continuations that precedence allows. A derivation precedence forbids is not a valid continuation.
- **The guards.** perigrin ruled "Regenerate everything". There is no precedence off-switch, and every guard runs with precedence on:
  - the textbook reference recognizer and `FuzzRecognize` gain precedence;
  - the set digest, the forest digest and the brute-force outside reference are regenerated;
  - each regeneration is its own reviewed commit, so the old golden and the new one can be compared.

### 4.5 Production and the referees

`ParseWith` stops calling `(*Forest).Constrain`. `Constrain` (V1's evaluator) and `ConstrainReference` (the old judgement) stay as test-only code beside the differentials.

The forest-equality differentials need V1's output to compare against: the old unpruned recognizer, then Build, then `Constrain`. Under "Regenerate everything" there is no off-switch in production, so the old recognizer survives as a test-only copy, the way `ConstrainReference` did in V1.

## 5. Correctness

The guards are written before the recognizer changes.

- **Forest equality.** V2's forest is compared with V1's (the old recognizer, Build, `LibraryOperators`, then `Constrain`) on:
  - every in-tree corpus file;
  - every waypoint2 row;
  - every referee sentence;
  - every generated sentence, seed 4 and seed 44.

  The comparison covers the verdict, the reachable node set, and each node's kept alternatives in order.
- **The perl.git reference.** `make chart-precedence-reference` grows the same comparison over perl.git `t/`.
- **Feasibility as tests.** §3's four counters become tests that fail if a grammar change makes them non-zero, so V2's assumptions are checked mechanically rather than remembered. The four are: per-rule reduce conflicts, parent conflicts, context conflicts, and post-narrowing terminal conflicts.
- **Leo.** A test where a Leo chain would cross a completion that precedence refuses. It must agree with plain Earley.
- **The existing differentials stay.** Production against `ConstrainReference` in forest, chart and cmd/ratchet, and the bison toys. "Production" now means the recognizer's forest.
- **The known-limit pins stay green.** V2 does not change their answers.
- **Keys.** The 27 answer keys equal 8c3317079's. `items-per-token` and `set-reuse` do not rise.

## 6. Left to the builder

- How the context and the refusal records are laid out in a set, and how the cache keys on them.
- How the Leo memo carries the context.
- Where the checks sit in `builder.complete` and `builder.scan`, and how they read V1's dense tables without maps.
- Whether the `speed-stage-precedence-aggregate-longlived` key is retired, or kept to time V1's pass as a reference.
- How the test-only copy of the old recognizer is kept: a copy, or a build-tagged path.

## 7. Exit and Stops

The exit is a proposal for assessment to settle, not a ruling:

- a quiet, interleaved B A B A B A verdict reading. Before is pu with V1, after the baseline retake. After is pu with V2;
- `speed-ratio-aggregate-longlived`'s after-median below its before-minimum;
- the 27 answer keys unchanged;
- `items-per-token` and `set-reuse` not risen;
- the differentials green.

Stops, reported to perigrin and never tuned toward:

- a forest differs from V1's on any input;
- an answer key moves, or a file's verdict changes;
- `speed-stage-recognize-aggregate-longlived` rises above 1.0;
- `items-per-token` or `set-reuse` rises;
- a §3 feasibility counter becomes non-zero;
- anything that would need a criterion change, or an amendment of a started issue.

## 8. Before V2

Retake `internal/chart/testdata/ratchet/baseline.txt` on current pu, as its own pull request. It still holds pre-V1 speeds, for example `speed-ratio-aggregate-longlived` 5.0350.

## 9. Process

These carry over from the chart-perly-w3 postmortem:

- Export the chain to `docs/assessments/<milestone>/chain/` after refinement, after chain-review and at close, in small docs PRs.
- Record rulings in `docs/assessments/chart-precedence-v2/rulings.md` on the day they are made.
- Runners execute a prepared driver script.
- Every close sets `T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit`, and `--timeout 1800` where it is accepted.
- A README section touched by several issues has one owner: a final docs issue that rewrites it whole.
- Milestone criteria are keyed on the commit range or the PR set, not on issue ids.

## 10. Approaches considered

1. **Full B + C in the recognizer, with the post-pass retired.** Chosen by perigrin: "1". This takes the whole precedence stage, about 0.68, plus about 16% less forest downstream. The risk sits in the recognizer, and the forest-equality differentials answer it. §3's feasibility check showed that the LR(0)-state items need no automaton split.
2. **Full B + C in the recognizer, with V1's pass kept in production.** Lower risk, but most of the gain is left on the table. This is close to the prototype: 4.10 against V1's 4.12.
3. **A set of contexts per item, never split.** No item growth, but it prunes less and still needs the pass. Splitting costs about 0.01% (§3), so a union saves nothing.
