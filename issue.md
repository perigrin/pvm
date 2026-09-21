---
title: The glossary is the token assertion's dictionary, and five entries is a start
state: in-progress
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c35d-6e43-7c0d-8828-ef7659402e39
blocks:
- 01a0c35e-1b37-7304-ba39-65a6ef57d4e0
created: 2026-09-21T09:48:13.598812961Z
updated: 2026-09-21T14:05:30.503775724Z
sessions:
- start_sha: 8c8e7a783a7f5481cbe0417328175ca7e6127fe6
  end_sha: ""
  commits: 0
  started_at: 2026-09-21T14:05:30.503775724Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T14:05:30.503775724Z
---

Token assertions are the corpus's only defence against a mis-lex, and they
are written in English: `one numeric literal whose text is ".5"`. That
English has to mean exactly one thing, or the assertion is not a check.

`conformance/GLOSSARY.md` has five categories -- numeric literal, string
literal, word, variable, operator -- each with its boundary measured against
perl. `internal/conformance/categories.go` maps those names to `lexer.Kind`.
That file is the coupling: where a glossary entry becomes evaluable.

## Why five is not enough, measured

The three spiked files use two categories. Tier 01 alone needs more before
it is covered: v-strings (`v5.42` is not a number), heredoc bodies, and
quote-like operators (`qw`, `qq`, `tr`) all have boundaries a corpus file
will want to assert, and none are in the glossary.

## The entry that proves the glossary is doing work

`-1` is TWO tokens. `perl -MO=Concise -e 'my $x = -1'` shows
`const[IV -1] s/FOLD` -- the optimiser folded the negation away, so the
optree cannot adjudicate and only the token stream can. That entry is why
the glossary exists and should stay first.

## Acceptance

- [ ] A category named in an expect-tokens line but absent from categories.go fails the file loudly, naming the category, rather than silently asserting nothing (`go test ./internal/conformance/ -run '^TestUnknownCategoryFails$' -count=1 -v | grep -q '^--- PASS: TestUnknownCategoryFails'`)
- [ ] Every category in categories.go has a glossary entry and every entry has a category, checked by test rather than by review (`go test ./internal/conformance/ -run '^TestGlossaryMatchesCategories$' -count=1 -v | grep -q '^--- PASS: TestGlossaryMatchesCategories'`)
- [ ] Each new category's boundary is asserted against the measured perl behaviour (`go test ./internal/conformance/ -run '^TestCategoryBoundaries$' -count=1 -v | grep -q '^--- PASS: TestCategoryBoundaries'`)
