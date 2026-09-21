#!perl
# `field` and `method` emit no op of their own. A class body holding both
# compiles to a bare block of `nextstate`s and nothing else.
#
# TIER 11 oo
# INTRODUCES field declaration and method declaration
# USES nothing from a later tier
#
# `field $x = 5;` contributes one `nextstate` and NO store: the
# initialiser is compiled into a field-init tree that is not reachable
# from the main optree. The method body is likewise its own CV, so the
# `methstart` that opens it is not in this file's measured op stream --
# `perl -MO=Concise,-exec` with no sub named dumps the main program alone,
# and naming it takes `-exec,Foo::m`. That is the finding, not an
# omission: the op stream of a file using `class` says almost nothing
# about the class.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'use feature "class"; no warnings "experimental::class"; class Foo { field $x = 5; method m { $x } } my $o = Foo->new; print $o->m, "\n"'
#   5
#
# the whole class body, at file scope:
#
#   3  <{> enterloop(next->5 last->5 redo->4) v
#   4  <;> nextstate(Foo 7 a.pl:4) v:%,us,*,&,{,$,fea=15,0x10
#   5  <2> leaveloop vK/2
#
# and the method's own optree, reachable only by naming it:
#
#   $ perl -MO=Concise,-exec,Foo::m ...
#   Foo::m:
#   1  <+> methstart() v
#   2  <;> nextstate(Foo 10 a.pl:4) v:%,us,*,&,$,fea=15,0x10
#   3  <0> padsv[$x:FAKE:] s
#   4  <1> leavesub[1 ref] K/REFC,1
#
# `methstart` has no classic counterpart: a `sub` doing the same job opens
# with `shift`, tier 02's op. Two spellings of one object system, two
# disjoint prefixes.

--- source
use feature 'class';
no warnings 'experimental::class';
class Foo {
    field $x = 5;
    method m { $x }
}
my $o = Foo->new;
print $o->m, "\n";

--- expect output
5

--- expect parses
