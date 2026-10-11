---
title: Canon writes perl op names for unary plus and drops return EXPR entirely
state: done
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T20:30:56.927867322Z
updated: 2026-10-11T10:47:56.637482471Z
sessions:
- start_sha: 971b903102dc4a335660172f3b1b5ecad5cf3330
  end_sha: 971b903102dc4a335660172f3b1b5ecad5cf3330
  commits: 0
  started_at: 2026-10-11T10:47:56.405661787Z
  ended_at: 2026-10-11T10:47:56.637482471Z
transitions:
- state: in-progress
  actor: human:git-zhi
  timestamp: 2026-10-11T10:47:56.405661787Z
- state: done
  actor: human:git-zhi
  timestamp: 2026-10-11T10:47:56.637482471Z
---

Canon writes perl's OP NAME where the source had an operator, and
`return EXPR` loses its whole expression. Two defects in `Canon`, both
found while the sub-attribute work (01a0de97) made the surrounding code
reachable for the first time.

Neither is caused by that change: each reproduces on a two-statement
program with no attribute anywhere.

## Measured 2026-09-26 at `35549df0` plus 01a0de97's change

    my $v; +$v;
      canon "my $v;pos $v;"
      tree  (unary "pos" (term "$v"))

    my $v; return $v || $v;
      canon "my $v;return ||; "
      tree  (loop_control "return" (binary "||" (term "$v") (term "$v")))

Both trees are RIGHT and both emissions are WRONG, so `Unknown` is 0 for
each and the parse ratchet sees nothing. Only the canon fixpoint ratchet
notices, and only once the code around them stops being an Unknown.

## Defect 1: `prefixName` leaks an op name into the emission

`term.go`'s `prefixName` maps the operator text to perl's OP NAME --
`+` to `pos`, `-` to `neg`, `\` to `ref` -- and `Canon`'s Unary arm
writes `n.Text`. So a unary plus re-emits as the `pos` builtin, which is
a DIFFERENT PROGRAM: `pos $v` returns the match position of `$v`.

`-` and `\` have the same shape and are not yet observed to bite, because
`neg $v` and `ref $v` are also real builtins that a re-parse accepts
silently.

## Defect 2: `Canon`'s LoopControl arm writes `c.Text`, not the child

`canon.go:353` emits `n.Text` then each child's `Text` with no recursion.
That is right for `next LABEL`, where the child is a bareword, and wrong
for `return EXPR`: a Binary's `Text` is its operator, so
`return $v || $v` emits `return ||` and both operands are gone.

`return` shares the LoopControl kind with `last`/`next`/`redo`, which is
why one arm serves two shapes that need different emissions.

## Reach

`lvalue_unary_plus.t` and `attribute_method_named_sub_deparse.t` are the
two T1 files where 01a0de97 moved `t1canon.ratchet` from 0 to 1 -- the
only rises in that change, and both of them are these two defects
becoming visible rather than anything regressing.

### Re-measured (2026-10-11)

Measured at pu `24bfac4a`. `my $v; +$v;` canons as `my $v;+$v;`, a fixpoint, not `pos $v`. `my $v; return $v || $v;` canons as `my $v;return $v || $v;`, both operands present, as perl Deparses it. `next LABEL;` and `last LABEL;` canon as themselves. `lvalue_unary_plus.t` and `attribute_method_named_sub_deparse.t` are both 0 in `t1canon.ratchet`. Fixed by c80ab769, the return operand, and 26ab62a3, unary plus.

The criteria ran `-run TestCanon` unanchored, which matches eleven tests, with no PASS grep. They now name TestCanonUnaryPlusIsPlus, TestReturnList and TestLoopControlWithoutModifier. TestReturnList's `my $v; return $v || $v;` case was added in test/pgt-reconcile; it fails as `my $v;return ||;` when canon's LoopControl arm writes a child's Text again. The ratchet criterion checks the two files' lines, and TestCanonTokenIdentity holds every file at 0 there.

The conformance and whole-suite commands carry `-timeout=30m`: internal/conformance took 696s in a full run, past go test's 10-minute default, and timed out without it.

## Acceptance criteria

- [ ] `my $v; +$v;` canons to a unary plus and not to `pos` (`go test ./internal/parse/ -run '^TestCanonUnaryPlusIsPlus$' -count=1 -v | grep -q '^--- PASS: TestCanonUnaryPlusIsPlus'`)
- [ ] `my $v; return $v || $v;` canons with both operands present (`go test ./internal/parse/ -run '^TestReturnList$' -count=1 -v | grep -q '^--- PASS: TestReturnList'`)
- [ ] `next LABEL` and `last LABEL` still emit the bareword target, no regression (`go test ./internal/parse/ -run '^TestLoopControlWithoutModifier$' -count=1 -v | grep -q '^--- PASS: TestLoopControlWithoutModifier'`)
- [ ] REGRESSION GUARD: `lvalue_unary_plus.t` and `attribute_method_named_sub_deparse.t` are 0 in `t1canon.ratchet` and stay identical (`grep -q '^0 lvalue_unary_plus.t$' internal/parse/testdata/t1canon.ratchet && grep -q '^0 attribute_method_named_sub_deparse.t$' internal/parse/testdata/t1canon.ratchet && go test ./internal/parse/ -run '^TestCanonTokenIdentity$' -count=1 -v | grep -q '^--- PASS: TestCanonTokenIdentity'`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1 -timeout=30m`)
