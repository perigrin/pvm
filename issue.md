---
title: Write a Paren node's parens once, and forgive only extra nesting
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5ebe-7169-b857-996ae2953394
- 01a168d5-d519-79b1-ae4f-e988e96cf893
- 01a1690b-4745-7ef7-9285-0a9f97d86924
blocks:
- 01a16602-5f1a-7d56-a91d-5fa2c478d608
- 01a16602-5f77-78ef-a0c7-538e0a241cdc
created: 2026-10-23T01:08:34.900061622Z
updated: 2026-10-23T14:34:37.290512873Z
---

## Context

- paths: `internal/parse/canon.go`, `internal/parse/fidelity.go`, `internal/parse/fidelity_test.go`, `internal/parse/canon_test.go`, `internal/parse/testdata/t1fidelity.ratchet`, `internal/parse/testdata/t1canon.ratchet`, `internal/parse/testdata/perlgittfaithful.ratchet`, `internal/parse/testdata/m3probes/canon-paren-once.txt` (new), `docs/specs/perl-parser/04-expressions.md`
- commands: `go test ./internal/parse/ -run '^(TestCanonWritesParenOnce|TestFaithfulForgivesOnlyExtraNesting|TestFaithfulForgivenessIsClosed|TestCanonCatchesMisgrouping|TestCanonQuotientIsNarrow)$' -count=1 -v`
- entrypoints: `internal/parse/canon.go` (~344, the `List`/`Paren` case), `internal/parse/fidelity.go` (`faithfulTokens`, `matchParen`)

Once the context-changing dropped parens have been fixed as tree defects, the remaining dropped-paren failures are parens perl records but canon does not write. Measured at 054e75e6 (M3, round 8): canon writes a source paren on two paths, the `List` case (its own `(`/`)`) and a `Term` with `Paren` written as its source text (`writesParens`, canon.go:~709, `n.Kind == Term`), which reproduces nesting (`my $z = (($x))` re-emits `(($x))`); every other `Paren` node gets no pair at all (`ok(($y eq "x"))` re-emits `ok($y eq "x")`). `Paren` is a bool. Paren-once changes both paths: the Term path writes one pair instead of its source text's nesting, and a non-List, non-Term `Paren` node (CmpChain, Binary, a Call such as `(shift)->foo`) gets its one pair.

**perigrin's ruling (2026-10-22, revised after measurement).** Where a paren does not change meaning, perl builds the same ops but keeps the paren as one boolean bit, OPf_PARENS (`P`), on the op. `my $z = ($x)` is `padsv sP` vs `padsv s`, and `(($y eq "x"))` and `($y eq "x")` are identical trees. So **canon writes a `Paren` node's parens ONCE**, and **`Faithful` forgives extra nesting**: a source paren pair nested directly inside another pair around the same node (`((x))` vs `(x)`). This supersedes the earlier same-day ruling that `Faithful` forgive redundant parens.

**perigrin's ruling (2026-10-23, chain-review round 7): `Faithful` forgives exactly TWO differences, and the set is closed:** a block's last `;` (`last-semicolon`) and extra paren nesting (`extra-nesting`). Canon is precedence-aware (the named-unary, `not`, and list-op issues that block this one), and the call-paren rule is removed by the rule-removal issue, so at this issue's start Faithful forgives `last-semicolon` alone. This supersedes the round-5 "exactly three" ruling and the round-6 narrowing of the call-paren rule (A28). A wrong tree's grouping now shows up as a paren the source lacks, by construction, with no forgiveness rule to leak through (A28, A29). `faithfulTokens` and TestFaithfulForgivenessIsClosed come from the comparator issue; the `ParenNode` bit on `CanonToken` marks each pair canon wrote from a node's `Paren` flag, and the nesting rule forgives a source pair nested directly inside such a pair and nothing else. Heredoc form, sub-less special subs and block-dereference braces get no forgiveness: the tree records them and canon writes them, as it does `Paren`. A new rule needs perigrin's ruling with its perl measurement, recorded in the test's comment; no issue adds one.

**Ratchet moves (L2, chain-review round 7).** Paren-once makes files faithful, never unfaithful, because it writes parens the tree records from the source. The only way a file can drop here is a `Paren` psc set where the source had none: a tree defect, which the OPf_PARENS check classes as invented. So a file that goes from faithful to unfaithful blocks the commit; it is fixed or escalated, never re-baselined downward. The milestone's "the T1, canon, fidelity ratchets do not regress" holds as written. Gains are re-baselined in the same commit with every moved file explained.

## Steps

