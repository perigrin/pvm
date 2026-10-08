---
title: "CORE.pmt has no declaration form for ?: and ->"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-08T01:32:14.443077918Z
updated: 2026-10-08T01:32:14.443077918Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt`, `docs/rfcs/0001-type-libraries.md`

perlop's table names `?:` (right, three operands) and `->` (dereference and method call). Neither fits a declaration form: `:infix`/`:prefix`/`:postfix` take one or two operands, and `->`'s right side is a subscript, call or method name, not a typed operand. Outside the precedence addendum; needs perigrin's ruling on a form for each. Toward the goal (perigrin, 2026-10-08) that every operator is defined in CORE.pmt.

## Steps

- [ ] Propose a form for `?:` and for `->` to perigrin; record the ruling in RFC 0001
- [ ] Declare them, failing tests first; `make test`; commit

## Acceptance criteria

- [ ] ?: and -> have CORE.pmt lines (`go test ./internal/parse/ -run '^TestCoreDeclaresTernaryAndArrow$' -count=1 -v | grep -q '^--- PASS: TestCoreDeclaresTernaryAndArrow'`)
