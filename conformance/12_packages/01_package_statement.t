#!perl
# `package Foo;` switches the current package for the rest of the enclosing
# block or file, and emits no runtime op.
#
# TIER 12 packages
# INTRODUCES the package statement
# USES nothing from a later tier
#
# The whole effect is on the compiler. The only trace in the op stream is
# the package name inside `nextstate(Greet ...)`, which is a field of an op
# tier 01 already claims, not an op of this tier's own.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'package Greet; sub hello { "hello" } package main; print Greet::hello(), "\n"'
#   hello

--- source
package Greet;
sub hello { return "hello" }
package main;
print Greet::hello(), "\n";

--- expect output
hello

--- expect parses
