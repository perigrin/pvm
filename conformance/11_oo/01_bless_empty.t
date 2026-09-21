#!perl
# `bless {}, $c` blesses an EMPTY anonymous hash, and the optimiser has a
# dedicated op for that case: `emptyavhv`, not `anonhash`.
#
# TIER 11 oo
# INTRODUCES bless, and the empty anonymous hash it is given
# USES nothing from a later tier
#
# This is the file that looks simplest and carries the op nobody expects.
# `02_bless_populated.t` is the same construct with contents in the hash,
# and it emits `pushmark`, `padhv`, `anonhash` instead -- three ops for
# what reads as the same thing. A parser that learns the populated form
# learns nothing about this one.
#
# `ref $o` rather than `$o` itself: printing the object gives
# `Foo=HASH(0x55d3...)`, whose address changes every run. `ref` gives the
# stable string the blessing actually installed, which is the thing this
# file is about. `ref` is tier 08's op.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $c = "Foo"; my $o = bless {}, $c; print ref $o, "\n"'
#   Foo
#
# and the op stream, where the hash never becomes an `anonhash`:
#
#   6  <0> emptyavhv[t3] s/ANONHASH
#   7  <0> padsv[$c:1,3] s
#   8  <@> bless sK/2

--- source
my $c = "Foo";
my $o = bless {}, $c;
print ref $o, "\n";

--- expect output
Foo

--- expect parses
