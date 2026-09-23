#!perl
# `[10, 20, 30]` builds an array and yields a reference to it in one op,
# `anonlist`, with no named array anywhere in the program.
#
# TIER 08 references
# INTRODUCES the anonymous array constructor
# USES my, print, scalar, ref
#
# The anonymous HASH constructor is deliberately NOT here. `{}` and
# `{ %a }` compile to `emptyavhv` and `anonhash`, which tier 11 claims
# where objects are built; writing one in this tier would use an op a
# LATER tier owns and the lint would refuse the file. `[...]` alone is
# tier 08's.
#
# `ref` is what makes the assertion deterministic. A reference stringifies
# as `ARRAY(0x55d3...)`, whose address changes every run, so this prints
# the stable category name instead.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $r = [10, 20, 30]; print scalar(@$r), " ", ref($r), "\n"'
#   3 ARRAY

--- source
my $r = [10, 20, 30];
print scalar(@$r), " ", ref($r), "\n";

--- expect output
3 ARRAY

--- expect parses

--- expect tokens
one operator whose text is "["
