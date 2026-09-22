#!perl
# `unpack` reads the same opaque template `pack` writes, and returns a
# LIST where `pack` returns one string.
#
# TIER 13 opaque
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# 12_pack_template.t establishes that the template is one string literal.
# This file exists because `unpack` is a SEPARATE OP with a separate
# arity, and a compiler that implemented `pack` and stopped would pass 12
# and fail here. Measured, the two differ in the optree beyond their
# names:
#
#   i  <@> pack[t5] sK/2       -- scalar context, one value out
#   b  <@> unpack lK/2         -- LIST context, `l` where pack has `s`
#
# `pack` folds over constant arguments and so does `unpack` above it, so
# the source uses the same `$ENV{X} // default` idiom for the same
# reason 12 does: a template applied to a constant string is a compile-
# time computation and the op is gone before the optree exists.
#
# THE WRONG PARSE THIS RULES OUT: `unpack` read as a scalar-returning
# call. The pinned output is three numbers, which only exists if the
# result was a list of three that `@c` absorbed. A compiler that gave
# `unpack` `pack`'s scalar arity would print one value here and the
# output pin would catch it -- so this is one of the few files in a tier
# of token facts where the BEHAVIOUR carries a real claim.
#
# `C3` is 12's template, deliberately: the round trip is the assertion.
# 12 packs 65, 66, 67 into `ABC` and this unpacks `ABC` back into 65, 66,
# 67. Either file alone could be satisfied by a compiler that had the
# template wrong in a compensating way; the pair cannot.
#
# `"@c"` interpolates the array with `$"` between elements, which is a
# space by default -- that is the `gvsv[*"]` and `join` in the optree,
# tiers 02 and 03. `print @c` would print `656667` with no separator and
# would not show the list had three elements.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $s = $ENV{X} // "ABC"; my @c = unpack("C3", $s); print "@c\n";' | od -c
#   0000000   6   5       6   6       6   7  \n
#   0000011

--- source
my $s = $ENV{X} // "ABC";
my @c = unpack("C3", $s);
print "@c\n";

--- expect output
65 66 67

--- expect parses

--- expect tokens
one string literal whose text is "\"C3\""
no word whose text is "C3"
