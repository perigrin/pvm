#!perl
# Without the `class` feature, `field $x` is a METHOD CALL on `$x` --
# `$x->field` -- and `class Foo { }` is an indirect method call whose
# block is an anonymous hash.
#
# TIER 11 oo
# INTRODUCES nothing of its own
# USES eval from 06_control and method dispatch from this tier
#
# `07_class_empty.t`, `08_class_field_method.t` and `09_class_adjust.t`
# are the enabled halves and each carries the feature at the top. None of
# them pins what the same bytes mean without it.
#
# MEASURED perl 5.42.0:
#
#   $ perl -MO=Deparse -e 'my $x; field $x;'
#   my $x;
#   $x->field;
#
#   $ perl -MO=Deparse -e 'class Foo { }'
#   'Foo'->class({});
#
# Both compile clean. `field` becomes an ordinary method call and `class`
# becomes an INDIRECT method call whose brace group is read as a hashref
# constructor -- the same indirect-object shape `06_indirect_new.t`
# measures for `new Foo`.
#
# THE ASYMMETRY IS THE PART WORTH RECORDING, because it decides what a
# file can claim. Measured:
#
#   class Foo { }                      compiles -- indirect method call
#   class Foo { field $x; method m {} }  SYNTAX ERROR near "; method "
#
# So the silent reparse applies only to the EMPTY form. A populated class
# body is a hard error without the feature, which means
# `08_class_field_method.t` has no silent-reparse partner to write and
# `07_class_empty.t` sits exactly in the dangerous window.
#
# This file pins the `field` half, because it is observable without the
# indirect-object spelling: the method call dies at runtime on an
# undefined invocant, which the eval traps so STDOUT stays pinnable. The
# `class` half is recorded above rather than written, since
# `'Foo'->class({})` would need a `class` sub in scope to produce output
# and that sub would then be the subject rather than the reparse.
#
# THE TOKEN LAYER CANNOT SEPARATE THE TWO READINGS, and saying so
# is the honest version of a claim this file got wrong twice.
#
# `one word whose text is "field"` is true under BOTH readings --
# the keyword lexes as one word whether it is a keyword or a
# method name -- so on its own it says nothing about which reading
# applies. That much an earlier draft had right.
#
# What it concluded was wrong. It added
# `no operator whose text is "->"`, reasoning that a method call
# is written with an arrow and this source contains none, so a
# lexer producing one would have manufactured it. A LEXER CANNOT
# MANUFACTURE BYTES THAT ARE NOT THERE: `scanOperator` matches its
# table against `l.src` at each position, so a token whose text is
# `->` requires those two bytes in the source. The fact could
# never fail, and a fact that cannot fail asserts nothing.
#
# The arrow came from perl's DEPARSE of the ungated reading,
# quoted above -- a claim read off `B::Deparse` and written as if
# it were a claim about tokens. The reparse is the PARSER
# reinterpreting the same tokens, not the lexer emitting different
# ones, which is exactly why the two readings are dangerous.
#
# So the positive fact stays and does the work it can: the keyword
# lexes as ONE word, not split and not swallowed. THE OUTPUT IS
# THE DISCRIMINATOR, and it has to be -- identical bytes lex
# identically, and only running them tells the readings apart.

--- source
my $x;
my $r = eval { field $x; 1 };
print defined $r ? "ran" : "died", "\n";

--- expect output
died

--- expect parses

--- expect tokens
one word whose text is "field"
