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
# THE ANGLE SPLIT, asserted here rather than described. `<$fh>` and
# `<*.c>` share a spelling and belong to different tiers: the first is
# this tier's IO construct and the second is tier 13's delimiting
# problem. Measured, they part at the optree -- `padsv readline` against
# `pushmark const gv glob` -- and they are the same thing to a lexer,
# which cannot tell a handle name from a pattern without knowing what
# the name means.
#
# So the token facts are where this tier's half of the split is made. The
# `parses` bit above cannot make it: measured, our parser reads `<`,
# `$fh`, `>` as a perfectly good comparison chain and returns ZERO
# Unknowns for it, so a lexer that split the angles would satisfy every
# behavioural assertion in this file unnoticed. The second fact is the
# one that falsifies that -- a split lexer emits an `<` operator, and an
# unsplit one never does.
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

--- expect tokens
one readline operator whose text is "<$fh>"
no operator whose text is "<"
