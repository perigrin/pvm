---
title: "`f 1, || 2` refuses: a comma before an operator ends the list only for assignment"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T22:20:52.960997107Z
updated: 2026-09-29T22:27:29.696167094Z
sessions:
- start_sha: e5da762ddfc9da30d8dd6e16c6d2fe3d8f008058
  end_sha: 1bb6f94e6508cdb97d0a1f778c5f5fbf7dac20ac
  commits: 1
  started_at: 2026-09-29T22:20:53.896953992Z
  ended_at: 2026-09-29T22:27:29.696167094Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T22:20:53.896953992Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T22:27:29.696167094Z
observed_paths:
- internal/parse/commaassign_test.go
- internal/parse/expr.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/testdata/t1.ratchet
- internal/parse/testdata/t1fidelity.ratchet
---

A comma followed by an operator that cannot begin an element ends the list, and perl then applies the operator to the list operator's call. The rule was limited to assignment operators (and too wide there): `ok open(...), '...', || _diag $!` refused, and `f 1, += 2`, which perl rejects, was accepted.

Measured with perl 5.42, with `sub f {}` declared:

    f 1, || 2;     '???' unless f 1;
    f 1, && 2;     '???' if f 1;
    f 1, == 2;     f(1) == 2;
    f 1, ? 1 : 2;  f(1) ? '???' : '???';
    f 1, . "x";    f(1) . 'x';
    f 1, += 2;     syntax error
    f 1, ** 2;     reads a glob: f 1, *2

perl.git t/io/open.t:281.

## Acceptance criteria

- [ ] The list ends before every operator perl accepts there, and not the others (`go test ./internal/parse/ -run '^TestTrailingCommaBeforeOperator$' -count=1 -v | grep -q '^--- PASS: TestTrailingCommaBeforeOperator'`)
