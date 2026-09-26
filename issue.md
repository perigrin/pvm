---
title: Canon writes perl op names for unary plus and drops return EXPR entirely
state: pending
urgency: normal
milestone: perl-git-t-parses
created: 2026-09-26T20:30:56.927867322Z
updated: 2026-09-26T20:30:56.927867322Z
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

## Acceptance criteria

- [ ] `my $v; +$v;` canons to a unary plus and not to `pos` (`go test ./internal/parse/ -run TestCanon -count=1`)
- [ ] `my $v; return $v || $v;` canons with both operands present (`go test ./internal/parse/ -run TestCanon -count=1`)
- [ ] `next LABEL` and `last LABEL` still emit the bareword target, no regression (`go test ./internal/parse/ -run TestCanon -count=1`)
- [ ] REGRESSION GUARD: `lvalue_unary_plus.t` and `attribute_method_named_sub_deparse.t` return to 0 in `t1canon.ratchet` (`go test ./internal/parse/ -run TestCanonRatchet -count=1`)
- [ ] REGRESSION GUARD: the whole suite holds (`go test ./... -count=1`)
