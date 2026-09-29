---
title: "`my $w = bar $r` splits into two statements at Unknown=0 where the other positions refuse"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T07:46:54.458353983Z
updated: 2026-09-29T08:18:14.562412167Z
sessions:
- start_sha: a0dcdb8fa37d37e59dc0b241cb1a2fe1bbc97bae
  end_sha: 908e6f5973a65e853f2dce0be7bafa7f7495192e
  commits: 1
  started_at: 2026-09-29T08:06:26.220450854Z
  ended_at: 2026-09-29T08:18:14.562412167Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T08:06:26.220450854Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T08:18:14.562412167Z
observed_paths:
- internal/parse/declsplit_test.go
- internal/parse/parse.go
- internal/parse/refusal.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1canon.ratchet
- internal/parse/testdata/t1fidelity.ratchet
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
