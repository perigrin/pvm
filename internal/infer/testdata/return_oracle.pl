# ABOUTME: Calls a module's subs with their oracle call cases and records each returned value's lattice type.
# ABOUTME: The observations are observe_types.pl's, emitted as its JSON on exit.

use v5.36;
no strict 'refs';
use B ();

# A sub calling exit dies instead, so perl does not end with the subs
# after it uncalled.
BEGIN { *CORE::GLOBAL::exit = sub { die "exit\n" } }

# Usage: return_oracle.pl OBSERVER CASES MODULE NAME:CONTEXT:FORM...
#
# CASES is a perl file whose value is a list of sub name => [ [args], ... ].
# CONTEXT is how each case is called, scalar or list. FORM is sub, a call
# of MODULE::NAME, or method, a call of the method MODULE->can(NAME) finds
# with MODULE as its invocant. Each call is made inside eval: a value it
# returns is observed under NAME:CONTEXT:FORM:N, for its N arguments, then
# the call itself as the type Returned, so an empty list is seen to return;
# a death is observed as the type Died. Once every sub is called, the
# driver prints "complete" before the observations, so a run perl ended
# early is told from a finished one.
my ( $observer, $cases_file, $module, @subs ) = @ARGV;

# A module's END may set perl's exit status, as Test::Builder's does when
# no plan was declared. This END, defined before any other, runs last.
my $complete;
END { $? = 0 if $complete }

# The subs' own output goes to STDERR, out of the report on STDOUT, and so
# does a module's that dups STDOUT as it loads, as Test::Builder does.
open my $report, '>&', \*STDOUT or die "cannot dup STDOUT: $!\n";
open STDOUT, '>&', \*STDERR or die "cannot point STDOUT at STDERR: $!\n";

require $observer;
( my $file = "$module.pm" ) =~ s{::}{/}g;
require $file;

# A module whose import re-exports a third module's names straight
# into its caller -- FileHandle's import calls
# Exporter::export('Fcntl', $callpkg) for a bare `use`, never defining
# the name in FileHandle:: itself (measured on 5.42.0: FileHandle::
# F_DUPFD is false even after `require FileHandle`) -- leaves such a
# name only in whatever package did a bare `use FileHandle;`. This
# bare prime, done once per module as held_to_perl.pl's does, is where
# a sub-form lookup falls back for a name neither $module:: nor a
# by-name reprime materializes (see the matching comment below).
eval "package Held::OracleBarePrime; use $module; 1";

my %cases = do $cases_file;
die "cannot read call cases $cases_file: ", $@ || $!, "\n" if $@ || !%cases;

for (@subs) {
    my ( $name, $context, $form ) = split /:/;
    my $calls = $cases{$name} or die "$name has no call case and no exemption\n";
    @$calls or die "$name has no call case\n";
    # A sub perl does not define dies when called, which is no death a
    # sub declared never to return is held to.
    my $code;
    if ( $form eq 'method' ) {
        $code = $module->can($name) or die "perl defines no method $name\n";
    }
    else {
        my $primed;
        if ( !callable("${module}::$name") && callable("Held::OracleBarePrime::$name") ) {
            $code = \&{"Held::OracleBarePrime::$name"};
        }
        if ( !$code && !callable("${module}::$name") ) {
            # A name a bare require does not materialize, measured --
            # Fcntl's S_ISWHT, whose sibling S_IS* subs are compiled in
            # regardless -- gets a stub only once it is asked for by
            # name.
            state $n = 0;
            $primed = "Held::OraclePrimeByName@{[$n++]}";
            eval "package $primed; use $module qw($name); 1";
        }
        if ( !$code && callable("${module}::$name") ) {
            $code = \&{"${module}::$name"};
        }
        elsif ( !$code && $primed && callable("${primed}::$name") ) {
            # A module with its own import, like
            # Locale::Maketext::Simple's, installs a glob straight into
            # the caller's package and never into its own (measured):
            # the only place left to find it is the package it was just
            # primed into.
            $code = \&{"${primed}::$name"};
        }
        $code or die "perl defines no sub $name\n";
    }
    for my $i ( keys @$calls ) {
        my @args = @{ $calls->[$i] };
        my $key  = join ':', $name, $context, $form, scalar @args;
        unshift @args, $module if $form eq 'method';
        my @values;
        my $returned =
          $context eq 'list'
          ? eval { @values = $code->(@args); 1 }
          : eval { @values = scalar $code->(@args); 1 };
        if ($returned) {
            ::__observe( $i, $key, $_ ) for @values;
            push our @OBSERVED, [ $i, $key, 'Returned' ];
        }
        else {
            push our @OBSERVED, [ $i, $key, 'Died' ];
        }
    }
}

open STDOUT, '>&', $report or die "cannot restore STDOUT: $!\n";
say 'complete';
$complete = 1;

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
