<!-- ABOUTME: Design for a research spike: an Earley chart parser for Perl, built beside the recursive-descent parser. -->
<!-- ABOUTME: States the question, the architecture, the phases with their gates, and the kill criteria; the verdict is perigrin's. -->

# A chart parser for Perl: design for a research spike

**Status:** design for a spike, written 2026-10-29, revised after
crochet:assess round 1 (`docs/assessments/2026-10-29-chart-parser-design.md`).
Not a decision. If the spike is accepted, this becomes a chapter of
`docs/specs/perl-parser/`; until then nothing here describes the parser pvm
ships.

## 1. Why

`docs/plans/2026-09-05-parser-prior-art.md` §A1.3 chose hand-written recursive
descent because it was the only measured architecture in which *"the lexer can
ask the parser what it is looking at."* pvm's parser as shipped keeps that
choice's results (T2 56/56, perl.git `t/` 619/619, M3-directory re-emission
347/347, WRONG = 0), but not its mechanism: its lexer runs up front, with its
own `expect` approximation, and hedges where it cannot know
(`internal/lexer/expect.go`, the `Word` case). An interleaved chart lexer
would give pvm that property for the first time.

What strains is context. How Perl parses depends on what is known at that
point: declared subs, prototypes, features, imports. In the current parser
that knowledge is implicit state. About 49 open v0.1 issues are cases where
the parser disagrees with perl, most of them context-dependent (a
title-keyword count, approximate). The M2 incremental work needed hand-written
invalidation (`sameState`, `keepsLead`, `peekEnd`, `openSquares`), and nearly
every gate review found another piece of hidden state.

An Earley chart parser can make the dependency explicit: the parser reports
every reading and the condition that selects it, and something outside the
parser decides.

## 2. The question

> Can an Earley chart parser reach parity with the floor pvm stands on, at a
> speed worth pursuing, and does it have a better chance than the current
> parser at the parse half of M2 (incremental, correct under edit) and at
> M4–M6 (prototypes and ambiguity, real-world code, conformance)?

Four kinds of evidence answer it, each assigned to a phase in §8:

1. **Defect yield** (phase 3). For each open "perl disagrees" issue a probe
   touches: can the chart fix it, where does the fix land (lexer tiebreak,
   resolver rule or grammar), and how big is it?
2. **Head-to-head on milestone corpora** (phase 3). M4: `t/comp/proto.t`,
   `t/op/signatures.t`, T1. M5: the T4 regex files and CPAN. M6: `t/op`, with
   re-emission excluded, because the chart has no canon.
3. **Cost of change** (phase 3), for each M4–M6 probe, measured on both
   parsers.
4. **Incremental potential** (phase 2 early; phase 5 in full): how soon
   Earley sets match again after an edit, then M2's equivalence contract.
   This covers the parse half of M2 only. The chart cannot feed analysis
   (§11), and M2's ruled target is tree plus published analysis.

### Out of scope

Replacing the production parser; choosing the downstream IR; feeding `lower`,
`infer` or `psc`; a B::SoN semantic oracle (§11).

### Output

A recommendation after phase 3, with two verdicts that may differ: one on the
M4–M6 route and one on the incremental route. Phases 4 and 5 run only if
perigrin rules both positive. The code is kept, and stays experimental until
the spike is accepted.

## 3. Design rules

1. **The grammar is static; the context is dynamic.** Imports and pragmas do
   not change the grammar. They change how tokens are classified and which
   reading applies.
2. **Context enters at two places only:** before the chart (token
   classification, in `chart/lex`) and after it (pruning, in
   `chart/resolve`). Never inside the chart. Facts in Earley items would make
   otherwise-identical sets differ, which kills set interning and with it the
   speed and the incremental route.
3. **Only compile-time facts may prune.** A reading may be removed only by a
   fact perl could know at compile time: a prototype, a declared sub, a
   feature, an import, a library declaration the oracle has vetted. Perl
   always decides from such facts (`foo / 2` divides unless `sub foo` has been
   seen; toke.c `yyl_just_a_word`); an ambiguity in pvm is a fact pvm lacks,
   so every branch has a resolvable condition. Inferred runtime types never
   prune; where they contradict perl's reading, that is a diagnostic.
4. **Ambiguity may stay.** A language server needs the facts true in every
   reading, not one parse. A reading is WRONG if it asserts what perl does
   not; perl's reading among several live ones is *covered but undecided*. §7
   says how each is measured.
