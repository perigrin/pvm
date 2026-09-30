---
title: "`opt_five | opt_three` refuses: `|` does not end a sub's arguments"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T18:03:39.216391287Z
updated: 2026-09-30T18:03:39.216391287Z
---

`opt_five | opt_three` refuses. `|` is not in the operators that end a sub's argument list, so the parser tries to read it as the start of an argument.

Measured with perl 5.42 and -MO=Deparse, for `sub f { 1 }` and for `sub g (;$) { 1 }` alike: `f | 2` is `f() | 2`, and `^`, `>`, `>=` and `>>` behave the same way. `&` and `%` begin a term: `f & 2` is `f(&2)`.

PerlOnJava unit/subroutine_prototype_args.t:263-266.

## Acceptance criteria

- [ ] `| ^ > >= >>` end a sub's argument list (`go test ./internal/parse/ -run '^TestInfixEndsArguments$' -count=1 -v | grep -q '^--- PASS: TestInfixEndsArguments'`)
