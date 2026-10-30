<!-- ABOUTME: Design for the second attempt at the chart-parser spike: perl's own grammar over pvm's own tokens. -->
<!-- ABOUTME: States what carries over from chart-waypoint, the token interface, the waypoints, the gates and the kill criteria. -->

# A chart parser for Perl, second attempt: perl's grammar over pvm's tokens

**Status:** design for a spike, written 2026-10-30 from perigrin's decisions
recorded in the `chart-waypoint` milestone's `## Records`. It is the successor
to `docs/plans/2026-10-29-chart-parser-design.md` (the first attempt), not an
amendment of it. If the spike's result is accepted, this becomes a chapter of
`docs/specs/perl-parser/`; until then nothing here describes the parser pvm
ships.

## 1. Why a second attempt

The first attempt (`chart-waypoint`) was stopped by its §9 speed kill:
lexing and recognition ran 3.9–5.0× recursive descent over the corpus,
against a 3× line (`docs/postmortems/chart-waypoint.md`). The postmortem's
finding is that the grammar was **translated faithfully rather than fitted**.
Guacamole is a scanless Marpa grammar; pvm's lexer is not. The bridge between
them compiled every G0 lexeme to a regexp and ran every predicate of a token's
kind on every token: 110 ms of 150 on std5k.pl, against the production
lexer's 9 ms. Most of that milestone's amendments trace to the same mismatch.

The parts that were not Guacamole's held up. The milestone review compared the
recognizer against an independent textbook Earley recognizer over 770 corpus
files, 3,000 mutated files and 11 adversarial grammars and found no
disagreement. Recognition alone was within about 1.5–1.9× recursive descent
untuned, and one keying fix (Q1, §5.3) took std5k.pl from 34.9 ms to 18.3 ms
warm.

So the second attempt keeps the machinery and replaces the grammar and the
token interface with ones fitted to pvm's components: **perl's own grammar,
`perly.y`, stated over `internal/lexer`'s own tokens** — the same ground truth
the recursive-descent parser was built from.

## 2. The question

> Is an up-front Earley chart parser over pvm's lexer, using a grammar ported
> from `perly.y`, within 3× of recursive descent while accepting at least
> what recursive descent accepts?

**Up-front only.** The lexer runs ahead of the chart with its own PL_expect
machine, exactly as for recursive descent, so the comparison isolates the
parsing algorithm. Interleaved lexing — the lexer asking `Expected(pos)` —
is part of what a chart offers and is **not** tested here. A kill in this
spike kills *up-front mode*; the chart approach stays open until interleaved
mode has been tried (perigrin, 2026-10-30).

**Scope is all of Perl.** The target is perl's own grammar, at parity with
recursive descent or a superset of it. There is no subset of our own
definition; the corpus labels are perl's verdicts.

### Out of scope

- Interleaved lexing (§13).
- Changes to `internal/lexer`, `internal/parse` or anything that consumes
  them (§11).
- Incremental reparse and convergence. They follow only a positive verdict.

## 3. Design rules

Rules 1–6 of the first attempt's §3 carry over unchanged: the grammar is
static and context dynamic; context enters only before the chart and after it,
never inside Earley items; only perl's own rules may prune; ambiguity may
stay; outside the accepted language, refuse, never misparse; weights may
order readings, never decide them. This attempt adds:

7. **Fitted, not faithful.** The grammar is stated in the vocabulary of the
   components it runs on. A terminal is something `internal/lexer` produces.
   No construct of the source grammar is carried over because it was there.
8. **No run-time translation.** Everything the grammar implies about tokens
   is compiled ahead of time. Per token, chart/lex does at most one hash
   lookup or one array index; no regexp, no allocation (§7).
9. **One source of truth per fact.** The grammar file is the only place a
   rule, a precedence level or a token class is stated. The generator derives
   everything else and refuses to emit tables when the grammar is
   inconsistent (§5.2).
10. **Speed is gated from the first recognizer issue**, by benchmarks and by
    `lex-overhead`, not left to a single checkpoint.
11. **Algorithmic components are tested against an outside reference**, not
    only against themselves under different options.