- [ ] Measure with the pinned perl 5.44.0 first. Run `perl -MO=Concise` on `my $z = ($x)` (with `my $x`; for a package `$x`, the `P` sits on a nulled `ex-rv2sv sKP`), on `ok(($y eq "x"))` vs `ok((($y eq "x")))`, and on `print ((1))+2` vs `print (1)+2`. Run `perl -MO=Deparse,-p` on each extent negative's source below. Quote the output in the tests' comments.
- [ ] RED: extend TestFaithfulForgivenessIsClosed. Add the second rule, `extra-nesting` (`ok((($y eq "x")))` against `ok(($y eq "x"))`, the same `seq sKPM` at 5.44), so the enumeration is exactly `last-semicolon` and `extra-nesting`. Add the extent negatives, each a canon from a hand-built wrong tree judged by the pinned perl 5.44.0 and quoted, and each unfaithful against its source with no forgiveness:
  - `defined($x||$y)` for `defined $x || $y`, where Deparse,-p gives `(defined($x) or $y)` (A28);
  - `not($x or 1)` for `not $x or 1` (A28);
  - `my($x)=@_` for `my $x = @_` (padsv_store, not aassign) (A28);
  - `local($x)=1` for `local $x = 1` (sassign, not aassign) (A28);
  - `$y x(3)` for `$y x 3` (A28);
  - `push(@a, 1 or die)` for `push @a, 1 or die`, where Deparse,-p gives `(push(@a, 1) or die)` (A29);
  - `foo(1, bar(2), 3)` for `foo 1, bar 2, 3` with both subs declared, where Deparse,-p gives `foo(1, bar(2, 3))` (A29).
  
  The test's comment says that a user sub with a `($)` prototype the file cannot see is M4 territory and is not covered. Run it and watch the extent negatives and the `extra-nesting` fixture fail.
