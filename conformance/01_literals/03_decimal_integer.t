#!perl
# A decimal integer is one numeric literal, and it is the baseline every
# other spelling in this tier deviates from.
#
# TIER 01 literals
# INTRODUCES numeric literal
# USES nothing from a later tier
#
# The tier had twelve construct files and none of them was a plain
# integer: `02_decimal.t` is `0.5`, the FRACTIONAL boundary, and the rest
# are hexadecimal, binary, octal, exponent, v-string and so on -- every
# one a deviation from a baseline no file stated. `= 42;` appeared only in
# `00_adjacency.t`, which introduces nothing, so the tier's simplest
# construct was the one it never asserted.
#
# That is not a hypothetical gap. `TestTierLiteralsCoversGlossary` names
# `decimal integer` among the fifteen boundaries this tier owns, and the
# check passed because it joined the adjacency file's source in with the
# construct files'. Excluding that file made three boundaries fail at
# once, this being the third.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'my $x = 42; print "$x\n"'
#   ... const[IV 42] ... padsv_store ... multiconcat ... print ...
#
# `const[IV 42]`, an INTEGER const, where `02_decimal.t`'s `0.5` gives
# `const[NV 0.5]`. Same op, different SV type, and the optree is the only
# place that difference is visible -- output prints `42` either way. The
# token fact is what pins the spelling.

--- source
my $x = 42;
print "$x\n";

--- expect output
42

--- expect parses

--- expect tokens
one numeric literal whose text is "42"
no operator whose text is "."
