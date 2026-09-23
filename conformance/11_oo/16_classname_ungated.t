#!perl
# Without the `class` feature, `__CLASS__` is a FILEHANDLE, and
# `print __CLASS__;` prints `$_` to it.
#
# TIER 11 oo
# INTRODUCES nothing of its own
# USES nothing from a later tier
# STATUS refuses as of this file. Issue 01a0cc04-339f-749a-a630-6f21bf45b60a. Refusal trailing_tokens.
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
# THE NEGATIVE FACT IS THE HALF THAT DISCRIMINATES, and a first draft
# omitted it. `one word whose text is "__CLASS__"` is true under BOTH readings
# -- the keyword lexes as one word whether it is a keyword or a method
# name -- so on its own it says nothing about which reading applies.
#
# What separates them is that perl reads this as a METHOD CALL. A method
# call is spelled with an arrow when written out, and this source
# contains no arrow anywhere -- so a lexer that produced one would have
# manufactured it. `10_isa_infix.t` and `10_io/07_say.t` both guard
# their equivalent claim exactly this way; this file cited them as models
# and left out the half that does the work.

--- source
print __CLASS__;
print "after\n";

--- expect output
after

--- expect parses

--- expect tokens
one word whose text is "__CLASS__"
no operator whose text is "->"
