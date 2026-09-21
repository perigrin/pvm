#!perl
# Every construct this tier introduces, in one body, each adjacent to
# another.
#
# TIER 08 references
# INTRODUCES nothing of its own
# USES sub, my, print, ref
#
# The tier's other files are one construct each, which is what makes them
# diagnosable: when `05_brace_deref.t` refuses, the construct that refused
# is the only one present. That same property is why such a corpus cannot
# reach an ADJACENCY bug -- a parser that handles every construct alone
# and mis-handles a pair goes green over the pair.
#
# Here the constructs are the backslash in both contexts, the anonymous
# array, the arrow, the three brace derefs, the code reference, and `ref`
# -- adjacent within one statement where the construct allows it, and on
# consecutive statements where it does not. The print statement puts six
# of them side by side in a single list, which is where a parser that
# handles `$$s[0]` and `${$l}[1]` separately but not next to each other
# would show it.
#
# The tier's declared prerequisite is 07_subroutines, so the adjacency
# pairs with it: `sub twice` is tier 07's, `\&twice` is this tier's, and
# `$c->(21)` is the arrow applied to the result -- the two tiers touching
# in one expression rather than in two separate files.
#
# `\(@a)` is the one that needs care. It distributes: `@r` holds a list
# of SCALAR references, one per element of `@a`, so `${$r[1]}` is 20 and
# `$r[1]->[0]` dies with "Not an ARRAY reference". That failure is what
# this file's first draft printed.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/08_references/00_adjacency.t
#   ARRAY 10 20 30 40 42

--- source
sub twice { return $_[0] * 2 }
my @a = (10, 20);
my $s = \@a;
my @r = \(@a);
my $l = [30, 40];
my $c = \&twice;
print ref($s), " ", $$s[0], " ", ${$r[1]}, " ", $l->[0], " ", ${$l}[1], " ", $c->(21), "\n";

--- expect output
ARRAY 10 20 30 40 42

--- expect parses
