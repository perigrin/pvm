---
title: A bare /.../ matching against $_ is not reported as a match site
state: in-progress
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-19T18:17:06.945277771Z
updated: 2026-09-19T20:59:15.555455554Z
sessions:
- start_sha: 34b103f60e9548502ff3afe1f4cbb943b3397863
  end_sha: ""
  commits: 0
  started_at: 2026-09-19T20:59:15.555455554Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T20:59:15.555455554Z
---

A pattern with no `=~` matches against `$_`, and perl emits the match op for
it:

    $ cat p.pl
    $_ = "abc";
    my @w = /(\w+)/g;
    $ perl -MO=Concise,-exec p.pl | grep -c match
    1

`Sites` reports a match only from two places (sites.go:194-238): a Binary
`=~`/`!~` whose right operand is a pattern, and a Term that isBarePattern.
Neither fires for a pattern standing alone as a statement's expression,
because the tree holds it as a Term reached through a path the Term case
does not take -- measure which before changing anything.

53 of the 73 WRONG match sites are this shape, and re/pat.t alone holds 35:

    230  my @words = /(\w+)/g;
    236  while (/\w+/g) {
    270  $t1++ if /$pat1/o;
    301  ok( /abc/g && $` eq "", $message);
    325  /\Gc/gc;
    926  ok(/b(?{$foo = $_; $bar = pos})c/g, $message);

Every one is a bare pattern in a boolean, list or statement position.

## What must not regress

s/// is subst and tr/// is trans, neither of which is this marker
(sites.go:307-310). isBarePattern exists to exclude them by spelling, and
whatever fixes this must keep that exclusion -- `split /,/` compiles to a
split op with NO match op, which §7.1.2 records, so a rule that reported
every regex literal would over-report.

## The remainder is span attribution

    re/reg_pmod.t:43   x4    "my $ok ="
    run/switches.t:105 x4    "ok("
    op/exec.t:114      x3    one long condition
    op/numconvert.t:226 x2

Same first-line-of-a-multi-line-statement shape as 01a0badc, and likely
fixed by it.

## Acceptance

- [ ] A bare `/.../` reports a match site (`go test ./internal/parse/ -run '^TestBarePatternIsAMatchSite$' -count=1 -v | grep -q '^--- PASS: TestBarePatternIsAMatchSite'`)
- [ ] s///, tr/// and split's pattern still do not (`go test ./internal/parse/ -run '^TestBarePatternIsAMatchSite$' -count=1 -v | grep -q '^--- PASS: TestBarePatternIsAMatchSite'`)
- [ ] WRONG match sites fall, measured (`go test ./internal/parseoracle/ -run '^TestGoParserWrongIsZero$' -count=1 -v | grep -q '^--- PASS: TestGoParserWrongIsZero'`)

Found by 01a0ba79's ranked table. Plan: M1. Spec: §7.1.2.
