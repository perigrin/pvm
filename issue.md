---
title: Canon writes () after an unresolved bareword, turning a string into a call
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-24T03:38:23.373117767Z
updated: 2026-10-24T14:40:49.775467954Z
---

Found by the independent review of NL-b (01a1692f-65de, exec/nlb 6b75a5aa) on 2026-10-24. The defect predates NL-b: base 5cb7269a writes `bar()` in all these cases.

## Context

- paths: `internal/parse/canon.go`, wherever a bareword becomes a zero-argument Call
- entrypoints: emitBeforeFat and the zero-argument Call path

Canon writes `()` after a zero-argument call even when no sub of that name is visible. Perl reads an unresolved bareword as a string, so adding the parens turns it into a call. Example:
- `my %h = ((lc bar) => 1)`, with no sub bar.
- Source Deparse: `lc('bar'), 1`.
- Canon writes `lc bar() => 1`, which Deparse reads as `lc(bar()), 1`.

`ucfirst`, `defined`, a `($)`-prototype sub, `$x = bar`, `1 + bar` and `-bar` behave the same way. Faithful reports all of these false. Behind them is a wider question: when should psc read an unresolved bareword as a Call? Measure under `use strict` too, where perl refuses many barewords.

## Acceptance criteria

### Positive Scenarios

- [ ] Faithful probe rows for `my %h = ((lc bar) => 1)`, `$x = bar` and `1 + bar`, with no sub bar (`go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'`)
