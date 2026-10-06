---
title: "TestSubjectTimeoutKillsGrandchildren flakes under load: pidfile \"\" read before written"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T00:14:57.380667974Z
updated: 2026-10-06T00:14:57.380667974Z
---

## Context

- paths: `internal/parseoracle/subject_test.go:231` (`TestSubjectTimeoutKillsGrandchildren`)
- commands: `make test`; `go test ./internal/parseoracle/ -run '^TestSubjectTimeoutKillsGrandchildren$' -count=20`

`TestSubjectTimeoutKillsGrandchildren` failed `make test` twice on 2026-10-05/06 (during the type-libraries milestone, issues 01a0fd9b-4e65 and 01a10177-49d9), each time with `pidfile "": strconv.Atoi`, while another agent's `make test` ran concurrently. It passed alone 3 times each time and the re-run suite was green. The empty pidfile suggests the test reads the grandchild's pidfile before the grandchild has written it -- a race the test, not the code under test, may own. Reproduce under load before fixing; the fix must remove the race, not lengthen a sleep.

## Steps

- [ ] Reproduce: run the test with `-count=50` under parallel CPU load (e.g. alongside `go test ./internal/parse/`); record the failure rate
- [ ] Find where the pidfile is read and what guarantees it is written first; fix the ordering
- [ ] Re-run under the same load with `-count=50`: zero failures; `make test`; commit

## Acceptance criteria

- [ ] the test survives 50 runs (`go test ./internal/parseoracle/ -run '^TestSubjectTimeoutKillsGrandchildren$' -count=50`)
