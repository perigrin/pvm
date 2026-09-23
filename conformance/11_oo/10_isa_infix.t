#!perl
# `isa` is an INFIX OPERATOR, not a method. `$o isa Foo` emits a binary
# `isa` op with no call machinery at all -- no `entersub`, no
# `method_named` -- where `$o->isa("Foo")` emits both.
#
# TIER 11 oo
# INTRODUCES the infix isa operator
# USES my, print, bless, package, %ENV, //, string interpolation
# MEASURED perl 5.42.0
#
# 29 of T1's 986 files use the word and the corpus named neither
# spelling. The two are not one construct with two syntaxes: they are
# two unrelated parses that happen to answer the same question, and the
# whole tier is about dispatch, so a tier that names only the method
# spelling has named the half that is NOT distinctive.
#
# MEASURED perl 5.42.0, the two spellings in one program:
#
#   my $infix  = $o isa Foo;      a  <2> isa sK/2
#   my $method = $o->isa("Foo");  h  <.> method_named[PV "isa"] s
#                                 i  <1> entersub[t5] sKRS/TARG,STRICT
#
# The infix form is a BINARY OPERATOR, arity 2, and dispatch never
# happens. The method form goes through the same `method_named` and
# `entersub` every other call in this tier uses. One word, two parses,
# and only one of them is this tier's dispatch machinery.
#
# IT IS FEATURE-GATED, which is the second half of its parse and the
# reason this file opens with `use v5.36`. Measured, without the feature
# the infix spelling is not a weaker parse but a SYNTAX ERROR:
#
#   $ perl -e 'package Foo; ... my $o=Foo->new; print(($o isa Foo)?1:0)'
#   syntax error at -e line 1, near "$o isa Foo"
#
# So a lexer cannot decide what `isa` is from the token alone; whether
# the word is an operator at all depends on a pragma earlier in the
# file. That is the same shape as tier 07's finding about prototypes --
# the call site does not carry what decides its parse -- met here through
# a feature gate rather than a declaration.
#
# It is NO LONGER EXPERIMENTAL. Measured under 5.42.0, `use v5.36` with
# `$o isa Foo` emits no warning, so this file needs no `no warnings`
# line where `00_adjacency.t` needs one for `class`.
#
# THE OUTPUT IS THE FALSIFYING HALF AND THE FALSE CASE IS WHAT MAKES IT
# ONE. Measured:
#
#   $ perl conformance/11_oo/10_isa_infix.t
#   1|
#
# `$yes` is 1 and `$no` is the EMPTY STRING -- perl's false, which
# interpolates to nothing -- so the two run together as `1` and the `|`
# is what proves a second value was printed at all. A parser that read
# `isa` as always true prints `11|`; one that read it as always false
# prints `|`. Both wrong readings are one byte from the right one, and
# without the false case neither would be.
#
# `bless {}, $ENV{X} // "Foo"` rather than `Foo->new`, for two reasons
# and both matter. The `//` keeps the class name a RUNTIME value, so
# nothing folds -- `X` is unset when the runner executes the file. And
# `Foo->new` would put an ARROW in this source, which is exactly what
# the token fact below says is absent.
#
# THE TOKEN FACTS ARE WHAT SEPARATE THE TWO SPELLINGS, because nothing
# else in this file can. `$o isa Foo` and `$o->isa("Foo")` both print
# `1`, so behaviour cannot tell them apart -- the same limit
# `06_indirect_new.t` records for `new Foo` against `Foo->new`, and the
# reason this tier asserts on tokens more than any other.
#
#   - NO `->` ANYWHERE. The method spelling cannot be written without
#     one, so a parser that quietly rewrote the infix form into a method
#     call -- the obvious way to "handle" a word it does not know as an
#     operator -- introduces an arrow this file says is absent. This is
#     the negative that `03_method_named.t`'s positive makes
#     non-vacuous: a lexer that never emitted `->` would pass here and
#     fail there.
#
#   - NO `(` ANYWHERE. `$o->isa("Foo")` cannot be written without one,
#     and neither can any other parenthesised call, so this is the
#     second half of the same guard as the arrow: both characters the
#     method spelling requires are declared absent, and a rewrite has to
#     produce at least one of them.
#
#   - NO STRING SPELLED `"Bar"` in operator position. In the infix form
#     the right operand is a BAREWORD; in the method form it is a quoted
#     string argument. Our lexer gives those different kinds -- `Word`
#     against `Quote` -- so a rewrite into the method spelling has to
#     produce a string literal, and this fact is what catches it. It is
#     asserted on `Bar` and not on `Foo` because `"Foo"` IS present, as
#     `bless`'s second argument, where it is a string in both spellings
#     and says nothing at all. `Bar` reaches the token stream only as a
#     bareword operand, so a quoted one there could only have come from
#     a rewrite.
#
# WHAT THESE FACTS CANNOT SAY is how many `isa`s there are. There are
# TWO -- the true case and the false one -- and the grammar admits only
# `one` and `no`, so the count that would name the operator directly is
# not expressible. `GLOSSARY.md` records that limit; the three facts
# above are what this source can claim that a broken lex fails.

--- source
use v5.36;
package Foo;
package Bar;
package main;
my $o = bless {}, $ENV{X} // "Foo";
my $yes = $o isa Foo;
my $no = $o isa Bar;
print "$yes$no|\n";

--- expect output
1|

--- expect parses

--- expect tokens
no operator whose text is "->"
one word whose text is "main"
one string literal whose text is "\"Foo\""
