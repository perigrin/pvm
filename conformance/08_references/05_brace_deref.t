#!perl
# `${$r}[2]` means what `$r->[2]` means and does NOT compile to the same
# ops: the brace form leaves the pad lookup as a separate `padsv`.
#
# TIER 08 references
# INTRODUCES the brace dereference of a scalar-held array reference
# USES my, print, the anonymous array constructor
#
# MEASURED perl 5.42.0, the whole print statement:
#
#   a  <0> pushmark s
#   b  <0> padsv[$r:1,4] sM/DREFAV
#   c  <+> multideref(->[2]) sK
#   d  <$> const[PV "\n"] s
#   e  <@> print vK
#
# Against `04_arrow_deref.t`'s single `multideref($r->[2])`: same result,
# same op NAMES plus one that tier 01 already owns. The difference is
# visible only in the op COUNT and the private flag, so a tier lint
# comparing name sets cannot distinguish this file from that one. That is
# why the pair exists and why both carry token claims -- this file has no
# arrow at all, which no op-level check can see.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $r = [10, 20, 30]; print ${$r}[2], "\n"'
#   30

--- source
my $r = [10, 20, 30];
print ${$r}[2], "\n";

--- expect output
30

--- expect parses

--- expect tokens
no operator whose text is "->"
no variable whose text is "${$r}"
