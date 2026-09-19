---
title: A lexical hash is not an rv2hv, and the subject cannot tell which it is
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-19T18:15:49.871374091Z
updated: 2026-09-19T23:15:53.031501368Z
---

**The title is wrong and git-zhi cannot retitle.** Read it as "the hash
marker over-reports what perl folds away; hedge the undecidable". The
lexical-versus-package framing was my filing hypothesis and the measurement
below refutes it.

48 WRONG rv2hv sites over the 620-file corpus at 25483d2b. This issue was
filed as a lexical-versus-package scope problem. **It is not**, and the
measurement below is what the work should be planned against rather than the
original diagnosis.

## What the 48 actually are

Every WRONG rv2hv site, dumped with its source line and counted by shape:

    hash SLICE  @h{...}, @$r{...}, ->@{...}    ~20
    delete, usually `delete local`              14
    nested deref %{$h{a}{b}}                     ~8
    a plain package %h                           ~2

Barely any are the `keys(%h)` case this issue was filed for. Examples:

    op/array.t:73       @bar{keys %foo} = values %foo;
    run/locale.t:44     delete local @ENV{'LANGUAGE', 'LANG', ...};
    porting/libperl.t:472  for my $symbol (sort keys %{$symbols{data}{$dtype}})
    op/postfixderef.t:104  $spring2{"foo"}->@* = (1,2,3);

## Scope matters, but not the way this issue assumed

It is not "lexical declarations are invisible to us". Measured on 5.42.0:

    my %h;    keys %h                       0 rv2hv   (padhv)
    our %h;   keys %h                       1 rv2hv
    state %h; keys %h                       1 rv2hv   <-- NOT lexical here
    local %h; keys %h                       2 rv2hv

    my %foo, %bar;  @bar{keys %foo} = ...   0 rv2hv
    our %h, %g;     @g{keys %h}    = ...    2 rv2hv

So `state` behaves like a package hash for this marker even though it is a
lexical declaration, and a slice's op count depends on the scope of the
hashes it slices. A scope stack keyed on `my` -- which is what I proposed --
gets `state` wrong and does nothing for the slice cases where both hashes
are already lexical and we over-report anyway.

## And some are not about scope at all

`@$a{'abc','def'}` emits 0 rv2hv on its own:

    my $r = {a=>1}; my @x = @$r{'a','b'};   0 rv2hv

but op/avhv.t:131 has the same shape inside an `eval { }` block and IS in
the WRONG list, which makes it an op-attribution question rather than a
detection one.

## Why hedging is the plan

`isHashTerm` tests a leaf's TEXT for a leading `%`. Text cannot carry scope,
cannot know which accesses perl's optimiser folds to multideref, and cannot
see the block structure that moves attribution. The marker's own doc records
the asymmetry that makes hedging safe:

> a lexical `%h` is padhv and `$h{a}` folds to multideref, so perl emits
> FEWER of these than the source has hash accesses, never more -- which is
> the direction presence-matching tolerates

A hedge scores WIDER, which is the honest bucket for "I see a hash access
and cannot tell whether perl will emit a dereference for it". It is also the
bucket the four-bucket design exists to reward, and exports.go's three-tier
rule is the existing precedent for refusing rather than guessing.

Tracking scope is NOT deferred, it is **dropped**: the measurement says it
would fix at most the ~2 plain-package sites while getting `state` wrong,
and the other 46 need facts a parser cannot have.

## Scope of the change

Hedge the hash site where the reading is undecidable, commit where it is
not. The line to draw needs its own measurement -- hedging EVERY hash site
would take exact down as well as WRONG, and the exact floor exists to stop
exactly that. Start by hedging the four shapes above and measure both
numbers.

## Acceptance

- [ ] An undecidable hash access hedges rather than committing (`go test ./internal/parse/ -run '^TestUndecidableHashHedges$' -count=1 -v | grep -q '^--- PASS: TestUndecidableHashHedges'`)
- [ ] A decidable one still commits, so exact does not collapse (`go test ./internal/parse/ -run '^TestDecidableHashStillCommits$' -count=1 -v | grep -q '^--- PASS: TestDecidableHashStillCommits'`)
- [ ] WRONG rv2hv falls and the exact floor holds, both measured (`go test ./internal/parseoracle/ -run '^TestGoParserExactFloor$' -count=1 -v | grep -q '^--- PASS: TestGoParserExactFloor'`)

Measured 2026-09-19. Plan: M1. Spec: §7.1.2.
