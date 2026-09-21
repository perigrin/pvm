#!perl
# A `<<'EOT'` heredoc does not interpolate: `$name` in the body is five
# characters of text, not a variable.
#
# TIER 13 opaque
# INTRODUCES the non-interpolating heredoc
# USES nothing from a later tier
# STATUS refuses as of 6c231691, and this was NOT a known gap. Cites this file.
#
# Refuses exactly as 01 does, and for the same reason: the lexer produces
# the two heredoc tokens correctly and the parser produces one Unknown. The
# quoting of the opener is not what the parser trips on.
#
# This is the pair to 01, and the pair is the point. The two spellings
# differ by one quote character and produce DIFFERENT strings, which makes
# this the one heredoc property whose behaviour is observable: 01 prints
# `hello world` and this prints `hello $name` with `$name` in scope and
# holding `world`. Everything else about a heredoc -- where it starts,
# where it ends, that it is a heredoc at all -- is invisible below the
# token stream.
#
# Measured, this compiles to a plain `const[PV "hello $name\n"]`: no
# multiconcat, because there is nothing to interpolate.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $name = "world"; my $h = <<'\''EOT'\'';
#   hello $name
#   EOT
#   print $h;'
#   hello $name

--- source
my $name = "world";
my $h = <<'EOT';
hello $name
EOT
print $h;
print "$name\n";

--- expect parses

--- expect output
hello $name
world

--- expect tokens
one heredoc opener whose text is "<<'EOT'"
one heredoc body whose text is "hello $name\nEOT\n"
