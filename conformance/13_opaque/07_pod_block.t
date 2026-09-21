#!perl
# A pod block between two statements compiles to NOTHING. The only place it
# is observable is the token stream.
#
# TIER 13 opaque
# INTRODUCES the pod block
# USES nothing from a later tier
#
# Measured, this source and the same source with the pod deleted emit the
# IDENTICAL op sequence -- const, padsv_store, pushmark, padsv,
# multiconcat, print -- differing only in the line numbers recorded on the
# nextstates. So the `expect output` here cannot fail for the reason the
# file exists: a lexer that read the pod as prose, as code, or as nothing
# at all would print `1` either way.
#
# That is why this file asserts the pod token. It is the whole assertion;
# the rest is scaffolding proving the program still ran.
#
# The block runs from a `=` in column 1 followed by an identifier through
# the `=cut` line, and the terminator is part of the token, for the same
# reason a heredoc body's is: every byte belongs to the block.
#
# MEASURED perl 5.42.0, with the pod block between the two statements:
#
#   $ perl pod.pl
#   1

--- source
my $x = 1;

=pod

Prose perl never compiles.

=cut

print "$x\n";

--- expect parses

--- expect output
1

--- expect tokens
one pod block whose text is "=pod\n\nProse perl never compiles.\n\n=cut\n"
