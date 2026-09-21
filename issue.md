---
title: A refusal citation captures 13 of 36 characters, and UUIDv7 collides on the first 8
state: done
urgency: normal
milestone: m3-conformance-corpus
blocks:
- 01a0c35e-1b37-7304-ba39-65a6ef57d4e0
created: 2026-09-21T14:03:18.076018197Z
updated: 2026-09-21T22:36:19.848401017Z
sessions:
- start_sha: 0ce515cbfe0ba1f48eb9b741cb1fc5d8c389454d
  end_sha: a79a8f82a052e2b580b63516bef139bc0282466b
  commits: 1
  started_at: 2026-09-21T19:00:02.292865539Z
  ended_at: 2026-09-21T19:03:27.032386284Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-09-21T19:00:02.292865539Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-09-21T19:03:27.032386284Z
observed_paths:
- conformance/01_literals/03_leading_decimal.t
- conformance/11_oo/00_adjacency.t
- conformance/README.md
- internal/conformance/conformance_test.go
- internal/conformance/file.go
---

`internal/conformance/file.go` captures 13 of a 36-character issue id:

    \bIssue ([0-9a-f]{8}-[0-9a-f]{4})\b

## Measured against the project's own ids

Checked against the 27 unique ids in
`docs/plans/2026-09-21-deferred-chain-m1-m2.md`:

    8-char prefix     COLLIDES -- 6 issues share 01a0c13f, 4 share
                      01a0c115, and five more pairs collide
    13-char prefix    unique across all 27

So it works today, and it works for a reason that will not hold. These are
UUIDv7: the first 8 characters are a millisecond timestamp, which is
exactly why a batch-created chain collides there. Uniqueness rests on the
next 4 characters -- 16 bits. At 27 issues the birthday probability is
already around 0.5%, and this corpus is designed to grow.

## Why this is Minor rather than Critical

`f.Refuses` is consumed only as a display string: `run.go` interpolates it
into `t.Errorf` and `t.Skipf`. It is never compared, looked up, or matched
against the chain. A collision misleads a human reading test output; it
does not misroute anything.

That changes if the runner ever verifies that a cited issue resolves --
which the runner issue's acceptance criteria require. A 13-character
citation cannot be looked up unambiguously, so this blocks that criterion
rather than merely preceding it.

## The fix keeps existing files working

Accept the full id and keep the short form as a fallback:

    \bIssue ([0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12}|[0-9a-f]{8}-[0-9a-f]{4})\b

Existing files citing the short form keep parsing. `conformance/README.md`
and `conformance/01_literals/03_leading_decimal.t` should move to the full
id to set the precedent.

Found by the review gate on 1e70de78.

## Acceptance criteria

- [ ] A full 36-character issue id is captured whole (`go test ./internal/conformance/ -run '^TestRefusalIssueFullID$' -count=1 -v | grep -q '^--- PASS: TestRefusalIssueFullID'`)
- [ ] A short 13-character citation still parses, so existing files keep working (`go test ./internal/conformance/ -run '^TestRefusalIssueShortIDStillWorks$' -count=1 -v | grep -q '^--- PASS: TestRefusalIssueShortIDStillWorks'`)
- [ ] The corpus files and README cite full ids (`! grep -qE 'Issue [0-9a-f]{8}-[0-9a-f]{4}[^-]' conformance/README.md conformance/01_literals/*.t`)
- [ ] The whole corpus still runs green (`go test ./internal/conformance/ -count=1`)
