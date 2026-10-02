---
title: :context(...) selects a declaration by calling context
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fa46-ece2-7c78-87ed-84da0b5d65f9
blocks:
- 01a0fa46-fccf-7f53-a476-c02f8f07616d
- 01a0fafc-0dd7-7cd6-bfc3-262c0b8019eb
created: 2026-10-02T01:42:28.423071525Z
updated: 2026-10-02T05:00:24.592252302Z
---

RFC 0001, "`:context(...)`": prototype sigils name the contexts a declaration answers for -- `:context($)` scalar, `:context(@)` list, `:context()` void, `:context($@)` both; no attribute means every context, and absent differs from empty. A call in a context no declaration answers for has no type (`sub sort :context(@) ...`). Boolean context counts as scalar (measured: wantarray reports scalar in `if (f())`). Re-mark the RFC section *Implemented*.

## Acceptance criteria

- [ ] :context parses and selects (`go test ./internal/parse/ -run '^TestContextAttribute$' -count=1 -v | grep -q '^--- PASS: TestContextAttribute'`)
- [ ] The RFC section is implemented (`grep -q '^### .*:context.*(\*Implemented\*' docs/rfcs/0001-type-libraries.md`)
