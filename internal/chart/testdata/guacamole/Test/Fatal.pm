package Test::Fatal;
# ABOUTME: Load-only stand-in for Test::Fatal, which three Guacamole tests import and none call.
# ABOUTME: Lets extract.pl run the suite on a perl without Test::Fatal installed; exception() refuses to run.

use strict;
use warnings;

use Exporter "import";
our @EXPORT_OK = qw(exception);
our @EXPORT    = qw(exception);

sub exception (&) { die "Test::Fatal stand-in: exception() is not recorded\n" }

1;
