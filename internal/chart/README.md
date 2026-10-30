<!-- ABOUTME: Orientation for internal/chart, the chart-parser research spike beside internal/parse. -->
<!-- ABOUTME: Covers the grammar's provenance, the Guacamole corpus, the convergence edits, the chart-ratchet output contract, the CORE.tpm cross-check, the expectation conflicting pairs, the recognizer, and the token cut's reviewed rewrite list. -->

# internal/chart

An Earley chart parser for Perl, built as a research spike beside the
recursive-descent parser in `internal/parse` and measured against it. The
design is `docs/plans/2026-10-29-chart-parser-design.md`. Nothing outside the
spike consumes this tree.

## The grammar

`grammar/testdata/guacamole/guacamole.slif` is Guacamole's Marpa SLIF grammar,
vendored at a fixed revision. `grammar/testdata/guacamole/PROVENANCE` says
where it came from. `grammar.ReadSLIF` reads the subset of SLIF it uses.

## The Guacamole corpus

`testdata/guacamole/` holds the waypoint's corpus: every `parses` and
`parsent` argument of Guacamole's own suite, recorded by running it with the
stand-in `testdata/guacamole/Guacamole/Test.pm`, one case file per unique
text. Each case is `preamble.pl` (one line) followed by Guacamole's bytes
unchanged, so the whitespace pre-commit hooks skip the case directories.
`stubs/` supplies the modules cases `use` that perl cannot find, and
`perl-rejects.txt` lists the `parses` cases `perl -c` still rejects with the
stubs as `PERL5LIB`. Those cases are not scored. `RejectsDisagree` is the
rule that list obeys.

`extract.pl` regenerates the whole directory from a Guacamole checkout.
`PROVENANCE` gives the revision and the command. Nothing here reads the
checkout afterwards. `TestGuacamoleCasesUnique` and
`TestGuacamoleCasesStartWithPreamble` guard the files, and the ratchet's
`rejects-disagree` holds the list to perl's verdicts across the whole corpus.

## The convergence edits

`testdata/convergence-edits.jsonl` holds the 200 single-statement edits to
`testdata/std5k.pl` that §8 fixes before the first step.
`testdata/convergence-edits-guacamole.jsonl` holds the same kind of edits to
the corpus's three multi-statement `parses` cases, those with at least two
editable statements after the preamble. Each line is one edit:
`offset`, `removed` and `inserted` bytes, and the statement span
`stmt_start`..`stmt_end` it lies in. The corpus list also names its `case`.
`drawEdits` draws both lists with a fixed seed, over recursive descent's
`;`-terminated statements that hold no other such statement. The guards hold
each edit inside its statement and require the edited file to parse with no
refusal. **The lists are fixed once committed.** A later step may add a list,
but must never redraw one: `-chart.update-convergence-edits` exists only to
write a list for the first time.

## The `make chart-ratchet` output contract

`make chart-ratchet` runs `cmd/ratchet`. Every later issue in the spike closes
against it, so this section is a contract. The spec stays in `docs/plans`
until perigrin accepts the spike.

- **Line form.** `make -s chart-ratchet` prints one line per measure on
  stdout, `chart-ratchet: <key> <value>`, and nothing else there. Diagnostics
  go to stderr.
- **`na`.** A key whose component has not been built prints `na`. The issue
  that builds the component makes it a number and updates the baseline in the
  same commit.
- **Exit rule.** The target exits non-zero only on a regression against
  `testdata/ratchet/baseline.txt`, never because the spike's exit is still
  unmet. A regression names the key, the baseline and the value on stderr. A
  measurement that cannot be taken, such as a missing `std5k.pl`, exits
  non-zero and names what is missing, rather than printing a number. The two
  layers return different codes. `cmd/ratchet` exits 1 on a regression and 2
  on a missing or malformed input. `make` exits 2 whenever its recipe fails,
  so `make chart-ratchet` returns 2 for both, and only stderr tells them
  apart.
- **Speed** is reported and never gated, so the target stays deterministic.
- **Keys** may be added and are never renamed. `baseline.txt` holds every key
  the ratchet prints, one line each in the line form, reported-only keys
  included. A key is added to it in the commit that adds the key. A key
  printed without a baseline line, or a baseline line no key prints, is a
  regression.
