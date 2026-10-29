<!-- ABOUTME: Orientation for internal/chart, the chart-parser research spike beside internal/parse. -->
<!-- ABOUTME: Covers the grammar's provenance, the Guacamole corpus, the convergence edits, the chart-ratchet output contract, and the token cut's reviewed rewrite list. -->

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
  non-zero and names what is missing, rather than printing a number.
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
  `${XDG_CACHE_HOME:-$HOME/.cache}/pvm/chart-ratchet-oracle`, and never
  `parseoracle.DefaultCacheDir`, because the cache key omits the environment.
- **Budget.** One run takes at most 120 s on the development machine. A cold
  cache takes about 25 s today.
- **Mainline.** `make test` never runs the ratchet. Its CI job,
  `.github/workflows/chart-ratchet.yml`, runs on changes under
  `internal/chart/**` and `internal/lexer/**` and is advisory
  (`continue-on-error: true`).

The regression rules are **may not rise** for an exit count, **may not fall**
for a corpus size or the chart's defect-yield side, and **reported only** for
everything else. Under every rule, a number may not become `na`.

| Key | Meaning | Rule |
|---|---|---|
| `parses-scored` | Guacamole `parses` cases less `perl-rejects.txt` | may not fall |
| `parsent-scored` | Guacamole `parsent` cases | may not fall |
| `parses-refused` | scored `parses` cases the chart refuses | may not rise |
| `parses-wrong-sites` | oracle sites graded WRONG for the chart | may not rise |
| `yield-violations` | forest yields that are not their input | may not rise |
| `oracle-below-rd` | the chart's oracle scores below recursive descent's, 0 to 2 | may not rise |
| `std5k-refused` | statements of `std5k.pl` the chart refuses | may not rise |
| `parsent-accepted` | `parsent` cases the chart accepts | may not rise |
| `rejects-disagree` | `parses` cases where `perl -c`, with the stubs, disagrees with `perl-rejects.txt` (`chart.RejectsDisagree`); stderr names each | may not rise |
| `speed-ratio-aggregate` | chart time over `parse.Parse` time, across the corpus | reported only |
| `speed-ratio-std5k` | the same on `std5k.pl` | reported only |
| `set-reuse` | the share of Earley sets served by interning | reported only |
| `convergence-edits` | edits in `testdata/convergence-edits.jsonl` | reported only |
| `convergence-within-8` | the share of those edits converging within the statement plus 8 tokens | reported only |
| `defect-yield-listed` | issues in `testdata/defect-yield.txt` | reported only |
| `defect-yield-sites` | oracle sites exact for the chart and WRONG for recursive descent | reported only |
| `defect-yield-settled` | listed issues the chart settles where recursive descent does not | may not fall |
| `defect-yield-rd-settled` | frozen defect-yield cases recursive descent is graded exact on with a marker in play | reported only |
| `oracle-sites-decided` | sites the chart reports decided (§7.1) | reported only |
| `expect-states-unreproduced` | in-force `PL_expect` states the expectation classes do not reproduce | reported only |
| `rd-oracle-exact` | recursive descent's oracle exact rate on the scored `parses` cases | reported only |
| `rd-oracle-coverage` | recursive descent's share of those cases graded anything but no-answer | reported only |
| `rd-parse-ms-std5k` | `parse.Parse` on `std5k.pl`, best of 5, in ms | reported only |
| `rd-parse-ms-corpus` | `parse.Parse` over the scored `parses` cases, best of 3, in ms | reported only |

Keys the later issues add, and the rule each is held to:
`defect-yield-chart-exact` may not fall, and `lex-unmapped`, `leo-disagree`
and `recovery-on-accepted` may not rise.

To take a baseline after an intended change, run
`go run ./internal/chart/cmd/ratchet > internal/chart/testdata/ratchet/baseline.txt`
and review the diff. That run exits 1, because the redirect empties the
baseline before the ratchet reads it, but its stdout is complete.

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
one token (`$ foo`, `< $fh >`). Inside a G0 rule it may not.

Where a Guacamole rule derives more than pvm's token can show, the predicate
admits less, so the chart refuses rather than misparses:

- **`Diamond` and `DoubleDiamond`** use `ScalarVar`, where Guacamole has
  `VarScalar`. A readline holds a name, not subscripts, so `<$x[0]>` is
  refused. perl reads it as a glob.
- **`$#{BLOCK}`** is split as Guacamole splits it:
  `DerefVariable ::= ArrayTopOpen StatementSeq RBrace`, where `ArrayTopOpen`
  is a `Variable` token whose text is `$#{`. pvm's lexer emits the whole
  construct as one `Variable` token, and no predicate admits that composite
  token. `chart/lex` (issue 01a18656-2ec0) owns splitting it into
  `ArrayTopOpen`, the re-lexed interior and `}`. Until that lands, nothing
  recognizes `$#{...}`, which is expected. After the split, the grammar
  checks the block, so `$#{ $foo->{'bar'} }` parses and `$#{1+}` is refused.
  `internal/lexer` does not change.

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
