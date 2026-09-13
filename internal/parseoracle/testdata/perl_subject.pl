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
# The reference question is answered from the optree. Every srefgen is a
# reference perl took, and each is reported as a reference site rather than a
# call site, because the contract asks for the conclusion and not for where it
# was reached. Perl resolves prototypes itself, so a \@-prototype call is a
# reference here where a static parser can only hedge.
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
my $refs = grep { /^\s*\S+\s+<[^>]*>\s+srefgen\b/ } split /\n/, $concise;

my $sites = join ',', ('{"kind":"reference","took_reference":true}') x $refs;
print qq({"ok":true,"call_sites":[$sites]}\n);
