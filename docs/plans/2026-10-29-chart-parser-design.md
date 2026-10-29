<!-- ABOUTME: Design for a research spike: an Earley chart parser for Perl, built beside the recursive-descent parser. -->
<!-- ABOUTME: States the question, the architecture, the ratchet with its gates, and the kill criteria; the verdict is perigrin's. -->

# A chart parser for Perl: design for a research spike

**Status:** design for a spike, written 2026-10-29. Not a decision. If the
spike is accepted, this becomes a chapter of `docs/specs/perl-parser/`; until
then nothing here describes the parser pvm ships.

## 1. Why

`docs/plans/2026-09-05-parser-prior-art.md` §A1.3 chose hand-written recursive
descent because it was the only measured architecture in which *"the lexer can
ask the parser what it is looking at."* pvm's parser as shipped keeps that
choice's results (T2 56/56, perl.git `t/` 619/619, M3-directory re-emission
347/347, WRONG = 0), but not its mechanism: its lexer runs up front with its
own `expect` approximation, and hedges where it cannot know
(`internal/lexer/expect.go`, the `Word` case). A chart lexer that asks the
chart what it expects would give pvm that property for the first time.

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
> on? Does it have a better chance than the current parser at M4–M6
> (prototypes and ambiguity, real-world code, conformance) and at the parse
> half of M2 (incremental, correct under edit), at a speed worth pursuing?

