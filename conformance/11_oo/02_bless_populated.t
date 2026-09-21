#!perl
# `bless { %a }, $c` builds a real anonymous hash first: `pushmark`,
# `padhv`, `anonhash`, and only then `bless`.
#
# TIER 11 oo
# INTRODUCES the populated anonymous hash a constructor blesses
# USES nothing from a later tier
#
# The pair with `01_bless_empty.t` is the point. One construct in the
# source -- an anonymous hash handed to `bless` -- compiles to two
# unrelated op streams depending on whether the braces are empty. The
# corpus has to carry both or it measures half of `bless`.
#
# `%a` is a tier 02 hash and `aassign`, `padhv` are tier 02's ops; what
# this file introduces is `anonhash` and the `bless` around it.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my %a = (n => 1); my $o = bless { %a }, "Foo"; print ref $o, "\n"'
#   Foo
#
# and the op stream, which has no `emptyavhv` anywhere:
#
#   d  <0> pushmark s
#   e  <0> padhv[%a:1,4] l
#   f  <@> anonhash sK*/1
#   g  <0> padsv[$c:2,4] s
#   h  <@> bless sK/2

--- source
my %a = (n => 1);
my $c = "Foo";
my $o = bless { %a }, $c;
print ref $o, "\n";

--- expect output
Foo

--- expect parses
