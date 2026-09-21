#!perl
# A format's picture line is NOT Perl. `@<<<<<<<` is a left-justified
# column, and a lexer reading it as Perl sees an array variable followed by
# three left-shift operators.
#
# TIER 13 opaque
# INTRODUCES the format picture field
# USES nothing from a later tier
# STATUS refuses as of 9750d03b. Cites this file. Refusal unimplemented_statement.
#
# 06_format_write.t deliberately writes a picture of plain text, because
# its subject is the `write` and a fixed line keeps the output byte-exact
# without introducing column padding. Its comment defers the field case to
# "a file that declares padding as its subject". This is that file, and the
# subject is narrower than padding: it is that the body is OPAQUE.
#
# WHY A PLAIN PICTURE LINE CANNOT MAKE THIS CLAIM, which is the whole
# reason this file exists beside 06. `a fixed report line` inside a format
# body is delimited identically by a lexer that treats the body as opaque
# and by one that lexes it as Perl -- three Words either way, and the
# region's extent is unchanged. The two readings are indistinguishable, so
# such a file measures DELIMITING and says nothing about NOT LEXING, which
# is the other half of this tier's thesis. `@<<<<<<< @>>>` separates them:
# opaque it is one token, lexed as Perl it is `@` variables and shift
# operators that would leave the body's bytes as ordinary tokens.
#
# Measured against our lexer, the body arrives as a single FormatBody token
# whose text runs from the first picture line through the lone `.`, with no
# token inside it. The parser produces one Unknown over the declaration --
# the same gap 06 records, and for the same reason: there is no grammar
# rule for `format NAME = BODY`. So the picture field is already opaque and
# what is missing is above the lexer.
#
# The argument line under the picture -- `$name,   $qty` -- IS ordinary
# Perl to perl, evaluated when `write` runs. It is inside the opaque region
# all the same: the lexer does not distinguish picture lines from argument
# lines, and it does not need to, because both belong to the format and
# neither is lexed here.
#
# The variables are `our` rather than `my` because a format body is
# compiled in its own scope and cannot see a lexical declared beside it;
# measured, `my $name` leaves the field empty. `our` is tier 05.
#
# The output has NO TRAILING WHITESPACE on either line, which the field
# widths were chosen for: `widget` fills 6 of the 8 columns
# `@<<<<<<<` sets and `7` is right-justified into `@>>>`, so the padding
# falls between the two fields rather than after the last one. A picture
# whose last field were left-justified would end its line with spaces, and
# the repository's `trailing-whitespace` hook would strip them out of the
# pinned output.
#
# MEASURED perl 5.42.0:
#
#   $ perl format_picture.pl | od -c
#   0000000   w   i   d   g   e   t                           7  \n   a   f
#   0000020   t   e   r       w   r   i   t   e  \n
#   0000032

--- source
our $name = "widget";
our $qty  = 7;
format STDOUT =
@<<<<<<< @>>>
$name,   $qty
.
write;
print "after write\n";

--- expect output
widget      7
after write

--- expect parses

--- expect tokens
one format body whose text is "@<<<<<<< @>>>\n$name,   $qty\n.\n"
