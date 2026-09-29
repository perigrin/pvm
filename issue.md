---
title: A dereferenced scalar as a foreach variable, `for ${*$f} (...)`, is read as the condition
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T12:03:54.094791562Z
updated: 2026-09-29T12:22:26.574692929Z
sessions:
- start_sha: 06f95ce2afd792f1c807c1e28af6aa913a2fda95
  end_sha: 4058d489710e76ad523e76f9f1f555f013cd58b9
  commits: 1
  started_at: 2026-09-29T12:03:55.569282834Z
  ended_at: 2026-09-29T12:22:26.574692929Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T12:03:55.569282834Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T12:22:26.574692929Z
observed_paths:
- internal/parse/control.go
- internal/parse/forderef_test.go
- internal/parse/testdata/perlgitt.ratchet
---

A foreach loop variable may be a dereferenced scalar -- `for ${*$f} (5,11,33)`, `for $$r (1,2)` -- and the loop-variable slot took only a plain Variable or a declarator, so `${*$f}` fell to the head and was read as a condition.

Measured with perl 5.42:

    $ perl -MO=Deparse -e 'no strict "refs"; for ${*$f} (5,11,33) { $i++ } for $$r (1,2) { 1 }'
    foreach ${*$f;} (5, 11, 33) {
    foreach $$r (1, 2) {

perl.git t/op/for.t:767.

## Acceptance criteria

- [ ] A deref'd scalar is a foreach loop variable (`go test ./internal/parse/ -run '^TestForeachDerefLoopVariable$' -count=1 -v | grep -q '^--- PASS: TestForeachDerefLoopVariable'`)