12. **A token is a superposition of what it could be; grammar and scope
    collapse it.** What a token *could be* is fixed by perl (the static table)
    or by a library's declarations (the declaration-time extension, §7.1).
    The **grammar** collapses it during recognition: a reading no item expects
    dies there. **Scope** — which declarations are in force at this position,
    whether `foo` is a declared sub, which features are on — collapses it
    after the chart in up-front mode (resolve, §9.4), or at lexing in
    interleaved mode (ruby slippers, §13). Scope is held by the parser beside
    the chart, never in Earley items, so interning is untouched. This refines
    rule 2: "before the chart" is only ever *what a token could be*; "after"
    decides *what it is here*.

## 4. Sources

| Source | Revision | Role |
|---|---|---|
| `perly.y` | perl blead at 94e50866 (2026-07-16), `PERL_VERSION 45` (`patchlevel.h:40-41`), `/opt/corpora/perl5` | Ground truth for every rule, precedence level and token class. About 302 alternatives over 67 nonterminals; 33 precedence lines (`perly.y:150-182`); 39 `%prec` overrides |
| `toke.c` (same tree) | as above | How each `perly.y` token type is decided, for the token table |
| `docs/specs/perl-parser/` chapters 2–5 | on `pu` | The recursive-descent parser's derivation from the same sources: lexical structure, lexer feedback, the 32 precedence levels, statements |
| `internal/lexer` | on `pu` | The token kinds the grammar is stated over (20 kinds, `lexer.go`) and its keyword tables (`keyword.go`, `perlkeywords.go`) |
| Aycock & Horspool, "Practical Earley Parsing" (2002) | | The split LR(0) automaton with nullable closure (carried over) |
| Aycock & Horspool, "Schrödinger's Token" (2001) | | Tokens with several types at once, resolved by the parser (§7.3) |
| YAEP | 7aba93e | Set interning, the transition cache, integer terminal codes |
| Leo (1991) | | Right recursion in linear items (carried over) |
| Kegler, Ruby Slippers (Marpa) | | The lexer asking the parser what is expected, for interleaved mode (§13) |

`perly.y` is read, cited and hand-ported. It is not vendored and not read by
any tool.

## 5. Components

### 5.1 Carried over from chart-waypoint

On `pu` at 4fffe4d6, verified by the milestone review:

- **`grammar`'s generator:** the LR(0) automaton with nullable closure,
  transition and reduce lists, expected-terminal sets, Leo eligibility per
  (state, symbol) with an empty β, byte-identical regeneration
  (`tables-drift`).
- **`recognize`:** interned sets over (state, distance) items, the transition
  cache with its dependency rule, lookahead pruning, `Expected(pos)`, Leo
  memos in a side table outside set identity, and the `Options` switches that
  the differential tests are built on.
- **`cmd/ratchet`:** the regression-only comparison against a committed
  baseline, and its `na` rule.
- **Fixtures:** `std5k.pl` and its generator; the Guacamole suite's extracted
  snippets (`parses`/`parsent`), relabelled by perl's verdict (§9.1).

### 5.2 New

- **`grammar/perl.go`:** the grammar as Go data, one rule per line, each with
  a `Cite("perly.y:N")`. It holds rules, precedence declarations and token
  classes:

  ```go
  var Perl = Grammar{
      Start: "grammar",
      Rules: []Rule{
          R("stmtseq", Seq()).Cite("perly.y:825"),
          R("stmtseq", Seq("stmtseq", "fullstmt")).Cite("perly.y:826"),
          R("term", Seq("term", "ADDOP", "term")).Cite("perly.y:1430"),
          // …
      },
      Prec: []Level{Left("OROR", "DORDOR").Cite("perly.y:162"), /* … */},
      Tokens: []Token{
          Lit("LSTOP", lexer.Word, "print", "push" /* … */),
          Kind("THING", lexer.Number),
          // …
      },
  }
  ```

  It is reviewed like code. Bison actions and mid-rule actions are not
  ported. Where `perly.y` relies on `toke.c` having already decided
  something, the port says so in a comment beside the rule. Two token
  families are ported with a different mechanism:

  - **`GRAM*` start forms** (`perly.y:187-268`). `toke.c` never lexes these;
    the caller pushes one to choose what `yyparse` parses: a program
    (`perl.c:2701`, string `eval` at `pp_ctl.c:4337`), or a block, statement,
    statement sequence, expression or signature for the `parse_*` perlapi
    entry points that XS keyword plugins use (`toke.c:14216-14642`). In the
    chart they are **start symbols**: all seven are ported as start rules,
    and `Recognize` takes the start symbol as a parameter. `fullstmt` is
    also recursive descent's `Repair` unit, which incremental work needs
    later.
  - **`PLUGIN_*` operators** (`perly.y:153-178`, rules at `1256-1470`):
    infix operators registered through XS::Parse::Infix, tokenised by
    precedence class (`tokentype_for_plugop`, `toke.c:9264`). Recursive
    descent supports them through `.tpm` declarations (`sub ⊕ :infix(ADD)`,
    `internal/parse/operators.go`), and `internal/lexer` emits such a symbol
    as an `Operator` wherever it stands (`scan.go:1227`). All eleven classes
    and their rules are ported; §7.1 says how a symbol reaches them.

