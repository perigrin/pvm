# ABOUTME: Shared by held_to_perl.pl and return_oracle.pl: one rule for whether a call to a sub
# ABOUTME: named by its fully-qualified name runs code, the "found" check each driver's own lookup uses.
use v5.36;
no strict 'refs';
use B ();

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

1;
