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
updated: 2026-10-23T22:00:12.522950417Z
---

## Context

- paths: `internal/parse/canon.go`, `internal/parse/testdata/t1canon.ratchet`, `internal/parse/testdata/t1fidelity.ratchet`, `internal/parse/fidelity.go`, `internal/parse/fidelity_test.go`, `internal/parse/testdata/faithful_queue.txt`, `internal/parse/testdata/perlgittfaithful.ratchet`
- commands: `go test ./internal/parse/ -run '^TestFaithfulStatementBoundaries$' -count=1 -v`
- entrypoints: `internal/parse/fidelity.go` (`matchForgivingSemicolons`), `internal/parse/canon.go` (block and statement emission)

These are the `stmt-boundary` rows of faithful_queue.txt, where canon and source disagree on a `;`, a `}`, a block dereference's braces or where the token stream ends, rather than on a paren. 5f77's class table routes all three here by rule (J10, K10, L4, chain-review round 7): a source `{` after a `DerefSigil` or `FuncSigil`, an `EOF` on either side, and a `;` or `}` on either side with canon not `(`. At 2026-10-22, examples were `;`/`}` placement (op/closure.t, op/goto.t), a source `{` where canon writes the variable (block deref `%{$a}` vs `%$a`: op/kvaslice.t, op/sort.t, uni/fold.t), and a token-count mismatch (318 vs 369: lib/universal.t). Each is reduced to a minimal fixture and judged against perl. It is then fixed in the parser or canon; the class must reach zero. **perigrin's ruling (2026-10-23, chain-review round 7, superseding round 5): `Faithful` forgives exactly two differences (a block's last `;`, extra paren nesting), and the set is closed.** A block dereference's braces (`%{$a}` vs `%$a`, `@{$x}` vs `@$x`) are recorded in the tree, and canon writes them as the source did, the way `Paren` works. **They are recorded as a FLAG only (P24, chain-review round 6).** It is a boolean on the deref node, like `Paren`, with no change to its Kind, Text or children. sites.go keys its rv2hv decisions on node Text (`n.Text == "{"` at ~202, 258, 277, 285, 299 and 337; `text[1] != '{'` in isNamedHash and isNamedVariable at ~674-688; hidesStructure at ~521-536). A flag cannot change what those read, so this issue stays independent of the rv2hv issues and needs no edge to them. TestDerefBraceFlagLeavesSitesUnchanged holds that. `;`/`}` placement is fixed in the tree or canon too. A difference that seems to need a new forgiveness rule goes to perigrin with its perl measurement; this issue never adds one. Any block-form seed (`print {`, `sort {`, `grep {`, `bless {` against canon `(`) left after the `not`/list-op issue is NOT this class. It is a call with a block argument, so it belongs to `canon-adds-paren-term` (the term issue). Scope is the rows of these classes in faithful_queue.txt when this issue starts, not a fixed file list. The 2026-10-22 first-difference list is only an example, because the Paren change exposes later differences. A fix that changes a row's class re-classifies it, and if that class's owner has already closed, this issue fixes it; the zero-row criterion covers every earlier owner's class to hold that (J1, chain-review round 5). The triage issues run in sequence after the dropped-paren rows issue (3323), which owns the `dropped-paren` rows and itself follows the OPf_PARENS check (OPf_PARENS, then dropped-paren rows, then heredoc/sub-less, then statement-boundary, then operator, then term), because they all edit canon.go, fidelity.go, fidelity_test.go, faithful_queue.txt and perlgittfaithful.ratchet. **Why the sequence stays (A23, chain-review round 3):** the shared-generated-files rule lets parallel branches regenerate faithful_queue.txt and the ratchets. It does not cover canon.go, fidelity.go and fidelity_test.go, which are hand-written and edited by every one of these issues. Running them in sequence also gives 5f77's handover rule a defined order: a fix that moves a row into a class whose owner has closed is handled by the issue that moved it, which only works if owners close in a known order. The cost is a longer critical chain, accepted over hand-merging four sets of canon/fidelity changes.

## Steps

- [ ] FIRST, before any fix: read this class's `# causes <class> = N` line in faithful_queue.txt's header (5f77). If N is over 8, split this issue by cause using 5f77's sibling template before starting. This issue's body then keeps only its own cause group, and the cumulative class-zero criterion moves to the last sibling (J8, chain-review round 6).
- [ ] For each `stmt-boundary` row, reduce the first difference to a minimal snippet and measure it with the pinned perl 5.44.0, using `perl -MO=Deparse,-p` or `perl -MO=Concise`. Note whether the tree is wrong (a parser defect) or canon is wrong (the tree records the spelling but canon does not write it). Either way the fix is in the parser or canon; a "cosmetic" difference is not forgiven, it goes to perigrin as a proposed rule with its measurement.
- [ ] RED: add each reduced snippet to TestFaithfulStatementBoundaries, with the perl measurement in a comment. Run it and watch the cases fail.
- [ ] GREEN: fix one cause at a time in the parser or canon, running the test after each. `Faithful` is not changed.
- [ ] RED/GREEN TestDerefBraceFlagLeavesSitesUnchanged: for every snippet in internal/parse/testdata/m3probes/ and a fixture of `@{$x}`, `%{$a}`, `${$a->[0]}` and `@{$r->{k}}` shapes, `Sites` output is identical with and without the brace flag.
- [ ] RED/GREEN TestDerefBraceFlagClearedEqualsBracelessTree (P24, chain-review round 7): for each brace-form fixture, the tree with the flag cleared equals the tree of the braceless spelling (`%{$a}` against `%$a`, `@{$x}` against `@$x`), so the flag cannot hide a shape change.
- [ ] Remove the fixed rows from faithful_queue.txt, and re-baseline perlgittfaithful.ratchet, t1canon.ratchet and t1fidelity.ratchet in the same commit, naming each moved file (P25). T1 has `@{$x}` forms, and parse(canon(x)) must restore the flag for the canon fixpoint.
- [ ] Add every reduced case, with perl's facts, as probe file `internal/parse/testdata/m3probes/stmt-boundary.txt`.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus, and adds them to this issue's probe file under internal/parse/testdata/m3probes/ so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] If a row proves unfixable, stop and escalate it to perigrin with the perl measurement. Never leave it in the queue, and never close this issue with a row of its class left (perigrin's ruling, chain-review round 4).
- [ ] Run `make test`, then commit.

**Review note from 30a8 (2026-10-23).** TestDerefBraceFlagLeavesSitesUnchanged's fixtures have no subscripted brace deref. Add `${$h}{a}{b}` and `@{$h}{qw(a b)}`. Canon re-emits `${$h}{a}` as `$$h{a}` today, and 30a8's probe row for it carries no `faithful` fact until this issue lands.

## Acceptance criteria

### Positive Scenarios

- [ ] Each reduced statement-boundary case re-emits faithfully (`go test ./internal/parse/ -run '^TestFaithfulStatementBoundaries$' -count=1 -v | grep -q '^--- PASS: TestFaithfulStatementBoundaries'`)
- [ ] The work queue still matches the tree after the fixes (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueMatchesTree$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueMatchesTree'`)
- [ ] The reduced cases are committed as perl-judged probes, held by TestM3ProbesHeldToPerl (`sh -c 'test -s internal/parse/testdata/m3probes/stmt-boundary.txt && go test ./internal/parse/ -run "^TestM3ProbesHeldToPerl$" -count=1 -v | grep -q "^--- PASS: TestM3ProbesHeldToPerl"'`)
- [ ] No row of this issue's class, or of any earlier owner's class (dropped-paren, heredoc, subless-special-sub, stmt-boundary), is left in the work queue; rows a fix here exposed in a closed owner's class are fixed here (`sh -c 'f=internal/parse/testdata/faithful_queue.txt; test -f $f && ! grep -Eq "^[^[:space:]#]+[[:space:]]+(dropped-paren|heredoc|subless-special-sub|stmt-boundary)[[:space:]]" $f'`)
- [ ] Regression guard: the forgiven set is still exactly the two ruled differences, a block's last `;` and extra nesting (`go test ./internal/parse/ -run '^TestFaithfulForgivenessIsClosed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulForgivenessIsClosed'`)
- [ ] The deref-brace flag changes nothing Sites reads: Sites output on the probe corpus and the deref fixtures is unchanged (`go test ./internal/parse/ -run '^TestDerefBraceFlagLeavesSitesUnchanged$' -count=1 -v | grep -q '^--- PASS: TestDerefBraceFlagLeavesSitesUnchanged'`)
- [ ] The flag is only a flag: with it cleared, a brace-form tree equals the braceless spelling's tree (`go test ./internal/parse/ -run '^TestDerefBraceFlagClearedEqualsBracelessTree$' -count=1 -v | grep -q '^--- PASS: TestDerefBraceFlagClearedEqualsBracelessTree'`)
- [ ] Regression guard: the T1 canon fixpoint ratchet matches its re-baselined file (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)

### Negative Scenarios

- [ ] canon writes a block dereference's braces exactly where the source had them, and never forgives their absence. perl 5.42 Deparses `${$a->[0]}` as `${$a->[0];}` and `$$a->[0]` as `$$a->[0]`, which are different trees, so source `${$a->[0]}` against canon `$$a->[0]` stays unfaithful (`go test ./internal/parse/ -run '^TestFaithfulBlockDerefBracesKeptWhenGroupingChanges$' -count=1 -v | grep -q '^--- PASS: TestFaithfulBlockDerefBracesKeptWhenGroupingChanges'`)
- [ ] A missing `;` between two statements is a difference; the existing forgiveness covers only a block's last statement (`go test ./internal/parse/ -run '^TestFaithfulDoesNotForgiveMissingInnerSemicolon$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDoesNotForgiveMissingInnerSemicolon'`)
- [ ] A `}` placed so that a statement moves into or out of a block is a difference, even when the token multiset is unchanged (`go test ./internal/parse/ -run '^TestFaithfulDetectsStatementMovedAcrossBrace$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDetectsStatementMovedAcrossBrace'`)
- [ ] A token-count mismatch, as in lib/universal.t (318 vs 369), is never resolved by truncating or skipping the longer side's tail (`go test ./internal/parse/ -run '^TestFaithfulTokenCountMismatchIsADifference$' -count=1 -v | grep -q '^--- PASS: TestFaithfulTokenCountMismatchIsADifference'`)
