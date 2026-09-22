#!perl
# `split`'s FIRST ARGUMENT IS A PATTERN, not an expression. `split /,/, $s`
# puts a regex literal in an argument slot, where a parser reading `/` as
# division produces a tree perl never builds.
#
# TIER 09 regex
# INTRODUCES the pattern in an argument slot
# USES nothing from a later tier
#
# The tier's README predicted this file: it recorded `split` as emitting
# its own op and being "deliberately out of scope ... if they are added
# later, the list grows by exactly those two". This is one of the two, and
# the list grows by `split`.
#
# THE FAILURE IS SILENT, which is why the file is written the way it is.
# `split /,/, $s` and `split ",", $s` behave identically on every ordinary
# input, so a parser that mistook the pattern for a division, recovered,
# and produced a string would print exactly what a correct one prints.
#
# Except on ONE input, and it is the file's whole measurement. A single
# SPACE as split's first argument is perl's awk-compatibility special
# case: the STRING `" "` means "split on runs of whitespace and discard
# leading whitespace", while the PATTERN `/ /` means what it says, one
# space. Measured under 5.42.0 on the string `"  a b "`:
#
#   $ perl -e 'print join "|", split / /, "  a b "'
#   ||a|b
#   $ perl -e 'print join "|", split " ", "  a b "'
#   a|b
#
# So the two spellings are a PARSE apart, and perl itself distinguishes
# them at parse time rather than at runtime.
#
# AND THE OPTREE CANNOT SEE IT. Measured, both spellings emit the same op
# with the same pattern text printed inside it:
#
#   split(/" "/ => @p:2,3)[t4] vK/LVINTRO,ASSIGN,LEX,IMPLIM
#
# Byte-identical, for two programs that print different things. That is
# this tier's standing argument -- delimiters erase themselves in the
# optree -- in its sharpest form yet, because here the erasure hides a
# difference the OUTPUT can still see.
#
# The token facts are the only place the pattern/string distinction can
# be asserted, and they are COUNTED rather than merely forbidden. Two
# facts, each falsifying a different mis-lex:
#
#   no operator whose text is "/"
#
# forbids the division reading. A lexer that scanned `/ /` as two divide
# operators around a space would emit two `Operator("/")` tokens where the
# pattern belongs, and this fact fails. It is the same negative claim
# `01_bare_match.t` makes, for the same glossary reason: a bare-slash
# pattern is spelled with delimiters alone, so `quote-like operator` --
# defined as a quote spelled with an operator NAME -- offers no positive
# category to assert it under.
#
#   one string literal whose text is "\" \""
#
# is the COUNTING half, and it is the sharper of the two. This source
# holds exactly ONE quoted string spelled `" "`: the second split's
# argument. The first split's `/ /` must NOT be one. A lexer that read a
# slash-delimited pattern as a string literal -- treating `/` as a quote
# character, which is the other way to get this wrong -- would produce
# TWO tokens matching that text and fail the count. Measured, ours emits
# `Quote("/ /")` for the pattern and `Quote("\" \"")` for the string, and
# only the second is a string literal by the glossary's rule.
#
# Neither fact can be written as `one word whose text is "split"`: the
# source calls `split` twice, on purpose, since the whole measurement is
# the two spellings side by side. The facts grammar has only `one` and
# `no`, so the count that matters is the one that CAN be pinned at one.
#
# The subject reads $ENV{X}, unset when the runner executes the file,
# because `split` on a constant string is a candidate for folding and a
# folded split measures nothing.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/09_regex/10_split_pattern.t
#   4 [  a b] 2 [a b]

--- source
my $s = $ENV{X} // "  a b ";
my @pat = split / /, $s;
my @str = split " ", $s;
print scalar(@pat), " [@pat] ", scalar(@str), " [@str]\n";

--- expect output
4 [  a b] 2 [a b]

--- expect parses

--- expect tokens
no operator whose text is "/"
one string literal whose text is "\" \""
