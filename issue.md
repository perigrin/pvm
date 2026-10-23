---
title: Re-emit heredoc bodies and sub-less special subs faithfully
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5f77-78ef-a0c7-538e0a241cdc
- 01a167da-3323-7ce0-a627-abf5b99603d8
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
- 01a16602-6030-7faa-8408-bb3a8cb03f12
created: 2026-10-22T23:46:34.069628305Z
updated: 2026-10-23T08:21:59.596075066Z
---

## Context

- paths: `internal/parse/fidelity.go`, `internal/parse/canon.go`, `internal/parse/fidelity_test.go`, `internal/parse/testdata/faithful_queue.txt`, `internal/parse/testdata/perlgittfaithful.ratchet`
- commands: `go test ./internal/parse/ -run '^(TestCanonWritesHeredocForm|TestCanonWritesSublessSpecialSub|TestFaithfulForgivenessIsClosed)$' -count=1 -v`
- entrypoints: `internal/parse/fidelity.go` (significant-token extraction), `internal/parse/canon.go` (sub declarations)

These are the `heredoc` and `subless-special-sub` rows of faithful_queue.txt. **perigrin's ruling (2026-10-23, chain-review round 5): `Faithful` forgives exactly three differences (call parens, a block's last `;`, extra paren nesting), and the set is closed.** Neither cause here may be forgiven. The tree records each spelling, and canon writes it back, the way `Paren` works.
1. Heredocs: canon writes a heredoc in heredoc form (`<<EOT`, `<<~EOT`, `<<"EOT"`, `<<'EOT'`, with the body and terminator after the line), recorded in the tree. Both sides then tokenize the same way. At 2026-10-22, examples were op/debug.t and op/threads.t, and, behind dropped parens, op/bop.t, op/eval.t and op/pack.t.
2. Sub-less special subs: the tree records that `DESTROY { ... }` was written without `sub`, and canon writes it that way. Only names perl accepts in that form qualify (measured below). Examples were op/hash.t and op/vec.t.

Neither cause is a grouping question, so neither may weaken misgrouping detection. A difference that seems to need a new forgiveness rule goes to perigrin with its perl measurement; this issue never adds one. Scope is the rows of these classes in faithful_queue.txt when this issue starts, not a fixed file list. The 2026-10-22 first-difference list is only an example, because the Paren change exposes later differences. A fix that changes a row's class re-classifies it, and if that class's owner has already closed, this issue fixes it; the zero-row criterion covers every earlier owner's class to hold that (J1, chain-review round 5). The triage issues run in sequence after the OPf_PARENS issue, which owns the `dropped-paren` rows (OPf_PARENS, then heredoc/sub-less, then statement-boundary, then operator, then term), because they all edit canon.go, fidelity.go, fidelity_test.go, faithful_queue.txt and perlgittfaithful.ratchet. **Why the sequence stays (A23, chain-review round 3):** the shared-generated-files rule lets parallel branches regenerate faithful_queue.txt and the ratchets. It does not cover canon.go, fidelity.go and fidelity_test.go, which are hand-written and edited by every one of these issues. Running them in sequence also gives 5f77's handover rule a defined order: a fix that moves a row into a class whose owner has closed is handled by the issue that moved it, which only works if owners close in a known order. The cost is a longer critical chain, accepted over hand-merging four sets of canon/fidelity changes.

## Steps

