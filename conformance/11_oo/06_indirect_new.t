#!perl
# `new Foo` and `Foo->new` compile to the SAME op stream. Indirect object
# notation is a lexing problem, not a compilation one.
#
# TIER 11 oo
# INTRODUCES indirect object notation for a constructor call
# USES nothing from a later tier
#
# Both spellings emit `pushmark`, `const[PV "Foo"] sM/BARE`,
# `method_named[PV "new"]`, `entersub`. Nothing downstream of the lexer
# can tell them apart, so a corpus that only checks behaviour or the
# optree cannot see the difference at all -- the same shape as the `5e-1`
# case in tier 01, and the reason the format carries token facts.
#
# MEASURED perl 5.42.0 -- the two spellings, diffed:
#
#   $ perl -MO=Concise,-exec -e 'package Foo; sub new { bless {}, shift } package main; my $o = new Foo; print ref $o'
#   $ perl -MO=Concise,-exec -e 'package Foo; sub new { bless {}, shift } package main; my $o = Foo->new; print ref $o'
#
# differ in no op, only in the line numbers:
#
#   3  <0> pushmark s
#   4  <$> const[PV "Foo"] sM/BARE
#   5  <.> method_named[PV "new"] s
#   6  <1> entersub[t2] sKRS/TARG
#
# and the program prints:
#
#   Foo

--- source
package Foo;
sub new { my $c = shift; return bless {}, $c }
package main;
my $o = new Foo;
print ref $o, "\n";

--- expect output
Foo

--- expect parses

--- expect tokens
no operator whose text is "->"
