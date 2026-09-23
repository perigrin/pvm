#!perl
# `..` in LIST context is the range operator, and one `..` in source
# emits THREE ops.
#
# TIER 03 context
# INTRODUCES the range operator
# USES nothing from a later tier
#
# `..` is three operators wearing one spelling, and before this file the
# corpus claimed none of them. The only `..` anywhere was
# `for my $i (1 .. 2)` in tier 07's adjacency file, which is loop
# scaffolding rather than a claim.
#
# Which reading applies is decided by CONTEXT, which is this tier's
# subject: in list context it builds a list, in scalar context it is a
# stateful flip-flop. `12_range_string.t` and `13_flipflop.t` are the
# other two readings.
#
# THE ENDPOINT IS A RUNTIME VALUE FOR A PLACEMENT REASON, NOT A PARSING
# ONE. Measured perl 5.42.0:
#
#   my @r = (1 .. 3)                no range, flip or flop op
#   my @r = (1 .. scalar(@a))       range flip flop
#
# Constant folding is an OPTIMISER effect and reaches only the op stream.
# `(1 .. 3)` lexes as `Operator("..")` and parses to a range either way;
# the fold happens afterwards and changes nothing this file claims. So a
# constant-endpoint file would be a perfectly good parsing claim.
#
# What it could not do is let this tier DECLARE `range` in its
# INTRODUCES block: the lint checks that a claimed op is actually
# emitted, and a folded range emits none. The endpoint is `scalar(@a)`
# to satisfy that declaration, not to make the construct parse.
#
# `scalar(@a)` rather than tier 04's `$ENV{X} // <default>` because `dor`
# is tier 04's op and out of this tier's budget -- and because a
# scalar-context array is this tier's own subject, which makes the
# operand a second small claim rather than scaffolding.
#
# ONE `..` EMITS ALL THREE OPS, which is worth stating because it defeats
# the obvious discriminator. `range`, `flip` and `flop` all appear for a
# plain list range -- the optimiser builds the flip-flop machinery even
# where the context means it can never be used. So the op set does NOT
# separate the three readings and the weight falls on OUTPUT, which
# separates them cleanly.
#
#   $ perl -e 'my @a=(1,2,3); my @r = (1 .. scalar(@a)); print "@r"'
#   1 2 3

--- source
my @a = (1, 2, 3);
my @r = (1 .. scalar(@a));
print "@r\n";

--- expect output
1 2 3

--- expect parses

--- expect tokens
one operator whose text is ".."
no operator whose text is "."
