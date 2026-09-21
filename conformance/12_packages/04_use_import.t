#!perl
# `use POSIX;` loads the file AND calls `POSIX->import`, which is the half
# of `use` that `require` does not have.
#
# TIER 12 packages
# INTRODUCES the use statement with an import
# USES nothing from a later tier
#
# This is the file that makes `use` FALSIFIABLE. `03_use_pragma.t` pins
# `ok`, and a copy of that file with both `use` lines deleted prints `ok`
# too -- measured -- so it asserts that the lexer did not choke on the
# lines, and nothing more. A pragma cannot do better: its whole effect is
# lexical and compile-time, and `$^H` and `${^WARNING_BITS}` read the
# caller's scope rather than the file's when consulted at run time.
#
# A module with an exporter can. `use POSIX;` puts `floor` into `main::`,
# and deleting the `use` line changes both pinned lines -- so the output is
# a measurement of the statement rather than of its absence of syntax
# errors.
#
# It is the PAIR with `05_use_empty_list.t` that carries the tier's claim.
# Same module, same two observations, one statement apart:
#
#   use POSIX;      loaded yes    imported yes
#   use POSIX ();   loaded yes    imported no
#
# The first column is `require`'s half of `use` and the second is
# `import`'s. Neither emits an op -- the op stream for this file and for
# `05` is byte-identical, and identical again to the same two `print`
# statements alone -- so the symbol table is where the difference is
# visible at all. That is the tier's defining difficulty stated as a
# measurement: `use` is real, and no optree records it.
#
# POSIX is core, its import list is not versioned in any way this file
# observes, and nothing here depends on what it prints, because it prints
# nothing.
#
# MEASURED perl 5.42.0:
#
#   $ perl -e 'use POSIX; print +($INC{"POSIX.pm"} ? "yes" : "no"), ($main::{"floor"} ? "yes" : "no"), "\n"'
#   yesyes

--- source
use POSIX;
print "POSIX loaded: ", ($INC{"POSIX.pm"} ? "yes" : "no"), "\n";
print "floor imported: ", ($main::{"floor"} ? "yes" : "no"), "\n";

--- expect output
POSIX loaded: yes
floor imported: yes

--- expect parses
