---
title: "Sigils are structural: the lexer swallows whole dereferences"
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0ac52-a68c-750b-92d4-852abfc54a2f
created: 2026-09-17T03:04:49.380252705Z
updated: 2026-09-17T05:28:00.097813211Z
---

An attempt at this measured worse than it looks, and the reason is issue 01a0ac52.

## What was tried

The lexer's `scanVarName` swallows `${EXPR}` and `$$x` whole, so the parser
sees a leaf:

    "${$h->{k}};"   [source_file statement term]

The fix: a `DerefSigil` token for a sigil applied to an EXPRESSION rather
than a name, with `${name}` staying a Variable because the braces there are
punctuation around an identifier (`"${foo}bar"`). The parser turns it into a
`Unary` whose operand is the interior. That part works:

    "${$h->{k}};"   [source_file statement unary index term call]
    "@{[ 1, 2 ]};"  [source_file statement unary anon_array term term]
    "$$x;"          [source_file statement unary term]

## Why it is not committed

Measured on the PerlOnJava corpus (1,935 files) against 3b9422eb:

                      before    after
    files clean       856       854
    Unknown nodes     8,432     8,155
    round-trip        0 fail    0 fail

    27 files worse, 10 files better

The aggregate -277 is not what it appears. 335 of the 345 improvements come
from THREE files:

    -136  module/Class-MethodMaker/t/array.t   (180 -> 44)
    -122  module/Class-MethodMaker/t/hash.t    (167 -> 45)
     -77  module/Class-MethodMaker/t/scalar.t  (131 -> 54)

against 27 files made slightly worse, including two that were clean and are
no longer (`unit/indirect_object_constructor.t`, `unit/subs_pragma.t`). A
change that fixes one construct in three files and breaks something small in
27 others is not ready, whatever the total says.

## The blocker, which is this issue's own subject

Splitting the sigil off the brace exposes a latent bug in the expect state
machine. `expect.go`'s CloseBracket case applies perl's right-paren
lookahead to EVERY closer:

    if t.nextIsOpenBrace {
        return XBlock
    }

The comment above it says "A `)` with a `{` next", and cites
`yyl_rightparen` (toke.c:7156) -- which handles `)` alone. The code checks
no byte. So `${$y}{Keys}` classifies its second brace as a BLOCK:

    DerefSigil   "$"      block=false
    Operator     "{"      block=false
    Variable     "$y"     block=false
    CloseBracket "}"      block=false
    Operator     "{"      block=true     <- wrong
    Word         "Keys"   block=false

and the statement parses as a bare block. Nothing exposed this before,
because `${$y}` lexed as one Variable token and no CloseBracket ever
preceded the subscript.

Restricting the rule to `)` is a one-line change and it is in the stash. It
takes Unknown from 8,459 to 8,155, and it is the point at which this stops
being a dereference fix and becomes brace classification -- 01a0ac52's
subject. Two more term-position brace decisions sit behind it.

## What is already done and should be reused

Stashed as `deref-lexer-only-01CRysQ`, SHA af8ca8f76b87e29d7837dedb2cce1848fddd9423:

- `DerefSigil` token kind, with the `${name}` vs `${EXPR}` discrimination,
  including the whitespace case `$ {*$glob}{Keys}` (legal, measured on perl
  5.42.0) and the caret control variables `${^TAINT}` (a NAME, caught by
  TestLexDotTGoldenStream).
- `expect.go`: DerefSigil leaves XTerm, since a sigil has not produced a
  value yet.
- `term.go`: the `Unary` term, with the braces of `${EXPR}` treated as
  GROUPING rather than as an AnonHash constructor.
- `call.go`: `bpDeref = 300`, above `->` at 290, so `$$x[0]` is `${$x}[0]`.
- `internal/parse/deref_test.go`: four tests covering the interior, the
  round trip, `${name}` staying a name, and `$$` staying the process id.
- Two ratchet updates that are genuine improvements: `comp/parser.t` 1 -> 0,
  and a golden-stream change where POD inside a dereference block now lexes
  as an expression (verified against perl: the block does contain POD and an
  expression, and evaluates).

## Acceptance

- [ ] `${$h->{k}}`, `@{[ ... ]}`, `$$x`, `@$x` each have their interior in
      the tree, as children rather than as swallowed bytes.
- [ ] `${name}` stays one Variable; `$$` stays the process id; `${^TAINT}`
      stays a name.
- [ ] No file in the PerlOnJava corpus gains Unknown nodes. The three
      Class-MethodMaker files are expected to improve; the 27 regressions
      must be understood and fixed, not averaged away.
- [ ] Files clean does not decrease.
- [ ] Round-trip stays at zero failures.
- [ ] The `nextIsOpenBrace` rule applies to `)` only, with a test that
      `${$y}{Keys}` and `$a[0]{k}` classify both braces as terms.

Blocked by 01a0ac52.
