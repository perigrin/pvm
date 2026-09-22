#!perl
# `tie VARIABLE, CLASS, LIST` takes a VARIABLE where every other call in
# this tier takes an expression, and hands the trailing LIST to a
# constructor it names by string.
#
# TIER 11 oo
# INTRODUCES the tie builtin
# USES my, print, bless, package, sub, return, %ENV, //
# MEASURED perl 5.42.0
#
# `tie` and `tied` were UNCLAIMED by every tier of this corpus -- a
# search of all fourteen INTRODUCES blocks finds neither -- and `tie` is
# the one whose SHAPE is distinctive. Three things about the argument
# list are true of no other construct here:
#
#   - The first argument is a VARIABLE, not an expression. `tie my $x,
#     ...` DECLARES it in the same breath; measured, the pad slot is
#     created with `/LVINTRO` by the `tie` statement itself.
#   - The second is a CLASS NAMED BY STRING, resolved at runtime.
#   - The trailing LIST is passed to the constructor, so the extent of
#     the argument list decides what the class receives.
#
# A parser that read `tie` as an ordinary named operator gets the second
# and third right and the first wrong: `my $x` in argument position is
# a declaration, and nothing in the call site says so.
#
# MEASURED perl 5.42.0, this file, main program only:
#
#   b  <0> pushmark s
#   c  <0> padsv[$x:41,42] sRM/LVINTRO   <- the VARIABLE, declared here
#   d  <+> multideref($ENV{"X"}) sK
#   e  <|> dor(other->f) sK/1
#   f      <$> const[PV "Counter"] s     <- the CLASS, a runtime string
#   g  <$> const[PV "a"] s               <- the LIST, two more elements
#   h  <$> const[PV "b"] s
#   i  <@> tie vK/3
#
# `vK/3` is the arity: THREE children under the mark, which is variable,
# class and the two-element list flattened. The constructor's two
# arguments are not a pair the op can see -- `tie` pushes a flat list and
# `TIESCALAR` shifts what it wants off `@_`. A parser that stopped the
# argument extent at the class name emits `vK/2` and the program still
# runs, printing the same thing, which is why the arity is recorded here
# and the OUTPUT cannot carry this half of the claim.
#
# WHY THIS TIER, and it was measured into place rather than argued.
# `07_subroutines` was the first proposal, on the ground that `TIESCALAR`
# and `FETCH` are ordinary named subs -- which is true and is not
# sufficient. `tie` REQUIRES the constructor to return a BLESSED object,
# and a constructor that does not fails SILENTLY:
#
#   $ perl -e 'package C; sub TIESCALAR { my $s = "x"; return \$s }
#       sub FETCH { return 42 }
#       package main; tie my $c,"C"; print "[", $c, "]\n";'
#   []
#
# No error, no warning, an empty value. So the blessed constructor is not
# optional, and the minimal one emits `emptyavhv` and `bless` -- BOTH of
# them this tier's, and both counted against this file because `opsOf`
# reads inside a file's own CVs. Tier 11 is the EARLIEST tier that can
# hold `tie`, and `07_subroutines/README.md` records the same measurement
# from the other side.
#
# `bless {}, "Counter"` inside the constructor rather than `$_[0]`: the
# class arrives as a string this file already names, and `$_[0]` would
# add tier 02's `aelemfast` for nothing this construct is about.
#
# `$ENV{X} // "Counter"` keeps the class name a RUNTIME value, so the
# `tie` cannot fold to a constant; `X` is unset when the runner executes
# the file. The same idiom, for the same reason, as `10_isa_infix.t`.
#
# THE OUTPUT IS THE DISPATCH. `42` is `FETCH`'s return value reached by
# reading `$x`, and the read looks like an ordinary scalar read in the
# source -- `print $x` -- which is the whole point of the construct: the
# call is INVISIBLE at the call site. Measured:
#
#   $ perl conformance/11_oo/12_tie_variable.t
#   42
#
# A parser that compiled `print $x` as a plain pad read prints the empty
# string, because an untied `my $x` is undef. There is no spelling of
# this construct in which the dispatch is written down.
#
# THE TOKEN FACT pins `tie` as ONE word. It is the count that a lexer
# folding the keyword into the declaration beside it -- `tie my` read as
# one thing, which is how the construct reads -- gets wrong.
#
# THE SUBSTRING WORRY IS NOT REAL, and recording that is worth a
# paragraph because it is not obvious and it cost a draft. The sibling
# construct is spelled `tied`, and `tie` is a prefix of it, so the
# obvious fear is that a file containing both cannot count either.
# Measured against the checker rather than guessed: `checkTokenFact`
# (`internal/conformance/fact.go:50`) compares `tokenText == text` on the
# token's FULL text and never as a substring, so `one word whose text is
# "tie"` counts exactly one even in a file that also spells `tied`. This
# source does not, but `00_adjacency.t` holds both and relies on exactly
# this. `13_tied_boolean.t` is the sibling.
#
# A NEGATIVE `no word whose text is "tied"` was drafted here and
# WITHDRAWN, and the reason is the substring paragraph above read
# backwards. `checkTokenFact` runs against the SOURCE section alone
# (`run.go:146` passes `src`), so the fact would pass -- but it would
# assert nothing, because with full-text comparison there was never a way
# for `tied` to appear in a source that does not write it. A negative
# earns its place by ruling out a lex a positive cannot; this one rules
# out nothing, and a fact that cannot fail is noise in a file whose other
# claim is a count.

--- source
package Counter;
sub TIESCALAR { return bless {}, "Counter" }
sub FETCH { return 42 }
package main;
tie my $x, $ENV{X} // "Counter", "a", "b";
print $x, "\n";

--- expect output
42

--- expect parses

--- expect tokens
one word whose text is "tie"
