<!-- ABOUTME: Design for a research spike: an Earley chart parser for Perl, built beside the recursive-descent parser. -->
<!-- ABOUTME: States the question, the architecture, the phases with their gates, and the kill criteria; the verdict is perigrin's. -->

# A chart parser for Perl: design for a research spike

**Status:** design for a spike, written 2026-10-29. Not a decision. If the
spike is accepted, this becomes a chapter of `docs/specs/perl-parser/`; until
then nothing here describes the parser pvm ships.

## 1. Why

`docs/plans/2026-09-05-parser-prior-art.md` §A1.3 chose hand-written recursive
descent because it was the only measured architecture in which *"the lexer
can ask the parser what it is looking at."* That choice has held: T2 56/56,
perl.git `t/` 619/619, M3-directory re-emission 347/347, WRONG = 0.

What strains is context. How Perl parses depends on what is known at that
point: declared subs, prototypes, features, imports. In the current parser
that knowledge is implicit parser state. Two kinds of evidence show the cost:

- About 49 of the open v0.1 issues are cases where the parser disagrees with
  perl, and most of them are context-dependent (a bareword the lexer cannot
  know is declared, `say` read as the builtin whatever the feature state, a
  bare `v5`, an imported builtin override). The count is from a title-keyword
  sort and is approximate.
- The M2 incremental work needed hand-written invalidation (`sameState`,
  `keepsLead`, `peekEnd`, `openSquares`), and nearly every gate review found
  another piece of hidden parser state.

An Earley chart parser can make the dependency explicit instead of implicit:
the parser reports every reading and the condition that selects it, and
something outside the parser decides.

## 2. The question

> Can an Earley chart parser reach parity with the floor pvm stands on, at a
> speed worth pursuing, and does it have a better chance than the current
> parser at M2 (incremental, correct under edit) and M4–M6 (prototypes and
> ambiguity, real-world code, conformance)?

The answer rests on four kinds of evidence, each measured in a phase of §8:

1. **Defect yield.** For each open "perl disagrees" issue: can the chart fix
   it, where does the fix land (lexer tiebreak, resolver rule or grammar),
   and how big is the fix?
2. **Head-to-head on milestone corpora.** M4: `t/comp/proto.t`,
   `t/op/signatures.t`, T1. M5: the T4 regex files and CPAN. M6: `t/op`.
3. **Cost of change.** For each contextual piece rebuilt in phase 2, and for
   each M4–M6 probe: lines, places touched, where it landed, review cycles.
4. **Incremental potential,** measured last: how soon Earley sets match again
   after an edit, and M2's own equivalence contract.

### Out of scope

