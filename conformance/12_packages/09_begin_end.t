#!perl
# `BEGIN` runs at COMPILE time, so a block written last executes first.
#
# TIER 12 packages
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# `BEGIN` appears in 99 of T1's 986 files and in no corpus file. Nor did
# `END`, `__PACKAGE__`, `__FILE__`, `__LINE__` or `use constant`. These
# are the constructs where parsing and execution INTERLEAVE -- where a
# construct changes what the rest of the parse sees -- and the corpus
# made no claim about any of it.
#
# MEASURED perl 5.42.0. The ops say nothing:
#
#   $ perl -MO=Concise,-exec -e 'BEGIN { print "B\n" } print "main\n"'
#   ... const ... print ... const ... print ...
#
# Two prints and two constants, exactly as two ordinary statements give.
# `BEGIN` and `END` emit no op of their own: they are compile-time
# instructions, and the optree records only what ends up running. So the
# op lint cannot see this construct and the claim has to be ORDERING.
#
#   $ perl -e 'END { print "end\n" } BEGIN { print "begin\n" }
#              print "main\n"'
#   begin
#   main
#   end
#
# THE SOURCE ORDER IS END, BEGIN, PRINT AND THE OUTPUT IS THE REVERSE OF
# NEITHER. That is the whole file: a parser that treated the two phasers
# as ordinary blocks would run them where they appear and print
# `end begin main`, and one that ignored phasing entirely would print
# them in source order too. Only a parser that knows `BEGIN` is hoisted
# to compile time and `END` deferred to exit produces this sequence.
#
# The token facts count the two phaser names. They are barewords followed
# by a block with no semicolon and no parens -- structurally identical to
# a sub call with a hashref argument, which is what a lexer that did not
# know them would produce.

--- source
END { print "end\n" }
BEGIN { print "begin\n" }
print "main\n";

--- expect output
begin
main
end

--- expect parses

--- expect tokens
one word whose text is "BEGIN"
one word whose text is "END"
