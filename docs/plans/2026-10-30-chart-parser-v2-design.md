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
alone fell from 34.9 to 18.3 ms on std5k.pl once its keys were bytes, against
recursive descent's 18.6–30.8 ms on the same machine (postmortem).

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
parsing algorithm. That machine is the feedback from toke.c that
`docs/specs/perl-parser/01-scope.md` §1.2 said a transcription of `perly.y`
would lack; `perly.y`'s conflicts are kept as ambiguity and resolved in the
forest (§8), not inherited as an LALR table.

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
the corpus labels are perl's verdicts. Three parity limits are stated, not
hidden: quote-like bodies are single tokens (interpolated code, `s///e`
bodies and `qw` lists are not recognized), as they are for recursive descent
today; library-declared operators are waypoint 2's (§9.4); and keyword
plugins, whose syntax is defined by loaded XS code, are excluded.

**Waypoint 1's speed figure is a lower bound.** Precedence, the forest and
scope all cost time after recognition (§8), so waypoint 1 can kill the spike
but cannot pass it; the verdict on speed is waypoint 2's checkpoint 3.

### Out of scope

- Interleaved lexing (§13); incremental reparse and convergence.
- Changes to any package outside `internal/chart` (§11).

## 3. Design rules

Rules 1–6 of the first attempt's §3 carry over: the grammar is static and
context dynamic; context enters only before the chart and after it, never
inside Earley items; only perl's own rules may prune; ambiguity may stay;
outside the accepted language, refuse, never misparse; weights may order
readings, never decide them. Rule 5 holds at waypoint 1 except for
refusals that depend on precedence or scope, which hold from waypoint 2 (§8, §9.4). Rule 6 is sharpened, not relaxed: a constraint
derived from perl's own declarations may decide which readings exist, and
refuse a span where none does, after the chart (§8); a weight may still only order. This attempt adds:

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
    already computed, and its neighbours (§7). The **grammar** collapses it
    during recognition: a reading no item expects dies there. **Scope**
    (declared subs, imports, features) collapses what remains, after the chart
    in waypoint 2, as precedence does (§8). The recognizer stays Boolean and
    its items stay (state, distance).

## 4. Sources

| Source | Revision | Role |
|---|---|---|
| `perly.y` | perl blead 94e50866 (2026-07-16), `PERL_VERSION 45` (`patchlevel.h:40-41`), `/opt/corpora/perl5` | Ground truth: 348 alternatives over 114 nonterminals (bison 3.8.2 report), 32 precedence lines (`perly.y:150-182`), 39 `%prec` uses |
| `toke.c` (same tree) | as above | How each token type is decided, and which tokens it invents |
| `docs/specs/perl-parser/` chapters 2–6 | on `pu` | The recursive-descent parser's derivation from the same sources; chapter 4's levels; chapter 6's lexer checkpoints |
| `internal/lexer` | on `pu` | The tokens and flags the grammar is stated over |
| Aycock & Horspool (2002, 2001) | | The split LR(0) automaton; tokens with several types ("Schrödinger's Token") |
| Afroozeh et al. (2013); Afroozeh & Izmaylova (2016) | | Operator precedence as declarative constraints over a forest; this design departs from their safety property and refuses as bison does (§8) |
| Goodman (1999) | | Semiring parsing: values computed over a parse forest (§8) |
| YAEP 7aba93e; Leo (1991) | | Interning, the transition cache; linear right recursion |

`perly.y` is read and cited by people; no build step reads it. The
precedence referee (§8) is generated from it offline and committed.
Hand-off items H1–H5 are listed in the assessment minute,
`docs/assessments/chart-parser-v2-design.md`, under "Handed to refinement".

## 5. Components

### 5.1 Carried over from chart-waypoint (on `pu` at 4fffe4d6)

- **`grammar`'s generator:** the LR(0) automaton with nullable closure,
  transition and reduce lists, expected-terminal sets, Leo eligibility,
  byte-identical regeneration.
