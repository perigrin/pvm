# ABOUTME: Calls a module's subs with their oracle call cases and records each returned value's lattice type.
# ABOUTME: The observations are observe_types.pl's, emitted as its JSON on exit.

use v5.36;
no strict 'refs';

# Usage: return_oracle.pl OBSERVER CASES MODULE NAME:CONTEXT...
#
# CASES is a perl file whose value is a list of sub name => [ [args], ... ].
# CONTEXT is how each case is called: scalar, list, or none for a sub
# declared never to return, which is called inside eval and must die.
my ( $observer, $cases_file, $module, @subs ) = @ARGV;
require $observer;
( my $file = "$module.pm" ) =~ s{::}{/}g;
require $file;

my %cases = do $cases_file;
die "cannot read call cases $cases_file: ", $@ || $!, "\n" if $@ || !%cases;

for (@subs) {
    my ( $name, $context ) = split /:/;
    my $calls = $cases{$name} or die "$name has no call case and no exemption\n";
    my $code  = \&{"${module}::$name"};
    for my $i ( keys @$calls ) {
        my @args = @{ $calls->[$i] };
        if ( $context eq 'none' ) {
            # A return is observed, and fails against None; a death is not.
            my $value = eval { scalar $code->(@args) };
            ::__observe( $i, $name, $value ) unless $@;
        }
        elsif ( $context eq 'list' ) {
            ::__observe( $i, $name, $_ ) for $code->(@args);
        }
        else {
            ::__observe( $i, $name, scalar $code->(@args) );
        }
    }
}
