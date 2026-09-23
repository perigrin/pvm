#!perl
# Every construct this tier introduces, in one body, each adjacent to
# another, and paired with the declared prerequisite.
#
# TIER 12 packages
# INTRODUCES nothing of its own
# USES nothing from a later tier
#
# The tier's other files are one construct each, which is what makes them
# diagnosable: when `06_require_bareword.t` refuses, `require` is the only
# construct present. That same property is why a corpus of such files
# cannot reach an ADJACENCY bug -- a parser that handles every construct
# alone and mis-handles a pair goes green over the pair. Tier 11 has the
# case the design was written for: `class Foo { ADJUST { 1 } }` parses and
# `class Foo { ADJUST { 1 } method m { 2 } }` does not.
#
# The pairs this file puts next to each other:
#
#   an importing `use` immediately followed by an empty-list `use`
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
# EVERY PINNED LINE IS FALSIFIABLE BY DELETING A STATEMENT, which is what
# a tier of no-op constructs has instead of an op check. Measured, each
# deletion changes the output:
#
#   drop `package Greet;`    `Greet::hello` is undefined and the program dies
#   drop `package Louder {`  likewise for `Louder::shout`
#   drop `use POSIX;`        `use:` reads `no` where it read `yes`
#   drop `use Fcntl ();`     nothing to drop -- and see below
#   drop either `require`    `inc:` loses a `yes`
#   drop `Greet->import`     the first line of output goes
#
# `use Fcntl ();` is the one whose deletion does NOT change the output,
# because its whole claim is that it imports nothing: it is pinned by
# CONTRAST with the `use POSIX;` above it. `LOCK_EX` is chosen over
# `O_RDONLY` for exactly that reason -- POSIX exports `O_RDONLY` too, so
# pinning it would have read `yes` whether or not `Fcntl` was ever loaded,
# and the pair would have measured one statement twice. Measured on 5.42.0:
# POSIX exports `O_RDONLY` and does not export `LOCK_EX`.
#
# MEASURED perl 5.42.0:
#
#   $ perl conformance/12_packages/00_adjacency.t
#   begin
#   import tag
#   hello HELLO
#   inc: yesyes
#   use: yesno
#   pkg main line 300 file adj tag c
#   end
#
# THE COMPILE-PHASE SLICE WRAPS THE WHOLE BODY, which is the adjacency
# claim it makes. `BEGIN` is written AFTER `END` in the source and runs
# first; `END` runs after the last statement. Neither emits an op, so
# the ordering is the only evidence either exists, and a parser that
# treated them as ordinary blocks would print them where they appear.
#
# `pkg main line 300 file adj` is four compile-time constructs in one
# statement. The `#line` directive above it is a `#` at column zero that
# is NOT a comment: it rewrites `__LINE__` and `__FILE__` for everything
# below, which is why the line reports 300 and `adj` rather than its
# real position. `__FILE__` is only pinnable at all because of that --
# without the directive it reports the runner's temp path.
#
# `TAG` is the bareword `use constant` installed, and it prints `c`
# rather than `TAG`, which is what separates an installed sub from a
# bareword string.
#
# `expect output` is written before `expect parses` rather than last. The
# blank line after it is what carries the output's own trailing newline,
# and a blank line at END of file is what `end-of-file-fixer` strips.

--- source
use POSIX;
use Fcntl ();
use constant TAG => "c";
END { print "end\n" }
BEGIN { print "begin\n" }
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
print "use: ", ($main::{"floor"} ? "yes" : "no"), ($main::{"LOCK_EX"} ? "yes" : "no"), "\n";
#line 300 "adj"
print "pkg ", __PACKAGE__, " line ", __LINE__, " file ", __FILE__, " tag ", TAG, "\n";

--- expect output
begin
import tag
hello HELLO
inc: yesyes
use: yesno
pkg main line 300 file adj tag c
end

--- expect parses