- [ ] Measure with the pinned perl 5.44.0 first. Run `perl -MO=Deparse` and `perl -c` on `DESTROY { 1 }`, `AUTOLOAD { 1 }`, `BEGIN { 1 }` and `foo { 1 }` (the last should not be a declaration), and on a heredoc inside a call's argument list. Quote the results in the tests' comments.
- [ ] RED: write TestCanonWritesHeredocForm, with minimal fixtures reduced from op/debug.t and op/threads.t. Each heredoc re-emits in its source form, and the file is faithful with no forgiveness. Run it and watch it fail.
- [ ] GREEN: record the heredoc form in the tree and write it in canon. Do not touch `Faithful`. Run the test and watch it pass.
- [ ] RED: write TestCanonWritesSublessSpecialSub for `DESTROY {...}` and the other names perl accepted above. Each re-emits without `sub`, and is faithful with no forgiveness. Run it and watch it fail.
- [ ] GREEN: record the sub-less form in the tree and write it in canon, only for the names perl accepts. Do not touch `Faithful`. Run the test and watch it pass, together with TestFaithfulForgivenessIsClosed.
- [ ] Remove the fixed rows from faithful_queue.txt and re-baseline perlgittfaithful.ratchet in the same commit, naming each moved file in the message.
- [ ] Add every reduced case, with perl's facts, as probe file `internal/parse/testdata/m3probes/heredoc-subless.txt`.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus, and adds them to this issue's probe file under internal/parse/testdata/m3probes/ so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] If a row proves unfixable, stop and escalate it to perigrin with the perl measurement. Never leave it in the queue, and never close this issue with a row of its class left (perigrin's ruling, chain-review round 4).
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] canon writes each heredoc in its source form, and the file is faithful without any forgiveness (`go test ./internal/parse/ -run '^TestCanonWritesHeredocForm$' -count=1 -v | grep -q '^--- PASS: TestCanonWritesHeredocForm'`)
- [ ] canon writes a sub-less special sub without `sub`, for exactly the names perl accepts, and the file is faithful without any forgiveness (`go test ./internal/parse/ -run '^TestCanonWritesSublessSpecialSub$' -count=1 -v | grep -q '^--- PASS: TestCanonWritesSublessSpecialSub'`)
- [ ] The work queue still matches the tree after the fixes (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestFaithfulQueueMatchesTree$' -count=1 -v | grep -q '^--- PASS: TestFaithfulQueueMatchesTree'`)
- [ ] No row of this issue's class, or of any earlier owner's class (dropped-paren, heredoc, subless-special-sub), is left in the work queue; rows a fix here exposed in a closed owner's class are fixed here (`sh -c 'f=internal/parse/testdata/faithful_queue.txt; test -f $f && ! grep -Eq "^[^[:space:]#]+[[:space:]]+(dropped-paren|heredoc|subless-special-sub)[[:space:]]" $f'`)
- [ ] The reduced cases are committed as perl-judged probes (`test -s internal/parse/testdata/m3probes/heredoc-subless.txt`)
- [ ] Regression guard: the forgiven set is still exactly the three ruled differences (`go test ./internal/parse/ -run '^TestFaithfulForgivenessIsClosed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulForgivenessIsClosed'`)

### Negative Scenarios

- [ ] A heredoc body that differs between source and canon is still a difference; writing the heredoc form must not make its content invisible (`go test ./internal/parse/ -run '^TestFaithfulHeredocBodyChangeIsADifference$' -count=1 -v | grep -q '^--- PASS: TestFaithfulHeredocBodyChangeIsADifference'`)
- [ ] An open paren or a semicolon inside a heredoc body never pairs with, or offsets, a paren or statement boundary in the code around it (`go test ./internal/parse/ -run '^TestFaithfulHeredocBodyPunctuationIsInert$' -count=1 -v | grep -q '^--- PASS: TestFaithfulHeredocBodyPunctuationIsInert'`)
- [ ] Two heredocs started on one line, such as `f(<<A, <<B)`, keep their bodies in order; swapping them is a difference (`go test ./internal/parse/ -run '^TestFaithfulTwoHeredocsKeepOrder$' -count=1 -v | grep -q '^--- PASS: TestFaithfulTwoHeredocsKeepOrder'`)
- [ ] An indented `<<~EOT` heredoc and an interpolating `<<"EOT"` heredoc follow the same rule as a plain one, with no source-side tokens leaking (`go test ./internal/parse/ -run '^TestFaithfulIndentedAndInterpolatingHeredocs$' -count=1 -v | grep -q '^--- PASS: TestFaithfulIndentedAndInterpolatingHeredocs'`)
- [ ] canon never writes the sub-less form for a name perl does not accept. perl 5.42 Deparses `DESTROY { 1 }`, `AUTOLOAD { 1 }` and `BEGIN { 1 }` as `sub NAME {...}`, but leaves `foo { 1 }` and `new { 1 }` as an expression statement, so source `foo { 1 }` never re-emits as a declaration (`go test ./internal/parse/ -run '^TestFaithfulSublessFormRejectsOrdinaryNames$' -count=1 -v | grep -q '^--- PASS: TestFaithfulSublessFormRejectsOrdinaryNames'`)
- [ ] A sub-less special sub whose body or name differs between source and canon is still a difference (`go test ./internal/parse/ -run '^TestFaithfulSublessSpecialSubBodyStillCompared$' -count=1 -v | grep -q '^--- PASS: TestFaithfulSublessSpecialSubBodyStillCompared'`)
