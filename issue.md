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
updated: 2026-10-23T12:56:22.540126431Z
---

**The title is stale and git-zhi cannot retitle.** Read it as "Seed the re-emission work queue with every unfaithful M3-dir file". Under perigrin's round-4 ruling the queue is a work list that must be empty at close, not a register of located defects (J6/A27, chain-review round 5).

## Context

- paths: `internal/parse/perlgitt_faithful_test.go`, `internal/parse/testdata/faithful_queue.txt` (new), `.github/workflows/ci.yml` (the skip-guard loop, ~207), `internal/parse/pin_test.go` (TestCIRunsTheRatchets)
- commands: `PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueMatchesTree$' -count=1 -v`
- entrypoints: `internal/parse/perlgitt_faithful_test.go` (TestPerlGitTFaithfulRatchet)

**perigrin's ruling (2026-10-23, chain-review round 4): re-emission is 100% faithful on the M3 dirs, with no defects file.** Every M3-dir file that parses must re-emit faithfully when the milestone closes. `testdata/faithful_queue.txt` is a WORK QUEUE, not an accepted-defects register. It holds one `path:line class reason` row per M3-dir file that is unfaithful today. Each triage issue empties its own class, and the queue must be empty when the milestone closes. A row that proves unfixable is escalated to perigrin; it is never parked.

TestFaithfulQueueMatchesTree asserts the queue is exactly the current state. It fails on an unfaithful M3-dir file (`op mro uni lib test_pl`) that has no row, on a row whose file has become faithful, and on a row whose line is not the file's first difference. TestPerlGitTFaithfulOnM3Dirs is the 100% assertion with no exception list; it belongs to the last re-emission issue and to the milestone's resolution, not to this one.

This issue seeds the queue with every M3-dir failure that remains after canon writes `Paren` once and `Faithful` forgives extra nesting. Each row gets one cause class, and each class has one owning issue that FIXES its rows and brings the class to zero:

- `dropped-paren` (a source paren canon still does not write): the dropped-paren rows issue, which follows the OPf_PARENS check;
- `heredoc` and `subless-special-sub`: the heredoc/sub-less issue;
- `stmt-boundary`: the statement-boundary issue;
- `canon-adds-paren-op` (the source token at the first difference is an operator or named op): the operator issue;
- `canon-adds-paren-term` (a term, literal, bareword or `;`/`{`): the term issue. This includes the block-form calls (`print {`, `sort {`, `grep {`, `bless {` against canon `(`), which are a call with a block argument, not a statement boundary;
- `other`: anything else, owned by the final re-emission issue, which fixes those rows and asserts 100%.

**Sizing is by causes, not rows, and the cause is computed (P22, J1, J7, J8, P23, chain-review rounds 5 and 6).** A row shows only a file's FIRST difference, and reaching zero means fixing every difference in the file.
- Measured at 054e75e6:
  - fresh4 (round 5): 1174 differing places in 168 files, about 17 files exposing a later row into a class whose owner has closed, and 26 distinct leading signatures (69 in all).
  - pushback (round 5): 169 failing files, 39 still differing with every paren dropped, and a simulated post-Paren-once seed of 87 rows (about 43 canon-adds-paren, 21 block-form, 8 heredoc, 6 subless, 8 other or dropped-paren).
- **Row format:** `path:line class cause reason`, with single spaces.
- **Cause:** a computed signature, `<source token kind>/<canon token kind>` at the file's first difference (lexer token kinds, e.g. `Word/LParen`, `HeredocBody/LBrace`). TestFaithfulQueueMatchesTree computes it and logs it for every unfaithful file.
- **Class:** also computed, from the cause by a fixed table in the test:
  - source `(` that canon does not write → `dropped-paren`;
  - source heredoc body or terminator → `heredoc`;
  - canon `sub` where the source has a special-sub name → `subless-special-sub`;
  - source and canon both in `;` `{` `}`, with canon not `(` → `stmt-boundary`;
  - canon `(` after an operator or named op → `canon-adds-paren-op`;
  - canon `(` after a term, literal, bareword, `;` or `{` → `canon-adds-paren-term`;
  - anything else → `other`.
  
  The test fails on any row whose class or cause disagrees with the computed one, so relabelling a hard row as `other` cannot dodge a class (P23).
- **Header:** the queue file carries one `# causes <class> = N` comment line per class, all seven, with N the number of distinct causes in that class's current rows. TestFaithfulQueueRowsAreWellFormed parses them. TestFaithfulQueueMatchesTree requires them to equal the computed counts, so they are always current (J9, H9).
- **Split rule, mechanical, applied at every owner's start (the dropped-paren rows issue included):** if the owner's class has more than 8 in its `# causes` line when the owner starts, the owner splits before doing any fix. The template below is used. The class-zero line moves to the last sibling, and the owner's body keeps only its own cause group's criteria.

The class is decided from the tree as it stands after the Paren change, not from the 2026-10-22 scratch list. A file's first difference is all a row shows, so fixing it can expose a later difference of another class. The row then changes class rather than disappearing. The issue that makes that change re-classifies the row, and if the new class's owner has already closed, it handles the row itself.

