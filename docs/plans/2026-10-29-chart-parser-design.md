<!-- ABOUTME: Design for a research spike: an Earley chart parser for Perl, built beside the recursive-descent parser. -->
<!-- ABOUTME: States the question, the architecture, the phases with their gates, and the kill criteria; the verdict is perigrin's. -->

# A chart parser for Perl: design for a research spike

**Status:** design for a spike, written 2026-10-29, revised after crochet:assess
rounds 1 and 2 (`docs/assessments/2026-10-29-chart-parser-design.md`). Not a
decision. If the spike is accepted, this becomes a chapter of
`docs/specs/perl-parser/`; until then nothing here describes the parser pvm
ships.

## 1. Why

`docs/plans/2026-09-05-parser-prior-art.md` §A1.3 chose hand-written recursive
descent because it was the only measured architecture in which *"the lexer can
ask the parser what it is looking at."* pvm's parser as shipped keeps that
choice's results (T2 56/56, perl.git `t/` 619/619, M3-directory re-emission
347/347, WRONG = 0), but not its mechanism: its lexer runs up front with its
own `expect` approximation, and hedges where it cannot know
(`internal/lexer/expect.go`, the `Word` case). An interleaved chart lexer, one
that asks the chart what it expects, would give pvm that property for the
first time. Whether that is affordable is the first thing this spike measures.

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

> Is an Earley chart parser worth taking to parity with the floor pvm stands
> on? Specifically: does it have a better chance than the current parser at
> M4–M6 (prototypes and ambiguity, real-world code, conformance), and at the
> parse half of M2 (incremental, correct under edit), at a speed worth
> pursuing?

Four kinds of evidence answer it, each assigned to a phase in §8:

1. **Speed and incremental potential, early** (phase 0): cold recognition time
   against recursive descent, set reuse, and how soon Earley sets match again
   after an edit.
2. **Cost of change and defect yield** (phase 2): for each M4–M6 probe,
   measured on both parsers, and the open "perl disagrees" issues each probe
   settles.
3. **Head-to-head on milestone corpora** (phase 3, once the chart parses those
   files): M4 `t/comp/proto.t`, `t/op/signatures.t`, T1; M5 the T4 regex files
   and CPAN; M6 `t/op`, without re-emission, because the chart has no canon.
4. **M2's contract** (phase 4): equivalence under both incremental fuzzers,
   and repair latency against the parse share of M2's budget.

The chart cannot feed analysis (§11), and M2's ruled target is a re-parsed tree
*and* its published analysis in under 10 ms p99. So no result here establishes
M2's latency gate; phase 4 reports the parse share only.

### Out of scope

Replacing the production parser; choosing the downstream IR; feeding `lower`,
`infer` or `psc`; a B::SoN semantic oracle (§11).

### Output

A recommendation after phase 2, with two verdicts that may differ: the M4–M6
route and the incremental route. A positive M4–M6 verdict alone unlocks
phase 3 (parity); a negative incremental verdict means M2 keeps its snapshot
design and phase 4 is skipped (perigrin's ruling, 2026-10-29). The code is
kept, and stays experimental until the spike is accepted.

## 3. Design rules

1. **The grammar is static; the context is dynamic.** Imports and pragmas
   change how tokens are classified and which reading applies, not the
   grammar.
2. **Context enters at two places only:** before the chart (token
   classification, in `chart/lex`) and after it (pruning, in
   `chart/resolve`). Never inside the chart: facts in Earley items would make
   otherwise-identical sets differ, which kills set interning.
3. **Only compile-time facts may prune.** A reading may be removed only by a
   fact perl had at compile time: a prototype, a declared sub, a feature, an
   import, a library declaration the oracle has vetted. Perl decides every
   case from such facts (`foo / 2` divides unless `sub foo` has been seen;
   toke.c `yyl_just_a_word`). Some of those facts pvm can obtain only by
   running code (a prototype installed by a string `eval` in `BEGIN`; spec
   03 §3.0), and those readings stay ambiguous. Inferred runtime types never
   prune; where they contradict perl's reading, that is a diagnostic.
4. **Ambiguity may stay.** A language server needs the facts true in every
   reading. A reading is WRONG if it asserts what perl does not; perl's
   reading among several live ones is *covered but undecided*. §7 says how
   each is measured.
5. **Outside the accepted language, refuse; never misparse.** Pest scored
   99.996% and still lost (prior-art §A1.3) because it misparsed silently.
6. **Weights may order readings, never decide them.** Corpus frequencies may
   order what a user is shown. A lexeme priority is not a weight: it is a
   deterministic tiebreak rule, like perl's own, applied in `chart/lex`.

## 4. Sources

pvm's chart parser is an independent design in Go, borrowing from three
sources where the constraints match.

**Guacamole v0.008** (`xsawyerx/guacamole` at 2838253) is the grammar and the
definition of Standard Perl (perigrin's ruling, 2026-10-29). Its G1 rules are
used as written above pvm's token level: the `Statement`/`BlockStatement`
split, the four-way precedence clones (`ExprValueU`/`0`/`L`/`R`), the
`NonBrace*` hierarchy for block versus hash, and its per-keyword builtin
rules (about 200, e.g. `OpKeywordTimeExpr ::= OpKeywordTime LParen RParen`).
Rules below pvm's token level are collapsed (§5). Chalk's
`perl.bnf`@00399607 is not the seed: it had lost the clones and `NonBrace*`
in earlier simplifications (28c6029e and after), and accepts `1 2`.

**Chalk** (`perigrin/chalk`) contributes techniques: integer core-item IDs,
nullable sets and Aycock–Horspool dot advancement (`LR0DFA.pm`), terminal
clustering, the relative-distance chart, its seven ambiguity classes
(`docs/architecture/ambiguity-classes.md`), and Leo with an equivalence test.
It diverges in being scanless, resolving ambiguity during the parse through
semirings (it builds IR directly), and having no forest.

**YAEP** (`vnmakarov/yaep` at 7aba93e) contributes set interning: situations,
kernel-deduplicated cores, deduplicated distance vectors, whole sets interned
as (core, distances), per-(core, symbol) transition and reduce lists, and the
(set, token, lookahead) transition cache. Its forest is post-hoc: items hold no
links, and origins are found by walking the shared sets (`make_parse`).

## 5. Components

A new tree, `internal/chart/`. It depends on `internal/lexer`. `chart/compare`
imports `internal/parse` to export its trees, and the grammar generator may
import it at `go generate` time.

### `chart/grammar`

Reads the **subset of Marpa's SLIF notation that Guacamole uses** (perigrin's
ruling, 2026-10-29): G1 rules (`::=`, `|`), `:lexeme ~ X priority => n`,
`:discard`, and `lexeme default = latm => 1`. `action =>` adverbs are
ignored. Guacamole has no SLIF prioritised (`||`) rules.

**The token cut** (perigrin's ruling, 2026-10-29). G1 builds some single pvm
tokens from several lexemes (`VarScalar ::= SigilScalar VarIdentExpr
ElemSeq0`). Rules below pvm's token level are collapsed into **token
predicates** (lexer kind plus text), listed beside the grammar and reviewed;
each remaining G0 lexeme maps to one, because one lexer kind covers many
lexemes.

**Builtin shapes** come from G1's per-keyword rules, as written. CORE.tpm is a
cross-check: the generator, importing `internal/parse` at `go generate` time,
lists every builtin whose CORE.tpm shape differs from G1's, and that list is
reviewed. It is not a second source.

A generator, run through a Makefile target (CLAUDE.md: the Makefile for all
builds), builds the tables ahead of time: core items, the LR(0) automaton with
nullable closure, per-state transition and reduce lists, per-state
expected-terminal sets and Leo eligibility, and expectation classes.

**Expectation classes** are the projection of a position's expected-terminal
set onto the token pairs that conflict lexically in this grammar: `/` as
division or regex, `<` as less-than or readline, `<<` as shift or heredoc, and
`%`, `&`, `*`, `?`, `-X`, `x`. (`{` is not among them: G1 has one `LBrace` and
tells block from hash with `NonBrace*`.) A class says which side or sides are
expected, over the union of every state in the Earley set at that position.

### `chart/lex`

**Two modes, chosen by phase 0.**

- **Up front:** the production lexer, with its own `expect` machine, produces
  the token stream before recognition, as `internal/parse` uses it today. No
  expectation classes are involved. This is phase 0's mode.
- **Interleaved:** before each token the lexer reads the expectation class
  from the current Earley set and scans only the kinds it allows (Marpa's
  `latm`). This delivers §1's property. It needs `internal/lexer` refactored
  so a scanner takes its expectation and context as parameters instead of
  advancing its own state (perigrin's ruling, 2026-10-29): mainline work
  through `TestLexerRatchet`, counted in phase 1's budget, done only if phase 0
  passes, and kept after a stopped spike only on its own merits.

**One coupled choice: lexing mode, lookahead, list shape and Leo.** An
interleaved lexer cannot supply lookahead, because lexing token *k+1* needs
the set at *k+1*. So interleaved means lookahead 0, where G1's right-recursive
`StatementSeq` is quadratic without Leo (a toy YAEP probe: 0.007 s at
lookahead 1 against 87.6 s at lookahead 0, over 100k tokens). Interleaved
therefore needs Leo from phase 1, and has lower set reuse by construction.
Phase 0 re-measures this on G1.

**Ties.** Where several expected kinds match, Guacamole's lexeme priorities and
perl's own heuristics (`{` and `intuit_more`) decide. Where pvm lacks a fact
perl had (rule 3), the lexer emits a **token branch**.

**Input model.** Positions are the nodes of a token lattice: byte offsets of
token boundaries, each paired with a full lexer state. A branch forks the
lattice into paths of possibly different token counts. Two paths **rejoin**
only where both byte offset and lexer state are equal (the rule
`TokenCache.resyncsAt` and `State.Equal` already use), so `foo <<EOF;` does
not rejoin after `EOF`, where only one path has a heredoc pending. A branch
not rejoined within a bounded number of tokens is refused. Relative distances
count lattice nodes; sets intern outside branches only.

### `chart/recognize`

A YAEP-style Earley recognizer over the tables and the lattice: interned sets,
relative distances, the transition cache (lookahead 0 and no cache inside a
branch), and Leo memos stored by reference, with each resolved top in a
per-position side table outside the set's identity. It exposes
`Expected(pos)`.

### `chart/forest`

A Scott-style SPPF built **after** recognition, by walking the interned sets
as YAEP's `make_parse` does. Items never hold forest pointers: Scott's
in-recognition construction would defeat interning. Packed nodes may carry a
condition, and Leo's skipped intermediate nodes are recovered during the walk.

### `chart/resolve`

Prunes packed nodes using compile-time facts (rule 3); the rest stays
ambiguous. Built only as phase 2's probes need it.

### `chart/compare`

Exports each parser's tree as spans and diffs the two per file, under a
normalisation fixed before phase 1: byte spans, a trivia attachment rule,
unary-chain collapse, statement-list flattening, and a label map that is
explicit for Chalk's seven ambiguity classes and coarse elsewhere.

## 6. Data flow

```
source bytes
   │
   ▼
chart/lex ── up front: production expect machine
   │         interleaved: expectation class per lattice node ──┐
   │  lexeme priorities and perl's tiebreaks; a token branch    │
   │  where pvm lacks a fact                                     │
   ▼                                                             │
token lattice (lossless: trivia attached)                        │
   │                                                             │
   ▼                                                             │
chart/recognize ── interned sets (state, distances) ─────────────┘
   │  a reading dies at its first unexpected token
   │  if every reading dies: error recovery (`error` symbol, minimal skip)
   ▼
chart/forest ── post-hoc SPPF; packed nodes may carry conditions
   │
   ▼
chart/resolve ── compile-time facts only; the rest stays ambiguous
   │
   ├──▶ chart/compare: normalised spans → diff against recursive descent
   │                    → disagreement → the arbiter (§7.4)
   ├──▶ parseoracle: forest → SubjectFacts (§7.1)
   └──▶ future: LSP features ask "do all readings agree?"
```

## 7. The parity contract

Parity with the recursive-descent parser means all four:

1. **The oracle,** with the forest reported as a subject. A marker site
   present in every live reading is reported decided; one present in only some
   readings is reported unresolved, which scores *wider*, never exact.
   Coverage and exactness must be at or above recursive descent's measured
   score on the same corpus, and WRONG = 0. The oracle sees only five marker
   kinds and never sees surplus (`compareMarkers`): of the 730 perl-accepted
   `parses` cases (participant A's count, which keeps the suite's repeats),
   only 68 carry a marker site. So it is necessary, not sufficient.
2. **Yield.** The yield of every packed node is exactly its byte span, and the
   root's yield is the source.
3. **A per-file span diff** against recursive descent, under §5's
   normalisation, wherever both produce one reading.
4. **An arbiter for every disagreement,** built in phase 1 on perl's
   `B::Deparse` (perigrin's ruling, 2026-10-29). Where `perl -MO=Deparse,-p` of
   recursive descent's canon output equals that of the source (pvm's existing
   Faithful check), recursive descent's grouping is perl's, and a chart span
   that differs is wrong. Otherwise, including where Deparse loses the form
   (it folds `1 + 2 * 3` to `7`), the disagreement is **undecided**, and
   undecided disagreements may be at most 10% of all disagreements on the
   corpus. This also catches asserted readings perl does not take.

Canon re-emission stays recursive descent's, so the chart's M6 verdict is on
M6 without re-emission. Chalk's re-emission lowers its SoN IR, not a CST, and
does not serve here.

## 8. Phases and gates

The order is perigrin's (rulings of 2026-10-29): a cheap measurement first,
then Standard Perl, then the probes, then parity only on a positive M4–M6
verdict.

**Budgets.** Before each phase starts, refinement records its budget in the
milestone as a calendar estimate and an issue count; the calendar estimate
governs. The orchestrating session checks the budget at every issue close. A
phase open beyond twice its budget forces a decision by perigrin, continue or
stop, recorded with its reason; continuing records a new budget, and the rule
applies again.

**Fixed inputs.** The speed comparisons run on `edit5k-standard.pl`, a copy of
`internal/psc/testdata/latency/edit5k.pl` with its 70 heredocs rewritten as
Standard Perl, committed beside it. Recursive descent is timed on that same
file. "Cold parse" means `parse.Parse` on the recursive-descent side, and
lexing, recognition and forest construction on the chart's, without
`chart/resolve`.

### Phase 0: measure first

- **Build:** the grammar reader and generator (with the token cut), and
  `chart/recognize`, recognition only, over tokens from today's production
  lexer, up front.
- **Measure:** cold recognition time and memory against recursive descent on
  `edit5k-standard.pl`; Leo on and off; set reuse; automaton state count;
  Aycock–Horspool states against per-item cores; and convergence: over 200
  single-statement edits to `edit5k-standard.pl`, the number of tokens past
  the edit until the new sets match the old ones.
- **Gate** (thresholds written before the numbers arrive):
  - **speed kill:** recognition alone above 3× recursive descent's cold
    parse, unless removing the named hot spots brings it under 3×;
  - **incremental:** if fewer than 80% of the edits converge within the
    edited statement plus 8 tokens, the incremental verdict is "keep the
    snapshot design";
  - a fired kill ends the spike unless perigrin records an override and its
    reason.

### Phase 1: Standard Perl

- **Build:** the `internal/lexer` refactor and interleaved `chart/lex` (if
  phase 0 kept interleaving viable), `chart/forest`, `chart/compare`, the
  arbiter (§7.4), the oracle subject adapter, and one token-branch case: an
  undeclared word followed by `/`, which exercises the lattice and its rejoin
  rule.
- **Corpus:** Guacamole's suite, extracted by running its tests with a
  recording `Guacamole::Test` (791 `parses` and 126 `parsent` calls, 769 and
  124 unique), checked in and reviewed, under a declared preamble
  (`use feature ':5.36'`). `parsent` means "not Standard Perl", so its labels
  are correct by definition. Cases needing a missing module run with stubs on
  `@INC`. A `parses` case perl still rejects is a Guacamole erratum, excluded
  from scoring as spec 07 §7.5.2 excludes any input perl fails to compile.
- **Gate:**
  - every perl-accepted `parses` case accepted, under all four parts of §7;
  - every `parsent` case refused, with an Error node;
  - the expectation-class comparison: an explicit map from each of spec 03's
    eleven `PL_expect` states to the conflicts it settles, and the number of
    states whose conflict decisions the generated classes do not reproduce.
    More than 2 forces a recorded decision by perigrin, continue or stop.

### Phase 2: M4–M6 probes

- **The sample** is fixed before the phase starts: two or three probes per
  milestone, drawn from open issues. Candidates: `first { ... } @list`; a
  named unary versus a list operator with an unknown prototype;
  `sort $coderef @list`; `s///e`; `(?{ ... })`; unusual delimiters; failing
  constructs from `t/base/lex.t` and `t/comp/parser.t`.
- **Each probe builds the excluded features it needs** (perigrin's ruling,
  2026-10-29), such as prototype-driven parses or paren-less calls for
  `first { ... } @list`, with that cost charged to the probe. It is measured on
  both parsers: lines changed, places touched, and where the fix landed. Shared
  chart infrastructure is charged in equal parts to every probe that uses it.
  On the recursive-descent side, a closed issue's cost is its PR; an open one
  is implemented as ordinary mainline work through mainline gates. Which side
  goes first alternates.
- **"Clearly cheaper"** means at most half the lines and half the places
  touched. A feature Standard Perl excludes, built for a probe, is not a
  grammar hack; a rule added for a single probe that no Perl construct
  justifies is. **The M4–M6 verdict is positive** if at least half the probes
  are clearly cheaper on the chart and none needed such a hack.
- **Gate:** the recommendation, with both verdicts.

### Phase 3: parity (only if the M4–M6 verdict is positive)

The rest of Guacamole's exclusion list, one feature at a time, skipping what
phase 2 already built: auto-quoting; `print {$fh}` forms and bareword
filehandles; unusual quote delimiters; heredocs; paren-less calls;
prototype-driven parses; indirect objects; `given`/`when`. Then require/use
loading, which perl.git `t/`'s 619/619 depends on. Each step records `parsent`
cases flipped to correct parses. The milestone-corpus head-to-heads (§2,
evidence 3) run here, once the chart parses those files.
**Gate:** §7 parity with the floor: T2 56/56, perl.git `t/` 619/619, T1 and the
corpus and site ratchets at or above recursive descent's.

### Phase 4: incremental (only if the incremental verdict is positive)

The full convergence measures over phase 3's corpora: (a) tokens past the edit
until the sets match again, (b) the fraction of sets that match after that, and
(c) the transition-cache hit rate. Then M2's contract: a repair returns a tree
equal to a fresh parse or refuses, under FuzzIncremental and
FuzzIncrementalSequence retargeted at the chart, with repair latency reported
against the parse share of M2's 10 ms. Context edits invalidate everything
after them, by design.

## 9. Risks and kill criteria

Fixed before work starts. A fired kill ends the spike unless perigrin records
an override and its reason.

| Risk | Early signal | Kill criterion |
|---|---|---|
| Speed | Phase 0 | Phase 0's speed threshold. Recursive descent's cold parse on `edit5k.pl` is 23.89 ms p50 and 33.42 ms p99; a YAEP-class extrapolation for 5 kLOC is about 13.6 ms. Both are hypotheses, re-measured on `edit5k-standard.pl` |
| Silent misparse | A `parsent` case accepted; WRONG > 0; an arbiter-decided disagreement lost by the chart | A WRONG or lost disagreement that is fixed only by a refusal which drops coverage below recursive descent's on that file |
| Forest blow-up | Readings per packed node: the sum-product count of its sub-forest, stopping at nested packed nodes | A packed node in a real file with more than 2^10 readings after resolution, where perl's reading is decidable from facts pvm has |
| Undecided disagreements | The arbiter's undecided count | More than 10% of disagreements undecided at a phase gate |
| Expectation classes | Phase 1's state count | More than 2 of 11 states unreproduced: a recorded decision |
| Context leaking into the chart | Set reuse falling as probes add pieces | Any fix needing facts inside items: a design review |
| No incremental convergence (Wiren 1989) | Phase 0's convergence measure | The incremental verdict becomes "keep the snapshot design" |
| Probes show no advantage | The phase-2 table | Phase 2's threshold not met |
| Drift and stall | Budgets (§8) | A phase beyond twice its budget forces a recorded decision |

## 10. Testing

pvm's conventions apply: TDD, the pre-commit hooks, `make test` for code.

- **Generator:** golden tests on small grammars (nullables, left and right
  recursion); byte-identical tables on regeneration.
- **Recognizer:** positions with equal content share one interned set; Leo on
  and off give identical forests; a test that **fails if Leo does nothing**
  (`S → a S` at lookahead 0 must be linear); the round-1 ambiguity probes
  (`1 2`, `$x $y`) re-run on the collapsed grammar.
- **Lattice and forest:** `foo <<EOF;` does not rejoin early; every packed
  node's yield is its span.
- **Lexer refactor:** `TestLexerRatchet` and the existing lexer corpus,
  unchanged; each scanner yields today's tokens given today's context.
- **Compare and arbiter:** the normalisation tested against hand-built trees,
  with one check that must pass for each that must fail: a known disagreement
  is reported and decided, identical trees report none.
- **Ratchets:** chart ratchets fail in both directions, like pvm's.
  `internal/chart` builds behind a `chart` build tag, so mainline
  `make test` does not run it; a `make test-chart` target runs it, and gates
  the spike's work.
- **Fuzzing:** recognizer and forest invariants from phase 0; the incremental
  fuzzers in phase 4.
- **Isolation:** `internal/parse`, `lower`, `infer` and `psc` do not change for
  the chart. Two things are ordinary mainline work outside the isolation: the
  `internal/lexer` refactor, and recursive-descent probe work.

## 11. Future work, not in this spike

- **A B::SoN semantic oracle.** B::SoN (`perigrin/perl5-son`) turns perl's
  compiled op tree into a Sea-of-Nodes graph, serialized as deterministic JSON
  (`docs/plans/2026-04-07-son-json-serialization-design.md` there). A
  resolver-to-SoN builder would let pvm diff a reading's semantics against
  `perl -MO=SoN,json`. The v1 schema carries no source positions, and a graph
  is one resolved meaning.
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
- perl: `perly.y` writes `parser->expect` in 14 places; toke.c sets it from the
  symbol table.
- Rebel's YAEP analysis: `pages/incremental-chart-parsing.md` in perigrin's
  commonplacebook, commit 48d0d34.
- Kegel, the argument that Perl cannot be statically parsed. Cited from
  memory.
- pvm: `docs/plans/2026-09-05-parser-prior-art.md` §A1.3,
  `docs/plans/2026-09-04-perl-parser-decision.md`,
  `docs/plans/2026-09-21-cst-is-the-source-of-truth.md`,
  `docs/specs/perl-parser/03-lexer-feedback.md`, `07-conformance.md`,
  `internal/parseoracle/README.md`.
