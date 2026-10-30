<!-- ABOUTME: Design for the second attempt at the chart-parser spike: perl's own grammar over pvm's own tokens. -->
<!-- ABOUTME: States what carries over from chart-waypoint, the token interface, the waypoints, the gates and the kill criteria. -->

# A chart parser for Perl, second attempt: perl's grammar over pvm's tokens

**Status:** design for a spike, written 2026-10-30 from perigrin's decisions
recorded in the `chart-waypoint` milestone's `## Records` and in this
decision's assessment (`docs/assessments/chart-parser-v2-design.md`). It is
the successor to `docs/plans/2026-10-29-chart-parser-design.md` (the first
attempt), not an amendment of it. If the spike's result is accepted, this
becomes a chapter of `docs/specs/perl-parser/`; until then nothing here
describes the parser pvm ships. Implementation detail the assessment moved
out of this document is listed in the assessment, for refinement.

## 1. Why

**Why a chart at all.** The first design's §1 still holds: a chart keeps every
reading perl's grammar allows, so a language server can report the facts true
in every reading and refuse rather than misparse (prior-art §A1.3); it gives
principled error recovery; and an Earley chart's sets can be compared across an
edit, which is the convergence M2's incremental work needs. Recursive descent
commits to one reading as it goes and gets none of these for free.

**Why a second attempt.** The first attempt (`chart-waypoint`) was stopped by
its §9 speed kill: lexing and recognition ran 3.9–5.0× recursive descent over
the corpus, against a 3× line (`docs/postmortems/chart-waypoint.md`). The
grammar was **translated faithfully rather than fitted**: the bridge from
Guacamole's scanless lexemes to pvm's tokens ran a regexp per terminal per
token, 110 ms of 150 on std5k.pl against the production lexer's 9 ms. The
recognizer held up: an independent textbook recognizer agreed with it on 770
corpus files, 3,000 mutated files and 11 adversarial grammars, and recognition
alone was about 1.5–1.9× recursive descent untuned.

So this attempt keeps the machinery and replaces the grammar and the token
interface with ones fitted to pvm's components: **perl's own grammar,
`perly.y`, restated over `internal/lexer`'s tokens**, the ground truth the
recursive-descent parser was built from.

## 2. The question, and what its answer decides

> Is an up-front Earley chart parser over pvm's lexer, using a grammar ported
> from `perly.y`, within 3× of recursive descent in whole-file throughput, while
> accepting every file both perl and recursive descent accept?

**Up-front only.** The lexer runs ahead of the chart with its own PL_expect
machine, exactly as for recursive descent, so the comparison isolates the
parsing algorithm.

**What a verdict decides.**
- *Positive:* waypoint 2 builds the forest and resolve and re-measures speed
  with them (§9.4). If that also holds, the chart is a candidate engine for
  M2's incremental work, and the interleaved spike (§13) is designed next.
- *Negative (a kill):* up-front mode is closed. It does **not** close
  interleaved mode. Interleaved lexing does more total work than up-front,
  but it can return an answer *sooner* at a position, so it is judged on time
  to an answer, M2's latency, not on whole-file throughput (perigrin,
  2026-10-30). An up-front kill is recorded as evidence about throughput for
  that spike to weigh, not as its verdict.

**Scope is all of Perl** at parity with recursive descent or a superset of it;
the corpus labels are perl's verdicts. Two parity limits are stated, not
hidden: quote-like bodies are single tokens (interpolated code, `s///e`
bodies and `qw` lists are not recognized), as they are for recursive descent
today; and library-declared operators are waypoint 2's (§9.4).

### Out of scope

- Interleaved lexing (§13); incremental reparse and convergence.
- Changes to any package outside `internal/chart` (§11).

## 3. Design rules

Rules 1–6 of the first attempt's §3 carry over: the grammar is static and
context dynamic; context enters only before the chart and after it, never
inside Earley items; only perl's own rules may prune; ambiguity may stay;
outside the accepted language, refuse, never misparse; weights may order
readings, never decide them. This attempt adds:

7. **Fitted, not faithful.** The grammar is stated in the vocabulary of the
   tokens `internal/lexer` actually emits. Where toke.c invents or reshapes
   tokens, the rule is re-fitted to the lexer's stream, and says so (§5.2).
