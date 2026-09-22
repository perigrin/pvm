#!perl
# Every construct this tier introduces, in one body, each adjacent to
# another -- and the one file in the corpus written because the bug it
# catches was already known.
#
# TIER 11 oo
# INTRODUCES nothing of its own
# USES nothing from a later tier
# STATUS refuses as of 6c231691. Refusal trailing_tokens. Issue 01a0c35f-a487-7aa8-934b-4e68dd0fbaa1.
#
# The code names WHICH SITE declines, and the issue names which bug
# somebody believed this was. They are different promises and the file
# makes both. Measured, the Unknown starts at the `ADJUST` keyword and
# runs to the end of the `method` after it -- the expression parser reads
# the ADJUST block, then finds the `method` before the terminator, which
# is `trailing_tokens`. Without the code, a refusal that drifted to some
# other site would leave this file reading as though the ADJUST bug were
# still what it measured, and the citation would be stale.
#
# MEASURED against our parser, and this is the case the whole
# adjacency-file design was written for:
#
#   class Foo { ADJUST { 1 } }                   parses, 0 Unknowns
#   class Foo { ADJUST { 1 } method m { 2 } }    1 Unknown, swallowing both
#
# `ADJUST` alone parses -- that is `09_class_adjust.t`, and it is green.
# `ADJUST` followed by anything -- `method`, `field`, or a second `ADJUST`
# -- does not. A corpus of one construct per file goes green over this by
# CONSTRUCTION, because every construct in it is measured in isolation and
# every construct in isolation passes. Only a file that puts them next to
# each other can see it, which is why each tier carries one.
#
# Here that is: a `class` with a `field`, an `ADJUST` and a `method` in
# one body; a classic `bless` of an empty anon hash; a compile-time method
# call and a dynamic one on the result, an infix `isa` and a `can` chain.
# The class side and the classic side are adjacent to each other as well,
# because a parser that switches modes on `class` has to switch back.
#
# `$yes` and `$code` are the argument-extent slice's two constructs
# (issue 01a0c730), adjacent to the dispatch they are not:
#
#   - `$b isa Bar` is an INFIX OPERATOR. Measured, it emits `<2> isa`
#     with no `entersub` and no `method_named` at all, so it sits in this
#     body beside four method calls that emit both, and a parser that
#     read it as a fifth would be caught by the mixture and by nothing
#     else. `10_isa_infix.t` is the one-construct half. It is what the
#     `'isa'` in the feature list above is for: without it the infix
#     spelling is a SYNTAX ERROR rather than a weaker parse.
#
#   - `$b->can("hi")->($b)` is TWO calls through one chain and only the
#     first is a method call; the second arrow dereferences the code ref
#     `can` returned. Measured, two `entersub` under one `method_named`.
#     `11_can_chain.t` is the one-construct half, and this body is where
#     that chain stands next to the plain `$b->hi` and the dynamic
#     `$b->$name` it must not be confused with.
#
# The tier's declared prerequisite is `08_references`, and `bless` is the
# whole of it: the blessed hash below is a tier 08 reference plus a
# string. Pairing with tier 10 instead would assert nothing -- this tier
# uses no file handle.
#
# MEASURED perl 5.42.0, which accepts all of it:
#
#   $ perl conformance/11_oo/00_adjacency.t
#   Foo2Barbarbar1bar
#
# `Foo2` is `ref($c)` then `$c->m`, which ADJUST raised from 1 to 2;
# `Barbarbar` is `ref($b)` then the same method reached two ways. Nothing
# prints a raw object: `Bar=HASH(0x...)` carries an address that changes
# every run, so `ref` is what the file asserts on.
#
# `expect output` is written before `expect parses` rather than last. The
# blank line after it carries the output's own trailing newline, and a
# blank line at END of file is what `end-of-file-fixer` strips.

--- source
use feature 'class', 'isa';
no warnings 'experimental::class';
class Foo {
    field $x = 1;
    ADJUST { $x = 2 }
    method m { $x }
}
package Bar;
sub hi { return "bar" }
package main;
my $c = Foo->new;
my $b = bless {}, "Bar";
my $name = "hi";
my $yes = $b isa Bar;
my $code = $b->can("hi")->($b);
print ref($c), $c->m, ref($b), $b->hi, $b->$name, $yes, $code, "\n";

--- expect output
Foo2Barbarbar1bar

--- expect parses
