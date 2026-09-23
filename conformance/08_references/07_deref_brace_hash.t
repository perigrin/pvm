#!perl
# `%{$r}` dereferences a scalar as a whole hash, and `${$r}{a}` reaches
# one of its values -- both without any anonymous hash constructor.
#
# TIER 08 references
# INTRODUCES the %{ } and ${ }{ } dereferences
# USES my, print, scalar, keys, the reference operator
#
# This is the tier's other hard marker, `deref-brace`.
#
# The hash is a NAMED one taken a reference to, not `{ a => 1 }`. The
# anonymous hash constructor compiles to `emptyavhv`/`anonhash`, which
# tier 11 claims where objects are built, and a tier-08 file emitting a
# tier-11 op is exactly what the lint refuses. `\%h` reaches the same
# place through `srefgen`, which this tier owns.
#
# `scalar(keys %copy)` emits no `keys` op: measured, the optimiser fuses
# it into `padhv[%copy] sM/KEYS`. Another instance of more construct,
# fewer ops.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my %h = (a => 1, b => 2); my $r = \%h; my %copy = %{$r}; print scalar(keys %copy), " ", ${$r}{a}, "\n"'
#   2 1

--- source
my %h = (a => 1, b => 2);
my $r = \%h;
my %copy = %{$r};
print scalar(keys %copy), " ", ${$r}{a}, "\n";

--- expect output
2 1

--- expect parses

--- expect tokens
no operator whose text is "->"
one operator whose text is "\\"