- **Inputs.** The ratchet reads only the working tree: no commit, no
  `refs/zhi`, no network. Defect yield reads `testdata/defect-yield.txt` and
  `testdata/defect-yield/`, never the chain. The ratchet sets `PERL5LIB` to
  the corpus stubs itself. Its oracle cache is its own directory,
  `pvm/chart-ratchet-oracle/<hash of stubs/>` under Go's `os.UserCacheDir()`
  (`${XDG_CACHE_HOME:-$HOME/.cache}` on Linux), and never
  `parseoracle.DefaultCacheDir`. The cache key omits the environment, so
  changing the stubs moves the run to a fresh directory. Where no user cache
  directory exists, the ratchet exits 2 rather than writing the cache into
  the tree.
- **Budget.** One run takes at most 120 s on the development machine. A cold
  cache takes about 25 s today.
- **Mainline.** `make test` never runs the ratchet. Its CI job,
  `.github/workflows/chart-ratchet.yml`, runs on changes under
  `internal/chart/**` and `internal/lexer/**` and is advisory
  (`continue-on-error: true`).

The regression rules are **may not rise** for an exit count, **may not fall**
for a corpus size or the chart's defect-yield side, and **reported only** for
everything else. Under every rule, a number may not become `na`, and a key
whose baseline is still `na` may not print a number. That fails with
"baseline still na; take a baseline", so a component's first number is
committed to the baseline with the component.

| Key | Meaning | Rule |
|---|---|---|
| `parses-scored` | Guacamole `parses` cases less `perl-rejects.txt` | may not fall |
| `parsent-scored` | Guacamole `parsent` cases | may not fall |
| `parses-refused` | scored `parses` cases the recognizer rejects; stderr names each | may not rise |
| `parses-wrong-sites` | oracle sites graded WRONG for the chart | may not rise |
| `yield-violations` | forest yields that are not their input | may not rise |
| `oracle-below-rd` | the chart's oracle scores below recursive descent's, 0 to 2 | may not rise |
| `std5k-refused` | 1 if the recognizer rejects `std5k.pl`, else 0; stderr names it | may not rise |
| `parsent-accepted` | `parsent` cases the recognizer accepts; with no error recovery yet, every rejection counts as refused; stderr names each accepted | may not rise |
| `rejects-disagree` | `parses` cases where `perl -c`, with the stubs, disagrees with `perl-rejects.txt` (`chart.RejectsDisagree`); stderr names each | may not rise |
| `speed-ratio-aggregate` | chart time over `parse.Parse` time, across the corpus | reported only |
| `speed-ratio-std5k` | the same on `std5k.pl` | reported only |
| `set-reuse` | the share of lattice positions, over `std5k.pl` and the scored `parses` cases, whose Earley set was already interned; one recognizer serves every file | reported only |
| `convergence-edits` | edits in `testdata/convergence-edits.jsonl` | reported only |
| `convergence-within-8` | the share of those edits converging within the statement plus 8 tokens | reported only |
| `defect-yield-listed` | issues in `testdata/defect-yield.txt` | reported only |
| `defect-yield-sites` | oracle sites exact for the chart and WRONG for recursive descent | reported only |
| `defect-yield-settled` | listed issues the chart settles where recursive descent does not | may not fall |
| `oracle-sites-decided` | sites the chart reports decided (§7.1) | reported only |
| `expect-states-unreproduced` | in-force `PL_expect` states the expectation classes do not reproduce | reported only |
| `rd-oracle-exact` | recursive descent's oracle exact rate on the scored `parses` cases | reported only |
| `rd-oracle-coverage` | recursive descent's share of those cases graded anything but no-answer | reported only |
| `rd-parse-ms-std5k` | `parse.Parse` on `std5k.pl`, best of 5, in ms | reported only |
| `rd-parse-ms-corpus` | `parse.Parse` over the scored `parses` cases, best of 3, in ms | reported only |
| `tables-drift` | files under `grammar/testdata/tables/` that differ from what `make chart-tables` writes, regenerated into a temporary directory; stderr names each | may not rise (baseline 0) |
| `core-tpm-differences` | builtins whose CORE.tpm shape differs from G1's, or that only one of them has (`grammar/testdata/core-tpm-differences.txt`) | reported only |
| `lex-unmapped` | files among `std5k.pl` and the scored `parses` cases whose up-front `chart/lex` stream has a token no predicate admits, or does not concatenate back to the source; stderr names each file and byte offset | may not rise |
| `expectation-conflicts` | conflicting pairs the generator observed, the lines of `grammar/testdata/tables/conflicts.txt` (see "Expected terminals, Leo and the conflicting pairs") | reported only |
| `leo-disagree` | files among `std5k.pl` and the scored `parses` cases recognized differently with Leo on and off, in acceptance or in the number of distinct sets; stderr names each | may not rise (baseline 0) |