- [ ] RED: write TestCanonWritesParenOnce in canon_test.go. A `Paren` node on a non-`List` (`(shift)->foo`, `my $z = ($x)`, `ok(($y eq "x"))`) re-emits with exactly one pair, and `((x))` re-emits as `(x)` on the Term path too (`my $z = (($x))` becomes `my $z = ($x)`). Extend d519's TestCanonMarksParenNodePairs with `ok(($y eq "x"))`: the inner pair marked ParenNode, the call's pair not (M3). Run both and watch them fail.
- [ ] GREEN: make canon write a `Paren` node's parens once, marking the pair `ParenNode`. Run it and watch it pass.
- [ ] RED: write TestFaithfulForgivesOnlyExtraNesting in fidelity_test.go. `((x))` in the source against canon's `(x)` is faithful, and so is `(((x)))`. A source pair canon does not write at all (no `Paren` on the node) is still a difference. Run it and watch it fail.
- [ ] GREEN: implement `extra-nesting` in `faithfulTokens`, limited to a source pair nested directly inside a `ParenNode` pair around the same node. Run it, TestFaithfulForgivenessIsClosed, TestCanonCatchesMisgrouping and TestCanonQuotientIsNarrow, and watch them all pass.
- [ ] Add the measured cases as probe file `internal/parse/testdata/m3probes/canon-paren-once.txt`.
- [ ] Re-baseline t1canon.ratchet, t1fidelity.ratchet and perlgittfaithful.ratchet in the same commit. The message names each moved file and why it moved. A file that goes from faithful to unfaithful blocks the commit: it is an invented `Paren` (a tree defect) to fix or escalate, never a reason to re-baseline downward.
- [ ] Update docs/specs/perl-parser/04-expressions.md §4.10. Remove the stale "Measured 2026-09-17 at 3b9422eb ... `($x);` parses ... with no `Paren` node" claim, which §4.14.3's `Paren` row contradicts, and state the ruling.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus, and adds them to this issue's probe file under internal/parse/testdata/m3probes/ so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] canon writes a Paren node's parens exactly once (`go test ./internal/parse/ -run '^TestCanonWritesParenOnce$' -count=1 -v | grep -q '^--- PASS: TestCanonWritesParenOnce'`)
- [ ] Faithful forgives only extra nesting around the same node (`go test ./internal/parse/ -run '^TestFaithfulForgivesOnlyExtraNesting$' -count=1 -v | grep -q '^--- PASS: TestFaithfulForgivesOnlyExtraNesting'`)
- [ ] The perl.git t/ faithful ratchet matches the tree (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchet'`)
- [ ] The perl.git t/ faithful count is above the start recorded in m3start.txt (`sh -c 's=$(sed -n "s/^faithful_start_all = \([0-9][0-9]*\)$/\1/p" internal/parseoracle/testdata/ratchet/m3start.txt); n=$(sed -n "s|^# \([0-9][0-9]*\)/619 faithful.*|\1|p" internal/parse/testdata/perlgittfaithful.ratchet); [ -n "$s" ] && [ -n "$n" ] && [ "$n" -gt "$s" ]'`)
- [ ] Regression guard (passes today and after any re-baseline): the T1 canon ratchet matches its file (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)
- [ ] Regression guard (passes today and after any re-baseline): the T1 fidelity ratchet matches its file (`go test ./internal/parse/ -run '^TestCanonFidelityRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityRatchet'`)
- [ ] Regression guard: canon still catches misgrouping (`go test ./internal/parse/ -run '^TestCanonCatchesMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonCatchesMisgrouping'`)
- [ ] Regression guard: the forgiven quotient stays narrow (`go test ./internal/parse/ -run '^TestCanonQuotientIsNarrow$' -count=1 -v | grep -q '^--- PASS: TestCanonQuotientIsNarrow'`)
- [ ] The measured cases are committed as perl-judged probes, and the probe test holds them (`sh -c 'test -s internal/parse/testdata/m3probes/canon-paren-once.txt && go test ./internal/parse/ -run "^TestM3ProbesHeldToPerl$" -count=1 -v | grep -q "^--- PASS: TestM3ProbesHeldToPerl"'`)
- [ ] Faithful forgives exactly the two ruled differences, each with a perl-measured fixture, and the seven extent negatives are unfaithful (`go test ./internal/parse/ -run '^TestFaithfulForgivenessIsClosed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulForgivenessIsClosed'`)

### Negative Scenarios

- [ ] Source `my ($x) = @_` against canon `my $x = @_` is unfaithful. perl 5.42 builds `padrange` + `aassign` for the first and `padsv_store` for the second, so this paren is never forgiven (`go test ./internal/parse/ -run '^TestFaithfulRejectsDroppedParenAfterMy$' -count=1 -v | grep -q '^--- PASS: TestFaithfulRejectsDroppedParenAfterMy'`)
- [ ] Source `($y) x 3` against canon `$y x 3` is unfaithful. perl measures `repeat vK/DOLIST,2` for the first and `repeat vK/2` for the second (`go test ./internal/parse/ -run '^TestFaithfulRejectsDroppedParenBeforeRepeat$' -count=1 -v | grep -q '^--- PASS: TestFaithfulRejectsDroppedParenBeforeRepeat'`)
- [ ] Source `local ($x) = 1` against canon `local $x = 1` (for `our $x`) is unfaithful. perl measures `aassign` for the first and `sassign` for the second (`go test ./internal/parse/ -run '^TestFaithfulRejectsDroppedParenAfterLocal$' -count=1 -v | grep -q '^--- PASS: TestFaithfulRejectsDroppedParenAfterLocal'`)
- [ ] The nesting rule does not forgive a call's argument parens as an outer pair, because they are not a ParenNode pair. perl measures `print ((1))+2` as `const[IV 1] sP` and `print (1)+2` as `const[IV 1] s`, so source `print((1))+2` against canon `print(1)+2` is unfaithful (`go test ./internal/parse/ -run '^TestFaithfulDoesNotTreatCallParensAsNesting$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDoesNotTreatCallParensAsNesting'`)
- [ ] A pair nested inside another pair but around a different node is not forgiven: source `((1) + 2)` against canon `(1 + 2)` is unfaithful (`go test ./internal/parse/ -run '^TestFaithfulDoesNotForgiveInnerPairAroundSubterm$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDoesNotForgiveInnerPairAroundSubterm'`)
- [ ] Doubled parens in a context-changing position are forgiven only down to one pair, never to none. perl measures `(($y)) x 3` and `($y) x 3` identically, both as `repeat vK/DOLIST,2`, but source `(($y)) x 3` against canon `$y x 3` is unfaithful (`go test ./internal/parse/ -run '^TestFaithfulNestingForgivesToOnePairNotZero$' -count=1 -v | grep -q '^--- PASS: TestFaithfulNestingForgivesToOnePairNotZero'`)
- [ ] canon keeps a call-argument pair around a grouped term: `print ((1+2)*3)` re-emits with both pairs, never as `print (1+2)*3` (`go test ./internal/parse/ -run '^TestCanonKeepsCallArgParenAroundGroupedTerm$' -count=1 -v | grep -q '^--- PASS: TestCanonKeepsCallArgParenAroundGroupedTerm'`)
- [ ] `my ($p, $q) = (1, 2)` and `my $p, $q` are different trees; canon re-emits each as written, and Faithful rejects either against the other (`go test ./internal/parse/ -run '^TestFaithfulDistinguishesMyListFromMyComma$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDistinguishesMyListFromMyComma'`)
