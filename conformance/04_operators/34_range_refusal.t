#!perl
# `1 .. 2 .. 3` is a SYNTAX ERROR, and level 11 is the ONE `%nonassoc`
# level in perly.y that means what the keyword suggests.
#
# TIER 04 operators
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my @x = (1 .. 2 .. 3);'
#   syntax error at -e line 1, near "2 .."
#
# WHY THIS LEVEL AND NOT THE OTHER TEN. This tier's README measures all
# eleven `%nonassoc` levels against perl and finds that EIGHT of them do
# not reject repetition at all -- `print print @a`, `scalar scalar @a`
# and `require require` all compile, because a conflict needs a left
# operand to fight over and those are prefix forms whose second
# occurrence is the first one's ARGUMENT. Five more carry no lexable
# operator. Level 11 is the remainder: a genuine nonassoc that refuses.
#
# So a parser that reads the eleven `%nonassoc` lines and refuses
# repetition at each is wrong about eight of them, and a parser that
# reads them and refuses NONE is wrong about this one. The corpus needs
# the rejecting case asserted or only the accepting direction is
# claimed.
#
# ITS PARTNER IS `33_nonassoc_refusal.t`, which carries perlop's level
# 13 -- `1 <=> 2 <=> 3`, also a syntax error, and a level that is NOT
# one of perly.y's eleven. The two files are different claims: 33 says a
# perlop level documented `chain/na` has halves that disagree, and this
# one says a perly.y `%nonassoc` level actually is one. Neither implies
# the other, which is why both exist.
#
# ONE REFUSAL PER FILE is forced by the format -- perl stops at the
# first error, so a source holding both would only ever report one.
#
# THE OP BUDGET DOES NOT APPLY HERE, and the reason is worth stating
# because `..` is tier 03's `range`/`flip`/`flop` and a `parses` file
# spelling it in this tier would fail the dependency lint. A `parsent`
# file emits NO OPS AT ALL: perl builds no optree for a program it will
# not compile, and `lint_test.go` returns early on `ExpectParsent` for
# exactly that reason. The refusal is what makes the file legal here.
#
# THE TOKEN FACT IS ONE NEGATIVE AND IT IS REACHABLE IN THIS SOURCE,
# which is the test a negative has to pass and the test this corpus has
# now failed seven times. A fact is a claim only if the named text can
# appear, and that needs TWO things: the spelling must be an entry in
# the lexer operator table, and it must be reachable from this source.
# `.` is both -- a table entry, and four of them are written here
# inside the two `..` operators -- so the fact fails exactly when the
# lexer splits a range into two concatenations.
#
# `...` was drafted alongside it and DROPPED. It is a table entry, so
# it passes the first test; but this source writes `1 .. 2 .. 3` with
# spaces and holds no three consecutive dots, so no lexing of it could
# produce one. The fact would have been unfalsifiable -- the same
# mistake as `+++` and `+(` before it, reached by the second route
# rather than the first.

--- source
my @x = (1 .. 2 .. 3);

--- expect parsent

--- expect tokens
no operator whose text is "."
