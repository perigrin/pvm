#!perl
# A package hash reaches its storage through `rv2hv`, where a lexical hash
# is a `padhv` alone.
#
# TIER 02 variables
# INTRODUCES package hash
# USES nothing from a later tier
#
# The element access is `multideref` either way -- the optimiser folds the
# glob lookup into the deref chain just as it folds the pad lookup -- so
# the difference the file pins is in naming the hash itself, not in
# subscripting it. `scalar(keys %::h)` is what forces the bare name into
# the optree, and that is `gv` then `rv2hv`.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e '%::h = (a => 1, b => 2); print scalar(keys %::h), "\n"; print $::h{a}, "\n"'
#   2
#   1

--- source
%::h = (a => 1, b => 2);
print scalar(keys %::h), "\n";
print $::h{a}, "\n";

--- expect output
2
1

--- expect parses
