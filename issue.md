---
title: Judge canon-added parens on terms and argument lists against perl
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-608d-71ed-8ab2-66a9951d9957
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
- 01a16795-75e8-7787-81ad-79e130324139
created: 2026-10-22T23:46:34.345543908Z
updated: 2026-10-23T13:55:22.542293255Z
---

## Context

- paths: `internal/parse/canon.go`, `internal/parse/fidelity.go`, the parser files the fix touches, `internal/parse/fidelity_test.go`, `internal/parse/testdata/faithful_queue.txt`, `internal/parse/testdata/perlgittfaithful.ratchet`
- commands: `go test ./internal/parse/ -run '^TestCanonAddedParensOnTermsMatchPerl$' -count=1 -v`
- entrypoints: `internal/parse/canon.go` (binding-power parens, the follow power, the list-op branch of the `Call` case), `internal/parse/call.go` (`ArgParen`)

These are the `canon-adds-paren-term` rows of faithful_queue.txt, where canon writes a `(` the source lacks after a call name, a list op, a bareword, a term, a literal, `;` or `{` (5f77's rule 8), including any block-form call (`print {`, `sort {`, `grep {`, `bless {` against canon `(`: about 21 at the pushback round-5 simulation, most of which the `not`/list-op issue should have made faithful) still left. At 2026-10-22, examples were op/cmpchain.t, op/avhv.t and op/tr.t. Under the round-7 ruling canon writes a call's parens only where the source had them (`ArgParen`) or where binding power requires them, plus a `Paren` node's pair. So an added paren here is one of two things: a Call whose `ArgParen` the parser failed to record, or whose list extent canon's binding power gets wrong (a canon defect), or a tree that groups the arguments differently from perl (a parser misgrouping). There is no forgiveness to apply. That makes these the most important rows in the list. No case may be called a misgrouping, or forgiven, before it is judged against perl with `perl -MO=Deparse,-p` or an op tree. Scope is the rows of these classes in faithful_queue.txt when this issue starts, not a fixed file list. The 2026-10-22 first-difference list is only an example, because the Paren change exposes later differences. A fix that changes a row's class re-classifies it, and if that class's owner has already closed, this issue fixes it; the zero-row criterion covers every earlier owner's class to hold that (J1, chain-review round 5). The triage issues run in sequence after the dropped-paren rows issue (3323), which owns the `dropped-paren` rows and itself follows the OPf_PARENS check (OPf_PARENS, then dropped-paren rows, then heredoc/sub-less, then statement-boundary, then operator, then term), because they all edit canon.go, fidelity.go, fidelity_test.go, faithful_queue.txt and perlgittfaithful.ratchet. **Why the sequence stays (A23, chain-review round 3):** the shared-generated-files rule lets parallel branches regenerate faithful_queue.txt and the ratchets. It does not cover canon.go, fidelity.go and fidelity_test.go, which are hand-written and edited by every one of these issues. Running them in sequence also gives 5f77's handover rule a defined order: a fix that moves a row into a class whose owner has closed is handled by the issue that moved it, which only works if owners close in a known order. The cost is a longer critical chain, accepted over hand-merging four sets of canon/fidelity changes.

## Steps

- [ ] FIRST, before any fix: read this class's `# causes <class> = N` line in faithful_queue.txt's header (5f77). If N is over 8, split this issue by cause using 5f77's sibling template before starting. This issue's body then keeps only its own cause group, and the cumulative class-zero criterion moves to the last sibling (J8, chain-review round 6).
- [ ] For each `canon-adds-paren-term` row, reduce the first difference to a minimal snippet and run the pinned perl 5.44.0's `perl -MO=Deparse,-p` on it. Record whether perl's grouping matches the tree's (then canon or the `ArgParen` recording is wrong) or differs (a parser misgrouping). Neither is fixed by a forgiveness rule.
- [ ] RED: add each reduced snippet to TestCanonAddedParensOnTermsMatchPerl, asserting perl's grouping and quoting the Deparse output in a comment. Run it and watch the cases fail.
- [ ] GREEN: fix one cause at a time. A misgrouping is fixed in the parser. A canon defect is fixed in canon or in the parser's `ArgParen` recording. `Faithful` is not changed (perigrin's round-7 ruling: the forgiven set is closed at a block's last `;` and extra paren nesting). A difference that seems to need a new rule goes to perigrin with its perl measurement. Each row is fixed one way or the other; the class must reach zero. Run the test after each fix.
- [ ] Run TestCanonCatchesMisgrouping and TestCanonQuotientIsNarrow after every fix, and add each misgrouping found here as a named case to TestCanonCatchesMisgrouping.
- [ ] Remove the fixed rows from faithful_queue.txt, and re-baseline perlgittfaithful.ratchet (and the T1 ratchets if they move) in the same commit, naming each moved file.
- [ ] Add every reduced case, with perl's facts, as probe file `internal/parse/testdata/m3probes/canon-paren-term.txt`.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus, and adds them to this issue's probe file under internal/parse/testdata/m3probes/ so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] If a row proves unfixable, stop and escalate it to perigrin with the perl measurement. Never leave it in the queue, and never close this issue with a row of its class left (perigrin's ruling, chain-review round 4).
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] Each reduced case groups as perl's Deparse,-p does (`go test ./internal/parse/ -run '^TestCanonAddedParensOnTermsMatchPerl$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParensOnTermsMatchPerl'`)
- [ ] Regression guard: canon still catches misgrouping (`go test ./internal/parse/ -run '^TestCanonCatchesMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonCatchesMisgrouping'`)
- [ ] The work queue still matches the tree after the fixes (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueMatchesTree$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueMatchesTree'`)
- [ ] The perl.git t/ faithful ratchet matches the tree after this issue's re-baseline (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchet'`)
- [ ] The reduced cases are committed as perl-judged probes, held by TestM3ProbesHeldToPerl (`sh -c 'test -s internal/parse/testdata/m3probes/canon-paren-term.txt && go test ./internal/parse/ -run "^TestM3ProbesHeldToPerl$" -count=1 -v | grep -q "^--- PASS: TestM3ProbesHeldToPerl"'`)
- [ ] No row of this issue's class, or of any earlier owner's class (dropped-paren, heredoc, subless-special-sub, stmt-boundary, canon-adds-paren-op, canon-adds-paren-term), is left in the work queue; rows a fix here exposed in a closed owner's class are fixed here (`sh -c 'f=internal/parse/testdata/faithful_queue.txt; test -f $f && ! grep -Eq "^[^[:space:]#]+[[:space:]]+(dropped-paren|heredoc|subless-special-sub|stmt-boundary|canon-adds-paren-op|canon-adds-paren-term)[[:space:]]" $f'`)
- [ ] Regression guard: the forgiven set is still exactly the two ruled differences, a block's last `;` and extra nesting (`go test ./internal/parse/ -run '^TestFaithfulForgivenessIsClosed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulForgivenessIsClosed'`)

### Negative Scenarios

- [ ] A nested parenless call keeps perl's argument grouping. perl 5.42 Deparses `foo 1, bar 2, 3` (both subs predeclared) as `foo(1, bar(2, 3))`, so a hand-built tree grouped `foo(1, bar(2), 3)` emits `foo 1, bar (2), 3` and stays unfaithful by construction (A29, round 7) (`go test ./internal/parse/ -run '^TestCanonAddedParenOnNestedListOpKeepsPerlGrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenOnNestedListOpKeepsPerlGrouping'`)
- [ ] A comparison chain is not regrouped. perl Deparses `1 < $x < 3` with no inner parens, while `(1 < $x) < 3` compiles to two separate `lt` ops (the inner one `sKP/2`), so a canon `(1 < $x) < 3` is a misgrouping (`go test ./internal/parse/ -run '^TestCanonAddedParenOnCmpChainIsMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenOnCmpChainIsMisgrouping'`)
- [ ] No forgiveness covers a call's argument list: `print (1+2)*3` parsed as `print((1+2)*3)` emits `print ((1+2)*3)`, which is a different token sequence from the source and unfaithful (`go test ./internal/parse/ -run '^TestParenlessCallForgivenessIgnoresInnerTermParens$' -count=1 -v | grep -q '^--- PASS: TestParenlessCallForgivenessIgnoresInnerTermParens'`)
- [ ] A bareword or filehandle that perl reads as a call argument, not a call, is not given call parens and then forgiven (`go test ./internal/parse/ -run '^TestCanonAddedParenOnBarewordMatchesPerl$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenOnBarewordMatchesPerl'`)
