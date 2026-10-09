---
title: Hedge rv2hv for a ->{...} hash read whose multideref folding pvm cannot predict
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T11:07:58.763658294Z
updated: 2026-10-09T11:08:15.764399999Z
---

## Context

- paths: `internal/parse/sites.go` (rv2hv for `->{...}`), `internal/parseoracle/ratchet_test.go` (`knownWrong`), `internal/parseoracle/testdata/ratchet/baseline.txt`, `docs/specs/perl-parser/00-findings.md` §0.11
- commands: `make test`; the oracle sweep at `PARSEORACLE_TIMEOUT=8m` with `-parseoracle.corpus` (and `-parseoracle.update`)

Four of the nine files still WRONG commit to an rv2hv answer for a `->{...}` hash read that perl compiles otherwise (rv2hv vs multideref folding): japh/abigail.t, op/gmagic.t, op/ref.t, porting/libperl.t. Per the plan's M1 rule (unknown, not guessed), internal/parse hedges the rv2hv marker for the `->{...}` reads whose folding it cannot predict; the real rule is 01a12059-30a8-722e-bbc3-52066b8610cb (no milestone). Hedge only the shapes measured to be unpredictable -- the ones perl decides the same way every time stay decided, so exact files stay exact. Re-measure each file's WRONG findings first.

## Steps

- [ ] Re-measure the four files' WRONG findings and the forms with Concise; failing test first; hedge the unpredictable shapes; re-baseline (each move WRONG to wider or better, no exact file worse); shrink the knownWrong pin; findings §0.11 subsection; `make test`; commit

## Acceptance criteria

- [ ] an arrow hash read whose folding pvm cannot predict hedges its rv2hv (`go test ./internal/parse/ -run '^TestArrowHashReadHedgesRv2hv$' -count=1 -v | grep -q '^--- PASS: TestArrowHashReadHedgesRv2hv'`)
