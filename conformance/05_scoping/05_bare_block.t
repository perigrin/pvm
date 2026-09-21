#!perl
# A bare `{ ... }` at statement position is a SCOPE: declarations inside
# it are gone on the far side, and it executes exactly once.
#
# TIER 05 scoping
# INTRODUCES bare block
# USES nothing from a later tier
#
# This is the tier's own op pair, and the surprising one. A bare block
# compiles to `enterloop`/`leaveloop` -- perl builds it as a LOOP THAT
# RUNS ONCE, so that `last`, `next` and `redo` have something to act on
# inside it. The braces look like grouping and the optree says loop.
#
# The contrast is measured, because the same braces produce different
# ops depending on what precedes them. Under 5.42.0:
#
#   $ perl -MO=Concise,-exec -e '{ print "a\n"; }'
#   ... enter nextstate enterloop nextstate pushmark const print leaveloop leave
#
#   $ perl -MO=Concise,-exec -e 'my $c = 1; if ($c) { my $y = 2; print "$y\n"; }'
#   ... and enter nextstate const padsv_store ... print leave leave
#
# The `if` block gets `enter`/`leave`, tier 01's ops, because nothing can
# jump out of it; only the bare form gets the loop pair. (With a constant
# condition -- `if (1) { ... }` -- the block ops vanish entirely, folded
# away, which is the optimiser-erases-constructs warning in miniature.)
# So `enterloop`/`leaveloop` belong to this tier and not to 06_control,
# even though 06 is where loops live.
#
# The loop-ness is not probed behaviourally here. Showing it would take a
# `last` inside the block, and `last` compiles to a `last` op that no
# tier at or before 05 claims -- measured; the op stream carries a literal
# `last` between `print` and `leaveloop`. That probe belongs with the loop
# control tier. What this file asserts is the part inside the budget: the
# scope closes, and the block is entered once and not repeated.
#
# The second, empty-ish block on line 7 is what shows "once": a construct
# perl builds as a loop prints `again` a single time. Its ops are the
# same pair -- `{ 1; }` keeps its `enterloop`/`leaveloop` under 5.42.0,
# measured rather than assumed, so an optimiser that erased a do-nothing
# block would be visible as a missing op rather than as identical output.
#
# MEASURED perl 5.42.0:
#
#   $ perl -w conformance/05_scoping/05_bare_block.t
#   in 2
#   out 1
#   again

--- source
my $x = 1;
{
  my $x = 2;
  print "in $x\n";
}
print "out $x\n";
{
  print "again\n";
}

--- expect output
in 2
out 1
again

--- expect parses
