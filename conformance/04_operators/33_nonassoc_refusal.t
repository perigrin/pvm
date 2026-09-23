#!perl
# `1 <=> 2 <=> 3` is a SYNTAX ERROR, and the file next to this one shows
# that `1 < 2 < 3` is not. Both are perlop's comparison band.
#
# TIER 04 operators
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'print 1 <=> 2 <=> 3'
#   syntax error at -e line 1, near "2 <=>"
#
#   $ perl -e 'print "a" cmp "b" cmp "c"'
#   syntax error at -e line 1, near ""b" cmp"
#
#   $ perl -e 'print 1 == 1 == 1'
#   1
#
# THE HAZARD IS WITHIN ONE LEVEL. perlop documents level 13 as
# `chain/na`, and the three operators above are all in it. `==` chains.
# `<=>` and `cmp` refuse. A parser that models the level as ONE
# associativity class is wrong whichever class it picks, and nothing in
# the corpus said so before this file.
#
# THIS IS A `parsent` FILE because that is what the claim is: perl
# itself rejects the source, so there is no output to pin and no tree to
# describe. An `expect parses` file could only assert the opposite.
#
# ONE REFUSAL PER FILE is forced by the format -- a file has one source
# and perl stops at the first error -- so this file carries `<=>` and
# the README's other measured refusals stay described there. What
# changes is that the level-13 split is now asserted somewhere that
# fails when it stops being true, rather than living only in prose.
#
# The operands are constants, which is the one place in this tier where
# that is correct: the file never runs. Constant folding is an optimiser
# pass over an op stream, and a source perl refuses to compile has no op
# stream to fold.
#
# THE TOKEN FACTS ARE NEGATIVES AND BOTH ARE FALSIFIABLE. A refusing
# file has no output to fall back on, so the token layer is the only
# thing here besides the refusal itself, and it asserts that the lexer
# got `<=>` RIGHT even though the parser rejects what follows.
#
# `<=` and `>` are both entries in the lexer operator table, so both
# spellings are reachable and each negative is a claim rather than a
# name nothing can produce. The table is checked longest-first exactly
# so `<=>` is not read as `<=` then `>`, and these two facts are that
# rule stated from the corpus side: if either appeared, the longest-first
# order had failed and the refusal below would be about the wrong thing.
#
# Counting would be the stronger form -- two `<=>` and not four tokens --
# but `fact.go` has only `one` and `no`, and its own comment says the
# counting form arrives with the first file that needs it. This file
# does not need it: the two negatives fail on the same lexer error a
# count would catch.

--- source
print 1 <=> 2 <=> 3;

--- expect parsent

--- expect tokens
no operator whose text is "<="
no operator whose text is ">"
