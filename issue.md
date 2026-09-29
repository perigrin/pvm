---
title: "`my $w = bar $r` splits into two statements at Unknown=0 where the other positions refuse"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T07:46:54.458353983Z
updated: 2026-09-29T07:46:54.458353983Z
---

`WORD $var` with an unknown WORD refuses everywhere except a declaration's initialiser, where it splits the statement in two at Unknown=0:

    $w = bar $r, 1;        unknown=1   (refuses, as 01a0ebb0-6329-7a13-b398-a5120ec9bbe8 intends)
    f(bar $r, 1);          unknown=1
    my $w = bar $r;        unknown=0   canon: my $w = bar();$r;
    my $w = bar $r, 1;     unknown=0   canon: my $w = bar();$r , 1;

perl reads it as an indirect method call, one statement. Measured with perl 5.42:

    $ perl -MO=Deparse -e 'my $w = Foo::bar $r, 1;'
    my $w = $r->Foo::bar, '???';

Until 01a0ebb0 can read `WORD $var`, the declaration path must refuse as the others do, not emit two statements.

## Acceptance criteria

- [ ] `my $w = bar $r` refuses rather than splitting (`go test ./internal/parse/ -run '^TestDeclarationInitialiserDoesNotSplit$' -count=1 -v | grep -q '^--- PASS: TestDeclarationInitialiserDoesNotSplit'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
