#!perl
# `require strict;` is the one construct in this tier that reaches run time,
# and the one op this tier introduces.
#
# TIER 12 packages
# INTRODUCES the require statement
# USES nothing from a later tier
#
# The bareword-to-filename rewrite is the compiler's: by the time the op
# runs, `strict` has become the string `"strict.pm"`, emitted as
# `const[PV "strict.pm"] s/BARE` and consumed by `require`. So `require`
# takes a filename, never a package name, and the bareword spelling is
# sugar the op stream cannot see.
#
# `strict` is chosen because it is core, prints nothing, and its version
# does not reach the output. The file observes the load through `%INC`
# rather than through anything the module itself emits.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'require strict; print +($INC{"strict.pm"} ? "yes" : "no"), "\n"'
#   yes

--- source
require strict;
print "loaded ", ($INC{"strict.pm"} ? "yes" : "no"), "\n";

--- expect output
loaded yes

--- expect parses
