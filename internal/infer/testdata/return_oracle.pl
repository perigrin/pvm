# ABOUTME: Calls a module's subs with their oracle call cases and records each returned value's lattice type.
# ABOUTME: The observations are observe_types.pl's, emitted as its JSON on exit.

use v5.36;
no strict 'refs';
use B ();

# A sub calling exit dies instead, so perl does not end with the subs
# after it uncalled.
BEGIN { *CORE::GLOBAL::exit = sub { die "exit\n" } }

# Usage: return_oracle.pl OBSERVER CASES MODULE NAME:CONTEXT...
#
# CASES is a perl file whose value is a list of sub name => [ [args], ... ].
# CONTEXT is how each case is called: scalar, list, or none for a sub
# declared never to return, which is called inside eval and must die. Each
# value is observed under NAME:CONTEXT. Once every sub is called, the
# driver prints "complete" before the observations, so a run perl ended
# early is told from a finished one.
my ( $observer, $cases_file, $module, @subs ) = @ARGV;
require $observer;
( my $file = "$module.pm" ) =~ s{::}{/}g;
require $file;

my %cases = do $cases_file;
die "cannot read call cases $cases_file: ", $@ || $!, "\n" if $@ || !%cases;

# The subs' own output goes to STDERR, out of the report on STDOUT.
open my $report, '>&', \*STDOUT or die "cannot dup STDOUT: $!\n";
open STDOUT, '>&', \*STDERR or die "cannot point STDOUT at STDERR: $!\n";

for (@subs) {
    my ( $name, $context ) = split /:/;
    my $calls = $cases{$name} or die "$name has no call case and no exemption\n";
    @$calls or die "$name has no call case\n";
    # A sub perl does not define dies when called, which is no death a
    # sub declared never to return is held to.
    callable("${module}::$name") or die "perl defines no sub $name\n";
    my $code = \&{"${module}::$name"};
    for my $i ( keys @$calls ) {
        my @args = @{ $calls->[$i] };
        if ( $context eq 'none' ) {
            # A return is observed, and fails against None; a death is not.
            my $value = eval { scalar $code->(@args) };
            ::__observe( $i, "$name:$context", $value ) unless $@;
        }
        elsif ( $context eq 'list' ) {
            ::__observe( $i, "$name:$context", $_ ) for $code->(@args);
        }
        else {
            ::__observe( $i, "$name:$context", scalar $code->(@args) );
        }
    }
}

open STDOUT, '>&', $report or die "cannot restore STDOUT: $!\n";
say 'complete';

# callable is whether a call to the sub named $name runs code, by the rule
# internal/parse/testdata/held_to_perl.pl measures installs with: one that
# is defined, or one that exists undefined in a package with an AUTOLOAD
# to run in its place.
sub callable ($name) {
    return 0 if !exists &$name;
    return 1 if defined &$name;
    my $home = B::svref_2object( \&$name )->GV->STASH->NAME;
    return defined &{"${home}::AUTOLOAD"};
}
