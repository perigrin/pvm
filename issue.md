---
title: A site spans to the end of its statement, where perl's nextstate puts it
state: in-progress
urgency: normal
milestone: m1-parse-the-core
blocks:
- 01a0a8cd-8fcc-72a8-9ba1-c916eee9992d
created: 2026-10-09T05:00:08.930045688Z
updated: 2026-10-09T05:18:52.828583469Z
sessions:
- start_sha: 93c26e8152e5d8e6b6914956ecf5014d163c7342
  end_sha: ""
  commits: 0
  started_at: 2026-10-09T05:18:52.828583469Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T05:18:52.828583469Z
---

## Context

- commands: `make test` (green before every commit; never `--no-verify`); the oracle's corpus ratchet `go test ./internal/parseoracle/ -run '^TestRatchetCorpus$' -count=1 -v -parseoracle.corpus` (~2 min; regenerate with `-parseoracle.update` in the same commit when verdicts move, listing every moved file); the parse ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet|TestT2HoldsAll56)$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; corpus: perl.git t/ through the shim (see `internal/parseoracle/ratchet_test.go`); the oracle's markers and verdicts: `internal/parseoracle/compare_facts.go`, docs/specs/perl-parser/00-findings.md §0.11
- the subject's sites: `internal/parse/sites.go` (`decidedMarkers`, `hidesStructure`), `internal/parse/cmd/subject`
- NEVER use pkill or killall

Found by the WRONG triage of 2026-10-09 at 98d02849 (139 WRONG findings across the 27 WRONG files of the regenerated baseline, each cause confirmed by a minimal repro run through the oracle). Goal: WRONG = 0 measured against internal/parse (01a0a8cd).

**S.** perl's nextstate line for a multi-line statement can be a later line, up to the closing `);`/`};` (copline moves past an anon sub, a `when` block, a bareword `$h{foo}`); `sites.go` ends each site at its own node. Repro: `my @a;\nmy @r = (\@a, sub {\n 1;\n});` -> perl line 4, subject 2..2. Files: comp/proto.t:540,544,589x2,594x2 (and P1's 532,536,580,584), japh/abigail.t anonhash:211, op/bop.t anonhash:706x6, op/switch.t match:209 srefgen:711, porting/dual-life.t:60. A one-line trial cleared 15 findings in these files and added none. Side effect measured: op/sprintf2.t:452 went WRONG -> wider because the subject pools every site under its top-level statement, so a wider span lets an unrelated hedge excuse more; the pooling itself is a separate question (filed apart), not this issue's.

## Steps

- [ ] Write failing test `TestSiteSpansToStatementEnd` from the repro; implement; re-baseline, listing moved files (report any WRONG -> wider move caused by pooling, as sprintf2.t:452); `make test`; commit

## Acceptance criteria

- [ ] a site spans to its statement end (`go test ./internal/parse/ -run '^TestSiteSpansToStatementEnd$' -count=1 -v | grep -q '^--- PASS: TestSiteSpansToStatementEnd'`)
