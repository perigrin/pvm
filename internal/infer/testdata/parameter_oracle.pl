# ABOUTME: Observes each oracle call case's own arguments' types, before any call is made.
# ABOUTME: The observations are observe_types.pl's, emitted as its JSON on exit.

use v5.36;
no strict 'refs';

# Usage: parameter_oracle.pl OBSERVER CASES MODULE NAME:CONTEXT:FORM...
#
# CASES, CONTEXT and FORM are return_oracle.pl's own (see its header):
# the same call-case file, read the same way, for the same
# NAME:CONTEXT:FORM list the return-type oracle already built from a
# module's declared signatures. Unlike return_oracle.pl, the named sub
# is never called: only its own call cases' arguments are observed,
# each under NAME:CONTEXT:FORM:N:J for its N arguments and the J'th one
# (0-based), so the parameter oracle can hold a declaration's parameter
# types to what perl actually built for each case, independent of
# whatever the sub does with it once called.
my ( $observer, $cases_file, $module, @subs ) = @ARGV;

# A module's own load-time output goes to STDERR, out of the report on
# STDOUT, as return_oracle.pl's does.
open my $report, '>&', \*STDOUT or die "cannot dup STDOUT: $!\n";
open STDOUT, '>&', \*STDERR or die "cannot point STDOUT at STDERR: $!\n";

require $observer;
( my $file = "$module.pm" ) =~ s{::}{/}g;
require $file;

my %cases = do $cases_file;
die "cannot read call cases $cases_file: ", $@ || $!, "\n" if $@ || ( !%cases && @subs );
# INSTANCE names a constructor call for the invocant some method cases
# need (return_oracle.pl's own comment explains why); the parameter
# oracle never builds or calls anything, so it has no use for it here.
delete $cases{INSTANCE};

for (@subs) {
    my ( $name, $context, $form ) = split /:/;
    my $calls = $cases{$name} or die "$name has no call case and no exemption\n";
    @$calls or die "$name has no call case\n";
    for my $i ( keys @$calls ) {
        my @args = @{ $calls->[$i] };
        my $key = join ':', $name, $context, $form, scalar @args;
        ::__observe( $i, "$key:$_", $args[$_] ) for 0 .. $#args;
    }
}

open STDOUT, '>&', $report or die "cannot restore STDOUT: $!\n";
binmode STDOUT;
say 'complete';
