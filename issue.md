---
title: Hold the parse-without-error floor on the 593 files perl compiles
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a1664d-69f0-715c-b060-c8f648ddfcd3
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
created: 2026-10-22T23:46:33.601895885Z
updated: 2026-10-23T01:09:00.107305943Z
---

## Context

- paths: `internal/parse/compiled_test.go` (new), `internal/parse/testdata/compiledclean.txt` (new), `.github/workflows/ci.yml` (the skip-guard loop, ~207)
- commands: `PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestParseWithoutErrorOnCompiledCorpus$' -count=1 -v`
- entrypoints: `internal/parse/perlgitt_test.go:114` (TestPerlGitTRatchet, `perlGitTRoots`, `perlRejects`), `internal/parseoracle/testdata/ratchet/baseline.txt`, `internal/parseoracle/testdata/ratchet/m3start.txt`

M3's "parse without error" metric is "≥ 90% of the compilable corpus". Under 5.42, "593" was the graded set, not a compile count: six of those files did not compile under 5.42. Under the 5.44.0 pin, perl compiles all 594 non-excluded files, and the re-pin issue records the compiled count in m3start.txt (`compiled`). TestParseWithoutErrorOnCompiledCorpus reads the graded set from `../parseoracle/testdata/ratchet/baseline.txt` (status neither excluded nor error). It asserts the set's size equals m3start.txt's `compiled`, naming any difference. It parses each file through `Session.ParseFileFrom` and asserts that >= 90% have 0 Unknown. A refusal in internal/parse is an Unknown node, so 0 Unknown means parsed without error. It writes the clean count and the unclean files to `testdata/compiledclean.txt` (header `# N/M clean`), which the docs issue quotes. Test-only; no parser change.

## Steps

- [ ] Judge comp/final_line_num.t with the pinned perl first. The oracle reports "perl reported success but produced no optree" while `perlRejects` treats it as rejected. Quote `perl -c`'s exit status and output at 5.44.0 in the test's comment, and make the test treat the file the way perl does.
- [ ] Create `internal/parse/compiled_test.go` with two `// ABOUTME:` lines.
- [ ] RED: write the test reading the expected size from m3start.txt. Run it once against a fixture m3start with the count off by one, and watch it fail, so the count check is shown to bite.
- [ ] GREEN: implement the parse loop, reusing TestPerlGitTRatchet's helpers rather than copying them. Write compiledclean.txt under the existing update flag, and fail when the tree disagrees with it. Run the test and watch it pass.
- [ ] Add TestParseWithoutErrorOnCompiledCorpus to ci.yml's "Fail if a corpus-backed ratchet skipped" loop, and confirm that the job sets the corpus env the test needs.
- [ ] Run `make test`, then commit, quoting the measured clean count in the message.

## Acceptance criteria

### Positive Scenarios

- [ ] At least 90% of the files perl 5.44.0 compiles parse with 0 Unknown, and the graded set matches the recorded compile count (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestParseWithoutErrorOnCompiledCorpus$' -count=1 -v | grep -q '^--- PASS: TestParseWithoutErrorOnCompiledCorpus'`)
- [ ] The clean count is committed for the docs to quote (`grep -Eq '^# [0-9]+/[0-9]+ clean' internal/parse/testdata/compiledclean.txt`)
- [ ] CI fails if the test skips (`grep -q TestParseWithoutErrorOnCompiledCorpus .github/workflows/ci.yml`)

### Negative Scenarios

- [ ] Reading the graded set fails when baseline.txt is missing, empty or unparseable, rather than yielding 0 files and a 0/0 pass (`go test ./internal/parse/ -run '^TestGradedSetFailsWithoutBaseline$' -count=1 -v | grep -q '^--- PASS: TestGradedSetFailsWithoutBaseline'`)
- [ ] The graded-set reader keeps exact, wider and no-answer rows and drops only excluded and error rows, checked on a fixture baseline holding one row of each status (`go test ./internal/parse/ -run '^TestGradedSetExcludesOnlyExcludedAndError$' -count=1 -v | grep -q '^--- PASS: TestGradedSetExcludesOnlyExcludedAndError'`)
- [ ] A graded file absent from PERL5_CORPUS (wrong pin, partial checkout) counts as unclean and is named, never skipped out of the graded set (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestCompiledCorpusMissingFileCountsAsUnclean$' -count=1 -v | grep -q '^--- PASS: TestCompiledCorpusMissingFileCountsAsUnclean'`)
- [ ] The 90% floor cannot hide a swap or a slide down to the 90% line: the test pins the unclean set, failing when any file outside it (today only comp/final_line_num.t) has an Unknown, and when a listed file becomes clean (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestParseWithoutErrorPinsUncleanSet$' -count=1 -v | grep -q '^--- PASS: TestParseWithoutErrorPinsUncleanSet'`)
- [ ] comp/final_line_num.t is refused on the line perl rejects; a refusal on any other line fails (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlRejectedFileRefusedOnPerlsLine$' -count=1 -v | grep -q '^--- PASS: TestPerlRejectedFileRefusedOnPerlsLine'`)
