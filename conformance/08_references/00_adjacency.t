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
# array, the arrow, all three whole-aggregate dereferences, the brace
# element derefs, the code reference, and `ref` -- adjacent within one
# statement where the construct allows it, and on consecutive statements
# where it does not. The two print statements put ten of them side by
# side, which is where a parser that handles `$$s[0]` and `${$l}[1]`
# separately but not next to each other would show it.
#
# THE THREE SPELLINGS OF ONE OP are all here on consecutive lines:
# `@{$s}`, `@$s` and `$s->@*` each emit `rv2av` and nothing in the op
# stream separates them, so a parser that reads two of the three and
# guesses at the third produces an identical optree for the wrong
# program. Three lines apart is where that guess has to hold.
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
# The hash is a NAMED one taken a reference to, not `{ k => 1 }`. The
# anonymous hash constructor compiles to `emptyavhv`/`anonhash`, which
# tier 11 claims, and a tier-08 file emitting a tier-11 op is what the
# dependency lint refuses.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/08_references/00_adjacency.t
#   ARRAY 10 20 30 40 42
#   10 20 10 1

--- source
sub twice { return $_[0] * 2 }
my @a = (10, 20);
my %h = (k => 1);
my $s = \@a;
my $hr = \%h;
my @r = \(@a);
my $l = [30, 40];
my $c = \&twice;
my @at = @{$s};
my @sig = @$s;
my @post = $s->@*;
my %copy = %{$hr};
print ref($s), " ", $$s[0], " ", ${$r[1]}, " ", $l->[0], " ", ${$l}[1], " ", $c->(21), "\n";
print $at[0], " ", $sig[1], " ", $post[0], " ", $copy{k}, "\n";

--- expect output
ARRAY 10 20 30 40 42
10 20 10 1

--- expect parses
