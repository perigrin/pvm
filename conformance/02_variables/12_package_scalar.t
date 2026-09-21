#!perl
# `$::x` names a package scalar without declaring it, and assigning to one
# is where plain `sassign` enters the corpus.
#
# TIER 02 variables
# INTRODUCES package scalar
# USES nothing from a later tier
#
# Tier 01's `my $x = 0.5` emits `padsv_store` and no `sassign` at all --
# the lexical store is one op, not an assignment over a variable. The
# package scalar is the first place the two halves separate: `gvsv` fetches
# the glob's scalar slot and `sassign` puts the value in it. So the op a
# reader would look for in tier 01 is introduced four constructs into tier
# 02, by the spelling nobody writes first.
#
# `$::x` rather than `$main::x`: `::` with an empty package name IS `main`,
# and the short spelling is the one that makes the lexing question visible
# -- whether `$::` is a sigil plus a name that begins with a separator.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e '$::x = 1; print $::x, "\n"'
#   1

--- source
$::x = 1;
print $::x, "\n";

--- expect output
1

--- expect parses
