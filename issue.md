---
title: "A refaliasing foreach variable, `for \\my $x (...)`, splits the loop from its list"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T14:33:47.579219797Z
updated: 2026-09-29T14:44:31.287479997Z
sessions:
- start_sha: a2b774c555f0d3426b100328495938681de26fe7
  end_sha: e60f44cabd8d3937bca2a7fcf7e39f7eccbe7b58
  commits: 1
  started_at: 2026-09-29T14:33:49.229480093Z
  ended_at: 2026-09-29T14:44:31.287479997Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T14:33:49.229480093Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T14:44:31.287479997Z
observed_paths:
- internal/parse/control.go
- internal/parse/refaliasfor_test.go
- internal/parse/testdata/perlgitt.ratchet
---

A foreach loop variable that is a reference to alias through -- `for \my $x (...)`, `for my \$x (...)`, `for \$x (...)`, `for \@a`, `for \%h`, `for \&f` -- leaves the `foreach` as an empty loop statement, and the rest parses as an expression statement where the `\` operand swallows the list as a call on the variable. A wrong tree at Unknown=0.

perly.y's KW_FOR productions take `my_refgen my_var` and `REFGEN refgen_topic` before the list. Measured with perl 5.42:

    $ perl -MO=Deparse -e 'use feature qw(refaliasing declared_refs); no warnings;
          for my \$a (\1) { 1 } for \my $b (\2) { 2 } for \$c (\3) { 3 }
          for \@d ([4]) { 4 } for \%e ({}) { 5 } for \&f (sub {}) { 6 }'
    foreach \my $a (\1) {
    foreach \my $b (\2) {
    foreach \$c (\3) {
    foreach \@d ([4]) {
    foreach \%e ({}) {
    foreach \&f (sub {

perl.git t/op/decl-refs.t:127, t/op/lvref.t.

## Acceptance criteria

- [ ] A refaliasing loop variable is one loop with its list and block, round-trips, and canons to a fixpoint (`go test ./internal/parse/ -run '^TestRefaliasLoopVariable$' -count=1 -v | grep -q '^--- PASS: TestRefaliasLoopVariable'`)