8. **No run-time translation.** Everything the grammar implies about tokens
   is compiled ahead of time; chart/lex does a constant, allocation-free
   lookup per token, with no regexp.
9. **One source of truth per fact, and a check where two remain.** The grammar
   file is the only place a rule, precedence level or token class is stated.
   Where a fact is also held elsewhere in pvm (builtin shapes in CORE.tpm), a
   test lists every difference for review.
10. **Speed is measured from the first recognizer issue** and gated wherever a
    gate can fail (§9.3).
11. **Algorithmic components are tested against an outside reference**, not
    only against themselves under different options.
12. **A token is a superposition; three things collapse it.** What a token
    *could be* is decided by its own text and kind, the flags the lexer has
    already computed, and the next token (§7). The **grammar** collapses it
    during recognition: a reading no item expects dies there. **Scope**
    (declared subs, imports, features) collapses what remains, after the chart
    in waypoint 2. Scope is never held in Earley items.

## 4. Sources

| Source | Revision | Role |
|---|---|---|
| `perly.y` | perl blead 94e50866 (2026-07-16), `PERL_VERSION 45` (`patchlevel.h:40-41`), `/opt/corpora/perl5` | Ground truth: 348 alternatives over 114 nonterminals (bison 3.8.2 report), 32 precedence lines (`perly.y:150-182`), 39 `%prec` uses |
| `toke.c` (same tree) | as above | How each token type is decided, and which tokens it invents |
| `docs/specs/perl-parser/` chapters 2–6 | on `pu` | The recursive-descent parser's derivation from the same sources; chapter 4's levels; chapter 6's lexer checkpoints |
| `internal/lexer` | on `pu` | The tokens and flags the grammar is stated over |
| Aycock & Horspool (2002, 2001) | | The split LR(0) automaton; tokens with several types ("Schrödinger's Token") |
| Afroozeh et al. (2013) | | Operator precedence as a filter that accepts what conflict resolution accepts (§8) |
| YAEP 7aba93e; Leo (1991) | | Interning, the transition cache; linear right recursion |

`perly.y` is read and cited by people; no build step reads it.

## 5. Components

### 5.1 Carried over from chart-waypoint (on `pu` at 4fffe4d6)

- **`grammar`'s generator:** the LR(0) automaton with nullable closure,
  transition and reduce lists, expected-terminal sets, Leo eligibility,
  byte-identical regeneration.
- **`recognize`:** interned sets, the transition cache with its dependency
  rule, lookahead pruning, `Expected(pos)`, Leo memos outside set identity,
  and the `Options` switches the differential tests are built on.
- **`cmd/ratchet`'s comparison,** with two changes this attempt needs: keys
  are reset with a new baseline, and some keys gain absolute gates (§9.5).
- **Fixtures:** `std5k.pl` and its generator; the Guacamole suite's extracted
  snippets, relabelled by perl's verdict (§9.1).

Before the new grammar lands, the carried-over code is separated from the
discarded code so that removing the latter is a single deletion that leaves
the kept packages compiling and tested, and the milestone review's findings
against the kept code are fixed. Both are refinement's to size; the
assessment lists what they involve.

### 5.2 New

- **The grammar, as Go data in `internal/chart/grammar`,** one rule per
  line. Each rule carries a citation: `perly.y:N` for a rule ported as
  written, or a *fitted* citation (`perly.y:N, fitted: <what the lexer emits
  instead>`) for a rule restated over the lexer's stream. Bison actions and
  mid-rule actions are not ported. The families of rules expected to need
  fitting are named so refinement can size them: braces and the `;` toke.c
  inserts before `}`; quote-likes and interpolation (single tokens here);
  formats; labels; attributes; postfix dereference; heredocs (bodies are
  trivia).
- **Start forms.** The `GRAM*` tokens are start selectors, not lexed tokens
  (`perly.y:187-268`). All seven are ported as start rules. Waypoint 1 grades
  the program form; the others are tested only where they lex the same from
  statement start, since the lexer exports only statement-start states, and
  that limit is recorded.
