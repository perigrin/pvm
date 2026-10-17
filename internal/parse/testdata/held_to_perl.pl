# ABOUTME: Measures a module in perl for the tests that hold its .tpm to perl: where it loads from,
# ABOUTME: its subs' prototypes, and what each `use` of it installs in a fresh caller package.

use v5.36;
no strict 'refs';
use B ();
use Scalar::Util qw(refaddr);

# Usage:
#   held_to_perl.pl inc MODULE              prints the file perl loaded MODULE from
#   held_to_perl.pl prototypes MODULE SUB... prints SUB, then "none", "missing" or "proto" and it
#   held_to_perl.pl installs MODULE LIST...  prints each import LIST ("" for a bare use), then
#                                            "installs" and the names it installs, or "dies" and why
# Fields are tab-separated, one line per argument. `installs` also asks for
# every name in the module's @EXPORT and @EXPORT_OK, an `&name` as `name`.
my ( $mode, $module, @args ) = @ARGV;
( my $file = "$module.pm" ) =~ s{::}{/}g;
eval { require $file; 1 } or die "cannot require $module: $@";

if ( $mode eq 'inc' ) {
    say $INC{$file};
}
elsif ( $mode eq 'prototypes' ) {
    for my $name (@args) {
        my $code = exists &{"${module}::$name"} ? \&{"${module}::$name"} : $module->can($name);
        my $proto = $code && prototype($code);
        say join "\t", $name, !$code ? 'missing' : defined $proto ? ( 'proto', $proto ) : 'none';
    }
}
elsif ( $mode eq 'installs' ) {
    my %seen;
    my @lists = grep { !$seen{$_}++ } @args,
      map { s/^&//r } @{"${module}::EXPORT"}, @{"${module}::EXPORT_OK"};
    my $n = 0;
    for my $list (@lists) {
        my $caller = 'Held::Caller' . $n++;
        if ( !eval "package $caller; use $module " . ( $list eq '' ? '' : "qw($list)" ) . "; 1" ) {
            say join "\t", $list, 'dies', $@ =~ s/\n.*//sr;
            next;
        }
        say join "\t", $list, 'installs', installed($caller);
    }
}
else {
    die "unknown mode $mode\n";
}

# installed is what a `use` put in $caller: a sub a call there reaches,
# and a variable whose slot is the module's own.
sub installed ($caller) {
    my @names;
    for my $name ( sort grep { !/::\z/ } keys %{"${caller}::"} ) {
        push @names, $name if callable("${caller}::$name");
        for ( [ '$', 'SCALAR' ], [ '@', 'ARRAY' ], [ '%', 'HASH' ] ) {
            my ( $sigil, $slot ) = @$_;
            my $mine = *{"${caller}::$name"}{$slot} or next;
            my $its  = *{"${module}::$name"}{$slot} or next;
            push @names, "$sigil$name" if refaddr($mine) == refaddr($its);
        }
    }
    return @names;
}

# callable is whether a call to the sub named $name runs code: one that is
# defined, or one that exists undefined in a package with an AUTOLOAD to
# run in its place. Exists, not defined: measured on 5.42.0, a bare `use
# POSIX` installs 586 subs, 124 of them declared for its AUTOLOAD and not
# yet defined. An undefined sub with no AUTOLOAD is no sub: Exporter makes
# such a stub for `use Carp qw(verbose)`, an @EXPORT_FAIL flag, and a call
# to it dies "Undefined subroutine &Carp::verbose".
sub callable ($name) {
    return 0 if !exists &$name;
    return 1 if defined &$name;
    my $home = B::svref_2object( \&$name )->GV->STASH->NAME;
    return defined &{"${home}::AUTOLOAD"};
}
