#!perl
# `ADJUST` runs after construction and emits no op of its own. ALONE in a
# class body, it also parses under our parser -- which is the half of the
# bug that makes `00_adjacency.t` necessary.
#
# TIER 11 oo
# INTRODUCES an ADJUST block
# USES nothing from a later tier
#
# `ADJUST` becomes an anonymous sub in the class stash with no CODE slot,
# so B::Concise cannot name it and its body is unreachable from any dump.
# What the main optree shows is a `nextstate` where the block was.
#
# This file holds ADJUST by itself on purpose. `00_adjacency.t` is the
# same construct with a `method` after it, and the pair is the whole
# demonstration: one construct per file goes green over an adjacency bug
# by construction, because every construct in such a corpus is measured
# alone.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'use feature "class"; no warnings "experimental::class"; class Foo { field $x = 1; ADJUST { $x = 2 } } print ref(Foo->new), "\n"'
#   Foo
#
# and the class body, ADJUST included, is three ops:
#
#   3  <{> enterloop(next->5 last->5 redo->4) v
#   4  <;> nextstate(Foo 5 f.pl:4) v:%,{,fea=15
#   5  <2> leaveloop vK/2

--- source
use feature 'class';
no warnings 'experimental::class';
class Foo {
    field $x = 1;
    ADJUST { $x = 2 }
}
print ref(Foo->new), "\n";

--- expect output
Foo

--- expect parses
