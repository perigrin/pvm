#!/usr/bin/env perl
# ABOUTME: Perl's own parser as a subject of the fidelity harness, through the subprocess contract.
# ABOUTME: A second implementation that is not Go; measured against the oracle it must score exact wherever it compiles.
use strict;
use warnings;

# Usage: perl_subject.pl FILE
#
# The harness runs this from the corpus directory and hands it a path relative
# to there, exactly as it runs the oracle. The answer is one JSON object on
# stdout and exit 0. A file that does not compile is reported as ok:false, not
# as a non-zero exit: the harness reads a non-zero exit as "the subject broke"
# and records a runner error, never a verdict.
#
# Every question is answered from the optree. Every srefgen is a reference
# perl took, and each is reported as a reference site rather than a call
# site, because the contract asks for the conclusion and not for where it was
# reached. Perl resolves prototypes itself, so a \@-prototype call is a
# reference here where a static parser can only hedge. The other four kinds
# are the ops that witness them: rv2hv is a hash, match is a match, readline
# (and rcatline, the optimiser's spelling of `$x .= <FH>`) is a readline,
# anonhash -- and emptyavhv flagged ANONHASH, which is `{}` -- is an anonhash.
#
# ponytail: taint (-T) shebangs are not honoured; such a file reports ok:false
# and scores no-answer. Read the shebang the way run.go's wantsTaint does if
# this subject is ever run over the full corpus.

my $file = shift or die "usage: $0 FILE\n";

(my $quoted = $file) =~ s/'/'\\''/g;
my $concise = `perl -MO=Concise,-exec -c '$quoted' 2>&1`;
if ($? != 0) {
    print qq({"ok":false}\n);
    exit 0;
}

# "5  <1> srefgen sK/1 ->6" -- the op name is the third field.
#
# Each reference is reported with the line of the statement it belongs to,
# because the comparison matches per statement: a site without a line spans
# nothing perl attributed a reference to, and accounts for nothing. The line
# comes from the nearest preceding nextstate/dbstate, which is the same COP
# the oracle reads, so the two sides name the same statement by construction.
#
#   "1  <;> nextstate(main 2 explicit_ref.pl:5) v:{"
my %kind = (
    srefgen => 'reference', refgen => 'reference',
    rv2hv => 'hash',
    match => 'match',
    readline => 'readline', rcatline => 'readline',
    anonhash => 'anonhash',
);

my @sites;
my $line = 0;
for my $op (split /\n/, $concise) {
    if ($op =~ /^\s*\S+\s+<;>\s+(?:next|db)state\([^)]*:(\d+)\)/) {
        $line = $1;
        next;
    }
    # "3  <0> emptyavhv[$x:1,2] v/LVINTRO,ANONHASH,TARGMY" -- the op name,
    # then its target and flags, where emptyavhv says which of {} and []
    # it is.
    next unless $op =~ /^\s*\S+\s+<[^>]*>\s+(\w+)(.*)$/;
    my ($name, $rest) = ($1, $2);
    my $kind = $kind{$name};
    $kind = 'anonhash' if $name eq 'emptyavhv' && $rest =~ /\bANONHASH\b/;
    next unless $kind;
    push @sites, $kind eq 'reference'
        ? qq({"line":$line,"kind":"reference","took_reference":true})
        : qq({"line":$line,"kind":"$kind"});
}

my $sites = join ',', @sites;
print qq({"ok":true,"call_sites":[$sites]}\n);
