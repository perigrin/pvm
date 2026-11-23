# crochet:assess round 5: codebase lens

**Subject.** `docs/plans/2026-11-23-chart-stateless-precedence.md`, blob **d97415a147b923df589b27bed3cad0317f6bbfe2** (`git hash-object`). The worktree HEAD is 157b0f820, the minute commit on top of 2782fd181. This is revision 5, "V1 now": a state-free, hash-free judgement in the forest only.

**Prototypes.** The evidence directory's `stateless.diff.gz` and `combined.diff.gz` decompress to byte-identical copies of the diffs I reviewed in round 4 (md5 4e53425c… and 077c8ffb…).

**What I ran.** The combined prototype applied to scratch copies and loaded through `go test -overlay`; the worktree is untouched. Files are in `.../assess-w3/codebase/ov4/`.

## 1. Disposition of C23 to C29

| Id | Disposition |
|---|---|
| C23 | **Moot.** V2 is deferred. Recognition is unchanged, so `print $fh [];` keeps `{Byte:0, Reason:precedence}`. §2 records the V2 effect as the reason for deferral. |
| C24 | **Released** against §3: "What this does not amend. Recognition stays Boolean … Precedence is still applied after the chart, as a constraint on the forest, as v2 §2, §8 and rule 6 have it." That statement is now true of what is decided. |
| C25 | **Released.** The key half is released against §3.1's shift side: "the spine-top item …, the rule-position item before the shift, and the terminal shifted". The guard half is released against §3.2: "returns an error through `NewPrecedence`'s existing error path". |
| C26, C27, C28, C29 | **Moot.** They concern Build, the recognition contract, the Leo check and the recognizer's parent table, all V2. |

I am the holder releasing my own findings, so I may be biased toward finding the bar met. Each release above quotes the text it rests on.

## 2. §3.1: are the keys exactly what V1 uses?

**Yes**, read against `combined.diff`'s V1 (`NewStateless`, `v1C.eval/seq/mark`, `reduces`, `shiftForbid`).

**Reduce side.** `reduces(r, t, pi)` reads `reduce[r*nt+t]`. For the mixed pairs it also reads `parentForbid[mixIdx][pi]`, where:
- `pi = item(r', m-1)` for a right child;
- `pi = item(r', 0)` for the left child of a two-symbol prefix;
- `pi = -1` elsewhere.

I generated V1's mixed list on the committed tables (`TestZZAssess5Mixed`). It has exactly six pairs, as §3.1 says:
- `myterm ::= ( expr )` and `myterm ::= ( )` before `[`;
- `myterm ::= scalar` before `[` and before `{`;
- `indirob ::= scalar` before `[` and before `{`.

**Shift side.** `shiftForbid(pend, item(r, m-1), f)`:
- `pend` is the spine top;
- at the root, `pend` is `item(r, 0)` of each alternative's start rule (`rootTop`);
- `f` is the right child's first terminal.

That matches §3.1. The memo key additionally carries the terminal after the span (eval) and the required first terminal (mark). These are Constrain's own memo dimensions, not judgement keys, so §3.1 is right not to list them.

## 3. §3.2: the guard

**Accurate.**
- `NewStateless` fails when a mixed pair stays mixed by parent item, and when a shift key is mixed by spine top. It currently panics; §3.2 moves that to an error.
- "Proves the tables are well-defined, not that the evaluator supplies the right key" is the correct limit.

**A generation-failure toy is feasible.** A sketch of a reduce-side failure:

```
%left LOW
%left t
S : p E | q E | q F ;
E : A ;
A : e %prec LOW ;
F : e t ;
```

- The parent item `E → . A` sits in both the state after `p` and the state after `q`.
- After `q e`, the state holds `A → e .` and `F → e . t`. It can shift `t`, so it resolves shift (LOW < t) and forbids the reduction.
- After `p e`, nothing shifts `t`, so the reduction is allowed.
- The pair is therefore mixed by parent item, and generation must fail.

A shift-side analogue is built the same way, with one spine top in two states that differ in a reduction on `f`. I did not run either toy.

## 4. §3.4: identical output

**Achievable.**
- Build is unchanged, so the kept set compares by alternative index on the same forest.
- `Expected` depends only on the chart, so it is equal trivially.
- Reason and byte are equal when the verdict is equal, because a precedence refusal is always `{Byte:0, Reason:precedence}` and later refusals depend on the kept set.
- Both digests unchanged is now true by construction.

### C30: Partial. "The port's refused cases" are not precedence-refused fixtures

