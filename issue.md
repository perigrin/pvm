---
title: A builtin overridden by an import still parses as the builtin
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-23T21:37:50.149043545Z
updated: 2026-10-24T14:40:49.39780315Z
---

Found by the NU worker (01a1690b-1d10, exec/nu, 945135e2) on 2026-10-23, and measured with perl 5.44.0. The defect predates NU.

## Context

- paths: `internal/parse/shape.go`, `internal/parse/call.go`
- entrypoints: `keywordShape`, which never asks whether a builtin has been overridden by an import

A builtin overridden by an import still parses as the builtin. Example: `use subs 'chdir'; chdir $x, 1`. psc parses `chdir` as the builtin named unary, `(chdir($x), 1)`. Perl calls `main::chdir($x, 1)` as a list operator. Canon writes the source's own spelling either way, so Faithful cannot catch this; only a tree check or the oracle can. NU's probe file carries the row, and the row's comment names this defect.

The same gap applies to `use subs` and to a `CORE::GLOBAL::` override, and to any `use Module` whose import overrides a builtin (`use Time::HiRes 'sleep'`). Production psc never runs perl, so an override that psc cannot see statically must hedge, not guess.

## Acceptance criteria

### Positive Scenarios

- [ ] A tree-level probe row for `use subs 'chdir'; chdir $x, 1` that asserts the list-operator reading (`go test ./internal/parse/ -run '^TestM3ProbesHeldToPerl$' -count=1 | grep -q '^ok'`)
