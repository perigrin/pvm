---
title: TestMirrorSelectionStrategies orders 20 ms and 30 ms sleeps by wall clock and fails under load
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-29T13:54:19.462807812Z
updated: 2026-09-29T13:54:19.462807812Z
---

`internal/cpan/mirror_manager_test.go`'s TestMirrorSelectionStrategies starts four local servers that sleep 50, 20, 80 and 30 ms and asserts StrategyFastest picks the 20 ms one. The 20 and 30 ms servers are 10 ms apart, and under load the scheduler can reorder them, so `make test` fails on a package no change touched:

    --- FAIL: TestMirrorSelectionStrategies (0.08s)
        mirror_manager_test.go:490: StrategyFastest should return fastest mirror http://127.0.0.1:45319, got http://127.0.0.1:41445

Seen once on 2026-09-29 while a `go vet` ran in another worktree; `go test ./internal/cpan -run TestMirrorSelectionStrategies -count=10` alone passes. The same shape as 01a0ec70-53e6-7676-8d0d-7aa04fd5cef3 (TestAnalyzer_SlowestOperations).

The test is about the strategy's choice, not about sleep. It can feed the manager measured latencies directly, or widen the gaps so ordering survives scheduling jitter.

## Acceptance criteria

- [ ] StrategyFastest's test does not depend on 10 ms of scheduling (`go test ./internal/cpan/ -run '^TestMirrorSelectionStrategies$' -count=20 -cpu=1 | grep -q '^ok'`)
