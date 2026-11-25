# V2 feasibility: can an item's precedence context be "(state, origin, one canonical context)" on perly.y?

Measured 2026-11-25 on pu c383f861 (a scratch worktree, since removed). Answers `builder-check.md` Q1 (all three bullets) and Q2. Nothing was committed.

**Tables.** The committed perly.y tables: 772 split-automaton states, 350 rules, 1,212 core items (1,224 in V1's automaton, which adds the 4 refused rules), 237 terminals plus EOF, and 46 canonical spine tops (items with a `topAt` row). V1's tables hold 674 forbidden (top, at, f) shift triples and 6 parent-dependent ("mixed") reduce pairs.

**Corpus.** In-tree corpora, with `T1_CORPUS=/opt/corpora/PerlOnJava/src/test/resources/unit`: 1,881 files. 1,842 got a forest, 3 were refused by `LibraryOperators`, and 1,839 were judged. That is the same population as design §3. perl.git `t/` was not measured.

**Code.** The scratch code is kept for reproduction in `feasibility-code/` beside this file, with a `.txt` suffix so the module build ignores it. Drop the suffix when copying:

- `zz_feas.go` and `zz_feas_test.go` go in `internal/chart/forest/`.
- `zz_feas_ratchet_test.go` goes in `internal/chart/cmd/ratchet/` as `zz_feas_test.go`.

To run them: `go test ./internal/chart/forest -run '^TestZZ' -v`, then `go test ./internal/chart/cmd/ratchet -run '^TestZZFeasCorpus$' -v` with T1_CORPUS set as above.

## 1. Per-rule reduce conflict inside a state (Q1, third bullet)

**Count: 0.** Only one state reduces two or more rules with the same LHS: state 439, which reduces r271 and r275. Both are `term ::= term ARROW CAST_STAR`; perly.y has that rule twice with the same text. Their V1 reduce verdicts agree on every terminal.

**How it was measured.** For each state in `tab.States`, the `Reduces` list was grouped by LHS. For every pair of rules in a group, and every t in FOLLOW(LHS) (EOF included), the test compared `reduceAt[r*nt+t]`. Where both values are 2 (parent-dependent), it also compared their whole `parentForbid` rows.

**Non-vacuous.** perly.y has 649 same-LHS rule pairs whose verdicts differ for some t. Each pair was put in a synthetic state, and the counter counted all 649.

**Consequence.** `builder.complete`'s per-LHS completion loses nothing. On perly.y the per-rule verdict is a per-(state, LHS) verdict.

## 2. Parent-item conflict inside a state (Q1, second bullet)

**Count: 0.** There are six mixed pairs:

| Rule | Before | Forbidden only under the parent item |
|---|---|---|
| r300 `myterm ::= ( expr )` | `[` | `myattrterm: KW_MY REFGEN . myterm attrlist` |
| r301 `myterm ::= ( )` | `[` | the same |
| r302 `myterm ::= scalar` | `{` | the same |
| r302 `myterm ::= scalar` | `[` | the same |
| r347 `indirob ::= scalar` | `{` | `listop: LSTOP . indirob listexpr` or `listop: FUNC ( . indirob expr )` |
| r347 `indirob ::= scalar` | `[` | the same two items |

60 states hold two or more core items with the dot before the same nonterminal. Only 2 of them, states 105 and 248, wait on a parent-dependent symbol. Both hold `[myattrterm: KW_MY . myterm attrlist]` and `[myattrterm: KW_MY . myterm]`, and neither of those forbids anything. No state holds a forbidding parent item next to a non-forbidding one that waits on the same symbol.

**How it was measured.** In each state, the core items with a nonterminal B after the dot were grouped by B. For every pair in a group, and every (r, t) with LHS(r) = B and `reduceAt` = 2, the test compared `parentForbid[mixAt*items + pi]`.

**Non-vacuous.** Across the whole item set, 18 pairs of items wait on the same parent-dependent symbol and have different verdicts: `indirob` is awaited by 10 items and `myterm` by 3. Each pair was put in a synthetic state, and the counter counted all 18.

**Consequence.** The parent-dependent verdict is a function of the parent *state* (and the rule and t). It can be applied when each parent state item advances in `complete`, keyed by its state. No automaton split is needed.

## 3. Spine-top conflict (Q1, first bullet)

### 3.0 Fact that reframes the question

**Every one of the 674 forbidden (top, at, f) triples has a terminal right after `at`'s dot, and f is that terminal.** No triple has a nonterminal next (0), and no triple has f differing from the next terminal (0). So on perly.y a forbidden shift is always the scan made by the item that holds `at` itself. The builder's "two levels up" case (a predicted item scanning f on behalf of `at`) never forbids anything. The context that matters is the one carried by the state item (S, o) that holds `at` when it scans.

### 3.1 Dynamic measure, on the corpus

**Instrument.** The instrument is a copy of V1's state-free `eval`/`seq`, run from the root on each forest after `LibraryOperators`. Every non-empty shift check `seq` makes is recorded as (rule origin, position j = the right child's start, `at`, canonical top, f). Each record was then attributed to every chart item (S, o) in set j whose state S contains `at`. The slipper's closures (`Chart.Slipped`) are included: without them, 4,380 records went unmatched, all at positions where the slipper inserted a `;`. With them, 0 are unmatched. 8,372 records sit in more than one state item at the same (origin, position). They are attributed to all of those items, which can only overcount conflicts.

