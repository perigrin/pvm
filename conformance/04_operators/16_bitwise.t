#!perl
# `&`, `|` and `^` are BINARY OPERATORS, and this tier had none of them.
#
# TIER 04 operators
# INTRODUCES the bitwise operators
# USES nothing from a later tier
#
# perlop's levels 14 and 15 had no coverage at all. Measured across the
# corpus source before this file, every `&` was the CODE SIGIL -- `\&foo`,
# `&$code` -- and every `|` was either the `//` defined-or or a regex
# alternation. `^` appeared nowhere outside a regex.
#
# That is a lexical hazard, not merely a missing operator: the same three
# characters mean something else in the places the corpus already used
# them, so a lexer that only ever saw `\&foo` could read `$a & $b` as a
# code reference and pass every file in the tree.
#
# MEASURED perl 5.42.0. Both operands must be RUNTIME values or the
# operator does not survive to be measured:
#
#   $ perl -MO=Concise,-exec -e 'print 12 & 10'
#   ... const[IV 8] ... print ...          NO bit_and op
#
#   $ perl -MO=Concise,-exec -e 'my $a = $ENV{X} // 12; print $a & 10'
#   ... bit_and ... print ...
#
# The same trap this tier's README records for arithmetic: the optimiser
# erases the construct the file is about. `$ENV{X}` is unset when the
# runner executes, so the `//` default is what arrives, at run time.
#
#   $ perl -e 'my $a = $ENV{X} // 12; print $a & 10, " ", $a | 3, " ", $a ^ 3'
#   8 15 15
#
# `12 & 10` is 8, `12 | 3` is 15 and `12 ^ 3` is 15 -- the last two agree
# by arithmetic accident on these operands, which is why the file prints
# all three rather than one: a parser that confused `|` with `^` would
# still print 15 twice, and the `&` is what separates them.

--- source
my $a = $ENV{X} // 12;
print $a & 10, " ", $a | 3, " ", $a ^ 3, "\n";

--- expect output
8 15 15

--- expect parses

--- expect tokens
one operator whose text is "&"
one operator whose text is "|"
one operator whose text is "^"