- **`recognize`:** interned sets, the transition cache with its dependency
  rule, lookahead pruning, `Expected(pos)`, Leo memos outside set identity,
  and the `Options` switches the differential tests are built on. Its input
  changes from terminal names to terminal-set ids (hand-off H2).
- **`cmd/ratchet`'s comparison,** with its keys reset to a new baseline.
- **Fixtures:** `std5k.pl` and its generator; the Guacamole snippets,
  relabelled by perl's verdict; the 200 convergence edits, unused until §13.

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
  mid-rule actions are not ported. An alternative whose action refuses
  *unconditionally* is not ported as an accepting rule; the ledger marks it
  "refused in action". The four comparison-chain alternatives among them
  (`perly.y:1466-1487`) derive a dangling operator (`$a <=> $b <=>`), which
  the rules as written already refuse; the chained *sentence* they exist to
  refuse (`$a <=> $b <=> $c`) has another derivation, so its refusal is §8's.
  The ledger marks them "refused by precedence"; the dangling form keeps a
  waypoint-1 refusal test (§12). An action that
  refuses only under a syntactic condition (a `catch` without a variable)
  becomes a fitted rule that excludes that case, and the alternative that
  case needs (`catch_paren: empty`) is excluded with that reason. An action
  that refuses under scope or features (`method` and `ADJUST` outside a
  class, signatures not enabled) is a scope refusal, waypoint 2's (§9.4). Where `perly.y` spells as several
  terminals what the lexer emits as one token, the rule is fitted to the
  token (a token maps to a *set* of terminals, never a sequence); where the
  lexer emits several tokens for one terminal, the rule is fitted to the
  tokens. The families needing fitting are variables and indirect objects;
  braces and the `;` toke.c inserts; quote-likes and interpolation;
  formats, labels, attributes and postfix dereference; heredocs; forced
  version tokens; `NOAMP` and method-call tokens; the scope-decided `*SUB`
  and `PRIVATEREF`; phasers. Their detail is in the hand-off (H4).
- **Start forms.** The `GRAM*` tokens are start selectors, not lexed tokens
  (`perly.y:187-268`). All seven are ported as start rules. Waypoint 1 grades
  the program form; the others are tested only where they lex the same from
  statement start, since the lexer exports only statement-start states, and
  that limit is recorded.
- **`PLUGIN_*` operators.** The eleven classes (`perly.y:153-178`, rules at
  `1256-1470`) are ported as grammar. How a library operator reaches them is
  waypoint 2's (§9.4).
- **The token table (§7).**
- **Generator refusals.** The generator emits no tables for an inconsistent
  grammar: an uncited rule or a citation outside the reference list; a literal
  the lexer does not lex to its class's kind and text inside a short context
  string for that class; a terminal with no token class (the `PLUGIN_*`
  terminals, waypoint 2's, the `GRAM*` start selectors, and the keyword-plugin
  tokens `PLUGEXPR`/`PLUGSTMT`, excluded, excepted); an
  unreachable or unproductive nonterminal; a precedence token used by no rule
  and no `%prec`. The exact checks are in the hand-off (H5).
- **The port ledger.** A committed reference list of `perly.y`'s alternatives
  at the pinned revision anchors it; every alternative is marked ported,
  fitted or excluded with a reason, and a test holds the grammar's citations
  and the ledger equal (§9.2).
- **A thin chart/lex:** the production lexer up front, trivia attached to the
  next significant token (heredoc bodies to their heredoc's opening token
  instead), then one table lookup per token.
- **A reference recognizer (test-only):** a plain Earley recognizer used by
  the fuzz target, toy grammars and a corpus sample.

### 5.3 Discarded

The SLIF reader, the token cut, chart/lex's predicate mapping, the
observed-conflict probes, the Guacamole CORE.tpm cross-check and Guacamole's
grammar and tables, in one dedicated commit. The CORE.tpm cross-check is
replaced, not dropped: the port's builtin classes are compared with
`parse.CoreShapes()` under a stated correspondence of shapes to token types,
and every difference is listed.

## 6. Data flow

