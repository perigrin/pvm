---
title: "The repeat operator touching its count lexes as a word: \"ab\" x4 canons as a call to x4()"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-28T23:15:22.949588881Z
updated: 2026-10-11T15:17:45.257464401Z
sessions:
- start_sha: d8029d82fa384bc5d781d771297fa742dfb4b01e
  end_sha: f196c8a1cfd85c12aac31986c2dfd014375fbf79
  commits: 1
  started_at: 2026-09-28T23:31:12.597472111Z
  ended_at: 2026-09-28T23:42:23.422688199Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-28T23:31:12.597472111Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-28T23:42:23.422688199Z
observed_paths:
- internal/lexer/repeatcount_test.go
- internal/lexer/scan.go
- internal/parse/testdata/perlgitt.ratchet
- internal/parse/testdata/t1.ratchet
---

The repetition operator written with its count touching it -- `"ab"x4`, `"ab" x4`
-- lexes as one Word `x4`, so the tree is the string followed by a CALL to a
sub named x4, at Unknown=0:

    my $x = "ab" x4;     canon  my $x = "ab";x4();       WRONG

perl reads `x` then `4` in operator position, even with a `sub x4` declared --
measured on 5.42.0, `sub x4 {9} my $x = "ab" x4;` deparses to
`my $x = 'abababab';`. toke.c's `x` case in yyl_keylookup checks for a digit
after the `x` when an operator is expected.

Reach: 3 perl.git t/ files spell it (op/write.t's `"'NN'\n"x4` among them).

### Whole-suite criteria removed (2026-10-11)

perigrin, on a milestone verify that would have run for 9+ hours: "Why are
we doing 73 full test suite runs?!?" Criteria that only re-ran a whole suite
or package (`go test ./...`, `make test`, the whole conformance or parse
package) restated "make test passed", which gates every commit, while CI
gates every PR -- and `git zhi verify` re-ran each copy for every done issue.
They are removed; this issue's named tests and ratchet checks remain.

## Acceptance criteria

- [ ] `x` followed by digits in operator position is the repeat operator and a number (`go test ./internal/lexer/ -run '^TestRepeatTouchingItsCount$' -count=1 -v | grep -q '^--- PASS: TestRepeatTouchingItsCount'`)
- [ ] NEGATIVE: in term position `x4` is still a word (`go test ./internal/lexer/ -run '^TestRepeatTouchingItsCount$' -count=1 -v | grep -q '^--- PASS: TestRepeatTouchingItsCount'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
