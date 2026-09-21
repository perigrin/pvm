#!perl
# `<<~EOT` strips the terminator's indentation from every body line, so the
# heredoc can be indented with the code around it.
#
# TIER 13 opaque
# INTRODUCES the indentation-stripping heredoc
# USES nothing from a later tier
# STATUS refuses as of 6c231691, and this was NOT a known gap. Cites this file.
#
# One Unknown, same as 01 and 02. All three heredoc spellings lex correctly
# and parse to the same refusal, which is the useful shape: the gap is one
# gap, not three.
#
# The `~` belongs to the OPENER token, not to a separate operator: the
# opener's text is `<<~EOT`, three characters plus the terminator name.
# And the stripping happens in the lexer, so by the time an op exists the
# body is already `trimmed\n` -- identical to what an unindented `<<EOT`
# would have produced. The body token still carries the leading spaces,
# which is what distinguishes this file from one that simply did not
# indent.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $h = <<~EOT;
#       trimmed
#       EOT
#   print $h;'
#   trimmed

--- source
my $h = <<~EOT;
    trimmed
    EOT
print $h;

--- expect parses

--- expect output
trimmed

--- expect tokens
one heredoc opener whose text is "<<~EOT"
one heredoc body whose text is "    trimmed\n    EOT\n"
