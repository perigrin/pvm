---
title: "TestMirrorManagerSelectBestMirror flakes under load: the fastest local mirror is chosen by wall-clock latency"
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-06T13:34:39.22177952Z
updated: 2026-10-06T13:34:39.22177952Z
---

## Context

- paths: `internal/cpan/mirror_manager_test.go` (`TestMirrorManagerSelectBestMirror`)
- commands: `go test ./internal/cpan/ -run '^TestMirrorManagerSelectBestMirror$' -count=50`, under parallel load

On 2026-10-06, during the type-libraries milestone with several agents running test suites (load ~20), `make test` failed once: "StrategyFastest should select fastest mirror http://127.0.0.1:45751, got http://127.0.0.1:46329". Run alone 10 times it passes. The test picks the fastest of local test servers by measured latency, so scheduler delay under load can reorder them. The fix must not depend on wall-clock ordering being stable under load (e.g. widen the latency gap substantially, or inject the measured latencies), not lengthen a timeout.

## Steps

- [ ] Reproduce under parallel load with `-count=50`
- [ ] Remove the dependence on wall-clock ordering; re-run under load with no failure; `make test`; commit

## Acceptance criteria

- [ ] the test survives 50 runs (`go test ./internal/cpan/ -run '^TestMirrorManagerSelectBestMirror$' -count=50`)