The spike is one ratchet towards parity (perigrin's ruling, 2026-10-29), and
every step produces evidence:

1. **Speed,** end to end against recursive descent, at every step (§8).
2. **Cost of change,** for every step, measured on both parsers. The M4–M6
   cases are early steps, so their costs are the M4–M6 evidence.
3. **Defect yield:** the open "perl disagrees" issues each step settles, and
   every disagreement the arbiter (§7.4) decides in the chart's favour.
4. **Incremental convergence,** at every step: how soon Earley sets match
   again after an edit. Then, once parity is reached, M2's equivalence
   contract.

The chart cannot feed analysis (§11), and M2's ruled target is a re-parsed tree
*and* its published analysis in under 10 ms p99, so no result here establishes
M2's latency gate.

### Out of scope

Replacing the production parser; choosing the downstream IR; feeding `lower`,
`infer` or `psc`; a B::SoN semantic oracle (§11).

### Output

A recommendation at the checkpoint after the M4–M6 steps (§8), with two
verdicts that may differ: the M4–M6 route and the incremental route. A
positive M4–M6 verdict continues the ratchet to parity. The incremental route
is pursued only after parity, so a positive incremental verdict without a
positive M4–M6 verdict ends the spike at the recommendation. The code is kept,
and stays experimental until the spike is accepted.

## 3. Design rules

1. **The grammar is static; the context is dynamic.** Imports and pragmas
   change how tokens are classified and which reading applies, not the
   grammar.
2. **Context enters at two places only:** before the chart (token
   classification, in `chart/lex`) and after it (pruning, in
   `chart/resolve`). Never inside the chart: facts in Earley items would make
   otherwise-identical sets differ, which kills set interning.
3. **Only perl's own rules may prune.** A reading may be removed by a fact
   perl had at compile time (a prototype, a declared sub, a feature, an
   import, a library declaration the oracle has vetted), or by a deterministic
   rule perl's lexer applies (`intuit_more` for `{`). Some compile-time facts
   pvm can obtain only by running code (a prototype installed by a string
   `eval` in `BEGIN`; spec 03 §3.0), and those readings stay ambiguous.
   Inferred runtime types never prune; where they contradict perl's reading,
   that is a diagnostic.
4. **Ambiguity may stay.** A language server needs the facts true in every
   reading. A reading is WRONG if it asserts what perl does not; perl's
   reading among several live ones is *covered but undecided*. A statement
   whose readings exceed a fixed cap collapses to an Error node.
5. **Outside the accepted language, refuse; never misparse.** Pest scored
   99.996% and still lost (prior-art §A1.3) because it misparsed silently.
6. **Weights may order readings, never decide them.** A lexeme priority is a
   deterministic tiebreak rule, like perl's own, not a weight.

## 4. Sources

pvm's chart parser is an independent design in Go, borrowing from three
sources where the constraints match.

**Guacamole v0.008** (`xsawyerx/guacamole` at 2838253) is the starting
grammar, and its definition of Standard Perl is the ratchet's first waypoint
(perigrin's rulings, 2026-10-29). Its G1 rules supply the `Statement`/
`BlockStatement` split, the four-way precedence clones (`ExprValueU`/`0`/`L`/
`R`), the `NonBrace*` hierarchy for block versus hash, and per-keyword
builtin rules (about 200, e.g. `OpKeywordTimeExpr ::= OpKeywordTime LParen
RParen`). Rules at or below pvm's token level are rewritten (§5).

**Chalk** (`perigrin/chalk`) contributes techniques: integer core-item IDs,
nullable sets and Aycock–Horspool dot advancement (`LR0DFA.pm`), terminal
clustering, the relative-distance chart, its seven ambiguity classes
(`docs/architecture/ambiguity-classes.md`), and Leo with an equivalence test.
It is scanless, resolves ambiguity during the parse through semirings, and has
no forest; pvm does none of those.

**YAEP** (`vnmakarov/yaep` at 7aba93e) contributes set interning: situations,
kernel-deduplicated cores, deduplicated distance vectors, whole sets interned
as (core, distances), per-(core, symbol) transition and reduce lists, and the
(set, token, lookahead) transition cache. Its forest is post-hoc: items hold no
links, and origins are found by walking the shared sets (`make_parse`).

## 5. Components

A new tree, `internal/chart/`, depending on `internal/lexer`. `chart/compare`
and the grammar generator also import `internal/parse`.

### `chart/grammar`

Reads the **subset of Marpa's SLIF notation that Guacamole uses** (perigrin's
ruling, 2026-10-29): G1 rules (`::=`, `|`), sequence rules (`X ::= Y*`),
`:lexeme ~ X priority => n`, `:discard`, and `lexeme default = latm => 1`.
`action =>` and `name =>` adverbs are ignored.

**The token cut** (perigrin's ruling, 2026-10-29). G1 builds some single pvm
tokens from several lexemes. Every rule at or below pvm's token level is
rewritten in terms of **token predicates** (lexer kind plus text): a rule
wholly below it collapses into a predicate, and a rule that straddles it is
split (`VarScalar ::= SigilScalar VarIdentExpr ElemSeq0` becomes `VarScalar
::= ScalarVar ElemSeq0`). Every rewrite is listed beside the grammar and
reviewed. The cut also carries G0's lexeme priorities as static rules: keyword
text satisfies `SubName` (method names, definitions) but not
`SubNameNonQLike` (calls), so `print(1)` has one derivation. `Pod`,
`DataSection` and `FormatBody` tokens are trivia.

**Builtin shapes** come from G1's per-keyword rules. CORE.tpm is a
cross-check: the generator lists every builtin whose CORE.tpm shape differs
from G1's, for review. Reading CORE.tpm needs `internal/parse`'s unexported
shape table exported, a third exception to the isolation in §10.

A generator, run through a Makefile target, builds the tables ahead of time:
core items, the LR(0) automaton with nullable closure, per-state transition
and reduce lists, per-state expected-terminal sets and Leo eligibility, and
expectation classes.

**Expectation classes** project a position's expected-terminal set, the union
over every state in the Earley set there, onto the token pairs that conflict
lexically in the grammar in force. In Standard Perl those are `/` as division
or regex and `<` as less-than or readline, plus `%`, `&`, `*`, `-X` and `x`.
`{`, `<<` and `?` join the list only as the ratchet adds what makes them
conflict (block/hash in perl's sense, heredocs, `?PATTERN?`).

### `chart/lex`

**Two modes.** The ratchet starts **up front**: the production lexer, with its
own `expect` machine, produces the token stream before recognition, as
`internal/parse` uses it today. It cannot express a lexical ambiguity; it
resolves each one by its own default. **Interleaved** lexing arrives at the
first ratchet step whose feature needs a token branch: before each token the
lexer reads the expectation class from the current Earley set and scans only
the kinds it allows (Marpa's `latm`). That step includes refactoring
`internal/lexer` so a scanner takes its expectation and context as parameters
instead of advancing its own state (perigrin's ruling, 2026-10-29). The
refactor is mainline work through `TestLexerRatchet`, and is kept after a
stopped spike only on its own merits.

**One coupled choice: lexing mode, lookahead, list shape and Leo.** An
interleaved lexer cannot supply lookahead, because lexing token *k+1* needs
the set at *k+1*. So interleaved means lookahead 0, where G1's right-recursive
`StatementSeq` is quadratic without Leo (a toy YAEP probe: 0.007 s at
lookahead 1 against 87.6 s at lookahead 0, over 100k tokens). Interleaved
therefore needs working Leo, and has lower set reuse by construction. The
every-step speed measure (§8) times whichever mode is in force.

**Ties.** Where several expected kinds match, Guacamole's lexeme priorities and
perl's deterministic lexer rules decide (rule 3). Where pvm lacks a fact perl
had, the interleaved lexer emits a **token branch**.

**Input model.** Positions are the nodes of a token lattice: byte offsets of
token boundaries, each paired with a lexer state. A branch forks the lattice
into paths of possibly different token counts. Two paths **rejoin** only where
both byte offset and lexer state are equal (the rule `TokenCache.resyncsAt`
and `State.Equal` already use); the state the chart compares excludes the
production `expect` field, which interleaved mode does not use. A branch not
rejoined within a bounded number of tokens is refused. Relative distances
count lattice nodes; sets intern outside branches only.

### `chart/recognize`

A YAEP-style Earley recognizer over the tables and the lattice: interned sets,
relative distances, the transition cache (lookahead 0 and no cache inside a
branch), Leo memos stored by reference with each resolved top in a
per-position side table outside the set's identity, and error recovery
through an `error` terminal the generator adds to statement rules, skipping
the fewest tokens. It exposes `Expected(pos)`.

### `chart/forest`

A Scott-style SPPF built **after** recognition, by walking the interned sets
as YAEP's `make_parse` does. Items never hold forest pointers: Scott's
in-recognition construction would defeat interning. Packed nodes may carry a
condition, and Leo's skipped intermediate nodes are recovered during the walk.

### `chart/resolve`

Prunes packed nodes by rule 3; the rest stays ambiguous. Built as the ratchet
needs it.

### `chart/compare`

Exports each parser's tree as spans and diffs them per file, under a
normalisation fixed at the first step: byte spans, a trivia attachment rule,
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
   │  if every reading dies: error recovery (`error`, fewest skipped)
   ▼
chart/forest ── post-hoc SPPF; packed nodes may carry conditions
   │
   ▼
chart/resolve ── perl's own rules only; the rest stays ambiguous
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
   kinds and never sees surplus (`compareMarkers`), so it is necessary, not
   sufficient.
2. **Yield.** The yield of every packed node is exactly its byte span, and the
   root's yield is the source.
3. **A per-file span diff** against recursive descent, under §5's
   normalisation, wherever both produce one reading.
4. **An arbiter for every disagreement,** on perl's own `B::Deparse`
   (perigrin's ruling, 2026-10-29). It is two-sided. For the statement a
   disagreement falls in, each parser's reading is emitted fully parenthesised,
   and so is the source. Every literal outside a quote-like is first replaced
   by a distinct variable (`1+2*3` becomes `$k1+$k2*$k3`), so constant folding
   cannot hide grouping. All three go through `perl -MO=Deparse,-p`. The parser
   whose emission deparses as the source does is right. If both do, the shape
   difference is legitimate. If neither does, or the renaming changes the
   parse, the disagreement is **undecided**. Parity needs no disagreement
   decided against the chart, and undecided disagreements at no more than 10%
   of the corpus's disagreements. A disagreement decided for the chart is a
   finding against recursive descent, and counts as defect yield.

Canon re-emission stays recursive descent's, so the chart's M6 verdict is on
M6 without re-emission.

## 8. The ratchet

One ratchet from Standard Perl to parity (perigrin's rulings, 2026-10-29).
Each step adds one feature, keeps everything already accepted accepted, and is
measured before the next begins.

**Every step measures,** against recursive descent, on every file both parsers
accept:

- **speed,** end to end: lexing, recognition and forest construction on the
  chart's side, `parse.Parse` on recursive descent's. The aggregate ratio is
  recorded at every step, so the trend is visible, and edit5k.pl joins the set
  as soon as the chart accepts it;
- **cost of change:** lines changed and places touched on each side, where the
  fix landed, and, for features recursive descent already has, its cost from
  the PR that added them. The share of a step spent building infrastructure
  later steps reuse is reported separately;
- **defect yield:** open issues settled, and arbiter decisions for the chart;
- **convergence,** over a fixed set of 200 single-statement edits per corpus,
  chosen before the first step from the corpus's statements: the number of
  tokens past the edit until the new sets match the old.

**Budgets.** Before a step starts, refinement records its calendar estimate in
the milestone. A step open beyond twice its estimate forces a decision by
perigrin, continue or stop, recorded with its reason; continuing records a new
estimate, and the rule applies again.

### The waypoint: Standard Perl

The first step builds the recognizer, the forest, error recovery,
`chart/compare`, the arbiter and the oracle adapter, with the production lexer
up front. Its corpus is Guacamole's suite, extracted by running its tests with
a recording `Guacamole::Test` (791 `parses` and 126 `parsent` calls), checked
in and reviewed, under a declared preamble (`use feature ':5.36'`).

- Cases needing a missing module run with stub modules on `@INC`, each
  declaring a `$VERSION` that satisfies every request.
- `parsent` means "not Standard Perl", so its labels are correct by
  definition. A `parses` case perl still rejects is a Guacamole erratum,
  excluded from scoring, as spec 07 §7.5.2 excludes any input perl fails to
  compile.
- **Exit:** every perl-accepted `parses` case accepted under §7, and every
  `parsent` case refused with an Error node. The expectation classes are then
  compared with spec 03's eleven `PL_expect` states, through an explicit map
  from each state to the conflicts it settles, and the number of states whose
  decisions the classes do not reproduce is recorded. More than 2 forces a
  recorded decision by perigrin.

### The M4–M6 steps, then the checkpoint

The next steps are the M4–M6 cases, fixed before the waypoint exits and drawn
from open issues: `first { ... } @list`; a named unary versus a list operator
with an unknown prototype; `sort $coderef @list`; `s///e`; `(?{ ... })`;
unusual delimiters; and failing constructs from `t/base/lex.t` and
`t/comp/parser.t`. Each builds the Standard-Perl exclusions it needs, such as
prototype-driven parses for `first { ... } @list`. Recursive-descent work on
an open issue is ordinary mainline work through mainline gates. Which side
goes first alternates.

**Checkpoint.** A step is *clearly cheaper* on the chart if it needs at most
half the lines and half the places touched; the comparison is reported with
and without the reusable-infrastructure share. The M4–M6 verdict is positive
if at least half these steps are clearly cheaper and none needed a rule no
Perl construct justifies. The incremental verdict is "keep the snapshot
design" if fewer than 80% of the convergence edits converge within the edited
statement plus 8 tokens. The recommendation is made here.

### On to parity (only if the M4–M6 verdict is positive)

The rest of Guacamole's exclusion list, one feature per step, skipping what
the M4–M6 steps built: auto-quoting; `print {$fh}` forms and bareword
filehandles; unusual quote delimiters; heredocs; paren-less calls;
prototype-driven parses; indirect objects; `given`/`when`. Then require/use
loading, which perl.git `t/`'s 619/619 depends on. Each step records `parsent`
cases flipped to correct parses. The milestone-corpus head-to-heads (M4
`t/comp/proto.t`, `t/op/signatures.t`, T1; M5 the T4 regex files and CPAN; M6
`t/op`) run as those files become parseable.
**Exit:** §7 parity with the floor: T2 56/56, perl.git `t/` 619/619, T1 and the
corpus and site ratchets at or above recursive descent's.

### Incremental (only after parity, and only if that verdict is positive)

The full convergence measures over the parity corpora: (a) tokens past the
edit until the sets match again, (b) the fraction of sets that match after
that, and (c) the transition-cache hit rate. Then M2's contract: a repair
returns a tree equal to a fresh parse or refuses, under FuzzIncremental and
FuzzIncrementalSequence retargeted at the chart, with repair latency reported
against the parse share of M2's 10 ms.

## 9. Risks and kill criteria

Fixed before work starts. A fired kill ends the spike unless perigrin records
an override and its reason.

| Risk | Signal | Kill criterion |
|---|---|---|
| Speed | Every step's ratio | The aggregate ratio above 3× at any step. Recursive descent's cold parse on edit5k.pl is 23.89 ms p50, 33.42 ms p99; a YAEP-class extrapolation for 5 kLOC is about 13.6 ms |
| Silent misparse | A `parsent` case accepted; WRONG > 0; a disagreement decided against the chart | One that is fixed only by a refusal which drops coverage below recursive descent's on that file |
| Undecided disagreements | The arbiter's undecided count | More than 10% of disagreements undecided at the checkpoint or at parity |
| Forest blow-up | Readings per statement: the product over the packed nodes inside it | More than 1% of a corpus's statements collapsed by rule 4's cap (2^10 readings) |
| Expectation classes | The waypoint's state count | More than 2 of 11 states unreproduced: a recorded decision |
| Context leaking into the chart | Set reuse falling as steps add features | Any fix needing facts inside items: a design review |
| No incremental convergence (Wiren 1989) | Every step's convergence measure | The checkpoint's incremental threshold |
| No advantage | The M4–M6 step costs | The checkpoint's M4–M6 threshold |
| Drift and stall | Budgets (§8) | A step beyond twice its estimate forces a recorded decision |

## 10. Testing

pvm's conventions apply: TDD, the pre-commit hooks, `make test` for code.
`internal/chart` is ordinary mainline code: its tests run in `make test` and
the hooks like any package's. The spike's numeric ratchets (speed ratio, cost
tables, corpus counts) run under `make chart-ratchet`, which mainline does not
gate on, and which runs on any change under `internal/chart` or
`internal/lexer`.

- **Generator:** golden tests on small grammars (nullables, left and right
  recursion); byte-identical tables on regeneration.
- **Recognizer:** positions with equal content share one interned set; Leo on
  and off give identical forests; a test that **fails if Leo does nothing**
  (`S → a S` at lookahead 0 must be linear); `1 2`, `$x $y` and `print(1)`
  have zero, zero and one derivations.
- **Lattice and forest:** a branch rejoins only where offset and state agree,
  tested on a toy grammar until the ratchet adds heredocs; every packed node's
  yield is its span.
- **Lexer refactor:** `TestLexerRatchet` and the existing lexer corpus,
  unchanged; each scanner yields today's tokens given today's context.
- **Compare and arbiter:** the normalisation tested against hand-built trees,
  with one check that must pass for each that must fail: a disagreement where
  the chart is right is decided for it, one where recursive descent is right
  is decided for recursive descent, and identical trees report none.
- **Fuzzing:** recognizer invariants from the first step, forest invariants
  from the step that builds the forest, and the incremental fuzzers after
  parity.
- **Isolation:** `internal/parse`, `lower`, `infer` and `psc` do not change for
  the chart, with three exceptions that are ordinary mainline work: the
  `internal/lexer` refactor, recursive-descent probe work, and exporting
  `internal/parse`'s builtin shape table.

## 11. Future work, not in this spike

- **A B::SoN semantic oracle.** B::SoN (`perigrin/perl5-son`) turns perl's
  compiled op tree into a Sea-of-Nodes graph. A resolver-to-SoN builder would
  let pvm diff a reading's semantics against `perl -MO=SoN,json`.
- **The downstream IR.** Until it is settled the chart cannot feed analysis,
  which is why §2 covers only the parse half of M2.

## 12. References

- Aycock and Horspool, "Practical Earley Parsing" (2002); Aycock's thesis
  (2001).
- Leo (1991). Scott (2008), "SPPF-style parsing from Earley recognisers".
  Wiren (1989), EACL. Wagner and Graham (PLDI 1997).
- YAEP at 7aba93e: `src/yaep.c`, `doc/Internals.txt`.
- Guacamole v0.008 at 2838253: `lib/Guacamole.pm`, `lib/standard.pm`, `t/`.
- Chalk: `ARCHITECTURE.md`, `docs/architecture/earley-parser.md`,
  `docs/architecture/ambiguity-classes.md`.
- perl: `perly.y` writes `parser->expect` in 14 places and tests it in 4;
  toke.c sets it from the symbol table.
- pvm: `docs/plans/2026-09-05-parser-prior-art.md` §A1.3,
  `docs/plans/2026-09-04-perl-parser-decision.md`,
  `docs/plans/2026-09-21-cst-is-the-source-of-truth.md`,
  `docs/specs/perl-parser/03-lexer-feedback.md`, `07-conformance.md`,
  `internal/parseoracle/README.md`.
