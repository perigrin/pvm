---
title: List every unfaithful M3-dir file as a located defect
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a1664d-75d4-76eb-aad2-10b322f366a9
blocks:
- 01a16602-5fd5-7cc4-9645-4539ca73e839
- 01a16602-5f1a-7d56-a91d-5fa2c478d608
created: 2026-10-22T23:46:33.975890599Z
updated: 2026-10-23T07:06:25.171126028Z
---

## Context

- paths: `internal/parse/perlgitt_faithful_test.go`, `internal/parse/testdata/faithful_queue.txt` (new), `.github/workflows/ci.yml` (the skip-guard loop, ~207), `internal/parse/pin_test.go` (TestCIRunsTheRatchets)
- commands: `PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueMatchesTree$' -count=1 -v`
- entrypoints: `internal/parse/perlgitt_faithful_test.go` (TestPerlGitTFaithfulRatchet)

**perigrin's ruling (2026-10-23, chain-review round 4): re-emission is 100% faithful on the M3 dirs, with no defects file.** Every M3-dir file that parses must re-emit faithfully when the milestone closes. `testdata/faithful_queue.txt` is a WORK QUEUE, not an accepted-defects register. It holds one `path:line class reason` row per M3-dir file that is unfaithful today. Each triage issue empties its own class, and the queue must be empty when the milestone closes. A row that proves unfixable is escalated to perigrin; it is never parked.

TestFaithfulQueueMatchesTree asserts the queue is exactly the current state. It fails on an unfaithful M3-dir file (`op mro uni lib test_pl`) that has no row, on a row whose file has become faithful, and on a row whose line is not the file's first difference. TestPerlGitTFaithfulOnM3Dirs is the 100% assertion with no exception list; it belongs to the last re-emission issue and to the milestone's resolution, not to this one.

This issue seeds the queue with every M3-dir failure that remains after canon writes `Paren` once and `Faithful` forgives extra nesting. Each row gets one cause class, and each class has one owning issue that FIXES its rows and brings the class to zero:

- `dropped-paren` (a source paren canon still does not write): the OPf_PARENS issue;
- `heredoc` and `subless-special-sub`: the heredoc/sub-less issue;
- `stmt-boundary`: the statement-boundary issue;
- `canon-adds-paren-op` (the source token at the first difference is an operator or named op): the operator issue;
- `canon-adds-paren-term` (a term, literal, bareword or `;`/`{`): the term issue;
- `other`: anything else, owned by the final re-emission issue, which fixes those rows and asserts 100%.

Sizing, from the 2026-10-22 first-difference list measured at 054e75e6 (scratchpad faith.log) plus the 13 files pushback round 1 found behind dropped parens: about 14 heredoc, 6 sub-less, 15 statement-boundary, 11 operator, 19 term. The true counts come from this issue's seeding.

The class is decided from the tree as it stands after the Paren change, not from the 2026-10-22 scratch list. A file's first difference is all a row shows, so fixing it can expose a later difference of another class. The row then changes class rather than disappearing. The issue that makes that change re-classifies the row, and if the new class's owner has already closed, it handles the row itself.

## Steps

- [ ] RED: write TestFaithfulQueueRowsAreWellFormed, a unit test that every row of faithful_queue.txt parses as `path:line class reason` with one of the seven known classes. A file holding only `#` comment lines is a valid, empty queue. Run it against a malformed fixture and watch it fail; then implement the row parser and watch it pass.
- [ ] RED: write TestFaithfulQueueMatchesTree. Run it with an empty queue and watch it fail, listing every unfaithful M3-dir file with its first differing line.
- [ ] GREEN: seed faithful_queue.txt from that output, classifying each row by reading the failing line. Run the test and watch it pass.
- [ ] Add TestFaithfulQueueMatchesTree to ci.yml's "Fail if a corpus-backed ratchet skipped" loop and to TestCIRunsTheRatchets' list (internal/parse/pin_test.go:71), and confirm that the job sets the corpus env it needs.
- [ ] Run `make test`, then commit, with the per-class counts in the message.

## Acceptance criteria

### Positive Scenarios

- [ ] Every queue row names a path, line, known cause class and reason (`go test ./internal/parse/ -run '^TestFaithfulQueueRowsAreWellFormed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueRowsAreWellFormed'`)
- [ ] The work queue is exactly the M3-dir files that parse and re-emit unfaithfully, each at its first differing line (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueMatchesTree$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueMatchesTree'`)
- [ ] CI fails if the queue check skips: the test is in TestCIRunsTheRatchets' list, and that test passes (`sh -c 'grep -q "\"TestFaithfulQueueMatchesTree\"," internal/parse/pin_test.go && go test ./internal/parse/ -run "^TestCIRunsTheRatchets$" -count=1 -v | grep -q "^--- PASS: TestCIRunsTheRatchets"'`)

### Negative Scenarios

- [ ] The M3-dir walk fails unless it visits every M3-dir file outside `perlRejects` (347 files in the M3 dirs); a missing dir or unset corpus cannot pass against an empty queue (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulOnM3DirsVisitsEveryFile$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulOnM3DirsVisitsEveryFile'`)
- [ ] An M3-dir file that stops parsing (gains an Unknown) fails the check, rather than leaving the "files that parse" set unnoticed (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulOnM3DirsFailsOnUnparsedFile$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulOnM3DirsFailsOnUnparsedFile'`)
- [ ] A queue row naming a file outside the M3 dirs, or a file that does not exist in the corpus, fails (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueRowsNameM3Files$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueRowsNameM3Files'`)
- [ ] Two rows for the same path fail, so a file cannot sit in two classes (`go test ./internal/parse/ -run '^TestFaithfulQueueRowsAreUnique$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueRowsAreUnique'`)
- [ ] A row whose line is not the file's first differing line fails, so a fix that moves the first difference later shows up as a changed row (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueLineIsFirstDifference$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueLineIsFirstDifference'`)
- [ ] A row with an empty reason, a zero line, or a class outside the seven known classes is rejected (`go test ./internal/parse/ -run '^TestFaithfulQueueRowsRejectMalformed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueRowsRejectMalformed'`)
