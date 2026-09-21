#!perl
# A substitution mutates its target in place and emits one `subst` op. No
# assignment, no block, no dereference -- which is the measurement behind
# this tier's claim that it could sit at 03.
#
# TIER 09 regex
# INTRODUCES s/// substitution
# USES nothing from a later tier
#
# The replacement `z` arrives as a `const` pushed before the `subst`, not
# as a second argument to it, and there is no `sassign`: `$s` is named in
# the subst op's own flags exactly as the match op names its target.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my $s = "abc"; $s =~ s/a/z/; print "$s\n"'
#   zbc

--- source
my $s = "abc";
$s =~ s/a/z/;
print "$s\n";

--- expect output
zbc

--- expect parses

--- expect tokens
one quote-like operator whose text is "s/a/z/"
