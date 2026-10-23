---
title: Write a Paren node's parens once, and forgive only extra nesting
state: pending
urgency: normal
milestone: m3-oracle-at-scale
blocked_by:
- 01a16602-5ebe-7169-b857-996ae2953394
blocks:
- 01a16602-5f1a-7d56-a91d-5fa2c478d608
- 01a16602-5f77-78ef-a0c7-538e0a241cdc
created: 2026-10-23T01:08:34.900061622Z
updated: 2026-10-23T08:20:59.159446193Z
---

## Context

- paths: `internal/parse/canon.go`, `internal/parse/fidelity.go`, `internal/parse/fidelity_test.go`, `internal/parse/canon_test.go`, `internal/parse/testdata/t1fidelity.ratchet`, `internal/parse/testdata/t1canon.ratchet`, `internal/parse/testdata/perlgittfaithful.ratchet`, `internal/parse/testdata/m3probes/canon-paren-once.txt` (new), `docs/specs/perl-parser/04-expressions.md`
- commands: `go test ./internal/parse/ -run '^(TestCanonWritesParenOnce|TestFaithfulForgivesOnlyExtraNesting|TestFaithfulTokens|TestCanonCatchesMisgrouping|TestCanonQuotientIsNarrow)$' -count=1 -v`
- entrypoints: `internal/parse/canon.go` (~344, the `List`/`Paren` case), `internal/parse/fidelity.go` (`Faithful`, `matchParen`)

Once the context-changing dropped parens have been fixed as tree defects, the remaining dropped-paren failures are parens perl records but canon does not write. canon writes a source paren only for a `List`, and `Paren` is a bool.

**perigrin's ruling (2026-10-22, revised after measurement).** Where a paren does not change meaning, perl builds the same ops but keeps the paren as one boolean bit, OPf_PARENS (`P`), on the op. `my $z = ($x)` is `padsv sP` vs `padsv s`, and `(($y eq "x"))` and `($y eq "x")` are identical trees. So **canon writes a `Paren` node's parens ONCE**, and **`Faithful` forgives ONLY extra nesting**: a source paren pair nested directly inside another pair around the same node (`((x))` vs `(x)`). This supersedes the earlier same-day ruling that `Faithful` forgive redundant parens.

`Faithful(n *Node, src []byte)` takes a tree, so a test that says "source A against canon B is unfaithful" needs either a hand-built tree or a comparator over token sequences. This issue splits that comparator out of `Faithful` as `faithfulTokens(canon []CanonToken, src []string) bool`, with `Faithful` calling it. The negatives here and in the triage issues that follow use it. **How it tells call-argument parens from a Paren pair (F6, chain-review round 2):** the source side cannot tell them apart from text alone, so canon tags its own tokens by role as it emits them. A `(`/`)` is `paren-node` when canon wrote it for a node with `Paren` set, `call-args` when it opened a call's argument list (including a parenless call's acquired parens), and `binding` when binding power required it. A source paren pair is forgiven as extra nesting ONLY when it sits directly inside a source pair aligned with a canon `paren-node` pair, around the same tokens. A pair aligned with `call-args` or `binding` never makes an inner pair forgivable. So source `print((1))+2` against canon `print(1)+2` stays unfaithful: the outer pair is `call-args`, and `1` has no Paren. Source `print ((1))+2` against canon `print((1))+2` is faithful as written, and `ok(((($y eq "x"))))` against canon `ok(($y eq "x"))` is forgiven down to the one `paren-node` pair.

**perigrin's ruling (2026-10-23, chain-review round 5): `Faithful` forgives exactly three differences, and the set is closed:** call parens, a block's last `;`, and extra paren nesting. `faithfulTokens` reports which rule forgave each difference (`call-parens`, `last-semicolon`, `extra-nesting`). TestFaithfulForgivenessIsClosed enumerates those three, each with a perl-measured fixture, and fails when any difference is forgiven by an unlisted rule. Heredoc form, sub-less special subs and block-dereference braces get no forgiveness: the tree records them and canon writes them, as it does `Paren`. A new rule needs perigrin's ruling with its perl measurement, recorded in the test's comment; no issue adds one.

**What fidelity.go forgives today (read 2026-10-23 at 054e75e6).** It has exactly the two rules its comment names:
1. call parens, keyed on `(` right after a name in the emission, with the arguments matched unparenthesised;
2. a block's last `;`, before the emission's `}`.

Rule 2 is also applied inside a forgiven call paren (`matchForgivingSemicolons`), for a block argument such as `print({$fh;} "x")`. That is the same rule one level down, not a fourth. Separately, `significantTokens` drops Whitespace, Comment, Pod and DataSection tokens before comparing. That is token selection, not a forgiven difference, but a skip added there would bypass the rule list. So TestFaithfulForgivenessIsClosed also pins the dropped token kinds to exactly those four.

