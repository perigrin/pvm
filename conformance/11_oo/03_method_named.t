#!perl
# `$o->hi`, with the method name written out, resolves the name at compile
# time: `method_named[PV "hi"]`.
#
# TIER 11 oo
# INTRODUCES method call with a compile-time name
# USES nothing from a later tier
#
# This is the spelling everyone measures, and measuring only it would miss
# `04_method_dynamic.t` and `05_method_super.t`, which are different ops
# for what a reader calls the same construct. The `entersub` that follows
# is tier 07's; what this tier adds is how the callee is found.
#
# `package Foo;` carries the method. Measured, a bare `package` statement
# emits NO runtime op -- the op stream below starts at the `bless` on line
# 4 -- so the classic files can name packages without borrowing tier 12.
# The `sub hi` body is not in this dump either: `perl -MO=Concise,-exec`
# with no sub named dumps the main program alone.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'package Foo; sub hi { return "hi" } package main; my $o = bless {}, "Foo"; print $o->hi, "\n"'
#   hi
#
# and the dispatch:
#
#   b  <.> method_named[PV "hi"] l
#   c  <1> entersub[t3] lKRS/TARG

--- source
package Foo;
sub hi { return "hi" }
package main;
my $o = bless {}, "Foo";
print $o->hi, "\n";

--- expect output
hi

--- expect parses
