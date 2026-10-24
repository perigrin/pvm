---
title: print FH (LIST) parses the filehandle as a call
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-24T02:04:17.860956737Z
updated: 2026-10-24T14:40:49.589577426Z
---

Found by the NL-b worker (01a1692f-65de, exec/nlb 3630cbed) on 2026-10-24. The defect predates NL-b.

## Context

- paths: `internal/parse/indirect.go`, `internal/parse/call.go`
- entrypoints: the filehandle slot of print, printf and say, when parens follow it

The parser reads `print STDOUT (1, 2)` as a call `STDOUT(1, 2)`. perl reads it as printing the list `(1, 2)` to STDOUT: the bareword is a filehandle, and the parens belong to the list. Canon writes it back unchanged, so Faithful cannot see the wrong tree. Measure the related forms with perl 5.44: `print STDERR ("x")`, `printf STDOUT ("%s", 1)`, and `say STDOUT (1)` under `use feature 'say'`.

**Also from the 75d4 review (2026-10-24):** `print STDOUT ((1))` re-emits as `print STDOUT((1))`, and Faithful says true because it never sees whitespace. perl reads the re-emitted form as `print(STDOUT(1))`, against the source's `print(STDOUT 1)`. So the misparse produces a meaning change that Faithful cannot see. Add it as a probe row.

## Acceptance criteria

### Positive Scenarios

- [ ] A tree-level probe row asserting that `print STDOUT (1, 2)` prints a list to a filehandle (`go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'`)
