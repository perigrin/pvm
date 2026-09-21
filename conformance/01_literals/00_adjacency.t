#!perl
# Every construct this tier introduces, in one body, each adjacent to
# another.
#
# TIER 01 literals
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# The tier's other files are one construct each, which is what makes them
# diagnosable: when `03_leading_decimal.t` refuses, the construct that
# refused is the only one present. That same property is why a corpus of
# such files cannot reach an ADJACENCY bug -- a parser that handles every
# construct alone and mis-handles a pair goes green over the pair.
#
# MEASURED perl 5.42.0, and this is not hypothetical:
#
#   class Foo { ADJUST { 1 } }                   0 Unknowns
#   class Foo { ADJUST { 1 } method m { 2 } }    1 Unknown, swallowing both
#
# `ADJUST` alone parses; `ADJUST` followed by anything does not. No
# one-construct-per-file corpus can ever see that, because every file is
# one construct by definition.
#
# So each tier carries one file where its constructs sit next to each
# other. Here that is an integer, a decimal, a single-quoted string, a
# double-quoted string with interpolation, and a `qw` list -- adjacent
# within one statement where the construct allows it, and on consecutive
# statements where it does not.
#
# This tier depends on nothing, so there is no earlier tier to pair with;
# the adjacency is entirely within 01.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/01_literals/00_adjacency.t
#   42 0.5 plain 42-0.5 abc
#
# `qw(a b c)` prints as `abc` rather than `a b c`: in a print LIST the
# three words are separate arguments and $, is unset, so nothing separates
# them. Binding them to an array instead would print `a b c` -- but it
# would also emit aassign, padav and join, which are tier 02's array
# machinery and unclaimed here. The lint caught that, and this is the
# version that keeps `qw` in the tier that owns it.
#
# `expect output` is written before `expect parses` rather than last. The
# blank line after it is what carries the output's own trailing newline,
# and a blank line at END of file is what `end-of-file-fixer` strips. A
# section after it puts the blank line in the middle of the file, where
# the hook has no quarrel with it.

--- source
my $int = 42;
my $dec = 0.5;
my $sq = 'plain';
my $dq = "$int-$dec";
print "$int $dec $sq $dq ", qw(a b c), "\n";

--- expect output
42 0.5 plain 42-0.5 abc

--- expect parses
