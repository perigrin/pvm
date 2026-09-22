#!perl
# `tr///` is the only quote-like besides `s///` that takes TWO delimited
# regions -- and its second region is NOT A PATTERN and neither is its
# first. That asymmetry is the trap.
#
# TIER 09 regex
# INTRODUCES the second quote-like operator with two regions
# USES nothing from a later tier
#
# The tier's README predicted this file alongside `10_split_pattern.t`:
# `split` and `tr///` "each emit their own op (`split`, `trans`) and are
# deliberately out of scope ... if they are added later, the list grows by
# exactly those two." This is the second of the two, and the list grows by
# `trans`.
#
# WHY THE TWO-REGION SHAPE IS THE LEXICAL CLAIM. A lexer that knows only
# `s` takes a second region reads `tr/./Z/` as `tr/./` followed by the
# stray tokens `Z` and `/`, which is a different token stream and a
# different program. Nothing about `tr`'s first region announces that a
# second one follows; the operator NAME is the only signal, so a lexer
# must carry a table of which quote-like names take two regions, and
# `tr` (with its synonym `y`) is the entry most often missing from it.
#
# AND `tr` IS NOT `s` WHERE IT COUNTS. Both take two regions and there the
# resemblance stops. Measured under 5.42.0 on the same subject and the
# same-looking source:
#
#   $ perl -e '$_ = "a.c"; tr/./Z/; print'
#   aZc
#   $ perl -e '$_ = "a.c"; s/./Z/;  print'
#   Z.c
#
# `tr`'s `.` is the CHARACTER dot; `s`'s `.` is the metacharacter matching
# anything. So a parser that desugared `tr` into `s` -- an easy thing to
# do when the syntax is this similar -- produces a program that compiles,
# runs, and prints the wrong answer. The file runs both spellings on one
# subject so the two results sit side by side in the output, and a
# desugared `tr` prints `Z.c 1 Z.c` instead of `aZc 1 Z.c`.
#
# THE RETURN VALUE IS THE SECOND BEHAVIOURAL PIN. `tr` returns the COUNT
# of characters it transliterated, where a bare match returns a boolean.
# Measured, `"a.c" =~ tr/./Z/` is 1. Pinning it costs one variable and
# catches a `tr` whose result was modelled on a match's -- the count and
# the boolean agree at 1 here, which is why the count is pinned beside the
# two strings rather than trusted to carry the claim alone.
#
# The op is `trans`, and it is one of this tier's three additions.
# Measured:
#
#   $ perl -MO=Concise,-exec -e 'my $t = "a.c"; $t =~ tr/./Z/'
#   ... <"> trans[$t:1,2] sP/TRANS=ONLY_UTF8_INVARIANTS ...
#
# One op, no `match`, no `regcomp`, no `subst` -- which is the optree
# agreeing that `tr` is not a regex construct at all despite wearing the
# syntax of one. It sits in this tier because the LEXER cannot tell them
# apart, and the lexer is what this tier is about.
#
# THE REPLACEMENT CHARACTER IS `Z` AND NOT `X`, which looks arbitrary and
# is not. `no word whose text is "Z"` is the fact that falsifies a lexer
# leaking the second region -- and with `X` as the replacement the fact is
# false before `tr` is reached, because `$ENV{X}` puts a `Word("X")` in
# the source's first line. The runner caught it. A negative token fact is
# only a claim about the construct if the character it names appears
# nowhere else in the file.
#
# The subject reads $ENV{X}, unset when the runner executes the file,
# because a `tr` on a constant is a folding candidate and a folded `tr`
# measures nothing.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/09_regex/11_transliteration.t
#   aZc 1 Z.c

--- source
my $s = $ENV{X} // "a.c";
my $t = $s;
my $n = ($t =~ tr/./Z/);
my $u = $s;
$u =~ s/./Z/;
print "$t $n $u\n";

--- expect output
aZc 1 Z.c

--- expect parses

--- expect tokens
one quote-like operator whose text is "tr/./Z/"
one quote-like operator whose text is "s/./Z/"
no word whose text is "Z"
