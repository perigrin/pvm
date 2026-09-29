---
title: "`f 1, || 2` refuses: a comma before an operator ends the list only for assignment"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T22:20:52.960997107Z
updated: 2026-09-29T22:20:52.960997107Z
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
