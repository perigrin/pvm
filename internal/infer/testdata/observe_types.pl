# ABOUTME: Runs a Perl script under instrumentation and reports each assignment's observed value type.
# ABOUTME: Provides ground truth for measuring PSC inference precision without hand annotation.

use v5.36;
use strict;
use warnings;
no warnings 'experimental::builtin';
use B ();
use Scalar::Util qw(blessed reftype looks_like_number);
use builtin qw(is_bool);

# WHAT THIS IS FOR.
#
# PSC's coverage number says how OFTEN it assigns a type. It says nothing
# about whether the type is RIGHT, and the two move independently: a change
# that turns Unknowns into confident-but-wrong types improves coverage and
# makes the checker worse.
#
# The literature calls these recall and precision (TypeEvalPy), and the hard
# part is ground truth — perl has no annotations to harvest. So perl itself is
# the oracle: run the program, observe what each variable ACTUALLY holds, and
# compare against what PSC predicted statically.
#
# The limit is honest and worth stating: this covers only executed paths. A
# branch never taken contributes nothing. That is a KNOWN limit, unlike a
# hand-written table's unknown one.

# observe maps a live value to the name of its PSC lattice type, using the
# paper's membership tests rather than SV flags. SV flags record a value's
# HISTORY — "hi" reads as PVNV once anything has numified it — where the
# paper's question is what the value IS.
sub observe ($v) {
    return 'Undef' if !defined $v;

    if (ref $v) {
        my $rt = reftype($v) // '';
        return 'Regex'  if $rt eq 'REGEXP';
        return 'Object' if defined blessed($v);
        return 'ArrayRef' if $rt eq 'ARRAY';
        return 'HashRef'  if $rt eq 'HASH';
        return 'CodeRef'  if $rt eq 'CODE';
        return 'GlobRef'  if $rt eq 'GLOB';
        return 'ScalarRef';
    }

    # Boolean first: a marked boolean also looks like an integer.
    return 'Bool' if is_bool($v);

    # Int before Num, since every integer is also a number.
    return 'Int' if $v =~ /\A-?[0-9]+\z/;
    return 'Num' if looks_like_number($v);
    return 'Str';
}

# The script under test appends observations to @::OBSERVED as
# [line, varname, type] triples by calling ::__observe.
our @OBSERVED;

sub __observe ($line, $name, $value) {
    push @OBSERVED, [ $line, $name, observe($value) ];
    return $value;
}

# Emit the observations as JSON on exit, hand-rolled to avoid a non-core
# dependency.
sub __emit {
    my @out;
    for my $o (@OBSERVED) {
        my ($line, $name, $type) = @$o;
        $name =~ s/(["\\])/\\$1/g;
        push @out, qq[{"line":$line,"name":"$name","type":"$type"}];
    }
    print STDOUT '[', join( ',', @out ), "]\n";
}

END { __emit() }

1;
