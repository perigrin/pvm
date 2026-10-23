---
title: Triage statement-boundary re-emission differences in the M3 dirs
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5fd5-7cc4-9645-4539ca73e839
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
- 01a16602-608d-71ed-8ab2-66a9951d9957
created: 2026-10-22T23:46:34.159970536Z
updated: 2026-10-23T07:06:31.096172789Z
---

## Context

- paths: `internal/parse/canon.go`, `internal/parse/fidelity.go`, `internal/parse/fidelity_test.go`, `internal/parse/testdata/faithful_queue.txt`, `internal/parse/testdata/perlgittfaithful.ratchet`
- commands: `go test ./internal/parse/ -run '^TestFaithfulStatementBoundaries$' -count=1 -v`
- entrypoints: `internal/parse/fidelity.go` (`matchForgivingSemicolons`), `internal/parse/canon.go` (block and statement emission)

These are the `stmt-boundary` rows of faithful_queue.txt, where canon and source disagree on a `;`, a `}` or a block dereference's braces rather than on a paren. At 2026-10-22, examples were `;`/`}` placement (op/closure.t, op/goto.t), a source `{` where canon writes the variable (block deref `%{$a}` vs `%$a`: op/kvaslice.t, op/sort.t, uni/fold.t), and a token-count mismatch (318 vs 369: lib/universal.t). Each is reduced to a minimal fixture and judged against perl. It is then fixed; the class must reach zero. Scope is the rows of these classes in faithful_queue.txt when this issue starts, not a fixed file list. The 2026-10-22 first-difference list is only an example, because the Paren change exposes later differences. A fix that changes a row's class re-classifies it, and if that class's owner has already closed, this issue handles it. The triage issues run in sequence after the OPf_PARENS issue, which owns the `dropped-paren` rows (OPf_PARENS, then heredoc/sub-less, then statement-boundary, then operator, then term), because they all edit canon.go, fidelity.go, fidelity_test.go, faithful_queue.txt and perlgittfaithful.ratchet. **Why the sequence stays (A23, chain-review round 3):** the shared-generated-files rule lets parallel branches regenerate faithful_queue.txt and the ratchets. It does not cover canon.go, fidelity.go and fidelity_test.go, which are hand-written and edited by every one of these issues. Running them in sequence also gives 5f77's handover rule a defined order: a fix that moves a row into a class whose owner has closed is handled by the issue that moved it, which only works if owners close in a known order. The cost is a longer critical chain, accepted over hand-merging four sets of canon/fidelity changes.

## Steps

- [ ] For each `stmt-boundary` row, reduce the first difference to a minimal snippet and measure it with the pinned perl 5.44.0, using `perl -MO=Deparse,-p` or `perl -MO=Concise`. Note whether the tree is wrong (a parser defect), canon is wrong, or the difference is cosmetic.
- [ ] RED: add each reduced snippet to TestFaithfulStatementBoundaries, with the perl measurement in a comment. Run it and watch the cases fail.
- [ ] GREEN: fix one cause at a time, running the test after each. Do not widen the forgiven set beyond what the measurement shows is cosmetic.
- [ ] Remove the fixed rows from faithful_queue.txt, and re-baseline perlgittfaithful.ratchet in the same commit, naming each moved file.
- [ ] Add every reduced case, with perl's facts, as probe file `internal/parse/testdata/m3probes/stmt-boundary.txt`.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus, and adds them to this issue's probe file under internal/parse/testdata/m3probes/ so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] If a row proves unfixable, stop and escalate it to perigrin with the perl measurement. Never leave it in the queue, and never close this issue with a row of its class left (perigrin's ruling, chain-review round 4).
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] Each reduced statement-boundary case re-emits faithfully (`go test ./internal/parse/ -run '^TestFaithfulStatementBoundaries$' -count=1 -v | grep -q '^--- PASS: TestFaithfulStatementBoundaries'`)
- [ ] The work queue still matches the tree after the fixes (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueMatchesTree$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueMatchesTree'`)
- [ ] The reduced cases are committed as perl-judged probes (`test -s internal/parse/testdata/m3probes/stmt-boundary.txt`)
- [ ] No statement-boundary row is left in the work queue (`sh -c 'f=internal/parse/testdata/faithful_queue.txt; test -f $f && ! grep -Eq "^[^ #]+ (stmt-boundary) " $f'`)

### Negative Scenarios

- [ ] A block dereference's braces are forgiven only where they change no grouping. perl 5.42 Deparses `${$a->[0]}` as `${$a->[0];}` and `$$a->[0]` as `$$a->[0]`, which are different trees, so source `${$a->[0]}` against canon `$$a->[0]` stays unfaithful (`go test ./internal/parse/ -run '^TestFaithfulBlockDerefBracesKeptWhenGroupingChanges$' -count=1 -v | grep -q '^--- PASS: TestFaithfulBlockDerefBracesKeptWhenGroupingChanges'`)
- [ ] A missing `;` between two statements is a difference; the existing forgiveness covers only a block's last statement (`go test ./internal/parse/ -run '^TestFaithfulDoesNotForgiveMissingInnerSemicolon$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDoesNotForgiveMissingInnerSemicolon'`)
- [ ] A `}` placed so that a statement moves into or out of a block is a difference, even when the token multiset is unchanged (`go test ./internal/parse/ -run '^TestFaithfulDetectsStatementMovedAcrossBrace$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDetectsStatementMovedAcrossBrace'`)
- [ ] A token-count mismatch, as in lib/universal.t (318 vs 369), is never resolved by truncating or skipping the longer side's tail (`go test ./internal/parse/ -run '^TestFaithfulTokenCountMismatchIsADifference$' -count=1 -v | grep -q '^--- PASS: TestFaithfulTokenCountMismatchIsADifference'`)
