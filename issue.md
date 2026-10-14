---
title: A list op over a flattened %hash sees only the value types, not the keys
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-14T18:01:06.332913862Z
updated: 2026-10-14T18:01:06.332913862Z
---

## Context

- paths: `internal/infer/lowered_calls.go` (`aggElemType`), `internal/infer/symbols.go` (`Symbol.KeyTypes`, element types)
- commands: `go test ./internal/infer/`; perl 5.42 `/home/perigrin/.local/bin/perl -w`

Found 2026-10-14 by the vars false-positive fixes (01a13b5c-7537, 01a13b5c-a052): a hash's element type now holds only its VALUES (correct for `$h{k}` and `values %h`). `aggElemType` reuses it when `sort`, `grep` or `reverse` read a `%hash` -- but those flatten keys AND values, so their element type should join the keys' Str in. `my %h = ("abc" => 1); (grep {1} %h)[0] + 1` -- perl 5.42 -w warns `Argument "abc" isn't numeric`; PSC no longer does. (Before, quoted keys added Str by accident and bareword keys never did.)

## Steps

- [ ] Failing test first; a list op reading a flattened %hash sees Str|<value type>; `values %h` keeps the value type; `keys %h` is Str; no new false positive on `sort %h` (measure); `make test`; commit

## Acceptance criteria

- [ ] a flattened hash's elements include its keys (`go test ./internal/infer/ -run '^TestFlattenedHashIncludesKeys$' -count=1 -v | grep -q '^--- PASS: TestFlattenedHashIncludesKeys'`)
