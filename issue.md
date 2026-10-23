---
title: Judge canon-added parens on operators against perl
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-6030-7faa-8408-bb3a8cb03f12
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
- 01a16602-60e9-7343-a630-7a93da5ee389
created: 2026-10-22T23:46:34.2538574Z
updated: 2026-10-23T01:11:03.360509546Z
---

## Context

- paths: `internal/parse/canon.go`, `internal/parse/fidelity.go`, the parser files the fix touches, `internal/parse/fidelity_test.go`, `internal/parse/testdata/faithful_defects.txt`, `internal/parse/testdata/perlgittfaithful.ratchet`
- commands: `go test ./internal/parse/ -run '^TestCanonAddedParensOnOperatorsMatchPerl$' -count=1 -v`
- entrypoints: `internal/parse/canon.go` (binding-power parens), `internal/parse/fidelity.go` (parenless-call forgiveness)

These are the `canon-adds-paren-op` rows of faithful_defects.txt, where canon writes a `(` the source lacks, and the source token at the first difference is an operator or named op (`not`, unary `-`/`+`, `\`, `defined`, `eval`, `shift`, `sub`, `my`). At 2026-10-22, examples were op/not.t, op/negate.t and op/undef.t. canon writes the parens binding power implies plus a `Paren` node's pair, so an added paren is either an unforgiven parenless call or a misgrouped tree. That makes these the most important rows in the list. No case may be called a misgrouping, or forgiven, before it is judged against perl with `perl -MO=Deparse,-p` or an op tree. Scope is the rows of these classes in faithful_defects.txt when this issue starts, not a fixed file list. The 2026-10-22 first-difference list is only an example, because the Paren change exposes later differences. A fix that changes a row's class re-classifies it, and if that class's owner has already closed, this issue handles it. The triage issues run in sequence (heredoc/sub-less, then statement-boundary, then operator, then term) because they all edit canon.go, fidelity.go, fidelity_test.go, faithful_defects.txt and perlgittfaithful.ratchet.

## Steps

- [ ] For each `canon-adds-paren-op` row, reduce the first difference to a minimal snippet and run the pinned perl 5.44.0's `perl -MO=Deparse,-p` on it. Record whether perl's grouping matches the tree's (an unforgiven parenless call, cosmetic) or differs (a parser misgrouping).
- [ ] RED: add each reduced snippet to TestCanonAddedParensOnOperatorsMatchPerl, asserting perl's grouping and quoting the Deparse output in a comment. Run it and watch the cases fail.
- [ ] GREEN: fix one cause at a time. A misgrouping is fixed in the parser. A parenless call is forgiven only through the existing parenless-call rule, widened no further than the measurement supports. Run the test after each fix.
- [ ] Run TestCanonCatchesMisgrouping and TestCanonQuotientIsNarrow after every fix, and add each misgrouping found here as a named case to TestCanonCatchesMisgrouping.
- [ ] Remove the fixed rows from faithful_defects.txt. Give every kept row its line and the Deparse result that explains it, and re-baseline perlgittfaithful.ratchet (and the T1 ratchets if they move) in the same commit, naming each moved file.
- [ ] Add every reduced case, with perl's facts, as probe file `internal/parse/testdata/m3probes/canon-paren-op.txt`.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus, and adds them to this issue's probe file under internal/parse/testdata/m3probes/ so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] A parser change can move the oracle ratchets, because Sites() reads the tree. Before committing, run the sweep (`PATH=$HOME/.local/share/pvm/versions/5.44.0/bin:$PATH PERL5_CORPUS=$HOME/dev/perl5 PARSEORACLE_SHIM=$HOME/.cache/pvm/oracleshim PARSEORACLE_TIMEOUT=8m go test ./internal/parseoracle/ -run '^(TestRatchetCorpus|TestSiteRatchetCorpus)$' -parseoracle.corpus -count=1 -timeout 40m -v`). If baseline.txt or sites.txt moves, re-baseline with `-parseoracle.update` in the same commit and explain each moved file. WRONG must stay 0.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] Each reduced case groups as perl's Deparse,-p does, or is pinned with that measurement (`go test ./internal/parse/ -run '^TestCanonAddedParensOnOperatorsMatchPerl$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParensOnOperatorsMatchPerl'`)
- [ ] The defects list and the M3-dir check agree after the fixes (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulOnM3Dirs$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulOnM3Dirs'`)
- [ ] The reduced cases are committed as perl-judged probes (`test -s internal/parse/testdata/m3probes/canon-paren-op.txt`)

### Negative Scenarios

- [ ] A named unary op's parens are never forgiven when they would regroup its operand. perl 5.42 Deparses `defined $x || $y` as `(defined($x) or $y);` (the grouping, not the spelling of the operator, is what counts), so a canon `defined($x || $y)` is a misgrouping and stays unfaithful (`go test ./internal/parse/ -run '^TestCanonAddedParenOnNamedUnaryRejectsRegrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenOnNamedUnaryRejectsRegrouping'`)
- [ ] `not` keeps its low precedence. perl Deparses `my $v = not $x or 1` as `((my $v = (!$x)) or '???')`, so a canon that writes `not($x or 1)` is a misgrouping and stays unfaithful (`go test ./internal/parse/ -run '^TestCanonAddedParenOnNotKeepsLowPrecedence$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenOnNotKeepsLowPrecedence'`)
- [ ] Unary minus binds looser than `**`. perl folds `-2 ** 2` to `-4.0`, so a canon `(-2) ** 2` is a misgrouping and stays unfaithful (`go test ./internal/parse/ -run '^TestCanonAddedParenOnNegateBindsLooserThanPow$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenOnNegateBindsLooserThanPow'`)
- [ ] A paren canon adds after `\` is never forgiven. perl compiles `\(@a)` to `refgen` over a list and `\@a` to `srefgen`, so a canon `\(@a)` against source `\@a` is a different tree (`go test ./internal/parse/ -run '^TestCanonAddedParenAfterBackslashIsNotForgiven$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenAfterBackslashIsNotForgiven'`)
- [ ] Every row kept in faithful_defects.txt for the `canon-adds-paren-op` class carries its Deparse,-p result in the reason, so no `canon-adds-paren-op` row is parked without a measurement (`go test ./internal/parse/ -run '^TestCanonAddsParenOpRowsQuoteDeparse$' -count=1 -v | grep -q '^--- PASS: TestCanonAddsParenOpRowsQuoteDeparse'`)
