---
title: "A phaser brace is a block: BEGIN { } leaves the lexer in the wrong state"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-18T22:28:03.205765477Z
updated: 2026-09-18T22:28:34.201042644Z
---

The lexer has no phaser table, so `BEGIN` is an ordinary bareword. It leaves
XTerm, its `{` is classified as an anonymous hash rather than a block, and its
`}` therefore leaves XOperator instead of XState. The block after it is then
read as a hashref too:

    BEGIN { my $x = 1; }
    { use Foo; }            <- Unknown

Measured: the second statement is Unknown, while every other preceding form
works.

    sub f { 1 }      then a bare block    parses
    my %h = (a=>1);  then a bare block    parses
    if (1) { 2 }     then a bare block    parses
    BEGIN { ... }    then a bare block    UNKNOWN

`sub NAME {` is already right, through `afterDeclName` in
`internal/lexer/expect.go` -- perl's PREBLOCK at toke.c:6636. The five phasers
need the same treatment and for the same reason.

## It is currently propped up by another rule

`expect.go`'s CloseBracket case returns XBlock for ANY closer followed by `{`,
not just `)`. perl's own rule (toke.c yyl_rightparen, 7156) is `)` only.
Narrowing ours to match regresses 16 T1 files, and this defect is why: the
broad rule is what rescues the bare block after a `BEGIN` whose braces were
misclassified. One defect propping up another.

So this issue has two halves and the second is the point:

1. Classify the phaser brace as a block.
2. THEN narrow the CloseBracket rule to `)` and confirm the 16 files hold.

Measured at 01a0ac52 while working the intuit_curly lookahead, and named in a
comment at the CloseBracket rule so the next reader does not narrow it and
wonder what broke.

## Acceptance criteria

- [ ] `BEGIN { ... }` followed by a bare block parses, and so do END, CHECK,
      INIT and UNITCHECK
- [ ] The brace of `BEGIN {` is OpensBlock on the token, not decided again in
      the parser
- [ ] The CloseBracket rule is narrowed to `)`, matching yyl_rightparen, with
      no T1 regression
- [ ] The comment at the CloseBracket rule saying this is tracked is replaced
      by what the code now does
- [ ] A test names the shape, not the implementation: `BEGIN`-then-bare-block
      parses because a phaser takes a block, not because of an expect state
