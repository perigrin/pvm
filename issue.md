---
title: A call with parens before => lowers as a string
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-24T02:04:17.921523729Z
updated: 2026-10-24T02:04:17.921523729Z
---

Found by the NL-b worker (01a1692f-65de, exec/nlb 3630cbed) on 2026-10-24. The defect predates NL-b.

## Context

- paths: `internal/lower/literals.go`
- entrypoints: lowering of a `=>` pair's left operand

Lowering treats a zero-argument Call on the left of `=>` as a string even when the source wrote call parens. In `(foo() => 1)` perl calls `foo` and the pair is `(foo(), 1)`. Only a bare `foo => 1` is autoquoted. The parser records the parens as ArgParen, but lowering ignores them, so inference sees a string where perl sees foo's return value.

## Acceptance criteria

### Positive Scenarios

- `go test ./internal/lower/ -run 'FatCommaCallWithParens' -count=1 -v | grep -q '^--- PASS'` (a lowering test: `(foo() => 1)` lowers to a call, and `(foo => 1)` lowers to a string)
