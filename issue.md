---
title: "Three neighbours of print's handle rule: - A is a string, a bareword handle before -1/&y, and the rule after any list operator"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-10T03:02:47.317350536Z
updated: 2026-10-10T03:02:47.317350536Z
---

## Context

- paths: `internal/lexer/lexer.go` (`handleSlotTakesTerm`, `handleTakers`), `internal/parse/call.go` (`termAfterScalar`, `parseFilehandleSlot`)
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

Parked 2026-10-10 by 01a12378-0abd (a `$var` before a term-start after print is the filehandle). Three neighbours of that rule, measured on 5.42:

- `- A` with no parens is the string `"-A"` to perl (`--- A` is `--'-A'`); internal/parse reads `-(A())`.
- A bareword handle: `print FOO -1` and `print FOO &y` take FOO as the filehandle in perl.
- perl applies the `$var` rule after EVERY list operator (`PL_last_lop`), user subs included; the lexer covers only the built-in handle-takers (noted as a limit in lexer.go).

## Steps

- [ ] Measure each with Deparse; failing tests first; implement as perl's rule requires; ratchets toward perl only; `make fuzz-parser`; `make test`; commit

## Acceptance criteria

- [ ] the bareword and minus-bareword neighbours of the handle rule follow perl (`go test ./internal/parse/ -run '^TestHandleRuleNeighbours$' -count=1 -v | grep -q '^--- PASS: TestHandleRuleNeighbours'`)
