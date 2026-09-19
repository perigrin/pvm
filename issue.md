---
title: A bare /.../ matching against $_ is not reported as a match site
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-19T18:17:06.945277771Z
updated: 2026-09-19T21:11:02.928119055Z
sessions:
- start_sha: 34b103f60e9548502ff3afe1f4cbb943b3397863
  end_sha: 25483d2b5bfff159cfc023dec9b050ce0c57c4b3
  commits: 1
  started_at: 2026-09-19T20:59:15.555455554Z
  ended_at: 2026-09-19T21:11:02.928119055Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T20:59:15.555455554Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-19T21:11:02.928119055Z
observed_paths:
- internal/parse/barepattern_test.go
- internal/parse/sites.go
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

- [x] A bare `/.../` reports a match site (`go test ./internal/parse/ -run '^TestBarePatternIsAMatchSite$' -count=1 -v | grep -q '^--- PASS: TestBarePatternIsAMatchSite'`)
- [x] s///, tr/// and split's pattern still do not (`go test ./internal/parse/ -run '^TestBarePatternIsAMatchSite$' -count=1 -v | grep -q '^--- PASS: TestBarePatternIsAMatchSite'`)
- [x] WRONG match sites fall, measured (`go test ./internal/parseoracle/ -run '^TestGoParserWrongIsZero$' -count=1 -v | grep -q '^--- PASS: TestGoParserWrongIsZero'`)

Found by 01a0ba79's ranked table. Plan: M1. Spec: §7.1.2.


## Done, 2026-09-19 (25483d2b)

    corpus match sites   50 -> 4
    corpus WRONG sites  142 -> 96, files 45 -> 38
    corpus exact        403 -> 405
    no other marker moved
    full suite green, 0 FAIL across the repo

## The issue's diagnosis was wrong, and the cause was one line

I filed this expecting the pattern to be reached "through a path the Term
case does not take". Dumping the trees showed the Term right there with the
correct text -- `/(\w+)/g` -- and the predicate simply rejecting it:

    case len(text) >= 2 && text[0] == '/' && text[len(text)-1] == '/':

The text must END in `/`, so `/abc/` was reported and `/abc/g` was not: it
ends in the flag letter. Every pattern in re/pat.t carries a flag, which is
why all 35 of its sites scored WRONG.

The closing delimiter is still checked rather than trusting the leading `/`
alone -- the lexer emits a Quote token only where a pattern is possible, but
a leaf that begins with `/` and never closes is not one.

## Both directions pinned

s/// is subst, tr/// and y/// are trans, qr// compiles without matching --
three different ops, none of them this marker -- each tested WITH and
WITHOUT flags, since the flags are what this change loosened. Two division
rows too (`$a / $b`, `6/2/3`): a rule that accepts a trailing character is
exactly the kind that starts reading arithmetic as a regex.

## The remaining 4 are not what this issue predicted

Every file the old table named is gone. The four now are:

    io/eintr_print.t:22  skip_all("does not match platform whitelist")
    op/dump.t:25         skip_all("no point in dumping on $^O")
    op/groups.t:27       skip_all("No 'id' or 'groups'") if
    io/fs.t:38           my $accurate_timestamps =

Three are `skip_all(...)`, where perl matches a pattern somewhere this
parser does not see one -- most likely inside test.pl's own code reached
through the call, which is a different question from the source text in
front of us. The fourth is a multi-line assignment.

Listed rather than diagnosed: four sites is below the bar for guessing, and
whoever picks this up has the lines.
