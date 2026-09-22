#!perl
# `pos` is LVALUE-CAPABLE and tied to the regex engine: it reads and
# WRITES the match position `//g` leaves behind on a scalar.
#
# TIER 09 regex
# INTRODUCES the match position, read and assigned
# USES nothing from a later tier
#
# THE LVALUE FORM IS WHY THIS IS A PARSE AND NOT A CALL. `pos($s) = 0` puts
# a named operator on the LEFT of an assignment, which almost nothing in
# Perl does -- measured, the optree puts the `pos` op underneath
# `sassign`'s left arm:
#
#   $ perl -MO=Concise,-exec -e 'my $s = "x"; pos($s) = 0;'
#   ... <1> pos[t3] sKRM*/1
#       <2> sassign vKS/2
#
# A parser that modelled `pos` as an ordinary named unary returning a
# value would refuse that assignment or silently discard it, and the
# second case is the dangerous one: the program still runs.
#
# THE OUTPUT REPORTS THE DISCARD, which is the file's whole design. `//g`
# on a scalar advances a stored position, so two successive `/b/g` matches
# against `"abcabc"` find the two `b` characters in turn. Measured:
#
#   without the reset:  2 5     the second match continues from 2
#   with    the reset:  2 2     the second match restarts from 0
#
# So a parser that dropped `pos($s) = 0` prints `2 5` where this file pins
# `2 2`. Nothing else in the source differs, and the discarded statement is
# the only thing that could account for it.
#
# The op is `pos`, this tier's fifth, and it is the same op in both
# positions -- rvalue `my $a = pos($s)` and lvalue `pos($s) = 0` differ in
# the flags perl prints, not in the op. So the op claim is earned by
# either spelling and the LVALUE claim is behavioural, which is the
# division this corpus keeps running into.
#
# THE TOKEN FACT IS ABOUT THE MODIFIER, not about `pos`. `pos` lexes as an
# ordinary `Word` and this source holds three of them, so the facts
# grammar's `one`/`no` cannot pin it. What the facts CAN pin is the thing
# `pos` is useless without: `/b/g` is a pattern carrying a MODIFIER, and
# the modifier is part of the quote token. A lexer that stopped at the
# closing delimiter would leave `g` behind as a stray `Word("g")` -- a
# program that still parses, since a bareword is legal there, and still
# runs, since a non-global match sets pos too. It would just print `2 2`
# for the wrong reason.
#
# So both facts are negative, and they bracket the pattern from each
# side: `no word whose text is "g"` forbids the modifier escaping the
# quote, and `no word whose text is "b"` forbids the pattern BODY escaping
# it, which is what a lexer reading the slashes as division would produce.
# Together they say the three characters `b/g` were consumed by the quote
# and by nothing else.
#
# A positive fact is not available and the reason is the glossary's, not
# an omission. The source spells the match `m/b/g` twice, deliberately --
# once before the reset and once after, since the measurement IS the two
# matches -- so `one quote-like operator whose text is "m/b/g"` would be
# false at a count of two, and the facts grammar has only `one` and `no`.
# Spelling one of them differently to make the count work would put a
# second delimiter form in the file for no reason but the assertion.
#
# The subject reads $ENV{X}, unset when the runner executes the file,
# because a match against a constant is a folding candidate.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/09_regex/12_pos.t
#   2 2

--- source
my $s = $ENV{X} // "abcabc";
$s =~ m/b/g;
my $a = pos($s);
pos($s) = 0;
$s =~ m/b/g;
my $b = pos($s);
print "$a $b\n";

--- expect output
2 2

--- expect parses

--- expect tokens
no word whose text is "g"
no word whose text is "b"
