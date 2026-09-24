# 06_control

`if`/`unless`, `while`/`until`, `for`/`foreach`, the postfix forms, `goto`,
and the two block-valued expressions `do BLOCK` and `eval BLOCK`.

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

    cond_expr die enteriter entertry exit goto iter last leavetry next redo
    time unstack

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

  And `elsif` is a THIRD form of the same keyword, compiling to NESTED
  `cond_expr` -- one per condition, no op of its own. So three source
  constructs share two ops between them, and what distinguishes `elsif` is
  the shape of the nesting, which `-exec` order cannot show.
  `04_elsif.t` pins it behaviourally instead, by making the MIDDLE branch
  the one taken: neither a lone `if` nor an `else` can produce that. Its
  token facts count rather than forbid, since a lexer reading `elsif` as
  `else` followed by `if` would emit two words spelled `else` in a file
  that contains one.

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

`entertry` and `leavetry` are `eval BLOCK`, and they are here because
`eval` is TWO constructs wearing one keyword that share no op at all.
Measured, `eval { 1 }` emits `entertry`/`leavetry` and `eval "1"` emits
`entereval`. The string form compiles Perl the compiler never saw and
belongs to `14_recursive`, which claims `entereval` already; the block
form compiles nothing new and marks a FRAME the runtime can unwind back
to. That is a control transfer in the same family as `next`, `last` and
`goto` -- a jump whose target the construct names and whose distinction is
that nothing in the source spells the jump. Both ops were unclaimed by
every tier before `14_eval_block.t` was written, so claiming them here is
additive and `claimonce_test.go` finds no collision.

`die` is NOW this tier's, and the order the two files were written in is
why `14_eval_block.t` reads as it does. When that file was written `die`
was claimed by no tier at all, so the obvious `eval { die "x" }` reached
forward and the lint refused it; it traps a division by zero instead,
exercising the same frame with `divide`, an op tier 04 already owns. The
budget picked the operand. `15_die.t` then claimed `die` here on the
argument the `entertry` paragraph above makes -- a control transfer whose
target the construct names and which nothing in the source spells -- which
was only available once the frame itself was spelled in this tier. The
older file is left as measured: it is still the file that shows `entertry`
catching a failure the source does not raise by name.

`exit` joins it and is the same family with the opposite reach: `die`
unwinds to the nearest `entertry`, `exit` unwinds past every one of them.
`16_exit.t` is that measurement -- an `exit` inside an `eval BLOCK`, with
the frame visibly built in the optree and the statement after the `eval`
compiled and never reached. Neither op was claimed by any tier before
these two files, so both additions are additive and `claimonce_test.go`
finds no collision. `10_io`'s README records the `die` gap from the other
side: its files `open` unchecked because `or die` would have widened that
tier's set by an op belonging "to a tier that does not exist yet", and
notes that the check can come back when a tier claims `die`. This is that
tier.

The three `do` spellings resolve to two files and one measurement that
nothing here can carry. `do BLOCK` in expression position emits no op of
its own -- an `enter`/`leave` pair with an `s` flag, tier 01's ops -- and
`do { 1 }` emits nothing at all, folding to a bare `const`. `do BLOCK
while COND` is a loop that emits `unstack` and no `enterloop`, the same
shape as the postfix modifier. `do FILE` emits `dofile`, which is a
file-loading op in tier 12's family rather than a control construct, and
is left out of this tier deliberately. `do SUB` is not measured at all
because it does not exist: under 5.42.0 `do f()` is a syntax error, so the
third of the "three unrelated parses" is a parse perl no longer has.

Two erasures worth recording, because they are how a file in this tier can
silently measure nothing. `if (1) { print "y" } else { print "n" }` emits
`pushmark`, `const` and `print` and no branch op whatsoever -- the else
branch is gone from the binary. `while (0) { print "y" }` emits `enter`,
`nextstate` and `leave`: an empty program. A corpus file with a constant
condition passes its output test while testing none of this tier's ops,
which is why every file here uses a runtime value and why the ops LINT the
declared tier rather than deriving it.

## What writing the corpus changed

Nothing was removed from INTRODUCES. All twelve ops are emitted by files
in this tier, and the corpus-wide claim check passes against them. Three
things the files measured that this README had asserted without one:

`goto` is backed by `goto LABEL` and `goto $target` only. Both were
measured and both fit the budget -- `goto("SKIP")` is a `<">` op carrying a
constant, `goto $t` a `<1>` unary over a pad slot, so the two spellings
differ in arity and op class as well as in when the target is known.
`goto &sub` was measured OUT, as this README predicted: it needs `rv2cv`
and `srefgen`, which tiers 07 and 08 claim. The op NAME is legitimately
this tier's because two spellings emit it here; the frame-replacing
spelling waits for the tier that supplies frames. Both jumps in
`12_goto.t` go forward to a label at the same scope depth, because perl
warns on a `goto` into or out of a construct and the harness compares
output byte for byte.

The postfix loop modifier cost an op the first time it was written.
`print $i-- while $i > 0` emits `postdec`, which no tier at or before 06
claims, so `11_postfix_while.t` spells the body `$i = $i - 1` instead. The
file's subject is the loop frame, not the decrement, and the lint was
right to refuse the shorter spelling.

`redo` needs a guard that is false at run time. It is the one jump with no
terminating spelling of its own -- an unguarded `redo` restarts the body
forever -- so both `10_next_last_redo.t` and the adjacency file put it
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
`12_goto.t` is the file, and the note above already measures that the tier
backs `goto` with `goto LABEL` and `goto $target` only.

The probe column is the SOURCE probe verbatim, `·` standing for a
significant trailing space; the trailing space is what keeps the probe off
`gotoward`. `TestEveryHardMarkerPlaced` in `internal/conformance` requires
a file here that the probe finds.

## FILE ORDER

    conditionals
    loops
    jumps
    blocks-and-terminations

The tier's topics, which are its grouping: a case lives in the topic
whose subject it shares, and the topic file name IS the group name.

The three conditional cases are the `and`/`or`/`cond_expr` argument --
`if` and `unless` differ by exactly one op, and `if`/`else` is a
different op again -- and that argument is a comparison between
neighbours, which is why they share a topic rather than being three
files a reader has to notice are related.

`blocks-and-terminations.md` is three former groups in one topic:
block-valued expressions, terminations, and the niladic builtins. They
were separate ranges when a group had to be a run of numbered files;
they are one document now because what they have in common -- a
construct whose VALUE or whose EXIT is the subject -- reads better as
one.
