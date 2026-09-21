#!perl
# `package Foo { ... }` scopes the package switch to a block, and the block
# is a bare block: the op stream is `enterloop`, `stub`, `leaveloop`.
#
# TIER 12 packages
# INTRODUCES the package block form
# USES nothing from a later tier
#
# This is the same construct as `01_package_statement.t` with a scope
# attached, and it is the spelling `class Foo { ... }` also uses -- which is
# why tier 11 could not be ordered after this one. Its three ops belong to
# tiers 05 and 11; nothing in the stream says `package`.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'package Louder { sub shout { "HELLO" } } print Louder::shout(), "\n"'
#   HELLO

--- source
package Louder {
    sub shout { return "HELLO" }
}
print Louder::shout(), "\n";

--- expect output
HELLO

--- expect parses