- **`PLUGIN_*` operators.** The eleven classes (`perly.y:153-178`, rules at
  `1256-1470`) are ported as grammar. How a library operator reaches them is
  waypoint 2's (§9.4).
- **The precedence filter (§8).**
- **The token table (§7).**
- **Generator refusals.** No tables are emitted when a rule lacks a citation
  or cites a line outside `perly.y`'s range; a token class names a literal
  that `internal/lexer` does not lex to that kind and text (checked by lexing
  it); a nonterminal is unreachable or derives nothing; a precedence token is
  used by no rule or `%prec`; or a sequence rule's right-hand side is not one
  symbol.
- **The port ledger.** Every `perly.y` alternative appears once in a reviewed
  list as ported, fitted or excluded with a reason; a test holds the grammar's
  citations and the ledger equal (§9.2).
- **A thin chart/lex:** the production lexer up front, trivia attached to the
  next significant token, then one table lookup per token.
- **A reference recognizer (test-only):** a plain Earley recognizer used by
  the fuzz target, toy grammars and a corpus sample.

### 5.3 Discarded

The SLIF reader, the token cut, chart/lex's predicate mapping, the
observed-conflict probes, the Guacamole CORE.tpm cross-check and Guacamole's
grammar and tables, in one dedicated commit. The CORE.tpm cross-check is
replaced, not dropped: the port's builtin classes (`UNIOP`, `LSTOP`, `FUNC0`
and kin) are compared with `parse.CoreShapes()` and every difference listed.

## 6. Data flow

```
source ──internal/lexer (up front)──▶ tokens + trivia + flags
       ──chart/lex: one lookup per token──▶ terminal-set ids
       ──recognize (interned sets, cache, Leo, precedence filter)──▶ chart
       ──accepted, or refused at the first empty set with Expected there
```

Waypoint 2 adds the forest, resolve (scope collapse), and the oracle.

## 7. The token interface

**The key.** A token's terminal set is looked up by its lexer kind, its text,
the flags the lexer has already computed (`OpensBlock`, `TermFollows` and
similar), its sigil or quote-prefix class, and the class of the next
significant token. The lexer has already decided block-or-hash at `{`; the
key keeps that decision instead of re-deciding it. The next token's class is
what toke.c reads to tell `FUNC` from `LSTOP`, `FUNC1` from `UNIOP`, and an
auto-quoted word before `=>`.

**Compiled.** The generator builds the table from the grammar's token classes,
with unions precomputed, so the lookup is constant-time, allocation-free and
regexp-free, and yields an integer terminal-set id. The recognizer works only
in ids. An empty set refuses the file at that byte (`lex-unmapped`).

**Several terminals.** Where the key still cannot decide (a bareword that may
be a declared sub, an indirect object), the token maps to every terminal it
could be. The grammar collapses what it can during recognition; scope
collapses the rest in waypoint 2. Ambiguities only the next character decides
are covered by the key's lookahead class, not left to scope.

**What it costs.** Superposition moves cost from the lexer into the
recognizer, so §9 reports terminals per token and Earley items per token
alongside chart/lex's own overhead.

## 8. Precedence: a conflict filter

`perly.y`'s `%left`/`%right`/`%nonassoc` lines and `%prec` uses do not
stratify the grammar; bison applies them only where a shift/reduce conflict
arises. A lower-precedence prefix operator may therefore be a right operand
(`$a + defined $b`, `my $s = join ",", @a`), and `%prec` resolves conflicts on
rules that are not operators at all (most of the 39 uses). A level-by-level
rewrite would refuse code perl accepts.

So the rules stay as `perly.y` writes them, and the generator derives, from the
declared levels and associativity, the filter that rejects exactly the
derivations bison's resolution rejects (perigrin, 2026-10-30). The filter is
static data applied at completion; it holds no facts and keeps items
context-free. Its contract is bison's accepted language, tested on:

- chains in each associativity, and `%nonassoc` refusing a chain;
- lower-precedence prefix operands on the right (`$a + defined $b`,
  `$a * defined $b + 1`, `$a + not $b, 1`) and list operators after `=`;
- `%prec` on non-operator rules;
- every level declared, cross-checked against spec chapter 4's table and
  `parse.LevelsAccountedFor()`, with differences listed.