**Results.** 574,274 shift records were attributed to 517,358 chart items:

| Measure | Count |
|---|---|
| Items whose shifting core items carry 2 or more canonical tops (`none` counts as one) | 1,246 |
| ...because one core item has several tops (predictors merging) | 1,246 |
| ...because different core items carry different tops | **0** |
| Items whose tops give 2 or more distinct verdict vectors over the item's shifts, so they need copies | 92 (all same-core, 0 cross-core) |
| Over-keep: a top of another core item in the same item allows a shift that every real top of `at` forbids | **0** |

The 92 same-core items are the design's "split where it matters" case. Design §3 measured the same thing per node: 0.24% of nodes met two contexts and 0.010% needed a split.

**Non-vacuous.** `TestZZM3NonVacuous` uses real perly.y verdicts. Under the top `{expr: expr ANDOP . expr}`, the item `[expr: expr . ANDOP expr]` is forbidden to shift ANDOP. The test builds three cases:

- The same item with another core item whose top is `none`. It counts as cross-core, needs copies and over-keeps.
- The same core item under both tops. It counts as same-core only.
- One shared top. It is not counted.

### 3.2 Static measure, over every state (an over-approximation)

**Model.** Every core item of a kernel state item (S, o) belongs to a rule that began at o. That rule was predicted in a predicted state E at o. Its spine top is fixed inside E: it is the kernel item of the predicting kernel state K, or an item of E advanced over a nullable, with the left corner inheriting. The start rules hang from their own start item. The test computed each rule's canonical tops per E, taking the union over every K whose Epsilon is E. It carried those tops along transitions to every state reachable from E, and took the union again over every E that reaches a state.

**Soundness check.** 0 of the 574,274 dynamic records carry a top that the static model lacks for that state and core item.

**Per state.** There are 32,089 (state, core item, top) facts, and 932 core items carry two or more tops. 163 states have two or more tops among their shifting items:

- In 161 of them the extra tops all sit on one core item, the same-core case.
- In 2 of them, states **563** and **660** (after `my \ (`), core items differ. `term`/`subscripted: ( . expr ...` hang from `{myattrterm: KW_MY REFGEN . term}`, while `myterm: ( . expr )` hangs from `none`. In both states every top gives the same verdict vector, so they need no copies and over-keep nothing.
- Only states 275 (`term . OP term`) and 393 (`expr . ANDOP/OROP expr`) need two verdict vectors, and both are the same-core case.

**Per predictor K.** There are 261 predictors (kernel states with an Epsilon, plus the start state). Only 3 mix tops among the rules they begin:

| Predictor | Kernel items | Rules begun hang from |
|---|---|---|
| K=90 | `termdo: KW_DO . term`, `termdo: KW_DO . block` | `{termdo: KW_DO . term}` and `none` |
| K=159 | `term: UNIOP .`, `term: UNIOP . block`, `term: UNIOP . term` | `{term: UNIOP . term}` and `none` |
| K=324 | `myattrterm: KW_MY REFGEN . myterm attrlist`, `myattrterm: KW_MY REFGEN . term` | `{myattrterm: KW_MY REFGEN . term}` and `none` |

For each of the three, a single context gives every shift of every rule begun there the verdict of that rule's real top. That holds over all states reachable at that origin. The context is the predictor's one non-`none` top. Only K=324 has a rule carrying two tops, and their verdicts agree.

The criterion discriminates. Carrying `none` instead over-keeps. One example is in state 275 under K=324: shifting ADDOP after `[termbinop: term . ADDOP term]` is forbidden under `{myattrterm: KW_MY REFGEN . term}` but allowed under `none`. The same happens under K=90 and K=159 for `&&`, `=`, `==` and others.

### 3.3 Limits of this measure

- The dynamic measure only sees shift checks that V1's `eval` reaches from the root, so it misses derivations that never reach the forest. The static model covers all states, but it does not model Leo memos (which skip completions), the slipper, or branched lattices.
- The static union over E and K over-approximates merges: it counts K and E that may never meet at one origin. That makes the zero counts stronger, not weaker.
- Each predicted item's context is a function of the predicting kernel *state* K, not of K's own context. Kernel items always have dot > 0, so each one is a spine top in its own right. Items must still carry it, because S alone does not determine K.

