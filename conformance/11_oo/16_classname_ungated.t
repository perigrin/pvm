#!perl
# Without the `class` feature, `__CLASS__` is a FILEHANDLE, and
# `print __CLASS__;` prints `$_` to it.
#
# TIER 11 oo
# INTRODUCES nothing of its own
# USES nothing from a later tier
# STATUS parses as of `01a0cf3e`. `print __CLASS__;` refused
# because `__CLASS__` is all-caps and was taken into `print`'s
# filehandle slot -- the same bug `10_compile_tokens.t` records,
# reached through a different keyword.
#
# This is the `isa` finding in a different keyword. `10_isa_infix.t`
# records that `$o isa Foo` without the feature parses as
# `print {$o} isa(Foo)` -- a filehandle print. `__CLASS__` does the same
# thing one step more directly.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Deparse -e 'print __CLASS__;'
#   print __CLASS__ $_;
#
#   $ perl -MO=Concise -e 'print __CLASS__;'
#   ... rv2gv sKR/1 ...
#       gv[*__CLASS__] s ...
#
# `rv2gv` over `*__CLASS__` -- a GLOB. Perl reads the bareword as a
# filehandle name, `$_` as the thing to print, and the statement as a
# print to a handle nobody opened. It compiles clean, prints nothing, and
# exits 0.
#
# THAT SILENCE IS THE HAZARD. `state` and `field` die at runtime, loudly
# enough that a test notices. This one SUCCEEDS and produces no output,
# so a program that meant to print a class name prints nothing and says
# nothing about why.
#
# The file makes the silence observable by printing a sentinel after it:
# the `after` arrives, the class name does not, and a parser that read
# `__CLASS__` as a term would have printed something before it.
#
# THE TOKEN LAYER CANNOT SEPARATE THE TWO READINGS, and saying so
# is the honest version of a claim this file got wrong twice.
#
# `one word whose text is "__CLASS__"` is true under BOTH readings --
# the keyword lexes as one word whether it is a keyword or the
# FILEHANDLE this file measures -- so on its own it says nothing
# about which reading applies. That much an earlier draft had right.
#
# What it concluded was wrong. It added
# `no operator whose text is "->"`, reasoning that the ungated
# reading is a method call, that a method call is written with an
# arrow, and that a lexer producing one here would have
# manufactured it. A LEXER CANNOT MANUFACTURE BYTES THAT ARE NOT
# THERE: `scanOperator` matches its table against `l.src` at each
# position, so a token whose text is `->` requires those two bytes
# in the source. The fact could never fail, and a fact that cannot
# fail asserts nothing.
#
# AND THE PREMISE WAS WRONG TOO, which is the sharper half.
# Measured, the ungated reading here is NOT a method call:
#
#   $ perl -MO=Deparse -e 'print __CLASS__;'
#   print __CLASS__ $_;
#
# That is a FILEHANDLE print, as this file's own opening line
# says and its `gv[*__CLASS__]` measurement shows. There is no
# arrow in the deparse to have read the claim off. The fact was
# copied from `14_state_ungated.t`, whose deparse DOES show
# `$n->state`, without checking that it applied here.
#
# So the positive fact stays and does the work it can: the keyword
# lexes as ONE word, not split and not swallowed. THE OUTPUT IS
# THE DISCRIMINATOR, and it has to be -- identical bytes lex
# identically, and only running them tells the readings apart.

--- source
print __CLASS__;
print "after\n";

--- expect output
after

--- expect parses

--- expect tokens
one word whose text is "__CLASS__"
