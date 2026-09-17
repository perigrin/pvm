---
title: "Sigils are structural: the lexer swallows whole dereferences"
state: pending
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0ac52-a68c-750b-92d4-852abfc54a2f
- 01a0a70f-3ccc-77bb-902d-9a0b9d8050a6
created: 2026-09-17T03:04:49.380252705Z
updated: 2026-09-17T14:24:46.166879028Z
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

Measured on T1 graded (PerlOnJava `unit/*.t`, 986 files, top level only),
against d2ebe02a:

                      before    after
    files clean       408       406
    Unknown nodes     4,944     4,986
    round-trip        0 fail    0 fail

Worse on both axes.

An earlier run of the same change looked favourable -- 8,432 -> 8,155
Unknown, a gain of 277 -- and that measurement was against the WRONG CORPUS:
1,935 files found by walking the whole of `src/test/resources`, including
`module/`. T1 graded is `unit/*.t` at top level. The three files supplying
335 of those 345 improvements were all Class-MethodMaker, under `module/`,
outside T1 entirely:

    -136  module/Class-MethodMaker/t/array.t   (180 -> 44)
    -122  module/Class-MethodMaker/t/hash.t    (167 -> 45)
     -77  module/Class-MethodMaker/t/scalar.t  (131 -> 54)

The decision to hold the change was right; the reason given for it was not.
Both are recorded so the next attempt measures the right set from the start.

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
is a real fix regardless of this issue -- `$a[0]{k}` has the same defect and
needs no dereference to reach it -- and it is the point at which this stops
being a dereference fix and becomes brace classification, 01a0ac52's
subject. Two more term-position brace decisions sit behind it.

That one-liner is worth landing on its own, with its own measurement, before
the dereference work resumes: it is independently correct, and separating it
means the next T1 number attributes cleanly to one change rather than two.

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

- [ ] `${$h->{k}}`, `@{[ ... ]}`, `$$x`, `@$x` each have their interior in the tree as children (`go test ./internal/parse/ -run '^TestDereferenceHasInterior$' -count=1 -v | grep -q '^--- PASS: TestDereferenceHasInterior'`)
- [ ] `${name}` stays one Variable, `$$` stays the process id, `${^TAINT}` stays a name (`go test ./internal/parse/ -run '^TestBracedNameIsNotADereference$|^TestPunctuationVariablesStillLex$' -count=1 -v | grep -q '^--- PASS'`)
- [ ] Every dereference form still round-trips (`go test ./internal/parse/ -run '^TestDereferenceRoundTrips$' -count=1 -v | grep -q '^--- PASS: TestDereferenceRoundTrips'`)
- [ ] T1 clean does not fall below 408 and T1 Unknown does not rise above 4,944, measured on `unit/*.t` at top level, 986 files (`go test ./internal/parse/ -run '^TestT1Ratchet$' -count=1 -v | grep -q '^--- PASS: TestT1Ratchet'`)
- [ ] No individual T1 file gains Unknown nodes; a net improvement that hides per-file regressions fails (`go test ./internal/parse/ -run '^TestT1PerFileNoRegression$' -count=1 -v | grep -q '^--- PASS: TestT1PerFileNoRegression'`)
- [ ] The whole suite is green (`go test ./internal/... 2>&1 | grep -qv FAIL`)

Blocked by 01a0ac52.

Note on the ratchet ACs: TestT1Ratchet and TestT1PerFileNoRegression do not
exist yet. They belong to the M1 gate (01a0a70f-3ccc), which owns the parse
ratchet and has none of its eight AC tests written. Until that lands, these
two criteria are unrunnable and this issue cannot be completed -- which is
the correct state, not a defect in the criteria.
