#!perl
# `import` is an ordinary method call on a package name. There is no import
# op; `use` gets one by calling it, and the op stream cannot tell that call
# from any other.
#
# TIER 12 packages
# INTRODUCES the import call
# USES nothing from a later tier
#
# `Marker->import("tag")` emits `pushmark`, `const[PV "Marker"] sM/BARE`,
# `const[PV "tag"] sM`, `method_named[PV "import"]`, `entersub` -- tiers 01,
# 07 and 11 between them and nothing left over. This tier's claim is that
# `import` is not special, and the way to demonstrate that is to call it
# directly and get the same stream `use` would have produced.
#
# Called at run time rather than from `BEGIN` so the output order is the
# source order. Under `BEGIN` the same call runs during compilation and
# prints first, which is true of any BEGIN block and is tier 05's business,
# not this tier's.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'package Marker; sub import { print "import $_[0] $_[1]\n" } package main; Marker->import("tag"); print "after\n"'
#   import Marker tag
#   after

--- source
package Marker;
sub import { print "import $_[0] $_[1]\n" }
package main;
Marker->import("tag");
print "after\n";

--- expect output
import Marker tag
after

--- expect parses
