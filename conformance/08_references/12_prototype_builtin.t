#!perl
# `prototype \&f` hands back, as a runtime STRING, the very text that
# changed how calls to `f` parse. The parser's own input is readable as
# data.
#
# TIER 08 references
# INTRODUCES prototype
# USES sub, my, print, the code reference
#
# WHY THIS TIER AND NOT 07. `prototype` takes a CODE REFERENCE, and a
# code reference is this tier's construct. Measured, `prototype \&f`
# emits `rv2cv` and `srefgen`, and both are THIS tier's: the README names
# `rv2cv` as appearing ONLY for `\&foo` on a named sub, which is exactly
# the argument `prototype` requires. Tier 07 declares a prototype and
# measures what it does to a call site, but tier 07 cannot WRITE this
# construct, because `\&f` is not available to it. Tier 08 is the
# earliest tier in which `prototype` is spellable at all, so the op lint
# and the placement argument agree rather than merely not conflicting.
#
# WHAT TIER 07 ALREADY CLAIMS, AND WHAT THIS ADDS.
# `07_subroutines/09_prototype_extent.t` measures the `($)` prototype's
# effect on PARSING: `print g 1, 2` cuts the argument extent to one and
# the second constant falls through to the enclosing `print`. That claim
# is about a CALL SITE, where the prototype is invisible and only the
# output reveals it. This file makes the complementary claim and repeats
# none of it: there is no parenless call here, and the prototyped sub is
# never called at all. What is here is the prototype coming back OUT as
# the two-character string `($)` -- the reflection tier 07 never asks
# for, and the reason the two files are not the same measurement wearing
# different numbers.
#
# THE WRONG PARSE THIS RULES OUT. `prototype \&f` is a word immediately
# followed by a reference operator applied to a sub sigil, with no
# parentheses to bracket the argument. A lexer that read `\&` as ONE
# token -- a plausible reading, since the two characters only ever occur
# together in this tier -- produces a program whose optree still contains
# `srefgen` and whose output is unchanged, because perl's ops and perl's
# bytes cannot see the difference. The token facts are the only place
# that reading dies: exactly one `\` operator, and NOT one whose text is
# `\&`.
#
# `prototype` is the one word in the file, which is also a fact worth
# pinning: a lexer that folded `prototype \&f` into a single term the way
# it might a quote-like operator would leave no separate `prototype`
# word behind. Its own text appears NOWHERE ELSE in the source, so the
# count is a real claim rather than a coincidence of the fixture.
#
# MEASURED perl 5.42.0, the whole file:
#
#   3  <#> gv[IV \"$"] s
#   4  <1> rv2cv[t3] lKRM/AMPER,TARG
#   5  <1> srefgen sK/1
#   6  <1> prototype sK/1
#
# `gv[IV \"$"]` rather than `gv[IV \&main::f]`: with a prototype in force
# perl has folded the sub's identity into its prototype string at compile
# time, the same substitution `07_subroutines/09_prototype_extent.t`
# records at its `entersub`. So the op stream does not even carry the
# sub's NAME here, which is a further reason the token facts must.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'sub f ($) { 1 } my $p = prototype \&f; print "[$p]\n"'
#   [$]

--- source
sub f ($) { 1 }
my $p = prototype \&f;
print "[$p]\n";

--- expect output
[$]

--- expect parses

--- expect tokens
one operator whose text is "\\"
one word whose text is "prototype"
no operator whose text is "\\&"
no operator whose text is "->"
