---
title: Decide how a builtin with no prototype opts out of prototype derivation
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-02T01:42:06.335001949Z
updated: 2026-10-02T01:42:06.335001949Z
---

RFC 0001 open question 9. Prototypes derive from typed signatures, but about 20 builtins have no prototype because their parse is their own (`print`'s filehandle slot, `sort`'s block). `print`'s typed signature would derive a prototype `print` does not have, and applying it would change how `print` parses. Absent and empty differ for `:prototype` already (`()` is niladic), so the opt-out needs its own spelling.

perigrin decides; the decision lands in RFC 0001 as `### Builtins that keep their own parse (*Decided*)`.

## Acceptance criteria

- [ ] The RFC records the decision (`grep -qF '### Builtins that keep their own parse (*Decided*' docs/rfcs/0001-type-libraries.md`)
