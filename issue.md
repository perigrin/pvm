---
title: return followed by or/and re-emits as a syntax error
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-24T02:36:35.50313554Z
updated: 2026-10-24T14:40:49.657664077Z
---

Found by the independent review of NL-b (01a1692f-65de, exec/nlb 3630cbed) on 2026-10-24. The defect predates NL-b; base 5cb7269a writes the same text.

## Context

- paths: `internal/parse/canon.go` and the parser's handling of `return` with a low-precedence operator after it
- entrypoints: `return` followed by `or`/`and`

`return or 1;` re-emits as `return;or 1`, and `sub f { return 1, 2 or die }` re-emits as `...;or die }}`. Both are syntax errors. Faithful already reports these lines false. Measure the tree with perl 5.44 first. If psc's tree ends the return at `or`, this is a tree defect, not just a canon defect.

## Acceptance criteria

### Positive Scenarios

- [ ] Faithful probe rows for `return or 1;` and `sub f { return 1, 2 or die }` (`go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'`)
