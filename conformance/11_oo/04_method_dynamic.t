#!perl
# `$o->$m`, with the method name in a variable, is a DIFFERENT op:
# `method`, resolved at run time, not `method_named`.
#
# TIER 11 oo
# INTRODUCES method call with the name in a scalar
# USES nothing from a later tier
#
# The source differs from `03_method_named.t` by one sigil and the op
# changes entirely. `method_named` carries the name as a constant in the
# op itself; `method` takes it off the stack, so the padsv that supplies
# it is part of the call rather than an argument to it.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'package Foo; sub hi { "hi" } package main; my $o = bless {}, "Foo"; my $m = "hi"; print $o->$m, "\n"'
#   hi
#
# and the dispatch, where the name arrives as an operand:
#
#   d  <0> padsv[$o:3,5] sM
#   e  <0> padsv[$m:4,5] s
#   f  <.> method lK/1
#   g  <1> entersub[t4] lKRS/TARG

--- source
package Foo;
sub hi { return "hi" }
package main;
my $o = bless {}, "Foo";
my $m = "hi";
print $o->$m, "\n";

--- expect output
hi

--- expect parses
