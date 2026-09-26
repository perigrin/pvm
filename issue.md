---
title: "return LIST drops every operand: `return 1, 2` canons as `return ,` at Unknown=0"
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T23:36:39.800469866Z
updated: 2026-09-26T23:36:39.800469866Z
---

## Measured finding

`return LIST` drops every operand, at Unknown=0, and does so WITH the
semicolon as well as without. Measured at `652f4792`:

```
return 1, 2;                canon: return ,;          both operands gone
sub f { return 1, 2; }      canon: sub f {return ,; }
sub f { return ($a, $b); }  canon: sub f {return ; }   parens and both gone
sub f { return 1 }          canon: sub f {return 1; }  correct, ONE operand
```

A single operand survives; a LIST does not. `parseReturn` calls
`p.parseExpr(bpListOp)` with the comment "Below the comma, so the whole list
belongs to the return" -- the intent is right and the result is not. At
`bpListOp` the comma is not consumed as a list separator, so what comes back is
an expression node whose operands never attach.

THE UNKNOWN COUNT CANNOT SEE THIS. It is zero for every spelling above. Only
canon shows it, which puts it in the same class as the eleven wrong-tree
constructs this milestone has already found -- the tree is not a parse of its
source, and a consumer walking it for a `return`'s values finds none.

This was found while fixing 01a0dfb8 (a bare `return` last in a block swallowed
the closing `}`). It is a SEPARATE cause: that bug was about whether
parseReturn looks for an operand, this one is about what it does with the
expression it gets. 01a0dfb8's test deliberately omits the `return LIST` cases
and says why, so that its assertions do not fail for a cause it does not fix.

Not yet measured: how many perl.git `t/` files this reaches. Because the count
is zero either way, the perlgitt ratchet cannot report it -- a canon-based
sweep is needed to size it, the way 01a0dfb8's two halves were isolated.

## Acceptance criteria

- [ ] `return 1, 2;` canons with both operands present rather than as `return ,;` `go test ./internal/parse -run TestReturnList -count=1`
- [ ] `sub f { return ($a, $b); }` keeps its operands `go test ./internal/parse -run TestReturnList -count=1`
- [ ] REGRESSION GUARD, passes today: a single-operand `return 1` and a bare `return` are unchanged `go test ./internal/parse -run 'TestLastInBlockWithoutSemicolon|TestBareLoopControlLastInBlock' -count=1`
- [ ] the perl.git `t/` ratchet is regenerated if it moves, and a canon sweep sizes the construct either way `go test ./internal/parse -run TestPerlGitTRatchet -count=1`
- [ ] the whole tree is green `go test ./... -count=1`
