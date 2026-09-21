#!perl
# The embedded regions hold content whose delimiter appears inside a
# nested quote: a reader that counts characters stops early, a reader
# that re-enters Perl does not.
#
# TIER 14 recursive
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# The tier's other four files hold INERT regions. `$n+1`, `$k = 5` and
# `$n = 7` are delimited identically by a lexer that counts braces and by
# one that parses Perl, so those files establish that the region is
# DELIMITED and leave the RE-ENTRY half of the tier's thesis unasserted.
# This file is the inverse of 13_opaque/11_data_not_perl.t: there the
# hostile content proved a region was NOT lexed, here it proves a region
# IS.
#
# MEASURED, AND THE TWO CONSTRUCTS DIFFER. Under 5.42.0 perl uses two
# different strategies to find the end of a re-entrant region, and the
# difference is observable:
#
#   - `(?{ ... })` is PARSED AS PERL. Measured, `/b(?{ $k =
#     length("}}}") })/` compiles and prints 3 -- the three braces inside
#     the string do NOT close the block. A brace counter would stop at
#     the first and hand the parser `(?{ $k = length("}`.
#   - `s{a}{ ... }e` COUNTS DELIMITERS. Measured, `s{a}{ $n +
#     length("}}") }e` is a syntax error in perl itself -- "Unmatched
#     right curly bracket at -e line 1" -- because the brace in the
#     string DOES close the replacement. The re-entry happens after the
#     region is cut out, not while it is being found.
#
# So the substitution's hostile content is PARENS rather than braces:
# `s{a}{ $n + length("))") }e` compiles and prints 3bc, and a reader that
# balanced some bracket other than its own delimiter would truncate it.
# That is the strongest falsifiable claim the construct admits, and
# writing braces there instead would pin a perl syntax error as though it
# were our bug.
#
# The braced `s{}{}e` spelling rather than `s///e` is forced by the same
# measurement from the other side: with a `/` delimiter, a `/` inside a
# string in the replacement is a syntax error in perl too. A delimiter
# that cannot appear inside the region leaves no hostile content to write.
#
# Every replacement here is unfoldable, as the tier's README requires:
# `$n + length("))")` keeps `substcont`, where a constant replacement
# would fold to `const s/FOLD` and erase the `/e` entirely. Measured,
# `length("))")` itself folds to `const[IV 2]`, but the `add` around it
# does not, so the frame survives.
#
# No `use re 'eval'` and no warnings suppression. The blocks are LITERAL,
# not interpolated from a variable, which is what the pragma gates.
#
# The token claims are where this file carries its weight, because the
# optree cannot see the distinction at all: measured, the `(?{ })` and
# the `qr//` block are inside their ops' PATTERN STRINGS, one `match` and
# one `qr`. The `no` claims name the hostile characters -- a lexer that
# stopped early would have spilled the region's tail out as ordinary
# tokens, and `length` would appear as a word beside the quote rather
# than inside it.
#
# MEASURED perl 5.42.0:
#
#   $ perl hostile_contents.pl | od -c
#   0000000   3   b   c       5  \n
#   0000006

--- source
my $s = "abc";
my $n = 1;
my $k = 0;
$s =~ s{a}{ $n + length("))") }e;
$s =~ /b(?{ $k = length("}}}") })/;
my $r = qr/c(?{ $k = $k + length("}}") })/;
$s =~ $r;
print "$s $k\n";

--- expect output
3bc 5

--- expect parses

--- expect tokens
one quote-like operator whose text is "s{a}{ $n + length(\"))\") }e"
one quote-like operator whose text is "qr/c(?{ $k = $k + length(\"}}\") })/"
no word whose text is "length"
no operator whose text is "+"