- **The precedence rewrite (generator):** `%left`/`%right`/`%nonassoc` and
  `%prec` become one nonterminal per level, giving an unambiguous expression
  grammar. A test cross-checks the derived levels against spec chapter 4's
  table.

- **The token table (generator, §7):** (lexer kind, text) and lexer kind to an
  interned terminal-set id.

- **Generator refusals.** The generator emits no tables when:
  - a rule has no citation, or cites a line outside `perly.y`'s range at the
    pinned revision;
  - a token class names a (kind, text) the lexer cannot produce, checked
    against `internal/lexer`'s own keyword and operator tables;
  - a nonterminal is unreachable or derives nothing;
  - a precedence token is used by no rule;
  - a sequence rule's right-hand side is not exactly one symbol (Q9).

- **A thin `chart/lex`:** the production lexer up front, trivia attached to
  the next significant token (lossless), then one table lookup per token.

- **A reference recognizer (test-only):** a plain textbook Earley recognizer,
  about 80 lines, used by the fuzz target, the toy-grammar tests and a corpus
  sample (rule 11).

### 5.3 Preparation of the carried-over code

From the milestone review (`## Records`, 2026-10-30), done before the new
grammar lands, each with a test that fails first:

- **Decouple.** `recognize.Guacamole` and `grammar.readGuacamoleCut` load SLIF
  and the cut. The recognizer and its tests must depend only on
  `*grammar.Tables`, so that removing the discarded code is a deletion.
- **Move the probe tooling.** About 280 lines of `grammar/expect.go` (probes,
  `FindConflicts`, the sampler) and the `Probes`/`Conflicts` fields of
  `Tables` belong to the discarded side. Only the expected-set computation
  carries over.
- **Q1:** token and lookahead classes are integer set ids computed once per
  token, never `fmt.Sprint` keys per node. This merges with §7's set ids.
- **Q2:** `Expected(pos)` with lookahead on is memoized per interned set and
  the origin sets it read, as the transition cache is, not rebuilt per call.
- **Q3:** FIRST and FOLLOW are computed by the generator and stored on
  `Tables`, not in every `recognize.New`.
- **Q4:** `leo-disagree` compares acceptance and `Expected`, not the count
  of distinct sets, which Leo may legitimately change.
- **Q7, Q8, Q10, Q11:** `Recognize` validates its lattice; `Recognizer` is
  documented as not safe for concurrent use; `WriteTables` deletes only the
  table files it owns; `parses-refused` excludes files the lexer could not
  map.
- **Cheap deletions:** the unread `Transition.Items`, the `postdot`/`after`
  duplicate, the `chart_noleo` build tag (`Options.NoLeo` covers it), and the
  ratchet rules for keys that are never printed.

### 5.4 Discarded

The SLIF reader, the token cut and its rewrite list, chart/lex's predicate
mapping, the observed-conflict probes, the CORE.tpm cross-check, Guacamole's
grammar file and its generated tables. They are removed in one dedicated
commit after §5.3, so the removal is a single visible deletion in history.

## 6. Data flow

```
source ──internal/lexer (up front, own PL_expect)──▶ tokens + trivia
       ──chart/lex: one lookup per token──▶ terminal-set ids
       ──recognize (interned sets, cache, Leo, lookahead = next id's set)──▶ chart
       ──Accepted / refusal at the first empty set, with Expected there
```

Waypoint 2 adds: chart → forest (built after recognition) → resolve (facts
outside items) → the oracle's per-site grading.

