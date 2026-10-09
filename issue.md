---
title: Hedge rv2hv for a ->{...} hash read whose multideref folding pvm cannot predict
state: done
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T11:07:58.763658294Z
updated: 2026-10-09T12:33:09.462513207Z
sessions:
- start_sha: 5ae2b3eb8cf8e8b2b63e86042dfbc046cf18a62c
  end_sha: a45869f9f87be4c7c68e6f15a2026fba6f7a9fd7
  commits: 2
  started_at: 2026-10-09T11:08:55.755849525Z
  ended_at: 2026-10-09T12:33:09.462513207Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T11:08:55.755849525Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-09T12:33:09.462513207Z
observed_paths:
- docs/specs/perl-parser/00-findings.md
- internal/parse/arrowhashread_test.go
- internal/parse/interphedge_test.go
- internal/parse/sites.go
- internal/parse/sites_decided_test.go
- internal/parseoracle/ratchet_test.go
- internal/parseoracle/testdata/ratchet/baseline.txt
---

## Context

- paths: `internal/parse/sites.go` (rv2hv for `->{...}`), `internal/parseoracle/ratchet_test.go` (`knownWrong`), `internal/parseoracle/testdata/ratchet/baseline.txt`, `docs/specs/perl-parser/00-findings.md` §0.11
- commands: `make test`; the oracle sweep at `PARSEORACLE_TIMEOUT=8m` with `-parseoracle.corpus` (and `-parseoracle.update`)

Four of the nine files still WRONG commit to an rv2hv answer for a `->{...}` hash read that perl compiles otherwise (rv2hv vs multideref folding): japh/abigail.t, op/gmagic.t, op/ref.t, porting/libperl.t. Per the plan's M1 rule (unknown, not guessed), internal/parse hedges the rv2hv marker for the `->{...}` reads whose folding it cannot predict; the real rule is 01a12059-30a8-722e-bbc3-52066b8610cb (milestone v0.1). Hedge only the shapes measured to be unpredictable -- the ones perl decides the same way every time stay decided, so exact files stay exact. Re-measure each file's WRONG findings first.

## Steps

- [ ] Re-measure the four files' WRONG findings and the forms with Concise; failing test first; hedge the unpredictable shapes; re-baseline (each move WRONG to wider or better, no exact file worse); shrink the knownWrong pin; findings §0.11 subsection; `make test`; commit

## Acceptance criteria

- [ ] an arrow hash read whose folding pvm cannot predict hedges its rv2hv (`go test ./internal/parse/ -run '^TestArrowHashReadHedgesRv2hv$' -count=1 -v | grep -q '^--- PASS: TestArrowHashReadHedgesRv2hv'`)
