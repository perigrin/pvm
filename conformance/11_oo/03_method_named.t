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
#
# THE TOKEN FACT, and why the direct spelling needs one as much as the
# indirect one does. `06_indirect_new.t` declares that `new Foo` holds NO
# arrow operator, which is the only place indirect object notation is
# visible at all -- both spellings emit the same four ops. But a negative
# alone is satisfied VACUOUSLY by a lexer that never emits `->`: one that
# folded the arrow into the word beside it, or dropped it as trivia,
# passes the indirect file's claim perfectly while getting every direct
# call in this tier wrong. This is the matching positive, and the pair is
# what makes either falsifiable.
#
# ONE arrow, not "at least one", which is why this file carries the claim
# rather than `00_adjacency.t`: the adjacency file makes four arrow calls
# and could only say something vaguer.

--- source
package Foo;
sub hi { return "hi" }
package main;
my $o = bless {}, "Foo";
print $o->hi, "\n";

--- expect output
hi

--- expect parses

--- expect tokens
one operator whose text is "->"
