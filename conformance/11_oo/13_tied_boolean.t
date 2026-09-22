#!perl
# `tied` asks a variable whether it is tied, and answers with the OBJECT
# the tie is bound to or with undef. In boolean position that is the
# whole construct, and it needs no `ref` to observe.
#
# TIER 11 oo
# INTRODUCES the tied builtin
# USES my, print, bless, package, sub, return, %ENV, //, ?:
# MEASURED perl 5.42.0
#
# `tied` is `tie`'s sibling and was unclaimed by every tier alongside it.
# `12_tie_variable.t` carries the binding; this file carries the
# QUESTION, which is a different parse: `tied` is a named unary taking a
# variable, where `tie` is a list operator taking a variable and a list.
#
# WHAT MAKES IT WORTH A FILE IS THAT THE ANSWER IS A VALUE AND NOT A
# FLAG. `tied $x` returns the object `TIESCALAR` blessed -- the same one
# `tie` returned -- so the construct is an accessor wearing a predicate's
# shape. A parser that read it as a boolean builtin gets every program in
# this file right and the next one wrong, which is why this header says
# so and the token fact does not: the boolean form is what the op budget
# permits here, and it is the form the corpus can assert.
#
# MEASURED perl 5.42.0, this file, main program only:
#
#   m  <0> padsv[$x:41,43] sRM
#   n  <1> tied sK/1            <- arity ONE, a named unary
#   o  <|> cond_expr(other->p) lK/1
#   q  <0> padsv[$plain:42,43] sRM
#   r  <1> tied sK/1
#   s  <|> cond_expr(other->t) lK/1
#
# `<1>` is the arity and it is the parse: ONE child, no pushmark, which
# is a named unary and not a list operator. `tie` two lines earlier emits
# `<@> tie vK/2` -- a list op with a mark. Two words a lexer would call
# one family, two different argument grammars, and the op stream says so.
#
# NO `ref` AND NO `defined`, which an earlier draft assumed were both
# required. Measured, `tied $x` in the condition of a ternary is enough:
# an untied variable gives undef, which is false, and a tied one gives a
# blessed reference, which is true. `ref(tied $x)` would add tier 08's
# `ref` and `defined(tied $y)` tier 04's `defined`; both are EARLIER
# tiers and would be legal, and neither is needed. The boolean form is
# the smaller claim and the one this file makes.
#
# BOTH CASES, which is what makes the output falsifying. `$x` is tied and
# `$plain` is not, so the line carries a true answer and a false one:
#
#   $ perl conformance/11_oo/13_tied_boolean.t
#   tu
#
# A parser that read `tied` as always true prints `tt`; one that read it
# as always false prints `uu`. Both wrong readings are one byte from the
# right one, and with only the tied variable neither would be visible --
# the same argument `10_isa_infix.t` makes for its false case.
#
# `$plain = 1` rather than an undef variable, because the question is
# whether the VARIABLE is tied and not whether its value is defined. An
# untied undef would let a parser that compiled `tied` as `defined` pass,
# and `1` is what rules that out: it is defined, it is true, and `tied`
# still answers `u`. That is the distinction the second half of the
# output exists to make.
#
# `$ENV{X} // "Counter"` keeps the class name a RUNTIME value so the
# `tie` cannot fold; `X` is unset when the runner executes the file.
# Same idiom and same reason as `12_tie_variable.t`.
#
# THE TOKEN FACT pins `tied` as ONE word, and the count is real: `tie`
# appears in this source too, on the line above, and `checkTokenFact`
# (`internal/conformance/fact.go:50`) compares the token's FULL text
# rather than a substring, so the two words are counted separately even
# though one is a prefix of the other. That is the property
# `12_tie_variable.t`'s header records and this file is where it is
# exercised: a source holding both, each counted once.
#
# WHAT THE FACT CANNOT SAY is that there are TWO `tied`s, which is the
# file's defining count -- the true case and the false one. The grammar
# admits only `one` and `no`, so `one word whose text is "tied"` would be
# FALSE of a correct lex of this source. The fact asserted is `tie`
# instead, which IS one here, and it is the honest half: it pins that the
# binding and the question reached the token stream as different words.
# `GLOSSARY.md` records the limit; `11_can_chain.t` hits the same one.

--- source
package Counter;
sub TIESCALAR { return bless {}, "Counter" }
sub FETCH { return 42 }
package main;
tie my $x, $ENV{X} // "Counter";
my $plain = 1;
print tied($x) ? "t" : "u", tied($plain) ? "t" : "u", "\n";

--- expect output
tu

--- expect parses

--- expect tokens
one word whose text is "tie"
