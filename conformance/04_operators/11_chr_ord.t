#!perl
# `chr` and `ord` are inverses over one character, and BOTH FOLD AWAY
# when their argument is a literal. This file writes each of them twice
# -- once on a runtime operand and once on a constant -- so the optree
# holds one op where the source holds two.
#
# TIER 04 operators
# INTRODUCES chr and ord
# USES my, print, %ENV, //, string interpolation
# MEASURED perl 5.42.0
#
# THE TIER'S STANDING TRAP, caught in the act. This tier's README states
# that an operator with only constant operands is erased before anything
# can observe it, and `01_arithmetic.t` honours it by never writing a
# constant pair at all. Here the constant form is written DELIBERATELY,
# beside the runtime one, because the erasure is the measurement:
#
#   $ perl -MO=Concise,-exec -e 'my $x = chr(74); print $x'
#   3  <$> const[PV "J"] s/FOLD
#   $ perl -MO=Concise,-exec -e 'my $x = ord("A"); print $x'
#   3  <$> const[IV 65] s/FOLD
#
# Not "chr with a folded argument" -- no `chr` op survives at all. The
# `s/FOLD` flag is the only trace that an operator was ever written, and
# `opsOf` collects op NAMES, so it sees a tier-01 literal.
#
# And measured on THIS FILE's source, which contains two `chr`:
#
#   $ perl -MO=Concise,-exec ... | grep -c '> chr'
#   1
#
# One op for two spellings. A corpus file that had written only the
# constant form would have declared `chr` in its tier and emitted
# nothing, and the lint -- which checks that a file's ops are CLAIMED,
# never that a claimed op is USED -- would have passed it.
#
# WHAT WAS EXPECTED AND IS WRONG. `chr` was assumed to survive a literal
# argument where `ord` and `sprintf` do not, on the reasoning that its
# result depends on the encoding pragma in scope. Measured above, it does
# not: 5.42.0 folds all three. The asymmetry is real but it runs the
# other way -- `index` and `substr` are the builtins here that DO survive
# constant arguments, and `12_index_sentinel.t` and `14_substr_arity.t`
# record that.
#
# ORD IS NOT A CHARACTER COUNT. Measured, `ord` takes the FIRST character
# of its argument and ignores the rest, which is why `ord($live)` below
# is an inverse of the `chr` above it rather than a fact about the
# string's length:
#
#   $ perl -e 'print ord("Jello")'
#   74
#
# WHY %ENV AND NOT $ARGV[0]. Both are runtime operands and both are tier
# 02's; this tier's README measures `$ARGV[0]` as the cheaper of the two
# by one op. `$ENV{X}` is used here because it is what the later files in
# this tier use, and `X` is unset when the runner executes the file, so
# the default is what arrives. The op it costs, `multideref`, is tier
# 02's and claimed there.
#
# THE TOKEN FACT COUNTS, AND THE COUNT IS THE FALSIFYING FORM. There is
# exactly ONE word spelled `ord` in this source, against TWO spelled
# `chr` -- and the grammar admits only `one` and `no`, so `chr` cannot be
# counted at all. `ord` can, and it is the useful one: a lexer that read
# `ord($c)` as anything but a word followed by a paren changes it.
#
# The negative fact is checked against the WHOLE source, `$ENV{X}`
# included: no `x` appears as a word anywhere here. `X` inside `$ENV{X}`
# is a hash subscript, and measured against our lexer it is a `Word`
# whose text is `X` -- capital, so `"x"` does not match it. That near
# miss is the reason this fact is worth asserting rather than assumed:
# `02_string.t` claims `one operator whose text is "x"` for the repetition
# operator, and this file claims the absence of the word, one case-fold
# away from each other.

--- source
my $n = $ENV{X} // 74;
my $live = chr($n);
my $folded = chr(74);
my $back = ord($live);
print "[$live][$folded][$back]\n";

--- expect output
[J][J][74]

--- expect parses

--- expect tokens
one word whose text is "ord"
