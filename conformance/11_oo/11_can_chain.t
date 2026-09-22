#!perl
# `$o->can("hi")->($o)` is TWO CALLS THROUGH ONE CHAIN, and only the
# first is a method call. The second arrow dereferences a code ref.
#
# TIER 11 oo
# INTRODUCES the ->can(...)->() chain
# USES my, print, bless, package, sub, return, %ENV, //
# MEASURED perl 5.42.0
#
# 132 of T1's 986 files -- 13% -- use `can`, and the corpus named it
# nowhere. It introduces NO OP OF ITS OWN, which is why it is worth a
# file rather than an entry in a list: the construct is entirely a
# question about how the two arrows are parsed, and a parser can get it
# wrong while emitting a plausible op stream.
#
# MEASURED perl 5.42.0, this file:
#
#   $ perl -MO=Concise,-exec conformance/11_oo/11_can_chain.t
#   b  <0> pushmark s
#   c  <0> padsv[$o:3,4] sM      <- the ARGUMENT of the second call
#   d  <0> pushmark s
#   e  <0> padsv[$o:3,4] sM      <- the INVOCANT of the first
#   f  <$> const[PV "hi"] sM
#   g  <.> method_named[PV "can"] s
#   h  <1> entersub[t4] sKRS/TARG    <- the method call
#   i  <1> entersub[t5] lKS/TARG     <- the CODE DEREF
#
# TWO `entersub`, ONE `method_named`. That asymmetry is the whole
# construct. A parser that read BOTH arrows as method calls emits two
# `method_named`; one that read the second arrow as part of the first
# call emits one `entersub`. Both are ordinary-looking op streams and
# both are wrong.
#
# The FLAGS say the same thing again: `h` is `sKRS`, a resolved method
# call, and `i` is `lKS` with no `R` -- a code-ref call. `opsOf`
# collects op NAMES and can see neither the counts nor the flags, which
# is why this file's claims are the output and the tokens.
#
# THE INVOCANT IS PASSED EXPLICITLY, and that is not a stylistic choice.
# `can` returns a bare CODE ref, and calling a code ref passes no
# receiver -- so `$o->can("hi")->()` calls `hi` with an EMPTY `@_`. The
# `$o` inside the second parens is the receiver the method call would
# have supplied and the deref does not. That is the semantic content of
# "only the first is a method call", and it is why the two arrows cannot
# be collapsed into one even though they read as one chain.
#
# `bless {}, $ENV{X} // "Foo"` gives the class name at RUNTIME, so
# nothing folds; `X` is unset when the runner executes the file.
#
# THE TOKEN FACTS, and what they falsify.
#
#   - `one string literal whose text is "hi"` pins the method NAME as a
#     STRING. `can` takes its argument as a string where `$o->hi` writes
#     the same name bare, so the two spellings a method name can wear are
#     pinned in two files: this one and `03_method_named.t`, which
#     asserts the arrow that precedes a bare one. A lexer that read
#     `"hi"` as a quote-like operator -- our lexer gives both the same
#     `Quote` kind and the glossary separates them by TEXT -- fails this.
#
#     The body returns `"called"` and not `"hi"`, which is what makes the
#     count of ONE true. Written the obvious way the source holds two
#     strings spelled `"hi"` -- the method name and the body's return --
#     and the fact then reads `got 2, want 1`. Found by the runner, and
#     recorded because the obvious spelling is the broken one.
#
#   - `one word whose text is "can"` pins that `can` reaches the token
#     stream as a WORD, once. A lexer that folded it into the arrow
#     beside it, or that emitted it twice while splitting the chain,
#     changes the count.
#
# WHAT THE FACTS CANNOT SAY is that there are exactly TWO arrows, which
# is the construct's defining count. The grammar admits only `one` and
# `no`, so `one operator whose text is "->"` would be FALSE of a correct
# lex of this file -- `03_method_named.t` can assert it only because its
# source makes a single call. That limit is `GLOSSARY.md`'s and is
# recorded here rather than worked around.

--- source
package Foo;
sub hi { return "called" }
package main;
my $o = bless {}, $ENV{X} // "Foo";
print $o->can("hi")->($o), "\n";

--- expect output
called

--- expect parses

--- expect tokens
one string literal whose text is "\"hi\""
one word whose text is "can"
