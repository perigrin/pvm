---
title: The oracle counts a cloned anon sub's ops once, not twice
state: in-progress
urgency: normal
milestone: m1-parse-the-core
blocks:
- 01a0a8cd-8fcc-72a8-9ba1-c916eee9992d
created: 2026-10-09T05:00:09.069779735Z
updated: 2026-10-09T05:00:57.604149967Z
sessions:
- start_sha: a341ba06c04d14ad11bda95e66345ad83f611576
  end_sha: ""
  commits: 0
  started_at: 2026-10-09T05:00:57.604149967Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-09T05:00:57.604149967Z
---

## Context

- commands: `make test` (green before every commit; never `--no-verify`); the oracle's corpus ratchet `go test ./internal/parseoracle/ -run '^TestRatchetCorpus$' -count=1 -v -parseoracle.corpus` (~2 min; regenerate with `-parseoracle.update` in the same commit when verdicts move, listing every moved file); the parse ratchets `go test ./internal/parse/ -run '^(TestParseRatchet|TestPerlGitTRatchet|TestCanonRatchet|TestT2HoldsAll56)$' -count=1`
- perl 5.42: `/home/perigrin/.local/bin/perl`; corpus: perl.git t/ through the shim (see `internal/parseoracle/ratchet_test.go`); the oracle's markers and verdicts: `internal/parseoracle/compare_facts.go`, docs/specs/perl-parser/00-findings.md §0.11
- the subject's sites: `internal/parse/sites.go` (`decidedMarkers`, `hidesStructure`), `internal/parse/cmd/subject`
- NEVER use pkill or killall

Found by the WRONG triage of 2026-10-09 at 98d02849 (139 WRONG findings across the 27 WRONG files of the regenerated baseline, each cause confirmed by a minimal repro run through the oracle). Goal: WRONG = 0 measured against internal/parse (01a0a8cd).

**O, an oracle bug.** parse_facts.pl dedupes CVs by `$$cv`; a cloned anon sub (a closure, or any anon sub under `$^P`) is a different CV sharing the prototype's ROOT, and `use overload` installs the clone in the stash, so the walk visits the same ops twice. Repro: `package P {\n my @a;\n use overload q(.) => sub { push @a, \$_[1]; $_[0] };\n}` -> perl srefgen [3,3]; non-closure [3]. Files: op/coreamp.t anonhash:23 srefgen:29 (the file sets `$^P |= 0x100`), op/concat2.t:31. Fix: dedupe the walk on `${$cv->ROOT}`; a trial cleared all three. This corrects perl's count, it is not a hedge.

## Steps

- [ ] Write failing test `TestOracleCountsClonedSubOnce` (the repro, via AskFile); dedupe on ROOT in `internal/parseoracle/testdata/parse_facts.pl`; re-baseline, listing moved files; `make test`; commit

## Acceptance criteria

- [ ] a cloned anon sub's ops are counted once (`go test ./internal/parseoracle/ -run '^TestOracleCountsClonedSubOnce$' -count=1 -v | grep -q '^--- PASS: TestOracleCountsClonedSubOnce'`)
