---
title: Re-emit heredoc bodies and sub-less special subs faithfully
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5f77-78ef-a0c7-538e0a241cdc
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
- 01a16602-6030-7faa-8408-bb3a8cb03f12
created: 2026-10-22T23:46:34.069628305Z
updated: 2026-10-23T01:35:40.723740867Z
---

## Context

- paths: `internal/parse/fidelity.go`, `internal/parse/canon.go`, `internal/parse/fidelity_test.go`, `internal/parse/testdata/faithful_defects.txt`, `internal/parse/testdata/perlgittfaithful.ratchet`
- commands: `go test ./internal/parse/ -run '^(TestFaithfulHeredocBodyIsNotAToken|TestFaithfulSublessSpecialSub)$' -count=1 -v`
- entrypoints: `internal/parse/fidelity.go` (significant-token extraction), `internal/parse/canon.go` (sub declarations)

These are the `heredoc` and `subless-special-sub` rows of faithful_defects.txt: two token-level causes. (1) Heredoc bodies show up as significant tokens on the source side (at 2026-10-22, e.g. op/debug.t, op/threads.t, and behind dropped parens op/bop.t, op/eval.t, op/pack.t). A heredoc body is a string literal's content, and the two sides must tokenize it alike. (2) canon writes `sub` where the source has `DESTROY` without it (e.g. op/hash.t, op/vec.t). perl accepts `sub`-less declarations for some names. Whether canon should write `sub` there (a forgiven difference) or write what the source wrote is decided by measuring perl. Neither cause is a grouping question, so neither may weaken misgrouping detection. Scope is the rows of these classes in faithful_defects.txt when this issue starts, not a fixed file list. The 2026-10-22 first-difference list is only an example, because the Paren change exposes later differences. A fix that changes a row's class re-classifies it, and if that class's owner has already closed, this issue handles it. The triage issues run in sequence after the OPf_PARENS issue, which owns the `dropped-paren` rows (OPf_PARENS, then heredoc/sub-less, then statement-boundary, then operator, then term), because they all edit canon.go, fidelity.go, fidelity_test.go, faithful_defects.txt and perlgittfaithful.ratchet.

## Steps

- [ ] Measure with the pinned perl 5.44.0 first. Run `perl -MO=Deparse` and `perl -c` on `DESTROY { 1 }`, `AUTOLOAD { 1 }`, `BEGIN { 1 }` and `foo { 1 }` (the last should not be a declaration), and on a heredoc inside a call's argument list. Quote the results in the tests' comments.
- [ ] RED: write TestFaithfulHeredocBodyIsNotAToken, minimal fixtures reduced from op/debug.t and op/threads.t. Run it and watch it fail.
- [ ] GREEN: fix the token extraction on whichever side is wrong. Run the test and watch it pass.
- [ ] RED: write TestFaithfulSublessSpecialSub for `DESTROY {...}` and the other names perl accepted above. Run it and watch it fail.
- [ ] GREEN: fix it, either as a forgiven difference limited to exactly the names perl accepts or in canon, whichever the measurement supports. Run the test and watch it pass.
- [ ] Remove the fixed rows from faithful_defects.txt and re-baseline perlgittfaithful.ratchet in the same commit, naming each moved file in the message.
- [ ] Add every reduced case, with perl's facts, as probe file `internal/parse/testdata/m3probes/heredoc-subless.txt`.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus, and adds them to this issue's probe file under internal/parse/testdata/m3probes/ so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] A heredoc body is compared as one literal, not as source tokens (`go test ./internal/parse/ -run '^TestFaithfulHeredocBodyIsNotAToken$' -count=1 -v | grep -q '^--- PASS: TestFaithfulHeredocBodyIsNotAToken'`)
- [ ] A sub-less special-sub declaration re-emits faithfully, for exactly the names perl accepts (`go test ./internal/parse/ -run '^TestFaithfulSublessSpecialSub$' -count=1 -v | grep -q '^--- PASS: TestFaithfulSublessSpecialSub'`)
- [ ] The defects list and the M3-dir check agree after the fixes (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulOnM3Dirs$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulOnM3Dirs'`)
- [ ] No heredoc or sub-less special-sub row remains in the defects list (`sh -c 'f=internal/parse/testdata/faithful_defects.txt; test -f $f && ! grep -Eq "^[^ ]+ (heredoc|subless-special-sub) " $f'`)
- [ ] The reduced cases are committed as perl-judged probes (`test -s internal/parse/testdata/m3probes/heredoc-subless.txt`)

### Negative Scenarios

- [ ] A heredoc body that differs between source and canon is still a difference; treating the body as one literal must not make its content invisible (`go test ./internal/parse/ -run '^TestFaithfulHeredocBodyChangeIsADifference$' -count=1 -v | grep -q '^--- PASS: TestFaithfulHeredocBodyChangeIsADifference'`)
- [ ] An open paren or a semicolon inside a heredoc body never pairs with, or offsets, a paren or statement boundary in the code around it (`go test ./internal/parse/ -run '^TestFaithfulHeredocBodyPunctuationIsInert$' -count=1 -v | grep -q '^--- PASS: TestFaithfulHeredocBodyPunctuationIsInert'`)
- [ ] Two heredocs started on one line, such as `f(<<A, <<B)`, keep their bodies in order; swapping them is a difference (`go test ./internal/parse/ -run '^TestFaithfulTwoHeredocsKeepOrder$' -count=1 -v | grep -q '^--- PASS: TestFaithfulTwoHeredocsKeepOrder'`)
- [ ] An indented `<<~EOT` heredoc and an interpolating `<<"EOT"` heredoc follow the same rule as a plain one, with no source-side tokens leaking (`go test ./internal/parse/ -run '^TestFaithfulIndentedAndInterpolatingHeredocs$' -count=1 -v | grep -q '^--- PASS: TestFaithfulIndentedAndInterpolatingHeredocs'`)
- [ ] The sub-less form is forgiven for no name perl does not accept. perl 5.42 Deparses `DESTROY { 1 }`, `AUTOLOAD { 1 }` and `BEGIN { 1 }` as `sub NAME {...}`, but leaves `foo { 1 }` and `new { 1 }` as an expression statement with no `sub`. A canon `sub foo` against source `foo { 1 }` stays unfaithful (`go test ./internal/parse/ -run '^TestFaithfulSublessFormRejectsOrdinaryNames$' -count=1 -v | grep -q '^--- PASS: TestFaithfulSublessFormRejectsOrdinaryNames'`)
- [ ] The sub-less forgiveness covers only the missing `sub` keyword; a different body or a different name after it is still a difference (`go test ./internal/parse/ -run '^TestFaithfulSublessSpecialSubBodyStillCompared$' -count=1 -v | grep -q '^--- PASS: TestFaithfulSublessSpecialSubBodyStillCompared'`)
