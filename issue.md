---
title: "Two declaration comma shapes still disagree with perl: list declarations, and a later low-precedence operator"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T11:12:05.694610041Z
updated: 2026-10-06T11:12:05.694610041Z
---

## Context

- paths: `internal/parse` (`parseVarDeclHead`, declaration initialisers), `internal/parse/lowlogical_test.go`
- commands: `make test`; ratchets; perl 5.42 `/home/perigrin/.local/bin/perl -MO=Deparse,-p`

Found by the alignment gate on 01a10dc1-3530-762f-9a03-177445629c6b, which fixed a statement-level declaration's initialiser to stop at the comma (602a1626). Two shapes still disagree with perl (neither caused by that change):

1. A list declaration: `my ($a,$b) = (1,2), 3;` and `local ($a,$b) = (1,2), 3;` keep the comma inside the declaration; perl: `((my($a,$b)=(1,2)), 3)`.
2. A declaration in the middle of a comma list before a low-precedence operator: `my $x=1, my $y=2 and 3` gives `(, (my $x 1) (and (my $y 2) 3))`; perl: `(((my $x=1),(my $y=2)) and 3)`.

## Steps

- [ ] Measure both shapes and their neighbours with Deparse,-p
- [ ] Write failing parse tests; fix; ratchets must not regress (report movement, checked against perl); `make test`; commit

## Acceptance criteria

- [ ] both declaration shapes parse as perl's Deparse shows (`go test ./internal/parse/ -run '^TestDeclarationCommaShapesMatchPerl$' -count=1 -v | grep -q '^--- PASS: TestDeclarationCommaShapesMatchPerl'`)
