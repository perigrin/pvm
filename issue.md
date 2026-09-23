---
title: Pattern internals are out of scope, decided rather than rediscovered
state: done
urgency: normal
milestone: corpus-complete
created: 2026-09-23T02:07:02.822021512Z
updated: 2026-09-23T19:40:23.612556801Z
sessions:
- start_sha: 66e6d10dfed49b493965299ae1c5d6fcc9088c25
  end_sha: 66e6d10dfed49b493965299ae1c5d6fcc9088c25
  commits: 0
  started_at: 2026-09-23T19:40:23.406164781Z
  ended_at: 2026-09-23T19:40:23.612556801Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-23T19:40:23.406164781Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-23T19:40:23.612556801Z
---

Record that pattern internals are OUT OF SCOPE for this corpus, and why,
so the decision is written down rather than rediscovered as a gap.

The five-axis measurement found that tier 09 tests regex DELIMITERS
exhaustively -- `//`, `m{}`, `m!!`, `m##`, and the nesting case
`m{a{b}c}` -- and pattern SYNTAX not at all. Across 1,141 lines of corpus
source there is no `(?:`, no `(?=`, no `(?<`, no `[^`, no `\d`, no
alternation, no quantifier, no `/x`, and no `$1` anywhere. T1 uses all of
them.

Stated as a gap, that reads alarming: a parser that treats a pattern body
as an opaque string passes the entire regex tier.

## Why it is not a gap, measured

A regex lexes as ONE token. Measured against our own lexer:

    $s =~ /(a)|[0-9]+/
      ->  Variable($s) Operator(=~) Quote(/(a)|[0-9]+/)

Everything between the delimiters is inside a single `Quote`. The
glossary's twelve categories have no vocabulary for pattern-internal
structure, and there is no token fact that can reach `(`, `|` or `[0-9]`
because no token exists for them.

So asserting what a pattern body contains would require one of two
things, and the plan forbids both:

- A new glossary category for pattern internals -- which makes the corpus
  assert a sub-language our lexer deliberately does not lex, and which an
  adopter whose lexer also treats patterns as opaque could not answer.
- A claim about tree shape. The plan is explicit: "No corpus file asserts
  CST shape, and none should" -- that coupling is what makes a corpus
  unusable by anyone whose tree differs.

## What tier 09 actually claims, and that it is correct

Its existing facts assert OPACITY, and they are right to:

    no operator whose text is "/"
    no word whose text is "abc"

That is the claim "the delimiter did not leak out and the contents were
not lexed as code" -- which is exactly what a Perl lexer must do, and is
a real and checkable property.

A parser that treats a pattern body as an opaque string is CORRECT by
this corpus, deliberately. Pattern semantics belong to the regex engine,
which is a separate language with its own compiler, and to a later
milestone if ever.

## What this issue produces

No corpus files. A section in `conformance/README.md` and a note in
`09_regex/README.md` recording the decision, the measurement behind it,
and the consequence -- so the next measurement against T1 or perlre
reports this as a scope boundary rather than as 37 missing files.

`14_recursive` is the one place the boundary is crossed, for `(?{ })`,
and for a stated reason: there the contents are PERL, not pattern syntax,
so the lexer must re-enter. That exception belongs in the note.

## Acceptance criteria

- [ ] The corpus README records that pattern internals are out of scope, with the measurement that decided it (`grep -q 'pattern internal' conformance/README.md`)
- [ ] Tier 09's README states that its claims are about delimiters and opacity rather than pattern syntax (`grep -qi 'opaque\|opacity' conformance/09_regex/README.md`)
- [ ] The note names 14_recursive as the deliberate exception, where the contents are Perl rather than pattern syntax (`grep -q '14_recursive' conformance/09_regex/README.md`)
- [ ] Tier 09 stays green with the documentation change (`go test ./internal/conformance/ -run '^TestTierRegex' -count=1 | grep -q '^ok'`)
- [ ] The whole conformance suite is green (`go test ./internal/conformance/ -count=1 | grep -q '^ok'`)
