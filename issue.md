---
title: "multi sub: several declarations for one name"
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-02T01:42:27.298629109Z
updated: 2026-10-02T01:42:27.298629109Z
---

RFC 0001, "Multi declarations": a declaration may be `multi sub`, several signatures for one name, selected at a call site by arity or by argument types that are not subtypes of each other. Measured: `each %h` gives (Str, v) and `each @a` (Int, v); `select(FH)` returns the previous handle and `select(r,w,e,t)` a count. When the call site cannot decide, the consumer joins the candidates. Bounded polymorphism (`abs`) needs no multi. Re-mark the RFC section *Implemented*.

## Acceptance criteria

- [ ] multi declarations parse and select by arity and argument type (`go test ./internal/parse/ -run '^TestMultiDeclarations$' -count=1 -v | grep -q '^--- PASS: TestMultiDeclarations'`)
- [ ] The RFC section is implemented (`grep -qF '### Multi declarations (*Implemented*' docs/rfcs/0001-type-libraries.md`)
