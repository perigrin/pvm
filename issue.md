---
title: "A declaring condition loses its parens in canon: `if (my $x = f())` becomes `if my $x = f()`"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T14:20:33.981610508Z
updated: 2026-09-29T14:31:31.303590597Z
sessions:
- start_sha: 580b9d981afc942c4fa785b856dca63174928fdb
  end_sha: a2b774c555f0d3426b100328495938681de26fe7
  commits: 1
  started_at: 2026-09-29T14:20:35.4974998Z
  ended_at: 2026-09-29T14:31:31.303590597Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T14:20:35.4974998Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T14:31:31.303590597Z
observed_paths:
- internal/parse/canon.go
- internal/parse/condeclcanon_test.go
- internal/parse/control.go
- internal/parse/parse.go
- internal/parse/testdata/t1canon.ratchet
- internal/parse/testdata/t1fidelity.ratchet
---

Canon takes ANY Declaration child of a Conditional or Loop for a foreach loop variable and writes it without parentheses, so a condition that declares loses them, as does a foreach list whose one element is an anonymous sub -- at Unknown=0:

    while (my $l = <F>) { 1 }     canon: while my $l = <F> {1;}
    if (my $x = f()) { 1 }        canon: if my $x = f() {1;}
    for my $y (sub {}) { 6 }      canon: for my $y sub {} {6;}

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'while (my $l = <F>) { 1 } if (my $x = f()) { 1 } for my $y (sub {}) { 6 }'
    while (defined(my $l = readline F)) {
    if (my $x = f()) {
    foreach my $y (sub {

The parser already marks a BARE loop variable LoopVar; the declared one should carry the flag too, and canon should read the flag rather than the Kind.

## Acceptance criteria

- [ ] A declaring condition and a one-element foreach list keep their parentheses in canon (`go test ./internal/parse/ -run '^TestConditionDeclarationCanon$' -count=1 -v | grep -q '^--- PASS: TestConditionDeclarationCanon'`)
