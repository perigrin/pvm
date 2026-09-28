---
title: "The repeat operator touching its count lexes as a word: \"ab\" x4 canons as a call to x4()"
state: in-progress
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-28T23:15:22.949588881Z
updated: 2026-09-28T23:31:12.597472111Z
sessions:
- start_sha: d8029d82fa384bc5d781d771297fa742dfb4b01e
  end_sha: ""
  commits: 0
  started_at: 2026-09-28T23:31:12.597472111Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-28T23:31:12.597472111Z
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

## Acceptance criteria

- [ ] `x` followed by digits in operator position is the repeat operator and a number (`go test ./internal/lexer/ -run '^TestRepeatTouchingItsCount$' -count=1 -v | grep -q '^--- PASS: TestRepeatTouchingItsCount'`)
- [ ] NEGATIVE: in term position `x4` is still a word (`go test ./internal/lexer/ -run '^TestRepeatTouchingItsCount$' -count=1 -v | grep -q '^--- PASS: TestRepeatTouchingItsCount'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