## 9. Waypoints and the ratchet

### 9.1 Corpora and the superset population

| Corpus | Labels |
|---|---|
| Guacamole snippets (`parses`, `parsent`) | perl 5.42's verdict under the declared preamble (perl accepts 78 of the 124 `parsent`) |
| `std5k.pl`, `edit5k.pl` | both compile under perl |
| T1 (`PerlOnJava/unit`, 986 files, top level) | perl's verdict and recursive descent's (it refuses 236) |
| perl.git `t/` (waypoint 2) | perl's verdict; the oracle's marker sites |

**Recursive descent accepts** a file when `parse.Parse`'s root reports no
refusal (`Node.Refused()` false). The **superset population** is every file
in the waypoint-1 corpora that **both perl and recursive descent accept**,
less files on `lex-unmapped`'s known list (§9.2). Files recursive descent
accepts but perl refuses are reported, not gated, since accepting them would
break rule 5.

### 9.2 Waypoint 1: recognizes Perl

Exit, all at once:

- **The port is complete:** every `perly.y` alternative is in the port ledger
  as ported, fitted or excluded with a reason, and every grammar citation is in
  the ledger. The generator's refusals pass; `tables-drift 0`.
- **`lex-unmapped`** equals a known list, each entry naming an open lexer
  issue; no unlisted file is unmapped.
- **Superset:** `rd-accepted-chart-refused 0` over the superset population,
  with the list printed on failure.
- **Precedence:** the §8 cases pass.
- **Builtins:** the CORE.tpm cross-check runs and its differences are listed.
- **`leo-disagree 0`**, defined as Leo on and off agreeing on acceptance and
  on `Expected` at every node.
- **Speed:** measured at the checkpoints of §9.3; `lex-overhead` gated.

### 9.3 Speed protocol

**Ratios.** Every ratio is chart over recursive descent, timed **alternately**
in one process (chart, RD, chart, RD…) so machine load affects both sides
alike. A **measurement** is 2 warm-up pairs then 9 timed pairs; the ratio is
the median of the 9 per-pair ratios, with min and max reported.

**Two lifetimes,** each covering every file exactly once per timed pass, so
no pass replays a cache warmed on its own inputs:
- **long-lived:** one recognizer per pass, reused across that pass's files,
  as an editor holds one across documents (this is the first attempt's
  measurement);
- **cold:** a fresh recognizer per file.

**Figures.** `std5k` is std5k.pl alone. The **aggregate** is the summed time
over the superset population, excluding std5k.pl and edit5k.pl, with each
corpus's share also reported so no single file dominates unseen. The kill
reads the long-lived std5k and aggregate ratios.

**Checkpoints.** A speed measurement is taken (1) when std5k.pl and the
superset population are first recognized, (2) at waypoint 1's exit, and (3) at
waypoint 2's exit, with the forest included. Before checkpoint 1, every
recognizer issue reports speed on whatever is recognized, for early warning.

**Gates that can fail.** The ratchet (`make chart-ratchet`) is part of every
recognizer issue's acceptance criteria, so its gated keys block each issue.
`lex-overhead` is chart/lex's time per token over `lexer.Tokenize`'s, measured
alternately at a size where the two runs of `Tokenize` against itself agree
within 10%, and gated at 1.3. Benchmarks exist for lexing, recognition and
`Expected`, and are run by the ratchet, not by `make test`.

### 9.4 Waypoint 2: parses Perl as recursive descent does

The forest; resolve, which collapses by scope; the readings count with
collapse above 2^10; the oracle's per-site grading with **WRONG 0** and exact
sites at or above recursive descent's; parity on perl.git `t/`; and the speed
measurement at checkpoint 3. Library-declared operators come here: how
recursive descent's `.tpm` operator system (its classes, relations and
fixities) reaches the chart's `PLUGIN_*` terminals, and whatever read-only
access to `internal/parse` that needs, is waypoint 2's decision (perigrin,
2026-10-30). Its issues are refined only after waypoint 1's verdict.

### 9.5 Ratchet keys

