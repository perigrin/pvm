#!perl
# A hash subscript that is itself an element lookup is the case the
# optimiser will not fold into a deref chain.
#
# TIER 02 variables
# INTRODUCES hash element with a computed subscript
# USES nothing from a later tier
#
# `$h{a}` is `multideref`. `$h{$k[0]}` is `helem` over an
# `aelemfast_lex` -- same construct in the source, different op, decided
# entirely by what is inside the braces. This file exists because it is the
# only way to reach `helem` at all, and a tier that claimed `helem` without
# it would be claiming an op no file emits.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'my %h = (a => 1, b => 2); my @k = ("b"); print $h{$k[0]}, "\n"'
#   2

--- source
my %h = (a => 1, b => 2);
my @k = ("b");
print $h{$k[0]}, "\n";

--- expect output
2

--- expect parses
