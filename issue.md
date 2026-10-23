---
title: Canon wraps an assignment in a ternary's middle in parens the source lacks
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-23T20:36:43.139953275Z
updated: 2026-10-23T20:36:43.139953275Z
---

Found in the independent review of 01a16602-5ebe (exec/5ebe, f80cbfbf), 2026-10-23. The defect predates 5ebe; base 4af1670f does the same.

## Context

- paths: `internal/parse/canon.go`
- entrypoints: canon's wrapping of a Ternary's middle operand

`1 ? my ($y) = 1 : 2` is not faithful: canon wraps the assignment in the ternary's middle in parens that the source doesn't have. Perl parses a ternary's middle as a full expression, so no wrap is needed there. NU's follow-power design gives the middle 0, per K14. Check whether that fix already covers this case once NU merges. If it does, close this issue with the evidence.

## Acceptance criteria

### Positive Scenarios

- `go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'` (a probe row for the snippet, faithful)
