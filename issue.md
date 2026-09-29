---
title: TestMirrorSelectionStrategies orders 20 ms and 30 ms sleeps by wall clock and fails under load
state: done
urgency: normal
milestone: v0.1
created: 2026-09-29T13:54:19.462807812Z
updated: 2026-09-29T16:08:55.846302162Z
sessions:
- start_sha: 5436629d157e3c7ed7335113d825ff3f9b4ecd44
  end_sha: 2d41ac8096c78c540a020722ba283b3528f9e495
  commits: 1
  started_at: 2026-09-29T16:03:29.005253751Z
  ended_at: 2026-09-29T16:08:55.846302162Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T16:03:29.005253751Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-29T16:08:55.846302162Z
observed_paths:
- internal/cpan/mirror_manager_test.go
---

`internal/cpan/mirror_manager_test.go`'s TestMirrorSelectionStrategies starts four local servers that sleep 50, 20, 80 and 30 ms and asserts StrategyFastest picks the 20 ms one. The 20 and 30 ms servers are 10 ms apart, and under load the scheduler can reorder them, so `make test` fails on a package no change touched:

    --- FAIL: TestMirrorSelectionStrategies (0.08s)
        mirror_manager_test.go:490: StrategyFastest should return fastest mirror http://127.0.0.1:45319, got http://127.0.0.1:41445

Seen once on 2026-09-29 while a `go vet` ran in another worktree; `go test ./internal/cpan -run TestMirrorSelectionStrategies -count=10` alone passes. The same shape as 01a0ec70-53e6-7676-8d0d-7aa04fd5cef3 (TestAnalyzer_SlowestOperations).

The test is about the strategy's choice, not about sleep. It can feed the manager measured latencies directly, or widen the gaps so ordering survives scheduling jitter.

## Acceptance criteria

- [ ] StrategyFastest's test does not depend on 10 ms of scheduling (`go test ./internal/cpan/ -run '^TestMirrorSelectionStrategies$' -count=20 -cpu=1 | grep -q '^ok'`)
