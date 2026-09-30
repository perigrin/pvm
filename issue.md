---
title: "LazyValue.Load loads twice: the write-lock double-check ignores an in-flight load"
state: in-progress
urgency: normal
milestone: v0.1
created: 2026-09-30T17:08:31.003430451Z
updated: 2026-09-30T17:08:33.182924704Z
sessions:
- start_sha: 4aff87bf40eedf08eb1e0b1db5cb540885f609d4
  end_sha: ""
  commits: 0
  started_at: 2026-09-30T17:08:33.182924704Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-30T17:08:33.182924704Z
---

TestLazyValueConcurrency fails intermittently, as in `make test` on 2026-09-30. The failure: "Expected load count 1, got 9", with each goroutine getting a different value. It reproduced once in 300 runs of `go test ./internal/memory -run TestLazyValueConcurrency -count=300`.

Cause: in LazyValue.Load, once a goroutine takes the write lock, the double-check tests `loaded` but not `loading`. A loader runs loadFunc with the lock released and `loaded` still false. So a second goroutine that takes the lock in that window passes the check, and starts its own load.

## Acceptance criteria

- [ ] Load waits for an in-flight load (`sh -c 'go test ./internal/memory -run "^TestLazyValueConcurrency$" -count=1000 2>&1 | grep -q "^ok"'`)
