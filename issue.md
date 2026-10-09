---
title: The conformance package takes ~284s, so verify's 300s per-command default kills m1's conformance criteria
state: pending
urgency: normal
milestone: v0.1
created: 2026-10-09T16:07:28.134352691Z
updated: 2026-10-09T16:07:28.134352691Z
---

## Context

- paths: `internal/conformance/*_test.go` (`TestCorpusLints`, `TestMdtestTopicsMakeTheSameClaims`)
- commands: `go test ./internal/conformance/ -count=1 -v`; `git zhi verify m1-parse-the-core`

Found 2026-10-09 finishing m1. `go test ./internal/conformance/ -count=1` takes ~284s on this 24-core machine alone (TestCorpusLints 101s, TestMdtestTopicsMakeTheSameClaims 52s, the rest small). `git zhi verify` kills each criterion command at 300s by default (`--timeout`), and eight done m1 issues have the criterion `go test ./internal/conformance/ -count=1 | grep -q '^ok'`, so a default verify fails one of them intermittently (seen twice: 01a0dd43's, then 01a0d087's) with no test failing. m1 was verified with `--timeout 900`. Fix the cause: make the slow tests fast (measure first -- what do they spend the time on?), or if they must be slow, give those criteria a narrower test. Do not raise a default to hide it.

## Steps

- [ ] Profile TestCorpusLints and TestMdtestTopicsMakeTheSameClaims; make the package run well under 300s (or narrow the criteria with perigrin's agreement); show a default `git zhi verify m1-parse-the-core` passes

## Acceptance criteria

- [ ] the conformance package runs in under 120 seconds (`sh -c 's=$(date +%s); go test ./internal/conformance/ -count=1 | grep -q "^ok" || exit 1; test $(( $(date +%s) - s )) -lt 120'`)
