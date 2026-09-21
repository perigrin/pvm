# 06_control

`if`/`unless`, `while`/`until`, `for`/`foreach`, the postfix forms, and `goto`.

## Why this tier sits here

A conditional needs something to test and a loop needs something to vary,
and both have to be runtime values or the construct does not survive
compilation. `if (1)` emits no conditional op at all -- measured, the whole
test vanishes and only the taken branch remains -- so every file here binds
a value first and branches on it second. That binding is tier 05's `my` in a
scope tier 05 defines, over an operand tier 04 built, and a loop body is a
block, which is tier 05's subject. Tier 06 cannot come before the thing
whose truth it tests exists.

Check 2 asks whether this tier could move earlier. It could not sit before
05, because the loop body and the `else` block are blocks and the loop
variable is a scoped declaration -- `enteriter` allocates a pad slot with
`LVINTRO`, which is tier 05's machinery. It could not sit before 04, because
a condition worth testing is a comparison, and `$i < $ENV{N}` is a tier 04
operator. Moving it later is possible in the partial order -- 07 could be
written without it -- but it would put the tier that introduces jumps after
the tier that introduces the frames they jump out of, and `goto &sub` is
exactly the construct that straddles that line.

What the next tier needs from it: tier 07 declares subroutines, and a
subroutine that does not branch or loop is a name for a single expression.
`return EXPR if COND` is the ordinary shape of an early return, and it is
this tier's postfix `if`. `goto &sub` is the other half -- it replaces the
current frame with another sub's, which is a control transfer that only
means something once tier 07 supplies frames.

## DEPENDS ON

    05_scoping

Conditions and loop variables also lean on 02, 03 and 04, but a tier names
the one dependency it cannot be written without, and that is the block.

`print` and `$ENV{...}` appear throughout as FIXTURES. A branch has to be
observable to be measured, and the condition has to be a value the optimiser
cannot see through -- `shift`, `$ARGV[0]` and `$ENV{X}` all serve. Neither is
this tier's subject.

## INTRODUCES

    cond_expr enteriter goto iter last next redo unstack

## Why those ops, and not the ones the source implies

Measured with `perl -MO=Concise,-exec` under 5.42.0. Four places where the
op stream contradicts a plain reading of the source:

- **There is no `if` op, and `if` is not one thing.** A bare `if (COND) {
  BLOCK }` compiles to `and`, the same op as `COND && BLOCK` -- tier 04's
  op, claimed there, arriving here for an entirely different construct. An
  `if`/`else` compiles to `cond_expr` instead, the same op the ternary `?:`
  emits. So a source-level construct maps to two different ops depending on
  whether an `else` is present, only one of them is new to this tier, and
  the other is shared with an operator tier 04 already spells differently.
  The corpus needs both files; neither alone lints the tier.

- **`unless` is `or`, not an inverted `if`.** The discriminating pair is
  exact: `if ($ENV{X}) { print "y" }` and `unless ($ENV{X}) { print "n" }`
  emit identical op streams except that position 4 is `and` in the first
  and `or` in the second. There is no negation op anywhere -- perl does not
  invert the condition, it picks the other short-circuit. Which means a
  parser that desugars `unless` to `if (!COND)` is producing a tree perl
  never builds. Both ops belong to tier 04; what this tier contributes is
  the measurement that the two keywords differ by exactly one of them.

- **Postfix is not a separate construct.** `print "y" if $ENV{X}` and `if
  ($ENV{X}) { print "y" }` emit the same ops in the same order, down to the
  `nextstate`. The block form's braces do not even add an `enter`/`leave`
  pair at this nesting. The two spellings are one construct to perl, so the
  corpus files that pair them assert an equality rather than a difference
  -- which is the useful claim, because a parser that treats them as
  unrelated grammar productions will still pass an output test.
  The postfix LOOP modifiers are the exception that proves it: `print "y"
  while $ENV{X}` emits `enter`/`leave` and `unstack` but NO `enterloop`,
  while the block `while` emits `enterloop`/`leaveloop`. Same modifier
  syntax, different machinery, because a postfix loop has no block for
  `next`/`last` to target.

- **`while` and `until` differ by the same one op as `if` and
  `unless`.** Both emit `enterloop`, `unstack` and `leaveloop`; the
  condition test is `and` for `while` and `or` for `until`. The loop
  family and the conditional family share their branch ops entirely -- what
  makes a loop a loop is `enterloop`/`unstack`, not the test.

`enterloop` and `leaveloop` are tier 05's, claimed there for the bare block:
a bare `{ ... }` is a loop that runs once, so the loop ops are emitted three
tiers before anything that looks like a loop. This tier adds `unstack`,
which the bare block does not emit, and that is the op that distinguishes a
loop that iterates from a block that does not.