## 7. The token interface

### 7.1 The table

The generator builds, from `perl.go`'s token classes:

- **an exact map** from (lexer kind, text) to a terminal-set id, for every
  literal a class names (`Word 'print'`, `Operator '+='`);
- **a per-kind array**, `[numKinds]`, from lexer kind to the terminal-set id
  of the whole-kind classes (`Kind("THING", lexer.Number)`);
- **a per-kind flag** saying whether any literal names that kind.

Unions are precomputed. An exact entry already includes every whole-kind
class that also admits its text, so a lookup never unions at run time.

**The declaration-time extension.** When `.tpm` declarations load, every
library operator they declare adds an exact entry from (`Operator`, its
symbol) to the `PLUGIN_*_OP` terminal of its class, unioned with whatever
the static table already gives that text. If two libraries declare the same
symbol at different classes, the entry holds both terminals. The extension
depends on which declarations are *available*, never on which are *in scope*
in a file, so it is built once per declaration load, not per file, and the
tables stay free of per-file state.

### 7.2 The lookup

Per significant token: if the kind's flag is set, look the text up in the
exact map, using Go's no-copy `map[string(bytes)]` form; on a miss, or if the
flag is clear, index the per-kind array. The result is a terminal-set id. The
recognizer reads a precomputed `[]termID` and bitset per set id; it never
sees names. An empty set refuses the file at that byte offset
(`lex-unmapped`).

### 7.3 Several terminals per token

Where `toke.c` decides a token's type from facts pvm does not have while
lexing — whether a bareword is a declared sub, what a prototype makes it,
block or anonymous hash at `{`, method call or indirect object, whether a
library operator's module is in scope — the table maps the token to every
terminal it could be: the token is in superposition ("Schrödinger's Token",
rule 12). The grammar collapses it during recognition, since only items
expecting one of its terminals advance. Whatever scope must still decide is
carried as separate readings. Waypoint 1 grades acceptance only, so those
readings are allowed there, and a scope-dependent acceptance appears as
`chart-accepted-perl-refused`. In waypoint 2, resolve collapses them using
the scope record (§9.4), and the oracle grades the result.

### 7.4 The overhead gate

`lex-overhead` is chart/lex's time divided by `lexer.Tokenize`'s, on
std5k.pl and on the corpus aggregate, under §9.3's protocol. Waypoint 1
requires it to be at most 1.3.

## 8. The precedence rewrite

The generator reads `perl.go`'s precedence levels in `perly.y`'s order
(`perly.y:150-182`, lowest first) and each rule's `%prec` override, and
expands every ambiguous `term` rule into level nonterminals:

- `%left` at level *k*: `E_k ::= E_k op E_{k+1} | E_{k+1}`;
- `%right`: `E_k ::= E_{k+1} op E_k | E_{k+1}`;
- `%nonassoc`: `E_k ::= E_{k+1} op E_{k+1} | E_{k+1}`, so the operator
  cannot chain;
- prefix and postfix operators and `%prec` rules are placed at their
  declared level.

Tests on toy grammars cover each associativity, `%nonassoc` refusing a chain,
and `%prec` overriding a token's own level. A test derives the level table
from `perl.go` and compares it with spec chapter 4's 32 levels and
`internal/parse`'s binding powers; any difference is listed for review, not
silently accepted.

## 9. Waypoints and the ratchet

### 9.1 Corpora

| Corpus | Labels | Used from |
|---|---|---|
| Guacamole snippets (`parses` 746, `parsent` 124 scored) | perl's verdict under the declared preamble | waypoint 1 |
| `std5k.pl`, `edit5k.pl` | both compile under perl | waypoint 1 |
| T1 (`PerlOnJava/unit`, 986 files, top level) | recursive descent's verdict and perl's | waypoint 1 |
| perl.git `t/` | perl's verdict; the oracle's marker sites | waypoint 2 |

"Recursive descent accepts" means `parse.Parse` returns no syntax error.

### 9.2 Waypoint 1: recognizes Perl

Exit, all at once:

- `perl.go` complete, with all seven start forms and all eleven `PLUGIN_*`
  classes; the generator's refusals pass; `tables-drift 0`. The corpora
  grade the program form; a test recognizes a sample of each other start
  form (a block, a statement, an expression, a signature).
