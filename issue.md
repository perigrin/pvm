---
title: internal/infer reads CORE.pmt and module declarations; signatures.go goes
state: pending
urgency: normal
milestone: type-libraries
created: 2026-10-02T01:42:42.231544132Z
updated: 2026-10-02T01:42:42.231544132Z
---

The last fold. Once m2 has moved internal/infer onto internal/parse (blocked by its retire issue), infer takes builtin and operator signatures from CORE.pmt and module signatures from `.pmt` declarations, and `internal/types/signatures.go` is deleted. This changes what `psc check` reports wherever a declaration is more precise, which is why it waits for m2's behaviour-preserving cutover.

## Acceptance criteria

- [ ] infer types a call from a .pmt declaration (`go test ./internal/infer/ -run '^TestInferReadsDeclarations$' -count=1 -v | grep -q '^--- PASS: TestInferReadsDeclarations'`)
- [ ] signatures.go is gone (`sh -c '! test -e internal/types/signatures.go'`)
