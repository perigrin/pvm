---
title: "LazyValue.Load loads twice: the write-lock double-check ignores an in-flight load"
state: pending
urgency: normal
milestone: v0.1
created: 2026-09-30T17:08:31.003430451Z
updated: 2026-09-30T17:08:31.003430451Z
---

TestLazyValueConcurrency fails intermittently, as in `make test` on 2026-09-30. The failure: "Expected load count 1, got 9", with each goroutine getting a different value. It reproduced once in 300 runs of `go test ./internal/memory -run TestLazyValueConcurrency -count=300`.

Cause: in LazyValue.Load, once a goroutine takes the write lock, the double-check tests `loaded` but not `loading`. A loader runs loadFunc with the lock released and `loaded` still false. So a second goroutine that takes the lock in that window passes the check, and starts its own load.

## Acceptance criteria

- [ ] Load waits for an in-flight load (`sh -c 'go test ./internal/memory -run "^TestLazyValueConcurrency$" -count=1000 2>&1 | grep -q "^ok"'`)
