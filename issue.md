---
title: A dereferenced scalar as a foreach variable, `for ${*$f} (...)`, is read as the condition
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T12:03:54.094791562Z
updated: 2026-09-29T12:03:54.094791562Z
---

A foreach loop variable may be a dereferenced scalar -- `for ${*$f} (5,11,33)`, `for $$r (1,2)` -- and the loop-variable slot took only a plain Variable or a declarator, so `${*$f}` fell to the head and was read as a condition.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'no strict "refs"; for ${*$f} (5,11,33) { $i++ } for $$r (1,2) { 1 }'
    foreach ${*$f;} (5, 11, 33) {
    foreach $$r (1, 2) {

perl.git t/op/for.t:766.

## Acceptance criteria

- [ ] A deref'd scalar is a foreach loop variable (`go test ./internal/parse/ -run '^TestForeachDerefLoopVariable$' -count=1 -v | grep -q '^--- PASS: TestForeachDerefLoopVariable'`)
