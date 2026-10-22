---
title: Judge canon-added parens on operators against perl
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5f77-78ef-a0c7-538e0a241cdc
blocks:
- 01a16602-61fe-7207-8f41-b24de8b9b960
created: 2026-10-22T23:46:34.2538574Z
updated: 2026-10-22T23:56:40.621779507Z
---

## Context

- paths: `internal/parse/canon.go`, `internal/parse/fidelity.go`, the parser files the fix touches, `internal/parse/fidelity_test.go`, `internal/parse/testdata/faithful_defects.txt`, `internal/parse/testdata/perlgittfaithful.ratchet`
- commands: `go test ./internal/parse/ -run '^TestCanonAddedParensOnOperatorsMatchPerl$' -count=1 -v`
- entrypoints: `internal/parse/canon.go` (binding-power parens), `internal/parse/fidelity.go` (parenless-call forgiveness)

These are the `canon-adds-paren` rows of faithful_defects.txt, where canon writes a `(` the source lacks, and the source token at the first difference is an operator or named op (`not`, unary `-`/`+`, `\`, `defined`, `eval`, `shift`, `sub`, `my`): 11 files. canon writes the parens binding power implies plus a `Paren` node's pair, so an added paren is either an unforgiven parenless call or a misgrouped tree. That makes these the most important rows in the list. No case may be called a misgrouping, or forgiven, before it is judged against perl with `perl -MO=Deparse,-p` or an op tree. Files: op/not.t, op/isa.t, op/negate.t, op/hash-rt85026.t, op/bless.t, uni/bless.t, op/undef.t, op/lexsub.t, uni/method.t, op/gmagic.t, op/lvref.t.

## Steps

- [ ] For each file, reduce the first difference to a minimal snippet and run `perl -MO=Deparse,-p` on it. Record whether perl's grouping matches the tree's (an unforgiven parenless call, cosmetic) or differs (a parser misgrouping).
- [ ] RED: add each reduced snippet to TestCanonAddedParensOnOperatorsMatchPerl, asserting perl's grouping and quoting the Deparse output in a comment. Run it and watch the cases fail.
- [ ] GREEN: fix one cause at a time. A misgrouping is fixed in the parser. A parenless call is forgiven only through the existing parenless-call rule, widened no further than the measurement supports. Run the test after each fix.
- [ ] Run TestCanonCatchesMisgrouping and TestCanonQuotientIsNarrow after every fix.
- [ ] Remove the fixed rows from faithful_defects.txt. Give every kept row its line and the Deparse result that explains it, and re-baseline perlgittfaithful.ratchet (and the T1 ratchets if they move) in the same commit, naming each moved file.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus. Re-checking only the named case is not enough.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] Each reduced case groups as perl's Deparse,-p does, or is pinned with that measurement (`go test ./internal/parse/ -run '^TestCanonAddedParensOnOperatorsMatchPerl$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParensOnOperatorsMatchPerl'`)
- [ ] canon still catches misgrouping (`go test ./internal/parse/ -run '^TestCanonCatchesMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonCatchesMisgrouping'`)
- [ ] The defects list and the M3-dir check agree after the fixes (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulOnM3Dirs$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulOnM3Dirs'`)
- [ ] The perl.git t/ faithful ratchet records the result (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchet'`)

### Negative Scenarios

- [ ] A named unary op's parens are never forgiven when they would regroup its operand. perl 5.42 Deparses `defined $x || $y` as `(defined($x) || $y)`, so a canon `defined($x || $y)` is a misgrouping and stays unfaithful (`go test ./internal/parse/ -run '^TestCanonAddedParenOnNamedUnaryRejectsRegrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenOnNamedUnaryRejectsRegrouping'`)
- [ ] `not` keeps its low precedence. perl Deparses `my $v = not $x or 1` as `((my $v = (!$x)) or '???')`, so a canon that writes `not($x or 1)` is a misgrouping and stays unfaithful (`go test ./internal/parse/ -run '^TestCanonAddedParenOnNotKeepsLowPrecedence$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenOnNotKeepsLowPrecedence'`)
- [ ] Unary minus binds looser than `**`. perl folds `-2 ** 2` to `-4.0`, so a canon `(-2) ** 2` is a misgrouping and stays unfaithful (`go test ./internal/parse/ -run '^TestCanonAddedParenOnNegateBindsLooserThanPow$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenOnNegateBindsLooserThanPow'`)
- [ ] A paren canon adds after `\` is never forgiven. perl compiles `\(@a)` to `refgen` over a list and `\@a` to `srefgen`, so a canon `\(@a)` against source `\@a` is a different tree (`go test ./internal/parse/ -run '^TestCanonAddedParenAfterBackslashIsNotForgiven$' -count=1 -v | grep -q '^--- PASS: TestCanonAddedParenAfterBackslashIsNotForgiven'`)
- [ ] The parenless-call rule is not widened to cover these rows; a case perl's Deparse,-p groups differently from the tree stays a difference, whichever rule is asked to forgive it (`go test ./internal/parse/ -run '^TestCanonQuotientIsNarrow$' -count=1 -v | grep -q '^--- PASS: TestCanonQuotientIsNarrow'`)
- [ ] Every row kept in faithful_defects.txt for this class carries its Deparse,-p result in the reason, so no `canon-adds-paren` row is parked without a measurement (`go test ./internal/parse/ -run '^TestCanonAddsParenRowsQuoteDeparse$' -count=1 -v | grep -q '^--- PASS: TestCanonAddsParenRowsQuoteDeparse'`)
