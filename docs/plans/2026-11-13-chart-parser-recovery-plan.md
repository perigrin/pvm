# Chart Parser Recovery Plan

*pvm `internal/chart`, after waypoint 2. Revision 2, 2026-11-13.*

Revision 2. It is grounded in the sources themselves this time: Aycock's thesis, Opedal et al., YAEP, Marpa and Chalk were read, not remembered. It answers the assessment's ten held findings and shows what each piece of work does to the figure that stopped waypoint 2.

## What changed from revision 1

- **The recognizer already is Aycock–Horspool.** `builder.complete` is the thesis's Figure 7.5 line for line. The slow part is the thesis's own completion loop. The fix (a postdot index) comes from YAEP and Marpa, not from Aycock.
- **R7, directly-executable code, leaves the plan.** The thesis shows it could not be compiled for a grammar the size of perly.y, and its completion step is the same full walk of the parent set that R1 removes.
- **The lexer workstream follows Schrödinger's Token, not recognizer-driven lexing.** The thesis argues against lexical feedback by name. pvm already scans a token as a set of terminals; the work is to use that more, inside `internal/chart`.
- **The split ε-DFA stays.** No source gives a measured reason to drop it.
- **Measurement comes first, not last** (assessment A1), and the plan now carries a projection onto 3.51 (A8).
- **A new option from perigrin's Chalk direction:** prune by precedence during recognition. The plan evaluates it, names its conflicts with the accepted spec, and recommends a measured spike of one narrow form.
- **The honest headline:** the recognizer and forest work together project to about 2.2–2.6 on the kill figure, before precedence, resolve and readings exist. That is under the old limit with 0.4–0.8 to spare, not comfortably under it.

## Where things stand

Waypoint 2 stopped at the §10 speed kill: forest-timed aggregate long-lived ratio 3.51 against a limit of 3 (milestone Records, 2026-11-13). perigrin's direction since: the limit was arbitrary, the benefits are real, investigate completion against Aycock–Horspool and YAEP, and bring the implementation in line before re-measuring.

| Measure (286,350 tokens)           | Chart      | Recursive descent | Source             |
|------------------------------------|------------|-------------------|--------------------|
| Shared lexer alone                 | 99 ms      | 99 ms             | orchestrator probe |
| Chart lex (lexer plus `lex.Map`)   | 122 ms     | —                 | orchestrator probe |
| Recognize, one recognizer per pass | 301 ms     | —                 | orchestrator probe |
| Recognize, review prototype        | 138–159 ms | —                 | review report      |
| RD parse, its own lexing included  | —          | 210 ms            | orchestrator probe |

**Re-measured for this revision** (my scratch probe, run through a Go build overlay so neither checkout was written; minimum of 7 passes, two runs per tree, load 0.7–1.8). The recognize figure here includes building the lattice, which the orchestrator's probe did before its timer.

