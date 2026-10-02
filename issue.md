---
title: The performance analyzer times its own ReadMemStats pause
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-02T05:56:24.88598886Z
updated: 2026-10-02T05:56:24.88598886Z
---

`performance.Analyzer.StartMeasurement` (internal/performance/analyzer.go) reads the clock first and then calls `getCurrentMemoryUsage`, which is `runtime.ReadMemStats`, a stop-the-world pause. `Finish` likewise reads the clock and then the memory. So every measured duration includes the analyzer's own bookkeeping on the start side, which under load can be milliseconds. Taking `startTime` after the memory and goroutine reads would exclude it.

Found while diagnosing a TestAnalyzer_SlowestOperations failure under make test (`Expected medium_op second, got fast_op`), which 0 of 300 isolated runs and 0 of 300 loaded runs reproduced; that test's margins were widened separately, since any stall between the clock reads (not only this one) could flip them.

## Acceptance criteria

- [ ] StartMeasurement reads the clock after its bookkeeping (`go test ./internal/performance/ -run '^TestMeasurementExcludesBookkeeping$' -count=1 -v | grep -q '^--- PASS: TestMeasurementExcludesBookkeeping'`)
