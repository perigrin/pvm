#!perl
# A data section's contents are not lexed: bytes perl would refuse as a
# program are read back verbatim as text.
#
# TIER 13 opaque
# INTRODUCES nothing of its own
# USES nothing from a later tier
# STATUS refuses as of 9750d03b. Cites this file. Refusal not_a_term.
#
# 08_data_section.t holds `one` and `two`. Two bare words are delimited the
# same way whether or not the section was lexed, so that file establishes
# that the marker ends the program text and leaves the NOT LEXED half of
# this tier's thesis unasserted.
#
# This file asserts it. The section holds `sub not_compiled { $x <=> }`,
# which is a syntax error as Perl -- `<=>` with no right operand -- and
# `q{ unbalanced`, whose quote never closes. Measured, perl compiles the
# program and prints both lines back: neither is a program, and neither was
# ever offered to the parser. Measured against our lexer, the whole section
# arrives as ONE DataSection token with nothing lexed inside, which is what
# the token fact below asserts.
#
# Refuses exactly as 08 does and with the same code. Both token facts pass;
# the Unknown is the parser having no rule for a trailing data section, not
# the lexer failing to delimit one. The unbalanced `q{` is the sharper case
# for that split: a lexer that had read the section would still be inside
# an unterminated quote at end of file.
#
# THE THIRD LINE OF OUTPUT IS BLANK, and it is the corpus format's own. A
# section body runs to the next `--- ` marker including the blank separator
# line before it, and inside a data section that blank line is DATA.
# 08_data_section.t records the same measurement with the full reasoning;
# it is repeated here because every data-section file inherits it and a
# file that dodged it would hide the property.
#
# The section is read in LIST context, as 08 is and for 08's reason: the
# `while (my $line = <DATA>)` spelling emits `defined`, which belongs to no
# tier yet.
#
# The content is written clear of any line beginning `--- `, which the
# corpus format would read as a section marker.
#
# MEASURED perl 5.42.0, against a file whose data section ends with a blank
# line, which is what the runner writes:
#
#   $ perl data_not_perl.pl | od -c
#   0000000   s   u   b       n   o   t   _   c   o   m   p   i   l   e   d
#   0000020       {       $   x       <   =   >       }  \n   q   {       u
#   0000040   n   b   a   l   a   n   c   e   d  \n  \n
#   0000053

--- source
my @lines = <DATA>;
print @lines;
__DATA__
sub not_compiled { $x <=> }
q{ unbalanced

--- expect output
sub not_compiled { $x <=> }
q{ unbalanced


--- expect parses

--- expect tokens
one readline operator whose text is "<DATA>"
one data section whose text is "__DATA__\nsub not_compiled { $x <=> }\nq{ unbalanced\n\n"
