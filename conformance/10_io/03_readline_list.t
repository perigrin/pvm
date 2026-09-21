#!perl
# `<$fh>` in list context reads EVERY remaining line: the same `readline`
# op, flagged `lK/1` rather than `sKS/1`.
#
# TIER 10 io
# INTRODUCES readline in list context
# USES nothing from a later tier
#
# This is the file the tier's placement rests on. Measured under 5.42.0:
#
#   my $l = <$fh>;      readline sKS/1
#   my @l = <$fh>;      readline lK/1
#
# One op, two behaviours, selected by what receives it. A parser that
# cannot say which context an expression is in cannot say what `<$fh>`
# returns, and it cannot learn that from the readline -- which is why this
# tier declares `DEPENDS ON 03_context` and cannot precede it.
#
# The handle holds two lines and the count is printed rather than the
# lines, because the count is what distinguishes this from
# `02_readline_scalar.t`. Printing the lines would show `one` first in
# both files and differ only in what followed.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'open(my $fh, "<", \"one\ntwo\n"); my @l = <$fh>; print scalar(@l), "\n"'
#   2

--- source
open(my $fh, "<", \"one\ntwo\n");
my @lines = <$fh>;
print scalar(@lines), "\n";
close($fh);

--- expect output
2

--- expect parses
