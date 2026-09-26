---
title: Three token facts call a Word an operator, and the glossary says the lexer is right
state: in-progress
urgency: normal
milestone: corpus-covers-t2
created: 2026-09-26T03:02:06.779257794Z
updated: 2026-09-26T04:48:57.594125598Z
sessions:
- start_sha: 58076b8abd0e49c3879c050fa5657dfac841914f
  end_sha: ""
  commits: 0
  started_at: 2026-09-26T04:43:08.977932447Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-26T04:43:08.977932447Z
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

## The glossary's own amendment rule, which cuts against this issue's framing

`GLOSSARY.md:12`:

> **It is derived from perl, not from us.** Where a boundary is
> contestable, the entry cites `perldata`/`perlop` and, where those are
> silent, a measured `perl` invocation.

So the glossary's CURRENT text is the derivation so far, not the
authority. And perlop does call `x`, `cmp`, `and`, `or`, `xor` operators
-- which is evidence FOR route 2 that the argument above omits. This issue
opened by treating the current `operator` entry as decisive; it is not, and
a decision taken on that reading would be taken on half the evidence.

Two more rules constrain the execution:

  - `GLOSSARY.md:21` -- every `##` heading is a category name, and
    `TestGlossaryMatchesCategories` reads them. **Under route 1 the
    decision prose must go under `###` or inside an existing entry**; a
    `## Word-shaped operators are not punctuation` heading satisfies the
    prose criterion and then FAILS the guard, because no `categories.go`
    row backs it.
  - `categories.go:15` -- a category may only be added once GLOSSARY.md
    defines it.

Route 2 also needs B::SoN's agreement, and there is no B::SoN in this
repo. If that agreement is not obtainable when this issue runs, ROUTE 1 IS
THE DEFAULT -- stated here so the blocker is named rather than discovered.

## Route 2 has a side effect route 1 does not

Adding the category makes `comparison.md:120`'s `xor` fact pass with no
edit to that case. Route 1 must rewrite that fact by hand. Either way the
case stays a SKIP on `not_a_term`, which `01a0dbe8` owns, and neither
route changes this issue's criteria -- but the asymmetry is real and is
recorded so nobody discovers it as a surprise.

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
   `and`, `or`, `not`, `xor`, `lt`..`ge`, `eq`, `ne`). Costs one `##`
   heading plus ONE TEXT-KEYED PREDICATE over `lexer.Word`. **The lexer
   does not change.**

An earlier draft priced route 2 as "the lexer follows", which is wrong by
a whole subsystem and would have decided this on a false cost. Measured:
`categories.go:39-45` already splits ONE lexer kind by TEXT --

    "string literal":      k == lexer.Quote && !lexer.HasQuoteOperator(text)
    "quote-like operator": k == lexer.Quote &&  lexer.HasQuoteOperator(text)

-- and the glossary's `quote-like operator` entry states the principle:
"Our lexer gives both the same `Quote` kind, so the category is decided by
the token's TEXT... That choice is `categories.go`'s, not the corpus's."

So the spelling list is a data table, not a refactor, and its length barely
moves the cost. Measure it for completeness, not for sizing.

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

The cheapest WRONG way to close this issue is to widen
`"operator": kind(lexer.Operator)` to accept `Word` -- the prose forbids it
and, measured, nothing else would catch it. Every negative operator fact in
the corpus names punctuation; the one word-shaped spelling (`"b"`, in
`numeric-radix.md`) is inside `0b1010` so no `Word("b")` token exists
either way, and `TestCategoryBoundaries` has a single `operator` row, on
`-`. Hence the new criterion: add a row asserting `operator` does NOT match
a word, so the prohibition has a check rather than a sentence.

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
- [ ] A named row proves `operator` still rejects a word (`go test ./internal/conformance/ -run '^TestCategoryBoundaries$' -count=1 -v | grep -qE 'PASS: TestCategoryBoundaries/a_word.shaped_operator_is_not_the_operator_category'`)
- [ ] REGRESSION GUARD, passes today: the suite is green (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
### Criterion corrected during execution

The boundary-row criterion was written `a_word_shaped_operator...` with an
underscore. Go keeps the HYPHEN from the row name, so the real subtest is
`a_word-shaped_operator_is_not_the_operator_category` and the criterion
could never have passed -- a broken pattern rather than a progress check,
which is the class the chain-review gate hunts and which I wrote anyway.
Now a regex with `.` for the separator, verified against the actual `-v`
output rather than against the name I expected.
