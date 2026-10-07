---
title: "catch, isa and method have a () prototype but no call reaches them: type them or exclude them"
state: done
urgency: normal
milestone: type-libraries
created: 2026-10-06T14:54:03.825712198Z
updated: 2026-10-07T16:24:47.638023598Z
sessions:
- start_sha: 25a995c59df828540f32cd08fdd2e41998b1b78f
  end_sha: 25a995c59df828540f32cd08fdd2e41998b1b78f
  commits: 0
  started_at: 2026-10-07T16:24:47.534612932Z
  ended_at: 2026-10-07T16:24:47.638023598Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-07T16:24:47.534612932Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-07T16:24:47.638023598Z
---

## Context

- paths: `internal/parse/declarations/CORE.pmt` (catch, isa, method), `internal/parse/coreprocess_test.go` (`uncallableKeywords`), RFC 0001 ("CORE.pmt: the builtin table")
- commands: `make test`; perl 5.42 `/home/perigrin/.local/bin/perl`

From 01a10c6e-988f-7aa6-811d-0e462abc442a, which typed every other builtin of its batch. perl 5.42 reports prototype `()` for `catch`, `isa` and `method`, so CORE.pmt carries a line for each -- but none is reachable as a call: `CORE::isa()` and `CORE::catch()` are syntax errors, `CORE::method()` dies "Cannot 'method' outside of a 'class'", and `&CORE::isa`, `&CORE::catch`, `&CORE::method` are undefined subroutines. They are syntax keywords that `prototype()` happens to answer for. With nothing callable to measure, they are held prototype-only (the test asserts so). isa is separately typed as the infix operator; an anonymous `method { 1 }` inside a class is a CODE ref. The milestone goal requires every builtin typed except the six glob-slot lines.

## Steps

- [ ] perigrin decides: type them (e.g. `() None`, since no call returns), exclude them from "every builtin typed" as keywords that are not functions (recorded in the RFC beside the glob-slot exception), or another spelling
- [ ] Apply it with a failing test; `make test`; commit

## Acceptance criteria

- [ ] catch, isa and method are either typed or recorded as excluded keywords (`go test ./internal/parse/ -run '^TestCoreUncallableKeywordsSettled$' -count=1 -v | grep -q '^--- PASS: TestCoreUncallableKeywordsSettled'`)
