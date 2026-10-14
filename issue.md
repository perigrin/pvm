---
title: "TestBuildPerlCancellation fails under load: the build finishes before the cancel"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-14T20:41:35.947496891Z
updated: 2026-10-14T20:41:35.947496891Z
---

## Context

- paths: `internal/perl/*_test.go` (`TestBuildPerlCancellation`)
- commands: `go test ./internal/perl/ -run '^TestBuildPerlCancellation$' -count=50`

Found 2026-10-14 by the cutover-1 calls fixes: `TestBuildPerlCancellation` failed one `make test` ("An error is expected but got nil" on a 2.9s build) while three other agents' `make test` runs loaded the machine, and passed on every rerun. Same shape as 01a1326d (TestAskCancellation) and 01a10e90 (TestSubjectTimeoutKillsGrandchildren): a fixed timing raced against work under load. Reproduce under load and remove the race -- do not lengthen a timeout.

## Steps

- [ ] Reproduce under load; make cancellation deterministic (cancel before the build can finish, or block the build on a channel the test controls); 50 of 50 under load; `make test`; commit

## Acceptance criteria

- [ ] the test survives 50 runs (`go test ./internal/perl/ -run '^TestBuildPerlCancellation$' -count=50 -v | grep -c '^--- PASS: TestBuildPerlCancellation' | grep -qx 50`)
