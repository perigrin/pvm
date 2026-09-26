---
title: Four token facts call a Word an operator, and the glossary says the lexer is right
state: pending
urgency: normal
milestone: corpus-covers-t2
blocks:
- 01a0d0c7-052d-7744-a682-fe39d2592920
created: 2026-09-26T02:07:19.16565421Z
updated: 2026-09-26T02:56:46.202975197Z
---

Four corpus token facts assert something the glossary forbids, and the
glossary is the corpus's contract with B::SoN and Chalk.

## Measured, at b0f1a9e9

THREE cases have a clean parse and fail only on a token fact:

    arithmetic.md    Concatenation and repetition      one operator "x"
    comparison.md    The string relations and `cmp`    one operator "cmp"
    comparison.md    `&&` and `and` at two precedences one operator "and"

A FOURTH case makes the same assertion about `xor` and is NOT in scope
here: `comparison.md`'s "The five logical operators" also carries
`refusal: not_a_term` (comparison.md:112), so it refuses on the parser as
well as on the fact. It is the only one of the four that does. Rewriting
its token fact would not make it pass, so it belongs with the parser
refusals, not with this decision.

Chain-review found this: the first draft called all four "clean", which is
false for that one, and the error made this issue and the citation issue
each assume the other owned the case.

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

## Criteria corrected by chain-review, 2026-09-26

Two of the first draft's criteria did not check what they claimed:

  - `! grep -c 'refuses: unfiled' comparison.md | grep -q '^3$'` asserts
    "the count is not exactly three", so it PASSES at one and two -- a
    partial fix satisfies it. Measured: counts 0, 1 and 2 all pass. It also
    passes if the file is deleted, since `grep -c` exits 0 on a readable
    file and the whole pipeline is negated.
  - `go test -run '^TestGlossary' | grep -q '^ok'` passes today, before any
    work, and passes under BOTH routes. It distinguishes nothing.

The replacements name the three cases that must go from SKIP to PASS.
That is outcome-neutral by construction: rewriting the facts and adding a
glossary category both make those cases pass, and nothing else does.
Each of the three case criteria was verified to FAIL now (rc=1). The
glossary/category-table criterion passes today and is labelled a
regression guard, because under route 2 it becomes the real gate --
`categories.go` gains a row and that test is what holds the two lists
together.

## Acceptance criteria

- [ ] The repetition case passes rather than skipping (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -q 'PASS: TestCorpus/arithmetic.md/Concatenation_and_repetition'`)
- [ ] The string-relations case passes rather than skipping (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -qE 'PASS: TestCorpus/comparison.md/The_string_relations_and_.cmp.'`)
- [ ] The two-precedences case passes rather than skipping (`go test ./internal/conformance/ -run '^TestCorpus$' -count=1 -v | grep -qE 'PASS: TestCorpus/comparison.md/.&&._and_.and._are_one_op_at_two_precedences'`)
- [ ] The decision is written where the next reader finds it (`grep -qiE 'word-shaped|not punctuation' conformance/GLOSSARY.md`)
- [ ] REGRESSION GUARD, passes today: the glossary and the category table still agree (`go test ./internal/conformance/ -run '^TestGlossaryMatchesCategories$' -count=1 -v | grep -q '^--- PASS: TestGlossaryMatchesCategories'`)
- [ ] The suite is green (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
