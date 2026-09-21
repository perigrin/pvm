#!perl
# Every construct this tier introduces, in one body, each adjacent to
# another.
#
# TIER 03 context
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# The tier's other files are one construct each, which makes them
# diagnosable: when `05_sort.t` refuses, the construct that refused is the
# only one present. That property is also why such files cannot reach an
# adjacency bug -- a parser that handles every construct alone and
# mis-handles a pair goes green over the pair.
#
# Here the constructs sit on consecutive statements: an array in scalar
# context, the same array interpolated, the comma operator in scalar
# context, `sort`, `reverse`, `localtime` through the count idiom, and
# `wantarray`. They are adjacent as statements rather than nested, because
# nesting them would need an operator to join them and this tier is before
# the operator tier.
#
# This tier depends on 02_variables, and the adjacency with the earlier
# tier is the first line: `@a` is tier 02's array, and every statement
# after it puts that same array in a different context. Pairing this tier
# with the one it depends on is the array being re-used, not a separate
# statement about arrays.
#
# `sort` is a default string sort, so `(3,1,2)` comes back `1 2 3`. The
# reverse of that array in scalar context is `213` -- the elements joined
# with nothing between them, then the string reversed -- not `2 1 3`.
#
# `$fields` is 9 because `localtime` in list context returns nine
# elements; the count idiom is what puts it in list context while still
# yielding a scalar to print. Pinning the elements themselves would make
# this file depend on the clock.
#
# Fewer ops appear than the constructs suggest. `padrange` fuses
# consecutive `my` declarations, and the comma operator in scalar context
# erases its own left operands -- so the op set is a UNION across the
# tier's files rather than a property of this one. It happens that this
# file does emit all six, but that is a fact about this program, not a
# rule the format requires.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/03_context/00_adjacency.t
#   3 3 1 2 6 1 2 3 213 9 1

--- source
my @a = (3, 1, 2);
my $count = @a;
my $interp = "@a";
my $last = (4, 5, 6);
my @sorted = sort @a;
my $rev = reverse @a;
my $fields = () = localtime;
my $want = wantarray;
my @seen = ($want);
print "$count $interp $last @sorted $rev $fields ", scalar(@seen), "\n";

--- expect output
3 3 1 2 6 1 2 3 213 9 1

--- expect parses
