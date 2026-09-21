#!perl
# A `<<"EOT"` heredoc's body starts on the NEXT line while the statement it
# sits in continues on the SAME one, so the body token arrives after the
# semicolon.
#
# TIER 13 opaque
# INTRODUCES the interpolating heredoc
# USES nothing from a later tier
# STATUS refuses as of 6c231691, and this was NOT a known gap. Cites this file. Refusal trailing_tokens.
#
# The code names the site, and the site is the informative part. Measured,
# the DECLARATION PARSES: `my $h = <<"EOT";` arrives as a declaration whose
# two terms are `$h` and the opener, with no Unknown in it. The Unknown
# spans what comes NEXT -- the body token and the `print` after it -- and
# its code is `trailing_tokens`, an expression that parsed with bytes
# remaining before the terminator.
#
# That is the heredoc's reordering stated as a parser failure. The body
# arrives after the semicolon that ended the statement it belongs to, so
# the parser meets it where a new statement should start and has no form
# for it. The gap is a grammar rule, not a lexer change.
#
# The LEXER handles the heredoc: it produces the opener and the body as two
# tokens, in the right order, with the right text -- the `expect tokens`
# facts below pass. The PARSER produces one Unknown over the statement. So
# the refusal is above the lexer, and the token facts are what say so: a
# corpus file asserting only `parses` would report the same failure and
# leave the boundary unlocated.
#
# The heredoc emits no op of its own: perl folds it into the same
# `multiconcat` that `my $h = "hello $name\n"` produces. So the output
# proves the CONTENT and proves nothing about the spelling, and the token
# facts are what say a heredoc was written.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $name = "world"; my $h = <<"EOT";
#   hello $name
#   EOT
#   print $h;'
#   hello world

--- source
my $name = "world";
my $h = <<"EOT";
hello $name
EOT
print $h;

--- expect parses

--- expect output
hello world

--- expect tokens
one heredoc opener whose text is "<<\"EOT\""
one heredoc body whose text is "hello $name\nEOT\n"
