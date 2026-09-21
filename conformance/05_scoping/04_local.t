#!perl
# `local $x` saves the package variable's value and restores it when the
# enclosing block ends, so the old value is back on the far side.
#
# TIER 05 scoping
# INTRODUCES dynamic scope
# USES nothing from a later tier
#
# This is the one form here that is not lexical. `my`, `our` and `state`
# all bind a NAME for a region of source; `local` binds a VALUE for a
# region of TIME. Nothing about the name changes -- `$x` is the same
# package scalar throughout -- and what the block changes is what that
# scalar holds while the block is running.
#
# The probe is the third print. Line 4 shows the localised 2; line 6
# shows 1 again, restored by the block exit, and no statement put it
# back. That restoration is the entire construct.
#
# Like `our`, it emits no op of its own: `gvsv` and `sassign` are tier
# 02's. The difference from `our` is a flag, `gvsv[*x] s/LVINTRO` against
# `s/OURINTR`, which the ops list does not record. The `enterloop` and
# `leaveloop` are the bare block's, this tier's own, and here they are
# load-bearing rather than incidental -- `leaveloop` is where the restore
# happens.
#
# No `use strict`, deliberately. Under strict, `local $x` on an
# undeclared package variable is a compile error ("Global symbol "$x"
# requires explicit package name") and would need an `our $x` first,
# making the file a test of two constructs. The corpus does not run under
# strict unless a file asks for it, so the plain assignment on line 1 is
# what brings `$main::x` into existence.
#
# MEASURED perl 5.42.0:
#
#   $ perl -w -e '$x = 1; { local $x = 2; print "$x\n"; } print "$x\n"'
#   2
#   1

--- source
$x = 1;
{
  local $x = 2;
  print "$x\n";
}
print "$x\n";

--- expect output
2
1

--- expect parses
