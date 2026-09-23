#!perl
# `$x += 2` is ONE operator token, not `+` and `=`, and no corpus file
# said so.
#
# TIER 04 operators
# INTRODUCES nothing of its own
# USES the arithmetic operators from 01_arithmetic.t
#
# Thirteen compound assignment operators, and before this file the corpus
# wrote none of them. perlop's level 20 is the second-largest level in the
# table; the corpus made no claim about any of it. Plain `=` appears in
# roughly 155 files binding a literal, which is scaffolding rather than a
# claim by this corpus's own rule.
#
# WHAT A PARSER GETS WRONG HERE, AND WHAT CATCHES IT. `$x += 2` and
# `$x = $x + 2` compute the same value, so OUTPUT cannot separate them.
# They are different trees, and this corpus does not assert trees. What
# separates them is the TOKEN STREAM, measured against our own lexer:
#
#   $x += 1        Variable($x) Operator(+=) Number(1)
#   $x = $x + 1    Variable($x) Operator(=) Variable($x)
#                  Operator(+) Number(1)
#
# One token against three. A lexer that split `+=` into `+` and `=` would
# produce a stream the parser could still make sense of -- and a wrong
# tree with correct output. That is precisely the case the token layer
# exists for, and why this file's claim is a count rather than a value.
#
# MEASURED perl 5.42.0. The ops are the BINARY ones, reused:
#
#   $x += 2   ->  add
#   $x -= 2   ->  subtract
#   $x *= 2   ->  multiply
#   $x /= 2   ->  divide
#
# So this file introduces nothing: every op it emits is already tier
# 04's from `01_arithmetic.t`. The whole of its content is the four token
# facts. `22_compound_string.t`, `23_compound_shortcircuit.t` and
# `24_compound_bitwise.t` are the rest of the family, split where the
# ops or the observability differ.
#
#   $ perl -e 'my $a=5; my $b=5; my $c=5; my $d=6;
#              $a+=2; $b-=2; $c*=2; $d/=2; print "$a $b $c $d"'
#   7 3 10 3

--- source
my $a = $ENV{X} // 5;
my $b = $ENV{X} // 5;
my $c = $ENV{X} // 5;
my $d = $ENV{Y} // 6;
$a += 2;
$b -= 2;
$c *= 2;
$d /= 2;
print "$a $b $c $d\n";

--- expect output
7 3 10 3

--- expect parses

--- expect tokens
one operator whose text is "+="
one operator whose text is "-="
one operator whose text is "*="
one operator whose text is "/="
