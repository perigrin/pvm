#!perl
# `@{$r}` dereferences a scalar as a whole array, and the op it emits --
# `rv2av` -- is tier 02's, reached here by a route tier 02 does not have.
#
# TIER 08 references
# INTRODUCES the @{ } dereference
# USES my, print, the reference operator, interpolation
#
# This is one of the tier's two hard markers, `deref-at`, and it is the
# spelling tier 11's method bodies are written in.
#
# `@{$r}`, `@$r` and `$r->@*` all emit `rv2av` and are indistinguishable
# in the op stream. The op set cannot tell which was written, which is
# the recurring shape of this tier: the `rv2*` family collapses several
# spellings each. Only a token claim separates them, so this file asserts
# it has no arrow.
#
# Interpolating the dereference inside a string is what makes the output
# deterministic AND keeps `join` in tier 03 where it belongs -- printing
# `@{$r}` as a bare list would run the three elements together as `102030`,
# which is true but reads as a bug.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @a = (10, 20, 30); my $r = \@a; print "@{$r}\n"'
#   10 20 30

--- source
my @a = (10, 20, 30);
my $r = \@a;
print "@{$r}\n";

--- expect output
10 20 30

--- expect parses

--- expect tokens
one operator whose text is "\\"
no operator whose text is "->"
