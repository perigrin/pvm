#!perl
# `index` reports failure as -1, not as undef, and that makes its result
# a NUMBER that is always defined -- so `//` cannot test it and a truth
# test is wrong at position zero.
#
# TIER 04 operators
# INTRODUCES index
# USES my, print, %ENV, //, +, string interpolation
# MEASURED perl 5.42.0
#
# `index` SURVIVES CONSTANT ARGUMENTS, which makes it one of the two
# exceptions to this tier's folding trap. Measured:
#
#   $ perl -MO=Concise,-exec -e 'my $x = index("hello","l"); print $x'
#   3  <$> const[PV "hello"] s
#   4  <$> const[PVMG "l"] s
#   5  <@> index[t2] sK/2
#
# Two plain consts, NO `/FOLD` flag, and a live `index` op. Compare
# `11_chr_ord.t`, where the same shape of call collapses to a single
# folded const. The optimiser folds `chr`, `ord` and `sprintf` and
# declines to fold `index` and `substr`, and nothing in the source shape
# predicts which. This file still takes a runtime operand, because the
# rule the tier writes down is about what a file may RELY on and not
# about what one perl release happens to do.
#
# THE SENTINEL IS THE CONSTRUCT. `index` is the only builtin in this
# tier whose "not found" answer is an ordinary in-range value:
#
#   $ perl -e 'my $s = "hello world"; print index($s,"z"), "\n"'
#   -1
#   $ perl -e 'my $s = "hello world"; print defined(index($s,"z")) ? 1 : 0'
#   1
#
# Defined, numeric, false-ish under `!` and TRUE under boolean test --
# `-1` is true in perl. So the three tests a reader reaches for all get
# it wrong in different directions: `//` never fires because the value is
# defined, `if (index(...))` is true for a miss and FALSE for a hit at
# position 0, and only `>= 0` is right. The file prints the sentinel
# rather than branching on it, because a branch would emit `cond_expr`,
# which is tier 06's op two tiers forward.
#
# THE THIRD ARGUMENT IS A POSITION, NOT A COUNT. `index($s, $sub, $pos)`
# starts the search at $pos and still reports an offset from the START of
# the string, so the second hit below is 7 and not 3. A reader expecting
# a count-from-here reads it as the latter.
#
# THE OP CANNOT SEE THE ARITY, which is `10_undef_arity.t`'s finding met
# again. Measured, the two-argument and three-argument calls both emit an
# op named `index`, differing only in the operand count `sK/2` against
# `sK/3`, and `opsOf` collects names. So this file asserts on OUTPUT: the
# three printed numbers are what a parser that dropped or reordered the
# third argument changes.
#
# THE BRACKETS ARE NOT DECORATION. Each printed number is wrapped in
# `[...]` for the reason this tier's README records: `-1` and `4` are
# fine, but a result that printed empty would leave a line ending in a
# space, and the repo's `trailing-whitespace` hook strips that out of the
# `--- expect output` block after staging, leaving a committed
# expectation one byte short of what perl prints.
#
# THE TOKEN FACTS COUNT WHAT THE SOURCE HAS ONE OF. There are THREE words
# spelled `index` here and the grammar admits only `one` and `no`, so the
# construct cannot be counted directly -- the same limit `09_named_unary.t`
# records for `defined`. `one operator whose text is "+"` pins the single
# arithmetic in the file, which is the expression computing the resume
# position; a parser that folded it or a lexer that split it changes the
# count. `no word whose text is "rindex"` is checked against the whole
# source including `$ENV{X}`: the word appears nowhere, and it is the
# assertion worth making because `rindex` is `index`'s mirror, emits an op
# named `rindex` that NO tier in this corpus claims, and a file that
# reached for it to widen the pair would fail the dependency lint.

--- source
my $s = $ENV{X} // "hello world";
my $first = index($s, "o");
my $next = index($s, "o", $first + 1);
my $none = index($s, "z");
print "[$first][$next][$none]\n";

--- expect output
[4][7][-1]

--- expect parses

--- expect tokens
one operator whose text is "+"