Keys the later issues add, and the rule each is held to:
`defect-yield-chart-exact` may not fall, and `recovery-on-accepted` may not
rise.

`lex-unmapped`'s baseline is not 0. Three scored cases are refused because
`chart/lex` re-lexes a `$#{` interior from a blank lexer state. The rule fails closed. It lexes the interior only when
nothing before the `$#{` can have moved the file's state off blank: `use` or
`no` of strict or warnings, and statements with no `use`, `no`, `sub`,
`class`, `BEGIN` or heredoc opener. Anything else refuses rather than guess,
an unknown module included. A `$#{` whose own interior holds one of those, or
a nested `$#{`, refuses too, and so does every `$#{` after a refused one,
whose interior went unread. Every corpus case's preamble is a `use feature`,
so every corpus `$#{` is refused until 01a1882d-d8a4 lets the production
lexer lex the interior in place.

`testdata/lex-unmapped-known.txt` lists each unmapped file, one per line as
`<file> <lexer issue>`. The ratchet holds the failing files to that list as a
set, and names any file that fails unlisted, or is listed and no longer fails,
as a regression. The lexer issues are mainline work. As each lands, its files
leave the list and the baseline falls with them. `std5k.pl` is never listed.

To take a baseline after an intended change, run
`go run ./internal/chart/cmd/ratchet > internal/chart/testdata/ratchet/baseline.txt`
and review the diff. That run exits 1, because the redirect empties the
baseline before the ratchet reads it, but its stdout is complete.

## The recognizer: `recognize`

`recognize.Recognize` runs a YAEP-style Earley recognizer over the generated
tables and a `chart/lex` lattice. Error recovery is not built yet.

- **Items** are an LR(0) state and the distance in lattice nodes back to the
  item's origin. A set is its items, sorted, interned as a (core, distances)
  pair, so two positions with equal content share one `*Set`. A node that an
  edge passes over is inside a branch: its set is built but not interned.
- **Lookahead.** In up-front mode a complete item's rule completes only if
  the next token, or the end of input, can follow its LHS. That stops
  right-recursive `StatementSeq` from completing back through every earlier
  statement, which is what keeps sets equal from one statement to the next.
  `Expected(pos)` is the union of the expected terminals of every state in
  the set at `pos` as recognition without lookahead builds it: it closes
  the set again under every completion, outside the interned content, so a
  refusal point still names what was expected there.
- **The transition cache** is keyed by (set, token, lookahead) and the
  edge's span. A key alone does not fix the result: completion reads the
  sets where complete items began, which are not part of the source set. So
  each entry records those sets, and a hit requires them to be the same.
  `TestCacheChecksOriginSets` fails without that check. The cache is not
  used inside a branch, nor where two edges meet, and a transition that read
  any set inside a branch is never cached (`TestCacheNotKeyedOnBranchSet`).
- **Leo memos** (Leo 1991) complete a right recursion in one step. A memo is
  made at a node for a symbol B only where the node's whole set holds one
  item awaiting B, and that item is a Leo item of the tables (`leo.txt`). It
  resolves the chain of completions to its top item, by reference to the
  memo at the item's origin. Memos live in a per-node side table, never in
  a set, so two nodes whose sets differ only in their tops share one set
  (`TestLeoTopOutsideSetIdentity`). With lookahead, a chain whose top cannot
  be followed by the next token is completed link by link, as without Leo.
  Every set a chain read is a transition-cache dependency. Leo never changes
  acceptance. At lookahead 0, where interleaved mode will run, it makes
  `S ::= A S | A` and Guacamole's `StatementSeq` linear, where without it
  they are quadratic (`TestLeoLinearAtLookaheadZero`,
  `TestLeoStatementSeqAtLookaheadZero`). Linear here means the item count,
  not time: with the cache on, each use of a chain walks it to record its
  sets, so time stays quadratic until memos are interned. A chain stops at
  a start-rule item from node 0, which acceptance must see. `Options.NoLeo` turns it off, and
  the `chart_noleo` build tag makes the memo step a no-op whatever the
  options say, so `go test -tags chart_noleo` shows those tests failing.
