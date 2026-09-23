#!perl
# A single-quoted string does not interpolate, and its backslash escapes
# are not the double-quoted set: `'a\nb'` is FOUR characters.
#
# TIER 01 literals
# INTRODUCES string literal
# USES nothing from a later tier
#
# The tier was chartered for "numbers, strings, quoting and qw" and shipped
# twelve numeric files and no string file at all. The three string
# boundaries its glossary check names were satisfied only by
# `00_adjacency.t`, whose own header says it introduces nothing -- so the
# tier asserted the numeric half of its vocabulary and borrowed the rest
# from a file that exists to compose, not to introduce.
#
# MEASURED perl 5.42.0. The op stream cannot see this construct at all:
#
#   $ perl -MO=Concise,-exec -e 'my $s = q{plain}; print $s'
#   ... const[PV "plain"] ... padsv_store ... print ...
#   $ perl -MO=Concise,-exec -e 'my $s = "plain"; print $s'
#   ... const[PV "plain"] ... padsv_store ... print ...
#
# Byte-identical. A single-quoted string, a double-quoted string with
# nothing to interpolate, and `q{...}` all arrive as one `const`, which is
# why this file's claim is LEXICAL and why GLOSSARY.md keeps `string
# literal` and `quote-like operator` as separate categories.
#
# The ESCAPE is what output can still see, and it is the discriminating
# half:
#
#   $ perl -e 'print length(q{a\nb})'      4
#   $ perl -e 'print length("a\nb")'       3
#
# So a lexer that applied double-quoted escape processing to a single
# quoted string would print `a`, a newline and `b`. `length` is claimed by
# no tier, so the count is not measured here: the OUTPUT is the four
# characters themselves, and a lexer that collapsed `\n` would print three
# characters across two lines instead.
#
# The token fact spells the backslash DOUBLED, `'a\\nb'`, because the
# fact's text is compared against the source bytes and the source holds a
# literal backslash. The file's own trailing `"\n"` is a second Quote
# token, which is why the fact counts `one` of a specific text rather than
# one Quote: a bare count would be two and the claim would be about the
# file's punctuation rather than about its construct.

--- source
my $p = 'plain';
my $s = 'a\nb';
print $p, $s, "\n";

--- expect output
plaina\nb

--- expect parses

--- expect tokens
one string literal whose text is "'a\\nb'"
