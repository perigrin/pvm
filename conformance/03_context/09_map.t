#!perl
# `map` takes either a BLOCK or an EXPRESSION, and the optree cannot tell
# you which was written. This file carries its weight in TOKEN FACTS.
#
# TIER 03 context
# INTRODUCES mapstart mapwhile
# USES nothing from a later tier
#
# THE CENTRAL FACT, and the reason this file is mostly lexical. The two
# call forms are a genuine parse fork -- `map BLOCK LIST` takes no comma
# after the block, `map EXPR, LIST` requires one -- and the OP STREAM is
# the same for both: same ops, same order, same flags, same targs.
# Measured, `my @b = map { $_ } @a` and `my @b = map($_, @a)` both give
#
#   pushmark, pushmark, padav[@a] lM, mapstart lK, mapwhile lK, gvsv[*_]
#
# so nothing an op-name lint reads distinguishes the halves. A parser
# that took the comma form for the block form would emit an identical
# optree and is caught only at the token stream, which is where this file
# asserts. The comma IS the fork, so the file is written to contain
# exactly one -- `qw()` builds the list without any -- and the token fact
# below says so. One comma is the expression form; a parser that admitted
# a comma after a block, or required one, changes that count.
#
# WHAT IS NOT IDENTICAL, and it is not the ops. The two optrees differ in
# the COP SEQUENCE RANGE B::Concise prints beside each lexical: under the
# block form `@a` is `padav[@a:1,5]` and under the expression form
# `padav[@a:1,3]`. That range is a lexical's visibility, and a BLOCK is a
# SCOPE -- measured, a bare `{ 1; }` standing where the map block stands
# advances the counter by exactly the same 2. So the block's scope is
# real and leaves a trace, but the trace is in pad metadata rather than
# in any op, and the op-name lint does not read it. "Byte-identical" is
# too strong; "identical in every op" is the measured claim.
#
# The fourth statement is the discriminating pair. `my $n = map { $_ } @a`
# emits `mapstart sK` where the two before it emit `mapstart lK` -- same
# op, same operand, and in scalar context the answer is the COUNT of what
# the list form returns. As everywhere in this tier, what separates the
# halves is the flag, not the op name.
#
# The list is opened up by `$ENV{M}` because a CONSTANT argument erases
# the construct. `$ENV{M}` is unset when the runner executes, so
# `"$ENV{M}a"` is the one-character string `a` -- opaque at compile time,
# fixed at run time, and the reason the first element comes back `a`
# rather than the `x` that `qw` put there.
#
# No arithmetic appears in the block on purpose. `map { $_ + 1 } @a` is
# the obvious demonstration and emits `add`, which is tier 04's, one tier
# forward. A bare `$_` is the smallest body that is still a body.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @a = qw(x y z);
#              $a[0] = "$ENV{M}a";
#              my @block = map { $_ } @a;
#              my @expr  = map($_, @a);
#              my $n = map { $_ } @a;
#              print "@block @expr $n\n";'
#   a y z a y z 3
#
#   $ perl -MO=Concise,-exec -e 'my @a=(1); my @b = map { $_ } @a;' | grep padav
#   6  <0> padav[@a:1,5] lRM*/LVINTRO
#   g  <0> padav[@b:4,5] lRM*/LVINTRO
#   $ perl -MO=Concise,-exec -e 'my @a=(1); my @b = map($_, @a);' | grep padav
#   6  <0> padav[@a:1,3] lRM*/LVINTRO
#   g  <0> padav[@b:2,3] lRM*/LVINTRO

--- source
my @a = qw(x y z);
$a[0] = "$ENV{M}a";
my @block = map { $_ } @a;
my @expr  = map($_, @a);
my $n = map { $_ } @a;
print "@block @expr $n\n";

--- expect output
a y z a y z 3

--- expect parses

--- expect tokens
one operator whose text is ","
one variable whose text is "@block"
one variable whose text is "@expr"
