#!perl
# `require $m;` takes the filename from a variable, and emits the same
# `require` op as the bareword form.
#
# TIER 12 packages
# INTRODUCES require with a runtime operand
# USES nothing from a later tier
#
# This is the file that shows the bareword form was sugar. `require strict`
# emits `const[PV "strict.pm"] s/BARE` then `require`; this emits `padsv`
# then `require`. One op, two sources, and only the operand differs --
# which is also why `require Foo::Bar` and `require "Foo/Bar.pm"` are the
# same statement and `require $m` with `$m` holding `"Foo::Bar"` is not.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $m = "warnings.pm"; require $m; print +($INC{$m} ? "yes" : "no"), "\n"'
#   yes

--- source
my $m = "warnings.pm";
require $m;
print "loaded ", ($INC{$m} ? "yes" : "no"), "\n";

--- expect output
loaded yes

--- expect parses