```
source ──internal/lexer (up front)──▶ tokens + trivia + flags
       ──chart/lex: one lookup per token──▶ terminal-set ids
       ──recognize (interned sets, cache, Leo; Boolean)──▶ chart
       ──accepted, or refused at the first empty set with Expected there
```

Waypoint 2 adds: chart → forest → precedence and scope constraints over the
forest (§8, §9.4) → the oracle.

## 7. The token interface

**The key.** A token's terminal set is looked up by its lexer kind, its text,
the flags the lexer has already computed (`OpensBlock` and `TermFollows`),
its sigil or quote-prefix class, the classes of the **two** previous
significant tokens, and the class of the next. A neighbour's class is its
lexer kind, its flags and, for words, operators and closing brackets, its text: never its own
terminal set, so the key does not recurse. The two previous tokens are what
make a lone word inside `{}` a string in `$h{foo}`, `$h->{foo}` and
`f()->{foo}` and a call in `+{ foo }`; the lexer's `OpensBlock` is false on
those braces and cannot, and which pairs mean which is the grammar's to state
and the table's to hold. A `{` neighbour's class carries whether it opens a
subscript or an anonymous hash, derived from its own predecessor as toke.c
does, so `$h{-foo}` and `+{-foo}` get different keys at `foo` though both are
three tokens back from it; this is still a class, not a terminal set, so the
key does not recurse. Heredoc bodies are not neighbours. Where the lexer has already
decided block-or-hash at `{`, the key keeps that decision. The next token's
class is what toke.c reads to tell `FUNC` from `LSTOP`, `FUNC1` from
`UNIOP`, and an auto-quoted word before `=>`.

**Compiled.** The generator builds the table from the grammar's token classes,
with unions precomputed, so the lookup is constant-time, allocation-free and
regexp-free (a sparse table refined by a neighbour only where a rule needs
it), and yields an integer terminal-set id. The recognizer works only
in ids. An empty set refuses the file at that byte (`lex-unmapped`).

**Several terminals.** Where the key still cannot decide (a bareword that may
be a declared sub, an indirect object), the token maps to every terminal it
could be. The grammar collapses what it can during recognition (unary or
binary `-`, and `x` as operator or word, are decided by position this way);
scope collapses the rest in waypoint 2. Ambiguities the neighbouring tokens
decide are covered by the key, not left to scope.

**What it costs.** Superposition moves cost from the lexer into the
recognizer, so §9 reports terminals per token and Earley items per token
alongside chart/lex's own overhead.

## 8. Precedence: in the forest, as bison resolves it

This section is the only statement of how precedence works; other sections
refer to it.

`perly.y`'s precedence declarations do not stratify the grammar; bison applies
them only where a shift/reduce conflict arises. Imitating that resolution
inside an Earley recognizer would be faithful, not fitted. So precedence
leaves recognition (perigrin, 2026-10-30): recognition uses `perly.y`'s rules
as written, is Boolean, and is ambiguous on operator chains, keeping every
derivation those rules allow.

**In waypoint 2's forest, precedence orders and refuses, as bison does**
(perigrin, 2026-10-30). The declared levels, associativity and the
non-chaining comparisons' `%prec` act on the forest in two steps. First,
where two derivations of one nonterminal cover one span, the one bison's
resolution would build is kept. Second, a derivation bison's resolution would
never build is forbidden: a `%nonassoc` operator chained with its own level
where bison's resolution is an error action (infix after infix, prefix with postfix,
but not prefix after prefix), and a non-chaining comparison chained with
another, including chains split by a tighter operator
(`1 .. 2 + 3 .. 4`, `$a <=> $b + 1 <=> $c`). A span whose every derivation is
forbidden is refused, and so is a file whose root is. These shapes are
examples; the definition is bison's resolution of each conflict, and the
generated tests below are what pin it. The value carried up the forest is
each derivation's exposed operator context at both edges, not one bit per
node.

This is narrower than the safe semantics of Afroozeh et al., which never
refuse: perl is bison, and bison refuses these on purpose. Participant E's
probe (minute, row E18 and round 6) showed this definition bison-exact on a
six-level operator ladder; that it holds over all of `perly.y` is what the
referee tests below establish, and §10 says what happens if it does not.

