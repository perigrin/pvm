#!perl
# The six string relations and `cmp`: seq, sne, slt, sgt, sle, sge and
# scmp. Same two operands as `03_numeric_comparison.t`, same relation,
# different ops -- and the difference is only the context each imposes.
#
# TIER 04 operators
# INTRODUCES string comparison operators
# USES nothing from a later tier
# STATUS refuses as of this file.
#
# The refusal is the token claim. Every operator in this file is spelled
# with letters, and our lexer files all seven under `Word` where the
# glossary calls them operators. `02_string.t` has the same refusal for
# `x`; this file is where it is densest, since the whole construct is
# word-spelled. See that file for why the claim is left failing.
#
# This file and `03_numeric_comparison.t` are the tier's argument for
# depending on 03. `$a == $b` and `$a eq $b` compile to two different ops
# over identical operands, so a corpus that had not yet named numeric
# versus string evaluation could not say why there are two.
#
# The `s` prefix belongs to the STRING ops: `eq` is the numeric `==` and
# `seq` is the string `eq`. See the sibling file.
#
# Results are bracketed for the same reason as the sibling: false is the
# empty string, and a line ending in a space does not survive the
# `trailing-whitespace` pre-commit hook.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $a = "aa"; my $b = "bb"; print "[", ($a eq $b), "][", ($a cmp $b), "]\n"'
#   [][-1]

--- source
my $a = $ARGV[0] // "aa";
my $b = $ARGV[1] // "bb";
print "seq [", ($a eq $b), "]\n";
print "sne [", ($a ne $b), "]\n";
print "slt [", ($a lt $b), "]\n";
print "sgt [", ($a gt $b), "]\n";
print "sle [", ($a le $b), "]\n";
print "sge [", ($a ge $b), "]\n";
print "scmp [", ($a cmp $b), "]\n";

--- expect output
seq []
sne [1]
slt [1]
sgt []
sle [1]
sge []
scmp [-1]

--- expect parses

--- expect tokens
one operator whose text is "cmp"
