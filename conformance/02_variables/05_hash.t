#!perl
# A hash is named with `%` and subscripted with braces; `$h{a}` is one
# element of it, and the bareword key needs no quotes.
#
# TIER 02 variables
# INTRODUCES hash variable
# USES nothing from a later tier
#
# `scalar(keys %h)` is here so the file can observe the hash as a whole
# without printing it: hash order is not guaranteed, so a file that printed
# the values would be pinning an expectation perl does not promise. The
# count is stable and the element lookup is what the file is about.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my %h = (a => 1, b => 2); print scalar(keys %h), "\n"; print $h{a}, "\n"'
#   2
#   1
#
# `$h{a}` emits no `helem`. It compiles to `multideref($h{"a"})`, the op
# that swallows most element access in this tier; `helem` survives only
# where the subscript is an expression, which is `06_hash_element_expr.t`.
# `keys` likewise emits no op of its own -- it becomes a FLAG on the
# `padhv`, `sM/KEYS`.

--- source
my %h = (a => 1, b => 2);
print scalar(keys %h), "\n";
print $h{a}, "\n";

--- expect output
2
1

--- expect parses