| Key | Rule |
|---|---|
| `tables-drift`, `rd-accepted-chart-refused`, `leo-disagree` | gated at 0 |
| `lex-unmapped` | equals its known list |
| `lex-overhead` | gated at ≤ 1.3 |
| `speed-ratio-{std5k,aggregate}-{longlived,cold}` | reported; the kill reads long-lived |
| `terminals-per-token`, `items-per-token`, `set-reuse`, `rd-accepted-perl-refused`, `chart-accepted-perl-refused` | reported |

The baseline is reset for this attempt: keys may be added or retired when the
reset lands, and never afterwards. A regression on a gated key fails the
ratchet; an `na` key that starts printing a number fails it.

## 10. Kill criteria

Each kills **up-front mode** (§2).

- **Speed:** a long-lived ratio above **3×** on std5k or the aggregate, as the
  median of three measurements at a checkpoint (§9.3).
- **Facts in items:** a fix would need facts inside Earley items.
- **Waypoint 2:** a WRONG site fixable only by a refusal that drops coverage
  below recursive descent's; or more than 1% (and at least 5) of a corpus's
  statements collapsing.

These are stops reported to perigrin, not targets to tune toward. A definition
amended twice by the orchestrator goes to perigrin before a third amendment;
no hypothesised count is written into a criterion as a fixed number.

## 11. Isolation

Only `internal/chart` changes. Every other package is read through what it
already exports, never changed; where something needed is not exported, that
is reported, not worked around. A lexer defect found here is filed against the
v0.1 milestone and listed on `lex-unmapped`'s known list until fixed.

## 12. Testing

- **Generator:** toy grammars for the precedence filter (§8), the token table
  (each key part, unions, multi-terminal tokens) and each refusal; each test
  shown to fail before it passes.
- **Grammar:** the port ledger; the level cross-check; the CORE.tpm
  cross-check.
- **Recognizer:** the carried-over tests and fuzz target over integer ids,
  with branched copies, Leo on/off and warm/cold, and the reference
  recognizer as an oracle for acceptance and `Expected`.
- **Superset, start forms, speed:** §9.

## 13. Future work, not in this spike

- **Interleaved lexing, through ruby slippers.** The lexer asks the parser
  what is expected at a position and what is in scope there, and may produce
  the token the parser expects. Its case is time to an answer (§2): its payoff
  is M2's editing loop, where mid-edit code is incomplete and a parser that can
  supply an expected `;` or `}` recovers by insertion. It is judged against
  chapter 6's design, in which re-lexing resumes from captured lexer
  checkpoints and converges on state equality; whether the chart's expected
  sets are a better resumption point is a hypothesis for that spike, not a
  premise here. It runs at lookahead 0, so `Expected(pos)` must be cheap: the
  first attempt measured it at 94–431 ms over std5k.pl's 26,289 positions
  against 37 ms for recognition.
- **Incremental reparse and convergence,** after parity.

## 14. Process and budget

The pipeline is the same: this decision, assessment to a fixed point,
refinement, chain-review, execute, review, postmortem; branches fast-forward
on `pu`; an `Implements:` trailer on every commit; a gate review before every
merge; `make test` at 100%; workers follow ponytail.

Proposed budgets, for perigrin to adjust: **waypoint 1, 20 days**, sized for a
port of 348 alternatives over 114 nonterminals; **waypoint 2, 20 days**,
refined after waypoint 1's verdict. Twice a waypoint's budget is a stop
reported to perigrin.

## 15. References

- `docs/plans/2026-10-29-chart-parser-design.md`; `docs/postmortems/chart-waypoint.md`;
  the `chart-waypoint` milestone's `## Records`;
  `docs/assessments/chart-parser-v2-design.md`.
- Aycock, J. and Horspool, R. N. (2002), "Practical Earley Parsing", *The
  Computer Journal* 45(6); (2001), "Schrödinger's Token", *Software: Practice
  and Experience* 31(8).
- Afroozeh, A., van den Brand, M., Johnstone, A., Scott, E. and Vinju, J.
  (2013), "Safe Specification of Operator Precedence Rules", *SLE 2013*.
- Kegler, J., "Ruby Slippers parsing" (Marpa).
- Leo, J. (1991), *TCS* 82. YAEP at 7aba93e.
