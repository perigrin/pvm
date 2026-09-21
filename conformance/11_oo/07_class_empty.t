#!perl
# An EMPTY `class` body compiles to a bare block containing `stub`.
#
# TIER 11 oo
# INTRODUCES an empty class declaration
# USES nothing from a later tier
#
# `class Empty {}` is a complete, working class -- `Empty->new` returns an
# object -- and the four keywords that make it so produce, between them,
# tier 05's `enterloop` and `leaveloop` plus one `stub`. There is no
# `class` op. The constructor perl generates is XS code with no Perl
# optree, so nothing this tier introduces can be measured through it
# either; `ref` on the result is what shows the class exists.
#
# MEASURED perl 5.42.0. Neither `use feature 'class'` nor `no warnings`
# emits a runtime op -- the stream below begins at the class body on line
# 3 -- so the pragmas the syntax requires cost this tier nothing and do
# not borrow tier 12's `use`.
#
#   $ perl -e 'use feature "class"; no warnings "experimental::class"; class Empty {} print ref(Empty->new), "\n"'
#   Empty
#
#   1  <0> enter v
#   2  <;> nextstate(main 3 d.pl:3) v:%,{,fea=15
#   3  <{> enterloop(next->5 last->5 redo->4) v
#   4  <0> stub v
#   5  <2> leaveloop vK/2

--- source
use feature 'class';
no warnings 'experimental::class';
class Empty {}
print ref(Empty->new), "\n";

--- expect output
Empty

--- expect parses
