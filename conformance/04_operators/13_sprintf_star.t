#!perl
# `sprintf`'s first argument is a FORMAT, and `%*d` makes the format
# consume an extra argument to supply its own width -- so how many
# arguments a call takes is decided by the format's CONTENTS and not by
# its syntax.
#
# TIER 04 operators
# INTRODUCES sprintf
# USES my, print, %ENV, //, string interpolation
# MEASURED perl 5.42.0
#
# `sprintf` FOLDS AWAY on constant arguments, exactly as `chr` and `ord`
# do and unlike `index` and `substr`. Measured:
#
#   $ perl -MO=Concise,-exec -e 'my $x = sprintf("%03d", 5); print $x'
#   3  <$> const[PV "005"] s/FOLD
#
# No `sprintf` op at all: the formatting happened at compile time and the
# result arrived as a tier-01 string literal. A file here written with a
# literal argument would declare `sprintf` in its tier and emit nothing,
# and the lint would pass it -- the lint checks that a file's ops are
# CLAIMED, never that a claimed op is USED. `$ENV{X} // 5` is what makes
# the op exist. This tier's README states that constraint for its
# arithmetic; it holds just as hard for its named operators.
#
# THE STAR IS THE PARSE, and it is why this file is `sprintf`'s rather
# than a formatting demonstration. A parser reading `sprintf($fmt, @args)`
# as "one format, then a list" is right about the SYNTAX and cannot say
# how the list is consumed, because the consumption rule lives inside a
# string:
#
#   $ perl -e 'printf "%*d|", 5, 42'
#      42|
#   $ perl -e 'printf "%-*d|", 5, 42'
#   42   |
#
# The same two arguments, read in the same order, and the `-` flag
# flips which side the padding lands on. `%*d` takes its width from the
# argument list, so ONE conversion has eaten TWO arguments -- a fact no
# amount of looking at the call's parentheses reveals. This is the same
# class of finding as `09_named_unary.t`'s: where an argument stops is
# not where the source suggests.
#
# THE FORMAT IS AN ORDINARY STRING TO THE LEXER, which is the reason the
# construct needs a corpus file at all. `"%*d"` is one `string literal`
# by the glossary, indistinguishable from `"abc"`, and nothing downstream
# of the lexer can tell that it will change the call's arity. A `printf`
# spelling would say the same thing about arity and drag in tier 10's
# filehandle question, so the `s` form is the one written here.
#
# ZERO PADDING IS A FLAG AND NOT A NUMBER. `%03d` is flag `0`, width `3`,
# conversion `d`; `%3d` pads with spaces. Both are measured below through
# the output rather than asserted, because the whole format language is
# invisible to the optree.
#
# AND THE OPTREE IS BLIND TO THE STAR TOO, which was not expected and is
# the measurement worth recording. The three calls emit three ops, all
# named `sprintf`, and the flag that looks like an argument count is the
# SAME on all three:
#
#   b  <@> sprintf[t4] sK/2      <- "%03d", $n          two children
#   i  <@> sprintf[t6] sK/2      <- "%*d",  $n, $n      THREE children
#
# `/2` is not an arity. The operands are the ops BETWEEN the `pushmark`
# and the `sprintf`, and there are two in the first call and three in the
# second -- but `-exec` prints them as a flat sequence, so counting them
# means knowing where the pushmark was. `opsOf` collects names and
# dedupes, so it reports `sprintf` once and can say nothing about either
# number. That is why this file asserts on output: the star's whole
# effect is a value.
#
# NO `printf` ANYWHERE, and the negative fact says so. `printf` emits an
# op named `prtf` that NO tier in this corpus claims, so a file reaching
# for it to compare the two spellings would fail the dependency lint --
# correctly, and for the same reason `12_index_sentinel.t` keeps `rindex`
# out. Checked against the whole source, `$ENV{X}` included: the text
# `printf` appears nowhere, and `print` is a different token, so a lexer
# that ran the two words together would break this claim, which is what
# it is for.
#
# The positive fact counts the ONE `%` operator in the source. Every
# other `%` here is inside a string literal and belongs to the format, so
# a lexer that scanned into the quoted text -- the failure this corpus
# exists to catch in `01_literals` -- would find four and fail.
# `01_arithmetic.t` asserts the same text for the modulo operator; here
# the same character is a format directive three times over, and only
# once an operator.

--- source
my $n = $ENV{X} // 5;
my $zero = sprintf("%03d", $n);
my $star = sprintf("%*d", $n, $n);
my $left = sprintf("%-*d", $n, $n);
print "[$zero][$star][$left][", $n % 3, "]\n";

--- expect output
[005][    5][5    ][2]

--- expect parses

--- expect tokens
one operator whose text is "%"
no word whose text is "printf"
