---
title: "One statement reported as many WRONG sites: multi-line span attribution"
state: done
urgency: normal
milestone: m1-parse-the-core
blocks:
- 01a0bae2-e4fe-74e9-82cb-562c3fa8adf6
created: 2026-09-19T18:10:16.339474961Z
updated: 2026-10-09T17:21:18.998278373Z
sessions:
- start_sha: afdd06c9562e8c8efaa4fb31bad3824739ffb0ba
  end_sha: 09e38fa97c039422168735f3da9b607b14677d80
  commits: 1
  started_at: 2026-09-19T20:25:32.848375655Z
  ended_at: 2026-09-19T20:38:21.082315925Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T20:25:32.848375655Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-19T20:38:21.082315925Z
observed_paths:
- internal/parseoracle/compare_facts.go
- internal/parseoracle/span_test.go
---

The same source line is reported WRONG many times over, which is one
statement whose span does not line up rather than many defects. Measured over
the 620-file corpus at afdd06c9, srefgen marker:

    op/ref.t:214      "for ("          x15
    op/pack.t:1435    "my @tests = ("  x6
    op/pack.t:1449    "push @tests,"   x6
    re/reg_mesg.t:874 "foreach my $ref ("  x3
    op/bop.t:497      "for ("          x1

31 of the 88 WRONG srefgen sites, and every one of those lines is the FIRST
line of a long multi-line statement.

## Why this is not a parser capability gap

compare_facts.go:106-117 states the rule: perl attributes every op to the
nearest preceding nextstate, and the subject reports each site with its
statement's SPAN (Line..EndLine). A perl site belongs to the innermost
subject statement whose span contains its line.

So either the subject is reporting a span that does not cover the lines perl
used, or it reports the same span repeatedly where one statement is meant.
Both are defects in `internal/parse/sites.go`, not in what the parser
understands. This is M1's, because M1 owns the subject contract.

## What to look at first

`op/ref.t:214` has fifteen sites on one line. Either:

  a. perl attributes all 15 ops to line 214 and the subject's statement span
     for that statement does not contain 214; or
  b. the subject emits 15 sites for one statement where one is meant.

The distinction matters: (a) is a span bug, (b) is a hedging bug. Dump the
subject's sites for that file and compare against the oracle's line list
before changing anything.

There is prior art for the shape of the answer: op/qr.t line 112 is already
handled (persite_test.go:26-38), where a block-bearing term resets copline
and perl's nextstate records the LAST line of the statement rather than the
first. A `for (` head spanning many lines may be the same phenomenon at the
other end.

## Not to be confused with

These are NOT prototype cases. The lines carry no call to a
prototyped sub -- `for (` and `push @tests,` are ordinary statements. M4
owns prototypes; this is separate and lands in M1.

## Acceptance criteria

- [x] The repeated-line sites resolve to one statement each, measured (`go test ./internal/parseoracle/ -run '^TestStatementSpanCoversItsOps$' -count=1 -v | grep -q '^--- PASS: TestStatementSpanCoversItsOps'`)
- [x] The cause is recorded: span not covering, or duplicate sites (`go test ./internal/parse/ -run '^TestSitesSpanMultilineStatement$' -count=1 -v | grep -q '^--- PASS: TestSitesSpanMultilineStatement'`)
- [x] WRONG srefgen sites fall, measured over the corpus (`go test ./internal/parseoracle/ -run '^TestGoParserWrongIsZero$' -count=1 -v | grep -q '^--- PASS: TestGoParserWrongIsZero'`)

Found by 01a0ba79's ranked table. Plan: M1. Spec: §7.2.


## op/lvref.t belongs here too, and it is not what it looks like

Five more sites, all reported at op/lvref.t:432:

    \($bb, @cc, %dd, &ee, $_==1 ? $ff : @ff, ...) =
        (\$BB, \@CC, \%DD, \&EE, \$FF, \@GG, \1, \2, \3);

I first read this as `\` DISTRIBUTING over a parenthesised list -- one
backslash, many references. Measured, and it is not:

    $ perl -MO=Concise,-exec -e 'my ($a,@b,%c); my @r = \($a, @b, %c);' | grep -c srefgen
    0
    $ perl -MO=Concise,-exec -e 'my ($a,@b); \($a, @b) = (\my $x, \my @y);' | grep -c srefgen
    0

The lvalue `\(...)` emits NO srefgen at all. The five come from the RIGHT
HAND SIDE, line 433, which carries five explicit backslashes:

    $ perl -MO=Concise,-exec -e 'my ($BB,@CC); my @x = (\$BB, \@CC);' | grep -c srefgen
    2

So this is the same defect as the rest of this issue -- ops on line 433
attributed to the statement beginning at 432 -- and not a parse rule. 36 of
the 88 WRONG srefgen sites are span attribution once lvref.t is counted.


## Done, 2026-09-19 (09e38fa9)

    corpus WRONG files   72 -> 51
    corpus WRONG sites  292 -> 163
    corpus exact        387 -> 401
    full suite green, 0 FAIL across the repo

    srefgen   88 -> 38      anonhash  35 ->  2
    match     73 -> 50      readline  25 -> 25
    rv2hv     71 -> 48

129 of the 292 sites were this one defect, spread across every marker.

## Both hypotheses in this issue were wrong

I proposed (a) the span not covering perl's line, or (b) the subject
emitting duplicate sites. Dumping op/ref.t showed neither:

    kind=reference line=214 end=215 took_ref=true
    kind=reference line=214 end=216 took_ref=true
    ... fifteen, one per backslash, every Line matching perl's

The subject was right. The defect was one layer down in `groupByStatement`,
which this issue did not name: it keyed a statement record on [start, end],
and since `Sites` deliberately ends each site at its OWN node
(sites.go:83-101), fifteen references in one statement became fifteen
records of one reference each. `innermost` picks one winner, its reference
is spent on perl's first site, and the rest score WRONG.

Keyed on the start line instead, with the end widening to the furthest node.

## AC 2's package was wrong, for the same reason

It named `./internal/parse/`, because I had assumed the cause was in
`Sites`. TestSitesSpanMultilineStatement lives in `parseoracle`, where the
code it tests is. Verified there rather than moving a test to the wrong
package to satisfy a stale command.

## Two guards, because a merge over-reaches easily

TestNestedStatementStillOwnsItsLines -- a `\` inside a sub body stays its
own record. That is the case the per-node end exists for
(t/class/destruct.t:22), and merging by start alone does not touch it: the
body's statement has its own start line.

TestOneReferenceStillAccountsForOne -- the merge changes which record a site
lands in, never how many a record holds.

persite_test.go's attribution tests all still pass, which is the real check.

## What this leaves visible

    readline  25   all the builtin           01a0bae1-b741
    match     50   bare /.../ against $_     01a0bae2-e481
    rv2hv     48   lexical vs package hash   01a0bae1-b76f
    srefgen   38   18 are pinned M4 work
    anonhash   2   almost all of it was this bug

01a0bae2-e4fe (anonhash) was blocked on this and should be re-measured
before it is worked: its 35 sites are now 2.
