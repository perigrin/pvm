#!perl
# `use POSIX ();` loads the file and suppresses the `import` call; the empty
# parentheses are not an empty argument list but the absence of one.
#
# TIER 12 packages
# INTRODUCES the use statement with an empty import list
# USES nothing from a later tier
#
# This is the pair that separates the two halves of `use`. `use POSIX;`
# would put `floor` into `main::`; `use POSIX ();` leaves it empty while
# `%INC` still records the load. Both halves are compile-time, so neither
# shows in the op stream -- the file has to observe them through the symbol
# table instead.
#
# POSIX is used rather than a pragma because a pragma's whole effect IS its
# import, so there would be nothing left to observe. Nothing here depends on
# POSIX's version or on anything it prints.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'use POSIX (); print +($INC{"POSIX.pm"} ? "yes" : "no"), ($main::{"floor"} ? "yes" : "no"), "\n"'
#   yesno

--- source
use POSIX ();
print "POSIX loaded: ", ($INC{"POSIX.pm"} ? "yes" : "no"), "\n";
print "floor imported: ", ($main::{"floor"} ? "yes" : "no"), "\n";

--- expect output
POSIX loaded: yes
floor imported: no

--- expect parses
