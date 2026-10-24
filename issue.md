---
title: A lexical sub named use or package is still misread
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-24T18:34:18.653363811Z
updated: 2026-10-24T18:34:18.653363811Z
---

Found by the independent review of 01a167da-3323 (exec/3323, 72d8c827) on 2026-10-24. 3323 taught psc that a lexical sub shadows a statement keyword. These two shapes are still wrong.

## Context

- paths: `internal/parse/use.go`, `internal/parse/` (the package, format and sub declaration paths)
- entrypoints: a statement keyword shadowed by a lexical sub

1. `my sub use {} use strict;` — perl reads this as `'strict'->use`, an indirect-object method call. psc reads it as a call `use(strict)`. 3323's parseTheRest correctly declines the keyword, but `WORD Bareword` then needs the same indirect-object handling as any other word. This shape is new with 3323 and contrived.
2. `my sub package {} package('x');` — perl reads a call, psc reads Unknown. This is safe, since Unknown is never WRONG. It needs the same lexical check in the package, format and sub declaration paths.

Measure each one with perl 5.44 using Concise and Deparse,-p.

## Acceptance criteria

### Positive Scenarios

- [ ] Both shapes give perl's tree (`go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'`)
