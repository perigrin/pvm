# Chart Parser Recovery Plan

*pvm `internal/chart`, after waypoint 2. Revision 4, 2026-11-13.*

Revision 4. It applies perigrin's scope for the run and answers the round-2 and round-3 assessments. Section 7 lists each finding from both rounds and what changed. The sources (Aycock's thesis, Opedal et al., YAEP, Marpa and Chalk) were read for revision 2, and that account stands.

## The run's scope (perigrin)

- **In scope:** P1–P5, M1–M3, R1–R5, F1–F2, and the C1 spike. **P5 is this plan's addition, beyond the P1–P4 perigrin scoped** (a forest-identity golden; round-2 finding B-B2). It is his to rule in or out, with D4.
- **Out of scope:** R6, R7, L and S. Each has one line in section 3 and nothing else.
- **C1** carries precedence as a value on items, Chalk-style: multiply to zero, and add keeps every survivor. It is built in a scratch worktree and never merged.
- **Merged code keeps items as (state, distance).**
- **Order:** P1–P3 and M1–M3 first. Every R and F issue then reports its stage key's change.
- **Milestone:** `chart-recovery`, new.
- **Decisions** D1, D2, D4, D5, D7 and D11 go to perigrin before chain-review. Nothing merges before D1, D2, D4 and D5 are ruled.
- **The headline is unchanged:** the recognizer and forest work together project to about 2.2–2.6 on the kill figure of 3.51, before precedence, resolve and readings exist.

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

Garbage collection: the runtime's own counter (`/cpu/classes/gc/total:cpu-seconds`) reads 0.43–0.49 CPU-seconds of collector work during an 0.76–0.85 s `forest.Parse` pass, and 0.12–0.14 during RD's 0.24 s. These are CPU-seconds summed over four cores against wall seconds, so they are not a share of anything; mark workers run in parallel. They show only that collection is heavy on both sides, not only the chart's.

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

### Schrödinger's Token and safe sets (thesis ch. 5, 6)

Both bear only on workstreams that are out of scope for this run (L and S). Revision 2 (commit 186a6af4) holds the full account; nothing below relies on it.

### Opedal et al., "Efficient Semiring-Weighted Earley Parsing"

