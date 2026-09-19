---
title: A brace before -> is a hashref, and two anonhash cases the tree misreads
state: in-progress
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0badc-a093-7e09-bdfb-002f4a73759f
created: 2026-09-19T18:17:07.068073736Z
updated: 2026-09-19T21:12:55.13834552Z
sessions:
- start_sha: 25483d2b5bfff159cfc023dec9b050ce0c57c4b3
  end_sha: 25483d2b5bfff159cfc023dec9b050ce0c57c4b3
  commits: 0
  started_at: 2026-09-19T21:11:22.140962055Z
  ended_at: 2026-09-19T21:12:55.13834552Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T21:11:22.140962055Z
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-19T21:12:55.13834552Z
---

35 WRONG anonhash sites, and most are span attribution -- but three are real
and each is the "is `{` a block or a hashref" question §7.1.2 names.

## The real cases

**`grep {a=>$_}->{a}`** (op/grep.t:105). The `->` makes the brace a HASHREF,
not grep's block. Measured:

    $ perl -MO=Concise,-exec -e 'my @res = grep {a=>$_}->{a}, ("x");' | grep -c anonhash
    1

**`use overload '%{}' => sub { +{} }`** (op/coreamp.t:23) and
**`'${}' => sub { \my $x }`**. `+{}` is the disambiguating unary plus that
forces a hashref reading, and perl emits NO anonhash for it:

    $ perl -MO=Concise,-exec -e 'my $x = sub { +{} };' | grep -c anonhash
    0

So the subject reporting one there is over-reporting, the opposite direction
from the grep case. Both are the same ambiguity decided differently.

**`\state %y = {1,2}`** (op/lvref.t:370) and **`bless {`** (op/sort.t:944,
op/switch.t:1228) need measuring individually.

## The rest is span attribution

    op/sprintf2.t:347           x8   "my @tests = ("
    mro/package_aliases.t       x6   "for("
    mro/package_aliases_utf8.t  x6   "for("
    op/bop.t:693                x3   "my %op_pairs = ("
    op/pos.t:138                x2   "my %expected = ("

A hash literal assigned to a hash emits no anonhash at all:

    $ perl -MO=Concise,-exec -e 'my %op_pairs = ( a => 1 );' | grep -c anonhash
    0

so these are the first line of a multi-line statement whose ops belong
elsewhere -- 01a0badc's shape. Fix that first and re-measure before touching
this; the residue is what this issue is actually about.

## Acceptance

- [ ] `grep {a=>$_}->{a}` reports an anonhash site (`go test ./internal/parse/ -run '^TestBraceBeforeArrowIsAHashref$' -count=1 -v | grep -q '^--- PASS: TestBraceBeforeArrowIsAHashref'`)
- [ ] `+{}` does NOT report one, matching perl (`go test ./internal/parse/ -run '^TestBraceBeforeArrowIsAHashref$' -count=1 -v | grep -q '^--- PASS: TestBraceBeforeArrowIsAHashref'`)
- [ ] The remaining cases are measured and recorded, blocked on 01a0badc (`go test ./internal/parseoracle/ -run '^TestGoParserWrongIsZero$' -count=1 -v | grep -q '^--- PASS: TestGoParserWrongIsZero'`)

Found by 01a0ba79's ranked table. Plan: M1. Spec: §7.1.2, §4.9.2.


## Re-measured after 01a0badc, 2026-09-19: 35 sites -> 2

This issue was filed at 35 anonhash sites and blocked on the span fix. That
fix landed (09e38fa9) and took it to 2. The span attribution WAS almost all
of it, as this issue predicted.

The two that remain:

    op/coreamp.t:23   use overload '%{}' => sub { +{} }
    op/undef.t:200    my $h= { baz => 1 }; my @k= keys %{($h=undef)||{}};

## What I established

**coreamp.t is a genuine over-report.** `+{}` is the disambiguating unary
plus, and perl emits NO anonhash for it:

    $ perl -MO=Concise,-exec -e 'my $x = sub { +{} };' | grep -c anonhash
    0

We report one. That is a claim perl contradicts, and it is the whole of
this site.

**undef.t is not what I assumed and I did not finish it.** I expected perl
to fold `{ baz => 1 }`, but measuring says otherwise:

    my $h = { baz => 1 };          1 anonhash
    my $h = {};                    0 anonhash, 1 emptyavhv/ANONHASH
    my $h = {}; my $j = {a=>1};    1 anonhash  (the empty one is emptyavhv)

and parse_facts.pl:142 folds emptyavhv-flagged-ANONHASH back into the
marker, so perl's total for that line should be 2 -- which is what we
report. The count agrees and the site is still WRONG, so the disagreement is
about WHICH STATEMENT owns one of them, not how many exist.

That needs the oracle's own line list for the file, which I did not run.
Paused there rather than guessing: two sites does not justify more
speculation, and a wrong diagnosis committed to an issue is worse than an
open question.

## Recommendation

Worth doing as a pair with whoever next touches anonhash, or folding into
the "hedge versus detect" decision 01a0bae1-b76f raises -- `+{}` may be
better answered by declining than by adding a special case for the unary
plus.
