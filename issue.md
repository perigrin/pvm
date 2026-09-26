---
title: Four token facts call a Word an operator, and the glossary says the lexer is right
state: pending
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T02:07:19.16565421Z
updated: 2026-09-26T02:07:19.16565421Z
---

Four corpus token facts assert something the glossary forbids, and the
glossary is the corpus's contract with B::SoN and Chalk.

## Measured, at b0f1a9e9

These four cases have a CLEAN parse and fail only on a token fact:

    arithmetic.md    Concatenation and repetition      one operator "x"
    comparison.md    The string relations and `cmp`    one operator "cmp"
    comparison.md    `&&` and `and` at two precedences one operator "and"
    comparison.md    The five logical operators        one operator "xor"

Each wants `one operator whose text is "..."` and gets 0, because the
lexer emits `Word`:

    ... Variable("$a") Word("x") Number("3") ...
    ... Variable("$a") Word("cmp") Variable("$b") ...

## The glossary says the LEXER is right

> **operator** -- Punctuation denoting an operation: `+`, `.`, `=~`,
> `->`, `?`, `:`.

> **word** -- A bare identifier: `foo`, `Foo::Bar`, `my`, `print`.
> Keywords are not distinguished from other identifiers at this layer.

`x` is not punctuation. By the glossary's own definition these are words,
which is exactly what the lexer produces. `01a0ce57` independently reached
the same conclusion in prose -- "the binary `x` is lexed correctly as a
Word" -- while the corpus scores the opposite. Two artifacts in this repo
disagree about one token.

## Why the fact was written, and why both layers are right

Perl emits `repeat` for `$a x 3` -- an operator AT THE OP LAYER:

    $ perl -MO=Concise -e 'my $a="ab"; my $b = $a x 3;'
      8  <2> repeat[t3] sK/2 ->9

`categories.go` maps the word "operator" to `lexer.Operator`, a TOKEN
kind. So a fact wanting to say "this spells perl's repeat op" had no
vocabulary for it and reached for the one word that appears in both
layers. Neither layer is wrong; the fact conflated them.

The glossary already draws this line for a neighbouring case and is worth
reading before deciding -- `quote-like operator` exists precisely because
`q{a}` is spelled with an operator NAME but is not punctuation.

## Why this is not a local fix

`GLOSSARY.md` is what lets an unknown block tag be ignored: B::SoN and
Chalk read the same corpus and answer in their own vocabulary because the
token vocabulary is agreed. Changing what `operator` means, or adding a
category, changes that contract. Do not quietly widen `categories.go` to
make four facts pass.

## The decision, stated as alternatives

1. **Rewrite the four facts in token vocabulary.** Cheapest, and it keeps
   the glossary's punctuation rule intact. Costs the ability to say "this
   is perl's repeat op" at the token layer -- which may be correct, since
   that IS an op-layer claim and the corpus has an `ir` block for those.
2. **Add a glossary category** for word-shaped operators (`x`, `cmp`,
   `and`, `or`, `not`, `xor`, `lt`..`ge`, `eq`, `ne`, `q`-family). Then
   the lexer follows and `categories.go` gains a row. Larger, and it needs
   B::SoN's agreement since it changes the shared vocabulary.

Recommend measuring which spellings would join the new category BEFORE
choosing, because the list decides how large option 2 is.

## Context

- paths: `conformance/GLOSSARY.md`, `internal/conformance/categories.go`, `conformance/mdtest/arithmetic.md`, `conformance/mdtest/comparison.md`
- commands: `go test ./internal/conformance/ -count=1`
- entrypoints: `internal/conformance/categories.go:27` (`"operator": kind(lexer.Operator)`), `conformance/GLOSSARY.md:131` (the `operator` entry)

## Acceptance criteria

- [ ] The glossary and the facts agree, whichever way it was decided (`go test ./internal/conformance/ -run '^TestGlossary' -count=1 | grep -q '^ok'`)
- [ ] No case in these two topics still refuses on a word-shaped operator fact (`! grep -A3 'parses: yes' conformance/mdtest/arithmetic.md | grep -q 'refuses: unfiled'`)
- [ ] The comparison topic's word-operator facts no longer sit unfiled (`! grep -c 'refuses: unfiled' conformance/mdtest/comparison.md | grep -q '^3$'`)
- [ ] The decision is written down where the next reader finds it (`grep -qiE 'word-shaped|not punctuation' conformance/GLOSSARY.md`)
- [ ] The suite is green (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