- **The empty program.** A source of only whitespace, comments and POD is
  accepted, as perl accepts it: the cut gives `Program` an empty
  alternative, which Guacamole's `Program ::= StatementSeq` lacks.

`make fuzz-chart-recognize` runs `FuzzRecognize`. It checks that recognition
never panics, that acceptance is the same with and without the cache, the
lookahead and Leo and with a dead-end branch added, and that `Expected` is the
same with and without lookahead.

## The token cut and its rewrite list

Guacamole's grammar builds some single pvm tokens out of several lexemes: `$x`
is one pvm `Variable` token, but the grammar spells it `SigilScalar
VarIdentExpr`. The token cut (`grammar.ApplyCut`, with the reviewed table
`grammar.GuacamoleCut`) rewrites the grammar so that its terminals are **token
predicates**: a pvm lexer kind, plus the text a token of that kind must have.
Each predicate's text is the language of the grammar symbols it stands for, so
`Quote ~ LitString` admits exactly the quoted strings Guacamole's
`LitString` rules derive. Inside a sequence that was G1 in Guacamole, the
whitespace Guacamole discards may appear, because pvm's lexer folds it into the
one token (`$ foo`). Inside a G0 rule it may not, nor inside a predicate
marked `without whitespace`.

Where a Guacamole rule derives more than pvm's token can show, the predicate
admits less, so the chart refuses rather than misparses:

