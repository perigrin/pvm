---
title: "`Y / X` after `use constant` opens a pattern: constants are not declared"
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-09-30T17:39:31.523098476Z
updated: 2026-09-30T17:39:31.523098476Z
---

`is(Y / X, 2, ...)` refuses when X and Y come from `use constant`. After an undeclared word the lexer expects a term, so `/ X, 2, 'constant division');\n...` lexes as a pattern. perl's lexer finds the constant sub in its symbol table and expects an operator. The parser doesn't record the names either.

Measured with perl 5.42: `prototype "main::X"` is "" for `use constant X => 2`, for the keys of `use constant { A => 1 }`, for a list constant, and for `sub Z () {...}`. `print Y / X` deparses as a constant-folded division.

PerlOnJava unit/constant.t:60.

## Acceptance criteria

- [ ] The lexer expects an operator after a declared constant (`go test ./internal/lexer/ -run '^TestDeclaredConstantDivides$' -count=1 -v | grep -q '^--- PASS: TestDeclaredConstantDivides'`)
- [ ] `use constant` declares its names with the empty prototype (`go test ./internal/parse/ -run '^TestUseConstantDeclares$' -count=1 -v | grep -q '^--- PASS: TestUseConstantDeclares'`)
