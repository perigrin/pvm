package Guacamole::Test;
# ABOUTME: Recording stand-in for Guacamole's own Guacamole::Test: writes each parses/parsent argument, parses nothing.
# ABOUTME: extract.pl puts this directory ahead of Guacamole's lib on @INC and names the record file in GUACAMOLE_RECORD.

use strict;
use warnings;

use Test::More;
use JSON::PP ();

use Exporter "import";

our @EXPORT = qw(
    parses
    parses_as
    parsent
    done_testing
);

my $json = JSON::PP->new->canonical->ascii;

sub _record {
    my ( $kind, $text ) = @_;
    my $path = $ENV{GUACAMOLE_RECORD}
        or die "GUACAMOLE_RECORD is not set\n";
    open my $fh, '>>', $path or die "open $path: $!\n";
    print {$fh} $json->encode( { kind => $kind, text => $text, test => $0 } ), "\n";
    close $fh or die "close $path: $!\n";
    local $Test::Builder::Level = $Test::Builder::Level + 2;
    pass("recorded $kind");
}

sub parses     { _record( parses  => $_[0] ); return [] }
sub parses_as  { _record( parses  => $_[0] ); return }
sub parsent    { _record( parsent => $_[0] ); return }

1;