**Split-sibling template (J7, P22).** A sibling issue is created with `git zhi issue add --body-file`, then read back. It is blocked by the previous sibling, or by the owner for the first one, and it blocks the next owner in the chain. Its body follows the chain's rules: Context, Steps, then a heading containing "criteria", one paren-wrapped command per line, positive and negative scenarios, and nothing after the block. It carries:
- **Context:** the class, the cause signatures it owns, and their row counts.
- **Steps:** the parent's steps, scoped to those causes, including the perl measurement, RED/GREEN, the per-fix review, the sweep when the parser moves, and the escalation rule.
- **Positive:**
  - a cause-zero line (`test -f` the queue, then `! grep -Eq` for rows with those causes);
  - TestFaithfulQueueMatchesTree;
  - a per-cause fixture test, `TestFaithful<Cause>MatchesPerl`, whose name is unique;
  - its probe file, backed by TestM3ProbesHeldToPerl;
  - TestFaithfulForgivenessIsClosed as a guard.
- **Negative:** at least one misgrouping or relabelling case for its causes. The parent's cumulative class-zero line moves to the LAST sibling only.

## Steps

- [ ] RED: write TestFaithfulQueueRowsAreWellFormed: every row parses as `path:line class cause reason`, with single spaces, one of the seven classes and a cause of the form `<kind>/<kind>`. The header holds one `# causes <class> = N` line for each of the seven classes. Run it against a malformed fixture and watch it fail; then implement the parser and watch it pass.
- [ ] RED/GREEN TestFaithfulQueueRowsRejectMalformed: rejects an empty reason, a zero line, an unknown class, a missing cause, a missing or duplicate `# causes` line, and any spacing other than one space.
- [ ] RED/GREEN TestPerlGitTFaithfulOnM3DirsVisitsEveryFile: the M3-dir walk visits every M3-dir file outside `perlRejects`, and fails on a missing dir or an unset corpus.
- [ ] RED/GREEN TestPerlGitTFaithfulOnM3DirsFailsOnUnparsedFile: an M3-dir file that gains an Unknown fails the walk.
- [ ] RED/GREEN TestFaithfulQueueRowsNameM3Files: a row naming a file outside the M3 dirs, or one missing from the corpus, fails.
- [ ] RED/GREEN TestFaithfulQueueRowsAreUnique: two rows for one path fail.
- [ ] RED/GREEN TestFaithfulQueueLineIsFirstDifference: a row whose line is not the file's first difference fails.
- [ ] RED: write TestFaithfulQueueMatchesTree. It computes each unfaithful M3-dir file's first difference, cause and class, logs them, compares rows and `# causes` counts with the queue, and fails both ways. Run it with an empty queue and watch it fail.
- [ ] RED/GREEN TestFaithfulQueueClassIsComputed: a fixture queue relabelling a `canon-adds-paren-op` row as `other`, or giving it a wrong cause, fails.
- [ ] GREEN: seed faithful_queue.txt from the computed output, with the header counts. Run the test and watch it pass.
- [ ] Add TestFaithfulQueueMatchesTree to ci.yml's "Fail if a corpus-backed ratchet skipped" loop and to TestCIRunsTheRatchets' list (internal/parse/pin_test.go:71), and confirm that the job sets the corpus env it needs.
- [ ] Apply the split rule to the first owner's class (dropped-paren) and to any class already over 8 causes: create the siblings from the template, and wire their edges. Record each class's row and cause counts, and each file's full difference depth, in the commit message too.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] Every queue row names a path, line, known class, computed cause and reason, and the header has all seven cause counts (`go test ./internal/parse/ -run '^TestFaithfulQueueRowsAreWellFormed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueRowsAreWellFormed'`)
- [ ] The work queue is exactly the M3-dir files that parse and re-emit unfaithfully, each at its first differing line (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueMatchesTree$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueMatchesTree'`)
- [ ] CI fails if the queue check skips: the test is in TestCIRunsTheRatchets' list, and that test passes (`sh -c 'grep -q "\"TestFaithfulQueueMatchesTree\"," internal/parse/pin_test.go && go test ./internal/parse/ -run "^TestCIRunsTheRatchets$" -count=1 -v | grep -q "^--- PASS: TestCIRunsTheRatchets"'`)
- [ ] The queue header carries a `# causes <class> = N` line for each of the seven classes, and they equal the counts the test computes (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueMatchesTree$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueMatchesTree'`)

### Negative Scenarios

- [ ] The M3-dir walk fails unless it visits every M3-dir file outside `perlRejects` (347 files in the M3 dirs); a missing dir or unset corpus cannot pass against an empty queue (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulOnM3DirsVisitsEveryFile$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulOnM3DirsVisitsEveryFile'`)
- [ ] An M3-dir file that stops parsing (gains an Unknown) fails the check, rather than leaving the "files that parse" set unnoticed (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulOnM3DirsFailsOnUnparsedFile$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulOnM3DirsFailsOnUnparsedFile'`)
- [ ] A queue row naming a file outside the M3 dirs, or a file that does not exist in the corpus, fails (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueRowsNameM3Files$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueRowsNameM3Files'`)
- [ ] Two rows for the same path fail, so a file cannot sit in two classes (`go test ./internal/parse/ -run '^TestFaithfulQueueRowsAreUnique$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueRowsAreUnique'`)
- [ ] A row whose line is not the file's first differing line fails, so a fix that moves the first difference later shows up as a changed row (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueLineIsFirstDifference$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueLineIsFirstDifference'`)
- [ ] A row with an empty reason, a zero line, a class outside the seven known classes, a missing cause, a missing `# causes` line, or fields separated by anything but one space is rejected (`go test ./internal/parse/ -run '^TestFaithfulQueueRowsRejectMalformed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueRowsRejectMalformed'`)

- [ ] A row whose class or cause disagrees with the one computed from its first difference fails, so relabelling a row as `other` cannot dodge its class (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueClassIsComputed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueClassIsComputed'`)
