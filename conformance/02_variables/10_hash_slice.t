#!perl
# `@h{...}` is a hash slice, and `delete` on a slice is the one spelling
# where `delete` survives as an op of its own.
#
# TIER 02 variables
# INTRODUCES hash slice
# USES nothing from a later tier
#
# `delete $h{a}` emits NO delete op: it compiles to
# `multideref($h{"a"}) vK/DELETE`, the construct surviving only as a flag
# on an op named after something else. Deleting a SLICE is different --
# the optimiser declines to build a multideref for it, so `delete vK/SLICE`
# appears as a real op. That is why the tier claims `delete` for the slice
# and not for the element, and why `exists` is not claimed at all: it has
# no op in any spelling.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my %h = (a=>1,b=>2,c=>3); delete @h{"a","b"}; print scalar(keys %h), "\n"; print( (@h{"c"}), "\n")'
#   1
#   3

--- source
my %h = (a => 1, b => 2, c => 3);
delete @h{"a", "b"};
print scalar(keys %h), "\n";
print( (@h{"c"}), "\n");

--- expect output
1
3

--- expect parses