- **A library operator:** a test with a `.tpm`-declared infix operator in
  scope is accepted, as recursive descent accepts it. With the library out
  of scope, the chart's acceptance is recorded as
  `chart-accepted-perl-refused`, to be collapsed by waypoint 2's resolve.
- `lex-unmapped`: every unmapped file is on a known list naming an open
  `internal/lexer` issue, and no unlisted file is unmapped. (A criterion may
  not require another milestone's issue to close.)
- **Superset:** `rd-accepted-chart-refused 0` over the waypoint-1 corpora. A
  test prints the list on failure.
- `chart-accepted-perl-refused` reported, not gated. Perl refuses some
  files for reasons that are not syntax (strict, arity), which only
  waypoint 2's oracle can separate.
- `lex-overhead` at most 1.3 (gate).
- `leo-disagree 0` under §5.3's Q4 definition; `set-reuse` reported.
- The speed measurement (§9.3), taken as soon as std5k.pl and the corpus
  are recognized, before any forest work.

### 9.3 Speed protocol

A **measurement** is 2 warm-up runs then 9 timed runs, with a GC before each,
reporting the median and the min/max. Each ratio is the chart's median over
recursive descent's (`parse.Parse`) median, in-process, on the same files.

| Key | Chart side | Use |
|---|---|---|
| `speed-ratio-std5k-longlived`, `speed-ratio-aggregate-longlived` | one recognizer for the whole measurement, reused across files | **the kill reads these** |
| `speed-ratio-std5k-cold`, `speed-ratio-aggregate-cold` | a fresh recognizer per file | reported |

Until the forest exists the chart side is lexing plus recognition, and the
report says so. Go benchmarks (`BenchmarkLexStd5k`,
`BenchmarkRecognizeStd5k`, `BenchmarkExpectedStd5k`) exist from the first
recognizer issue and run in `make test`'s package tests.

### 9.4 Waypoint 2: parses Perl as recursive descent does

The forest built after recognition; resolve; the readings count with collapse
above 2^10; the oracle's per-site grading with **WRONG 0** and exact sites at
or above recursive descent's; parity on perl.git `t/`. Its detailed issues
are refined only after waypoint 1's verdict.

**The scope record** is waypoint 2's interface, and the hook for interleaved
mode. It holds the facts in force at a position: declarations and imports
(including `.tpm` library operators), declared subs and their prototypes,
features and pragmas. The parser maintains it beside the chart as it
completes `use`, `no`, `package` and `sub` statements, never inside Earley
items. Waypoint 2's resolve reads it after the chart to collapse the
readings §7.3 carried. Interleaved mode's lexer will read the same record
while lexing (§13), so it is defined as an interface with those two
consumers, not as resolve's private state.

### 9.5 Ratchet keys

| Key | Rule |
|---|---|
| `tables-drift`, `rd-accepted-chart-refused`, `leo-disagree` | gated at 0 |
| `lex-unmapped` | equals its known list |
| `lex-overhead` | gated at ≤ 1.3 |
| `speed-ratio-*` | reported; the kill reads the long-lived pair |
| `chart-accepted-perl-refused`, `set-reuse`, `rd-parse-ms-*` | reported |

A regression on any gated key fails the ratchet; an `na` key that starts
printing a number fails it, as before.

## 10. Kill criteria

Each is a kill of **up-front mode** (§2).

- **Speed:** a long-lived ratio above **3×** on the aggregate or on std5k.pl,
  as the median of three measurements taken at any checkpoint.
- **Facts in items:** a fix would need facts inside Earley items.
- **Waypoint 2:** a WRONG site that can be fixed only by a refusal dropping
  coverage below recursive descent's; or more than 1% (and at least 5) of a
  corpus's statements collapsing.

These are stops, reported to perigrin, not targets to tune toward. Two
standing process rules from the postmortem apply: a definition amended twice
by the orchestrator goes to perigrin before a third amendment, and no
hypothesised count is written into a criterion as a fixed number.

## 11. Isolation

- Only `internal/chart` changes.
- `internal/lexer`, `internal/parse`, `internal/lower`, `internal/infer`,
  `internal/psc` and `internal/parseoracle` are read, never changed. The
  declaration-time extension (§7.1) reads `.tpm` operator declarations
  through whatever `internal/parse` already exports; if nothing suitable is
  exported, that is reported, not worked around. A lexer
  defect found here is filed against the v0.1 milestone, as
  01a1882d-d8a4 was, and appears on `lex-unmapped`'s known list until fixed.
- Nothing outside `internal/chart` imports it.

## 12. Testing

- **Generator:** toy-grammar tests for the precedence rewrite (each
  associativity, `%nonassoc`, `%prec`), the token table (exact, whole-kind,
  precomputed unions, multi-terminal tokens) and each refusal in §5.2. Each
  test is shown to fail before it passes.
- **Grammar:** every rule cited and in range; the level cross-check of §8.
- **Recognizer:** the carried-over tests and fuzz target, now over integer
  set ids, keeping branched copies, Leo on/off and warm/cold comparisons, and
  adding the reference recognizer as an oracle for acceptance and `Expected`.
- **Superset:** the listed `rd-accepted-chart-refused` test.
- **Start forms and library operators:** a sample of each non-program start
  form; a `.tpm`-declared operator in and out of scope (§9.2).
- **Benchmarks:** §9.3's three, from the first recognizer issue.

## 13. Future work, not in this spike

- **Interleaved lexing, through ruby slippers.** Before lexing at a
  position, the lexer asks the parser what is expected there
  (`Expected(pos)`) and what is in scope there (the scope record, §9.4). It
  intersects the two with what the token could be (§7.1), so the token
  arrives already collapsed, with fewer readings and nothing to prune
  afterwards. As in Marpa's ruby slippers, the lexer may also produce the
  token the parser expects: perl's `/` (divide or regex) and `{` (block or
  hash) decided by the grammar rather than by the lexer's own heuristics.
  This runs at lookahead 0, with Leo memos interned (the `shortcut:` in
  `recognize/leo.go`). It is required before the chart approach as a whole
  can be judged (§2).

  On complete files, ruby slippers mostly re-decides what the lexer's own
  PL_expect machine already decides. **Its larger payoff is M2, where the
  file is being edited:**
  - **Recovery by insertion.** Mid-edit code is usually incomplete. Where the
    parser expects a token the source lacks — a `;`, a `}`, a closing
    delimiter — the lexer can supply it as a virtual token, so the rest of
    the file still parses. This complements recovery by skipping through an
    `error` terminal.
  - **Re-lexing from the chart.** After an edit, the lexer restarts at the
    edit point from `Expected(pos)` and the scope record there, not from a
    replay of its own state from the top of the file, and stops when the
    sets converge with the old chart.

  That is why `Expected(pos)` is memoized (§5.3, Q2) and the scope record is
  an interface (§9.4): both are on M2's latency path.
