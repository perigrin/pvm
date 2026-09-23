#!perl
# `9 < 1 < 5` CHAINS in 5.42.0, and the answer is the opposite of what a
# left-associative parser computes. Same source, no diagnostic either
# way.
#
# TIER 04 operators
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# This tier's README carried the measurement and no file asserted it,
# which is the shape this milestone found three times: prose is what a
# reader trusts, and a claim living only in prose is a claim the corpus
# does not make. This file is the assertion.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'print "[", (9 < 1 < 5), "]\n"'
#   []
#   $ perl -e 'print "[", (1 < 5 < 9), "]\n"'
#   [1]
#
# WHY `9 < 1 < 5` IS THE DECISIVE ONE. Read left-associatively it is
# `(9 < 1) < 5`; `9 < 1` is false, perl's false is the empty string, and
# `"" < 5` is numerically `0 < 5`, which is TRUE. So a pre-5.32 parser
# prints `1`. Read as a chain it is `9 < 1 && 1 < 5`, whose first
# conjunct is false, so it prints the empty string. Opposite answers
# from one expression, and neither reading warns.
#
# THE SECOND LINE IS THE CONTROL and it has to be an ASCENDING chain to
# be one. `1 < 5 < 9` is true both ways, so a parser that gets the first
# line right by accident -- by always printing the empty string, say --
# fails here. `1 < 9 < 5` would NOT serve: it is false under both
# readings, and a file pinning two falses proves only that something
# printed nothing twice. That mistake was made and caught by
# measurement while writing this file.
#
# CHAINING IS UNCONDITIONAL in 5.42.0. No pragma, no feature gate, no
# `use v5.32`. That matters for placement: the version-gated constructs
# in tier 11 need their pragma to reparse and this does not, so it
# belongs with the operators rather than with the gates.
#
# THE OPERANDS ARE CONSTANTS, and this is the one place in the tier
# where that is right. The README's rule -- every operator needs a
# runtime operand or the optimiser erases it -- exists so an op-based
# claim is not made about an op perl never compiled. MEASURED, chaining
# is not foldable:
#
#   $ perl -MO=Concise,-exec -e 'print "[", (9 < 1 < 5), "]\n"'
#   ...
#   5  <$> const[IV 9] s
#   6  <$> const[IV 1] s
#   7  <1> cmpchain_dup sK/1
#
# The chain survives constant operands, because it is a PARSE-time
# n-ary grouping rather than a binary expression waiting to be folded.
# `TestTierOperatorsCmpchain` takes the same exception for the same
# reason, and its probe is constants too.
#
# THIS FILE DECLARES NO TOKEN FACT, and the reason is worth the space
# because the fact it carried looked falsifiable and was not.
#
# A NEGATIVE IS A CLAIM ONLY IF THE NAMED TEXT CAN APPEAR, and that
# needs TWO things, not one. The spelling must be an entry in the
# lexer's operator table -- `scanOperator` emits nothing else -- and it
# must be REACHABLE FROM THIS SOURCE, because the table is matched
# against the bytes at each position.
#
# The discarded fact was `no operator whose text is "<="`. It passed the
# first test, which is what made it look like a claim, and failed the
# second: this source writes `9 < 1 < 5` and `1 < 5 < 9`, which contain
# no `<=` anywhere, so no lexing of them could produce one. The fact
# could never fail.
#
# Its partner `33_nonassoc_refusal.t` keeps the same spelling and there
# it IS a claim, which is what makes the distinction concrete: `33`
# writes `<=>`, whose bytes contain `<=`, so a table checked in the
# wrong order splits it and the fact fires.
#
# THE OUTPUT IS THE WHOLE CLAIM HERE and it is a strong one: two
# readings of one source that differ by a whole bracket's worth of
# content, with no diagnostic to separate them.

--- source
print "[", (9 < 1 < 5), "]\n";
print "[", (1 < 5 < 9), "]\n";

--- expect output
[]
[1]

--- expect parses
