#!perl
# The five logical operators: and, or, dor, not and xor. `&&` is the op
# `and`, `||` is the op `or`, `//` is the op `dor`.
#
# TIER 04 operators
# INTRODUCES logical operators
# USES nothing from a later tier
# STATUS refuses as of this file.
#
# Two refusals, measured. The parser leaves one Unknown node over this
# body. And our lexer reads `xor` as `Word("xor")` where the glossary
# calls it an operator -- the same word-operator gap `02_string.t` and
# `04_string_comparison.t` record, here on a word whose symbolic twin
# (`^^`) perl does not spell, so `xor` is the ONLY way to write it and the
# gap cannot be sidestepped.
#
# `and`, `or` and `dor` short-circuit and so return one of their
# OPERANDS, not a boolean: `$a || $b` with `$a` true is `$a`. `not` and
# `xor` do not short-circuit and return a real boolean, so `not 1` is the
# empty string and `1 xor 0` is 1.
#
# The defaults are chosen so every branch is exercised without arguments:
# `$a` true and `$b` false makes `and` return the false `$b`, `or` return
# the true `$a`, and `xor` true.
#
# `dor` is the op `//` compiles to, and every file in this tier already
# emits it as part of its `$ARGV[0] // default` operand. It is claimed
# here rather than there because this is the file that measures it as a
# construct rather than as scaffolding.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $a = 1; my $b = 0; print "[", ($a && $b), "][", ($a xor $b), "]\n"'
#   [0][1]

--- source
my $a = $ARGV[0] // 1;
my $b = $ARGV[1] // 0;
my $and = ($a && $b);
my $or = ($a || $b);
my $dor = ($b // $a);
my $not = (not $a);
my $xor = ($a xor $b);
print "and [$and] or [$or] dor [$dor] not [$not] xor [$xor]\n";

--- expect output
and [0] or [1] dor [0] not [] xor [1]

--- expect parses

--- expect tokens
one operator whose text is "xor"
one operator whose text is "||"
