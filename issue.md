---
title: An empty {} at statement position is an anon hash, as toke.c reads it
state: done
urgency: normal
milestone: m1-parse-the-core
blocks:
- 01a0a8cd-8fcc-72a8-9ba1-c916eee9992d
created: 2026-10-09T05:00:09.531437085Z
updated: 2026-10-09T08:35:06.640931922Z
sessions:
- start_sha: d7c31487939158225e0589373b9e583fe7ed61fe
  end_sha: dd8abfe77cbc8881e7d098ea6f530d2fd51e4ed7
  commits: 3
  started_at: 2026-10-09T07:39:14.321870746Z
  ended_at: 2026-10-09T08:35:06.640931922Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T07:39:14.321870746Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-09T08:35:06.640931922Z
observed_paths:
- docs/specs/perl-parser/00-findings.md
- internal/parse/canon.go
- internal/parse/canon_test.go
- internal/parse/cmd/subject/main.go
- internal/parse/cmd/subject/main_test.go
- internal/parse/emptybraces_test.go
- internal/parse/parse.go
- internal/parse/perllines_test.go
- internal/parse/refusal.go
- internal/parse/refusal_test.go
- internal/parse/sites.go
- internal/parse/stmt_test.go
- internal/parse/term.go
- internal/parse/unclosed_test.go
- internal/parse/wrongtree_test.go
- internal/parseoracle/ratchet_test.go
- internal/parseoracle/testdata/ratchet/baseline.txt
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