5. **Outside the accepted language, refuse; never misparse.** Pest scored
   99.996% and still lost (prior-art §A1.3) because it misparsed silently.
6. **Weights may order, never decide.**

## 4. Sources

pvm's chart parser is an independent design in Go. It borrows from three
sources where the constraints match.

**Guacamole v0.008** (`xsawyerx/guacamole` at 2838253) is the grammar and the
definition of Standard Perl (perigrin's ruling, 2026-10-29). Its G1 rules are
used as written: the `Statement`/`BlockStatement` split, the four-way
precedence clones (`ExprValueU`/`0`/`L`/`R`) and the `NonBrace*` hierarchy for
block versus hash. Chalk's `perl.bnf`@00399607 is not the seed: it had lost
the clones and `NonBrace*` in earlier simplifications (28c6029e and after),
and accepts `1 2` and `$x $y`.

**Chalk** (`perigrin/chalk`) contributes proven techniques: integer core-item
IDs, nullable sets and Aycock–Horspool dot advancement (`LR0DFA.pm`), terminal
clustering, the relative-distance chart, its seven ambiguity classes
(`docs/architecture/ambiguity-classes.md`), and Leo with an equivalence test.
It diverges in being scanless, resolving ambiguity during the parse through
semirings (it builds IR directly), and having no forest.

**YAEP** (`vnmakarov/yaep` at 7aba93e) contributes set interning: situations,
kernel-deduplicated cores, deduplicated distance vectors, whole sets interned
as (core, distances), per-(core, symbol) transition and reduce lists, and the
(set, token, lookahead) transition cache. Its forest is post-hoc: it stores no
links in items and finds origins by walking the shared sets (`make_parse`).
Its DAG is not borrowed: it is per rule instance, carries costs rather than
conditions, and drops the concrete tree.

## 5. Components

A new tree, `internal/chart/`. It depends on `internal/lexer`; only
`chart/compare` also imports `internal/parse`.

### `chart/grammar`

Reads the **subset of Marpa's SLIF notation that Guacamole uses** (perigrin's
ruling, 2026-10-29): G1 rules (`::=`, `|`), `:lexeme ~ X priority => n`,
`:discard`, and `lexeme default = latm => 1`. `action =>` adverbs are
ignored. G0 rules (`~`) are not run: each lexeme name maps to a **token
predicate** (lexer kind plus text), because one lexer kind covers many
lexemes (every bracket is `Operator`, every quote-like is `Quote`) and keyword
sets overlap. Builtin shapes (named unary, list operator, niladic) come from
CORE.tpm, the one source `internal/parse` already reads, not from a second
table.

A generator run by `go generate` builds the tables ahead of time: core items,
the LR(0) automaton with nullable closure, per-state transition and reduce
lists, per-state expected-terminal sets and Leo eligibility, and expectation
classes.

**Expectation classes** are the projection of a position's expected-terminal
set onto the token pairs that conflict lexically: `/` as division or regex,
`{` as block or hash, `<` as less-than or readline, and `%`, `&`, `*`, `?`,
`-X`, `<<`, `x`. A class says which side or sides are expected. The lexer
sees the union over every state in the Earley set at that position.

### `chart/lex`

