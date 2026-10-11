---
title: "return LIST drops every operand: `return 1, 2` canons as `return ,` at Unknown=0"
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T23:36:39.800469866Z
updated: 2026-10-11T11:09:37.908996213Z
sessions:
- start_sha: 971b903102dc4a335660172f3b1b5ecad5cf3330
  end_sha: 971b903102dc4a335660172f3b1b5ecad5cf3330
  commits: 0
  started_at: 2026-10-11T11:09:37.710662966Z
  ended_at: 2026-10-11T11:09:37.908996213Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-11T11:09:37.710662966Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-11T11:09:37.908996213Z
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

### Re-measured (2026-10-11)

Measured at pu `24bfac4a`. `return 1, 2;` canons as `return 1 , 2;` and `sub f { return ($a, $b); }` as `sub f {return ($a, $b);}`, both at Unknown=0, as perl Deparses them. Fixed by c80ab769, merged in e87dee9c. The filed cause was wrong: the parse was right, and canon's LoopControl arm wrote each child's Text. The canon sweep the last criterion asked for is c80ab769's t1canon move, 158 -> 65 files, 95 better.

The commands were bare backticks without parens, which verify does not run. They now name TestReturnList, TestLastInBlockWithoutSemicolon, TestBareLoopControlLastInBlock and TestPerlGitTRatchet, each with a PASS grep.

The conformance and whole-suite commands carry `-timeout=30m`: internal/conformance took 696s in a full run, past go test's 10-minute default, and timed out without it.

## Acceptance criteria

- [ ] `return 1, 2;` canons with both operands present rather than as `return ,;` (`go test ./internal/parse/ -run '^TestReturnList$' -count=1 -v | grep -q '^--- PASS: TestReturnList'`)
- [ ] `sub f { return ($a, $b); }` keeps its operands (`go test ./internal/parse/ -run '^TestReturnList$' -count=1 -v | grep -q '^--- PASS: TestReturnList'`)
- [ ] REGRESSION GUARD, passes today: a single-operand `return 1` and a bare `return` are unchanged (`go test ./internal/parse/ -run '^TestLastInBlockWithoutSemicolon$' -count=1 -v | grep -q '^--- PASS: TestLastInBlockWithoutSemicolon' && go test ./internal/parse/ -run '^TestBareLoopControlLastInBlock$' -count=1 -v | grep -q '^--- PASS: TestBareLoopControlLastInBlock'`)
- [ ] the perl.git `t/` ratchet holds (`go test ./internal/parse/ -run '^TestPerlGitTRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTRatchet'`)
- [ ] the whole tree is green (`go test ./... -count=1 -timeout=30m`)
