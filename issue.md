---
title: Hedge the sites inside interpolation internal/parse does not parse
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T11:07:58.648536842Z
updated: 2026-10-09T11:08:13.04361167Z
---

## Context

- paths: `internal/parse/sites.go` (how sites are decided or hedged), `internal/parseoracle/ratchet_test.go` (`knownWrong`), `internal/parseoracle/testdata/ratchet/baseline.txt`, `docs/specs/perl-parser/00-findings.md` §0.11
- commands: `make test`; the oracle sweep at `PARSEORACLE_TIMEOUT=8m` with `-parseoracle.corpus` (and `-parseoracle.update`)

Five of the nine files still WRONG at 13c30099/d5f43f87 commit to a site perl reads differently inside an interpolated string or regex: comp/parser.t, op/coreamp.t, op/postfixderef.t, op/sub_lval.t, re/reg_eval_scope.t. The plan's M1 rule decides the approach: a construct not yet understood is reported as unknown, not guessed at -- that is what keeps WRONG at zero. So internal/parse hedges every marker inside an interpolated span it does not parse, and the real parse is 01a12059-3011-7a53-8249-e3f154026f6f (no milestone). Re-measure each file's WRONG findings first; any that are not interpolation belong to another issue.

## Steps

- [ ] Re-measure the five files' WRONG findings; failing test first; hedge markers inside unparsed interpolation; re-baseline (each move WRONG to wider or better, none worse); shrink the knownWrong pin; findings §0.11 subsection; `make test`; commit

## Acceptance criteria

- [ ] a marker inside unparsed interpolation is hedged, not decided (`go test ./internal/parse/ -run '^TestInterpolationHedgesItsSites$' -count=1 -v | grep -q '^--- PASS: TestInterpolationHedgesItsSites'`)
