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

--- source
my $x;
my $r = eval { field $x; 1 };
print defined $r ? "ran" : "died", "\n";

--- expect output
died

--- expect parses

--- expect tokens
one word whose text is "field"
