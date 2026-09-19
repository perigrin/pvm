---
title: "A phaser brace is a block: BEGIN { } leaves the lexer in the wrong state"
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-18T22:28:03.205765477Z
updated: 2026-09-19T01:11:11.251966061Z
sessions:
- start_sha: 1b21e398ccd962c900969acf3c1912a7c3aafa01
  end_sha: ""
  commits: 0
  started_at: 2026-09-19T01:11:11.251966061Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T01:11:11.251966061Z
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
- [ ] The CloseBracket rule is narrowed to `)`, matching yyl_rightparen, with
      no T1 regression
- [ ] The comments at the CloseBracket rule and in
      `TestRightParenRegressionIsNamed` are replaced by what the code then
      does
- [x] A test names the shape, not the implementation: a phaser takes a block
