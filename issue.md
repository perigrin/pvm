---
title: A brace before -> is a hashref, and two anonhash cases the tree misreads
state: in-progress
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0badc-a093-7e09-bdfb-002f4a73759f
- 01a0b835-1270-7b32-9f0a-18fe7c555bae
created: 2026-09-19T18:17:07.068073736Z
updated: 2026-10-09T09:12:17.979193617Z
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
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T03:26:34.37807385Z
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

### Re-measured (2026-10-09, at d2f26fe3)

The oracle now measures internal/parse (01a0b835), and this issue's cases have moved:

- op/grep.t is `exact`: `grep {a=>$_}->{a}` already reads as perl does (one anonhash: measured with Concise on 5.42).
- The `+{}` diagnosis was wrong. perl DOES emit an anonymous hash for `+{}` (`my $x = sub { +{} };` measures one anonhash/emptyavhv site), and the op/coreamp.t:23 site it cited was the oracle counting a cloned anon sub twice -- fixed by 01a11f08-6fee. The criterion that `+{}` reports no site contradicts perl and is replaced.
- `TestGoParserWrongIsZero` only runs with `PARSEORACLE_SHIM` set, so as a criterion verify sees it skip; the corpus ratchet's pin (`TestCorpusBaselineIsIntact`) is the WRONG gate now, and 01a0a8cd removes it.

What is left is to hold the two brace readings with a test, so they cannot regress.

## Steps

- [ ] Write `TestBraceBeforeArrowIsAHashref`: `grep {a=>$_}->{a}` reports one anonhash site and `+{}` reports one, each as perl 5.42 does (measure with Concise first); if it passes at once, show it is not vacuous by breaking the reading and reverting; `make test`; commit

## Acceptance criteria

- [ ] a brace before `->`, and `+{}`, each report the anonymous hash perl compiles (`go test ./internal/parse/ -run '^TestBraceBeforeArrowIsAHashref$' -count=1 -v | grep -q '^--- PASS: TestBraceBeforeArrowIsAHashref'`)
