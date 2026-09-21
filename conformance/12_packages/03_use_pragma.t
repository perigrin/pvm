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
# WHAT THIS FILE DOES NOT REACH, stated plainly because the pin does not
# say it. Measured: delete both `use` lines and this program still prints
# `ok`. So the output pin establishes that the lines LEX and the statement
# is delimited, and it establishes nothing about the load or the import.
# A pragma cannot do better -- its effect is lexical and compile-time, and
# `$^H` and `${^WARNING_BITS}` consulted at run time report the CALLER's
# scope, measured OFF inside the very file that turned them on.
#
# `04_use_import.t` is the file that reaches the rest, through a module
# with an exporter and the symbol table. This one is kept because `use`
# with a pragma is the common spelling and a lexer that mis-delimited it
# would fail here.
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
