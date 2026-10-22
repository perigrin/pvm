---
title: Classify dropped source parens by position, then write a Paren node's parens once
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5e5e-73cc-a210-0341ea517914
blocks:
- 01a16602-5f1a-7d56-a91d-5fa2c478d608
- 01a16602-5f77-78ef-a0c7-538e0a241cdc
created: 2026-10-22T23:46:33.790255282Z
updated: 2026-10-22T23:58:24.353971229Z
---

## Context

- paths: `internal/parse/canon.go`, `internal/parse/fidelity.go`, `internal/parse/fidelity_test.go`, `internal/parse/canon_test.go`, the parser files a tree-defect fix touches (e.g. `internal/parse/term.go`), `internal/parse/testdata/t1fidelity.ratchet`, `internal/parse/testdata/t1canon.ratchet`, `internal/parse/testdata/perlgittfaithful.ratchet`
- commands: `go test ./internal/parse/ -run '^(TestDroppedParenPositionsAreClassified|TestCanonWritesParenOnce|TestFaithfulForgivesOnlyExtraNesting|TestCanonCatchesMisgrouping|TestCanonQuotientIsNarrow)$' -count=1 -v`
- entrypoints: `internal/parse/canon.go` (~344, the `List`/`Paren` case), `internal/parse/fidelity.go` (`Faithful`, `matchParen`), `internal/parse/term.go` (`finishList`, ~723)

117 of the 169 M3-dir re-emission failures are `source has "(", canon has <something else>`: a source paren canon does not write. Examples are `(shift)->SUPER::foo` (mro/basic_05_c3.t:32), `ok(($y eq '...'), ...)` (mro/c3_with_overload.t:40) and `if (($Config{...} !~ /.../) )` (lib/commonsense.t:12). canon writes a source paren only for a `List`, and `Paren` is a bool.

**perigrin's ruling (2026-10-22, revised after measurement).** Measured on perl 5.42.0 with B::Concise. Where parens change meaning, perl builds different ops: `my ($x) = @_` is aassign but `my $x = @_` is padsv_store, and `($x) x 3` is a list repeat (`lK/DOLIST`) but `$x x 3` is a string repeat. Where they do not change meaning, perl keeps the paren as one boolean bit, OPf_PARENS (`P`), on the op: `my $z = ($x)` is `padsv sP` vs `padsv s`. `(($y eq "x"))` and `($y eq "x")` are identical trees. So **canon writes a `Paren` node's parens ONCE**, and **`Faithful` forgives ONLY extra nesting**: a source paren pair nested directly inside another around the same node (`((x))` vs `(x)`). Before anything is forgiven, the 117 failures are classified by position. A failure in a context-changing position (before `=`, before `x`, after `my`/`our`/`local`/`state`) is a tree defect, where psc read a list as a scalar, and it is fixed, not forgiven. This supersedes the earlier same-day ruling that `Faithful` forgive redundant parens.

## Steps

