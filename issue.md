---
title: "multi sub: several declarations for one name"
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fa46-a8d1-77c5-97d8-66cf3b4ec55d
blocks:
- 01a0fa46-f147-75f6-bd3c-4aa5f5644ee5
- 01a0fa46-fccf-7f53-a476-c02f8f07616d
- 01a0fafc-0dd7-7cd6-bfc3-262c0b8019eb
created: 2026-10-02T01:42:27.298629109Z
updated: 2026-10-02T05:00:23.434241708Z
---

RFC 0001, "Multi declarations": a declaration may be `multi sub`, several signatures for one name, selected at a call site by arity or by argument types that are not subtypes of each other. Measured: `each %h` gives (Str, v) and `each @a` (Int, v); `select(FH)` returns the previous handle and `select(r,w,e,t)` a count. When the call site cannot decide, the consumer joins the candidates. Bounded polymorphism (`abs`) needs no multi. Re-mark the RFC section *Implemented*.

## Acceptance criteria

- [ ] multi declarations parse and select by arity and argument type (`go test ./internal/parse/ -run '^TestMultiDeclarations$' -count=1 -v | grep -q '^--- PASS: TestMultiDeclarations'`)
- [ ] The RFC section is implemented (`grep -qF '### Multi declarations (*Implemented*' docs/rfcs/0001-type-libraries.md`)
