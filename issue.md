---
title: "The plan's M1 section states the gates the code enforces: the corpus ratchet, the exact-match ratchet, T2 at 56 of 56"
state: done
urgency: normal
milestone: m1-parse-the-core
blocked_by:
- 01a0c13f-e0e2-7b4f-a9d2-06f7edad1e7f
- 01a0a8cd-8fcc-72a8-9ba1-c916eee9992d
- 01a0b835-1270-7b32-9f0a-18fe7c555bae
created: 2026-10-09T03:25:53.001634674Z
updated: 2026-10-09T13:56:01.246271304Z
sessions:
- start_sha: 6c8caec9d5ca579d7bd396c7643edd89d935656b
  end_sha: 6c8caec9d5ca579d7bd396c7643edd89d935656b
  commits: 0
  started_at: 2026-10-09T13:56:01.04816327Z
  ended_at: 2026-10-09T13:56:01.246271304Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T13:56:01.04816327Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-09T13:56:01.246271304Z
---

## Context

- paths: `docs/plans/2026-09-05-parser-conformance-plan.md` (§5 "M1 -- Parses the core without error"), `internal/parse/t2_test.go`, `internal/parseoracle/ratchet_test.go`, `internal/conformance` (`TestCorpusRatchet`)
- commands: `make test`

The plan's M1 table still says "Parse without error 100% of T2", "≥ 70% of T1-easy", "Oracle exact ≥ 40% of markers attempted". The code has moved: `TestT2CoreParses` says the corpus ratchet (`TestCorpusRatchet`) is the gate, the oracle's per-file baseline ratchets `exact` in both directions, and perigrin's ruling (2026-10-09) is T2 at 56 of 56. Last in the chain: the numbers it states are the ones measured at the finish.

## Steps

- [ ] Rewrite the M1 section's table and prose to the gates the code enforces, with the numbers measured at the finish (T2, T1, exact-match starting rate, WRONG); `make test`; commit

## Acceptance criteria

- [ ] the plan names the corpus ratchet as M1's gate (`grep -q 'TestCorpusRatchet' docs/plans/2026-09-05-parser-conformance-plan.md`)
- [ ] the plan states T2 at 56 of 56 (`grep -q '56 of 56' docs/plans/2026-09-05-parser-conformance-plan.md`)
- [ ] the plan names the exact-match ratchet (`grep -qi 'exact-match ratchet' docs/plans/2026-09-05-parser-conformance-plan.md`)
