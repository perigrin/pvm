---
title: A bareword operator word as a hash subscript key parses as the operator
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-23T23:49:20.221374189Z
updated: 2026-10-23T23:49:20.221374189Z
---

Found by the independent review of NL-a (01a1690b-4745, exec/nla bc947a68) on 2026-10-23. The defect predates NL-a.

## Context

- paths: `internal/parse/expr.go` and `internal/parse/term.go`. These cover hash-subscript key parsing, which is where `not` is read as an operator.
- entrypoints: the bareword-key rule inside `{...}` subscripts

`$h{not};` and `$h->{not}{x}` parse `not` as the operator, with `}` as its operand. Canon then writes `$h{not }};`, which is a syntax error. Perl reads a lone bareword inside a subscript as a string key, `$h{"not"}`. Faithful already reports these lines false, so the ratchets count the files as unfaithful. Any M3-dir file with this pattern is listed by 01a16602-5f77.

Other operator words that perl quotes as subscript keys (`and`, `or`, `x`, `eq`, `lt`) may share the defect. Measure them with perl 5.44.

## Acceptance criteria

### Positive Scenarios

- `go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'` (probe rows for `$h{not}`, `$h->{not}{x}`, `$h{and}` and `$h{x}`, each faithful)
