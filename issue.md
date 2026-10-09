---
title: A postfix hash slice ->@{...} / ->%{...} reports perl's rv2hv
state: pending
urgency: normal
milestone: m1-parse-the-core
blocks:
- 01a0a8cd-8fcc-72a8-9ba1-c916eee9992d
created: 2026-10-09T05:00:09.225476484Z
updated: 2026-10-09T05:00:36.253902127Z
---

## Context

- commands: `make test` (green before every commit; never `--no-verify`); the oracle's corpus ratchet `go test ./internal/parseoracle/ -run '^TestRatchetCorpus$' -count=1 -v -parseoracle.corpus` (~2 min; regenerate with `-parseoracle.update` in the same commit when verdicts move, listing every moved file); the parse ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet|TestT2HoldsAll56)$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; corpus: perl.git t/ through the shim (see `internal/parseoracle/ratchet_test.go`); the oracle's markers and verdicts: `internal/parseoracle/compare_facts.go`, docs/specs/perl-parser/00-findings.md §0.11
- the subject's sites: `internal/parse/sites.go` (`decidedMarkers`, `hidesStructure`), `internal/parse/cmd/subject`
- NEVER use pkill or killall

Found by the WRONG triage of 2026-10-09 at 98d02849 (139 WRONG findings across the 27 WRONG files of the regenerated baseline, each cause confirmed by a minimal repro run through the oracle). Goal: WRONG = 0 measured against internal/parse (01a0a8cd).

**H2.** `->@{...}` and `->%{...}` always emit rv2hv in perl; sites.go has no case for them though it decides the equivalent `@$r{...}`. Repro: `my $h; my @x = $h->@{1,2};`, `my @x = {1..4}->%{1,3};`. Files: op/postfixderef.t:148,316.

## Steps

- [ ] Write failing test `TestPostfixHashSliceReportsRv2hv`; implement; re-baseline; `make test`; commit

## Acceptance criteria

- [ ] a postfix hash slice reports rv2hv (`go test ./internal/parse/ -run '^TestPostfixHashSliceReportsRv2hv$' -count=1 -v | grep -q '^--- PASS: TestPostfixHashSliceReportsRv2hv'`)