- **`Diamond` and `DoubleDiamond`** use `ScalarVar`, where Guacamole has
  `VarScalar`. A readline holds a name, not subscripts, so `<$x[0]>` is
  refused. perl reads it as a glob. Their predicates are also `without
  whitespace`: perl reads `<...>` as a readline only around a bare
  filehandle or a simple scalar (`scan_inputsymbol`), so `< $fh >` is a glob
  too, and refused (perigrin's ruling on 01a18656-2eec).
- **`$#{BLOCK}`** is split as Guacamole splits it:
  `DerefVariable ::= ArrayTopOpen StatementSeq RBrace`, where `ArrayTopOpen`
  is a `Variable` token whose text is `$#{`. pvm's lexer emits the whole
  construct as one `Variable` token, and no predicate admits that composite
  token. `chart/lex` (issue 01a18656-2ec0) owns splitting it into
  `ArrayTopOpen`, the re-lexed interior and `}`. Until that lands, nothing
  recognizes `$#{...}`, which is expected. After the split, the grammar
  checks the block, so `$#{ $foo->{'bar'} }` parses and `$#{1+}` is refused.
  `internal/lexer` does not change.

## The generated tables: `make chart-tables`

`make chart-tables` runs `grammar/cmd/gen`. It reads the vendored grammar,
applies the token cut, builds the tables (`grammar.BuildTables`), and writes
them to `grammar/testdata/tables/`, where they are committed. That directory
holds exactly the generator's files. The output depends only on the grammar's
rule order, never on map iteration or the clock, so regenerating it is
byte-identical (§10). `TestTablesGuacamole` and the ratchet's `tables-drift`
both hold the committed files to the generator's output. After a change to the
grammar or the cut, run `make chart-tables` and commit the result.

The tables are the split LR(0) automaton with nullable closure, Aycock and
Horspool's ε-DFA:

- A sequence rule is expanded first: `X ::= Y*` becomes `X ::=` and
  `X ::= X Y`, and `X ::= Y+` becomes `X ::= Y` and `X ::= X Y`.
- A **kernel** state holds the items a transition reached. A **predicted**
  state holds the items a kernel state predicts. Both are closed under dot
  advancement past nullable symbols.
- An Earley item's origin is its kernel's origin in a kernel state, and the
  current position in a predicted state.
- **The start state is state 0**, the start symbol's predicted state. The
  grammar has no augmented start rule, so the recognizer **accepts** when the
  last Earley set holds a complete item of one of the start symbol's rules
  whose origin is 0, meaning the item's rule appears in the reduce list of a
  state in that set.

Every line is space-separated numbers and names:

- **`symbols.txt`**: `<id> <kind> <name>`, where kind is `terminal`,
  `nonterminal` or `nullable` (a nonterminal that derives the empty string).
  The nonterminals come first, in order of their first rule, then the
  terminals in name order.
- **`rules.txt`**: `<id> <LHS> ::= <RHS>`, with sequences expanded. An empty
  RHS is an empty alternative.
- **`items.txt`**: `<id> <rule> <dot>`, the integer core items, rule by rule
  and dot by dot.
- **`states.txt`**: `start <symbol>`, then one line per state:
  `<id> kernel|predicted <ε-state> items <item ids> reduce <rule ids>`. The
  ε-state is a kernel state's predicted state, or `-1`. The reduce list holds
  the rules of the state's complete items.
- **`transitions.txt`**: one line per state, `<id> <symbol>:<target> ...`,
  in symbol order. The target is always a kernel state. A transition's items
  are the state's items whose dot is before the symbol, so they are not
  written out.
- **`expected.txt`**: one line per state, `<id> <terminal ids>`, the
  terminals after a dot in the state's items, in symbol order.
- **`leo.txt`**: one line per Leo item, `<id> <item>`. A Leo item is a
  state's one item awaiting some nonterminal B, with B its rule's last
  symbol and deriving the rule's own LHS at its right edge (`S ::= A . S`).
  Other items in the state do not matter, so `S ::= A . S` is a Leo item
  beside `S ::= A .`. The right edge reaches past trailing nullables, but an
  item with anything after B is not a Leo item: what follows B could still
  read a token, and the memo skips the item that would read it. This is
  necessary, not sufficient: the recognizer also requires the whole Earley
  set to hold one item awaiting B.
- **`probes.txt`**: `<terminal> "<text>"`, the generated samples the
  conflict finder lexed, one per printable character a pattern terminal's
  texts can start with. Literal texts are probed too and are not listed.
- **`conflicts.txt`**: one conflicting pair per line,
  `<text> <operator terminals> <term terminals> "<probe>"`.

These are separate files so that `states.txt` stays under the
check-added-large-files limit.

### Expected terminals, Leo and the conflicting pairs

An expectation class is formed at run time: `Expected(pos)`, the union of
the expected terminals of every state in a position's Earley set, projected
onto the conflicting pairs (spec §5). The pairs are observed, not derived
(perigrin's ruling, "Observe the lexer"). The generator lexes every probe
with the production lexer, once after `$a ` (`XOperator`) and once after
`$a = ` (`XTerm`); `internal/lexer` has no constructor for an expect state,
so its own state machine sets them. Where the two token streams differ, the
first differing tokens are mapped to the cut's terminals, and if both map to
some terminal that pair is listed, named by the text where the two readings
diverge. A reading no terminal admits, such as an unterminated regex, makes
no pair. A terminal list joined by commas is a token several terminals
admit.

A lexer bug therefore shows up as a wrong entry, and the review of
`conflicts.txt` is where it is caught: a pull request that changes the list
justifies each entry as one meaning in operator position and another in term
position. `TestGuacamoleConflictsIncludePerlOpTerm` requires perl's `/`, `<`,
`%`, `&`, `*` and `x`, and `TestGuacamoleConflictsExcludeLateJoiners` refuses
`{` and `?`. `-X` is not a pair: perl 5.42 lexes a filetest by its text, so
`$a -e` is a syntax error, and pvm's lexer agrees.

## The CORE.tpm cross-check

Builtin shapes come from G1's per-keyword `OpKeyword<X>Expr` rules
(`grammar.G1Shapes`). A builtin's shape is the widest any of its alternatives
takes, by the symbols after the keyword: `niladic` (nothing, `()` or a label),
`unary` (one term), `assign` (Guacamole's assign keyword operators, which
CORE.tpm has no shape for) or `list`. An argument symbol with no width is an
error, not a default.

`grammar/testdata/core-tpm-differences.txt` is committed and lists every
builtin whose shape in `internal/parse`'s CORE.tpm table (`parse.CoreShapes`)
differs from G1's, one per line: the builtin, G1's shape, CORE.tpm's. A
builtin only one side has is listed with `none` on the other side, never
skipped. `TestCoreTpmGuacamole` holds the list equal to what the cross-check
computes, byte for byte, and its failure names the list and the first line
that differs. The ratchet reports the count as `core-tpm-differences`. After
an intended change, regenerate the list:

```
go test ./internal/chart/grammar/ -run TestCoreTpmGuacamole -grammar.update-core-tpm
```

The list's review is the pull request's: a pull request that changes it walks
the changed lines in its description.

## Known divergences from Guacamole

Keyword priority is static here. Marpa arbitrates it only among the lexemes
acceptable at a position, so the two differ at a few positions:

- `$x = BEGIN()` is refused, though Marpa admits it.
- `print->foo` is admitted, though Marpa refuses it, because keyword text
  satisfies `SubNameExpr`.
- `print::foo()` is admitted. It is one pvm `Word`, which no keyword matches,
  while Marpa reads the keyword `print` first.
- `print: 1;`, a keyword as a label, diverges too. The gate review of #534
  found it.

Other divergences:

- A source with no tokens, only whitespace, comments and POD, is accepted
  through `empty Program ::=`. Guacamole refuses it; perl accepts it.
- `${foo}` is refused, though Guacamole accepts it through `DerefVariable`.
  pvm lexes it as one `Variable` token, and no predicate admits it.
- `Foo :: Bar`, with spaces, is refused. pvm lexes it as three tokens and
  perl refuses it. Guacamole accepts it.
- `:foo(bar)` is refused. pvm lexes the `(...)` of `:prototype(...)` as one
  `Prototype` token, which is what `SubAttrArgs` admits, but other attributes'
  arguments it lexes as code.

## Known ambiguities left for resolve or the interleaved lexer

Two lexeme priorities are not encoded. Marpa arbitrates them among the
lexemes acceptable at a position, which depends on the Earley set rather
than on any one rule, so they wait for `chart/resolve` or the interleaved
lexer. Until then the forest carries both readings:

- **`VersionNumber`**: `use Foo 1.2;` reads both as a version and as an
  import list, as do `no Foo 1.2;` and `require 1.2;`. **perl and Marpa read
  it only as a version**, and the chart must come to the same reading.
- **`OpAdd`**: where `OpUnary` is also acceptable, as after `sort $x`, `+`
  and `-` read both as binary and as unary. Marpa prefers the binary reading.

Every rewrite the cut makes is listed in
`grammar/testdata/token-cut-rewrites.txt`. It holds one rewrite per line and
nothing else. The first word of each line names the rewrite's form:

- **`split`**: a G1 rule that straddles pvm's token level, with the rule before
  and after. `split VarScalar ::= SigilScalar VarIdentExpr ElemSeq0 =>
  VarScalar ::= ScalarVar ElemSeq0` says that the sigil and the name are now
  one token, while the subscripts stay grammar.
- **`token`**: a terminal a split introduced, and its predicates.
  `token ScalarVar => Variable ~ SigilScalar VarIdentExpr` is a `Variable`
  token whose text `SigilScalar VarIdentExpr` derives.
- **`collapse`**: a rule wholly below pvm's token level, G1 (`::=`) or G0
  (`~`), with its alternatives, and the predicates it collapses into.
  `collapse OpMulti ~ '*' | '/' | '%' | 'x' => Operator '*' | Operator '/' |
  Operator '%' | Word 'x'` writes a literal's predicate as its kind and text.
  A G0 lexeme whose alternatives are all literals needs no table entry: each
  literal takes the kind pvm's lexer gives it, which `TestCutLiteralKinds`
  checks against the lexer.
- **`absorb`**: a G1 rule the cut left unreachable, because its text now
  lives inside a token, as `VarIdentExpr` does inside `ScalarVar`.
- **`priority`**: one line for every `:lexeme ~ X priority => n`, saying how
  the priority is carried. `except :priority` is **keyword priority**. A
  predicate marked `except :priority` refuses every word a priority lexeme
  matches, so an unqualified keyword is never a call's name: `print(1)` has one
  derivation, not two. A qualified `POSIX::exit` is one pvm `Word`, which no
  keyword matches, so it is still admitted. Keyword text still satisfies
  `SubNameExpr`, which names methods and definitions. Other lines say the lexer
  settles the priority, or that it is not encoded, and why.
- **`empty`**: an empty alternative the cut adds to a G1 symbol.
  `empty Program ::=` lets the grammar derive a source with no tokens.
- **`trivia`**: a lexer kind the recognizer skips, and the `:discard` it stands
  for, if any. `Pod`, `DataSection` and `FormatBody` are trivia because
  Guacamole has no rule for them.

`TestCutGuacamole` applies the cut to the vendored grammar and holds the
committed list equal to what the cut makes, byte for byte. A rewrite the cut
makes that is missing from the list fails, and so does a listed rewrite the
cut no longer makes. The failure names the list file and quotes the first line
that differs. When a change to the cut is intended, regenerate the list:

```
go test ./internal/chart/grammar/ -run TestCutGuacamole -grammar.update-rewrites
```

The list's review is the pull request's. A pull request that changes the list
walks the changed lines in its description, and the reviewer reads the list,
not the code that made it.
