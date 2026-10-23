---
title: Judge canon-added parens on terms and argument lists against perl
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-608d-71ed-8ab2-66a9951d9957
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
created: 2026-10-22T23:46:34.345543908Z
updated: 2026-10-23T01:11:04.82819858Z
---

## Context

- paths: `internal/parse/canon.go`, `internal/parse/fidelity.go`, the parser files the fix touches, `internal/parse/fidelity_test.go`, `internal/parse/testdata/faithful_defects.txt`, `internal/parse/testdata/perlgittfaithful.ratchet`
- commands: `go test ./internal/parse/ -run '^TestCanonAddedParensOnTermsMatchPerl$' -count=1 -v`
- entrypoints: `internal/parse/canon.go` (binding-power parens), `internal/parse/fidelity.go` (parenless-call forgiveness)

These are the `canon-adds-paren-term` rows of faithful_defects.txt, where canon writes a `(` the source lacks, and the source token at the first difference is a term, literal, bareword or `;`/`{` (most likely a list-op or call argument list canon parenthesises). At 2026-10-22, examples were op/cmpchain.t, op/avhv.t and op/tr.t. canon writes the parens binding power implies plus a `Paren` node's pair, so an added paren is either an unforgiven parenless call or a misgrouped tree. That makes these the most important rows in the list. No case may be called a misgrouping, or forgiven, before it is judged against perl with `perl -MO=Deparse,-p` or an op tree. It also owns the `other` rows, as the last issue in the triage sequence: each is classified, then fixed or kept with a perl measurement in its reason. Scope is the rows of these classes in faithful_defects.txt when this issue starts, not a fixed file list. The 2026-10-22 first-difference list is only an example, because the Paren change exposes later differences. A fix that changes a row's class re-classifies it, and if that class's owner has already closed, this issue handles it. The triage issues run in sequence (heredoc/sub-less, then statement-boundary, then operator, then term) because they all edit canon.go, fidelity.go, fidelity_test.go, faithful_defects.txt and perlgittfaithful.ratchet.

## Steps

- [ ] For each `canon-adds-paren-term` row and each `other` row, reduce the first difference to a minimal snippet and run the pinned perl 5.44.0's `perl -MO=Deparse,-p` on it. Record whether perl's grouping matches the tree's (an unforgiven parenless call, cosmetic) or differs (a parser misgrouping).
- [ ] RED: add each reduced snippet to TestCanonAddedParensOnTermsMatchPerl, asserting perl's grouping and quoting the Deparse output in a comment. Run it and watch the cases fail.
- [ ] GREEN: fix one cause at a time. A misgrouping is fixed in the parser. A parenless call is forgiven only through the existing parenless-call rule, widened no further than the measurement supports. Run the test after each fix.
- [ ] Run TestCanonCatchesMisgrouping and TestCanonQuotientIsNarrow after every fix, and add each misgrouping found here as a named case to TestCanonCatchesMisgrouping.
- [ ] Remove the fixed rows from faithful_defects.txt. Give every kept row its line and the Deparse result that explains it, and re-baseline perlgittfaithful.ratchet (and the T1 ratchets if they move) in the same commit, naming each moved file.
- [ ] Add every reduced case, with perl's facts, as probe file `internal/parse/testdata/m3probes/canon-paren-term.txt`.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus, and adds them to this issue's probe file under internal/parse/testdata/m3probes/ so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PATH=$HOME/.local/share/pvm/versions/5.44.0/bin:$PATH PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] RED/GREEN TestOtherDefectRowsQuotePerl: every kept `other` row quotes a perl measurement in its reason.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] Each reduced case groups as perl's Deparse,-p does, or is pinned with that measurement (`go test ./internal/parse/ -run '^TestCanonAddedParensOnTermsMatchPerl$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParensOnTermsMatchPerl'`)
- [ ] Regression guard, last in the triage sequence: canon still catches misgrouping (`go test ./internal/parse/ -run '^TestCanonCatchesMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonCatchesMisgrouping'`)
- [ ] The defects list and the M3-dir check agree after the fixes (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulOnM3Dirs$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulOnM3Dirs'`)
- [ ] The perl.git t/ faithful ratchet matches the tree after this issue's re-baseline (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchet'`)
- [ ] The reduced cases are committed as perl-judged probes (`test -s internal/parse/testdata/m3probes/canon-paren-term.txt`)
- [ ] No `other` row remains unclassified or unmeasured (`go test ./internal/parse/ -run '^TestOtherDefectRowsQuotePerl$' -count=1 -v | grep -q '^--- PASS: TestOtherDefectRowsQuotePerl'`)

### Negative Scenarios

- [ ] A nested parenless call keeps perl's argument grouping. perl 5.42 Deparses `foo 1, bar 2, 3` (both subs predeclared) as `foo(1, bar(2, 3))`, so a canon `foo(1, bar(2), 3)` is a misgrouping and stays unfaithful (`go test ./internal/parse/ -run '^TestCanonAddedParenOnNestedListOpKeepsPerlGrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenOnNestedListOpKeepsPerlGrouping'`)
- [ ] A comparison chain is not regrouped. perl Deparses `1 < $x < 3` with no inner parens, while `(1 < $x) < 3` compiles to two separate `lt` ops (the inner one `sKP/2`), so a canon `(1 < $x) < 3` is a misgrouping (`go test ./internal/parse/ -run '^TestCanonAddedParenOnCmpChainIsMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenOnCmpChainIsMisgrouping'`)
- [ ] The parenless-call forgiveness applies only to a call's own argument list; a pair canon adds around a term inside the arguments, such as `print (1+2)*3` written as `print((1+2)*3)`, is not forgiven by it (`go test ./internal/parse/ -run '^TestParenlessCallForgivenessIgnoresInnerTermParens$' -count=1 -v | grep -q '^--- PASS: TestParenlessCallForgivenessIgnoresInnerTermParens'`)
- [ ] A bareword or filehandle that perl reads as a call argument, not a call, is not given call parens and then forgiven (`go test ./internal/parse/ -run '^TestCanonAddedParenOnBarewordMatchesPerl$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenOnBarewordMatchesPerl'`)
- [ ] Every row kept in faithful_defects.txt for the `canon-adds-paren-term` class carries its Deparse,-p result in the reason (`go test ./internal/parse/ -run '^TestCanonAddsParenTermRowsQuoteDeparse$' -count=1 -v | grep -q '^--- PASS: TestCanonAddsParenTermRowsQuoteDeparse'`)
