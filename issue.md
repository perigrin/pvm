---
title: "One statement reported as many WRONG sites: multi-line span attribution"
state: pending
urgency: normal
milestone: m1-parse-the-core
blocks:
- 01a0bae2-e4fe-74e9-82cb-562c3fa8adf6
created: 2026-09-19T18:10:16.339474961Z
updated: 2026-09-19T18:17:31.367746494Z
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

## Acceptance

- [ ] The repeated-line sites resolve to one statement each, measured (`go test ./internal/parseoracle/ -run '^TestStatementSpanCoversItsOps$' -count=1 -v | grep -q '^--- PASS: TestStatementSpanCoversItsOps'`)
- [ ] The cause is recorded: span not covering, or duplicate sites (`go test ./internal/parse/ -run '^TestSitesSpanMultilineStatement$' -count=1 -v | grep -q '^--- PASS: TestSitesSpanMultilineStatement'`)
- [ ] WRONG srefgen sites fall, measured over the corpus (`go test ./internal/parseoracle/ -run '^TestGoParserWrongIsZero$' -count=1 -v | grep -q '^--- PASS: TestGoParserWrongIsZero'`)

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