`enteriter` and `iter` are the `foreach` family and are what distinguishes
it from C-style `for`: `for (my $i = 0; ...; $i++)` emits `enterloop` like a
`while`, while `foreach my $x (@ARGV)` emits `enteriter`/`iter` and no
`enterloop` at all. Both close with `leaveloop`. So "for" names two
unrelated optrees, and the tier claims the ops that are new in each.

`next`, `last` and `redo` are claimed here rather than deferred because
`enterloop` names all three of their targets in its own dump --
`enterloop(next->g last->j redo->4)` -- so the loop ops and the jump ops are
one measurement, not two. Labelled forms emit the same ops carrying a
string (`next("OUTER")`); unlabelled forms carry none. Same op, and the
corpus distinguishes them by the argument rather than by the op name.

`goto` covers three spellings that share one op name and nothing else:
`goto LABEL` emits `goto("L")` with the label as a constant argument,
`goto $target` emits a unary `goto` over a pad value, and `goto &sub` emits
`rv2cv`/`srefgen`/`goto` inside the sub body -- which is not visible in the
main optree at all, and needs `-exec,g` to see. Only the op name is claimed;
`rv2cv` and `srefgen` belong to tiers 07 and 08.

Two erasures worth recording, because they are how a file in this tier can
silently measure nothing. `if (1) { print "y" } else { print "n" }` emits
`pushmark`, `const` and `print` and no branch op whatsoever -- the else
branch is gone from the binary. `while (0) { print "y" }` emits `enter`,
`nextstate` and `leave`: an empty program. A corpus file with a constant
condition passes its output test while testing none of this tier's ops,
which is why every file here uses a runtime value and why the ops LINT the
declared tier rather than deriving it.

## What writing the corpus changed

Nothing was removed from INTRODUCES. All eight ops are emitted by files in
this tier, and the corpus-wide claim check passes against them. Three
things the files measured that this README had asserted without one:

`goto` is backed by `goto LABEL` and `goto $target` only. Both were
measured and both fit the budget -- `goto("SKIP")` is a `<">` op carrying a
constant, `goto $t` a `<1>` unary over a pad slot, so the two spellings
differ in arity and op class as well as in when the target is known.
`goto &sub` was measured OUT, as this README predicted: it needs `rv2cv`
and `srefgen`, which tiers 07 and 08 claim. The op NAME is legitimately
this tier's because two spellings emit it here; the frame-replacing
spelling waits for the tier that supplies frames. Both jumps in
`10_goto.t` go forward to a label at the same scope depth, because perl
warns on a `goto` into or out of a construct and the harness compares
output byte for byte.

The postfix loop modifier cost an op the first time it was written.
`print $i-- while $i > 0` emits `postdec`, which no tier at or before 06
claims, so `09_postfix_while.t` spells the body `$i = $i - 1` instead. The
file's subject is the loop frame, not the decrement, and the lint was
right to refuse the shorter spelling.

`redo` needs a guard that is false at run time. It is the one jump with no
terminating spelling of its own -- an unguarded `redo` restarts the body
forever -- so both `08_next_last_redo.t` and the adjacency file put it
behind `$ENV{R} // 0`. The op compiles, sits adjacent to the statements
around it, and is never taken, which is exactly what the tier needs to
measure and nothing more.

## HARD MARKERS

    goto	goto·

One of the twelve `hardMarkers` places here. The list came from
`internal/parse/easy_test.go`, which used it to carve an "easy" subset out
of T1; in a graded corpus each entry is a tier placement rather than a
filter, so that file was deleted and the placements moved to the tiers.

`goto` is the keyword-table row measured at 4.5% clean over twenty-two
files, and it is this tier because `goto LABEL` is a loop-control statement
in everything but name -- the same code path as `next`, `last` and `redo`,
which the section above records as the ops this tier introduces.
`10_goto.t` is the file, and the note above already measures that the tier
backs `goto` with `goto LABEL` and `goto $target` only.

The probe column is the SOURCE probe verbatim, `·` standing for a
significant trailing space; the trailing space is what keeps the probe off
`gotoward`. `TestEveryHardMarkerPlaced` in `internal/conformance` requires
a file here that the probe finds.

## FILE ORDER

    01-03	conditionals
    04-07	loops
    08-10	jumps

The numbering is chosen, not computed, and the sections above depend on
it. The three conditional files are the `and`/`or`/`cond_expr` argument --
`if` and `unless` differ by exactly one op, and `if`/`else` is a different
op again -- and that argument is a comparison between neighbours. The four
loop files are the `enterloop`/`unstack` family, with `enteriter` arriving
at `07_foreach.t` as the thing that distinguishes `foreach` from C-style
`for`. The three jump files are what `enterloop` names in its own dump.

A regeneration would sort these alphabetically and interleave the
families: `03_goto.t` would land between `02_foreach.t` and
`04_if_else.t`, putting a jump between a loop and a conditional and
separating `10_while.t` from `09_until.t`, which differ by one op and are
only legible side by side. That is why this tier is not `derived`: the
arrangement carries the argument, and alphabetical order would destroy it
while leaving every test passing.
