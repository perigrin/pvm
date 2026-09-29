---
title: "`$o->iters / 2` opens a pattern: after a method name an operator is expected"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T17:58:38.978184789Z
updated: 2026-09-29T18:11:01.160674733Z
sessions:
- start_sha: a2c28875f216619c6b27c3f4070515b2f4b5ffbc
  end_sha: 7bc912c2b1f6571957c11606f1ca1b8dfbec6e67
  commits: 1
  started_at: 2026-09-29T17:58:40.178282173Z
  ended_at: 2026-09-29T18:11:01.160674733Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T17:58:40.178282173Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T18:11:01.160674733Z
observed_paths:
- internal/lexer/expect.go
- internal/lexer/methodnameop_test.go
- internal/parse/testdata/perlgitt.ratchet
---

After a parenless method name the lexer expects a term, so an operator that can also start a term is misread: `$_[0]->iters / $_[0]->cpu_p` opens a pattern at the `/` that runs on past the statement. toke.c returns METHCALL0 for `->NAME` without a paren, and an operator follows.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'my $r = $o->iters / $o->cpu_p; my $s = $o->n <2;'
    my $r = $o->iters / $o->cpu_p;
    my $s = $o->n < 2;

perl.git t/benchmark/gh7094-speed-up-keys-on-empty-hash.t:57.

## Acceptance criteria

- [ ] After `->NAME` an operator is expected (`go test ./internal/lexer/ -run '^TestMethodNameThenOperator$' -count=1 -v | grep -q '^--- PASS: TestMethodNameThenOperator'`)