- Incremental reparse and convergence, after parity.

## 14. Process and budget

The pipeline is the same: this spec, `crochet:assess` to a fixed point,
refinement, chain-review, execute, review, postmortem. House rules: branches
fast-forward on `pu`; an `Implements:` trailer on every commit; a gate review
before every merge; `make test` at 100%; workers follow ponytail.

Proposed budgets, for perigrin to adjust: **waypoint 1, 15 days**, of which
the port of `perl.go` is the largest part; **waypoint 2, 20 days**, refined
after waypoint 1's verdict. Reaching twice a waypoint's budget is a stop
reported to perigrin.

## 15. References

- `docs/plans/2026-10-29-chart-parser-design.md`: the first attempt.
- `docs/postmortems/chart-waypoint.md`: why it stopped.
- The `chart-waypoint` milestone's `## Records`: the rulings and the review
  findings this document cites (§5.3).
- Aycock, J. and Horspool, R. N. (2002), "Practical Earley Parsing", *The
  Computer Journal* 45(6).
- Aycock, J. and Horspool, R. N. (2001), "Schrödinger's Token", *Software:
  Practice and Experience* 31(8).
- Kegler, J., "Ruby Slippers parsing" (Marpa; Ocean of Awareness, 2011) and
  the Marpa::R2 documentation of `terminals_expected` and alternative
  tokens.
- Leo, J. (1991), "A general context-free parsing algorithm running in
  linear time on every LR(k) grammar without using lookahead", *TCS* 82.
- YAEP at 7aba93e: `src/yaep.c`.
