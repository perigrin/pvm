---
title: Remove the call-paren forgiveness and re-home the block argument's last semicolon
state: pending
urgency: normal
milestone: m3-oracle-at-scale
created: 2026-10-23T14:34:59.350295778Z
updated: 2026-10-23T14:34:59.350295778Z
---

## Context

- paths: `internal/parse/fidelity.go` (the `isName` call-paren rule and `matchForgivingSemicolons`, removed; `Faithful`'s doc comment), `internal/parse/fidelity_test.go` (TestFaithfulForgivenessIsClosed's enumeration; TestCanonQuotientIsNarrow's comment "only the parenless-call paren is forgiven"), `internal/parse/listop_test.go`, `internal/parse/namedunary_test.go`, `internal/parse/notop_test.go` (the verdict halves of the wrong-tree tests), `internal/parse/testdata/t1fidelity.ratchet`, `internal/parse/testdata/perlgittfaithful.ratchet`, `internal/parseoracle/testdata/ratchet/m3start.txt` (read)
- commands: `go test ./internal/parse/ -run '^(TestFaithfulCallParenRuleIsGone|TestFaithfulForgivenessIsClosed|TestFaithfulLastSemicolonCoversBlockArgument|TestCanonWrongTreesAreUnfaithful)$' -count=1 -v`
- entrypoints: `internal/parse/fidelity.go` (`faithfulTokens`, `isName`, `matchForgivingSemicolons`)

**perigrin's ruling (2026-10-23, chain-review round 7): canon is precedence-aware, and the call-paren forgiveness rule is removed.** This is the last of the three issues that follow the named-unary one (P29 split, round 8). The measurement issue showed the design's counts and the list-op issue made canon write no call paren the source lacked, so the rule has nothing left to forgive. This issue removes it, which is where any remaining loss shows; it is also where the wrong-tree tests of the three canon issues gain their verdict assertions (M1), because while the rule existed those canons were still forgiven.

**What goes.** `faithfulTokens` loses the `isName` branch and `matchForgivingSemicolons`. The last-`;` rule in the main loop must then cover a block argument on its own (`print {$fh} "x"` is emitted `print {$fh;} "x"`), which the inner match used to handle one level down. TestFaithfulForgivenessIsClosed's enumeration becomes `last-semicolon` alone at this commit; the Paren-once issue adds `extra-nesting`. The code comments that describe the old rule are updated with it (M8, fresh7): `Faithful`'s doc comment in fidelity.go and TestCanonQuotientIsNarrow's comment in fidelity_test.go.

**Stop condition (M6).** The measurement issue predicted the counts with the rule off, per file and per statement. If removing the rule drops any FILE or any STATEMENT from faithful to unfaithful that the measurement did not predict and the list-op issue did not recover, the executor stops and brings perigrin the numbers and the first differences. A drop is never re-baselined.

**Why the rule removal lives here and not in the Paren-once issue.** The measurement's "no file lost" is a property of canon with the rule OFF. The rule can only come out once every Call canon parenthesises without a source paren is gone, and that is the list-op issue's last step. The Paren-once issue then starts from a Faithful with one rule and a canon that writes no paren the tree does not imply.

## Steps

- [ ] RED: write TestFaithfulCallParenRuleIsGone in fidelity_test.go: through `parse.FaithfulTokens`, source `print "x"` against canon tokens `print ( "x" )` is unfaithful, and so is `defined $x` against `defined ( $x )`. Run it and watch it fail.
- [ ] RED: write TestCanonWrongTreesAreUnfaithful (the verdict halves moved here under M1): each hand-built wrong tree of the three canon issues, run through `Canon` and judged with `Faithful` against its source, is unfaithful: `defined` over `||` (`defined ($x || $y)`), `ref` over `eq`, `defined` over `($x) . "a"`, `not` over `or` (`not ($x or 1)`), `push` over `or` (`push (@a, 1 or die)`), `foo(1, bar(2), 3)` (`foo 1, bar (2), 3`), and `print` over `($a), $b` (`print(($a), $b)`). Run it and watch every case fail while the rule forgives them.
- [ ] RED: write TestFaithfulLastSemicolonCoversBlockArgument: `print {$fh} "x"` emitted `print {$fh;} "x"` is faithful, and `print {$fh} "x"` against `print {$fh;} ("x")` is not. Run it; the first half fails once the inner match is gone, so write it before the removal and watch which half fails.
- [ ] GREEN: remove the `isName` branch and `matchForgivingSemicolons` from fidelity.go, make the last-`;` rule cover a block argument in the main loop, drop `call-parens` from TestFaithfulForgivenessIsClosed's enumeration (its fixture moves to TestFaithfulCallParenRuleIsGone as the negative), and update the two comments. Run the four tests, TestFaithfulForgivenessIsClosed, TestCanonCatchesMisgrouping and TestCanonQuotientIsNarrow, and watch them pass.
- [ ] Run TestCanonRatchet, TestCanonFidelityRatchet, TestParseRatchet, TestPerlGitTRatchet and TestPerlGitTFaithfulRatchet. Compare with the measurement issue's prediction: a file or statement that went from faithful to unfaithful is the stop condition, never a re-baseline. A file that became faithful is a gain: re-baseline in the same commit and explain each moved file.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus (a block argument with two statements, `print {$fh;} "x"` in the source, a nested block argument `map { grep { 1 } @a } @b`), and adds them to the list-op issue's probe file so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The forgiven set is exactly the rules in force, each with a perl-measured fixture, and call-parens is not among them (`go test ./internal/parse/ -run '^TestFaithfulForgivenessIsClosed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulForgivenessIsClosed'`)
- [ ] No file was lost: the perl.git t/ faithful count is at or above faithful_start_all in m3start.txt (`sh -c 's=$(sed -n "s/^faithful_start_all = \([0-9][0-9]*\)$/\1/p" internal/parseoracle/testdata/ratchet/m3start.txt); n=$(sed -n "s|^# \([0-9][0-9]*\)/619 faithful.*|\1|p" internal/parse/testdata/perlgittfaithful.ratchet); [ -n "$s" ] && [ -n "$n" ] && [ "$n" -ge "$s" ]'`)
- [ ] No M3-dir file was lost: the M3-dir faithful count is at or above faithful_start_m3 (`sh -c 's=$(sed -n "s/^faithful_start_m3 = \([0-9][0-9]*\)$/\1/p" internal/parseoracle/testdata/ratchet/m3start.txt); n=$(awk "\$1==\"0\" && \$2 ~ /^(op|mro|uni|lib|test_pl)\// {k++} END {print k+0}" internal/parse/testdata/perlgittfaithful.ratchet); [ -n "$s" ] && [ "$n" -ge "$s" ]'`)
- [ ] Regression guard (passes today and after any explained re-baseline): the T1 fidelity ratchet (`go test ./internal/parse/ -run '^TestCanonFidelityRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityRatchet'`)
- [ ] Regression guard (passes today and after any explained re-baseline): the perl.git t/ faithful ratchet (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchet'`)
- [ ] Regression guard: canon still catches misgrouping (`go test ./internal/parse/ -run '^TestCanonCatchesMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonCatchesMisgrouping'`)
- [ ] Regression guard: the forgiven quotient stays narrow (`go test ./internal/parse/ -run '^TestCanonQuotientIsNarrow$' -count=1 -v | grep -q '^--- PASS: TestCanonQuotientIsNarrow'`)
- [ ] The stale rule wording is gone from the code comments: fidelity_test.go:57's "only the parenless-call paren is forgiven" and fidelity.go's ABOUTME "the parens a parenless call must acquire" (`sh -c '! grep -q "only the parenless-call paren is forgiven" internal/parse/fidelity_test.go && ! grep -q "the parens a parenless call must acquire" internal/parse/fidelity.go'`)

### Negative Scenarios

- [ ] The call-paren rule is gone: source `print "x"` against canon `print("x")`, and `defined $x` against `defined($x)`, are unfaithful through the comparator (`go test ./internal/parse/ -run '^TestFaithfulCallParenRuleIsGone$' -count=1 -v | grep -q '^--- PASS: TestFaithfulCallParenRuleIsGone'`)
- [ ] Every hand-built wrong tree of the three canon issues is unfaithful against its source with no forgiveness, the leading-`(` cases included (`go test ./internal/parse/ -run '^TestCanonWrongTreesAreUnfaithful$' -count=1 -v | grep -q '^--- PASS: TestCanonWrongTreesAreUnfaithful'`)
- [ ] The last-`;` rule still covers a block argument without the removed inner match: `print {$fh} "x"` emitted `print {$fh;} "x"` is faithful, and `print {$fh} "x"` against `print {$fh;} ("x")` is not (`go test ./internal/parse/ -run '^TestFaithfulLastSemicolonCoversBlockArgument$' -count=1 -v | grep -q '^--- PASS: TestFaithfulLastSemicolonCoversBlockArgument'`)
