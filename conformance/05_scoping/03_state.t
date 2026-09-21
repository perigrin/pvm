#!perl
# `state $n = 1` declares a lexical whose INITIALISATION runs once, and
# it scopes like `my`: an inner `state $n` shadows an outer one.
#
# TIER 05 scoping
# INTRODUCES state declaration
# USES nothing from a later tier
#
# `once` is this tier's only op that a single construct owns outright.
# `state $n = 1` compiles to `once(other->...)` wrapped around an
# otherwise ordinary `padsv_store[$n] vKS/LVINTRO,STATE` -- the pad op is
# tier 01's, and `once` is the whole of what `state` adds.
#
# WHAT THIS FILE CANNOT SHOW, and why that is not a gap in the file. The
# point of `state` is that the value persists across calls, which needs a
# sub to call twice. `sub` is tier 07 and `entersub` is not in this
# tier's budget; a bare block is the only re-enterable-looking construct
# available and it is not re-enterable -- it runs exactly once, so `once`
# firing once is indistinguishable from an ordinary initialisation. The
# behavioural probe for persistence therefore belongs to tier 07, and
# what this file asserts instead is the part that IS observable here: the
# declaration compiles, and its scoping is lexical.
#
# The `use feature "state"` line is a tier 12 construct in a tier 05
# file, and it is affordable because it emits NO RUNTIME OP -- measured,
# not assumed. Under `-MO=Concise,-exec` the pragma leaves nothing in the
# op stream at all; it only flips a bit in the `nextstate` hints field
# (`v:%,{,fea=15`), and the lint reads op names, not hints. `use v5.36`
# would serve as well and was measured too: same ops, a different hints
# value. The explicit feature import is preferred because it names the
# one thing this file needs rather than dragging in a bundle.
#
# MEASURED perl 5.42.0:
#
#   $ perl -w -e 'use feature "state"; state $n = 1; { state $n = 2; print "$n\n"; } print "$n\n"'
#   2
#   1

--- source
use feature "state";
state $n = 1;
{
  state $n = 2;
  print "$n\n";
}
print "$n\n";

--- expect output
2
1

--- expect parses
