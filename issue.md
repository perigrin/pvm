---
title: An empty {} at statement position is an anon hash, as toke.c reads it
state: pending
urgency: normal
milestone: m1-parse-the-core
created: 2026-10-09T05:00:09.531437085Z
updated: 2026-10-09T05:00:09.531437085Z
---

## Context

- commands: `make test` (green before every commit; never `--no-verify`); the oracle's corpus ratchet `go test ./internal/parseoracle/ -run '^TestRatchetCorpus$' -count=1 -v -parseoracle.corpus` (~2 min; regenerate with `-parseoracle.update` in the same commit when verdicts move, listing every moved file); the parse ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet|TestT2HoldsAll56)$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; corpus: perl.git t/ through the shim (see `internal/parseoracle/ratchet_test.go`); the oracle's markers and verdicts: `internal/parseoracle/compare_facts.go`, docs/specs/perl-parser/00-findings.md §0.11
- the subject's sites: `internal/parse/sites.go` (`decidedMarkers`, `hidesStructure`), `internal/parse/cmd/subject`
- NEVER use pkill or killall

Found by the WRONG triage of 2026-10-09 at 98d02849 (139 WRONG findings across the 27 WRONG files of the regenerated baseline, each cause confirmed by a minimal repro run through the oracle). Goal: WRONG = 0 measured against internal/parse (01a0a8cd).

**B.** perl reads `{` followed by `}` (whitespace/comments allowed) at statement position as an anon hash (toke.c; spec §4.9.2); the subject reads a block. Repro: `my $h={a=>1};\n{}` -> perl anonhash (emptyavhv/ANONHASH). File: comp/parser.t anonhash:125 (the innermost of `{{{...}}}` at 120). Measure perl on `{ }`, `{ # c\n }`, nested forms.

## Steps

- [ ] Measure; write failing test `TestEmptyBracesAtStatementAreAnonHash` (negative: `{ 1 }` stays a block); implement in the parser; ratchets toward perl only; re-baseline; `make test`; commit

## Acceptance criteria

- [ ] empty braces at statement position are an anon hash (`go test ./internal/parse/ -run '^TestEmptyBracesAtStatementAreAnonHash$' -count=1 -v | grep -q '^--- PASS: TestEmptyBracesAtStatementAreAnonHash'`)
