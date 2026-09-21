#!perl
# `<$fh>` in scalar context reads ONE line: `readline sKS/1`.
#
# TIER 10 io
# INTRODUCES readline in scalar context
# USES nothing from a later tier
#
# This is the half of `readline` that a scalar receives, and it is the
# baseline `03_readline_list.t` deviates from. The source differs there by
# one sigil, and that one sigil changes how much of the handle is
# consumed -- one line here, every remaining line there -- without
# changing the op name.
#
# The handle holds two lines and this file prints one, so the output is
# also the evidence that the scalar form stopped. A file reading a
# one-line handle could not tell the two contexts apart by behaviour.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'open(my $fh, "<", \"one\ntwo\n"); my $l = <$fh>; print $l'
#   one

--- source
open(my $fh, "<", \"one\ntwo\n");
my $l = <$fh>;
print $l;
close($fh);

--- expect output
one

--- expect parses
