#!perl
# `substr` takes TWO, THREE or FOUR arguments, a negative offset counts
# from the end, and a missing length means "to the end" -- three
# different meanings for the same word, none of them visible in the op.
#
# TIER 04 operators
# INTRODUCES substr
# USES my, print, %ENV, //, string interpolation
# MEASURED perl 5.42.0
#
# `substr` SURVIVES CONSTANT ARGUMENTS, and is the second of the two
# exceptions to this tier's folding trap. Measured:
#
#   $ perl -MO=Concise,-exec -e 'my $x = substr("hello",1,3); print $x'
#   3  <$> const[PV "hello"] s
#   4  <$> const[IV 1] s
#   5  <$> const[IV 3] s
#   6  <@> substr[t2] sK/3
#
# Three plain consts with no `/FOLD` and a live op, where the same shape
# of call to `chr`, `ord` or `sprintf` collapses to a single folded
# constant. `12_index_sentinel.t` records the other exception. The
# runtime operand is still written here, because what a file may RELY on
# is the rule, and the exceptions are a release's behaviour rather than
# the language's.
#
# ARITY IS THE CONSTRUCT AND THE OP CANNOT REPORT IT. Measured on the
# three calls below:
#
#   substr($s, 6)        <@> substr[t4] sK/2
#   substr($s, 3, 2)     <@> substr[t6] sK/3
#   substr($s, -5, 3)    <@> substr[t9] sK/3
#
# One name, three calls, and `opsOf` dedupes to a single `substr`. This
# is `10_undef_arity.t`'s finding met a second time and from the other
# side: there the two arities of `undef` produced OPPOSITE results under
# one name, here three arities produce three different substrings under
# one name. An op-name lint is satisfied by a file that only ever spelled
# the two-argument form.
#
# THE NEGATIVE OFFSET IS NOT A NEGATIVE INDEX ERROR. `substr($s, -5, 3)`
# starts five characters from the END and takes three forward, which is
# `wor` of `hello world`. A reader who expects the length to run backwards
# too gets `orl` and a parser that clamped the offset at zero gets `hel`,
# so the printed value discriminates all three readings -- which is what
# this tier's README requires of a fixture before it counts as a
# measurement.
#
# A MISSING LENGTH IS NOT A ZERO LENGTH. `substr($s, 6)` is `world`, the
# whole remainder, and not the empty string. That is the two-argument
# form, and it is where a parser filling in a default of 0 for the absent
# third argument silently produces an empty result that no op name shows.
#
# THE OFFSET-ZERO CASE IS DELIBERATELY ABSENT, and the reason is a
# separate op. Measured, `substr($s, 0, 1)` in rvalue position compiles
# to an op named `substr_left`, which is an optimisation 5.42.0 applies
# only at offset zero:
#
#   $ perl -MO=Concise,-exec -e 'my $s = $ENV{X} // "hello"; print substr($s,0,1)'
#   ...  <@> substr_left[t3] sK/2
#
# `substr_left` is claimed by NO TIER in this corpus, so a file here
# spelling an offset-zero rvalue substr would fail the dependency lint --
# correctly, since the lint's job is to stop the union growing by
# accident. Every offset in this file is nonzero for that reason, and it
# is a real limit rather than a stylistic choice: the most ordinary
# spelling of `substr` anyone writes is the one this corpus cannot yet
# carry. `15_substr_lvalue.t` reaches offset zero by a route that emits
# plain `substr`, which is how the corpus covers the position at all.
#
# THE TOKEN FACTS. There are THREE words spelled `substr` here and the
# grammar admits only `one` and `no`, so the construct cannot be counted
# -- the limit `09_named_unary.t` and `10_undef_arity.t` both record.
# `one operator whose text is "-"` pins the single negation in the file,
# and it is the falsifying form for the offset case: `-5` is TWO tokens
# by the glossary, a negation operator and a numeric literal, so a lexer
# that folded the sign into the number would find none. That boundary is
# `GLOSSARY.md`'s and `01_literals` asserts the numeric half of it; this
# is the operator half.
#
# `no word whose text is "sprintf"` is checked against the whole source,
# `$ENV{X}` included, and holds: the text appears nowhere here. It is
# worth asserting because `13_sprintf_star.t` is this file's neighbour and
# the two builtins fold on OPPOSITE sides of the trap, so a file that
# drifted into borrowing the other's operand would break this first.

--- source
my $s = $ENV{X} // "hello world";
my $tail = substr($s, 6);
my $mid = substr($s, 3, 2);
my $neg = substr($s, -5, 3);
print "[$tail][$mid][$neg]\n";

--- expect output
[world][lo][wor]

--- expect parses

--- expect tokens
one operator whose text is "-"
no word whose text is "sprintf"
