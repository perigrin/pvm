---
title: Ratchet integration, file-name-keyed so a tier move is not a regression
state: in-progress
urgency: normal
milestone: m3-conformance-corpus
blocked_by:
- 01a0c35e-1b37-7304-ba39-65a6ef57d4e0
blocks:
- 01a0c360-01bf-7eab-b864-7d902d36fb97
created: 2026-09-21T09:50:00.652020362Z
updated: 2026-09-21T19:49:24.86582283Z
sessions:
- start_sha: 6036de4e5427d522f018fc1af4fd59dc3d760c06
  end_sha: ""
  commits: 0
  started_at: 2026-09-21T19:49:24.86582283Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T19:49:24.86582283Z
---

The corpus is a moving artifact by design: files get added because something
failed, and files move between tiers when measurement contradicts placement.
Both break a path-keyed baseline.

`internal/parse/testdata/t1.ratchet` is the existing mechanism. The corpus
needs the same guarantee -- movement in either direction fails -- keyed so
the two legitimate kinds of churn do not read as regressions.

## Why file-name-keyed rather than path-keyed

A file moving from `04_operators/` to `02_variables/` because the sweep
proved it belongs earlier is **the sweep working**. Path-keyed, that is a
deletion plus an addition, and the baseline reports a regression for a file
that did not change. The numbering is explicitly regenerable so tiers CAN be
reordered; a path-keyed ratchet would make the spec's own stated plan look
like breakage every time it happened.

## Acceptance

- [ ] A file renamed between tiers produces no regression in the baseline (`go test ./internal/conformance/ -run '^TestTierMoveIsNotARegression$' -count=1 -v | grep -q '^--- PASS: TestTierMoveIsNotARegression'`)
- [ ] A file that stops parsing fails the baseline, and so does one that starts parsing while still marked refusing (`go test ./internal/conformance/ -run '^TestRatchetFailsBothDirections$' -count=1 -v | grep -q '^--- PASS: TestRatchetFailsBothDirections'`)
- [ ] Adding a corpus file does not require regenerating the baseline by hand (`go test ./internal/conformance/ -run '^TestNewFileNeedsNoBaselineEdit$' -count=1 -v | grep -q '^--- PASS: TestNewFileNeedsNoBaselineEdit'`)
- [ ] A duplicate file name across tiers is rejected, since the key assumes uniqueness (`go test ./internal/conformance/ -run '^TestDuplicateFileNameRejected$' -count=1 -v | grep -q '^--- PASS: TestDuplicateFileNameRejected'`)
