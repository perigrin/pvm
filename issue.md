---
title: "A phaser brace is a block: BEGIN { } leaves the lexer in the wrong state"
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-18T22:28:03.205765477Z
updated: 2026-09-19T01:30:28.234190551Z
sessions:
- start_sha: 1b21e398ccd962c900969acf3c1912a7c3aafa01
  end_sha: 528b9d835526d7a2c3acb41646c44cf6fd145437
  commits: 1
  started_at: 2026-09-19T01:11:11.251966061Z
  ended_at: 2026-09-19T01:30:28.234190551Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T01:11:11.251966061Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-19T01:30:28.234190551Z
observed_paths:
- internal/lexer/brackets.go
- internal/lexer/expect.go
- internal/lexer/expect_test.go
- internal/lexer/intuit_test.go
- internal/lexer/lexer.go
- internal/lexer/scan.go
- internal/parse/testdata/t1.ratchet
---

The phaser half of this LANDED in 01a0ac52. What remains is the narrowing it
was blocking.

## What landed

`internal/lexer/keyword.go` gained a `phaser` table, and `expect.go`'s Word
case returns XBlock for one. A phaser's brace is always a block and never a
hash, so the word alone settles it -- no lookahead, the way `sub NAME` needs
none. perl reaches PREBLOCK for both.

Before it, `BEGIN` was an ordinary bareword: its `{` read as an anonymous
hash, its `}` left XOperator instead of XState, and a bare block after it was
read as a hashref too.

    BEGIN { my $x = 1; }
    { use Foo; }            <- Unknown

Covered by `TestRightParenRegressionIsNamed` for all five phasers.

## What remains

`expect.go`'s CloseBracket case still returns XBlock for ANY closer followed
by `{`. perl's own rule (toke.c yyl_rightparen, 7156) is `)` only.

Narrowing it cost 16 T1 files before the phaser table. It now costs ONE, and
that one is named:

    my @s = $r->@[ 2, 1 ];      postderef.t:12

`@[` lexes as a Variable, so no `[` ever reaches the bracket stack, and its
`]` pops whatever was underneath. The stack is corrupted from there on, and
the broad rule happens to paper over the next brace.

That is a postfix-slice sigil defect rather than a brace one, and 01a0ad52
owns it. This issue is the narrowing that follows.

## Acceptance criteria

- [x] `BEGIN { ... }` followed by a bare block parses, and so do END, CHECK,
      INIT and UNITCHECK
- [x] The brace of `BEGIN {` is OpensBlock on the token, not decided again in
      the parser
- [x] The CloseBracket rule is narrowed to `)`, matching yyl_rightparen, with
      no T1 regression
- [x] The comments at the CloseBracket rule and in
      `TestRightParenRegressionIsNamed` are replaced by what the code then
      does
- [x] A test names the shape, not the implementation: a phaser takes a block


## Done, 2026-09-19 (528b9d83)

    T1 clean       460      (unmoved)
    T1 Unknown   3,323 -> 3,322
    postderef.t      3 -> 2
    the rule     narrowed to `)`, at no cost

The phaser half landed in 79468414 and took the narrowing's cost from 16 T1
files to one. This commit removes the last one.

## The remaining file was a postfix slice

    my @s = $r->@[ 2, 1 ];      postderef.t:12

The sigil scanner took `@[` as a punctuation variable named `[`. No opener
reached the bracket stack, so the matching `]` popped whatever was underneath
and the stack stayed one entry SHORT for the rest of the file. The visible
symptom was a brace misread forty lines later, which is why this survived
three passes over brace classification -- the damage never appeared near its
cause.

    $ perl -MO=Deparse -e 'my $r=[1,2,3]; my @s = $r->@[0,1]; my @t = $r->@*;'
    my(@s) = @$r[0, 1];
    my(@t) = @$r;

`@*` and `%*` stay one token -- there the star IS the whole dereference -- so
the rule is "a sigil before `[` or `{` in XPostDeref", and
TestPostfixSliceSigilStopsAtTheBracket guards both halves plus the stack
staying balanced afterwards. Verified to go red without the fix.

## Two defects, each hiding the other

That is the shape worth remembering from this issue. The broad closer rule
was papering over BOTH the phaser brace and the swallowed bracket, so
narrowing it exposed 16 files at once and looked like one large regression.
Each fix revealed the next, and the third attempt cost nothing.

## closedSubscript is gone

It existed to stop `$a[0]{k}`, `$h{a}{b}` and `${$y}{Keys}` marking their
second brace as opening a block, which the broad rule did. A `)` never closes
a subscript, so the paren test subsumes it and keeping the check would be a
condition that can never fire. TestRightParenLookaheadIsParenOnly passed
before the removal and after.
