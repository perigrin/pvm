<!-- ABOUTME: Orientation for internal/chart, the chart-parser research spike beside internal/parse. -->
<!-- ABOUTME: Covers the grammar's provenance and the token cut's committed, reviewed rewrite list. -->

# internal/chart

An Earley chart parser for Perl, built as a research spike beside the
recursive-descent parser in `internal/parse` and measured against it. The
design is `docs/plans/2026-10-29-chart-parser-design.md`. Nothing outside the
spike consumes this tree.

## The grammar

`grammar/testdata/guacamole/guacamole.slif` is Guacamole's Marpa SLIF grammar,
vendored at a fixed revision. `grammar/testdata/guacamole/PROVENANCE` says
where it came from. `grammar.ReadSLIF` reads the subset of SLIF it uses.

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
