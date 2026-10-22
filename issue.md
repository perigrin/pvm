---
title: Hold the parse-without-error floor on the 593 files perl compiles
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
created: 2026-10-22T23:46:33.601895885Z
updated: 2026-10-22T23:56:40.353216281Z
---

## Context

- paths: `internal/parse/compiled_test.go` (new)
- commands: `PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestParseWithoutErrorOnCompiledCorpus$' -count=1 -v`
- entrypoints: `internal/parse/perlgitt_test.go:114` (TestPerlGitTRatchet, `perlGitTRoots`, `perlRejects`), `internal/parseoracle/testdata/ratchet/baseline.txt`

M3's "parse without error" metric already holds: of the 593 files perl compiles, 592 parse with 0 Unknown, and comp/final_line_num.t is refused on the same line perl rejects. Nothing commits that number. TestParseWithoutErrorOnCompiledCorpus reads the graded set from `../parseoracle/testdata/ratchet/baseline.txt` (status neither excluded nor error), asserts it is exactly 593 files, parses each through `Session.ParseFileFrom`, and asserts that >= 90% (>= 534) have 0 Unknown. A refusal in internal/parse is an Unknown node, so 0 Unknown means parsed without error. Test-only; no parser change.

## Steps

- [ ] Create `internal/parse/compiled_test.go` with two `// ABOUTME:` lines.
- [ ] RED: write the test with the graded-set count asserted as 593, then run it once against a deliberately wrong expected count (594) and watch it fail, so the count check is shown to bite.
- [ ] GREEN: set 593 and implement the parse loop, reusing TestPerlGitTRatchet's helpers rather than copying them. Log the clean count and every unclean file by name. Run it and watch it pass.
- [ ] Run `make test`, then commit, quoting the measured clean count in the message.

## Acceptance criteria

### Positive Scenarios

- [ ] At least 90% of the 593 perl-compiled files parse with 0 Unknown (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestParseWithoutErrorOnCompiledCorpus$' -count=1 -v | grep -q '^--- PASS: TestParseWithoutErrorOnCompiledCorpus'`)

### Negative Scenarios

- [ ] Reading the graded set fails when baseline.txt is missing, empty or unparseable, rather than yielding 0 files and a 0/0 pass (`go test ./internal/parse/ -run '^TestGradedSetFailsWithoutBaseline$' -count=1 -v | grep -q '^--- PASS: TestGradedSetFailsWithoutBaseline'`)
- [ ] The graded-set reader keeps exact, wider and no-answer rows and drops only excluded and error rows, checked on a fixture baseline holding one row of each status (`go test ./internal/parse/ -run '^TestGradedSetExcludesOnlyExcludedAndError$' -count=1 -v | grep -q '^--- PASS: TestGradedSetExcludesOnlyExcludedAndError'`)
- [ ] A graded file absent from PERL5_CORPUS (wrong pin, partial checkout) counts as unclean and is named, never skipped out of the 593 (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestCompiledCorpusMissingFileCountsAsUnclean$' -count=1 -v | grep -q '^--- PASS: TestCompiledCorpusMissingFileCountsAsUnclean'`)
- [ ] The 90% floor cannot hide a swap or a slide from 592 to 534: the test pins the unclean set, failing when any file outside it (today only comp/final_line_num.t) has an Unknown, and when a listed file becomes clean (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestParseWithoutErrorPinsUncleanSet$' -count=1 -v | grep -q '^--- PASS: TestParseWithoutErrorPinsUncleanSet'`)
- [ ] comp/final_line_num.t is refused on the line perl rejects; a refusal on any other line fails (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlRejectedFileRefusedOnPerlsLine$' -count=1 -v | grep -q '^--- PASS: TestPerlRejectedFileRefusedOnPerlsLine'`)
