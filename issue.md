---
title: Smartmatch against an array or hash reports the reference perl takes
state: in-progress
urgency: normal
milestone: m1-parse-the-core
blocks:
- 01a0a8cd-8fcc-72a8-9ba1-c916eee9992d
created: 2026-10-09T05:00:09.762705181Z
updated: 2026-10-09T07:15:14.631817417Z
sessions:
- start_sha: d7c31487939158225e0589373b9e583fe7ed61fe
  end_sha: ""
  commits: 0
  started_at: 2026-10-09T07:15:14.631817417Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T07:15:14.631817417Z
---

## Context

- commands: `make test` (green before every commit; never `--no-verify`); the oracle's corpus ratchet `go test ./internal/parseoracle/ -run '^TestRatchetCorpus$' -count=1 -v -parseoracle.corpus` (~2 min; regenerate with `-parseoracle.update` in the same commit when verdicts move, listing every moved file); the parse ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet|TestT2HoldsAll56)$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; corpus: perl.git t/ through the shim (see `internal/parseoracle/ratchet_test.go`); the oracle's markers and verdicts: `internal/parseoracle/compare_facts.go`, docs/specs/perl-parser/00-findings.md §0.11
- the subject's sites: `internal/parse/sites.go` (`decidedMarkers`, `hidesStructure`), `internal/parse/cmd/subject`
- NEVER use pkill or killall

Found by the WRONG triage of 2026-10-09 at 98d02849 (139 WRONG findings across the 27 WRONG files of the regenerated baseline, each cause confirmed by a minimal repro run through the oracle). Goal: WRONG = 0 measured against internal/parse (01a0a8cd).

**M.** Smartmatch takes an implicit reference to an array or hash operand: `when(@l)`, `1 ~~ @l` -> `\@l` (Deparse: `1 ~~ \@l`). Files: op/switch.t:1101,1120.

## Steps

- [ ] Write failing test `TestSmartmatchAggregateReportsReference`; implement; re-baseline; `make test`; commit

## Acceptance criteria

- [ ] smartmatch on an aggregate reports its reference (`go test ./internal/parse/ -run '^TestSmartmatchAggregateReportsReference$' -count=1 -v | grep -q '^--- PASS: TestSmartmatchAggregateReportsReference'`)
