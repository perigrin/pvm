#!perl
# A pod block's contents are not lexed: text perl would refuse as a program
# passes through a pod block without a complaint.
#
# TIER 13 opaque
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# 07_pod_block.t holds prose. Prose is delimited the same way whether or
# not its contents were lexed -- `Prose perl never compiles.` is four Words
# and a Semicolon to a lexer that reads it, and the pod block's extent is
# the same either way -- so that file establishes that a pod block is one
# token and leaves the NOT LEXED half of the tier's thesis unasserted.
#
# This file asserts it. The block holds an unclosed `(`, a `$this` that is
# not a declared variable and two stray semicolons. Measured, perl compiles
# and runs the program without noticing any of it: the pod is eaten by the
# lexer before the parser exists. Measured against our lexer, the same
# bytes arrive as ONE Pod token with nothing lexed inside.
#
# `=head1` rather than `=pod`, because the block's opener is any `=` in
# column 1 followed by an identifier and 07 already covers `=pod`. Both
# terminate at `=cut`, and the terminator is part of the token for the
# reason GLOSSARY.md gives: every byte of a pod block belongs to the block.
#
# This file PASSES. It is a passing file in a tier where five refuse, and
# that is the useful shape: the pod block is the construct this tier gets
# entirely right, opaque contents and all, so the tier's refusals cannot be
# read as "the lexer cannot handle opaque regions".
#
# The block's text is written clear of any line beginning `--- `, which the
# corpus format reads as a section marker with no awareness of Perl's own
# nesting. A pod block is one of the three places in this tier that can
# contain such a line.
#
# MEASURED perl 5.42.0:
#
#   $ perl pod_not_perl.pl
#   1

--- source
my $x = 1;

=head1 NOT PERL

$this is not( a variable; and this ; is not a statement

=cut

print "$x\n";

--- expect output
1

--- expect parses

--- expect tokens
one pod block whose text is "=head1 NOT PERL\n\n$this is not( a variable; and this ; is not a statement\n\n=cut\n"
