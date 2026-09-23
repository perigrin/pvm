#!perl
# `use constant` installs a BAREWORD that afterwards parses as a
# zero-argument sub call, so a construct changes the grammar for
# everything below it.
#
# TIER 12 packages
# INTRODUCES nothing of its own
# USES the use statement from 03_use_pragma.t
#
# MEASURED perl 5.42.0. The op stream cannot see it:
#
#   $ perl -MO=Concise,-exec -e 'use constant PI => 3; print PI, "\n"'
#   ... const[IV 3] ... print ...
#
# A `const`, exactly as `print 3` gives. The pragma runs at compile time,
# installs a sub, and the optimiser inlines the call -- so by the time
# the optree exists there is no trace of either the pragma or the call.
#
# THE BAREWORD IS A SUB, verified rather than assumed:
#
#   $ perl -e 'use constant PI => 3; print defined(&PI) ? "yes" : "no"'
#   yes
#
# So `PI` below the pragma is a CALL, not a string. That is what makes
# this a parsing claim rather than a fact about constants: the same
# bareword above the pragma would be a string under `no strict`, and a
# syntax error under `use strict`. A parser that read it as a bareword
# string would print `PI` where this file prints `3`.
#
# `PI + 1` is the discriminator. A bareword string would numify to 0 and
# the sum would be 1; the installed sub returns 3 and the sum is 4. One
# digit separates the two readings, and it is the reason the file adds
# rather than just printing.
#
# The token facts count `constant` and FORBID a string `PI`. The count
# of `PI` itself would be two -- once in the pragma, once in the call --
# which is a claim about the file's shape rather than its construct. What
# matters is that neither occurrence is a STRING: the fat comma autoquotes
# the left side into one, but that happens in the pragma's arguments, not
# in the token stream.
#
# The fat comma is incidental here and already claimed by tier 02, but
# worth noting for a reader: `PI => 3` autoquotes the left side, so the
# pragma receives the string `"PI"` and installs a sub by that name.

--- source
use constant PI => 3;
print PI + 1, "\n";

--- expect output
4

--- expect parses

--- expect tokens
one word whose text is "constant"
no string literal whose text is "PI"
