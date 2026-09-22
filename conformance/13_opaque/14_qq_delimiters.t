#!perl
# `qq` emits NO OP OF ITS OWN, and its delimiters are invisible to every
# observation but the token stream.
#
# TIER 13 opaque
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# THIS FILE CARRIES ALL OF ITS WEIGHT IN TOKEN FACTS, and that is the
# point of it rather than an admission. Measured, `qq` with an
# interpolation compiles to tier 01's `multiconcat` and `qq` without one
# compiles to tier 01's `const`:
#
#   $ perl -MO=Concise,-exec -e 'print qq{plain}'
#   ...
#   4  <$> const[PV "plain"] s
#
# So this tier REACHES `multiconcat` and introduces nothing here. A `qq`
# string and a `"` string are the same op with the same contents; the
# optree records what the string IS and never how it was spelled. Four
# spellings in the source below produce four ops indistinguishable from
# four ordinary double-quoted strings, exactly as the pod block produces
# nothing and the heredoc produces tier 01's const.
#
# WHAT IS PARSING-DISTINCTIVE ABOUT `qq` IS ARBITRARY DELIMITERS. `qq{}`,
# `qq()`, `qq[]` and `qq!!` are one operator with four terminators, and a
# lexer that hardcodes `"` or scans for a fixed closer gets three of them
# wrong. GLOSSARY.md already records that bracketing delimiters NEST, and
# this file is where the corpus measures it.
#
# THE WRONG PARSE THIS RULES OUT, stated as the token a broken lexer
# would produce: `Quote("qq{a{b}")`. A lexer that scans forward to the
# FIRST `}` ends the string there, leaving `c=$x}` as loose tokens and a
# stray CloseBracket. That is the single most likely qq bug and it is why
# the nesting case is here rather than a fourth plain delimiter. Measured
# under 5.42.0, perl counts the inner braces as content:
#
#   $ perl -e 'print length(qq{a{b}c}), "\n"'
#   5
#
# Five characters -- `a`, `{`, `b`, `}`, `c` -- so the inner `}` is not a
# terminator. The negative fact below names the wrong token directly, and
# `qq{a{b}` appears nowhere else in this file, so it is a real claim
# rather than a fact defeated by its own source.
#
# `$ENV{X} // 1` is the corpus idiom for a runtime value, and here it
# also gives the interpolating case something to interpolate that is not
# a constant. `$ENV{X}` lexes as Variable, Operator, Word, CloseBracket
# -- NOT as one token -- so it cannot satisfy or defeat either fact
# below.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $x = $ENV{X} // 1; print qq{a{b}c=$x}, qq(p), qq[q], qq!r!, "\n";'
#   a{b}c=1pqr
#
# and measured against our lexer, the four spellings arrive as four Quote
# tokens, the first of them spanning both inner braces.

--- source
my $x = $ENV{X} // 1;
print qq{a{b}c=$x}, qq(p), qq[q], qq!r!, "\n";

--- expect output
a{b}c=1pqr

--- expect parses

--- expect tokens
one quote-like operator whose text is "qq{a{b}c=$x}"
no quote-like operator whose text is "qq{a{b}"
one quote-like operator whose text is "qq(p)"
one quote-like operator whose text is "qq[q]"
one quote-like operator whose text is "qq!r!"
