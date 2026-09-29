---
title: "`Foo::` parses as a call: canon writes `Foo::()`"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T20:16:20.208540338Z
updated: 2026-09-29T20:16:20.208540338Z
---

A bareword ending in `::` parses as an unresolved call, so canon writes `bless [], Foo::` as `bless([] , Foo::())` and `Foo::->new` as `Foo::() -> new`. That is a wrong tree which round-trips. toke.c:8072 strips the `::` and makes the rest a constant without looking for a sub, so there is no callee to resolve.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'f(Foo::); my $x = Foo::;'
    f('Foo');
    my $x = 'Foo';

Found in perl.git t/op/gv.t:984, `bless [], _121242::`.

## Acceptance criteria

- [ ] `Foo::` is a Term and canon does not call it (`go test ./internal/parse/ -run '^TestClassNameBareword$' -count=1 -v | grep -q '^--- PASS: TestClassNameBareword'`)
