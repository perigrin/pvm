#!perl
# `$o->SUPER::hi()` is a third dispatch op again: `method_super`, which
# starts its search in the CURRENT package's @ISA rather than in the
# invocant's class.
#
# TIER 11 oo
# INTRODUCES SUPER:: method dispatch
# USES nothing from a later tier
#
# `SUPER::` resolves against the package the call is COMPILED in, not the
# one the object is blessed into, which is why this file makes the call at
# file scope inside `package Derived;` rather than from inside a method.
# Written the usual way -- `sub hi { $_[0]->SUPER::hi() }` -- the op would
# sit in the sub's own optree, and `perl -MO=Concise,-exec` with no sub
# named dumps the main program alone, so nothing would measure it. The
# file is shaped by what can be observed, and says so.
#
# `our @ISA = ("Base")` is tier 02's `gv`, `rv2av` and `aassign`.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'package Base; sub hi { "base" } package Derived; our @ISA = ("Base"); my $o = bless {}, "Derived"; print $o->SUPER::hi(), "\n"'
#   base
#
# and the third dispatch op:
#
#   i  <.> method_super[PV "hi"] l
#   j  <1> entersub[t7] lKRS/TARG

--- source
package Base;
sub hi { return "base" }
package Derived;
our @ISA = ("Base");
my $o = bless {}, "Derived";
print $o->SUPER::hi(), "\n";

--- expect output
base

--- expect parses
