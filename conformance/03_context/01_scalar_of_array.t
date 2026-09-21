#!perl
# `my $n = @a` and `my $n = scalar(@a)` are the same program. The keyword
# that names this tier's subject compiles to no op of its own.
#
# TIER 03 context
# INTRODUCES scalar context on an array
# USES nothing from a later tier
#
# This is the tier's baseline, and it is the file that proves the tier
# cannot be linted by op NAME alone. Both statements compile to
# `padav[@a] s` followed by `padsv_store` -- byte-identical optrees. What
# separates scalar context from list context is the `s` versus `l` FLAG
# B::Concise prints on `padav`, and a flag is not an op.
#
# So this file introduces no new op. It is here because a tier whose first
# file emits something new would imply the tier is a set of ops, and it is
# not; it is a set of contexts, of which one leaves no trace.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @a=(10,20,30); my $n=@a; my $m=scalar(@a); print "$n $m\n"'
#   3 3

--- source
my @a = (10, 20, 30);
my $n = @a;
my $m = scalar(@a);
print "$n $m\n";

--- expect output
3 3

--- expect parses

--- expect tokens
one word whose text is "scalar"
