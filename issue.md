---
title: List every unfaithful M3-dir file as a located defect
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocks:
- 01a16602-5fd5-7cc4-9645-4539ca73e839
- 01a16602-60e9-7343-a630-7a93da5ee389
created: 2026-10-22T23:46:33.975890599Z
updated: 2026-10-23T01:08:50.322537832Z
---

## Context

- paths: `internal/parse/perlgitt_faithful_test.go`, `internal/parse/testdata/faithful_defects.txt` (new), `.github/workflows/ci.yml` (the skip-guard loop, ~207)
- commands: `PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulOnM3Dirs$' -count=1 -v`
- entrypoints: `internal/parse/perlgitt_faithful_test.go` (TestPerlGitTFaithfulRatchet)

M3 sets re-emission at 100% of what parses. Files that parse and re-emit differently are defects with locations, not a percentage. TestPerlGitTFaithfulOnM3Dirs asserts that every M3-dir file (`op mro uni lib test_pl`) that parses is faithful or is listed in `testdata/faithful_defects.txt`, with one `path:line reason` per row. It fails on an unlisted unfaithful file, on a listed file that has become faithful, and on a listed line that no longer fails. This issue seeds the list with every M3-dir failure that remains after canon writes `Paren` once and `Faithful` forgives extra nesting. Each row gets one cause class, and each class has one owning issue:

- `heredoc` and `subless-special-sub`: the heredoc/sub-less issue;
- `stmt-boundary`: the statement-boundary issue;
- `canon-adds-paren-op` (the source token at the first difference is an operator or named op): the operator issue;
- `canon-adds-paren-term` (a term, literal, bareword or `;`/`{`): the term issue;
- `dropped-paren` (a source paren canon still does not write): the OPf_PARENS issue;
- `other`: anything else. The term issue, last in the triage sequence, owns it.

The class is decided from the tree as it stands after the Paren change, not from the 2026-10-22 scratch list. A file's first difference is all a row shows, so fixing it can expose a later difference of another class. The row then changes class rather than disappearing. The issue that makes that change re-classifies the row, and if the new class's owner has already closed, it handles the row itself.

## Steps

- [ ] RED: write TestFaithfulDefectsRowsAreWellFormed, a unit test that every row of faithful_defects.txt parses as `path:line class reason` with one of the seven known classes. Run it against a malformed fixture and watch it fail; then implement the row parser and watch it pass.
- [ ] RED: write TestPerlGitTFaithfulOnM3Dirs. Run it with an empty defects file and watch it fail, listing every unfaithful M3-dir file with its first differing line.
- [ ] GREEN: seed faithful_defects.txt from that output, classifying each row by reading the failing line. Run the test and watch it pass.
- [ ] Add TestPerlGitTFaithfulOnM3Dirs to ci.yml's "Fail if a corpus-backed ratchet skipped" loop, and confirm that the job sets the corpus env it needs.
- [ ] Run `make test`, then commit, with the per-class counts in the message.

## Acceptance criteria

### Positive Scenarios

- [ ] Every defects row names a path, line, known cause class and reason (`go test ./internal/parse/ -run '^TestFaithfulDefectsRowsAreWellFormed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDefectsRowsAreWellFormed'`)
- [ ] Every M3-dir file that parses is faithful or listed with its location (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulOnM3Dirs$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulOnM3Dirs'`)
- [ ] CI fails if the M3-dir check skips (`grep -q TestPerlGitTFaithfulOnM3Dirs .github/workflows/ci.yml`)

### Negative Scenarios

- [ ] The M3-dir walk fails unless it visits every M3-dir file outside `perlRejects` (347 files in the M3 dirs); a missing dir or unset corpus cannot pass against an empty defects list (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulOnM3DirsVisitsEveryFile$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulOnM3DirsVisitsEveryFile'`)
- [ ] An M3-dir file that stops parsing (gains an Unknown) fails the check, rather than leaving the "files that parse" set unnoticed (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulOnM3DirsFailsOnUnparsedFile$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulOnM3DirsFailsOnUnparsedFile'`)
- [ ] A defects row naming a file outside the M3 dirs, or a file that does not exist in the corpus, fails (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulDefectsRowsNameM3Files$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDefectsRowsNameM3Files'`)
- [ ] Two rows for the same path fail, so a file cannot be parked under two classes (`go test ./internal/parse/ -run '^TestFaithfulDefectsRowsAreUnique$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDefectsRowsAreUnique'`)
- [ ] A row whose line is not the file's first differing line fails, so a fix that moves the first difference later shows up as a changed row (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulDefectsLineIsFirstDifference$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDefectsLineIsFirstDifference'`)
- [ ] A row with an empty reason, a zero line, or a class outside the seven known classes is rejected (`go test ./internal/parse/ -run '^TestFaithfulDefectsRowsRejectMalformed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDefectsRowsRejectMalformed'`)
