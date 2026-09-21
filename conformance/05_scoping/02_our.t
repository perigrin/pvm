#!perl
# `our $x` declares a lexical ALIAS to the package variable `$main::x`,
# so the value survives being shadowed by a `my $x` and is still
# reachable by its qualified name from inside the shadow.
#
# TIER 05 scoping
# INTRODUCES package-variable declaration
# USES nothing from a later tier
#
# This is the file that says what `our` is not. `our $x = 1` looks like
# `my $x = 1` and is nothing like it: `my` makes a pad slot, `our` makes a
# lexically-scoped NAME for a symbol-table entry that already exists. The
# probe is the shadow -- inside the block `$x` is the `my`, and
# `$main::x` is the `our`'s referent, unshadowed and still 1.
#
# It emits no op of its own. `gvsv` and `sassign` are tier 02's, claimed
# there with the package scalar. What distinguishes `our` from a bare
# assignment is a flag -- `gvsv[*x] s/OURINTR` against plain `gvsv[*x] s`
# -- and the ops list does not record flags. The output is what
# distinguishes the constructs here, which is why the file exists even
# though the op stream has nothing new to show.
#
# No `use strict`, so the qualified `$main::x` needs no declaration of its
# own; the `our` on line 1 is what makes the unqualified `$x` on line 6
# refer to the same scalar.
#
# MEASURED perl 5.42.0:
#
#   $ perl -w -e 'our $x = 1; { my $x = 2; print "$x $main::x\n"; } print "$x\n"'
#   2 1
#   1

--- source
our $x = 1;
{
  my $x = 2;
  print "$x $main::x\n";
}
print "$x\n";

--- expect output
2 1
1

--- expect parses
