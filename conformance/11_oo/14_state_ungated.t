#!perl
# Without `use feature "state"`, `state $n = 0` is not a syntax error --
# it is a METHOD CALL on an undeclared scalar, used as an lvalue.
#
# TIER 11 oo
# INTRODUCES nothing of its own
# USES eval from 06_control and method dispatch from this tier
#
# `05_scoping/03_state.t` is the enabled half and carries `use feature "state"` at
# the top. Every version-gated file in this corpus does the same, and not
# one pinned what the bytes mean with the feature OFF -- which is half of
# what "version-gated" means, and the half where two previously-found
# bugs lived.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Deparse -e 'sub c { state $n = 0; $n = $n + 1; return $n }'
#   -e syntax OK
#   sub c { $n->state = 0; $n = $n + 1; return $n; }
#
# `-c` reports SYNTAX OK. The bytes compile clean and mean something
# else: `state` is read as a method name, `$n` as the invocant, and the
# whole thing as an lvalue. It fails only when the sub is called:
#
#   $ perl -e 'sub c { state $n = 0; $n = $n + 1; return $n } print c()'
#   Can't call method "state" on an undefined value
#
# THE FAILURE IS AT RUNTIME, which is what makes this worth a file. A
# parser that ignores the pragma produces a program perl accepts, and
# nothing at compile time says otherwise. The `eval` traps the death so
# STDOUT is pinnable -- without it the file would print nothing and exit
# 255, which says nothing about where it failed.
#
# `defined $r` is the discriminator. Under the feature `c()` returns 1
# and the file would print `ran`; ungated it dies inside the eval, `$r`
# is undef, and the file prints `died`. One word, and it separates the
# two readings of identical bytes.
#
# THE NEGATIVE FACT IS THE HALF THAT DISCRIMINATES, and a first draft
# omitted it. `one word whose text is "state"` is true under BOTH readings
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
sub c { state $n = 0; $n = $n + 1; return $n }
my $r = eval { c() };
print defined $r ? "ran" : "died", "\n";

--- expect output
died

--- expect parses

--- expect tokens
one word whose text is "state"
no operator whose text is "->"
