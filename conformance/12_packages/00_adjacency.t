#!perl
# Every construct this tier introduces, in one body, each adjacent to
# another, and paired with the declared prerequisite.
#
# TIER 12 packages
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# The tier's other files are one construct each, which is what makes them
# diagnosable: when `05_require_bareword.t` refuses, `require` is the only
# construct present. That same property is why a corpus of such files
# cannot reach an ADJACENCY bug -- a parser that handles every construct
# alone and mis-handles a pair goes green over the pair. Tier 11 has the
# case the design was written for: `class Foo { ADJUST { 1 } }` parses and
# `class Foo { ADJUST { 1 } method m { 2 } }` does not.
#
# The pairs this file puts next to each other:
#
#   a `package NAME;` immediately followed by a `package NAME { }`
#   a bareword `require` immediately followed by an expression `require`
#   a `package main;` immediately followed by a statement
#   an `import` sub DEFINED in one package and CALLED from another
#
# The last is the pairing with `07_subroutines`, this tier's declared
# prerequisite. `import` is the reason `use` is heavier than `require`, and
# it is a plain sub call; a file that declared the dependency without
# crossing the package boundary would assert nothing about it. Pairing with
# tier 11 instead would assert nothing at all -- this tier needs nothing
# from it.
#
# `Greet::hello` and `Louder::shout` are called by their fully qualified
# names, which is the third spelling of the package boundary in this file
# and the only one that appears in the op stream: `gv[*Greet::hello]`. The
# two `package` statements that put them there leave no trace.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/12_packages/00_adjacency.t
#   import tag
#   hello HELLO
#   inc: yesyes
#
# `expect output` is written before `expect parses` rather than last. The
# blank line after it is what carries the output's own trailing newline,
# and a blank line at END of file is what `end-of-file-fixer` strips.

--- source
package Greet;
sub hello { return "hello" }
sub import { print "import $_[1]\n" }
package Louder {
    sub shout { return "HELLO" }
}
package main;
require strict;
my $mod = "warnings.pm";
require $mod;
Greet->import("tag");
print Greet::hello(), " ", Louder::shout(), "\n";
print "inc: ", ($INC{"strict.pm"} ? "yes" : "no"), ($INC{"warnings.pm"} ? "yes" : "no"), "\n";

--- expect output
import tag
hello HELLO
inc: yesyes

--- expect parses
