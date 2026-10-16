---
title: Benchmark the lowering and analysis latency the parser plan asks for
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T01:16:38.234468817Z
updated: 2026-10-16T01:16:38.234468817Z
---

## Context

- found by: the agentic review of PR #486 (plan alignment), 2026-10-16: a deferral with no tracking issue

docs/specs/perl-parser/06-incremental-lsp.md §6.1.6 ("What would change the decision") and §6.11 step 4 ask for a measured lowering budget. No Benchmark* exists in internal/lower, internal/infer, internal/parse or internal/psc. Measure parse, lower and Analyze per file over T1 and perl lib, and record the p50/p99 against the plan's budget.

## Steps

- [ ] Add benchmarks for parse, lower and Analyze; record numbers in the spec; `make test`; commit

## Acceptance criteria

- [ ] a lowering benchmark exists (`go test ./internal/lower/ -run '^BenchmarkLowerT1$' -count=1 -v | grep -q '^--- PASS: BenchmarkLowerT1'`)
