---
title: A lexical hash is not an rv2hv, and the subject cannot tell which it is
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-19T18:15:49.871374091Z
updated: 2026-09-19T18:15:49.871374091Z
---

Perl emits rv2hv for a PACKAGE hash and not for a lexical one. Measured:

    my %h = (a=>1);   my $n = scalar(keys(%h));     0 rv2hv
    our %h; local %h; my $n = scalar(keys(%h));     2 rv2hv

A lexical `%h` is padhv; only a package hash needs the symbol-table
dereference the marker names. The subject reports a hash site for both,
because `isHashTerm` tests the leaf's TEXT for a leading `%` and text cannot
carry scope.

This is the single largest rv2hv cause. op/local.t is package hashes
throughout and holds 14 of the 71 WRONG sites:

    206  is(scalar(keys(%h)), 3);
    214  is(scalar(keys(%h)), 3);
    ... nine more of the same shape
    811  local @Flibbert::{<bar baz>};
    841  delete local $Grompits::{foo};

## Why this is not simply a fix

Deciding it needs scope tracking: which `%h` is in scope at this statement,
and was it declared `my` or reached through the symbol table. That is more
than the subject contract asks for today, and §7.1.2's own note says perl
emits FEWER of these than the source has hash accesses -- "which is the
direction presence-matching tolerates".

So there are two honest routes and the choice needs measuring:

1. **Track lexical declarations** well enough to answer `my %h` versus
   package `%h`. A `my` in the same file is cheap; `our` and imports are
   not.
2. **Hedge the hash site** where scope is undecidable, scoring wider rather
   than WRONG. The existing three-tier rule in exports.go is the precedent,
   and wider is honest.

Route 2 is available now and correct regardless; route 1 is the one that
raises the exact rate.

## A second, separate cause in the same marker

`@{$h{k}}` -- an ARRAY dereference of a hash element -- emits no rv2hv,
because perl folds it to multideref:

    @{$spring2{"foo"}} = (1,2,3);        0 rv2hv
    %{$symbols{data}{x}}                 2 rv2hv

op/ref.t:113, japh/abigail.t:199 and porting/libperl.t:212 are this shape.
The subject reports a hash site for the `{k}` subscript; perl reports none.
Folding is §7.1.2's stated tolerance direction, so these may be a hedging
question rather than a detection one -- measure before deciding.

## Acceptance

- [ ] The lexical-versus-package distinction is decided or hedged, measured either way (`go test ./internal/parse/ -run '^TestLexicalHashIsNotAPackageHash$' -count=1 -v | grep -q '^--- PASS: TestLexicalHashIsNotAPackageHash'`)
- [ ] `@{$h{k}}` is measured and its treatment recorded (`go test ./internal/parse/ -run '^TestArrayDerefOfHashElement$' -count=1 -v | grep -q '^--- PASS: TestArrayDerefOfHashElement'`)
- [ ] WRONG rv2hv sites fall, measured (`go test ./internal/parseoracle/ -run '^TestGoParserWrongIsZero$' -count=1 -v | grep -q '^--- PASS: TestGoParserWrongIsZero'`)

Found by 01a0ba79's ranked table. Plan: M1 or M4 depending on the route.
Spec: §7.1.2.
