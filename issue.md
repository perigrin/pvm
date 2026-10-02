---
title: internal/infer reads CORE.pmt and module declarations; signatures.go goes
state: pending
urgency: normal
milestone: type-libraries
blocked_by:
- 01a0fa46-fccf-7f53-a476-c02f8f07616d
- 01a0fa47-146b-7047-84b6-20ac55c1b0f6
- 01a0c115-df67-7d69-8883-73d40b0ae15d
created: 2026-10-02T01:42:42.231544132Z
updated: 2026-10-02T01:42:46.073902177Z
---

The last fold. Once m2 has moved internal/infer onto internal/parse (blocked by its retire issue), infer takes builtin and operator signatures from CORE.pmt and module signatures from `.pmt` declarations, and `internal/types/signatures.go` is deleted. This changes what `psc check` reports wherever a declaration is more precise, which is why it waits for m2's behaviour-preserving cutover.

## Acceptance criteria

- [ ] infer types a call from a .pmt declaration (`go test ./internal/infer/ -run '^TestInferReadsDeclarations$' -count=1 -v | grep -q '^--- PASS: TestInferReadsDeclarations'`)
- [ ] signatures.go is gone (`sh -c '! test -e internal/types/signatures.go'`)