## Steps

- [ ] Measure with the pinned perl 5.44.0 first. Run `perl -MO=Concise` on `my $z = ($x)` (with `my $x`; for a package `$x`, the `P` sits on a nulled `ex-rv2sv sKP`), on `ok(($y eq "x"))` vs `ok((($y eq "x")))`, and on `print ((1))+2` vs `print (1)+2`. Quote the output in the tests' comments.
- [ ] RED/GREEN: write TestFaithfulTokens, then split `faithfulTokens` out of `Faithful` with no change in behaviour, giving canon's emitted tokens their role tags. `faithfulTokens` returns, with its verdict, the rule that forgave each difference.
- [ ] RED/GREEN: write TestFaithfulForgivenessIsClosed. It holds one perl-measured fixture per rule (call parens: `print "x"` against `print("x")`; last `;`: `sub f { 1 }` against `sub f { 1; }`, measured `11`; extra nesting: `ok((($y eq "x")))` against `ok(($y eq "x"))`, the same `seq sKPM`), checks that each fixture is forgiven by its named rule, fails on any rule name outside those three, and pins `significantTokens`' dropped kinds to Whitespace, Comment, Pod and DataSection. TestCanonFidelityRatchet must not move.
- [ ] RED: write TestCanonWritesParenOnce in canon_test.go. A `Paren` node on a non-`List` (`(shift)->foo`, `my $z = ($x)`, `ok(($y eq "x"))`) re-emits with exactly one pair, and `((x))` re-emits as `(x)`. Run it and watch it fail.
- [ ] GREEN: make canon write a `Paren` node's parens once. Run it and watch it pass.
- [ ] RED: write TestFaithfulForgivesOnlyExtraNesting in fidelity_test.go. `((x))` in the source against canon's `(x)` is faithful, and so is `(((x)))`. A source pair canon does not write at all (no `Paren` on the node) is still a difference. Run it and watch it fail.
- [ ] GREEN: implement the forgiveness in `faithfulTokens`, limited to a pair nested directly inside another pair around the same node. Run it, TestCanonCatchesMisgrouping and TestCanonQuotientIsNarrow, and watch them all pass.
- [ ] Add the measured cases as probe file `internal/parse/testdata/m3probes/canon-paren-once.txt`.
- [ ] Re-baseline t1canon.ratchet, t1fidelity.ratchet and perlgittfaithful.ratchet in the same commit. The message names each moved file and why it moved. A file that goes from faithful to unfaithful blocks the commit.
- [ ] Update docs/specs/perl-parser/04-expressions.md §4.10. Remove the stale "Measured 2026-09-17 at 3b9422eb ... `($x);` parses ... with no `Paren` node" claim, which §4.14.3's `Paren` row contradicts, and state the ruling.
- [ ] Any change to internal/parse/sites.go, canon.go, fidelity.go or the parser gets a per-fix review against perl before merge. The reviewer re-runs the perl measurements and writes NEW adversarial probes outside the corpus, and adds them to this issue's probe file under internal/parse/testdata/m3probes/ so TestM3ProbesHeldToPerl runs them from then on. Re-checking only the named case is not enough.
- [ ] Run `make test`, then commit.

## Acceptance criteria

### Positive Scenarios

