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
#   held_to_perl.pl skeleton MODULE          prints a .tpm skeleton for MODULE: `package M;`, then
#                                            one `sub NAME :export|:export_ok(tags) (PARAMS) TODO;`
#                                            or `our TODO NAME :export...;` per name `installs`
#                                            would report, sorted
# Fields are tab-separated, one line per argument. `installs` also asks for
# every name in the module's @EXPORT and @EXPORT_OK, an `&name` as `name`,
# and every tag in its %EXPORT_TAGS, as `:tag`.
my ( $mode, $module, @args ) = @ARGV;
( my $file = "$module.pm" ) =~ s{::}{/}g;
eval { require $file; 1 } or die "cannot require $module: $@";

if ( $mode eq 'inc' ) {
    say $INC{$file};
}
elsif ( $mode eq 'prototypes' ) {
    # A module whose import forward-declares subs for its AUTOLOAD, as
    # POSIX's does (measured on 5.42.0: `exists &POSIX::stdin` is false
    # after `require POSIX`, true only once a use has run), has nothing
    # to check until one use has primed it.
    eval "package Held::Prime; use $module; 1";
    my $n = 0;
    for my $name (@args) {
        my $code = exists &{"${module}::$name"} ? \&{"${module}::$name"} : $module->can($name);
        # A module whose import re-exports a third module's names
        # straight into its caller -- FileHandle's import calls
        # Exporter::export('Fcntl', $callpkg) for a bare `use`, never
        # defining the name in FileHandle:: itself or anywhere
        # FileHandle->can reaches -- leaves the name only in whatever
        # package did the bare use. Measured on 5.42.0: FileHandle's
        # F_DUPFD exists in Held::Prime after the bare use above, is
        # not exported by name at all ("F_DUPFD" is not exported by the
        # FileHandle module), and FileHandle::F_DUPFD and
        # FileHandle->can('F_DUPFD') are both false.
        $code ||= exists &{"Held::Prime::$name"} ? \&{"Held::Prime::$name"} : undef;
        if ( !$code ) {
            # A name a bare `use` does not install, measured, still gets
            # its stub only once it is asked for by name -- Fcntl's
            # S_ISWHT, whose sibling S_IS* subs exist from a bare use,
            # does not. Prime it by name before calling it missing.
            my $primed = "Held::PrimeByName@{[$n++]}";
            eval "package $primed; use $module qw($name); 1";
            $code = exists &{"${module}::$name"} ? \&{"${module}::$name"} : $module->can($name);
            # A module with its own import, as Locale::Maketext::Simple's
            # and IO::Socket::IP's are, installs a glob straight into the
            # caller's package and never into its own (measured: unlike
            # POSIX's AUTOLOAD, nothing it does defines
            # Locale::Maketext::Simple::loc). The only place left to find
            # it is the package it was just primed into.
            $code ||= exists &{"${primed}::$name"} ? \&{"${primed}::$name"} : undef;
        }
        my $proto = $code && prototype($code);
        say join "\t", $name, !$code ? 'missing' : defined $proto ? ( 'proto', $proto ) : 'none';
    }
}
elsif ( $mode eq 'installs' ) {
    # A module whose lists are built inside its own import, as POSIX's
    # are (measured on 5.42.0: @EXPORT/@EXPORT_OK/%EXPORT_TAGS are empty
    # right after `require POSIX`, full only once a use has run), has
    # nothing to find in them until one use has primed it.
    eval "package Held::Prime; use $module; 1";
    my %seen;
    my @lists = grep { !$seen{$_}++ } @args,
      ( map { s/^&//r } @{"${module}::EXPORT"}, @{"${module}::EXPORT_OK"} ),
      map { ":$_" } sort keys %{"${module}::EXPORT_TAGS"};
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
elsif ( $mode eq 'skeleton' ) {
    my %seen;
    my @lists = grep { !$seen{$_}++ } ( map { s/^&//r } @{"${module}::EXPORT"}, @{"${module}::EXPORT_OK"} ),
      map { ":$_" } sort keys %{"${module}::EXPORT_TAGS"};
    my ( %bare, %ok, %tags );
    my $n = 0;
    for my $list ( '', @lists ) {
        my $caller = 'Held::Skel' . $n++;
        next unless eval "package $caller; use $module " . ( $list eq '' ? '' : "qw($list)" ) . "; 1";
        my @names = installed($caller);
        if    ( $list eq '' )       { $bare{$_} = 1 for @names }
        elsif ( $list =~ /^:(.*)/ ) { push @{ $tags{$_} }, $1 for @names }
        elsif ( grep { $_ eq $list } @names ) { $ok{$list} = 1 }
    }
    my %all = ( %bare, %ok, map { $_ => 1 } keys %tags );
    say "package $module;";
    for my $name ( sort keys %all ) {
        my ($sigil) = $name =~ /^([\$\@%])/;
        my $mark = $bare{$name} ? 'export' : 'export_ok';
        $mark .= '(' . join( ',', sort @{ $tags{$name} || [] } ) . ')' if $tags{$name};
        if ($sigil) {
            say "our TODO $name :$mark;";
            next;
        }
        my $code  = exists &{"${module}::$name"} ? \&{"${module}::$name"} : $module->can($name);
        my $proto = $code && prototype($code);
        $mark .= " :prototype($proto)" if defined $proto && $proto ne '';
        say "sub $name :$mark " . skeletonParams($proto) . " TODO;";
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

# skeletonParams is the typed parameter list `skeleton` writes for a sub
# whose perl prototype is proto (undef for none): $ -> Any, @ -> List,
# \@ -> Array \@, & -> Code, and ; marks every scalar after it optional,
# defaulted undef so its type need not change. No prototype takes
# whatever List @args the caller passes.
#
# ponytail: only $, @, \@, & and ; are mapped; a prototype using any
# other character (\$, \%, %, *, _, +) falls back to the same permissive
# (List @args) a missing prototype gets, rather than guess its type. Add
# that character here when a core module's skeleton needs it typed.
sub skeletonParams ($proto) {
    return '(List @args)' if !defined $proto;
    my @parts;
    my ( $optional, $letter, $i ) = ( 0, 'a', 0 );
    my @chars = split //, $proto;
    while ( $i < @chars ) {
        my $c = $chars[$i];
        if ( $c eq ';' ) { $optional = 1; $i++; next }
        if ( $c eq '\\' && $i + 1 < @chars && $chars[ $i + 1 ] eq '@' ) {
            push @parts, "Array \\\@$letter";
            $letter++;
            $i += 2;
            next;
        }
        if ( $c eq '$' ) {
            push @parts, "Any \$$letter" . ( $optional ? ' = undef' : '' );
            $letter++;
            $i++;
            next;
        }
        if ( $c eq '@' ) {
            push @parts, 'List @rest';
            $i++;
            next;
        }
        if ( $c eq '&' ) {
            push @parts, 'Code &cb';
            $i++;
            next;
        }
        return '(List @args)';
    }
    return @parts ? '(' . join( ', ', @parts ) . ')' : '()';
}
