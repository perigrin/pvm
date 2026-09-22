#!perl
# `exists` and `delete` on a hash ELEMENT are the two constructs this tier
# names that leave no op of their own behind.
#
# TIER 02 variables
# INTRODUCES exists, delete on an element
# USES nothing from a later tier
#
# The tier README's longest passage is about this and the tier shipped
# without a file that writes either word. Measured, `exists $h{a}` compiles
# to `multideref($h{"a"}) sK/EXISTS` and `delete $h{a}` to
# `multideref($h{"a"}) sK/DELETE`: both survive only as a FLAG on an op
# named after something else. A tier derived from op names alone would
# contain no notion of `exists` at all, and `10_hash_slice.t` would be the
# only evidence `delete` exists -- for the wrong reason, since the slice is
# the spelling where the optimiser DECLINES to fold and a real `delete` op
# appears. This file is the common path; that one is the exception.
#
# Both truth values of `exists` are here because the false one is a
# different claim. Perl's false is the empty string, so `print exists
# $h{z}` prints NOTHING and the pinned output carries a blank line -- a
# file that tested only the true case could not tell `exists` from a
# construct that always yields 1.
#
# `print exists $h{a}` rather than `print exists $h{a} ? 1 : 0`: measured,
# the conditional adds `cond_expr` and `0+(...)` adds `add`, and both are
# ops later tiers introduce. The plain print is the only spelling of this
# construct that stays inside the tier.
#
# `delete` in scalar context returns the value removed, which is what makes
# it observable without printing the hash: hash order is not guaranteed, so
# the count is what the file pins alongside it.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my %h = (a => 1, b => 2); print exists $h{a}, "\n"; print exists $h{z}, "\n"'
#   1
#
#   $ perl -e 'my %h = (a => 1, b => 2); my $gone = delete $h{a}; print $gone, "\n"; print scalar(keys %h), "\n"'
#   1
#   1

--- source
my %h = (a => 1, b => 2);
print exists $h{a}, "\n";
print exists $h{z}, "\n";
my $gone = delete $h{a};
print $gone, "\n";
print scalar(keys %h), "\n";

--- expect output
1

1
1

--- expect parses
