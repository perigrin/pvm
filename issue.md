---
title: TestAnalyzer_SlowestOperations orders sleeps by wall clock and fails under load
state: in-progress
urgency: normal
milestone: v0.1
created: 2026-09-29T09:12:59.621659274Z
updated: 2026-09-29T16:10:03.331405294Z
sessions:
- start_sha: 2d41ac8096c78c540a020722ba283b3528f9e495
  end_sha: ""
  commits: 0
  started_at: 2026-09-29T16:10:03.331405294Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-29T16:10:03.331405294Z
---

`internal/performance/analyzer_test.go`'s TestAnalyzer_SlowestOperations orders three operations by wall-clock time -- sleeps of 1ms, 10ms and 50ms -- and asserts the order. Under load the scheduler can stretch the 1ms sleep past the 10ms one, and `make test` fails on a package no change touched:

    --- FAIL: TestAnalyzer_SlowestOperations (0.08s)
        analyzer_test.go:130: Expected medium_op second, got fast_op

Seen once on 2026-09-29 while another package's tests ran in a second worktree; `go test ./internal/performance -count=5` alone passes.

The test is about the analyzer's ORDERING, not about sleep. It can record known durations directly -- backdating each measurement's start, since the test is in the package -- so the assertion does not depend on sleep. An injected clock in the analyzer would remove the wall clock entirely; nothing else needs one.

## Acceptance criteria

- [ ] The slowest-operations test does not sleep: each measurement is backdated by its duration, so the ordering does not depend on the scheduler (`go test ./internal/performance/ -run '^TestAnalyzer_SlowestOperations$' -count=20 -cpu=1 | grep -q '^ok'`)
