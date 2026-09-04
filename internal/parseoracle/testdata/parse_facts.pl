#!/usr/bin/env perl
# ABOUTME: Extracts perl's own parse decisions for a source file as JSON.
# ABOUTME: Ground truth for measuring a static parser's fidelity, not just its coverage.
use strict;
use warnings;

# Perl reports how it parsed something. The optree is the parse, resolved: a
# prototype that turned a list into a reference shows up as an srefgen that is
# simply absent without the prototype. That makes fidelity measurable instead
# of a matter of opinion.
#
# Usage: parse_facts.pl FILE
# Emits one JSON object on stdout. Never dies on a bad parse -- an unparsable
# file is a fact about the file, reported as ok:0.

my $file = shift or die "usage: $0 FILE\n";

# -exec walks the optree in execution order, which linearises it into a
# sequence a test can diff. Without it the output is a tree drawn in ASCII.
my $concise = capture("-MO=Concise,-exec", $file);
my $compiles = $? == 0;

my @ops;
for my $line (split /\n/, $concise) {
    # "5  <0> pushmark s" -- the op name is the third field.
    next unless $line =~ /^\s*\S+\s+<[^>]*>\s+(\w+)/;
    push @ops, $1;
}

# Prototypes are the single highest-value parse fact: they change the parse at
# every call site, and they are the thing a static parser cannot know without
# having already seen the definition.
my %proto;
my $protosrc = <<'PERL';
    for my $name (sort keys %main::) {
        no strict 'refs';
        next unless defined &{"main::$name"};
        my $p = prototype(\&{"main::$name"});
        next unless defined $p;
        print "PROTO\t$name\t$p\n";
    }
PERL

my $protoout = capture_with_end($file, $protosrc);
for my $line (split /\n/, $protoout) {
    $proto{$2} = $3 if $line =~ /^(PROTO)\t([^\t]*)\t(.*)$/;
}

print encode({
    file      => $file,
    ok        => $compiles ? 1 : 0,
    ops       => \@ops,
    op_count  => scalar(@ops),
    srefgen   => scalar(grep { $_ eq 'srefgen' } @ops),
    entersub  => scalar(grep { $_ eq 'entersub' } @ops),
    prototypes=> \%proto,
}), "\n";

# Run perl on the file with the given flags, returning combined output. The
# file is compiled, never run: everything here is a compile-time question.
# Perl's own tests chdir to t/ and `require "./test.pl"`, so they only compile
# from that directory. The caller passes the directory to run from; without it
# a third of the corpus reports a BEGIN failure that is about @INC, not syntax.
sub capture {
    my ($flags, $path) = @_;
    my $cmd = sprintf("perl %s -c %s 2>&1", $flags, quote($path));
    $cmd = sprintf("cd %s && %s", quote($ENV{ORACLE_CHDIR}), $cmd)
        if $ENV{ORACLE_CHDIR};
    return scalar `$cmd`;
}

# Prototypes have to be read after compilation but before execution, which is
# what CHECK is for. The probe is injected rather than written into the file.
sub capture_with_end {
    my ($path, $body) = @_;
    my $probe = "CHECK {\n$body\n}\n";
    open my $fh, '<', $path or return '';
    my $src = do { local $/; <$fh> };
    close $fh;
    my ($tmp) = "$path.oracle.$$.pl";
    open my $out, '>', $tmp or return '';
    print $out $probe, $src;
    close $out;
    my $cmd = "perl -c @{[quote($tmp)]} 2>&1";
    $cmd = sprintf("cd %s && %s", quote($ENV{ORACLE_CHDIR}), $cmd)
        if $ENV{ORACLE_CHDIR};
    my $res = scalar `$cmd`;
    unlink $tmp;
    return $res;
}

sub quote { my $s = shift; $s =~ s/'/'\\''/g; return "'$s'" }

# A dependency-free encoder: the harness must run against any perl, including
# one with no non-core modules installed.
sub encode {
    my ($d) = @_;
    if (ref $d eq 'HASH')  { return '{' . join(',', map { encode("$_") . ':' . encode($d->{$_}) } sort keys %$d) . '}' }
    if (ref $d eq 'ARRAY') { return '[' . join(',', map { encode($_) } @$d) . ']' }
    return 'null' unless defined $d;
    return $d if $d =~ /\A-?(?:0|[1-9][0-9]*)\z/;
    my $s = $d;
    $s =~ s/(["\\])/\\$1/g;
    $s =~ s/\n/\\n/g; $s =~ s/\r/\\r/g; $s =~ s/\t/\\t/g;
    $s =~ s/([\x00-\x1f])/sprintf("\\u%04x", ord $1)/ge;
    return '"' . $s . '"';
}
