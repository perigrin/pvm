#!perl
# `eval STRING` is this tier's subject with no delimiter at all: a region
# of PERL that the outer compiler never reads, handed to the lexer at
# RUNTIME and assembled from a value.
#
# TIER 14 recursive
# INTRODUCES eval STRING -- the undelimited re-entry
# USES nothing from a later tier
#
# Every other file in this tier hands back a region the lexer at least had
# to FIND: `s///e` has three delimiters, `(?{ })` has a brace pair inside
# a pattern, `qr//` freezes one. This has none. The operand is an ordinary
# string expression, and what it holds is not known until the expression
# has been evaluated -- which is why the tier's README calls this form "the
# same re-entry with no delimiter at all".
#
# WHY HERE AND NOT 06_control, since `eval` is one keyword. Because the
# two forms share no op. MEASURED perl 5.42.0:
#
#   $ perl -MO=Concise,-exec -e 'my $r = eval { 1 };'   entertry / leavetry
#   $ perl -MO=Concise,-exec -e 'my $r = eval "1";'     entereval
#
# The BLOCK form marks a frame to unwind to and compiles nothing new; it
# is a control transfer, it lives in `06_control/14_eval_block.t`, and
# `entertry`/`leavetry` are claimed there. This form compiles Perl the
# compiler never saw, which is exactly what `entereval` names and exactly
# what this tier is for. `entereval` is already in this tier's INTRODUCES,
# claimed for `s///ee`, so this file adds a second construct behind an
# existing claim rather than a new op.
#
# MEASURED, with X unset so the interpolated operand is a runtime value:
#
#   $ perl -MO=Concise,-exec -e 'my $n = $ENV{X} // 2;
#     my $r = eval "$n + 1"; print "$r\n";'
#   8  <0> padsv[$n:1,3] s
#   9  <+> multiconcat(" + 1",-1,4)[t5] sK/STRINGIFY
#   a  <1> entereval[t256] sK/1
#
# THERE IS NO `add` IN THAT STREAM, and that absence is the whole file.
# The program adds two numbers and emits no addition op, because the `+`
# is a character in a string at compile time and becomes an operator only
# inside the second compilation `entereval` triggers. `multiconcat` at 9
# BUILDS the program text; `entereval` at a compiles and runs it. Compare
# `01_subst_eval.t`, where `/e` puts the replacement's `add` in the outer
# stream: one `e` compiles the region with the program, two `e`s and a
# bare `eval` do not.
#
# The string is interpolated rather than constant on purpose, the same
# choice every file in this tier makes for the same reason: `eval "1 + 1"`
# still emits `entereval`, but a parser could constant-fold the operand
# and the file would no longer measure that the operand is a VALUE. With
# `$n` in it there is nothing to fold.
#
# The behavioural pin:
#
#   $ perl conformance/14_recursive/06_eval_string.t
#   3
#
# `3`, not `2 + 1`. A parser that treated the string as an ordinary string
# expression -- never re-entering -- would assign the text and print it
# verbatim, which is a one-token difference in the tree and a five-byte
# difference in the output.
#
# THE TOKEN CLAIM IS WHERE THE RE-ENTRY HAS NOT YET HAPPENED, stated the
# way `01_subst_eval.t` states it. At the lexical layer the operand is ONE
# string literal and the `+` inside it is not an operator token. A lexer
# that emitted that `+` as an operator would have lexed the program text
# instead of delimiting it -- the error this tier exists to catch. `$n` is
# declared on its own line and appears as a variable token there; only the
# `+` is unique to the eval'd region, so it is the one the `no` claim
# names.

--- source
my $n = $ENV{X} // 2;
my $r = eval "$n + 1";
print "$r\n";

--- expect output
3

--- expect parses

--- expect tokens
one string literal whose text is "\"$n + 1\""
no operator whose text is "+"