**Bison is the referee, offline.** The constraint is tested against a
bison-built parser of the pinned `perly.y`, with every action stripped
except those that call `yyerror`, which are kept as refusals. A sentence is
refused if bison reports any error, whatever `yyparse` returns, since
`sideff: error` recovers and returns success. Same acceptance and same grouping, which is the outside reference
rule 11 requires (the oracle's marker sites do not cover grouping). Bison and
a C compiler are needed only to regenerate the referee's output, never by
`make test` or the ratchet: a script, pinned to the `perly.y` revision,
writes each generated sentence, bison's verdict and bison's grouping to
committed fixtures, and the Go tests read the fixtures, as the perl labels
are read. Each sentence is stored as `perly.y` terminals and as source text
rendered from them, one sample lexeme per terminal from a committed rendering
table; the chart is run on the source text through `internal/lexer`, so the
fitted grammar is tested as it will run. A sentence is kept only if its
rendering lexes back to tokens whose classes map to the same terminals (an
inserted `;`, or `UNIOP (` read as `FUNC1`, fails this); a terminal with no
single-lexeme rendering, and every sentence that fails to round-trip, are
counted on a reviewed list, never dropped silently. The referee's grammar
omits `catch_paren: empty`, as the port does (§5.2), rather than reproduce
the action that refuses it. Perl is the oracle on the
corpora: waypoint 2's WRONG 0 grades accepted files, and §9.2's refusal gate
grades refused ones. The constraint covers every conflict bison resolves by
precedence, not only the operator ladder: `%prec` on a rule that is not an
operator (`indirob: scalar %prec PREC_LOW`, which refuses `print $fh []`) is
resolved the same way. Which forms are refused is found, not listed: the
tests generate sentences that `perly.y` with its precedence and `perly.y`
without it (GLR, every derivation kept) disagree on, and each disagreement is
a refusal or grouping test. The unit of coverage is a (rule, token)
resolution in bison's report. A resolution is covered by a sentence whose
verdict or grouping depends on it; one resolved as an error is covered only
with an accepted neighbour as well (`print $fh []` and `print $$fh[0]`), so
an over-refusing constraint fails as surely as an under-refusing one.
Coverage may not fall, and at waypoint 2's exit every resolution is covered
or is on a reviewed list with a reason (generation timed out, or the
rendering cannot reach it). Resolutions involving `PLUGIN_*` terminals are
counted separately until §9.4 decides how library operators reach them. The `PLUGIN_*` operators' refusals are
in scope with the library operators in waypoint 2 (§9.4); only the keyword
plugins `PLUGEXPR`/`PLUGSTMT` are excluded (§2).

**Waypoint 1 does no precedence, and over-accepts.** Precedence-dependent
refusals are waypoint 2's, so waypoint 1 accepts files perl refuses for
precedence alone; its refusal gate excludes them by their label (§9.1), and
their tests are waypoint 2's, the four comparison-chain alternatives of §5.2
included. The ambiguous chains' cost shows in
`items-per-token`; operator chains in real Perl are short (a probe recorded in
the assessment).

## 9. Waypoints and the ratchet

### 9.1 Corpora and the superset population

| Corpus | Labels |
|---|---|
| Guacamole snippets (`parses`, `parsent`) | the pinned perl's verdict under the declared preamble (perl 5.42 accepted 78 of the 124 `parsent`; blead 94e50866 accepts 79, so labels and counts are regenerated) |
| `std5k.pl`, `edit5k.pl` (`internal/psc/testdata/latency/`, read as files) | both compile under perl |
| T1 (`PerlOnJava/unit`, 986 files, top level) | perl's verdict and recursive descent's (it refuses 236) |
| perl.git `t/` (waypoint 2) | perl's verdict; the oracle's marker sites |

**Recursive descent accepts** a file when **no node** in `parse.Parse`'s
tree reports a refusal (`Node.Refused()` false at every node), the measure
behind the 236 above. The **superset population** is every file in the
waypoint-1 corpora that **both perl and recursive descent accept**, less files
on `lex-unmapped`'s known list or the lexer-defect list (§9.2, §11). Files recursive descent accepts but
perl refuses are reported, not gated, since accepting them would break rule 5.

**Perl's labels are pinned, not live.** They are produced once and
committed as a label file recording how they were made (the perl binary and
version, working directory, `PERL5LIB`, timeout). One perl produces both the
labels and their evidence: a `-DDEBUGGING` build of the pinned blead, built
offline (refinement sizes it). Populations do not change
with the machine that runs the ratchet. Each perl refusal carries a category
that is reviewed when the file is committed: **grammar** (the construct is not
in `perly.y`'s language, including its unconditional action refusals and
toke.c's fatal lexing errors); **precedence** (refused only by bison's
precedence resolution, §8, which the label shows by evidence rather than
judgment: perl's token stream for the file, aligned as §12 states and taken
up to the token where perl refused, is a viable prefix of `perly.y` without
precedence, and bison with precedence refuses at that token; this speaks for
perl's first error only, so a later grammar error in the same file is not
seen, and is recorded as a known limit); **scope** (refused in an action under scope or features, §5.2);
**semantic** (a compile-time check after parsing,
such as arity, prototypes or op.c's "Can't modify"); or **environment** (a
missing module, `-T`, a runtime failure). The `isa` and similar feature-gated
forms are labelled under the feature set the preamble declares. Syntax only the newer perl accepts shows in
`chart-accepted-perl-refused`.

### 9.2 Waypoint 1: recognizes Perl

Exit, all at once:

- **The port is complete:** every `perly.y` alternative is in the port ledger
  as ported, fitted or excluded with a reason, and every grammar citation is in
  the ledger. The generator's refusals pass; `tables-drift 0`.
- **`lex-unmapped`** equals its known list. Each entry names an open lexer
  issue and the byte where the chart refuses, and the gate review confirms
  the construct there is that issue's, so a gap in the chart's own table
  cannot be listed away. No unlisted file is unmapped. Files whose verdict
  traces to a wrong but mapped token are a separate list (§11), not this one.
- **Superset:** `rd-accepted-chart-refused 0` over the superset population,
  with the list printed on failure.
- **Refusal (rule 5):** `chart-accepted-both-refused 0`: the chart accepts no
  file labelled a **grammar** refusal that recursive descent also refuses,
  less files on either lexer list. Precedence and scope refusals are
  reported at waypoint 1 and join this gate at waypoint 2; semantic and
  environment refusals are reported
  (`chart-accepted-perl-refused`), never gated. The corpora hold few such
  files at waypoint 1, so §12's refusal tests carry most of rule 5 until
  perl.git `t/` arrives.
- **Exclusions are bounded:** an alternative may be excluded from the port
  only for a reason on a closed list (a keyword plugin, an unconditional
  action refusal, the alternative a fitted conditional refusal needs, bison's
  `error` recovery rule `sideff: error`), and the count of
  exclusions is reported.
- **Builtins and levels:** the CORE.tpm cross-check runs and its differences
  are listed; the levels declared in the grammar are compared with chapter
  4's table and `parse.LevelsAccountedFor()`, and differences listed.
- **`leo-disagree 0`**, defined as Leo on and off agreeing on acceptance and
  on `Expected` at every position.
- **Speed:** measured at the checkpoints of §9.3.

These are the waypoint's exit, met by its final issue. Every issue before it
holds the ratchet's no-regression rule, so it may not make any key worse, but
is not required to meet an exit gate the unfinished port cannot meet yet. A
key's baseline is taken by the first issue that makes it print a number, in
that issue's commit.

### 9.3 Speed protocol

**Ratios.** Every ratio is chart over recursive descent, timed **alternately**
in one process (chart, RD, chart, RD…) so machine load affects both sides
alike. Within a timed run, collection is left to the runtime, so a side
pays for the collection its own allocation causes during the run; a GC is
forced between runs, outside the timed span, so neither inherits the other's
garbage. Allocation per token is reported for both sides, since a run too
short to trigger a collection would otherwise hide it. A **measurement** is 2 warm-up pairs then 9 timed pairs;
the ratio is the median of the 9 per-pair ratios, with min and max reported.
`lex-overhead` is timed the same alternating way.

**Two lifetimes,** each covering every file exactly once per timed pass, so
no pass replays a cache warmed on its own inputs:
- **long-lived:** one recognizer per pass, reused across that pass's files,
  as an editor holds one across documents (this is the first attempt's
  measurement);
- **cold:** a fresh recognizer per file.

**Figures.** `std5k` is std5k.pl alone. The **aggregate** is the summed time
over the superset population, excluding std5k.pl, edit5k.pl and any file
that alone takes more than 10% of recursive descent's time over the
population (such as T1's `code_too_large.t`); excluded files and each corpus's
share are reported beside it. The kill reads the long-lived std5k and
aggregate ratios. For std5k.pl, long-lived and cold coincide, since one file
is one pass; the first attempt's std5k figure was warmer than either.

**Checkpoints.** A speed measurement is taken (1) as soon as std5k.pl is
recognized; (2) at waypoint 1's exit; (3) at waypoint 2's exit, forest and
constraints included. At checkpoint 1 the kill reads std5k, and reads the
aggregate only once at least half the population is recognized; the 10%
exclusion is always computed over the whole population. Earlier issues report
speed for warning. The kill may fire at any checkpoint.

**Gates that can fail.** The ratchet (`make chart-ratchet`) is in every
recognizer issue's acceptance criteria, with its no-regression rule on count
keys; timing keys are never compared strictly, since unchanged code varies
from run to run. The exit gates are the final issue's (§9.2), with one
exception the postmortem asked for: `lex-overhead`, chart/lex's total time
(including its call to `lexer.Tokenize`) over `lexer.Tokenize`'s on the same
files over the aggregate, is gated at an absolute 1.3, as the median of three
measurements over the aggregate (std5k's figure is reported, since one file
is noisier), in every issue that touches chart/lex or the token table, from
the first. Benchmarks for lexing, recognition
and `Expected` run in the ratchet, not in `make test`.

### 9.4 Waypoint 2: parses Perl as recursive descent does

The forest; the precedence constraint (§8); resolve, which collapses by
scope; the readings count, taken after precedence and scope have acted, with
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
| `tables-drift`, `leo-disagree` | may not rise; 0 at the waypoint exit |
| `rd-accepted-chart-refused` | may not rise; 0 at the waypoint exit |
| `lex-unmapped` | may not rise; equals its known list at the exit |
| `lex-overhead` | ≤ 1.3 (median of three) in every issue touching chart/lex or the table |
| `chart-accepted-both-refused` | may not rise; 0 at the exit |
| `speed-ratio-{std5k,aggregate}-{longlived,cold}` | reported; the kill reads long-lived |
| `terminals-per-token`, `tdt-unseen` | reported until the port is complete; may not rise without review from the issue that completes it, which takes their baselines; every `tdt-unseen` pair is on a reviewed list at the exit (§12) |
| `tdt-mismatch` | reported until the port is complete; 0 from then on |
| `tdt-aligned` | may not fall; at the exit every unaligned token is in a reviewed fitting family (§12) |
| `items-per-token`, `set-reuse`, `rd-accepted-perl-refused`, `chart-accepted-perl-refused` | reported |
| `speed-ratio-*`, `lex-overhead` | timing: never compared strictly between issues |

The baseline is reset for this attempt: keys may be added or retired when the
reset lands, and never afterwards. A regression on a gated key fails the
ratchet; an `na` key that starts printing a number fails it unless that issue
takes the key's baseline (§9.2).

## 10. Kill criteria

Each kills **up-front mode** (§2).

- **Speed:** a long-lived ratio above **3×** on std5k or the aggregate, as the
  median of three measurements at a checkpoint (§9.3).
- **Facts in items:** a fix would need facts inside Earley items.
- **Precedence not expressible:** a precedence resolution bison makes cannot
  be stated over the forest from a span's derivations and its neighbouring
  tokens (a stop to perigrin, not a kill until ruled).
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
v0.1 milestone and listed on `lex-unmapped`'s known list until fixed. A file
whose chart verdict, accept or refuse, traces to a lexer token that is wrong
but mapped goes on a separate **lexer-defect list**, each entry naming the
open lexer issue and the byte of the wrong token, under the same review; such
a file leaves both gates' populations.

## 12. Testing

- **Generator:** toy grammars for the token table (each key part, unions,
  multi-terminal tokens) and each refusal; each test shown to fail before it
  passes.
- **Token table, against toke.c:** the pinned `-DDEBUGGING` perl, run
  offline, records `-DT`'s token stream over the corpora, and a committed
  script aligns it to `internal/lexer`'s tokens; the aligned pairs are
  committed fixtures. Alignment is the real work and is stated here, not left
  to an issue: modules a file loads are stubbed or their tokens dropped by
  toke.c's nested-EOF bracketing, so only the file's own tokens remain; tokens
  are matched in order by text; toke.c's split, invented and sublexed tokens
  (the `$` and name of a variable, `force_next` tokens, quote interiors) are
  folded by the reviewed fitting families (H4) onto the pvm token they belong
  to. Three keys (§9.5): **`tdt-aligned`**, the fraction of pvm tokens
  aligned, may not fall, and at the exit every unaligned token belongs to a
  reviewed fitting family, so the check cannot pass by aligning less;
  **`tdt-mismatch`**, aligned tokens whose toke.c type is not in the chart's
  set, 0 once the port is complete; and **`tdt-unseen`**, the (key,
  terminal) pairs in the table that toke.c never produced over the corpora,
  each on a reviewed list at the exit, so an over-wide table shows as surely
  as a too-narrow one. These are the width limit waypoint 1 otherwise lacks:
  T1 holds no grammar refusals, so the refusal gate grades little at
  waypoint 1. This is the table's outside
  reference under rule 11.
- **Grammar:** the port ledger; the level cross-check; the CORE.tpm
  cross-check.
- **Recognizer:** the carried-over tests and fuzz target over integer ids,
  with branched copies, Leo on/off and warm/cold, and the reference
  recognizer as an oracle for acceptance and `Expected`.
- **Superset, start forms, speed:** §9.
- **Precedence (waypoint 2):** §8's generated sentences against the bison
  referee's fixtures; toy grammars for the edge-context mechanism, with a
  bison build of each toy as their referee.
- **Refusals:** each alternative the ledger marks "refused in action", and
  the dangling form of each "refused by precedence" one, has a
  waypoint-1 test that the chart refuses it, beside a nearby form perl
  accepts. Precedence refusals (§8), adjacent and split chains alike, are
  waypoint 2's tests, against the bison-built parser and perl.

## 13. Future work, not in this spike

- **Interleaved lexing, through ruby slippers.** The lexer asks the parser
  what is expected at a position and what is in scope there, and may produce
  the token the parser expects. Its case is time to an answer (§2): its payoff
  is M2's editing loop, where mid-edit code is incomplete and a parser that can
  supply an expected `;` or `}` recovers by insertion. It is judged against
  chapter 6's design, in which re-lexing resumes from captured lexer
  checkpoints and converges on state equality; whether the chart's expected
  sets are a better resumption point is a hypothesis for that spike. It is
  judged on M2's measure (an edit to a re-parsed tree under 10 ms at p99), and
  its known costs are recorded in the assessment.
- **Incremental reparse and convergence,** after parity, using the carried
  convergence edits. A positive verdict here makes the chart a candidate for
  M2 only once convergence is measured.

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
- Afroozeh, A. and Izmaylova, A. (2016), "Operator Precedence for
  Data-Dependent Grammars", *PEPM 2016*.
- Goodman, J. (1999), "Semiring Parsing", *Computational Linguistics* 25(4).
- Kegler, J., "Ruby Slippers parsing" (Marpa).
- Leo, J. (1991), *TCS* 82. YAEP at 7aba93e.