The scanners of `internal/lexer`, refactored so a scanner takes its
expectation as a parameter instead of reading and advancing the lexer's own
state (perigrin's ruling, 2026-10-29). That is a production change: it goes
through mainline gates and `TestLexerRatchet` as ordinary pvm work, outside
the spike's isolation (§10). There is then one lexer, driven by its own
`expect` machine for `internal/parse` and by expectation classes for the
chart.

The lexer scans only the kinds its class allows (Marpa's `latm`), so a reading
dies at its first unexpected token. Where several expected kinds match and
perl decides by heuristic (`{` through `intuit_more`), Guacamole's lexeme
priorities and perl's tiebreaks decide. Where pvm lacks a fact perl would
have (rule 3), the lexer emits a **token branch**.

**Input model.** Positions are the nodes of a token lattice: the byte offsets
of token boundaries. A branch forks the lattice into paths of possibly
different token counts and rejoins it at the first byte offset both paths
reach; a branch that has not rejoined after a bounded number of tokens is
refused. Relative distances count lattice nodes along a topological
numbering. Sets intern in branch-free regions; sets inside a branch are not
expected to.

**One coupled choice: lexing mode, lookahead, list shape and Leo.** Lookahead
needs the next token before the current set is built. Interleaved lexing
cannot supply it, because lexing token *k+1* needs the set at *k+1*. So:

- up front, with lookahead 1 and the right-recursive statement list, is the
  default. A YAEP probe ran 100k tokens in 0.007 s with 11 unique sets;
- interleaved forces lookahead 0, where the same list ran 87.6 s with 20,007
  unique sets. Leo must then work from phase 1, verified by the test in §10
  that fails when Leo does nothing.

Perl's own `PL_expect` is not evidence for up-front lexing: perly.y writes
`parser->expect` in 18 places, and toke.c sets it from the symbol table.
Whether grammar-derived classes can run up front is measured in phase 1.

### `chart/recognize`

A YAEP-style Earley recognizer over the generated tables and the lattice:
interned sets, relative distances, the transition cache, and Leo memos stored
by reference, with each resolved top cached in a per-position side table
outside the set's identity. It exposes `Expected(pos)`.

### `chart/forest`

A Scott-style SPPF built **after** recognition, by walking the interned sets
the way YAEP's `make_parse` does. Items never hold forest pointers: Scott's
in-recognition construction would defeat interning. Packed nodes may carry a
condition, and Leo's skipped intermediate nodes are recovered during the
walk.

### `chart/resolve`

Prunes packed nodes using compile-time facts (rule 3); the rest stays
ambiguous. Built only as the probes in phase 3 need it.

### `chart/compare`

Exports each parser's tree as spans and diffs the two per file, under a
normalisation fixed before phase 1: byte spans, unary-chain collapse,
statement-list flattening, and a label map that is explicit for Chalk's seven
ambiguity classes and coarse elsewhere.

## 6. Data flow

```
source bytes
   │
   ▼
chart/lex ── expectation class per lattice node ───────────────┐
   │  scans only the expected kinds; lexeme priorities and      │
   │  perl's tiebreaks; a token branch where pvm lacks a fact    │
   ▼                                                            │
token lattice (lossless: trivia attached)                       │
   │                                                            │
   ▼                                                            │
chart/recognize ── interned sets (state, distances) ────────────┘
   │  a reading dies at its first unexpected token
   │  if every reading dies: error recovery (`error` symbol, minimal skip)
   ▼
chart/forest ── post-hoc SPPF; packed nodes may carry conditions
   │
   ▼
chart/resolve ── compile-time facts only; the rest stays ambiguous
   │
   ├──▶ chart/compare: normalised spans → diff against recursive descent;
   │                    a disagreement → perl -MO=Deparse,-p decides
   ├──▶ parseoracle: forest → SubjectFacts (§7)
   └──▶ future: LSP features ask "do all readings agree?"
```

## 7. The parity contract

Parity with the recursive-descent parser means all four:

1. **The oracle,** with the forest reported as a subject. A marker site
   present in every live reading is reported decided. A site present in only
   some readings is reported unresolved, which scores *wider*, never exact.
   Coverage and exactness must be at or above the recursive-descent ratchet,
   and WRONG = 0. The oracle sees only five marker kinds and never sees
   surplus (`compareMarkers`): on the phase-1 corpus only 68 of 730 snippets
   carry a marker site. So it is necessary, not sufficient.
2. **Yield.** The yield of every packed node is exactly its byte span, and
   the root's yield is the source. (A round-trip from the tokens alone would
   pass by construction.)
3. **A per-file span diff** against recursive descent, under §5's
   normalisation, wherever both produce one reading.
4. **An arbiter for every disagreement:** perl's own grouping, from
   `perl -MO=Deparse,-p` (perigrin's ruling, 2026-10-29). The parser whose
   grouping matches perl's is right. If both match, the shape difference is
   legitimate (07 §7.5.2). Where Deparse loses the form, the disagreement is
   recorded as undecided, reported by count, and does not count towards
   parity. The same check catches an asserted reading perl does not take,
   which the oracle cannot see.

Canon re-emission stays the recursive-descent parser's, so the chart's M6
verdict is on M6 without re-emission. Chalk's re-emission lowers its SoN IR,
not a CST, and does not serve here.

## 8. Phases and gates

The order is perigrin's ruling (2026-10-29): Standard Perl, then cheap early
measures and the probes, then full parity only if those come back positive.

**Budgets.** Before each phase starts, refinement records its budget in the
milestone as an issue count and a calendar estimate. A phase open beyond
twice its budget forces a decision, continue or stop, recorded with its
reason. A cold-parse tripwire on Standard Perl inputs runs at every step,
as a warning.

### Phase 1: Standard Perl

- **Build:** the grammar reader and generator, `chart/lex` (with the lexer
  refactor), `chart/recognize`, `chart/forest`, `chart/compare`, the oracle
  subject adapter, and one token-branch case (an undeclared word followed by
  `/`) that exercises the lattice.
- **Corpus:** Guacamole's suite, extracted by running its tests with a
  recording `Guacamole::Test` (791 `parses`, 126 `parsent` by A's count),
  checked in, and reviewed. The feature bundle Standard Perl assumes is
  declared as a preamble (`use feature ':5.36'`). `parsent` means "not
  Standard Perl", not "not Perl", so its labels are correct by definition. A
  `parses` case perl rejects (23 under perl 5.42.0) is a Guacamole erratum:
  recorded, and refused (rule 5).
- **Gate:**
  - every `parses` case perl accepts, accepted, under all four parts of §7;
  - every `parsent` case and every erratum refused, with an Error node;
  - the expectation-class comparison done: an explicit map from each of
    spec 03's eleven `PL_expect` states to the conflicts it settles, and the
    count of conflict decisions where the generated classes differ. The
    lexing mode (up front or interleaved) is recorded from it.

### Phase 2: early measures, on Standard Perl

- Cold parse time and memory against recursive descent on the accepted
  phase-1 corpus, and on Standard-Perl source files.
- Leo on and off, set-reuse rate, automaton state count; Aycock–Horspool
  automaton states against per-item cores, as a measured option.
- Incremental measure (a): tokens past an edit until the new sets match the
  old ones again, over single-statement edits.
- **Gate:** perigrin's ruling, against thresholds written here before the
  numbers arrive: a cold parse above 3× recursive descent on the same input,
  with no single hot spot accounting for half the gap, is a speed kill;
  sets that rarely match again within one statement of the edit make the
  incremental verdict "keep the snapshot design".

### Phase 3: M4–M6 probes

- **The sample** is fixed before the phase starts: two or three probes per
  milestone, drawn from open issues. Candidates: `first { ... } @list`; a
  named unary versus a list operator with an unknown prototype; `sort
  $coderef @list`; `s///e`; `(?{ ... })`; unusual delimiters; failing
  constructs from `t/base/lex.t` and `t/comp/parser.t`.
- **Each probe** is built on the Standard grammar plus only the context
  pieces it needs (`chart/resolve` and lexer facts), and measured on both
  parsers: lines changed, places touched, where the fix landed. Shared chart
  infrastructure is charged in equal parts to every probe that uses it. On
  the recursive-descent side, a closed issue's cost is its PR; an open one is
  implemented as ordinary mainline work through mainline gates, outside the
  spike's isolation. Which side goes first alternates.
- **"Clearly cheaper"** means at most half the lines and half the places
  touched. The M4–M6 verdict is positive if at least half the probes are
  clearly cheaper on the chart and none needed a grammar rule written for a
  single probe.
- The defect-yield and head-to-head evidence (§2, 1 and 2) is collected here.
- **Gate:** the recommendation, with both verdicts.

### Phase 4: parity (only if both verdicts are positive)

Guacamole's exclusion list, one feature at a time: auto-quoting; `print
{$fh}` forms and bareword filehandles; unusual quote delimiters; heredocs;
paren-less calls; prototype-driven parses; indirect objects;
`given`/`when`. Then require/use loading, which perl.git `t/`'s 619/619
depends on. Each step records `parsent` cases flipped to correct parses.
**Gate:** §7 parity with the floor: T2 56/56, perl.git `t/` 619/619, T1 and
the corpus and site ratchets at or above recursive descent's.

### Phase 5: incremental (after phase 4)

Measures (b) and (c), then M2's contract: a repair returns a tree equal to a
fresh parse or refuses, under FuzzIncremental and FuzzIncrementalSequence
retargeted at the chart. Context edits invalidate everything after them, by
design.

## 9. Risks and kill criteria

Fixed before work starts.

| Risk | Early signal | Kill or decline criterion |
|---|---|---|
| Silent misparse | A `parsent` case or erratum accepted; WRONG > 0; a Deparse-decided disagreement lost by the chart | A WRONG or lost disagreement whose only fix is a refusal that drops coverage below recursive descent's on that file |
| Forest blow-up | Readings per symbol node, counted by sum-product over packed nodes | A symbol node in a real file with more than 2^10 readings after resolution, where perl's reading is not decidable by compile-time facts |
| Speed | Phase 2; the tripwire | Phase 2's threshold. Recursive descent's cold parse on edit5k.pl is 23.89 ms p50 and 33.42 ms p99; a YAEP-class extrapolation for 5 kLOC is about 13.6 ms. Both are hypotheses for phase 2 to test |
| Expectation classes differ | The phase-1 count | More than 2 of the 11 states' conflict decisions unmatched: investigated before phase 2 |
| Context leaking into the chart | Set reuse falling as probes add pieces | Any fix needing facts inside items: a design review |
| No incremental convergence (Wiren 1989) | Phase 2's measure (a) | The incremental verdict becomes "keep the snapshot design" |
| Probes show no advantage | The phase-3 table | Phase 3's threshold not met |
| Drift and stall | Budgets (§8) | A phase beyond twice its budget forces a recorded decision |

## 10. Testing

pvm's conventions apply: TDD, the pre-commit hooks, `make test` for code.

- **Generator:** golden tests on small grammars (nullables, left and right
  recursion); byte-identical tables on regeneration; expectation classes
  matching a hand computation.
- **Recognizer:** accept and reject sets of toy grammars; positions with
  equal content share one interned set; Leo on and off give identical
  forests; and a test that **fails if Leo does nothing**: item count or time
  on `S → a S` at lookahead 0 must be linear.
- **Forest:** every packed node's yield is its span; known-ambiguous toy
  inputs yield exactly their readings.
- **Lexer refactor:** `TestLexerRatchet` and the existing lexer corpus,
  unchanged under the refactor; each scanner yields today's tokens under the
  expectation it is given.
- **Compare:** the normalisation tested against hand-built trees, with one
  check that must pass for each that must fail: a known disagreement is
  reported, identical trees report none.
- **Ratchets:** chart ratchets fail in both directions, like pvm's, and gate
  the spike's branch, not mainline `make test`.
- **Fuzzing:** recognizer and forest invariants in phase 1; the incremental
  fuzzers in phase 5.
- **Isolation:** `internal/parse`, `lower`, `infer` and `psc` do not change
  for the chart. Two things are outside the isolation and are ordinary
  mainline work: the `internal/lexer` refactor, and recursive-descent probe
  work. `chart/compare` imports `internal/parse` to export its trees.

## 11. Future work, not in this spike

- **A B::SoN semantic oracle.** B::SoN (`perigrin/perl5-son`) turns perl's
  compiled op tree into a Sea-of-Nodes graph, serialized as deterministic JSON
  (`docs/plans/2026-04-07-son-json-serialization-design.md` there). A
  resolver-to-SoN builder would let pvm diff a reading's semantics against
  `perl -MO=SoN,json`. The v1 schema carries no source positions, so the IR
  cannot replace the forest for position-based LSP features, and a graph is
  one resolved meaning.
- **The downstream IR.** Until it is settled the chart cannot feed analysis,
  which is why §2 covers only the parse half of M2.

## 12. References

- Aycock and Horspool, "Practical Earley Parsing" (2002); Aycock's thesis
  (2001), in Chalk as `docs/Aycock_JohnDaniel_PhD_2001.pdf`.
- Leo (1991). Scott (2008), "SPPF-style parsing from Earley recognisers".
  Wiren (1989), EACL. Wagner and Graham (PLDI 1997).
- YAEP at 7aba93e: `src/yaep.c`, `doc/Internals.txt`; README benchmarks are
  the author's own.
- Guacamole v0.008 at 2838253: `lib/Guacamole.pm`, `lib/standard.pm`, `t/`.
- Chalk: `ARCHITECTURE.md`, `docs/architecture/earley-parser.md`,
  `docs/architecture/ambiguity-classes.md`.
- Rebel's YAEP analysis: `pages/incremental-chart-parsing.md` in perigrin's
  commonplacebook, commit 48d0d34.
- Kegel, the argument that Perl cannot be statically parsed. Cited from
  memory.
- pvm: `docs/plans/2026-09-05-parser-prior-art.md` §A1.3,
  `docs/plans/2026-09-04-perl-parser-decision.md`,
  `docs/plans/2026-09-21-cst-is-the-source-of-truth.md`,
  `docs/specs/perl-parser/03-lexer-feedback.md`, `07-conformance.md`,
  `internal/parseoracle/README.md`.
