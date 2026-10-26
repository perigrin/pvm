---
title: Judge canon-added parens on operators against perl
state: in-progress
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-6030-7faa-8408-bb3a8cb03f12
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
- 01a16602-60e9-7343-a630-7a93da5ee389
created: 2026-10-22T23:46:34.2538574Z
updated: 2026-10-26T00:16:33.417440461Z
sessions:
- start_sha: 3b77ae24bd593b2eb20e829fdc36fd5135584b9a
  end_sha: ""
  commits: 0
  started_at: 2026-10-26T00:16:33.417440461Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-26T00:16:33.417440461Z
---

## Context

- paths: `internal/parse/canon.go`, `internal/parse/fidelity.go`, the parser files the fix touches, `internal/parse/fidelity_test.go`, `internal/parse/testdata/faithful_queue.txt`, `internal/parse/testdata/perlgittfaithful.ratchet`
- commands: `go test ./internal/parse/ -run '^TestCanonAddedParensOnOperatorsMatchPerl$' -count=1 -v`
- entrypoints: `internal/parse/canon.go` (binding-power parens, the follow power, the named-unary and list-op branches of the `Call` case), `internal/parse/precedence.go` (`derivePowers`)

These are the `canon-adds-paren-op` rows of faithful_queue.txt, where canon writes a `(` the source lacks after an operator, a declarator, a named unary op, `not`, a word operator or `sub` (5f77's rule 7: `\my`, `- -`, `= not`, `** -`, `delete $::`). At 2026-10-22, examples were op/not.t, op/negate.t and op/undef.t. Under the round-7 ruling canon is precedence-aware and writes a call's parens only where the source had them (`ArgParen`) or where binding power requires them, plus a `Paren` node's pair. So an added paren here is one of two things: canon's binding-power table disagrees with perl's precedence (a canon defect, fixed in canon or `derivePowers`), or the tree groups the operands differently from perl (a parser misgrouping). There is no forgiveness to apply. That makes these the most important rows in the list. No case may be called a misgrouping, or forgiven, before it is judged against perl with `perl -MO=Deparse,-p` or an op tree. Scope is the rows of these classes in faithful_queue.txt when this issue starts, not a fixed file list. The 2026-10-22 first-difference list is only an example, because the Paren change exposes later differences. A fix that changes a row's class re-classifies it, and if that class's owner has already closed, this issue fixes it; the zero-row criterion covers every earlier owner's class to hold that (J1, chain-review round 5). The triage issues run in sequence after the dropped-paren rows issue (3323), which owns the `dropped-paren` rows and itself follows the OPf_PARENS check (OPf_PARENS, then dropped-paren rows, then heredoc/sub-less, then statement-boundary, then operator, then term), because they all edit canon.go, fidelity.go, fidelity_test.go, faithful_queue.txt and perlgittfaithful.ratchet. **Why the sequence stays (A23, chain-review round 3):** the shared-generated-files rule lets parallel branches regenerate faithful_queue.txt and the ratchets. It does not cover canon.go, fidelity.go and fidelity_test.go, which are hand-written and edited by every one of these issues. Running them in sequence also gives 5f77's handover rule a defined order: a fix that moves a row into a class whose owner has closed is handled by the issue that moved it, which only works if owners close in a known order. The cost is a longer critical chain, accepted over hand-merging four sets of canon/fidelity changes.

**These negatives are satisfiable by construction (A28, round-7 ruling).** The call-paren rule is gone, so a canon `defined($x || $y)` or `not($x or 1)` from a wrong tree is a paren the source lacks and unfaithful. Each negative builds the wrong tree by hand (regroup the parsed tree, as docs/measurements/2026-10-23-precedence-aware-canon/README.md did), runs `Canon`, and judges the result with `Faithful` against the source.

## Steps

- [ ] FIRST, before any fix: read this class's `# causes <class> = N` line in faithful_queue.txt's header (5f77). If N is over 8, split this issue by cause using 5f77's sibling template before starting. This issue's body then keeps only its own cause group, and the cumulative class-zero criterion moves to the last sibling (J8, chain-review round 6).
- [ ] For each `canon-adds-paren-op` row, reduce the first difference to a minimal snippet and run the pinned perl 5.44.0's `perl -MO=Deparse,-p` on it. Record whether perl's grouping matches the tree's (then canon's binding power for that operator is wrong) or differs (a parser misgrouping). Neither is fixed by a forgiveness rule.
- [ ] RED: add each reduced snippet to TestCanonAddedParensOnOperatorsMatchPerl, asserting perl's grouping and quoting the Deparse output in a comment. Run it and watch the cases fail.
- [ ] GREEN: fix one cause at a time. A misgrouping is fixed in the parser. A wrong binding power is fixed in canon's table or in `derivePowers`, with the CORE.tpm line it derives from. `Faithful` is not changed (perigrin's round-7 ruling: the forgiven set is closed at a block's last `;` and extra paren nesting). A difference that seems to need a new rule goes to perigrin with its perl measurement. Each row is fixed one way or the other; the class must reach zero. Run the test after each fix.
- [ ] Run TestCanonCatchesMisgrouping and TestCanonQuotientIsNarrow after every fix, and add each misgrouping found here as a named case to TestCanonCatchesMisgrouping.
- [ ] Remove the fixed rows from faithful_queue.txt, and re-baseline perlgittfaithful.ratchet (and the T1 ratchets if they move) in the same commit, naming each moved file.
- [ ] Add every reduced case, with perl's facts, as probe file `internal/parse/testdata/m3probes/canon-paren-op.txt`.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus, and adds them to this issue's probe file under internal/parse/testdata/m3probes/ so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] If a row proves unfixable, stop and escalate it to perigrin with the perl measurement. Never leave it in the queue, and never close this issue with a row of its class left (perigrin's ruling, chain-review round 4).
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] Each reduced case groups as perl's Deparse,-p does (`go test ./internal/parse/ -run '^TestCanonAddedParensOnOperatorsMatchPerl$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParensOnOperatorsMatchPerl'`)
- [ ] The work queue still matches the tree after the fixes (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueMatchesTree$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueMatchesTree'`)
- [ ] The reduced cases are committed as perl-judged probes, held by TestM3ProbesHeldToPerl (`sh -c 'test -s internal/parse/testdata/m3probes/canon-paren-op.txt && go test ./internal/parse/ -run "^TestM3ProbesHeldToPerl$" -count=1 -v | grep -q "^--- PASS: TestM3ProbesHeldToPerl"'`)
- [ ] No row of this issue's class, or of any earlier owner's class (dropped-paren, heredoc, subless-special-sub, stmt-boundary, canon-adds-paren-op), is left in the work queue; rows a fix here exposed in a closed owner's class are fixed here (`sh -c 'f=internal/parse/testdata/faithful_queue.txt; test -f $f && ! grep -Eq "^[^[:space:]#]+[[:space:]]+(dropped-paren|heredoc|subless-special-sub|stmt-boundary|canon-adds-paren-op)[[:space:]]" $f'`)
- [ ] Regression guard: the forgiven set is still exactly the two ruled differences, a block's last `;` and extra nesting (`go test ./internal/parse/ -run '^TestFaithfulForgivenessIsClosed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulForgivenessIsClosed'`)

### Negative Scenarios

- [ ] A named unary op's parens are never forgiven when they would regroup its operand. perl 5.42 Deparses `defined $x || $y` as `(defined($x) or $y);` (the grouping, not the spelling of the operator, is what counts), so a canon `defined($x || $y)` is a misgrouping and stays unfaithful (`go test ./internal/parse/ -run '^TestCanonAddedParenOnNamedUnaryRejectsRegrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenOnNamedUnaryRejectsRegrouping'`)
- [ ] `not` keeps its low precedence. perl Deparses `my $v = not $x or 1` as `((my $v = (!$x)) or '???')`, so a canon that writes `not($x or 1)` is a misgrouping and stays unfaithful (`go test ./internal/parse/ -run '^TestCanonAddedParenOnNotKeepsLowPrecedence$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenOnNotKeepsLowPrecedence'`)
- [ ] Unary minus binds looser than `**`. perl folds `-2 ** 2` to `-4.0`, so a canon `(-2) ** 2` is a misgrouping and stays unfaithful (`go test ./internal/parse/ -run '^TestCanonAddedParenOnNegateBindsLooserThanPow$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenOnNegateBindsLooserThanPow'`)
- [ ] A paren canon adds after `\` is never forgiven. perl compiles `\(@a)` to `refgen` over a list and `\@a` to `srefgen`, so a canon `\(@a)` against source `\@a` is a different tree (`go test ./internal/parse/ -run '^TestCanonAddedParenAfterBackslashIsNotForgiven$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenAfterBackslashIsNotForgiven'`)
