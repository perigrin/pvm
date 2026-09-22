#!perl
# A `pack` template is an opaque region spelled as a string: `C3` is a
# count and a type code to `pack` and three ordinary characters to the
# lexer, which reads it as a string literal and never looks inside.
#
# TIER 13 opaque
# INTRODUCES the pack template
# USES nothing from a later tier
#
# THE FOLDING TRAP, and it is why this file's argument comes from the
# environment. Measured, `pack` over constant arguments is not an op at
# all:
#
#   $ perl -MO=Concise,-exec -e 'print unpack("A3", pack("A3","abc"))'
#   1  <0> enter v
#   2  <;> nextstate(main 1 -e:1) v:{
#   3  <0> pushmark s
#   4  <$> const[PV "A3"] s
#   5  <$> const[PV "abc"] s/FOLD
#   6  <@> unpack lK/2
#   7  <@> print vK
#   8  <@> leave[1 ref] vKP/REFC
#
# There is NO `pack` op there. It ran at compile time and its result is
# the `s/FOLD` const the `unpack` reads. A file written that way would
# assert `pack` in its INTRODUCES block and emit no `pack`, so the lint
# would have nothing to check and a compiler that had never heard of
# `pack` would pass it. `$ENV{X}` is unset when the runner executes, so
# `//` yields the default and the value is a runtime one all the same.
#
# THE WRONG PARSE THIS RULES OUT: a lexer that reads `C3` -- or `A3`, or
# `x2` -- as anything but three characters of string. A template is the
# one argument shape in this tier that LOOKS like it wants lexing: `C3`
# is a count and a type, `A3 x N` is three fields, and the temptation to
# give the template its own scanner is exactly the temptation the format
# body's picture lines present. The token fact says it is one string
# literal, whole, with `C` and `3` never separated.
#
# `C3` rather than `A3`, because `A` pads with SPACES: `pack("A5","abc")`
# is `abc  `, and pinning that output would put trailing whitespace in
# this file for the repository's `trailing-whitespace` hook to strip. `C`
# takes ordinals and produces printable letters, and 65, 66, 67 are `A`,
# `B`, `C` in every encoding perl builds against.
#
# `length` would be the natural way to observe a packed string and NO
# TIER CLAIMS IT, so the output is the packed bytes themselves.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $n = $ENV{X} // 65; print pack("C3", $n, $n + 1, $n + 2), "\n";'
#   ABC
#
# and its optree, which does contain a `pack`:
#
#   ...
#   a  <$> const[PV "C3"] s
#   b  <0> padsv[$n:1,2] s
#   ...
#   i  <@> pack[t5] sK/2

--- source
my $n = $ENV{X} // 65;
print pack("C3", $n, $n + 1, $n + 2), "\n";

--- expect output
ABC

--- expect parses

--- expect tokens
one string literal whose text is "\"C3\""
no numeric literal whose text is "3"
