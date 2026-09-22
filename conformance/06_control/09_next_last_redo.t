#!perl
# `next`, `last` and `redo` are three nullary ops jumping to the three
# targets `enterloop`/`enteriter` already names in its own dump.
#
# TIER 06 control
# INTRODUCES the loop control statements
# USES nothing from a later tier
#
# They are claimed in this tier rather than deferred because the loop op's
# dump is where their targets live -- `enteriter(next->w last->z redo->g)`
# -- so the loop ops and the jump ops are one measurement, not two. A tier
# that claimed the loops and left the jumps for later would be claiming
# half of a single line of output.
#
# All three take no operands. The labelled forms emit the SAME op carrying
# a string -- `next("OUTER")` -- so the corpus distinguishes the spellings
# by the argument rather than by the op name, and a parser that produces a
# different node kind for the labelled form is producing a distinction
# perl does not make.
#
# `redo` is exercised under a guard that is false at runtime, which is the
# only way to emit the op without writing a loop that does not terminate:
# `$ENV{R}` is unset, so `// 0` is a stable false and the op is compiled,
# reachable and not taken.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'my @l = ("a","b","c");
#     foreach my $x (@l) { next if $x eq "b"; last if $x eq "c"; print $x }'
#   ...
#   f  <{> enteriter(next->w last->z redo->g)[$x:3,6] vKS/LVINTRO
#   ...
#   k  <|> and(other->l) vK/1
#   l      <0> next v*
#   ...
#   q  <|> and(other->r) vK/1
#   r      <0> last v*
#
# The `v*` flag is what marks a jump: the op never returns to its
# successor, so the stream's textual order is not its execution order.

--- source
my $r = $ENV{R} // 0;
my @l = ($ENV{A} // "a", "b", "c", "d");
foreach my $x (@l) {
    redo if $r;
    next if $x eq "b";
    last if $x eq "d";
    print $x;
}
print "\n";

--- expect output
ac

--- expect parses
