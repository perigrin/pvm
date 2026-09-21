#!perl
# `$r->[2]` is ONE op: a `multideref` that has absorbed the pad lookup of
# `$r` along with the subscript.
#
# TIER 08 references
# INTRODUCES the arrow dereference
# USES my, print, the anonymous array constructor
#
# This is the file `05_brace_deref.t` is compared against, and the pair is
# this tier's version of tier 01's `.5` problem: two spellings that mean
# the same thing and must still be told apart.
#
# MEASURED perl 5.42.0, the whole print statement:
#
#   a  <0> pushmark s
#   b  <+> multideref($r->[2]) sK
#   c  <$> const[PV "\n"] s
#   d  <@> print vK
#
# The subscript chain is fused. `$r->[2]` does not emit `padsv` then
# `rv2av` then `aelem`; it emits one op naming the whole path, base
# included. The brace form emits `padsv[$r] sM/DREFAV` FIRST and then
# `multideref(->[2])` with an EMPTY base -- two ops where this has one.
# Both spellings use the op NAME `multideref`, and the brace form's extra
# op is tier 01's `padsv`, so a lint comparing name sets alone sees the
# two files as identical. The `expect tokens` claim is the only check
# that reads the spelling rather than the result: this file has an arrow
# and no `${`, and `05_brace_deref.t` asserts the reverse.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $r = [10, 20, 30]; print $r->[2], "\n"'
#   30

--- source
my $r = [10, 20, 30];
print $r->[2], "\n";

--- expect output
30

--- expect parses

--- expect tokens
one operator whose text is "->"