| Stage, probe corpus                                              | Baseline 3487d258 | Review prototype              |
|------------------------------------------------------------------|-------------------|-------------------------------|
| RD parse                                                         | 234–236 ms        | 249–257 ms (same code; noise) |
| Chart lex                                                        | 127–132 ms        | 127–142 ms                    |
| Lex + recognize, one recognizer per pass (the spec's long-lived) | 520–569 ms        | 351–412 ms                    |
| Lex + recognize, fresh recognizer per file                       | 568–588 ms        | 436–498 ms                    |
| `forest.Parse` (lex, cold recognize, forest)                     | 761–783 ms        | 653–694 ms                    |
| Forest alone, by subtraction                                     | 193–195 ms        | 155–258 ms                    |
| **Lex + per-pass recognize + forest, over RD**                   | **3.05–3.24**     | **2.28–2.58**                 |

So the prototype's recognizer changes alone move the whole pipeline on this corpus from about 3.1 to about 2.4, with the forest untouched. The forest figure is a subtraction of two minimums and is noisy. This is not the ratchet's protocol and not the kill's corpus.

Garbage collection: the runtime's own counter (`/cpu/classes/gc/total:cpu-seconds`) reads 0.43–0.49 s of collector CPU during an 0.76–0.85 s `forest.Parse` pass, and 0.12–0.14 s during RD's 0.24 s. Collection is heavy on both sides, not only the chart's.

## 1. What the sources say

Each claim names where it was read. Anything not verified is marked unverified.

### Recognizer representation and the ε-DFA (Aycock thesis ch. 3, 4, 7)

- An item is an (LR(0) state, parent) pair. Each LR(0) state is split into a kernel state and a nonkernel state holding its predictions (§4.1.5). pvm's `Item{State, Dist}` and `State.Epsilon` are this.
- Nullable symbols are handled in the predictor: predicting a nullable B also moves the dot over it (§7.2, proved equal to Earley's sets in §7.3). Precomputing that gives the ε-DFA (§7.4, Theorem 4).
- **The thesis's completion walks the whole parent set.** Figure 7.5: `foreach (pstate, pparent) in S_parent: goto(pstate, A)`. It skips items whose parent is the current set. There is no index from symbol to waiting items anywhere in the thesis. pvm's `builder.complete` (`recognize.go`) is the same loop, with `goTo` a binary search.
- The thesis uses no lookahead (§3.2). It found prediction lookahead of little use with LR states: 16% of items were unused on its Java corpus and lookahead could remove at most 19% of those (§4.3).
- Chapter 7 reports **no timings**. The "about 2× over standard Earley" figure the earlier review recalled is not in the thesis. It may be in the 2002 journal paper, which I did not have: unverified.

### DEEP and SHALLOW against Bison (thesis ch. 4)

- **What DEEP is.** One block of C per DFA state, run as direct threaded code through GNU C's labels-as-values, with the threaded code rewritten while it runs and memory taken by mapping oversized regions with guard pages (§4.1.3, §4.1.7).
- **Its completion is still the full walk.** A final item sets `sym = A` and "calls the code for Earley set S_j": every item of the parent set runs its nonterminal switch (Figures 4.1 and 4.7).
- **Speed, expression grammar only** (11 rules, 25 states; 200 MHz Pentium, gcc 2.7.2.3, same flex scanner inside the times). Reading Figures 4.9 and 4.12 from the PDF at 50,000 tokens: Bison about 0.08–0.10 s, DEEP about 0.20 s, SHALLOW about 0.28 s, a standard Earley parser about 0.54 s. So DEEP is roughly 2–2.5× Bison there. These are my readings of plotted points, not printed numbers.
- **It did not scale.** On the Java 1.1 grammar (350 rules) DEEP generated over 19,000 lines of C, 98% in one function, which gcc could not compile in reasonable time (§4.3). The retreat, SHALLOW, interprets the terminal code and keeps only the nonterminal code compiled.
- **SHALLOW on real Java: 2 to 5 times slower than Bison** (§4.3, Figure 4.11, 3,350 files).
- **Code size:** SHALLOW's stripped executables were 1.5× (expression grammar) and 9.0× (Java) the size of Bison's (§4.5).
- The thesis's own verdict on shipping DEEP in its tool: "a definite 'not yet'" (§4.5).

### Schrödinger's Token (thesis ch. 5)

- A token whose type depends on context is returned as a *set* of types. A general parser repeats every type-dependent operation for each type in the set (§5.1, §5.3.2, Figure 5.4b). The grammar is unchanged.
- **The thesis argues against lexical feedback** (§5.2.1): it "couples the scanner and parser tightly", and the scanner can no longer "tokenize the entire input in a tight loop". One of the technique's five listed advantages is "no tight coupling between scanner and parser" (§5.5).
- It also rejects scannerless parsing on two counts: the grammar must spell out whitespace and comments, and a pushdown machine does per character what a finite automaton would (§5.2.7). It notes no timings existed then.
- Where token *boundaries* differ (Fortran's `DO57I`, C++'s `>>`), the scanner emits both sequences, padding the shorter with null tokens the grammar ignores (§5.4.3).
- **pvm already does the first half.** Spec rule 12 says "a token is a superposition". `builder.scan` loops over `cls.terms[e.Token.TerminalSet]`, which is Figure 5.4b. And `lex.Lattice` is a more general form of §5.4.3's padding.

### Early action and safe sets (thesis ch. 6)

- A set is *safe* if it has a final item, no other item shares a final item's last symbol before its dot, and no final item is an empty rule (§6.1, Properties 1–3; Lemma 1).
- At a safe set every final item is on a real derivation (Theorem 1) and the parse is locally unambiguous (Corollary 1). So trees for everything before it can be built at once, and the sets between the last safe set and this one can be deleted (§6.2).
- **Perl is the thesis's worst case.** On Perl 5.6.0's grammar, with its 485 conflicts kept: 61% of sets hold a final item, 31% of those are safe, the mean window is 7.5 tokens, and 59% of sets (70% of items) must be retained. Java retains 12%. The thesis blames right recursion (Tables 6.2–6.5, 6.7).
- Those counts came from an unmodified Earley parser. The thesis notes that LR states, lookahead and set compression all change them (§6.3).
- **Does it let later stages do less?** In principle yes: between two safe sets the forest, precedence and resolve could each work on one small window, and at a safe set no alternative needs searching. Nothing in the thesis measures that. For pvm it is unmeasured, and the memory half does not apply: pvm's sets are interned and shared (21,026 sets for 263,245 nodes), so there is little to delete.

### Opedal et al., "Efficient Semiring-Weighted Earley Parsing"

- **The speedups are grammar-constant folds.** EarleyFast splits predict and complete in two so that a completed nonterminal over a span is one item whatever rule made it, taking O(N³\|G\|\|R\|) to O(N³\|G\|) (§5, Table 1). EarleyFSA replaces dotted rules with states of one automaton for the whole grammar, sharing prefixes and suffixes (§7).
- **Measured on a different regime:** Penn Treebank grammars with 52,009 preterminal rules, sentences of 5–40 words, seconds per sentence, in Cython. EarleyFast is about 20× plain Earley and EarleyFSA a further 2.5× (§8, Figure 4).
- **For pvm's unweighted recognizer most of this is already had.** LR(0) states share prefixes as the automaton does, and a state's completed left-hand sides are grouped. One fold may still apply: complete each (symbol, origin) once per set. `builder.complete` walks the parent once per *item* that completes the symbol. Whether that repeats in practice is unmeasured; Marpa's notes say 268 of the Perl automaton's 271 completing states complete one symbol (`legacy.w`, "Statistics on completed LHS symbols").
- **For weights, the preconditions matter.** The guarantee is "same time and space as unweighted" when semiring operations are constant time (§1), after nullary rules and unary cycles are removed from the grammar (§6, App. E, F), and with a commutative ⊗. A non-commutative semiring needs the automaton form of the grammar (App. K).
- **What that means for the planned stages:**
  - The forest *is* a semiring value: Goodman's derivation semiring, non-commutative (App. A). An SPPF built after recognition is the usual way to hold it.
  - The readings count is the counting semiring, commutative. Over the forest it is a sum of products, exact because perly.y forests are acyclic (chart README, "Cycles"). Done inside the recognizer it would need the nullable rules removed first, which pvm's ε-DFA exists to avoid.
  - Precedence, as §8 defines it (each derivation's exposed operator context at both edges), is a finite relation and so a legitimate, non-commutative semiring. Opedal's route for that is the automaton grammar: the constraint is multiplied into the grammar, not carried beside it.
  - Scope and type facts are not weights on rules at all. They depend on position and declarations.

### YAEP and Marpa

- **YAEP** (`src/yaep.c`, `doc/Internals.txt`, README):
  - Items are dotted rules, not LR states. A set is a core plus a distance vector, each stored once (Internals 1–3).
  - Scan and completion iterate a per-(core, symbol) vector of waiting items (`core_symb_vect_find`, used in `build_new_set`). Duplicates are caught by a stamped array (`sit_dist_insert`).
  - It caches (set, token, lookahead) to at most 3 results (`MAX_CACHED_GOTO_RESULTS`), 70% reuse (Internals 8). Lookahead is worth "almost 2 times" (Internals 9).
  - **It declined the Aycock–Horspool automaton:** "It would complicate the implementation much and, in my estimation, could give only a few percent speedup" (Internals 10). That is an estimate, not a measurement.
  - It declined Leo (Internals 11).
  - **It lexes everything first** (`read_toks`), one symbol per token. Its C grammar handles typedef names by leaving the grammar ambiguous, where YACC uses lexer feedback (README, "Grammar").
  - **Its speed claim:** 2.5–6× slower than YACC parsing only, 1.6–3× with the scanner counted (README). It builds one tree by default (`one_parse_p = 1`), not a forest, and its YACC baseline runs almost no actions.
- **Marpa** (`marpa.w`, `legacy.w`):
  - Each set gets a postdot index of its items and Leo items by symbol ("Postdot item (PIM) code").
  - **It dropped the automaton:** "The finite automaton is not longer in use, but its special items (dotted rules which ignore nullables) remain" ("Aycock-Horspool item (AHM) code"). The sources I read do not say why: reason unverified. Two hints in `legacy.w`: the Perl automaton's 58 predicted states averaged 83.59 items each, and LR states caused duplicate and-nodes when building the forest ("CHAF duplicate and-nodes").
  - Its input model has both lexer styles: the application can ask what is expected (`marpa_r_terminals_expected`), and it can offer several tokens of several lengths at one place ("Read a token alternative").

### Chalk (perigrin's compiler; a design source, read-only)

- **Scanless.** Terminals are regexes matched at each character position; each LR(0) state's terminal map is tried once per position behind a cache. Its own note: 93% of attempted scans fail (`docs/architecture/earley-parser.md`, "Scan Caching").
- **It uses the LR(0) DFA only to cluster prediction and scanning.** Chart items are dotted rules: `chart[pos][core_id][rel_dist] = value`. Completion reads a static list of waiting rule positions per nonterminal (`%_waiting_core_ids` in `Earley.pm` `_complete`).
- **Every item carries a value.** `multiply` runs at each scan and completion and may return zero, which kills the item. `add` merges two derivations of one item by choosing a winner (`FilterComposite` `_filter_compare`; `Precedence.pm` `add` prefers the higher level number). Where no filter prefers either, the two are kept as a packed value (`_pack_survivors`).
- **It needs one parse, and says so:** "By SemanticAction, exactly one parse survives" (ARCHITECTURE.md). The known costs of choosing early are recorded in its own documents: a merge hook to recover state the loser carried, Leo chain extension disabled to keep values equal, and `_fixup_stmts` listed as a violation of its invariants.
- **It parses a restricted Perl.** Indirect objects and bareword resolution are excluded; heredocs, POD and formats are planned as a pre-lex layer outside the grammar; bare regexes are single-line (`ambiguity-classes.md`).
- **Speed:** about 5,000 lines in roughly a second with its XS build (`earley-parser.md`, "Measurement Results"). pvm recognizes `std5k.pl` in 13–19 ms. Chalk is a source of design, not a speed reference.
- It implements Aycock's safe sets, with Property 2 tightened because zero-width terminals break the original test ("Safe-Set Garbage Collection"). pvm's inserted `;` is zero-width, so the same care applies.

### Corrections to the earlier review and to revision 1

| It said                                                                                                               | The sources say                                                                                                                                                  |
|-----------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Revision 1's table: "Completion: YAEP / Marpa / Aycock–Horspool: per set core, an index from symbol to waiting items" | True of YAEP and Marpa. False of Aycock–Horspool: Figure 7.5 walks the whole parent set. pvm was already aligned with the paper; the paper is what is slow here. |
| Review: DEEP "claims speed comparable to bison on Java source, within a small constant factor"                        | DEEP was timed only on the 11-rule expression grammar. On Java it could not be compiled; SHALLOW was 2–5× slower than Bison (§4.3).                              |
| Review: Aycock–Horspool "roughly 2× over their own standard Earley parser on Python source"                           | Not in the thesis. Unverified.                                                                                                                                   |
| Revision 1: "Next state: table lookup in the precomputed automaton" for all three                                     | YAEP and Marpa have no automaton. They advance a dotted rule.                                                                                                    |
| Revision 1: "Lexing: Marpa: the lexer asks the recognizer"                                                            | Marpa allows that and also ambiguous tokens. YAEP lexes up front. The thesis argues for ambiguous tokens and against feedback.                                   |
| Review: YAEP's figure includes "tree building"                                                                        | One tree (`one_parse_p = 1`), not a forest. pvm's forest keeps every reading, so the comparison flatters YAEP.                                                   |
| Neither covered Schrödinger's Token, safe sets, Opedal or Chalk.                                                      | Covered above.                                                                                                                                                   |

The review's code findings and its prototype stand. I re-ran the prototype tree and saw the same direction of gain.

**What the literature expects, in one line.** A practical LR-state Earley recognizer runs 2–5× a table-driven LALR parser on a real grammar (thesis §4.3), and the fastest known Earley parser runs 2.5–6× YACC building one tree (YAEP README). pvm's lex plus recognize is 2.2× a recursive-descent parser that builds a tree. So recognition was never far outside the literature. The forest, kept whole, is the part no source prices.

## 2. Decisions the plan needs first

Nothing merges before D1–D5. The full list, with options and a recommendation each, is in section 8.

- **D1** what replaces the 3× limit. **D2** the spec's status. **D3** where the work lives. **D4** budget and due date. **D5** new and redefined ratchet keys.
- **D6** lexer direction. **D7** precedence pruning during recognition. These authorise spikes; neither blocks the recognizer or forest work.

## 3. Workstreams

Each is a pull request behind the usual gate unless marked *spike*. A spike is scratch work that ends in a written report and is never merged.

**On the criteria below.** Each is a runnable command that fails on 3487d258. None is a time (§9.3). Where a bound is a number, it is provisional: §10 forbids writing a hypothesised count into a criterion, so refinement takes each bound from the prototype measured under the ratchet. A criterion on a test uses `go test -v … | grep -q '^--- PASS: Name'`, because `go test -run` exits 0 when no test matches. A criterion on a key uses an `awk` with an `END` guard, because a missing key must fail.

### P · Prerequisites (first in the chain)

1.  **P1 · Key `Tables.Derived` by type** *(low risk)*

    The slot holds one untyped value (`grammar/lr0.go` `Derived`); a second user gets the first one's value. R4 and F2 are second users. Resolves A5.

    **Criterion:** `go test ./internal/chart/grammar -run '^TestDerivedTwoBuilders$' -v | grep -q '^--- PASS: TestDerivedTwoBuilders'`

2.  **P2 · Commit the set-identity guard** *(low risk)*

    A digest of every set at every node, `Inserted`, sampled `Expected` and the cached count, under five option sets, compared with a committed golden file. It exists in the review's scratch tree (`zz_digest_test.go`). Resolves A10.

    **Criterion:** `go test ./internal/chart/recognize -run '^TestSetDigestGolden$' -v | grep -q '^--- PASS: TestSetDigestGolden'`. Every R issue keeps it green without regenerating the golden file.

3.  **P3 · A brute-force toy grammar with `;` and `}`** *(low risk)*

    The slipper's derivations are checked today only for `{ 1; }` (review Q3). Resolves A6; must land before F2.

    **Criterion:** `go test ./internal/chart/forest -run '^TestForestSlipperBruteForce$' -v | grep -q '^--- PASS: TestForestSlipperBruteForce'`

4.  **P4 · Repair the criteria git-zhi cannot run, and lint new ones** *(needs rulings (D11))*

    Four merged issues (4202, 1b0f, 41c7, 01a1bd9f), milestone criterion 7 (green without its work) and the w1 `speed.go` pin. The lint is one script run at refinement: refuse a criterion with no command, one that passes on the unchanged tree, or one pinned to a file a later issue edits.

    **Criterion:** `sh internal/chart/offline/lint-criteria.sh chart-recovery` exits 0, and exits non-zero on a fixture milestone holding one of each defect.

### M · Measurement (before any R or F issue)

1.  **M1 · Time every stage that exists, each on its own key** *(needs D5)*

    The ratchet's chart side becomes lex, recognize and forest, reported together and separately, beside RD's lex and parse. This is VERDICT's (1f72) speed half, moved forward. It also fixes three review findings a re-measurement depends on: Q2 (`forest.Parse` and `recognizeSource` report `lex-unmapped` by different rules), Q9 (BenchForest can time nothing) and Q12 (the load gate compares load with `NumCPU`).

    **Criterion:** `make -s chart-ratchet | awk '$1=="speed-stage-forest-aggregate-longlived"{f=1} END{exit !f}'`, and `TestSpeedChartSideIsPipeline` passing. Guard: `TestLexOverheadDefinitionUnchanged`.

2.  **M2 · Say what "cold" means now** *(needs D5)*

    Since CTORONCE, "cold" reuses grammar tables an earlier `New` built (review Q1). Keep the key under an honest name and add a truly cold one only if perigrin wants it.

    **Criterion:** the renamed key prints a number and the old name is absent from the ratchet's output.

3.  **M3 · Report allocation and collector CPU per stage** *(low risk)*

    Allocations per token for recognize and forest separately, and collector CPU per stage read from `runtime/metrics` (`/cpu/classes/gc/total:cpu-seconds`) around each timed span. That is a counter, not a profile share, so it repeats. Reported, never gated. Answers A19.

    **Criterion:** `make -s chart-ratchet | awk '$1=="gc-cpu-share-forest"{f=1} END{exit !f}'`

**Every R and F issue then reports its own stage key before and after, under the ratchet's protocol.** The verdict reading is still taken once, at the end. This resolves A1 without breaking "before we re-measure anything": no verdict is retaken early.

### R · Recognizer

1.  **R1 · Postdot index per set core** *(low risk · prototyped)*

    For each interned core, on first use, map each symbol to the items waiting on it and their target states. Scan, completion and the Leo test read it. Branch sets keep the old loop, and the cache's trie records dependencies in a new order; both go in the issue.

    While here, count how often one set completes the same (symbol, origin) twice. If it is not rare, complete each pair once (Opedal's fold, §5.2).

    YAEP `core_symb_vect_find`; Marpa PIM. Not in Aycock.

    **Criterion:** a test that counts `goTo` probes per token over the corpus sample and requires at most a stated bound. Today about 5.6 (1.61M over 286k). `go test ./internal/chart/recognize -run '^TestCompletionProbesBounded$' -v | grep -q '^--- PASS'`. Guards: P2, `TestCacheMatchesNoCache`, the fuzz target.

2.  **R2 · No per-node maps or slices** *(medium risk · prototyped)*

    **The design is the prototype's**, stated here because revision 1 described a different one (A16): find duplicates by scanning, switch to a map at 24 items; carve `in` and `out` from one array; reuse the builder's buffers; copy items only for a new set. `Chart.Slipped` clones, which is where the prototype's one aliasing bug was.

    YAEP `sit_dist_insert` and its object stacks, adapted.

    **Criterion:** `make -s chart-ratchet | awk '$1=="alloc-per-token-chart"{f=1; ok=($2<=3.0)} END{exit !(f&&ok)}'`. Today 11.40 on the ratchet's key. Guards: P2, and a `-race` test on `Slipped`.

3.  **R3 · Make Leo cheap; do not change its default** *(low risk · prototyped)*

    Test eligibility in constant time from the index and store only real memos. Revision 1's "Leo off when lookahead is on" is dropped (A17): Leo is a carried-over component (§5.1), `leo-disagree` is gated, and the lexer work may want lookahead 0, where Leo is what keeps `elsif` chains linear. Changing the default is D9.

    **Criterion:** a test that counts whole-set scans by `memoAt` per node and requires at most a stated bound. Today about 1 per node (263,000 over 263,245). Guards: `leo-disagree 0`, `TestForestFromChart`.

4.  **R4 · Generate the next-state and waiting tables with the grammar** *(low risk)*

    This is where "the grammar is fixed" pays. The per-state part of R1's index, and a dense next-state table, are emitted by `make chart-tables` as data, not built at run time. After P1.

    **Criterion:** the R1 probe test's counter of binary searches reads 0; `tables-drift 0`.

5.  **R5 · Hash each set once in the interner** *(low risk · estimate only)*

    One saved hash in place of varint encoding and three map lookups (`sets.go` `intern`).

    YAEP `set_core.hash`, `dists_hash`.

    **Criterion:** a test counting map lookups per interned set, at most 1 (today 3). Guard: P2.

**Removed: R7, a directly-executable recognizer.** perigrin asked that it not be ruled out, so here is the case in full.

- DEEP's completion runs every item of the parent set (Figures 4.1, 4.7). pvm measured that walk at 1.61M probes for 360k hits. R1 removes the walk; DEEP would compile it.
- On a 350-rule grammar DEEP produced 19,000 lines in one function and could not be compiled (§4.3). perly.y has 348 alternatives and 770 states here.
- DEEP depends on computed goto, on rewriting threaded code while it runs, and on guard pages (§4.1.3). Go offers none of the three. That last sentence is my knowledge of Go, not a source I read for this.
- The measured prize was small: SHALLOW to DEEP is about 0.28 s to 0.20 s on the toy grammar (Figure 4.12).
- 69% of pvm's nodes are served from the transition cache and never run the builder at all.

What a fixed grammar does buy is R4: generated tables. Reopen R7 only if a profile after R1–R5 shows table dispatch as the largest remaining term.

**Removed for now: R6, cheaper transition-cache storage.** Its "bound entries per key" overlaps input 41df, which perigrin ruled out of waypoint 2 on 2026-11-10 ("Both stay out"). It is unprototyped and the riskiest change in the review. It returns only by D10.

### F · Forest

1.  **F1 · Pointer-free, index-based nodes** *(medium risk · unmeasured)*

    `forest.Node` and `Alt` become int32 indexes into slices that hold no pointers, so the collector has nothing to trace. No production caller reads `Node` or `Alt`; `cmd/ratchet/speed.go` uses only `forest.Bench` (A21). It touches no recognizer code, so it runs beside R1. It must land before any precedence or resolve issue, since it changes the API they are written against.

    **Criterion:** `go test ./internal/chart/forest -run '^TestForestNodesPointerFree$' -v | grep -q '^--- PASS'` (reflect over the node and alternative types). Guard: the 2,678-forest node-for-node reference. Reports M3's `gc-cpu-share-forest` before and after.

2.  **F2 · Size and reuse forest memory** *(low risk)*

    Size the first blocks from the file, build grammar-only forest data once (after P1), and reuse recognition's inserted-`;` results (after P3 and R2).

    **Criterion:** `make -s chart-ratchet | awk '$1=="alloc-per-token-forest"{f=1; ok=($2<=B)} END{exit !(f&&ok)}'`, with B taken from M3's first reading. Guard: P3.

### S · Safe sets *(spike)*

1.  **S1 · Count safe sets on pvm's own charts** *(spike)*

    Apply Properties 1–3 (thesis §6.1) to the items inside each state at each node, with Chalk's care for zero-width terminals, over the corpus. Report the share of safe nodes and the window-length distribution.

    Why bother when the thesis's Perl numbers are its worst: pvm's sets are pruned by lookahead, which the thesis says may raise the count (§6.3), and superposed tokens, which should lower it. Nobody knows which wins.

    If windows are short, the forest, precedence and resolve can each be written per window, and a safe node is a place an incremental reparse can stop. If they are long, drop the idea with a number attached.

    **Deliverable:** a report. If it is worth keeping, its merged form is one reported key, `safe-set-share`.

### L · Lexer shape

**This workstream is about maintainability, not the kill figure.** The shared lexer costs the chart what it costs RD (99 ms each). The chart's own mapping is 23 ms, about 0.12 of the 3.51. Superposition moves work into the recognizer, so it can only cost time (spec §7, "What it costs").

1.  **L1 · Type superposition, inside `internal/chart` only** *(spike · needs D6)*

    The token table decides a token's terminals from two previous neighbours and one next (spec §7). Each such decision is a guess made before the grammar is consulted. For three families, drop the neighbour from the key and emit the union instead: `{` as block, hash or subscript; a word as call, string or label; unary or binary operators. Let the grammar collapse them.

    Report, per family: token-table keys and lexer-list entries removed, and the change in `terminals-per-token`, `items-per-token`, `set-reuse` and the cache hit rate. Run it after R1, since its cost depends on R1.

    This needs no lexer export and no `Expected` at lex time, so A2 and A3 do not arise. Lookahead stays on: a wider set is still one set id, and `Recognizer.lookahead` stays cacheable.

    Thesis §5.1, §5.3.2; spec rule 12; YAEP's ambiguous-typedef C grammar.

2.  **L2 · Boundary superposition for `/`** *(scratch spike · cannot merge without a §11 ruling)*

    Regex-or-divide changes where later tokens start, so it cannot be a wider set. The thesis's answer is both token sequences side by side (§5.4.3); pvm's form is a lattice branch that rejoins.

    Two real obstacles, both from the code: `internal/lexer` exports no way to lex from a position under a chosen expectation (`checkpoint.go`), and sets inside a branch are neither interned nor cached, while the forest refuses non-linear lattices (chart README, "Linear lattices only").

    So the spike is a scratch copy of the lexer with one added entry point. It reports how many nodes sit inside branches, how long branches run before rejoining, and the cost with interning off there. Only with those numbers does perigrin decide on the export.

**Why not recognizer-driven lexing, and why not scanless.**

- *Recognizer-driven* (revision 1's L1) is lexical feedback. The thesis's objection is exactly pvm's situation: the lexer could no longer run ahead in one loop (§5.2.1). It also needs the set at a node before the token that the set's lookahead is built from (A3).
- *Scanless*, as Chalk does it, puts an Earley set at every character and tries each expected regex there. Chalk reports 93% of scan attempts failing, and keeps heredocs, POD and formats outside the grammar. For a parser whose problem is speed, that multiplies the number of sets by the mean token length. It also discards the lexer's symbol-table state, which every option still needs.
- Chalk's lesson worth keeping is narrower: its semirings do what toke.c's `PL_expect` does, after the fact (`ambiguity-classes.md`, "Relationship to toke.c"). L1 tests that same move in pvm, with the grammar doing the collapsing.

### C · Precedence during recognition *(spike; perigrin's Chalk direction)*

perigrin's point: Chalk prunes by precedence and type while it parses, it gave up the SPPF only because it needs one parse, and pvm could keep an SPPF of the survivors.

| Form                             | What it is                                                                                                                   | Keeps every surviving reading?                                                                               | Conflicts with the accepted spec                                                                                                                                                                                                                                                           | Effect on interning and the cache                                                                   |
|----------------------------------|------------------------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------|
| **C-a** Chalk's, ported          | A value on each item; `multiply` kills, `add` picks a winner                                                                 | No. `add` discards the loser. Chalk's rule (prefer the higher level) is a heuristic, not bison's resolution. | §8; rule 12 ("items stay (state, distance)"); rule 6 (a weight may order, never decide); §10 "facts in items"                                                                                                                                                                              | Breaks both: two nodes with equal items but different values are no longer one set.                 |
| **C-b** set-valued               | Each item carries the *set* of exposed operator contexts its derivations have (§8's own value), as Chalk's packed survivors  | Yes                                                                                                          | §8; rule 12; §10 "facts in items", squarely                                                                                                                                                                                                                                                | Survives only if the value is interned with the item, which makes it part of the item.              |
| **C-c** compiled into the tables | The same finite context, multiplied into the grammar by `make chart-tables`. The recognizer runs unchanged over more states. | Yes. The forest builder is unchanged and builds only what survived.                                          | §8's first paragraph ("precedence leaves recognition … recognition uses perly.y's rules as written"). Not rule 12: items stay (state, distance) and recognition stays Boolean. §10 only if a state that encodes operator context counts as a fact in an item, which is perigrin's to rule. | None in kind. More states, so possibly fewer shared sets. Size unmeasured. |
| **C-d** type or scope too        | Declared subs, imports, features deciding during recognition                                                                 | —                                                                                                            | Rule 2 (context never inside items); §10                                                                                                                                                                                                                                                   | Fatal. These depend on position, so no two nodes could share a set.                                 |

**How this relates to Opedal.** Opedal's system does attach a weight to every item, as Chalk does, and proves the cost stays within a constant factor. But its weights form a real semiring (⊕ distributes; nothing is discarded), they come from rules, and a non-commutative one needs the grammar in automaton form (App. K). Chalk's filters have a semiring's interface without its laws, which is fine for one parse. §8's precedence value is non-commutative. So the Opedal-shaped way to do precedence in the recognizer is C-c: multiply the constraint into the automaton.

**The dividing line.** A value that is finite and does not depend on position can ride in the recognizer without costing the sharing that makes it fast. Operator context is such a value. Scope and type are not. That is why C-c is worth a spike and C-d is not.

**What it could buy.** The precedence stage would not exist as a pass, and the forest would never hold the readings precedence forbids. How much that is worth depends on how much of today's forest is operator-chain ambiguity, which is unmeasured. The chart README notes operator chains are quadratic in items at lookahead 0 and that real chains are short.

1.  **C1 · Measure before designing** *(spike · needs D7)*

    Two numbers, both from code that exists. First, over the corpus forests, the share of packed alternatives whose rule is an operator rule or carries `%prec`. Second, the state count of a product automaton for a six-level operator ladder and then for all of perly.y's 32 precedence lines, built in scratch by the existing generator.

    If the first is small, precedence in the forest is cheap and §8 stands as written. If the second is large, C-c is dead. Only if the first is large and the second small does C-c become a design for perigrin to accept, with the bison referee (already built: 708 sentences, 716 resolutions) as its test.

**What this means for the split ε-DFA.** C-a and C-b need a value per dotted rule, and an LR state holds many. Chalk keeps its chart by rule position for that reason, and Marpa's notes record LR states complicating its forest. So choosing C-a or C-b would also mean giving up the state representation, and with it items-per-node of 4.46. C-c keeps it. This is the one place the sources bear on keeping the ε-DFA, and it argues for keeping it.

## 4. Projection onto the kill figure

**Units.** Recursive descent's aggregate time is 1.00. The kill reading is 3.51 with the forest and 2.20 without, so the forest is 3.51 − 2.20 = **1.31**. The 2.20 splits by the profile's CPU shares, lex 19 to recognize 45: lex = 2.20 × 19/64 = **0.65**, recognize = 2.20 × 45/64 = **1.55**. (Taking shares of 3.51 directly gives 0.67, 1.58 and 1.23; the two methods agree to within 0.08.)

**Factors, and where each comes from.**

- **R1–R3:** recognize × 0.50 (good) to × 0.70 (cautious). The review measured 0.46–0.53 on recognition alone. My runs today give 0.51–0.69 with lattice building inside the span (224–270 ms against 393–437 ms).
- **R4 + R5:** a further × 0.92. An estimate: `intern` was 17% of the prototype's profile and R5 might halve it. Unprototyped.
- **F2:** forest × 0.65–0.70. The over-engineering review's estimate for its cuts S2–S4 ("30–35% of build time"). Unprototyped.
- **F1:** unknown. No source or measurement prices it. The row below assumes forest × 0.80 and is labelled as an assumption.
- **L, S, C, M, P:** no projected change. L1 may raise recognize.

| Combination                        | Lex  | Recognize   | Forest      | Total (good) | Total (cautious) |
|------------------------------------|------|-------------|-------------|--------------|------------------|
| Today (the kill reading)           | 0.65 | 1.55        | 1.31        | 3.51         |                  |
| R1–R3                              | 0.65 | 0.78 / 1.09 | 1.31        | 2.74         | 3.05             |
| R1–R5                              | 0.65 | 0.71 / 1.00 | 1.31        | 2.67         | 2.96             |
| F2 alone                           | 0.65 | 1.55        | 0.85 / 0.92 | 3.05         | 3.12             |
| **R1–R5 + F2**                     | 0.65 | 0.71 / 1.00 | 0.85 / 0.92 | **2.21**     | **2.57**         |
| R1–R5 + F2 + F1 *(assumed × 0.80)* | 0.65 | 0.71 / 1.00 | 0.68 / 0.74 | 2.04         | 2.39             |
| Floor: lexing alone                | 0.65 | —           | —           | 0.65         |                  |

Arithmetic for the bold row: 1.55 × 0.50 × 0.92 = 0.71 and 1.55 × 0.70 × 0.92 = 1.00; 1.31 × 0.65 = 0.85 and 1.31 × 0.70 = 0.92; 0.65 + 0.71 + 0.85 = 2.21 and 0.65 + 1.00 + 0.92 = 2.57.

**A check against a measurement.** The R1–R3 row predicts a fall of 0.46–0.77. My probe, a different corpus and protocol, measured the whole pipeline falling from 3.05–3.24 to 2.28–2.58, about 0.5–0.9. The two agree in size.

**What the table does not contain.** Precedence, resolve and the readings count are unbuilt, and nothing here prices them. Under the old limit of 3 they would have 0.43–0.79 between them after the bold row. The forest alone took 1.31 when it arrived. So this plan does not show the full waypoint fitting under 3. It shows the two stages that exist coming down by about a third, and it leaves the limit question to D1.

**Caveats.** The shares are CPU, not wall time. Collector work was a further 35% of samples and is not attributed. The prototype was timed on a probe without perl.git, which is 38% of the aggregate. The factors are assumed independent.

## 5. The sixteen open waypoint-2 issues

Resolves A7. "Hold" means untouched and still pending; nothing is cancelled by this plan.

| Issue                                                    | Disposition                                                          | Why                                                                                               |
|----------------------------------------------------------|----------------------------------------------------------------------|---------------------------------------------------------------------------------------------------|
| VERDICT (1f72)                                           | Split. Its speed half is M1 now; its verdict half holds.             | It sat behind READINGS, which is why nothing timed the forest.                                    |
| KEYS (1fa1)                                              | Hold; re-refine.                                                     | Its ruling covers waypoint 2's keys. M's keys need their own (D5).                                |
| Precedence (1dc5, 1df0)                                  | Hold; re-refine after F1 and after C1 reports.                       | F1 changes the forest API they are written against; C1 may change where precedence lives.         |
| Resolve and scope (1e1e, 1e4b)                           | Hold; re-refine after F1.                                            | Same API. Scope stays after the chart under every option above.                                   |
| READINGS (1eb1)                                          | Hold; re-refine after F1.                                            | A sum of products over the acyclic forest.                                                        |
| Oracle (1eda, 1f12, 1fd7), PLUGIN (1f41), PRINTFH (4137) | Hold.                                                                | Downstream of precedence and resolve. PRINTFH is what milestone criterion 7 falsely reports done. |
| EXIT (2009)                                              | Hold; its criterion changes with D1.                                 | The resolution command hard-codes `lim = 3`.                                                      |
| COMPILEFMT (4189), TESTDEDUP (423a)                      | Free to proceed.                                                     | Independent of everything here.                                                                   |
| PERLGITPARITY (1d6a)                                     | Stays stopped. Its drafted split is filed only after the re-measure. | perl.git is 38% of the aggregate; widening the population mid-recovery moves the denominator.     |

## 6. Order

1.  **Decisions D1–D5.** D6 and D7 to authorise spikes.
2.  **P1, P2, P3, then M1–M3.** From here every issue reports its stage key.
3.  **R1 → R2 → R3**, with **F1 alongside**. F1 shares no code with R and is the largest single lever on the forest.
4.  **R4, R5**; **F2** once P3 and R2 are in.
5.  **Spikes, in scratch, from step 3 on:** S1 and C1 at once; L1 after R1; L2 last and only if D6 allows.
6.  **One verdict reading** under M, compared with section 4's projection.
7.  **Then decide** from the spike reports: C-c or §8 as written; the lexer export; whether R6 or R7 has earned a place. Re-refine the held issues against whatever D1 set.

The only hard dependencies are P1 before R4 and F2, P2 before R1, P3 and R2 before F2, F1 before any precedence or resolve issue, and M before everything that claims a speed effect.

## 7. The assessment's held findings

| Finding                                             | How this revision answers it                                                                                                                                                        | Status                             |
|-----------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------|
| **A1** measurement comes last                       | M1–M3 land before R1. Each issue reports its own delta under the ratchet's protocol. One verdict reading at the end.                                                                | Resolved                           |
| **A2** L1 cannot be built in `internal/chart`       | The new L1 is type superposition through the token table and needs no lexer call. The part that does need the lexer (L2) is a scratch spike that never merges without a §11 ruling. | Resolved for L1; L2 waits on D6    |
| **A3** `Expected` depends on the token being chosen | The premise is gone. Nothing reads `Expected` at lex time, and lookahead stays on. L1 reports the costs A3 asked for: nodes affected, items, set reuse, cache hits.                 | Resolved                           |
| **A4** the spec and the kill have no status         | Made a decision of its own, D2, with a recommendation. The plan cannot resolve it; nothing merges until perigrin does.                                                              | Not resolved by design: perigrin's |
| **A5** `Tables.Derived` is one untyped slot         | P1, first in the chain.                                                                                                                                                             | Resolved                           |
| **A6** F2 changes the one untested forest path      | P3 lands before F2.                                                                                                                                                                 | Resolved                           |
| **A7** sixteen issues without a disposition         | Section 5.                                                                                                                                                                          | Resolved, subject to D3            |
| **A8** no projection onto the kill figure           | Section 4, with arithmetic and a measured cross-check. It agrees with the assessor's 2.7 for the recognizer alone.                                                                  | Resolved                           |
| **A9** five decisions missing                       | D2 (spec status), D4 (budget and date), D5 (keys), D6 (the lexer and §11), D11 (criterion repairs).                                                                                 | Resolved                           |
| **A10** no committed set-identity guard             | P2, before R1.                                                                                                                                                                      | Resolved                           |

**The lesser findings.** A11: Q2, Q9 and Q12 are in M1; ponytail S7 and L4 are not taken up and are listed under D12. A12: D1's options now cite §2's latency measure and YAEP's own ratio. A13: the thesis and Opedal are read; R7 no longer rests on memory. A14: D6 asks only to authorise spikes. A15: the over-engineering review's ids are written "ponytail L1" and "ponytail L5" so they cannot be mistaken for this plan's L1 and L2. A16: R2 names the prototype's design and the ratchet's key. A17: R3 leaves the Leo default alone. A18: R6 is out. A19: M3 measures collector CPU with a counter. A20: every P item is an issue with a criterion. A21: F1's wording corrected.

## 8. Decisions for perigrin

1.  **D1 · What replaces the 3× limit**

    - **(a)** Keep 3× on the whole pipeline. The projection reaches it for lex, recognize and forest; nobody knows about the rest.
    - **(b)** A budget per stage in RD units, each stage's issue held to its share (the postmortem's first change).
    - **(c)** An absolute latency: §13's "an edit to a re-parsed tree under 10 ms at p99". That is the measure the chart is wanted for, but it belongs to interleaved mode, which is not built.
    - **(d)** No limit; report only.

    **Recommendation: (b), with numbers set after the verdict reading, not now.** Report every stage from M1 on. After the reading, perigrin names a whole-pipeline figure with the unbuilt stages' shares stated. Whatever is chosen stays a stop reported to him, not a target. The literature supports a looser figure than 3 for a parser that keeps every reading (thesis §4.3: 2–5× for recognition alone; YAEP: 2.5–6× with one tree).

2.  **D2 · The accepted spec's status**

    §2 and §10 say a speed kill closes up-front mode. Options: amend the accepted spec; write a successor decision that supersedes §2's question, §10's speed line and §9.5's key rule; or record a ruling in Records.

    **Recommendation: a successor decision.** An accepted document is frozen; this plan, once accepted, is what refinement carries onto a milestone.

3.  **D3 · Where the work lives**

    Reopen `chart-perly-w2`, or a new milestone. **Recommendation: a new milestone** for P, M, R, F and the spikes. Waypoint 2 keeps its verdict and its sixteen issues as section 5 disposes of them.

4.  **D4 · Budget and due date**

    §14 makes twice a waypoint's budget a stop, and waypoint 2 is due 2026-11-30. **Recommendation:** the new milestone gets its own budget, sized at refinement, covering P, M, R1–R5, F1–F2 and four spike reports. The spikes' follow-on work is not in it.

5.  **D5 · New and redefined ratchet keys**

    §9.5 says keys are never added after the reset. M1 and M3 add keys; M2 renames one. The precedent is the 2026-11-10 ruling "Yes, in one issue". **Recommendation: the same ruling again**, one issue taking every new baseline. R's counters are tests, not keys, to keep this ruling small.

6.  **D6 · Lexer direction**

    - **(a)** Leave it.
    - **(b)** Type superposition inside `internal/chart` (L1).
    - **(c)** Boundary superposition through lattice branches, which needs a new `internal/lexer` export and so a §11 ruling (L2).
    - **(d)** Recognizer-driven lexing.
    - **(e)** Scanless, as Chalk.

    **Recommendation: authorise L1 now and L2 as scratch only.** Decide the export when L2's report exists. Not (d): the thesis argues against it and the code cannot supply it. Not (e): it multiplies the sets and gives up the lexer's state. Neither spike is expected to help the kill figure.

7.  **D7 · Precedence during recognition**

    - **(a)** §8 as written: precedence over the forest.
    - **(b)** C-c: compile the operator context into the tables. Needs §8's first paragraph amended, and a ruling on whether a state that encodes context is a "fact in an item" under §10.
    - **(c)** C-a or C-b: values on items, as Chalk. Needs rule 12 and §10 withdrawn, and gives up set interning and the ε-DFA.
    - **(d)** Scope or type in the recognizer as well.

    **Recommendation: authorise the C1 measurement and decide between (a) and (b) on its two numbers.** Rule out (c) and (d) for up-front mode. perigrin is right that keeping an SPPF of the survivors is not hard, and under (b) it is free: the forest builder does not change.

8.  **D8 · Keep the split ε-DFA?**

    YAEP and Marpa do without it; neither gives a measurement against it. **Recommendation: keep it.** It is why a node holds 4.46 items. Revisit only if D7 goes to (c).

9.  **D9 · Leo's default**

    The review suggested Leo off when lookahead is on. **Recommendation: leave it on.** R3 removes its cost; turning it off buys little after that and would have to be undone for lookahead 0.

10. **D10 · R6 and R7**

    **Recommendation: both out of this plan.** R6 only if the 41df ruling is reopened and a post-R profile names the cache. R7 only if a post-R profile names table dispatch, and then as SHALLOW's shape, not DEEP's.

11. **D11 · Criterion repairs**

    Six: the four unrunnable criteria (4202, 1b0f, 41c7, 01a1bd9f), milestone criterion 7, and the w1 `speed.go` pin FOREST broke. Each earlier repair of this kind was a ruling. **Recommendation:** one ruling, "convert or move, as before", applied in P4.

12. **D12 · Held from waypoint 2**

    Unchanged by this plan and still his: the four PERLGITPARITY split decisions (aligner amendment 4; NOAMP as a fitted terminal; widening `overrides.txt`; admitting 01a1c678 into FEATWORDS); ponytail L5 (the 25-copy lex pass); ponytail L1 (test-only shared-table machinery); ponytail S7 and L4. **Recommendation:** leave all of them until after the verdict reading.

## 9. Risks and unknowns

- **The projection is arithmetic on shares.** Only R1–R3 are prototyped. F1, the largest hoped-for forest gain, has no number at all.
- **The unbuilt stages.** Precedence, resolve and readings have never been timed. The plan's end state is "two stages cheaper", not "the waypoint fits".
- **Collection.** Collector CPU is about half of wall time in every stage I measured, RD included. F1 addresses the forest's part; nothing here addresses the recognizer's beyond R2.
- **R2's buffer reuse** is where aliasing bugs hide. P2 and a race test are its net.
- **L1 may make recognition slower** and reduce set sharing, on which R's gains were measured. That is why it is a spike that runs after R1 and reports those counts.
- **C-c's state count** could be large enough to end the idea. C1 exists to find out cheaply.
- **Scope.** Recovery (P, M, R, F) is one project. The lexer and precedence questions are others. The plan keeps them as spikes so they cannot hold the recovery up.
- **This page is not a decision in the repository.** Until D2 is ruled and a document with criteria exists, refinement has nothing to carry.

### Not verified

- Aycock and Horspool's 2002 journal paper and Leo 1991: not available to me. The thesis covers the same construction but reports no timings for it.
- Why Marpa dropped the automaton: its source says that it did, not why.
- DEEP's ratios to Bison are read off scanned plots.
- That Go cannot express DEEP's threaded code is my knowledge of the language, not a source read here.
- Chalk: I read the four documents in full, `Earley.pm`'s `_complete`, and `Precedence.pm`'s and `FilterComposite.pm`'s `multiply`/`add` paths. I did not read every line of the other semirings.
- My timings: a scratch probe at load 0.7–1.8, two runs per tree, not the ratchet's protocol. `make test` and `make chart-ratchet` were not run.
- The safe-set rate, the repeat rate of (symbol, origin) completions, the operator share of forest ambiguity, and the product automaton's size are all unmeasured; each has a spike or a counter above.
- The quotation "recognition decides what is Perl; the forest decides which reading" is from the brief. The spec's §8 says it as "precedence leaves recognition … recognition uses perly.y's rules as written, is Boolean".

Sources: Aycock, PhD thesis (2001), ch. 3–7 and Figures 4.9, 4.11, 4.12 from the PDF; Opedal et al. (ACL 2023), §1–§8 and App. A, C, G, H, K, L; YAEP `src/yaep.c`, `doc/Internals.txt`, README; libmarpa `marpa.w`, `legacy.w`; Kegler's `ah2002_notes.ltx`; Chalk at e26caaff; pvm at origin/pu 352b3671 (code identical to 3487d258 under `internal/` except two READMEs); the v2 design spec; the `chart-perly-w2` milestone Records; the waypoint-2 postmortem; the assessment; the earlier review and its prototype.
