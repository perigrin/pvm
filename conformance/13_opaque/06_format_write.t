#!perl
# A `format NAME =` declaration's picture lines are NOT Perl, and the
# declaration itself emits no ops -- only the `write` that uses it does.
#
# TIER 13 opaque
# INTRODUCES the format declaration
# USES nothing from a later tier
# STATUS refuses as of 6c231691, and this was NOT a known gap. Cites this file. Refusal unimplemented_statement.
#
# `unimplemented_statement` is the right code here and it is not the same
# refusal the heredoc files carry. Those parse their statement and then
# meet a body token they have no form for; this one never starts, because
# `format` is a statement keyword the parser does not implement. Measured,
# the Unknown spans `format STDOUT =` through the `write;` that follows it
# -- the declaration and the statement after it swallowed together, which
# is what an unimplemented keyword does to whatever it cannot find an end
# for.
#
# The lexer delimits the format body correctly -- the `format body` fact
# below passes -- and the parser produces one Unknown over the declaration.
# So the picture lines are already opaque to the lexer, which is the hard
# half, and what is missing is a grammar rule for `format NAME = BODY`.
#
# The body runs from the `=` to a line holding a lone `.`, and it is the
# clearest case in the tier of a region the lexer must delimit without
# lexing: `@<<<<<` inside a picture line is a left-justified column, not an
# array sigil followed by two left shifts. This file's picture is plain
# text, which keeps the OUTPUT deterministic while the body token is what
# asserts the region was recognised.
#
# `enterwrite` is claimed by this tier, but it belongs to `write`. The
# `format` declaration compiles to NOTHING: measured, a file holding only a
# format and a `write` emits `enter`, `nextstate`, `enterwrite`, `leave`,
# and the declaration accounts for none of them.
#
# The picture is a fixed line rather than a `@<<<` field because a field
# pads its column with spaces to a width the picture sets, and the runner
# compares bytes. A fixed line keeps this file about the LEXING and leaves
# field padding to a file that declares padding as its subject.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'format STDOUT =
#   a fixed report line
#   .
#   write;
#   print "after write\n";' | od -c
#   0000000   a       f   i   x   e   d       r   e   p   o   r   t       l
#   0000020   i   n   e  \n   a   f   t   e   r       w   r   i   t   e  \n
#   0000040

--- source
format STDOUT =
a fixed report line
.
write;
print "after write\n";

--- expect parses

--- expect output
a fixed report line
after write

--- expect tokens
one format body whose text is "a fixed report line\n.\n"
one word whose text is "format"
