---
title: Declare perl's operators in CORE.pmt
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fa46-902d-73c3-b74c-c6bf23263f42
- 01a0fa46-a8d1-77c5-97d8-66cf3b4ec55d
blocks:
- 01a0fa47-2737-7fdb-bd69-63e254d5a770
created: 2026-10-02T01:42:37.41941306Z
updated: 2026-10-02T16:09:41.121209385Z
---

signatures.go's `binaryOps` and `unaryOps` become operator declarations in CORE.pmt, spelled as RFC 0001's "Operator declarations" section (provisional) gives them: `sub + :infix(ADD) (Num $x, Num $y) Num;`, with `:prefix`/`:postfix`, and multis for `..` (`:context`) and `x` (parenthesised left operand). There is no result rule: the declared type is an upper bound (Int <: Num), and narrowing is inference's. The parser reads an operator name where a sub name stands. Tests hold each operator's declaration equal to its signatures.go row and its `:infix` class equal to its level in the parser's precedence table. Re-mark the RFC's operator section *Implemented*.

## Acceptance criteria

- [ ] Operator declarations parse and match signatures.go (`go test ./internal/parse/ -run '^TestCoreOperatorDeclarations$' -count=1 -v | grep -q '^--- PASS: TestCoreOperatorDeclarations'`)
- [ ] The RFC section is implemented (`grep -qF '### Operator declarations (*Implemented*' docs/rfcs/0001-type-libraries.md`)
