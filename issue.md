---
title: "`$a and return 1` refuses: return is a list operator only at a statement start"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-29T07:13:37.6125308Z
updated: 2026-09-29T07:13:37.6125308Z
---

Only the statement-start path read `return` as a list operator. Inside an expression it went to the generic call path, so `$a and return 1;` refused on the `1`, and a bare `$a and return;` became a call to a sub named return. toke.c lexes it as `OLDLOP(OP_RETURN)` wherever it stands.

Measured with perl 5.42:

    $ perl -MO=Deparse,-p -e 'sub f { $a and return 1, 2; $a || return 3;
          my @r = (2, 3, return @a ? 1 : ()); $c ? return : uc($_); f(return) }'
    ($a and (return 1, 2));
    ($a or (return 3));
    (my(@r) = (2, 3, (return (@a ? 1 : ()))));
    ($c ? (return) : uc($_));
    f((return));

perl.git t/ spells it in op/pack.t:46 (`$err =~ $e and return 1;`), op/stat.t:689-695 and op/sprintf2.t:645, among others. The list must stop wherever any list operator's does, which includes a ternary's `:` (PerlOnJava unit/zz_nested_map_return.t has `$_ eq 'missing' ? return : uc($_)`).

## Acceptance criteria

- [ ] `return` inside an expression is a LoopControl taking its list, and canon is a fixpoint (`go test ./internal/parse/ -run '^TestReturnInsideAnExpression$' -count=1 -v | grep -q '^--- PASS: TestReturnInsideAnExpression'`)
- [ ] REGRESSION GUARD: loop controls keep their statement canon (`go test ./internal/parse/ -run '^TestLoopControlWithoutModifier$' -count=1 -v | grep -q '^--- PASS: TestLoopControlWithoutModifier'`)
- [ ] REGRESSION GUARD: the perl.git t/ ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