- **The speedups are grammar-constant folds.** EarleyFast splits predict and complete in two so that a completed nonterminal over a span is one item whatever rule made it, taking O(N³\|G\|\|R\|) to O(N³\|G\|) (§5, Table 1). EarleyFSA replaces dotted rules with states of one automaton for the whole grammar, sharing prefixes and suffixes (§7).
- **Measured on a different regime:** Penn Treebank grammars with 52,009 preterminal rules, sentences of 5–40 words, seconds per sentence, in Cython. EarleyFast is about 20× plain Earley and EarleyFSA a further 2.5× (§8, Figure 4).
- **For pvm's unweighted recognizer most of this is already had.** LR(0) states share prefixes as the automaton does, and a state's completed left-hand sides are grouped. One fold may still apply: complete each (symbol, origin) once per set. `builder.complete` walks the parent once per *item* that completes the symbol. Whether that repeats in practice is unmeasured; Marpa's notes say 268 of the Perl automaton's 271 completing states complete one symbol (`legacy.w`, "Statistics on completed LHS symbols").
- **For weights, the preconditions matter.** The guarantee is "same time and space as unweighted" when semiring operations are constant time (§1), after nullary rules and unary cycles are removed from the grammar (§6, App. E, F), and with a commutative ⊗. A non-commutative semiring needs the automaton form of the grammar (App. K).
- **What that means for the planned stages:**
  - The forest *is* a semiring value: Goodman's derivation semiring, non-commutative (App. A). An SPPF built after recognition is the usual way to hold it.
  - The readings count is the counting semiring, commutative. Over the forest it is a sum of products, exact because perly.y forests are acyclic (chart README, "Cycles"). Done inside the recognizer it would need the nullable rules removed first, which pvm's ε-DFA exists to avoid.
  - Precedence, as §8 defines it (each derivation's exposed operator context at both edges), is a finite relation and so a legitimate, non-commutative semiring. Opedal's route for that is the automaton grammar: the constraint is multiplied into the grammar, not carried beside it. perigrin has ruled that form out (section 3, C).
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
- **It needs one parse, and says so:** "By SemanticAction, exactly one parse survives" (ARCHITECTURE.md). The known costs of choosing early are recorded in its own documents: a merge hook to recover state the loser carried, Leo chain extension disabled to keep values equal (`earley-parser.md`, "Chain Extension (Disabled)"; guarded by `t/bootstrap/leo-graph-equivalence.t`), and `_fixup_stmts` listed as a violation of its invariants.
- **It parses a restricted Perl.** Indirect objects and bareword resolution are excluded; heredocs, POD and formats are planned as a pre-lex layer outside the grammar; bare regexes are single-line (`ambiguity-classes.md`).
- **Speed:** about 5,000 lines in roughly a second with its XS build (`earley-parser.md`, "Measurement Results"). pvm recognizes `std5k.pl` in 13–19 ms. Chalk is a source of design, not a speed reference.

### Corrections to the earlier review and to revision 1

| It said                                                                                                               | The sources say                                                                                                                                                  |
|-----------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Revision 1's table: "Completion: YAEP / Marpa / Aycock–Horspool: per set core, an index from symbol to waiting items" | True of YAEP and Marpa. False of Aycock–Horspool: Figure 7.5 walks the whole parent set. pvm was already aligned with the paper; the paper is what is slow here. |
| Review: DEEP "claims speed comparable to bison on Java source, within a small constant factor"                        | DEEP was timed only on the 11-rule expression grammar. On Java it could not be compiled; SHALLOW was 2–5× slower than Bison (§4.3).                              |
| Review: Aycock–Horspool "roughly 2× over their own standard Earley parser on Python source"                           | Not in the thesis. Unverified.                                                                                                                                   |
| Revision 1: "Next state: table lookup in the precomputed automaton" for all three                                     | YAEP and Marpa have no automaton. They advance a dotted rule.                                                                                                    |
| Revision 1: "Lexing: Marpa: the lexer asks the recognizer"                                                            | Marpa allows that and also ambiguous tokens. YAEP lexes up front. The thesis argues for ambiguous tokens and against feedback.                                   |
| Review: YAEP's figure includes "tree building"                                                                        | One tree (`one_parse_p = 1`), not a forest. pvm's forest keeps every reading, so the comparison flatters YAEP.                                                   |
| Neither covered Schrödinger's Token, safe sets, Opedal or Chalk.                                                      | Opedal and Chalk are covered above; the other two in revision 2.                                                                                                 |

The review's code findings and its prototype stand. I re-ran the prototype tree and saw the same direction of gain.

**What the literature expects, in one line.** A practical LR-state Earley recognizer runs 2–5× a table-driven LALR parser on a real grammar (thesis §4.3), and the fastest known Earley parser runs 2.5–6× YACC building one tree (YAEP README). pvm's lex plus recognize is 2.2× a recursive-descent parser that builds a tree. So recognition was never far outside the literature. The forest, kept whole, is the part no source prices.

## 2. Decisions the plan needs first

The full list, with options and a recommendation each, is in section 8.

- **Before chain-review, to perigrin:** D1 (what replaces the 3× limit, with provisional ceilings now), D2 (the spec's status), D4 (budget and due date), D5 (new and redefined ratchet keys), D7 (what C1's report may lead to), D11 (criterion repairs).
- **Nothing merges before D1, D2, D4 and D5 are ruled.** Refinement may write M1, M2 and the exit issue before then; chain-review refuses the chain until the rulings exist.
- **What refinement carries** is this file, committed under `docs/plans/`. D2 decides its standing against the accepted spec.

## 3. Workstreams

Each is a pull request behind the usual gate unless marked *spike*. A spike is scratch work that ends in a written report and is never merged.

**On the criteria below.**

- Each is a runnable command that fails on 3487d258. None is a time (§9.3).
- A criterion on a test uses `go test -v … | grep -q '^--- PASS: Name'`, because `go test -run` exits 0 when no test matches.
- The ratchet prints `chart-ratchet: <key> <value>` (`testdata/ratchet/baseline.txt`). So a criterion on a key tests field 2 for the key and field 3 for the value, with an `END` guard so a missing key fails.
- A bound written `B` is provisional. §10 forbids a hypothesised count in a criterion, so refinement takes each `B` from the ratchet: from M3's first reading, or from the review prototype run under the ratchet.
- A check that already passes today is a guard, never the red criterion.
- Test and key names here are new. None reuses a name a `chart-perly-w2` criterion already holds.

### P · Prerequisites

1.  **P1 · Key `Tables.Derived` by type** *(low risk)*

    The slot holds one untyped value (`grammar/lr0.go` `Derived`); a second user gets the first one's value. F2 is a second user.

    **Criterion:** `go test ./internal/chart/grammar -run '^TestDerivedTwoBuilders$' -v | grep -q '^--- PASS: TestDerivedTwoBuilders'`

2.  **P2 · Commit the set-identity guard** *(low risk)*

    A digest of every set at every node, `Inserted`, sampled `Expected` and the cached count, under five option sets, compared with a committed golden file. It exists in the review's scratch tree (`zz_digest_test.go`), where it logs and has no golden yet.

    **Criterion:** `go test ./internal/chart/recognize -run '^TestSetDigestGolden$' -v | grep -q '^--- PASS: TestSetDigestGolden'`. Every R issue keeps it green. **An R issue that needs the golden regenerated is a stop reported to perigrin.**

3.  **P3 · A brute-force toy grammar with `;` and `}`** *(low risk)*

    The slipper's derivations are checked today only for `{ 1; }` (review Q3). Must land before F2.

    **Criterion:** `go test ./internal/chart/forest -run '^TestForestSlipperBruteForce$' -v | grep -q '^--- PASS: TestForestSlipperBruteForce'`

4.  **P4 · Repair the criteria git-zhi cannot run, and lint new ones** *(needs D11)*

    Four merged issues (4202, 1b0f, 41c7, 01a1bd9f), milestone criterion 7 (green without its work) and the w1 `speed.go` pin. The lint is one script run at refinement: refuse a criterion with no command, one that passes on the unchanged tree, or one pinned to a file a later issue edits.

    - The lint's criterion is its fixture only. Running it on the live milestone at close would re-run every criterion on a baseline checkout and double a verify that already took eight attempts on w1.
    - The w1 pin asserts how many commits touched `speed.go`. M1, M2 and M3 each touch that file. So P4 lands before M1 if D11 is ruled by then; otherwise the repair converts the pin to a check that does not count commits.

    **Criterion:** `go test ./internal/chart -run '^TestCriteriaLint$' -v | grep -q '^--- PASS: TestCriteriaLint'`. The test runs the lint on two fixtures. On the bad one it requires a non-zero exit and one named line per defect; on the clean one it requires exit 0. A missing script fails both halves. (Revision 3's shell form passed when the script was absent, since `sh` then exits 127.)

5.  **P5 · Commit a forest-identity golden** *(low risk; new in revision 3)*

    F1 rewrites every forest node type, and its named guard in revision 2 (a node-for-node reference over 2,678 forests) was run once by FOREST's gate and never committed. The tests that are committed are written against `*Node`, so F1 would rewrite its own guards.

    P5 commits a digest of each corpus-sample forest taken from a text dump, in a fixed order: for a symbol node its symbol and span; for a rule-position node its rule, dot and span; and for each alternative its rule and split. The dump names no Go type, so it survives F1.

    **Criterion:** `go test ./internal/chart/forest -run '^TestForestDigestGolden$' -v | grep -q '^--- PASS: TestForestDigestGolden'`. F1 and F2 keep it green. **Regenerating it is a stop reported to perigrin.**

### M · Measurement (before any R or F issue)

1.  **M1 · Time every stage that exists, each on its own key** *(needs D5)*

    The ratchet reports lex, recognize and forest on the chart side, together and separately, beside RD's lex and parse. It also fixes two review findings a re-measurement depends on: Q2 (`forest.Parse` and `recognizeSource` report `lex-unmapped` by different rules) and Q9 (BenchForest can time nothing).

    Revision 2 also put review Q12 here (the load gate compares load with `NumCPU`). That gate's threshold is part of a ruled definition (2026-11-11, quiet machine only), so it is **dropped from M1** and listed under D12.

    **Criterion:** `make -s chart-ratchet | awk '$2=="speed-stage-forest-aggregate-longlived"{f=1} END{exit !f}'`, and `go test ./internal/chart/cmd/ratchet -run '^TestSpeedStagesTimed$' -v | grep -q '^--- PASS: TestSpeedStagesTimed'`. Guard: `TestLexOverheadDefinitionUnchanged`.

2.  **M2 · Say what "cold" means now** *(needs D5)*

    Since CTORONCE, "cold" reuses grammar tables an earlier `New` built (review Q1). The key is renamed for what it measures: a fresh recognizer per file.

    **Criterion:** `make -s chart-ratchet | awk '$2=="speed-ratio-aggregate-cold"{old=1} $2=="speed-ratio-aggregate-perfile"{new=1} END{exit !(new && !old)}'`

3.  **M3 · Report allocation per stage, and collector CPU for the whole side** *(needs D5)*

    - `alloc-per-token-recognize` and `alloc-per-token-forest`, each counted over its own stage. Today's `alloc-per-token-chart` covers lexing plus recognition, and after M1 the forest too, so no R or F bound may sit on it.
    - `gc-cpu-share-chart` and `gc-cpu-share-rd`: collector CPU-seconds over total process CPU-seconds, both read from `runtime/metrics` (`/cpu/classes/gc/total` over `/cpu/classes/total`) around one side's whole pass. That is the stated denominator.
    - There is **no per-stage collector key**. Collection is concurrent and process-wide: garbage made in recognition is collected during the forest's span. So F1 reports its stage time key and `alloc-per-token-forest`, with `gc-cpu-share-chart` as context only.
    - All reported, never gated.

    **Criterion:** `make -s chart-ratchet | awk '$2=="alloc-per-token-forest"{a=1} $2=="gc-cpu-share-chart"{g=1} END{exit !(a&&g)}'`

**Every R and F issue then reports its own stage key before and after, under the ratchet's protocol**, as a median of three measurements with their min and max. The verdict reading is taken once, at the end.

**An issue that shows no gain is reported before it merges.** If an R or F issue meets its counting criterion but its stage key's median stays inside the before-reading's min–max range, that goes to perigrin before the merge. R4, R5, F1 and F2 are unprototyped, and pvm's rule is to measure before optimizing.

### R · Recognizer

1.  **R1 · Postdot index per set core** *(low risk · prototyped)*

    For each interned core, on first use, map each symbol to the items waiting on it and their target states. Scan, completion and the Leo test read it. Branch sets keep the old loop, and the cache's trie records dependencies in a new order; both go in the issue.

    While here, count how often one set completes the same (symbol, origin) twice. If it is not rare, complete each pair once (Opedal's fold, §5.2).

    YAEP `core_symb_vect_find`; Marpa PIM. Not in Aycock.

    **Criterion:** a test that counts `goTo` probes per token over the corpus sample and requires at most `B`. Today about 5.6 (1.61M over 286k). `go test ./internal/chart/recognize -run '^TestCompletionProbesBounded$' -v | grep -q '^--- PASS: TestCompletionProbesBounded'`. Guards: P2, `TestCacheMatchesNoCache`, the fuzz target.

2.  **R2 · No per-node maps or slices** *(medium risk · prototyped)*

    The design is the prototype's: find duplicates by scanning, switch to a map at 24 items; carve `in` and `out` from one array; reuse the builder's buffers; copy items only for a new set. `Chart.Slipped` clones, which is where the prototype's one aliasing bug was.

    YAEP `sit_dist_insert` and its object stacks, adapted.

    **Criterion:** `make -s chart-ratchet | awk '$2=="alloc-per-token-recognize"{f=1; ok=($3<=B)} END{exit !(f&&ok)}'`. The key is M3's, recognition only. `B` is provisional: refinement takes it from the prototype run under the ratchet. The review's own meter read 8.2 falling to 1.2; the ratchet's meter reads differently (11.40 for lexing plus recognition), so neither figure is the bound. Guards: P2, and a `-race` test on `Slipped`.

3.  **R3 · Make Leo cheap; do not change its default** *(low risk · prototyped)*

    Test eligibility in constant time from the index and store only real memos. Leo stays on by default: it is a carried-over component (§5.1) and `leo-disagree` is gated.

    **Criterion:** a test that counts whole-set scans by `memoAt` per node and requires at most `B`. Today about 1 per node (263,000 over 263,245). `go test ./internal/chart/recognize -run '^TestLeoMemoScansBounded$' -v | grep -q '^--- PASS: TestLeoMemoScansBounded'`. Guards: `leo-disagree 0`, `TestForestFromChart`.

4.  **R4 · A dense next-state table, only if a profile asks for it** *(low risk · gated · unpriced)*

    Revision 2 said `make chart-tables` could emit this as data. That was wrong: `make chart-tables` writes only the text files the `tables-drift` key compares, and nothing loads them. Production tables are built at run time by `grammar.Generate`, once per tables value.

    So R4 is the smallest thing that removes the binary search: a dense state-by-symbol table built once in `buildGrammar`, beside the other grammar-only tables. It needs neither the generator nor P1.

    **Gate:** after R1, scan and completion over interned sets no longer call `goTo`. What still does: branch sets, `memoAt`'s fallback for branch sets, the slipper (`slipper.go`) and `finished.go`. R4 is refined only if a profile taken after R3 names `goTo` above the reading's noise. Otherwise it is dropped, and that is reported.

    **Criterion, if refined:** `go test ./internal/chart/recognize -run '^TestGoToIsTableLookup$' -v | grep -q '^--- PASS: TestGoToIsTableLookup'`, a test that counts binary searches over the corpus sample and requires 0. Guard: P2.

5.  **R5 · Hash each set once in the interner** *(low risk · estimate only)*

    One saved hash in place of varint encoding and three map lookups (`sets.go` `intern`).

    YAEP `set_core.hash`, `dists_hash`.

    **Criterion:** a test counting map lookups per interned set, at most 1 (today 3). `go test ./internal/chart/recognize -run '^TestInternOneLookup$' -v | grep -q '^--- PASS: TestInternOneLookup'`. Guard: P2.

**Out of scope: R6** (cheaper transition-cache storage). It overlaps input 41df, ruled out on 2026-11-10.

**Out of scope: R7** (a directly-executable recognizer). The thesis could not compile DEEP for a 350-rule grammar (§4.3); revision 2 holds the full case.

### F · Forest

1.  **F1 · Pointer-free, index-based nodes** *(medium risk · unmeasured)*

    `forest.Node` and `Alt` become int32 indexes into slices that hold no pointers, so the collector has nothing to trace. No production caller reads `Node` or `Alt`; `cmd/ratchet/speed.go` uses only `forest.Bench`. It touches no recognizer code, so it runs beside R1. It must land before F2 and before any precedence or resolve issue, since it changes the code they are written against.

    **Criterion:** `go test ./internal/chart/forest -run '^TestForestNodesPointerFree$' -v | grep -q '^--- PASS: TestForestNodesPointerFree'` (reflect over the node and alternative types). Guard: P5's golden, unregenerated. Reports the forest stage key and `alloc-per-token-forest` before and after.

2.  **F2 · Size and reuse forest memory** *(low risk)*

    Three parts, after P1, P3, P5, R2 and F1:

    - Size the first blocks from the file.
    - Build grammar-only forest data once per tables value (P1).
    - **Reuse inserted-`;` results, for built nodes only.** `Recognize` computes the slipper's closure only when it builds a node (`builder.slip`). A node served from the transition cache skips the builder, and most nodes are served that way. So the chart keeps the closure for the nodes it built, and the forest reads it there. For a cache-served node the forest calls `Chart.Slipped` and recomputes, exactly as today. **The cache is not touched and stores nothing new**; that would be R6.
    - **The two closures differ today, and F2's issue must reconcile them.** Recognition closes the slipper's items with Leo on (`slipper.go` `semiAt`); `Chart.Slipped` closes them with Leo off, link by link. A forest that reads recognition's closure for some nodes and `Slipped`'s for others must get the same forest from both. P3 and P5 are the guards.
    - The issue reports what share of inserted-`;` nodes the reuse covered. If the share is small, this part is dropped and the other two stand.

    **Criterion:** `make -s chart-ratchet | awk '$2=="alloc-per-token-forest"{f=1; ok=($3<=B)} END{exit !(f&&ok)}'`, with `B` taken from M3's first reading after F1. Guards: P3, P5.

**Out of scope: S** (counting safe sets). Not in this run.

**Out of scope: L** (lexer shape). Not in this run; D6 is not asked.

### C · Precedence on items *(one scratch spike; perigrin's Chalk direction)*

perigrin's point: Chalk prunes by precedence while it parses, it gave up the SPPF only because it needs one parse, and pvm could keep every surviving reading.

**The forms, and where each stands.**

| Form | What it is | Standing |
|---|---|---|
| Value on items, `add` picks a winner | Chalk as written (`Precedence.pm` `add`) | Not built. It discards readings, and its rule is a heuristic, not bison's resolution. |
| **Value on items, `add` keeps every survivor** | Chalk's packed survivors (`FilterComposite.pm` `_pack_survivors`) | **C1, as scoped.** Scratch only. |
| Compiled into the tables | The same finite context multiplied into the grammar | **Ruled out by perigrin:** it grows the grammar, and Earley's cost scales with grammar size. |
| Scope or type on items | Declared subs, imports, features | Out. These depend on position, so no two nodes could share a set. |

**Why the spec does not forbid the spike.** Rule 12 and §10's "facts in items" bind merged code. C1 is never merged, so it breaches neither.

1.  **C1 · Precedence as a value on items, measured** *(scratch spike; D7)*

    **What rides on an item.** An item is (state, distance), and an LR(0) state holds many dotted rules. A single value per item could not kill one rule inside a state. So in the spike:

    - **The value is per dotted rule.** A kernel item carries one entry for each kernel dotted rule of its state. Each entry is the set of exposed operator contexts (§8's own value: the context at both edges) of the derivations that reach that dotted rule there.
    - A dotted rule whose entry is empty is dead at that item. An item whose every entry is empty is dropped.
    - A predicted item carries no entries: every prediction starts at one.
    - **Multiply** runs at scan and at completion. It combines a waiting entry with what was scanned or completed, and yields the empty set where bison's resolution forbids the combination. The table of resolutions is the committed `testdata/precedence/bison-conflicts.txt`.
    - **Add** is set union. Nothing is discarded, so every surviving reading is kept.
    - **An item whose value grows must complete again.** `builder.add` never revisits an item it has seen. With add as union, a later derivation can widen the value of an item that has already completed, and what it completed into must then be widened too. The spike re-queues such an item until values stop changing. The cross-check below would expose a spike that skips this.
    - The spike keeps the split ε-DFA. Each transition needs a map from the source state's dotted rules to the target's, built in scratch from `Tables.Items`. Whether that stays tractable is itself a result.
    - One assumption to check first: no rule that carries precedence is nullable. The ε-DFA moves the dot over nullable symbols before any multiply could run.

    **How a set is identified.** Sets are interned today as (core, distances). The spike interns them as (core, distances, values). That is the growth it exists to measure: two nodes with equal items and different values are two sets.

    **How Leo carries the value.** A memo's `top` is added in one step, skipping the links between (`leo.go` `builder.leo`). A value must not skip them.

    - Chalk met this and chose one multiply per chain level: it disabled chain extension so that Leo-on and Leo-off values are equal, and holds that with `t/bootstrap/leo-graph-equivalence.t` (`earley-parser.md`, "Chain Extension (Disabled)").
    - The spike does the same. A memo stores no value. `builder.leo` already walks the chain link by link to record cache dependencies; the spike multiplies along that walk, and the top's value is the product.
    - `Chart.LeoLink` is unchanged. It reads links from the sets and stores nothing, and the skipped links are in no set. The tree-side check below recomputes their values.
    - **Guard:** Leo on and off agree on acceptance, and on the value of every item present in both runs, chain tops included, over the referee sentences and the corpus sample. Whole sets cannot be compared: a Leo set leaves out the links between by design, which is why `leo-disagree` compares acceptance and `Expected`, not sets. This is Chalk's test, fitted to that.
    - **A cost to report:** the walk recomputes the chain's product on every use, so a long right recursion costs quadratic time. The report says how much of the measured time that is.
    - `Chart.Slipped` already closes with Leo off, so the inserted `;` needs no separate rule.

    **What it measures,** each with and without values, on the same corpus the ratchet's aggregate uses:

    - items per token, and items dropped by a zero value;
    - distinct interned sets, and `set-reuse`;
    - the transition-cache hit rate;
    - agreement with the bison referee (below);
    - and, from the unchanged forest, the share of packed alternatives whose rule is an operator rule or carries `%prec`. That is what in-parse pruning could spare the forest.

    **What "agreement" means.** The fixtures are `testdata/precedence/sentences.txt` (708 sentences kept, each with bison's verdict and bison's grouping as rule-and-span pairs) and `resolutions.txt` (716 resolutions covered). `TestPrecedenceReferee` does not exist; it belongs to the pending issue 1df0. So the spike reads the fixtures directly.

    - **Verdict, required, all 708.** The spike accepts a sentence exactly where bison's verdict is `accept`.
    - **Which readings count.** A rendered token's set can hold terminals the sentence does not name (`foo` reads as more than one). Agreement is judged only over readings whose leaves are the sentence's own terminals. The spike gets that by **narrowing**: it lexes the rendering as usual, then replaces each token's set with the single terminal the sentence names at that position. The recognizer and the forest both run on the narrowed input, so both sides see the same readings. Tokens of a preamble keep their lexed sets. A sentence whose tokens do not line up one to one with its terminals is counted and left out.
    - The spike also runs each sentence un-narrowed and reports the difference. **A disagreement that comes only from another reading is not §10's "precedence not expressible" stop.** It is scope's to collapse, later.
    - **Grouping, required, every accepted sentence.** The spike enumerates the sentence's derivations from the unchanged forest (`Forest.Derivations`, with its limit), evaluates the same value bottom-up on each tree, and keeps the trees whose value is not empty. Agreement is that exactly one tree survives and its rule-and-span pairs are bison's. A sentence whose enumeration is cut is counted, not passed.
    - **A cross-check that comes free:** the recognizer with values accepts exactly where some tree survives the tree-side evaluation. On the narrowed input, a disagreement is a defect in the spike, or the first sign that §10's stop "precedence not expressible" is real.
    - **A resolution agrees** when every sentence covering it agrees. The report gives the count out of 716.

    **Its size.** This is not two numbers. The multiply rule and the tree-side evaluator are the unbuilt core of the precedence issues (1dc5, 1df0), built here in scratch. D4 must size it.

    **What its numbers can decide for merged code.** Merged items stay (state, distance), and the compiled form is ruled out. So exactly two outcomes are open, and perigrin rules between them (D7):

    - **§8 as written.** Precedence acts on the forest after recognition. The spike's tree-side evaluator and its agreement counts then go to the re-refinement of 1dc5 and 1df0 as evidence that the constraint is expressible.
    - **A successor design** that amends the spec (rule 12, and §10's "facts in items") to allow one finite precedence value on items. The spike's numbers are its evidence.

    The plan sets no threshold. It says which way each number argues: sets and cache hits barely moved, with a large operator share in the forest, argue for a successor design; sets multiplied or cache hits fallen argue for §8 as written. Either way nothing from C1 merges in this run.

    **Deliverable:** a report holding the figures above, the guard's result, and the two agreement counts.

## 4. Projection onto the kill figure

**Units.** Recursive descent's aggregate time is 1.00. The kill reading is 3.51 with the forest and 2.20 without, so the forest is 3.51 − 2.20 = **1.31**. The 2.20 splits by the profile's CPU shares, lex 19 to recognize 45: lex = 2.20 × 19/64 = **0.65**, recognize = 2.20 × 45/64 = **1.55**. (Taking shares of 3.51 directly gives 0.67, 1.58 and 1.23; the two methods agree to within 0.08.)

**Factors, and where each comes from.**

- **R1–R3:** recognize × 0.50 (good) to × 0.70 (cautious). The review measured 0.46–0.53 on recognition alone. My runs give 0.51–0.69 with lattice building inside the span (224–270 ms against 393–437 ms).
- **R5:** a further × 0.92. An estimate: `intern` was 17% of the prototype's profile and R5 might halve it. Unprototyped. **R4 contributes nothing here**; it is unpriced and gated.
- **F2:** forest × 0.65–0.70. The over-engineering review's estimate for its cuts S2–S4 ("30–35% of build time"). Unprototyped, and its inserted-`;` part now covers built nodes only.
- **F1:** unknown. No source or measurement prices it. The row below assumes forest × 0.80 and is labelled as an assumption.
- **P, M, C1:** no projected change. C1 is never merged.

| Combination                        | Lex  | Recognize   | Forest      | Total (good) | Total (cautious) |
|------------------------------------|------|-------------|-------------|--------------|------------------|
| Today (the kill reading)           | 0.65 | 1.55        | 1.31        | 3.51         |                  |
| R1–R3                              | 0.65 | 0.78 / 1.09 | 1.31        | 2.74         | 3.05             |
| R1–R3 + R5                         | 0.65 | 0.71 / 1.00 | 1.31        | 2.67         | 2.96             |
| F2 alone                           | 0.65 | 1.55        | 0.85 / 0.92 | 3.05         | 3.12             |
| **R1–R3 + R5 + F2**                | 0.65 | 0.71 / 1.00 | 0.85 / 0.92 | **2.21**     | **2.57**         |
| The same + F1 *(assumed × 0.80)*   | 0.65 | 0.71 / 1.00 | 0.68 / 0.74 | 2.04         | 2.39             |
| Floor: lexing alone                | 0.65 | —           | —           | 0.65         |                  |

Arithmetic for the bold row: 1.55 × 0.50 × 0.92 = 0.71 and 1.55 × 0.70 × 0.92 = 1.00; 1.31 × 0.65 = 0.85 and 1.31 × 0.70 = 0.92; 0.65 + 0.71 + 0.85 = 2.21 and 0.65 + 1.00 + 0.92 = 2.57.

**A check against a measurement.** The R1–R3 row predicts a fall of 0.46–0.77. My probe, a different corpus and protocol, measured the whole pipeline falling from 3.05–3.24 to 2.28–2.58, about 0.5–0.9. The two agree in size.

**What the table does not contain.** Precedence, resolve and the readings count are unbuilt, and nothing here prices them. Under the old limit of 3 they would have 0.43–0.79 between them after the bold row. The forest alone took 1.31 when it arrived. So this plan does not show the full waypoint fitting under 3. It shows the two stages that exist coming down by about a third.

**Caveats.**

- The shares are CPU, not wall time. Collector work was a further 35% of samples and is not attributed.
- The factors are assumed independent.
- The prototype was timed on a probe corpus, not the kill reading's aggregate. Both are without perl.git: the kill aggregate was 1,391 files and held none of it. (Revision 2 said perl.git "is 38% of the aggregate". That figure is from one non-quiet run on PERLGITPARITY's unmerged branch, and describes what the aggregate would become, not the kill reading.)

## 5. The sixteen open waypoint-2 issues

"Hold" means untouched and still pending. Nothing is cancelled, split or edited by this run.

| Issue | Disposition | Why |
|---|---|---|
| VERDICT (1f72) | Hold, unedited. Re-refined after the verdict reading. | M1 is a new issue in `chart-recovery` with its own test and key names, so none of 1f72's criteria turns green early. 1f72's speed criteria are restated against M1's keys when it is re-refined. Revision 2's "split" is withdrawn. |
| KEYS (1fa1) | Hold; re-refine. | Its ruling covers waypoint 2's keys. M's keys need their own (D5). |
| Precedence (1dc5, 1df0) | Hold; re-refine after F1 and after C1's report and D7. | F1 changes the forest code they are written against; D7 decides where precedence lives. |
| Resolve and scope (1e1e, 1e4b) | Hold; re-refine after F1. | Same code. Scope stays after the chart under every option. |
| READINGS (1eb1) | Hold; re-refine after F1. | A sum of products over the acyclic forest. |
| Oracle (1eda, 1f12, 1fd7), PLUGIN (1f41), PRINTFH (4137) | Hold. | Downstream of precedence and resolve. PRINTFH is what milestone criterion 7 falsely reports done. |
| EXIT (2009) | Hold; its criterion changes with D1. | The resolution command hard-codes `lim = 3`. |
| COMPILEFMT (4189), TESTDEDUP (423a) | Free to proceed. | Independent of everything here. |
| PERLGITPARITY (1d6a) | Stays stopped. Its drafted split is filed only after the verdict reading. | Admitting perl.git would make it about 38% of the aggregate (one non-quiet run on its branch). Widening the population mid-recovery moves the denominator. |

**Who re-refines the held issues:** the step after the verdict reading (section 6, step 7), under whatever D1 and D7 then say. It is outside this run's budget.

## 6. Order, exit and stops

### Order

1.  **Decisions** D1, D2, D4, D5, D7 and D11 to perigrin, before chain-review.
2.  **P1, P2, P3, then M1–M3.** P4 before M1 if D11 is ruled; P5 any time before F1.
3.  **R1 → R2 → R3**, with **F1 alongside** (after P5).
4.  **R5**; **F2** once P1, P3, R2 and F1 are in; **R4** only if the post-R3 profile names `goTo`.
5.  **C1**, in a scratch worktree, any time after P2. It shares no branch with the merged work.
6.  **One verdict reading** under M, compared with section 4's projection. The exit issue records it.
7.  **Afterwards, outside this run:** perigrin rules D7 on C1's report and revises D1's ceilings; the held issues are re-refined.

**Hard dependencies:** P2 before R1. P5 before F1. F1 before F2. P1, P3 and R2 before F2. M1–M3 before every R and F issue. F1 before any precedence or resolve issue.

### Exit: what `chart-recovery`'s resolution checks

The milestone closes on its last issue, which takes the verdict reading. Its resolution is one command, in the form `chart-perly-w2`'s takes, and it checks:

- **Every stage key prints a number:** the three `speed-stage-*-aggregate-longlived` keys, `alloc-per-token-recognize`, `alloc-per-token-forest`, `gc-cpu-share-chart`.
- **The gated count keys hold:** `tables-drift 0`, `leo-disagree 0`, `rd-accepted-chart-refused 0`, `chart-accepted-both-refused 0`, `tdt-mismatch 0`, `ledger-gaps 0`, `tdt-unseen-unlisted 0`, `tdt-unaligned-unlisted 0`, and `lex-unmapped` at its known list.
- **`lex-overhead` is at or under 1.3.**
- **Both goldens pass unregenerated:** `TestSetDigestGolden` and `TestForestDigestGolden`.
- **`make test` passes.**

It does not check a speed limit, because timing keys are never compared strictly (§9.3). The verdict reading itself (three quiet measurements, each stage's median, min and max) is recorded in the milestone's Records and in the chart README's status by the exit issue. C1's report is attached to Records.

**"Done" is:** the resolution green, the reading recorded beside section 4's projection, and C1's report delivered. It is not "the chart fits under a limit". That question goes back to perigrin with the reading.

### Stops, each reported to perigrin

- **A stage key above the ceiling D1 sets,** read as the median of three measurements at an issue's close. The ceiling is the highest of M1's three measurements, so unchanged code does not trip it.
- **The verdict reading above the cautious projection, 2.57,** for lex, recognize and forest together, with the same margin: it fires only if the lowest of the reading's three measurements is above 2.57. A stop to report, not a kill.
- **P2's golden, or P5's, needing regeneration.**
- **`lex-overhead` above 1.3.**
- **Facts in items in merged code:** any merged item that is not (state, distance).
- **An issue whose stage key does not move** (M, above): reported before it merges.
- **The usual list:** twice the budget D4 sets (§14); a definition amended twice by the orchestrator (§10); a change needed outside `internal/chart` (§11); a criterion that cannot run or passes without its work; a gated count key that regresses.

## 7. The round-2 and round-3 findings

Assessment A is the round-1 assessor, resumed. Assessment B is the fresh participant. A released all ten of round 1's held findings; they are not repeated here.

| Finding | What revision 3 does | Status |
|---|---|---|
| **A-B1, B-B1** C1 is not the spike perigrin scoped | C1 rewritten as the scoped spike (section 3, C): a value per dotted rule on each item; how sets are identified; how Leo carries the value, after Chalk; what agreement means; its size; what it can decide. The compiled form is marked ruled out. D7 rewritten. | Resolved |
| **A-B2, B-B3** no exit; budgets only after the verdict | Section 6 gives the resolution, what "done" is, and the stops. D1 now recommends provisional per-stage ceilings from the start. | Resolved, subject to D1 |
| **B-B2** F1's guard is not in the tree | P5 commits a forest-identity golden before F1. | Resolved |
| **B-B6** no rule for an issue that shows no gain | Stated under M and in the stops: reported to perigrin before it merges. | Resolved |
| **B-B4** R4's premise is wrong, and R4 is unpriced | R4 is now a dense table built once in `buildGrammar`, gated on a post-R3 profile. Section 4 credits it nothing. | Resolved |
| **B-B5** F2's reuse is unavailable on cache-served nodes | F2 reuses for built nodes only, recomputes for cache-served nodes as today, leaves the cache alone, and reports the share covered. | Resolved |
| **A-B3** criteria read the wrong field | Every key criterion tests field 2 for the key and field 3 for the value. | Resolved |
| **A-B4, B-B7** criteria without commands; a guard used as red | M2, R3, R4 and R5 have commands. `tables-drift 0` is no longer a criterion. | Resolved |
| **A-B5, B-B7** R2's bound is on a key that includes lexing | Bound moved to M3's `alloc-per-token-recognize`, as a provisional `B` taken from the ratchet. | Resolved |
| **A-B6** `TestSpeedChartSideIsPipeline` is taken | M1 uses `TestSpeedStagesTimed`. | Resolved |
| **B-B7** P4 lints the live milestone; P4 has no place in the order | Fixture-only criterion. P4 before M1 if D11 is ruled, else the pin is converted. | Resolved |
| **A-B7** the load gate is a ruled definition | Q12 dropped from M1 and listed under D12. | Resolved |
| **A-B7, B-B8** M3's collector figure has no denominator and is not per-stage | M3 states the denominator, reports per side not per stage, and F1 leans on the stage time key. | Resolved |
| **A-B7** section 9 says the plan is not in the repository | Corrected; it is committed under `docs/plans/`. | Resolved |
| **A-B7** R4 placed "after P1" | R4 no longer needs P1. | Resolved |
| **B-B8** F1 before F2 unstated | Added to the hard dependencies. | Resolved |
| **B-B8** VERDICT's split leaves 1f72 stale | The split is withdrawn; 1f72 holds unedited and M1 takes new names. | Resolved |
| **B-B8** D2's timing | D2 goes to perigrin before chain-review; refinement carries this file. | Resolved |
| **B-B9** the perl.git caveat | Corrected in section 4 and reworded in section 5. | Resolved |

**Round 3.** Assessment A's findings are G1–G6; assessment B's are N1–N4. Both seats released every round-2 finding.

| Finding | What revision 4 does | Status |
|---|---|---|
| **G1** C1's Leo guard cannot pass | The guard compares acceptance and the value of every item present in both runs, chain tops included. Whole sets are not compared. | Resolved |
| **G2** agreement ignores tokens with several terminals | Agreement is judged on input narrowed to the sentence's own terminals. Un-narrowed runs are reported, and a disagreement from another reading is not §10's stop. | Resolved |
| **G3, N1** a ceiling at one reading trips on noise | Each ceiling is the maximum of M1's three measurements. The 2.57 report fires only if the lowest of three is above it. | Resolved |
| **G4, N2** P4's criterion passes with the script missing | Replaced by `TestCriteriaLint`, which needs each defect named on the bad fixture and exit 0 on the clean one. | Resolved |
| **G5** exit checks omit three gated keys | `ledger-gaps`, `tdt-unseen-unlisted` and `tdt-unaligned-unlisted`, each 0, added. | Resolved |
| **G6** the Leo walk is quadratic | C1's report states the cost. | Resolved |
| **N4** `builder.add` never revisits an item | C1's brief requires an item whose value grows to complete again. | Resolved |
| **N3** the slipper's two closures differ in Leo | F2's issue must reconcile them, with P3 and P5 as guards. | Passed to F2's issue |
| **B4 note** `memoAt`'s branch-set fallback calls `goTo` | Added to R4's list of callers. | Resolved |
| **P5's provenance** | The scope section and D4 say P5 is the plan's addition, for perigrin to rule on. | Resolved |
| **P5's dump** | It includes a rule position's rule and dot. | Resolved |

**Not resolved by the plan, because they are perigrin's:** whether P5 is in, and the rulings D1, D2, D4, D5, D7 and D11 themselves.

## 8. Decisions for perigrin

D1, D2, D4, D5, D7 and D11 are asked before chain-review. The rest are recorded so the numbering holds.

1.  **D1 · What replaces the 3× limit**

    - **(a)** Keep 3× on the whole pipeline. The projection reaches it for lex, recognize and forest; nobody knows about the rest.
    - **(b)** A ceiling per stage in RD units, each stage's issues held to it (the postmortem's first change).
    - **(c)** An absolute latency: §13's "an edit to a re-parsed tree under 10 ms at p99". It belongs to interleaved mode, which is not built.
    - **(d)** No limit; report only.

    **Recommendation: (b), with provisional ceilings now and a revision after the verdict.**

    - **Now:** no stage may get worse than it is. Each ceiling is the **maximum of the three measurements** in that stage key's first quiet reading under M1, not their median. The kill reading's own nine measurements ran from 3.42 to 3.70, so a ceiling at the median would stop about half of all closes on noise. Section 4 expects the medians near **lex 0.65, recognize 1.55 and forest 1.31**; M1's issue writes the three ceilings down. `lex-overhead` keeps its 1.3.
    - **After the verdict reading:** perigrin revises the three, and names a whole-pipeline figure with the unbuilt stages' shares stated.
    - A breach is a stop reported to him, not a target and not a kill. The literature supports a looser whole figure than 3 for a parser that keeps every reading (thesis §4.3: 2–5× for recognition alone; YAEP: 2.5–6× with one tree).

2.  **D2 · The accepted spec's status**

    §2 and §10 say a speed kill closes up-front mode. Options: amend the accepted spec; accept this plan as a successor decision that supersedes §2's question, §10's speed line and §9.5's key rule for the recovery; or record a ruling in Records.

    **Recommendation: a successor decision, ruled before chain-review.** The accepted spec is frozen. This file is what refinement carries onto `chart-recovery`.

3.  **D3 · Where the work lives**

    Settled by the scope: a new milestone, `chart-recovery`. Waypoint 2 keeps its verdict and its sixteen issues as section 5 disposes of them.

4.  **D4 · Budget and due date**

    §14 makes twice a waypoint's budget a stop, and waypoint 2 is due 2026-11-30. **Recommendation:** `chart-recovery` gets its own budget and date, sized at refinement, covering P1–P4, M1–M3, R1–R5, F1–F2 and C1, plus P5 if he rules it in. **P5 is the plan's addition beyond the scoped P1–P4;** without it F1 has no corpus guard that survives its own rewrite. C1 is the item to size with care: its multiply rule and tree-side evaluator are the unbuilt core of two pending issues. Re-refining the held issues is not in it.

5.  **D5 · New and redefined ratchet keys**

    §9.5 says keys are never added after the reset. M1 and M3 add keys; M2 renames one. The precedent is the 2026-11-10 ruling "Yes, in one issue". **Recommendation: the same ruling again**, covering the three `speed-stage-*` keys, `alloc-per-token-recognize`, `alloc-per-token-forest`, the two `gc-cpu-share-*` keys and the rename of the cold key. R's counters are tests, not keys, to keep the ruling small.

6.  **D6 · Lexer direction.** Not asked. L is out of scope.

7.  **D7 · What C1's report may lead to**

    C1 itself is scoped: a value on items, add keeping every survivor, scratch only. What perigrin is asked before chain-review is to confirm the three choices the spike rests on:

    - the value is **per dotted rule** within an LR(0)-state item;
    - Leo multiplies **link by link**, with Leo-off equality as the guard;
    - agreement means **the verdict on all 708 and the grouping on every accepted sentence**.

    What he rules after the report is which of two outcomes follows for merged code:

    - **(a)** §8 as written: precedence over the forest. C1's evaluator and counts feed 1dc5 and 1df0.
    - **(b)** A successor design that amends rule 12 and §10 to allow one finite precedence value on items.

    The compiled form is ruled out, and scope or type on items stays out. **Recommendation: none between (a) and (b) until the numbers exist.** Section 3 says which way each number argues. Revision 2 recommended against values on items without having measured them; that recommendation is withdrawn.

8.  **D8 · Keep the split ε-DFA?** Not asked. Merged code keeps it. C1 keeps it too, and reports whether per-rule values stay tractable on LR(0) states.

9.  **D9 · Leo's default.** Not asked. It stays on; R3 removes its cost.

10. **D10 · R6 and R7.** Not asked. Both are out of scope.

11. **D11 · Criterion repairs**

    Six: the four unrunnable criteria (4202, 1b0f, 41c7, 01a1bd9f), milestone criterion 7, and the w1 `speed.go` pin FOREST broke. Each earlier repair of this kind was a ruling. **Recommendation:** one ruling, "convert or move, as before", applied in P4, with the pin converted to a check that does not count commits to `speed.go`.

12. **D12 · Held, not asked now**

    Still perigrin's, and unchanged by this run: the four PERLGITPARITY split decisions (aligner amendment 4; NOAMP as a fitted terminal; widening `overrides.txt`; admitting 01a1c678 into FEATWORDS); ponytail L5 (the 25-copy lex pass); ponytail L1 (test-only shared-table machinery); ponytail S7 and L4; and review Q12, the load gate's comparison with `NumCPU`, which touches the 2026-11-11 ruling. **Recommendation:** leave all of them until after the verdict reading.

## 9. Risks and unknowns

- **The projection is arithmetic on shares.** Only R1–R3 are prototyped. F1, the largest hoped-for forest gain, has no number at all.
- **The unbuilt stages.** Precedence, resolve and readings have never been timed. The run's end state is "two stages cheaper", not "the waypoint fits".
- **Collection.** Collector work is heavy in every stage I measured, RD included. F1 addresses the forest's part; nothing here addresses the recognizer's beyond R2. M3's side-wide key cannot attribute it to a stage.
- **R2's buffer reuse** is where aliasing bugs hide. P2 and a race test are its net.
- **R4, R5, F1 and F2 may show nothing.** The no-gain rule exists for them.
- **C1 is larger than it looks.** It builds the precedence constraint's core in scratch. If its tree-side evaluation cannot agree with bison, that is §10's "precedence not expressible" stop arriving early, which is worth knowing.
- **C1's per-rule value on LR(0) states** may not stay tractable. The spike reports that as a result, not a failure.
- **The provisional ceilings rest on M1's first reading.** If that reading is taken on a loaded machine, every later stop is measured against a soft number. M1's issue takes it quiet, three measurements, and the ceiling is their maximum.
- **This file is a plan, not yet a decision.** It is committed under `docs/plans/`. Its standing against the accepted spec is D2.

### Not verified

- Aycock and Horspool's 2002 journal paper and Leo 1991: not available to me. The thesis covers the same construction but reports no timings for it.
- Why Marpa dropped the automaton: its source says that it did, not why.
- DEEP's ratios to Bison are read off scanned plots.
- Chalk: I read the four documents in full, `Earley.pm`'s `_complete`, and `Precedence.pm`'s and `FilterComposite.pm`'s `multiply`/`add` paths. I did not read every line of the other semirings, and I did not open `t/bootstrap/leo-graph-equivalence.t`; I cite it from `Earley.pm`'s own comment.
- My timings: a scratch probe at load 0.7–1.8, two runs per tree, not the ratchet's protocol. `make test` and `make chart-ratchet` were not run.
- For revision 3 I checked three things in the tree: the ratchet's line format (`baseline.txt`), the precedence fixtures' columns (`sentences.txt`, `resolutions.txt`), and the Leo code (`leo.go`, `finished.go`). The other code claims the round-2 findings rest on (`buildGrammar`, `builder.slip`, `speed.go`'s allocation meter, the 2026-11-11 ruling, 1f72's criteria) are taken from the two assessments. So are round 3's: `semiAt` closing with Leo on, `memoAt`'s branch-set fallback, and the three extra gated keys. That `builder.add` never revisits a seen item I read myself in `recognize.go`.
- That `bison-conflicts.txt` is enough to state C1's multiply rule is an assumption. It is the first thing the spike finds out.
- Whether any precedence-bearing rule is nullable: unchecked; C1 checks it first.
- The repeat rate of (symbol, origin) completions and the operator share of forest ambiguity are unmeasured; R1 and C1 each count one.
- 4.46 items per node, 69% of nodes cache-served and 21,026 sets for 263,245 nodes are the review probe's figures. The ratchet's nearest keys read 4.874 and 0.712 on a different meter.

Sources: Aycock, PhD thesis (2001), ch. 3–7 and Figures 4.9, 4.11, 4.12 from the PDF; Opedal et al. (ACL 2023), §1–§8 and App. A, C, G, H, K, L; YAEP `src/yaep.c`, `doc/Internals.txt`, README; libmarpa `marpa.w`, `legacy.w`; Kegler's `ah2002_notes.ltx`; Chalk at e26caaff; pvm at origin/pu 352b3671 (code identical to 3487d258 under `internal/` except two READMEs); the v2 design spec; the `chart-perly-w2` milestone Records; the waypoint-2 postmortem; the round-1 and round-2 assessments; the earlier review and its prototype.
