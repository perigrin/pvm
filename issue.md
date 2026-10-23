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
updated: 2026-10-23T01:35:26.221677154Z
---

**The title is stale and git-zhi cannot retitle.** Read it as "Hold the parse-without-error floor on the compiled set recorded at the 5.44.0 pin". The floor is over compiled.txt, not over 593 files; 593 was the 5.42 graded set (chain-review round 2, A3/A13).

## Context

- paths: `internal/parse/compiled_test.go` (new), `internal/parseoracle/testdata/ratchet/compiled.txt` (read), `internal/parse/testdata/compiledclean.txt` (new), `.github/workflows/ci.yml` (the skip-guard loop, ~207)
- commands: `PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestParseWithoutErrorOnCompiledCorpus$' -count=1 -v`
- entrypoints: `internal/parse/perlgitt_test.go:114` (TestPerlGitTRatchet, `perlGitTRoots`, `perlRejects`), `internal/parseoracle/testdata/ratchet/baseline.txt`, `internal/parseoracle/testdata/ratchet/m3start.txt`

M3's "parse without error" metric is "≥ 90% of the compilable corpus", and goal (4) says it is measured against the current compile count. The re-pin issue records the compiled set by path in `../parseoracle/testdata/ratchet/compiled.txt`: files the pinned perl 5.44.0 accepts under `perl -c` (`-T` where the shebang has it), minus perl.git's `perlRejects`. TestParseWithoutErrorOnCompiledCorpus parses exactly that set through `Session.ParseFileFrom` and asserts that >= 90% have 0 Unknown. A refusal in internal/parse is an Unknown node, so 0 Unknown means parsed without error. Two files are named in the test's comment:
- re/pat_psycho.t IS in the set. perl compiles it, and parsing needs no oracle, even though the oracle sweep errors on it at 8m.
- comp/final_line_num.t is NOT in the set. It ends in `print 1+`, and its BEGIN `__DIE__` handler prints `ok 1` and exits 0, so `perl -c`'s rc 0 is the file's own trick.

The test logs, by name, every difference between the compiled set and baseline.txt's graded set. It never asserts the two are the same size. It writes the clean count and the unclean files to `testdata/compiledclean.txt` (header `# N/M clean`), which the docs issue quotes. Test-only; no parser change.

## Steps

- [ ] Create `internal/parse/compiled_test.go` with two `// ABOUTME:` lines.
- [ ] RED: write the test over compiled.txt, with its size checked against m3start.txt's `compiled`. Run it once against a fixture compiled.txt one line short of the count, and watch it fail, so the count check is shown to bite.
- [ ] GREEN: implement the parse loop, reusing TestPerlGitTRatchet's helpers rather than copying them. Write compiledclean.txt under the existing update flag, and fail when the tree disagrees with it. Run the test and watch it pass.
- [ ] Add TestParseWithoutErrorOnCompiledCorpus to ci.yml's "Fail if a corpus-backed ratchet skipped" loop, and confirm that the job sets the corpus env the test needs.
- [ ] Run `make test`, then commit, quoting the measured clean count in the message.

## Acceptance criteria

### Positive Scenarios

- [ ] At least 90% of the compiled set (compiled.txt, recorded with perl 5.44.0) parse with 0 Unknown (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestParseWithoutErrorOnCompiledCorpus$' -count=1 -v | grep -q '^--- PASS: TestParseWithoutErrorOnCompiledCorpus'`)
- [ ] The clean count is committed for the docs to quote (`grep -Eq '^# [0-9]+/[0-9]+ clean' internal/parse/testdata/compiledclean.txt`)
- [ ] CI fails if the test skips (`grep -q TestParseWithoutErrorOnCompiledCorpus .github/workflows/ci.yml`)

### Negative Scenarios

- [ ] Reading the compiled set fails when compiled.txt is missing, empty or unparseable, rather than yielding 0 files and a 0/0 pass (`go test ./internal/parse/ -run '^TestCompiledSetFailsWithoutFile$' -count=1 -v | grep -q '^--- PASS: TestCompiledSetFailsWithoutFile'`)
- [ ] The graded-set reader the difference log uses keeps exact, wider and no-answer rows and drops only excluded and error rows, checked on a fixture baseline holding one row of each status (`go test ./internal/parse/ -run '^TestGradedSetExcludesOnlyExcludedAndError$' -count=1 -v | grep -q '^--- PASS: TestGradedSetExcludesOnlyExcludedAndError'`)
- [ ] A compiled-set file absent from PERL5_CORPUS (wrong pin, partial checkout) counts as unclean and is named, never skipped out of the compiled set (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestCompiledCorpusMissingFileCountsAsUnclean$' -count=1 -v | grep -q '^--- PASS: TestCompiledCorpusMissingFileCountsAsUnclean'`)
- [ ] The 90% floor cannot hide a swap or a slide down to the 90% line: the test pins the unclean set, failing when any file outside it has an Unknown, and when a listed file becomes clean (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestParseWithoutErrorPinsUncleanSet$' -count=1 -v | grep -q '^--- PASS: TestParseWithoutErrorPinsUncleanSet'`)
- [ ] comp/final_line_num.t is refused on the line perl rejects; a refusal on any other line fails (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlRejectedFileRefusedOnPerlsLine$' -count=1 -v | grep -q '^--- PASS: TestPerlRejectedFileRefusedOnPerlsLine'`)