- [ ] Measure with perl first. Run `perl -MO=Concise` on `my ($x) = @_` vs `my $x = @_`, `($x) x 3` vs `$x x 3`, `($x) = f()`, `() = f()`, `local ($x) = 1`, `my $z = ($x)` and `ok((($y eq "x")))`. Quote the output in the tests' comments.
- [ ] Classify: list every one of the 117 dropped-paren failures with its position class (context-changing: before `=`, before `x`, after my/our/local/state; or not). Put the list in the commit message.
- [ ] RED: write TestDroppedParenPositionsAreClassified. For each context-changing failure, reduce it to a fixture asserting psc's tree matches perl's ops (a list where perl builds a list). Run it and watch any tree defect fail.
- [ ] GREEN: fix each tree defect in the parser, one at a time, running the test after each. If there are none, the test still pins the measured positions as lists.
- [ ] RED: write TestCanonWritesParenOnce in canon_test.go. A `Paren` node on a non-`List` (`(shift)->foo`, `my $z = ($x)`, `ok(($y eq "x"))`) re-emits with exactly one pair, and `((x))` re-emits as `(x)`. Run it and watch it fail.
- [ ] GREEN: make canon write a `Paren` node's parens once. Run it and watch it pass.
- [ ] RED: write TestFaithfulForgivesOnlyExtraNesting in fidelity_test.go. `((x))` in the source against canon's `(x)` is faithful, and so is `(((x)))`. A source pair canon does not write at all (no `Paren` on the node) is still a difference. Run it and watch it fail.
- [ ] GREEN: implement the forgiveness in `Faithful`, limited to a pair nested directly inside another pair around the same node. Run it, TestCanonCatchesMisgrouping and TestCanonQuotientIsNarrow, and watch them all pass.
- [ ] Re-baseline t1canon.ratchet, t1fidelity.ratchet and perlgittfaithful.ratchet in the same commit. The message names each moved file and why it moved; a file going from faithful to unfaithful must be explained by a tree-defect fix or it blocks the commit.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the Concise measurements and writes NEW adversarial probes outside the corpus (e.g. `my ($a, $b) = ...` vs `my $a, $b`, `(@a) x 2`, `local ($_) = ...`, `return (1)`, `print ((1+2)*3)`, a paren around a ternary arm). Re-checking only the named case is not enough.
- [ ] Run `make test`, then commit.
- [ ] Update docs/specs/perl-parser/04-expressions.md §4.10: remove the stale "Measured 2026-09-17 at 3b9422eb ... `($x);` parses ... with no `Paren` node" claim (contradicted by §4.14.3's `Paren` row) and state perigrin's ruling: canon writes a `Paren` node's parens once, as perl keeps one OPf_PARENS bit per op, and `Faithful` forgives only extra nesting

## Acceptance criteria

### Positive Scenarios

- [ ] Every context-changing dropped-paren position parses as perl's ops do (`go test ./internal/parse/ -run '^TestDroppedParenPositionsAreClassified$' -count=1 -v | grep -q '^--- PASS: TestDroppedParenPositionsAreClassified'`)
- [ ] canon writes a Paren node's parens exactly once (`go test ./internal/parse/ -run '^TestCanonWritesParenOnce$' -count=1 -v | grep -q '^--- PASS: TestCanonWritesParenOnce'`)
- [ ] Faithful forgives only extra nesting around the same node (`go test ./internal/parse/ -run '^TestFaithfulForgivesOnlyExtraNesting$' -count=1 -v | grep -q '^--- PASS: TestFaithfulForgivesOnlyExtraNesting'`)
- [ ] canon still catches misgrouping (`go test ./internal/parse/ -run '^TestCanonCatchesMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonCatchesMisgrouping'`)
- [ ] The forgiven quotient stays narrow (`go test ./internal/parse/ -run '^TestCanonQuotientIsNarrow$' -count=1 -v | grep -q '^--- PASS: TestCanonQuotientIsNarrow'`)
- [ ] The T1 canon ratchet matches its re-baselined file (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)
- [ ] The T1 fidelity ratchet matches its re-baselined file (`go test ./internal/parse/ -run '^TestCanonFidelityRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityRatchet'`)
- [ ] The perl.git t/ faithful ratchet records the gain (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchet'`)

### Negative Scenarios

- [ ] Source `my ($x) = @_` against canon `my $x = @_` is unfaithful. perl 5.42 builds `padrange` + `aassign` for the first and `padsv_store` for the second, so this paren is never forgiven (`go test ./internal/parse/ -run '^TestFaithfulRejectsDroppedParenAfterMy$' -count=1 -v | grep -q '^--- PASS: TestFaithfulRejectsDroppedParenAfterMy'`)
- [ ] Source `($y) x 3` against canon `$y x 3` is unfaithful. perl measures `repeat vK/DOLIST,2` for the first and `repeat vK/2` for the second (`go test ./internal/parse/ -run '^TestFaithfulRejectsDroppedParenBeforeRepeat$' -count=1 -v | grep -q '^--- PASS: TestFaithfulRejectsDroppedParenBeforeRepeat'`)
- [ ] Source `local ($x) = 1` against canon `local $x = 1` (for `our $x`) is unfaithful. perl measures `aassign` for the first and `sassign` for the second (`go test ./internal/parse/ -run '^TestFaithfulRejectsDroppedParenAfterLocal$' -count=1 -v | grep -q '^--- PASS: TestFaithfulRejectsDroppedParenAfterLocal'`)
- [ ] A paren after `\` changes context, and the classification treats it so: `\(@a)` is `refgen` over `padav lPM`, while `\@a` is `srefgen`. A dropped paren there is a tree defect, not a forgivable one (`go test ./internal/parse/ -run '^TestDroppedParenAfterBackslashIsContextChanging$' -count=1 -v | grep -q '^--- PASS: TestDroppedParenAfterBackslashIsContextChanging'`)
- [ ] The nesting rule does not forgive a call's argument parens as an outer pair. perl measures `print ((1))+2` as `const[IV 1] sP` and `print (1)+2` as `const[IV 1] s`, so source `print((1))+2` against canon `print(1)+2` is unfaithful (`go test ./internal/parse/ -run '^TestFaithfulDoesNotTreatCallParensAsNesting$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDoesNotTreatCallParensAsNesting'`)
- [ ] A pair nested inside another pair but around a different node is not forgiven: source `((1) + 2)` against canon `(1 + 2)` is unfaithful (`go test ./internal/parse/ -run '^TestFaithfulDoesNotForgiveInnerPairAroundSubterm$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDoesNotForgiveInnerPairAroundSubterm'`)
- [ ] Doubled parens in a context-changing position are forgiven only down to one pair, never to none. perl measures `(($y)) x 3` and `($y) x 3` identically, both as `repeat vK/DOLIST,2`, but source `(($y)) x 3` against canon `$y x 3` is unfaithful (`go test ./internal/parse/ -run '^TestFaithfulNestingForgivesToOnePairNotZero$' -count=1 -v | grep -q '^--- PASS: TestFaithfulNestingForgivesToOnePairNotZero'`)
- [ ] canon keeps a call-argument pair around a grouped term: `print ((1+2)*3)` re-emits with both pairs, never as `print (1+2)*3` (`go test ./internal/parse/ -run '^TestCanonKeepsCallArgParenAroundGroupedTerm$' -count=1 -v | grep -q '^--- PASS: TestCanonKeepsCallArgParenAroundGroupedTerm'`)
- [ ] The classification fixtures cannot pass empty: they hold at least one case from each context-changing position (before `=`, before `x`, after each of my/our/local/state, after `\`), and the test fails if any class has none (`go test ./internal/parse/ -run '^TestDroppedParenFixturesCoverEveryContextPosition$' -count=1 -v | grep -q '^--- PASS: TestDroppedParenFixturesCoverEveryContextPosition'`)
- [ ] `my ($p, $q) = (1, 2)` and `my $p, $q` are different trees; canon re-emits each as written, and Faithful rejects either against the other (`go test ./internal/parse/ -run '^TestFaithfulDistinguishesMyListFromMyComma$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDistinguishesMyListFromMyComma'`)
