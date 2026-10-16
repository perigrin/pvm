---
title: given/when, try/catch/finally, defer and format lower to a whole-statement Hole, so inference never reads their bodies
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-16T01:16:38.185891265Z
updated: 2026-10-16T01:16:38.185891265Z
---

## Context

- found by: the agentic review of PR #486 (plan alignment), 2026-10-16: a deferral with no tracking issue

internal/lower/stmt.go:150-155 lowers these statements to a single Hole (01a12a85-2fa3 scoped them out "until a consumer needs them"). Inference therefore never reads a try body: `try { my $x = 1 + "a"; } catch ($e) {}` is one Hole, and its diagnostic is lost. The vocabulary needs statement kinds for them (or a lowering into existing kinds).

## Steps

- [ ] Lower try/catch/finally, defer, given/when and format bodies to kinds inference walks; failing test first; `make test`; commit

## Acceptance criteria

- [ ] a try body is analysed (`go test ./internal/lower/ -run '^TestLowerTryBodyIsWalked$' -count=1 -v | grep -q '^--- PASS: TestLowerTryBodyIsWalked'`)
