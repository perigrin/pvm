#!perl
# `use strict;` is a BEGIN block, finished before there is a runtime optree,
# and it emits no op.
#
# TIER 12 packages
# INTRODUCES the use statement
# USES nothing from a later tier
#
# The op stream for this file is the stream for `my $x = "ok"; print "$x\n"`
# alone -- tier 01 and tier 05 -- differing only in the feature bits printed
# on `nextstate`, which are a field rather than an op. A tier derived from
# ops would conclude `use` is not in the language.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'use strict; use warnings; my $x = "ok"; print "$x\n"'
#   ok

--- source
use strict;
use warnings;
my $x = "ok";
print "$x\n";

--- expect output
ok

--- expect parses
