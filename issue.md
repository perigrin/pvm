---
title: "TestAskCancellation fails under load: Ask did not return after its context was cancelled"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-12T23:23:38.853610995Z
updated: 2026-10-12T23:23:38.853610995Z
---

## Context

- paths: `internal/parseoracle/oracle_test.go` (`TestAskCancellation`), `internal/parseoracle/oracle.go` (`AskFile`, `killDelay`)
- commands: `make test`; `go test ./internal/parseoracle/ -run '^TestAskCancellation$' -count=50`

Found 2026-10-12 by the m2 psc-analyze agent: `TestAskCancellation` failed one `make test` ("Ask did not return after its context was cancelled") while another session's `make test` ran on the same machine, and passed 3 of 3 alone. Same shape as 01a10e90 (TestSubjectTimeoutKillsGrandchildren), which raced a fixed timeout against process startup under load: reproduce it under load, find the race, and remove it -- do not lengthen a timeout.

## Steps

- [ ] Reproduce under load (e.g. a deliberately slow start, as 01a10e90 did); fix the ordering, not the timeout; 50 of 50 under load; `make test`; commit

## Acceptance criteria

- [ ] the test survives 50 runs (`go test ./internal/parseoracle/ -run '^TestAskCancellation$' -count=50 -v | grep -c '^--- PASS: TestAskCancellation' | grep -qx 50`)
