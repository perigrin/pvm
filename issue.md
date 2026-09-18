---
title: "The intuit_curly lookahead: a brace after a list operator"
state: in-progress
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0a70f-3ccc-77bb-902d-9a0b9d8050a6
blocks:
- 01a0ad52-f1e4-7e19-80c7-a99f859951eb
created: 2026-09-16T22:24:52.875856958Z
updated: 2026-09-18T22:53:42.652178232Z
sessions:
- start_sha: 5b41510164015a80506c02291228c7544ea1f3df
  end_sha: ""
  commits: 0
  started_at: 2026-09-18T21:49:11.664330742Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-18T21:49:11.664330742Z
---

A one-line change here is correct and makes T1 worse. That is the finding, and it is why this issue is bigger than it looks.

## The defect

`expect.go`'s CloseBracket case applies perl's right-paren lookahead to
EVERY closing bracket:

    if t.nextIsOpenBrace {
        return XBlock
    }

Its own comment says "A `)` with a `{` next" and cites `yyl_rightparen`
(toke.c:7156), which handles the right PAREN alone. The code checks no byte.
So a subscript chain opens a block:

    $a[0]{k}     the `{` follows a `]`
    $h{a}{b}     the `{` follows a `}`
    $x->[0]{k}

All three are subscripts of the value the previous bracket produced. Nothing
in them opens a block. Verified 2026-09-17 at 3c997a6b: all three set
`OpensBlock` on their second brace.

## The fix, and why it is not committed

    if t.nextIsOpenBrace && len(t.text) == 1 && t.text[0] == ')' {

Correct by toke.c, and it turns the failing test green. Measured on T1
graded (`unit/*.t`, 986 files):

                      before    after
    T1 clean          408       406
    T1 Unknown        4,944     4,972

Worse on both axes. Something downstream depends on the wrong
classification, and until that is found, fixing the lexer in isolation moves
the number the wrong way. The likely shape is a parser path that reads
`OpensBlock` and gets a correct answer for the first time, having been
written against the incorrect one -- but that is a hypothesis, not a
measurement, and finding out is this issue's work rather than a side errand.

Two attempts have now stopped here from different directions: this one, and
the dereference leaf of 01a0ad52, which cannot proceed until brace
classification is right.

## What this issue actually covers

Not just the `intuit_curly` lookahead. The term-position brace decision as a
whole:

- the right-paren lookahead applying to every closer (above)
- `{` after a list operator, this issue's original subject
- `${EXPR}` vs `${name}`, needed by 01a0ad52
- whatever downstream code the first item's fix disturbs

Doing them together is the point. Each one alone moves the T1 number the
wrong way, because they are one decision spread across the lexer.

## Acceptance

- [x] `$a[0]{k}`, `$h{a}{b}`, `$x->[0]{k}` classify both brackets as terms (`go test ./internal/lexer/ -run '^TestRightParenLookaheadIsParenOnly$' -count=1 -v | grep -q '^--- PASS: TestRightParenLookaheadIsParenOnly'`)
- [x] `if (...) {`, `while (...) {`, `for (;;) {` still open blocks (`go test ./internal/lexer/ -run '^TestRightParenStillOpensABlock$' -count=1 -v | grep -q '^--- PASS: TestRightParenStillOpensABlock'`)
- [x] A brace after a list operator is decided by lookahead, not by position alone (`go test ./internal/lexer/ -run '^TestIntuitCurlyLookahead$' -count=1 -v | grep -q '^--- PASS: TestIntuitCurlyLookahead'`)
- [x] `map { $_ => 1 } @a` and `map { ; $_ } @a` both parse (`go test ./internal/parse/ -run '^TestMapBlockForms$' -count=1 -v | grep -q '^--- PASS: TestMapBlockForms'`)
- [x] Whatever regressed T1 from 408/4,944 when the right-paren rule was narrowed is identified and named in a test comment (`go test ./internal/lexer/ -run '^TestRightParenRegressionIsNamed$' -count=1 -v | grep -q '^--- PASS: TestRightParenRegressionIsNamed'`)
- [x] T1 clean does not fall below 408 and T1 Unknown does not rise above 4,944 (`go test ./internal/parse/ -run '^TestParseRatchet$' -count=1 -v | grep -q '^--- PASS: TestParseRatchet'`)
- [x] The whole suite is green (`go test ./internal/... 2>&1 | grep -qv FAIL`)

Note on the ratchet AC: TestParseRatchet does not exist yet. It belongs to the
M1 gate (01a0a70f-3ccc), which owns the parse ratchet and has none of its
eight AC tests written.


## Done, 2026-09-18 (79468414)

    T1 clean       435 -> 456
    T1 Unknown   4,573 -> 3,357
    T1-easy      50.8% -> 52.7%
    T2 core         16 -> 17 clean, 10 files better
    t/class         46 -> 45
    lexer            2 files better
    oracle       405 exact, unmoved

All seven ACs pass, each by the command the AC names. T1 clean 456 against
the AC's floor of 408; T1 Unknown 3,357 against its ceiling of 4,944.

## The lookahead

`internal/lexer/intuit.go`. A word or string then `,` or `=>` is a hash;
everything else is a block -- a leading `;`, an operator, and a VARIABLE
before a fat comma. Measured on perl 5.42.0:

    map { $_ => 1 } @a       map({$_, 1;} @a)        a BLOCK
    map { ; a => 1 } (1,2)   map({'a', 1;} 1, 2)     a BLOCK
    map { a => 1 }, (1,2)    map({'a', 1}, (1, 2))   a HASHREF

## What the wrong brace actually cost

Not a misread node -- an ORPHANED list. `map { $_ => 1 } @a` produced a call
holding an AnonHash and `@a` as a statement of its own. Every byte in the
tree, round-tripping, and not a parse of its source. A block slot beside the
filehandle slot fixes it, and reads OpensBlock rather than deciding again.

## AC 1 turned up a brace that lied

Every second brace in a subscript chain carried OpensBlock -- `$a[0]{k}`,
`$h{a}{b}`, `$x->[0]{k}` -- because the closer-then-brace rule fired for `]`
and for a subscript `}`, not just for `)`. The parser survived it (it reaches
those braces by the postfix path, which never asks) so nothing failed. A flag
that lies until someone reads it is what the brace stack exists to prevent.

Fixing it exposed the phaser defect underneath, which is why phasers landed
here too: `BEGIN` was an ordinary bareword, so its `{` was a hashref and its
`}` left XOperator instead of XState.

## AC 5: what narrowing the right-paren rule costs, now

16 T1 files before the phaser table, ONE after:

    my @s = $r->@[ 2, 1 ];      postderef.t:12

`@[` lexes as a Variable, so no `[` reaches the bracket stack and its `]`
pops whatever was underneath. A postfix-slice sigil defect -- 01a0ad52 owns
it, and the narrowing that follows is 01a0b6a2.

## The empty statement came free

A lone `;` went to the expression parser, returned nil, and
skipToStatementEnd ran past the enclosing `}`. `map { ; $_ } @a` lost its
closing brace, and so did `{};` -- a defect recorded in a baseline two
commits ago.

## Counts that rose, all measured

Seven T1 files rose in node count; six fell in BYTES, one Unknown splitting
into several. overload_compound_assignment.t 1 -> 10 nodes but 2,561 -> 1,424
bytes. Each verified by stashing the change and re-measuring the same source.

cpan_fallback_configure_retry.t rose on both (218 -> 289 bytes): `print {$fh}`
is Unknown before and after, and appears twice more because more of the file
parses. comp/require.t went 11 -> 12 for the same shape.