`testdata/port/refused-where.txt` lists 336 refused cases: 282 `grammar` and 54 `table`, and **none** `precedence`. None of them reaches `Constrain`. The aggregate holds no precedence refusals either (round4-fresh.md: "the aggregate contains no precedence refusals"). So the differential test's refused-path coverage is:
- 22 referee sentences whose bison verdict is refuse;
- the toy referees;
- the 10 refused rows of `waypoint2Refusals`.

The strongest refused-path evidence in §2 is the fresh assessor's 185 precedence-refused operator sentences and 989 refused chain sentences. Those were generated in scratch, not committed. Fix: commit them, or their generator with a fixed seed, as a differential fixture, and drop or relabel "the port's refused cases". If this ships as written, the refused path of the new evaluator is checked on about 30 inputs.

## 5. §3.5: the referee judges production

**No import cycle.** `chart` imports `forest`; `forest` imports neither `chart` nor `cmd/ratchet`. **No circularity:**
- `precedence-disagree` and `precedence-covered` judge production against bison directly;
- the differential test ties production to the oracle;
- oracle-versus-bison keeps the oracle trustworthy where bison has no verdict, which is the aggregate.

### C31: Partial. Where the "test-only" oracle lives is undecidable as written

A `_test.go` file in `forest` is not visible to package `chart`'s tests or to `cmd/ratchet`. Yet:
- the aggregate population is built in `cmd/ratchet` (`aggregateFiles`, package `main`, not importable);
- the perly referee runs in package `chart`.

The oracle cannot move to another package either. It needs `Precedence`'s unexported automaton (`states`, `walk`, `reduces`, `shifts`). Both prototypes put it, and the comparison drivers, in non-test `zz_*.go` files in `forest`.

The decision should pick one of two options:
- (a) the oracle ships as an exported, documented-as-reference function in `forest`, and the differential runs from `cmd/ratchet` and `chart` tests;
- (b) the differential and oracle-versus-bison tests run inside `forest`'s own tests, with the population and the referee sentences read from testdata paths.

Option (b) keeps "test-only" literally true.

## 6. What an implementer of V1 could get wrong

### C32: Low, issue-level. Where the tables are built

Build them inside `NewPrecedence` and store them on `Precedence`. That is the only place the automaton's item lists exist: the prototype had to add `allRules`, `allItems`, `stItems` and `allBase`. It is also the only path that serves every `Precedence`: Parse's, the referee's own `newPrecedence`, and each toy's.

The prototypes did neither:
- stateless used a global `map[*Precedence]` under a mutex;
- combined used `grammar.Derived` keyed by tables, with the committed grammar.

Copying either would leave the referee and the toys on different tables from production. It would also put a lock or map lookup on every `Constrain` call.

### C33: Low, issue-level. Evaluator details that are easy to break

- **Root.** The root's spine top is per alternative, `item(r, 0)` of the start rule. `buildShift` admits a dot-0 top only for the start rule in state 0. Dropping either special case silently changes root judgements.
- **Canonicalisation.** It must drop `pend` only when that top has no forbid entry (`topIdx < 0`), and `pi` only when the node's symbol needs no parent. Dropping more merges memo entries that differ.
- **`pi = -1` on a mixed pair is allowed** (`reduces`: `pi < 0 ||`). That is safe on perly only because a mixed pair's left-hand side (`myterm`, `indirob`) is never the root, nor the left child of a prefix of three or more symbols. Make it an assertion, or a generation check, rather than a silent allow.
- **Empty right children skip the shift check,** as in `Constrain`.
- **Cached slices are shared.** `seq` returns the memoised slice when m is 1. Never append into a cached slice.
- **The per-call memo heads** `make([][]v1Ent, len(f.Nodes))` cost about 72 MB on 3-million-node forests. Allocation is ruled to come later ("get ourselves to 3.92 first"), but the stage key's cost should be read with it in mind.

## 7. What I did not examine

- I did not build or run the generation-failure toys.
- The no-map prototype (`no-map.diff.gz`).
- The other assessors' round-4 reports, beyond the evidence counts quoted.
- The README and ratchet-table text, which is not yet written.
- Timing.
- The 11,574-file and 35,083-sentence runs: I relied on the fresh assessor's report.

## 8. Recommendation: **modify**

Small changes; nothing blocking.

- **C30:** the named refused-path fixtures contain no precedence refusals. Commit the 1,174 precedence-refused sentences, or their generator, and relabel.
- **C31:** state where the oracle lives, so "test-only" and "differential over the aggregate" can both be true.

Issue-level: C32, C33.

§3.1, §3.2 and §3.5 are accurate against the prototype. §3.4 is achievable.
