#!perl
# `my $x` in an inner block shadows an outer `my $x` for the extent of
# that block, and the outer binding is unchanged when the block ends.
#
# TIER 05 scoping
# INTRODUCES lexical declaration
# USES nothing from a later tier
#
# This is the file where `my` stops being scaffolding. Tier 01 writes
# `my $x = 1` in every case, but only to hold a literal still; nothing
# there asks what the `my` does. Here the question is the scope, and the
# answer is visible only because two declarations of the SAME NAME are
# live at once and print different values.
#
# The op stream cannot tell this file from a tier 01 one. Both emit
# `const` then `padsv_store[$x] vKS/LVINTRO`, both tier 01's ops, and the
# two `$x` differ only by pad slot -- `[$x:1,5]` against `[$x:3,4]` in the
# concise output. That is the honest position of this whole tier: what
# distinguishes it is the declared subject, not the op set.
#
# The block brings `enterloop`/`leaveloop` with it, which is this tier's
# own op pair and the subject of 05_bare_block.t.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $x = 1; { my $x = 2; print "$x\n"; } print "$x\n"'
#   2
#   1

--- source
my $x = 1;
{
  my $x = 2;
  print "$x\n";
}
print "$x\n";

--- expect output
2
1

--- expect parses
