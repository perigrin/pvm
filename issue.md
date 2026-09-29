---
title: TestAnalyzer_SlowestOperations orders sleeps by wall clock and fails under load
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-29T09:12:59.621659274Z
updated: 2026-09-29T09:12:59.621659274Z
---

`internal/performance/analyzer_test.go`'s TestAnalyzer_SlowestOperations orders three operations by wall-clock time -- sleeps of 1ms, 10ms and 50ms -- and asserts the order. Under load the scheduler can stretch the 1ms sleep past the 10ms one, and `make test` fails on a package no change touched:

    --- FAIL: TestAnalyzer_SlowestOperations (0.08s)
        analyzer_test.go:130: Expected medium_op second, got fast_op

Seen once on 2026-09-29 while another package's tests ran in a second worktree; `go test ./internal/performance -count=5` alone passes.

The test is about the analyzer's ORDERING, not about sleep. It can record known durations directly (or through an injected clock) so the assertion does not depend on timing.

## Acceptance criteria

- [ ] The slowest-operations test does not read the wall clock (`go test ./internal/performance/ -run '^TestAnalyzer_SlowestOperations$' -count=20 -cpu=1 | grep -q '^ok'`)