- Replacing the production parser.
- Choosing the downstream IR (pvm's lowered tree or B::SoN).
- Feeding `lower`, `infer` or `psc`.
- A B::SoN semantic oracle (§11, future work).

### Output

A written recommendation with every number. It gives two verdicts, which may
differ: one on the M4–M6 route and one on the incremental route. The code
built along the way is kept, but stays experimental until the spike is
accepted.

## 3. Design rules

These were settled while designing and every component follows them.

1. **The grammar is static; the context is dynamic.** Imports and pragmas do
   not change the grammar. They change how tokens are classified and which
   reading applies.
2. **Context enters at two places only:** before the chart (token
   classification, in `chart/lex`) and after it (pruning, in
   `chart/resolve`). Never inside the chart. Facts inside Earley items would
   make otherwise-identical sets differ, which kills set interning and with
   it both the speed and the incremental route (rebel's YAEP analysis, §12).
3. **Only compile-time facts may prune.** A reading may be removed only by a
   fact perl itself could have known at compile time: a prototype, a declared
   sub, a feature, an import, a library declaration the oracle has vetted.
   An inferred runtime type never prunes, because perl never consults one
   when parsing. Where perl's reading fails to type-check and another reading
   would, that becomes a diagnostic.
4. **Ambiguity may stay.** A language server needs the facts true in every
   reading, not one parse. A WRONG parse is one that *asserts* a reading perl
   does not take; leaving perl's reading among several is *covered but
   undecided*. A cap collapses runaway regions to an Error node.
5. **Outside the accepted language, refuse; never misparse.** Pest scored
   99.996% and still lost (prior-art §A1.3) because it misparsed silently.
   Error recovery produces an Error node over the smallest skipped region.
6. **Weights may order, never decide.** Corpus frequencies may order
   suggestions shown to a user. They never select the tree.

## 4. What is borrowed and what diverges

Chalk (`perigrin/chalk`) already implements an Earley parser with most of the
techniques below, in Perl. pvm's chart parser is an independent design in Go
that borrows Chalk's proven answers where the constraints match.

| Borrowed | Diverges, because pvm's constraints differ |
|---|---|
| Integer core-item IDs, nullable sets, Aycock–Horspool dot advancement, terminal clustering per automaton state (`LR0DFA.pm`, `CoreItemIndex.pm`) | Terminals are `internal/lexer` token kinds, not inline regexes. Chalk is scanless; pvm keeps its lexer |
| The relative-distance chart, `chart[pos][core][dist]` | Whole sets are also interned, YAEP-style, which Chalk does not do |
| Leo, and Chalk's Leo-equivalence test as a practice (`t/bootstrap/leo-graph-equivalence.t`) | Leo memos are stored by reference ("the top of the chain I continue, *d* tokens back"), never as the absolute top origin, so they cannot break interning. Chalk's Leo items store `top_origin` |
| The BNF notation, rules only. The semantic actions live separately in `Actions.pm` and are not used | No in-parse semirings. Chalk resolves ambiguity while parsing because it builds IR directly; pvm builds a forest and resolves afterwards |
| The seven ambiguity classes (`docs/architecture/ambiguity-classes.md`) as test categories and bracket labels | — |
| "The grammar over-generates; later layers narrow" | The narrowing happens after the chart, under rule 3 |

From YAEP (`vnmakarov/yaep` at 7aba93e, per rebel's reading of the source):
interned situations, cores deduplicated by their kernel, distance vectors
deduplicated, whole sets interned as (core, distances), the per-(core, symbol)
transition and reduce lists, and the (set, token, lookahead) transition cache.
YAEP's DAG is not borrowed: it is per rule instance, carries costs rather than
conditions, and discards the concrete tree.

## 5. Components

A new tree, `internal/chart/`, that depends only on `internal/lexer`. The
recursive-descent parser is untouched.

### `chart/grammar`

A `.bnf` file in Chalk's notation, rules only, with terminals named as lexer
token kinds, plus a generator run by `go generate`. The tables are built ahead
of time because the grammar never changes at runtime (rule 1).

- **Seed:** `grammar/perl.bnf` at Chalk commit 00399607 (2025-10-20), the last
  full Guacamole-derived grammar before Chalk simplified it. It has 374 rules
  and 227 nonterminals, Guacamole's naming (`ExprAssign` … `ExprValue`,
  `OpListKeywordExpr`), and a right-recursive statement list, the shape set
  interning wants.
- **Translation:** from its `A -> B C` syntax to the current `::=` notation;
  drop `WS_OPT`, because the lexer carries trivia; map its `%PATTERN_n%`
  regexes to token kinds; add Guacamole v0.008's four-way precedence clones
  (e.g. `ExprAddU`/`ExprAdd0`/`ExprAddL`/`ExprAddR`), which stop list
  operators producing double parses and which the seed lacks.
- **Generated tables:** core items; the LR(0) automaton with Aycock–Horspool
  nullable closure; per-state transition and reduce lists; per-state
  expected-terminal sets; per-state Leo eligibility; and **expectation
  classes** (below).
- **Expectation classes.** The expected-terminal sets come in far fewer shapes
  than there are states. The generator compresses them into a small set of
  classes, which is `PL_expect` derived from the grammar rather than written
  by hand.
- **Assumption to confirm first:** the notation parses cleanly with its
  semiring parts ignored. perigrin believes it is separable.

### `chart/lex`

Reuses `internal/lexer`'s scanners (quote-likes, heredocs, numbers and so on)
but not its hand-written `PL_expect` state machine. At each position it reads
the expectation class and scans only the token kinds that class allows.

- **Earliest failure.** If only `=` can continue and the next character is
  `*`, that reading dies at that position. Earley's valid-prefix property
  makes this exact.
- **Perl's tiebreaks stay.** Where several expected kinds match and perl
  decides heuristically (`{` through `intuit_more` and its weighting, spec
  chapter 3), the lexer keeps just those heuristics. That is pruning by a
  rule perl applies at compile time, so rule 3 allows it.
- **Token branches.** Where even perl cannot decide statically, the lexer
  emits both readings as a two-way branch, and the forest carries the
  condition.
- **Up front or interleaved.** Perl's own `PL_expect` runs standalone from the
  previous token plus a bracket stack. If the generated classes can be driven
  the same way, lexing stays up front, keeping today's token cache. If not,
  lexing interleaves with the recognizer. Phase 1 decides by measurement
  (§8).

### `chart/recognize`

A YAEP-style Earley recognizer over the generated tables: sets interned as
(state, distances), relative distances, the transition cache, and Leo memos
by reference. It exposes `Expected(pos)` for `chart/lex`.

### `chart/forest`

A Scott SPPF built from the chart, over the lexer's tokens, so round-trip is
lossless. Packed nodes may carry a condition; in phase 1 none do. Leo's
shortcut must still yield the intermediate nodes the forest needs: Chalk
disabled Leo chain extension because baking intermediate completions into a
value made Leo trees differ from non-Leo ones.

### `chart/resolve` (phase 2)

Prunes packed nodes using compile-time facts (rule 3): prototypes, declared
subs, features, imports, vetted type-library declarations. Whatever it cannot
decide stays ambiguous.

### `chart/compare` (harness)

Exports each parser's tree as labelled spans (first token, last token,
label), PARSEVAL-style, and diffs the two parsers per file. Labels cover
Chalk's seven ambiguity classes, so a same-span disagreement such as block
versus hash still shows. It feeds the existing oracle and ratchet harness.

## 6. Data flow

```
source bytes
   │
   ▼
chart/lex ── expectation class per position ──────────────────┐
   │  scans only the expected kinds; perl's tiebreaks;        │
   │  a two-way token branch where perl cannot decide          │
   ▼                                                           │
token stream (lossless: trivia attached)                       │
   │                                                           │
   ▼                                                           │
chart/recognize ── interned sets (state, distances) ───────────┘
   │  a reading dies at its first unexpected token
   │  if every reading dies: error recovery (`error` symbol, minimal skip)
   ▼
chart/forest ── Scott SPPF; packed nodes may carry conditions
   │
   ▼
chart/resolve ── compile-time facts only; the rest stays ambiguous
   │
   ├──▶ chart/compare: labelled spans → bracket diff against recursive descent
   │                    → oracle and ratchet harness
   └──▶ future: LSP features ask "do all readings agree?"
        future: a resolver-to-SoN builder → B::SoN semantic oracle
```

How each outcome is graded:

- **Input outside the accepted language:** an Error node; the oracle counts it
  as a refusal, not WRONG.
- **Both readings of a token branch die:** an ordinary syntax error, located
  at the earliest token where no reading continues.
- **Ambiguity left after resolution:** not an error. It counts towards
  coverage if perl's reading is among the live ones, and not towards
  decisiveness.

## 7. The parity contract

Parity with the recursive-descent parser means all three of:

1. **The oracle:** WRONG = 0, with coverage and exactness at or above the
   current ratchet levels.
2. **Lossless round-trip** from the forest's tokens.
3. **A per-file bracket diff** against recursive descent, wherever both
   produce one reading: identical labelled spans. A difference is a finding
   for one parser or the other, and the oracle decides which.

Canon re-emission stays the recursive-descent parser's. The chart parser
inherits its guarantee on every file where their brackets agree. A canon for
the chart is needed only if recursive descent is retired, and is decided
then. Chalk's re-emission is a lowering from the SoN IR, not from a CST, so it
does not serve here.

## 8. Phases and gates

The phase order is perigrin's: Standard Perl first, then parity, then
performance, then incremental.

### Phase 1: Standard Perl

- **Build:**
  - the grammar translation and generator (tables and expectation classes);
  - `chart/lex`, `chart/recognize`, `chart/forest`;
  - the minimal token-branch case: an undeclared word followed by `/`, with
    both readings branched and chosen by `Expected`. Building it now proves
    the interface, the lookahead handling and conditional packed nodes early;
  - `chart/compare`.
- **Corpus:** Guacamole v0.008's own test suite, taken as valid: about 546
  `parses(...)` cases (must parse) and 87 `parsent(...)` cases (must refuse),
  in 38 files. Guacamole is not run.
- **Gate:**
  - every `parses` case accepted, WRONG = 0 against perl, round-trip lossless,
    and the bracket diff agreeing with recursive descent, or every
    disagreement decided by the oracle;
  - every `parsent` case refused with an Error node, never a plausible tree;
  - the expectation classes compared against spec chapter 3's eleven
    `PL_expect` states. The result is recorded, and decides up-front versus
    interleaved lexing. Agreement is evidence the grammar models Perl
    correctly; each difference is a grammar bug or a place perl's lexer knows
    something the grammar does not.

### Phase 2: parity

- **Ratchet order:** Guacamole's exclusion list, one feature at a time:
  auto-quoting; `print {$fh}` forms and bareword filehandles; unusual quote
  delimiters; heredocs; paren-less calls; prototype-driven parses; indirect
  objects; `given`/`when`. The order may change by yield.
- **Each step records:** where the fix landed, its size, the places it
  touched, the open defects it settled, and the `parsent` cases it flipped
  into correct, oracle-verified parses.
- **M4–M6 probes.** A sample of two or three per milestone, fixed before
  phase 2 starts and drawn from open issues, is implemented and costed on
  both sides. Candidates:
  - **M4:** `first { ... } @list`; a named unary versus a list operator with
    an unknown prototype; `sort $coderef @list`;
  - **M5:** `s///e`; `(?{ ... })`; unusual or nested delimiters;
  - **M6:** failing constructs from `t/base/lex.t` and `t/comp/parser.t`.

  On the recursive-descent side, the cost is the fixing PR's diff and review
  rounds where the issue is already closed, or the same measures from
  implementing it where it is open. Only probes whose issues are open in M4–M6
  anyway are implemented on the recursive-descent side, so that work is not
  wasted. Which side goes first alternates, against learning bias, and some
  probes sit late in the phase, to show whether chart costs fall as it
  matures. **"Noticeably easier"** means at least half the probes cost
  clearly less on the chart and none needs a grammar hack.
- **Gate:** parity (§7) with the current floor: T2 56/56, perl.git `t/`
  619/619, T1, and the corpus and site ratchets no lower.

### Phase 3: performance

- Leo on and off, with ahead-of-time tables throughout.
- Cold-parse time and memory against recursive descent, on edit5k.pl and the
  corpora.
- Set-reuse rate (YAEP reports about 70% on a cold parse) and the automaton's
  state count.
- A deep chained-ternary file, which is where Leo's worst-case value shows.
- **Gate:** perigrin rules on the numbers, which also sets a cold-open budget.
  A cold parse cannot meet 10 ms: YAEP-class C extrapolates to about 11 ms for
  5 kLOC, and a Go port will probably be slower. So M2's 10 ms target can only
  be met on the incremental path.

### Phase 4: incremental

- How many tokens past an edit until the new sets match the old ones again,
  what fraction match after that, and the cache hit rate (rebel's measures
  a to c).
- Then M2's contract: a repair returns a tree equal to a fresh parse or
  refuses, under FuzzIncremental and FuzzIncrementalSequence retargeted at
  the chart.
- Context edits (prototype, `use`, feature) correctly invalidate everything
  after them and get no incremental benefit.
- **Gate:** the evidence goes into the recommendation.

### Phase 5: recommendation

The M4, M5 and M6 head-to-heads, the defect-yield and cost-of-change
summaries, and the two verdicts.

### Throughout

A tripwire records cold-parse time on edit5k.pl at every step. It is an early
warning, not a gate.

## 9. Risks and kill criteria

Fixed before any work starts. A kill ends the spike early and the
recommendation says why.

| Risk | Early signal | Kill or decline criterion |
|---|---|---|
| Silent misparse | A `parsent` case accepted, or WRONG > 0 | Any WRONG that cannot be fixed by refusing |
| Forest blow-up | Ambiguity nodes per file, and the largest region | Real files needing more than 2^10 readings in one region after resolution, where perl's reading is not decidable by compile-time facts |
| Speed | The tripwire | At phase 3, a cold parse more than about 3× recursive descent's with no credible route to closing it; perigrin rules |
| Expectation classes differ from perl's | The phase-1 comparison | Not a kill. It decides the lexing mode. A large mismatch is investigated before phase 2 |
| Context leaking into the chart | Set reuse falling as phase 2 adds pieces | Any fix needing facts inside items breaks rule 2: a design review, not a patch |
| No incremental convergence (Wiren 1989) | Measure (a) in phase 4 | Not a kill. The incremental verdict becomes "keep the snapshot design for M2" |
| Probes show no advantage | The per-probe cost table | Fewer than half clearly cheaper, or any needing a grammar hack: evidence to decline |
| Drift and stall | Each phase gate's date | A gate open far longer than estimated triggers a decision on whether to continue |

## 10. Testing

pvm's conventions apply: TDD, `make test` at 100%, the pre-commit hooks.

- **`chart/grammar` generator:** golden tests on small grammars covering
  nullables, left and right recursion; generating twice gives byte-identical
  tables; expectation classes match hand computation on a toy grammar.
- **`chart/recognize`:** accept and reject sets of every toy grammar; Leo on
  and off give identical forests on everything (Chalk's practice); two
  positions with equal content share one interned set.
- **`chart/forest`:** span consistency of every packed node; lossless token
  round-trip; a known-ambiguous toy input yields exactly its readings.
- **`chart/lex`:** each reused scanner yields today's tokens on the existing
  lexer corpus under each class; the token-branch case carries both readings
  and a condition.
- **`chart/compare`:** bracket export against hand-built trees, with one check
  that must pass for each that must fail (CONTRIBUTING): a known disagreement
  is reported, and identical trees report none.
- **Ratchets,** in the existing harness: Guacamole `parses` accepted and
  `parsent` refused, bracket agreement, `parsent` cases flipped, probe costs,
  the tripwire. They are reported, but are not `make test` gates until the
  spike is accepted, so the experiment cannot block mainline work.
- **Fuzzing:** a recognizer and forest invariant fuzzer in phase 1; the
  incremental fuzzers in phase 4.
- **Isolation:** everything lives in `internal/chart/` and new test files.
  Nothing in `internal/parse`, `lower`, `infer` or `psc` changes.

## 11. Future work, not in this spike

- **A B::SoN semantic oracle.** B::SoN (`perigrin/perl5-son`) turns perl's
  compiled op tree into a Sea-of-Nodes graph, serialized as deterministic
  JSON (`docs/plans/2026-04-07-son-json-serialization-design.md` there).
  Chalk produces the same IR from source. A resolver-to-SoN builder would let
  pvm diff a resolved reading's *semantics* against
  `perl -MO=SoN,json`, a stronger check than markers. Two limits: the v1
  wire schema carries no source positions, so the IR cannot replace the
  forest for position-based LSP features; and a graph is one resolved
  meaning, so it can only be built after resolution.
- **The downstream IR.** Whether the chart's output eventually feeds pvm's
  lowered tree or the SoN IR is open. Until it is settled, the chart parser
  cannot feed analysis, so M2's tree-plus-analysis budget is measured for the
  parse half only.

## 12. References

- Aycock and Horspool, "Practical Earley Parsing" (2002); Aycock's PhD thesis
  (2001), also in Chalk as `docs/Aycock_JohnDaniel_PhD_2001.pdf`.
- Leo (1991), right-recursion optimization.
- Scott (2008), "SPPF-style parsing from Earley recognisers".
- Wiren (1989), incremental chart parsing (EACL); the non-locality of
  top-down prediction under edit.
- Wagner and Graham (PLDI 1997): keeping C/C++ ambiguity in a parse DAG cost
  under 0.5% space and 1% time.
- YAEP, `github.com/vnmakarov/yaep`. Its README's own benchmarks: 67K lines of
  C in 0.18 s parse-only against YACC's 0.07 s, at 26 MB; Marpa 3.48 s at
  516 MB. These are the author's numbers.
- Guacamole v0.008, `github.com/xsawyerx/guacamole` at 2838253: the definition
  of Standard Perl (`lib/standard.pm`) and its test suite.
- Chalk, `github.com/perigrin/chalk`: `ARCHITECTURE.md`,
  `docs/architecture/earley-parser.md`,
  `docs/architecture/ambiguity-classes.md`, and the seed grammar at 00399607.
- Rebel's YAEP analysis: `pages/incremental-chart-parsing.md` in perigrin's
  commonplacebook, commit 48d0d34, with each claim marked as read in the
  source, from YAEP's docs, or inferred.
- Kegel, the argument that Perl cannot be statically parsed (BEGIN plus
  prototypes). Cited from memory.
- pvm: `docs/plans/2026-09-05-parser-prior-art.md` §A1.3,
  `docs/plans/2026-09-04-perl-parser-decision.md`,
  `docs/specs/perl-parser/03-lexer-feedback.md`.
