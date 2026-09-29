---
title: "Canon drops a deref's braces: `${$h->{k}}` becomes `$$h->{k}`, a different program"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T11:49:30.695128033Z
updated: 2026-09-29T12:01:25.349403181Z
sessions:
- start_sha: 8ef76e349e30a181a72fd26c176383cbd4f31bc0
  end_sha: 06f95ce2afd792f1c807c1e28af6aa913a2fda95
  commits: 1
  started_at: 2026-09-29T11:49:31.65678076Z
  ended_at: 2026-09-29T12:01:25.349403181Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T11:49:31.65678076Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T12:01:25.349403181Z
observed_paths:
- internal/parse/canon.go
- internal/parse/derefcanon_test.go
- internal/parse/testdata/t1canon.ratchet
- internal/parse/testdata/t1fidelity.ratchet
---

A deref Unary's canon writes its sigil and then its operand, so the braces of `${EXPR}` are dropped whatever EXPR is. The bare `$$name` form reads only a scalar variable, and anything more binds differently: `${$h->{k}}` canonicalises as `$$h->{k}`, which is `${$h}->{k}`; `${\$x}` as `$\$x`; `${*$f}` as `$*$f`. Wrong programs at Unknown=0.

Measured with perl 5.42:

    $ perl -e 'my $h={k=>\"K"}; my $x = ${$h->{k}}; print "$x\n";
          my $y = eval q{$$h->{k}}; print defined $y ? "y=$y\n" : "err: $@";'
    K
    err: Not a SCALAR reference at (eval 1) line 1.

Deparse writes `${*$f;}`, `${$h->{'k'};}`, `${\$x;}` and `$$$r`.

## Acceptance criteria

- [ ] Canon keeps a deref's braces unless the operand is a plain scalar variable (`go test ./internal/parse/ -run '^TestDerefCanonKeepsBraces$' -count=1 -v | grep -q '^--- PASS: TestDerefCanonKeepsBraces'`)