## 4. Terminal sets (Q2)

**Corpus.** 420,935 tokens across all 1,881 files. For each token k, the test took its terminal set, removed `LibraryOperators`' dead (token, terminal) pairs (computed by a copy of its own loop under the pipeline's scope), and added `;` where the set holds `}`, because the slipper may insert one. Then, for chart set k:

- **Reduce.** For every core item that completes there, and every two terminals of the token that are both in FOLLOW(LHS), does `reduceAt` (or the `parentForbid` row) differ?
- **Shift.** For every core item there with dot > 0, and every two terminals of the token in FIRST(next symbol), does `shiftForbidden` differ? This was checked under every canonical top, and under the tops §3.1 observed at that (origin, position, item).

| | Before narrowing | After `libraryOperators`' narrowing |
|---|---|---|
| Tokens carrying any PLUGIN_* terminal | 1,110 | **0** |
| Tokens with 2 or more terminals | 74,343 | 73,547 |
| Tokens whose terminals get different reduce verdicts at a completion the chart holds | **117** | **0** |
| ...differences that only arise through the slipper's `;` | 0 | 0 |
| Tokens whose terminals get different shift verdicts | 0 | 0 |

All 117 unnarrowed differences are `r168 listexpr ::= term` before PLUGIN_ADD_OP versus PLUGIN_LOGICAL_AND_LOW_OP. These are the PLUGIN_* levels (3 LOW to 28 HIGH) disagreeing, and they are what shows the counter is not vacuous.

**Static check over all 1,880 terminal sets.** 1,789 sets have two or more terminals.

- The shift side cannot differ on any set. A forbidden shift is always of `at`'s own next terminal (§3.0), so two terminals can never both be shifted from one `at`.
- `;` and `}` (both at level 0) give the same reduce verdict for every rule.
- Before narrowing, 63 sets have a rule whose verdict differs between two of their terminals that are both in FOLLOW.
- After narrowing to at most one PLUGIN_*, whichever one is kept, exactly **1** set still does: set 1626, `{HASHBRACK, PERLY_BRACE_OPEN}`, for `r347 indirob ::= scalar`. Before HASHBRACK the reduction is always allowed (`reduceAt` 0); before `{` it depends on the parent (2). In the corpus, 102 tokens carry set 1626, and none has `indirob ::= scalar .` complete before it.

**Answer.** If V2 judges after the narrowing, or if the narrowing moves before recognition (it reads only the token stream and the scope), then on this corpus no token position carries terminals with differing verdicts. Statically the only exception is `{HASHBRACK, {}` after `indirob ::= scalar`. It is never reached here, but a recognizer would have to carry "allowed for HASHBRACK only" there or narrow by parent. If V2 judges before the narrowing, the order matters: 117 corpus tokens get different reduce verdicts across PLUGIN_* readings. So a recognizer judging over all eleven readings must either carry per-terminal allowances into the next scan or receive the narrowed sets.

## Verdict

On perly.y, "(state, origin, one canonical context)" is expressible as the design assumes. It needs neither per-core-item contexts nor an automaton split, provided three things hold:

1. **Reduce side.** Judge per LHS (Q1 third bullet: no state reduces two same-LHS rules with different verdicts). For the 6 parent-dependent pairs, judge at each parent state item's advance, keyed by its state (second bullet: no state holds two parent items that disagree).
2. **Shift side.** Every forbidden shift is a terminal scan by the item that holds `at`. That item's core items never need different canonical tops in a way that changes a verdict: dynamically 0, and statically only states 563 and 660, where the verdicts are equal.
   - The context is not literally "the item's spine top". Three predictors (`do`, UNIOP and `my \`) begin some rules under a real top and others under `none`.
   - The rule must be: the predicted item's context is the predictor's non-`none` canonical top (unique per predictor on perly.y), and items reached from it at that origin inherit it.
   - Items split only where different predictors merge at one origin: 92 corpus items, as in design §3.
3. **Terminal sets.** Narrow the `PLUGIN_*` terminals before V2 judges, or carry per-terminal allowances. Without narrowing, 117 corpus tokens disagree; with narrowing, none do. There is one static exception, `{HASHBRACK, {}` after `indirob ::= scalar`, which the corpus never reaches.

**Risks this does not rule out.** Leo memos skipping a refused completion, branched lattices, perl.git `t/`, and grammars other than perly.y. The result rests on perly.y's shape, so a check of it should be pinned (for example, the static §3.2 and §1–2 counters as tests) so that a grammar change cannot silently invalidate it.
