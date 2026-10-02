---
title: :unary marks a no-prototype builtin a named unary
state: cancelled
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fa46-9aff-7c99-b17b-7278e810608c
blocks:
- 01a0fa46-e28a-7469-998f-ebdd26d18494
created: 2026-10-02T01:42:22.455228461Z
updated: 2026-10-02T17:18:12.59211723Z
transitions:
- state: cancelled
  actor: human:git-zhi
  timestamp: 2026-10-02T17:18:12.59211723Z
---

RFC 0001, "Builtins with no prototype": about 20 builtins have no prototype; a sub with none is a list operator, so `:unary` marks the eight that parse as named unaries (`chomp chop defined delete do eval exists goto`, per `internal/parse/keyword.go`). Their special forms stay in the parser. Re-mark the RFC section *Implemented* once CORE.pmt declares them.

## Acceptance criteria

- [ ] CORE.pmt declares the no-prototype builtins and :unary is read (`go test ./internal/parse/ -run '^TestCoreUnaryBuiltins$' -count=1 -v | grep -q '^--- PASS: TestCoreUnaryBuiltins'`)
- [ ] The RFC section is implemented (`grep -qF '### Builtins with no prototype (*Implemented*' docs/rfcs/0001-type-libraries.md`)
