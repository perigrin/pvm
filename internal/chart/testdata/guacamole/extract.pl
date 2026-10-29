#!/usr/bin/env perl
# ABOUTME: Extracts Guacamole's parses/parsent corpus by running its t/ suite with the recording Guacamole::Test.
# ABOUTME: Also writes the stub modules and perl-rejects.txt from perl -c verdicts; see PROVENANCE for the command.

use strict;
use warnings;

use Cwd qw(abs_path);
use File::Basename qw(dirname);
use File::Find qw(find);
use File::Path qw(make_path remove_tree);
use File::Spec;
use File::Temp qw(tempfile);
use IPC::Open3 qw(open3);
use JSON::PP qw(decode_json);

my $checkout = shift or die "usage: $0 GUACAMOLE_CHECKOUT\n";
$checkout = abs_path($checkout);
my $here  = dirname( abs_path(__FILE__) );
my $stubs = "$here/stubs";

# shortcut: one fixed version for every stub; raise it if a case ever asks for more.
my $STUB_VERSION = '9999';

my $preamble = slurp("$here/preamble.pl");
$preamble =~ /\A[^\n]*\n\z/ or die "preamble.pl must be exactly one line\n";

# 1. Record every parses/parsent argument the suite makes.
my @tests;
find( { no_chdir => 1, wanted => sub { push @tests, $_ if /\.t\z/ } }, "$checkout/t" );
@tests = sort @tests;

my ( undef, $record ) = tempfile( UNLINK => 1 );
for my $t (@tests) {
    local $ENV{GUACAMOLE_RECORD} = $record;
    my ( $ok, $out ) = run( $checkout, $^X, "-I$here", $t );
    $ok or die "$t failed under the recorder:\n$out";
}

my ( %calls, %seen, %cases );    # kind => count; kind => {text}; kind => [[name, text]]
my %per_test;
open my $rfh, '<', $record or die "open $record: $!\n";
while ( my $line = <$rfh> ) {
    my $r = decode_json($line);
    my ( $kind, $text ) = @{$r}{qw(kind text)};
    utf8::encode($text) if utf8::is_utf8($text);
    $calls{$kind}++;
    next if $seen{$kind}{$text}++;
    ( my $stem = File::Spec->abs2rel( $r->{test}, "$checkout/t" ) ) =~ s{\.t\z}{};
    $stem =~ s{/}{-}g;
    my $n = ++$per_test{$kind}{$stem};
    push @{ $cases{$kind} }, [ sprintf( '%s-%03d.pl', $stem, $n ), $text ];
}
close $rfh;

# 2. Write one case file per unique text, preamble first.
for my $kind (qw(parses parsent)) {
    remove_tree("$here/$kind");
    make_path("$here/$kind");
    spew( "$here/$kind/$_->[0]", $preamble . $_->[1] ) for @{ $cases{$kind} };
}

# 3. Stub every module a parses case uses that perl cannot find, to a fixed point.
remove_tree($stubs);
make_path($stubs);
my @parses = map {"$here/parses/$_->[0]"} @{ $cases{parses} };
my %without = rejected_by_perl( undef, @parses );
my ( %stubbed, %rejected );
while (1) {
    %rejected = rejected_by_perl( $stubs, @parses );
    my $added = 0;
    for my $out ( values %rejected ) {
        while ( $out =~ /Can't locate (\S+)\.pm in \@INC/g ) {
            my $path = "$stubs/$1.pm";
            ( my $mod = $1 ) =~ s{/}{::}g;
            next if $stubbed{$mod}++;
            make_path( dirname($path) );
            spew( $path, "package $mod;\n# ABOUTME: Stub for a module a Guacamole parses case uses; written by extract.pl.\nour \$VERSION = '$STUB_VERSION';\n1;\n" );
            $added++;
        }
    }
    last unless $added;
}

# 4. The parses cases perl still rejects, by bare file name.
my @rejects = sort map { ( File::Spec->splitpath($_) )[2] } keys %rejected;
spew( "$here/perl-rejects.txt", join '', map {"$_\n"} @rejects );

my $recovered = keys(%without) - @rejects;
printf "calls: parses %d, parsent %d\n", $calls{parses}, $calls{parsent};
printf "unique: parses %d, parsent %d\n", scalar @{ $cases{parses} }, scalar @{ $cases{parsent} };
printf "perl rejects: %d without stubs, %d with %d stubs (%d recovered)\n",
    scalar keys %without, scalar @rejects, scalar keys %stubbed, $recovered;

# rejected_by_perl returns (file => perl's output) for each file perl -c rejects.
sub rejected_by_perl {
    my ( $lib, @files ) = @_;
    my %rejected;
    for my $f (@files) {
        my ( $ok, $out ) = perl_c( $lib, $f );
        $rejected{$f} = $out unless $ok;
    }
    return %rejected;
}

# perl_c runs perl -c on $file with $lib (or nothing) as PERL5LIB.
sub perl_c {
    my ( $lib, $file ) = @_;
    local $ENV{PERL5LIB} = $lib;
    delete $ENV{PERL5LIB} unless defined $lib;
    return run( $here, $^X, '-c', $file );
}

# run executes a command in $dir and returns (succeeded, combined output).
sub run {
    my ( $dir, @cmd ) = @_;
    my $cwd = Cwd::getcwd();
    chdir $dir or die "chdir $dir: $!\n";
    my $pid = open3( my $in, my $out, undef, @cmd );
    chdir $cwd or die "chdir $cwd: $!\n";
    close $in;
    my $text = do { local $/; <$out> } // '';
    waitpid $pid, 0;
    return ( $? == 0, $text );
}

sub slurp {
    my ($path) = @_;
    open my $fh, '<:raw', $path or die "open $path: $!\n";
    local $/;
    return scalar <$fh>;
}

sub spew {
    my ( $path, $bytes ) = @_;
    open my $fh, '>:raw', $path or die "open $path: $!\n";
    print {$fh} $bytes;
    close $fh or die "close $path: $!\n";
}
