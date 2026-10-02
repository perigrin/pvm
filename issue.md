---
title: Declare perl's operators in CORE.pmt
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fa46-902d-73c3-b74c-c6bf23263f42
created: 2026-10-02T01:42:37.41941306Z
updated: 2026-10-02T01:42:40.998729607Z
---

signatures.go's `binaryOps` and `unaryOps` become operator declarations in CORE.pmt, spelled as D2 decides (`sub + :infix (Num $x, Num $y) Num;`), with the result rule each needs. The parser reads an operator name where a sub name stands. A test holds each operator's declaration equal to its signatures.go row, except where D2's result rule deliberately corrects one (`Int + Int`). Re-mark the RFC's operator section *Implemented*.

## Acceptance criteria

- [ ] Operator declarations parse and match signatures.go (`go test ./internal/parse/ -run '^TestCoreOperatorDeclarations$' -count=1 -v | grep -q '^--- PASS: TestCoreOperatorDeclarations'`)
- [ ] The RFC section is implemented (`grep -qF '### Operator declarations (*Implemented*' docs/rfcs/0001-type-libraries.md`)
