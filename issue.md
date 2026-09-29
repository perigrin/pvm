---
title: "`Foo::` parses as a call: canon writes `Foo::()`"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T20:16:20.208540338Z
updated: 2026-09-29T20:16:21.731745433Z
sessions:
- start_sha: 3d5c9d7a176699f0e6a42e64ae83fb4f173e1e5c
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T20:16:21.731745433Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T20:16:21.731745433Z
---

A bareword ending in `::` parses as an unresolved call, so canon writes `bless [], Foo::` as `bless([] , Foo::())` and `Foo::->new` as `Foo::() -> new`. That is a wrong tree which round-trips. toke.c:8072 strips the `::` and makes the rest a constant without looking for a sub, so there is no callee to resolve.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'f(Foo::); my $x = Foo::;'
    f('Foo');
    my $x = 'Foo';

Found in perl.git t/op/gv.t:984, `bless [], _121242::`.

## Acceptance criteria

- [ ] `Foo::` is a Term and canon does not call it (`go test ./internal/parse/ -run '^TestClassNameBareword$' -count=1 -v | grep -q '^--- PASS: TestClassNameBareword'`)
