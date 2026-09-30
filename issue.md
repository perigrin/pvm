---
title: TestMemoryAlerts has a data race under -race (monitoring.go)
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-30T17:08:31.288502597Z
updated: 2026-09-30T17:08:31.288502597Z
---

`go test -race ./internal/memory` reports a DATA RACE in TestMemoryAlerts, between monitoring.go:83/99 and monitoring_test.go:196-214. It predates the LazyValue fix: it happens on the tree before that fix as well. `make test` doesn't run -race, so it doesn't gate commits today.

## Acceptance criteria

- [ ] The memory package is race-free (`sh -c 'go test -race ./internal/memory -count=3 2>&1 | grep -q "^ok"'`)
