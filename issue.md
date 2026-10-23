---
title: Triage statement-boundary re-emission differences in the M3 dirs
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5f77-78ef-a0c7-538e0a241cdc
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
created: 2026-10-22T23:46:34.159970536Z
updated: 2026-10-23T01:08:13.698716839Z
---

## Context

- paths: `internal/parse/canon.go`, `internal/parse/fidelity.go`, `internal/parse/fidelity_test.go`, `internal/parse/testdata/faithful_defects.txt`, `internal/parse/testdata/perlgittfaithful.ratchet`
- commands: `go test ./internal/parse/ -run '^TestFaithfulStatementBoundaries$' -count=1 -v`
- entrypoints: `internal/parse/fidelity.go` (`matchForgivingSemicolons`), `internal/parse/canon.go` (block and statement emission)

These are the `stmt-boundary` rows of faithful_defects.txt, where canon and source disagree on a `;`, a `}` or a block dereference's braces rather than on a paren. At 2026-10-22, examples were `;`/`}` placement (op/closure.t, op/goto.t), a source `{` where canon writes the variable (block deref `%{$a}` vs `%$a`: op/kvaslice.t, op/sort.t, uni/fold.t), and a token-count mismatch (318 vs 369: lib/universal.t). Each is reduced to a minimal fixture and judged against perl. It is then fixed, or kept in faithful_defects.txt with its line and the perl measurement that explains it. Scope is the rows of these classes in faithful_defects.txt when this issue starts, not a fixed file list. The 2026-10-22 first-difference list is only an example, because the Paren change exposes later differences. A fix that changes a row's class re-classifies it, and if that class's owner has already closed, this issue handles it. The triage issues run in sequence (heredoc/sub-less, then statement-boundary, then operator, then term) because they all edit canon.go, fidelity.go, fidelity_test.go, faithful_defects.txt and perlgittfaithful.ratchet.

## Steps

- [ ] For each `stmt-boundary` row, reduce the first difference to a minimal snippet and measure it with the pinned perl 5.44.0, using `perl -MO=Deparse,-p` or `perl -MO=Concise`. Note whether the tree is wrong (a parser defect), canon is wrong, or the difference is cosmetic.
- [ ] RED: add each reduced snippet to TestFaithfulStatementBoundaries, with the perl measurement in a comment. Run it and watch the cases fail.
- [ ] GREEN: fix one cause at a time, running the test after each. Do not widen the forgiven set beyond what the measurement shows is cosmetic.
- [ ] Remove the fixed rows from faithful_defects.txt. Give every kept row its line and reason, and re-baseline perlgittfaithful.ratchet in the same commit, naming each moved file.
- [ ] Add every reduced case, with perl's facts, as probe file `internal/parse/testdata/m3probes/stmt-boundary.txt`.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus, and adds them to this issue's probe file under internal/parse/testdata/m3probes/ so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PATH=$HOME/.local/perl-5.44.0/bin:$PATH PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] Each reduced statement-boundary case re-emits faithfully or is pinned with its perl measurement (`go test ./internal/parse/ -run '^TestFaithfulStatementBoundaries$' -count=1 -v | grep -q '^--- PASS: TestFaithfulStatementBoundaries'`)
- [ ] The defects list and the M3-dir check agree after the fixes (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulOnM3Dirs$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulOnM3Dirs'`)
- [ ] The reduced cases are committed as perl-judged probes (`test -s internal/parse/testdata/m3probes/stmt-boundary.txt`)

### Negative Scenarios

- [ ] A block dereference's braces are forgiven only where they change no grouping. perl 5.42 Deparses `${$a->[0]}` as `${$a->[0];}` and `$$a->[0]` as `$$a->[0]`, which are different trees, so source `${$a->[0]}` against canon `$$a->[0]` stays unfaithful (`go test ./internal/parse/ -run '^TestFaithfulBlockDerefBracesKeptWhenGroupingChanges$' -count=1 -v | grep -q '^--- PASS: TestFaithfulBlockDerefBracesKeptWhenGroupingChanges'`)
- [ ] A missing `;` between two statements is a difference; the existing forgiveness covers only a block's last statement (`go test ./internal/parse/ -run '^TestFaithfulDoesNotForgiveMissingInnerSemicolon$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDoesNotForgiveMissingInnerSemicolon'`)
- [ ] A `}` placed so that a statement moves into or out of a block is a difference, even when the token multiset is unchanged (`go test ./internal/parse/ -run '^TestFaithfulDetectsStatementMovedAcrossBrace$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDetectsStatementMovedAcrossBrace'`)
- [ ] A token-count mismatch, as in lib/universal.t (318 vs 369), is never resolved by truncating or skipping the longer side's tail (`go test ./internal/parse/ -run '^TestFaithfulTokenCountMismatchIsADifference$' -count=1 -v | grep -q '^--- PASS: TestFaithfulTokenCountMismatchIsADifference'`)