- [ ] The token-sequence comparator behaves as Faithful did before the split (`go test ./internal/parse/ -run '^TestFaithfulTokens$' -count=1 -v | grep -q '^--- PASS: TestFaithfulTokens'`)
- [ ] canon writes a Paren node's parens exactly once (`go test ./internal/parse/ -run '^TestCanonWritesParenOnce$' -count=1 -v | grep -q '^--- PASS: TestCanonWritesParenOnce'`)
- [ ] Faithful forgives only extra nesting around the same node (`go test ./internal/parse/ -run '^TestFaithfulForgivesOnlyExtraNesting$' -count=1 -v | grep -q '^--- PASS: TestFaithfulForgivesOnlyExtraNesting'`)
- [ ] The perl.git t/ faithful ratchet matches the tree (`PERL5_CORPUS=$HOME/dev/perl5 go test ./internal/parse/ -run '^TestPerlGitTFaithfulRatchet$' -count=1 -v | grep -q '^--- PASS: TestPerlGitTFaithfulRatchet'`)
- [ ] The perl.git t/ faithful count is above the start recorded in m3start.txt (`sh -c 's=$(sed -n "s/^faithful_start_all = \([0-9][0-9]*\)$/\1/p" internal/parseoracle/testdata/ratchet/m3start.txt); n=$(sed -n "s|^# \([0-9][0-9]*\)/619 faithful.*|\1|p" internal/parse/testdata/perlgittfaithful.ratchet); [ -n "$s" ] && [ -n "$n" ] && [ "$n" -gt "$s" ]'`)
- [ ] Regression guard (passes today and after any re-baseline): the T1 canon ratchet matches its file (`go test ./internal/parse/ -run '^TestCanonRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonRatchet'`)
- [ ] Regression guard (passes today and after any re-baseline): the T1 fidelity ratchet matches its file (`go test ./internal/parse/ -run '^TestCanonFidelityRatchet$' -count=1 -v | grep -q '^--- PASS: TestCanonFidelityRatchet'`)
- [ ] Regression guard: canon still catches misgrouping (`go test ./internal/parse/ -run '^TestCanonCatchesMisgrouping$' -count=1 -v | grep -q '^--- PASS: TestCanonCatchesMisgrouping'`)
- [ ] Regression guard: the forgiven quotient stays narrow (`go test ./internal/parse/ -run '^TestCanonQuotientIsNarrow$' -count=1 -v | grep -q '^--- PASS: TestCanonQuotientIsNarrow'`)
- [ ] The measured cases are committed as perl-judged probes (`test -s internal/parse/testdata/m3probes/canon-paren-once.txt`)
- [ ] Faithful forgives exactly the three ruled differences, each shown by a perl-measured fixture, and nothing else (`go test ./internal/parse/ -run '^TestFaithfulForgivenessIsClosed$' -count=1 -v | grep -q '^--- PASS: TestFaithfulForgivenessIsClosed'`)

### Negative Scenarios

- [ ] Source `my ($x) = @_` against canon `my $x = @_` is unfaithful. perl 5.42 builds `padrange` + `aassign` for the first and `padsv_store` for the second, so this paren is never forgiven (`go test ./internal/parse/ -run '^TestFaithfulRejectsDroppedParenAfterMy$' -count=1 -v | grep -q '^--- PASS: TestFaithfulRejectsDroppedParenAfterMy'`)
- [ ] Source `($y) x 3` against canon `$y x 3` is unfaithful. perl measures `repeat vK/DOLIST,2` for the first and `repeat vK/2` for the second (`go test ./internal/parse/ -run '^TestFaithfulRejectsDroppedParenBeforeRepeat$' -count=1 -v | grep -q '^--- PASS: TestFaithfulRejectsDroppedParenBeforeRepeat'`)
- [ ] Source `local ($x) = 1` against canon `local $x = 1` (for `our $x`) is unfaithful. perl measures `aassign` for the first and `sassign` for the second (`go test ./internal/parse/ -run '^TestFaithfulRejectsDroppedParenAfterLocal$' -count=1 -v | grep -q '^--- PASS: TestFaithfulRejectsDroppedParenAfterLocal'`)
- [ ] The nesting rule does not forgive a call's argument parens as an outer pair. perl measures `print ((1))+2` as `const[IV 1] sP` and `print (1)+2` as `const[IV 1] s`, so source `print((1))+2` against canon `print(1)+2` is unfaithful (`go test ./internal/parse/ -run '^TestFaithfulDoesNotTreatCallParensAsNesting$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDoesNotTreatCallParensAsNesting'`)
- [ ] A pair nested inside another pair but around a different node is not forgiven: source `((1) + 2)` against canon `(1 + 2)` is unfaithful (`go test ./internal/parse/ -run '^TestFaithfulDoesNotForgiveInnerPairAroundSubterm$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDoesNotForgiveInnerPairAroundSubterm'`)
- [ ] Doubled parens in a context-changing position are forgiven only down to one pair, never to none. perl measures `(($y)) x 3` and `($y) x 3` identically, both as `repeat vK/DOLIST,2`, but source `(($y)) x 3` against canon `$y x 3` is unfaithful (`go test ./internal/parse/ -run '^TestFaithfulNestingForgivesToOnePairNotZero$' -count=1 -v | grep -q '^--- PASS: TestFaithfulNestingForgivesToOnePairNotZero'`)
- [ ] canon keeps a call-argument pair around a grouped term: `print ((1+2)*3)` re-emits with both pairs, never as `print (1+2)*3` (`go test ./internal/parse/ -run '^TestCanonKeepsCallArgParenAroundGroupedTerm$' -count=1 -v | grep -q '^--- PASS: TestCanonKeepsCallArgParenAroundGroupedTerm'`)
- [ ] `my ($p, $q) = (1, 2)` and `my $p, $q` are different trees; canon re-emits each as written, and Faithful rejects either against the other (`go test ./internal/parse/ -run '^TestFaithfulDistinguishesMyListFromMyComma$' -count=1 -v | grep -q '^--- PASS: TestFaithfulDistinguishesMyListFromMyComma'`)
