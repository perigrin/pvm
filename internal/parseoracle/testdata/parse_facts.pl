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
#
# This run answers "did it compile" and supplies the op list. It does NOT
# supply the reference count: Concise dumps PL_main_root only, and a named
# sub, an anonymous sub or a BEGIN block is a separate CV that never appears
# in it. The probe further down walks every CV instead.
my $concise = capture("-MO=Concise,-exec", $file);
my $compiles = $? == 0;

my @ops;
for my $line (split /\n/, $concise) {
    # "5  <0> pushmark s" -- the op name is the third field.
    next unless $line =~ /^\s*\S+\s+<[^>]*>\s+(\w+)/;
    push @ops, $1;
}

# A failing compile leaves its diagnostic in the captured output, and that
# diagnostic is the only thing that separates "your parser is wrong" from
# "this machine lacks Config.pm". Exit status cannot tell them apart, so
# discarding it would make a ratchet built on this output encode the second
# as though it were the first.
my $stderr = $compiles ? '' : $concise;

# Prototypes are the single highest-value parse fact: they change the parse at
# every call site, and they are the thing a static parser cannot know without
# having already seen the definition.
#
# The same probe walks every CV perl compiled for this file and reports each
# reference-taking op with the line of the statement it belongs to. The
# population is deliberately "what the source could have written a backslash
# for": explicit `\` forms and prototype-forced references at call sites.
#
#   srefgen           \@a  \%h  \&f  \$x  \(&f)  \-1  and every \-prototype arg
#   refgen            \(@a)  \(@a,@b)  \my($x,$y)     -- one op per list, as
#                                                        the source has one \
#   const[IV \1]      \1  \"x"  \'s'   folded away: a backslash with no op
#   srefgen under goto  goto &NAME     excluded, see the walk
#
# Attribution is by the nearest enclosing nextstate/dbstate, which names the
# statement's FIRST line -- measured on multi-line calls, hash literals and
# if/elsif/for/while conditions -- and the COP's file, so subs a corpus file
# pulls in from t/test.pl are not counted against it.
my %proto;
my $protosrc = <<'PERL';
    no strict 'refs';
    no warnings;
    for my $name (sort keys %main::) {
        next unless defined &{"main::$name"};
        my $p = prototype(\&{"main::$name"});
        next unless defined $p;
        print "PROTO\t$name\t$p\n";
    }

    my (%seen, @sites);
    my ($walk, $walkcv);
    $walk = sub {
        my ($op, $cop, $cv, $parent) = @_;
        return unless ref $op && $$op;
        my $name = $op->name;
        if ($name eq 'srefgen' || $name eq 'refgen') {
            # `goto &NAME` is a goto whose operand perly.y parsed as an
            # entersub term; op.c's newLOOPEX then wraps that in a REFGEN.
            # The srefgen is goto's rewrite, not a reference the parse took
            # at a call site, and no subject could write a backslash for it.
            push @sites, [$cop->line, $name]
                if $cop && $cop->file eq $oracle_file
                && !($parent && $parent->name eq 'goto');
        }
        if ($op->flags & B::OPf_KIDS()) {
            # A COP names the statement for the siblings that FOLLOW it at
            # this level only; a nested block's COPs must not leak out to
            # the ops after the block. `sort { ... } f(@a)` attributes
            # f(@a) to the sort statement, not to the block's last line.
            for (my $k = $op->first; $$k; $k = $k->sibling) {
                if ($k->isa('B::COP')) { $cop = $k; next }
                $walk->($k, $cop, $cv, $op);
            }
        }
        elsif ($op->isa('B::PMOP')) {
            # A regex (?{ ... }) block hangs off the PMOP rather than being
            # a kid. B::Concise walks it the same way.
            my $code = $op->code_list;
            $walk->($code, $cop, $cv, $op) if ref $code && $code->isa('B::OP');
        }
        if ($name eq 'anoncode') {
            # An anonymous sub is its own CV, reachable only from the op
            # that closes over it. Under ithreads the CV lives in the pad.
            my $sv = $op->sv;
            $sv = (($cv->PADLIST->ARRAY)[1]->ARRAY)[$op->targ] unless $$sv;
            $walkcv->($sv);
        }
    };
    $walkcv = sub {
        my ($cv) = @_;
        return unless ref $cv && $cv->isa('B::CV') && !$seen{$$cv}++;
        my $root = $cv->ROOT;
        $walk->($root, undef, $cv, undef) if ref $root && $$root;
    };

    $seen{${B::main_cv()}}++;
    $walk->(B::main_root(), undef, B::main_cv(), undef);

    # BEGIN blocks are run and freed during compilation; the BEGIN at the top
    # of this probe asked perl to keep them (B::save_BEGINs) so they can be
    # walked here. CHECK/INIT/END are kept anyway.
    for my $av (B::begin_av(), B::unitcheck_av(), B::check_av(), B::init_av(), B::end_av()) {
        next unless ref $av && $av->isa('B::AV');
        $walkcv->($_) for $av->ARRAY;
    }

    # Named subs, in every package the file declared: walk the stash tree
    # from main::, taking only CVs this file defined. `#line` above set the
    # file name back to the original, so CvFILE matches.
    my %stash_seen;
    my $walkstash;
    $walkstash = sub {
        my ($pkg) = @_;
        my $stash = \%{"${pkg}::"};
        return if $stash_seen{$stash}++;
        for my $name (keys %$stash) {
            if ($name =~ /::\z/) {
                my $inner = substr($name, 0, -2);
                $walkstash->($pkg eq 'main' ? $inner : "${pkg}::$inner");
                next;
            }
            my $full = "${pkg}::$name";
            next unless defined &$full;
            my $cv = B::svref_2object(\&$full);
            next unless ($cv->FILE // '') eq $oracle_file;
            $walkcv->($cv);
        }
    };
    $walkstash->('main');

    print "SITE\t$_->[0]\t$_->[1]\n" for @sites;
    print "WALK\tok\n";
PERL

my $protoout = capture_with_end($file, $protosrc);
my (@sites, $walked);
for my $line (split /\n/, $protoout) {
    $proto{$2} = $3 if $line =~ /^(PROTO)\t([^\t]*)\t(.*)$/;
    push @sites, [$1, $2] if $line =~ /^SITE\t(\d+)\t(\w+)$/;
    $walked = 1 if $line eq "WALK\tok";
}

print encode({
    file      => $file,
    ok        => $compiles ? 1 : 0,
    ops       => \@ops,
    op_count  => scalar(@ops),
    srefgen   => scalar(grep { $_->[1] eq 'srefgen' } @sites),
    ref_lines => [ sort { $a <=> $b } map { $_->[0] } @sites ],
    walked    => $walked ? 1 : 0,
    entersub  => scalar(grep { $_ eq 'entersub' } @ops),
    prototypes=> \%proto,
    stderr    => $stderr,
}), "\n";

# Run perl on the file with the given flags, returning combined output. The
# file is compiled, never run: everything here is a compile-time question.
# Perl's own tests chdir to t/ and `require "./test.pl"`, so they only compile
# from that directory. The caller passes the directory to run from; without it
# a third of the corpus reports a BEGIN failure that is about @INC, not syntax.
sub capture {
    my ($flags, $path) = @_;
    my $cmd = sprintf("perl%s %s -c %s 2>&1", taint(), $flags, quote($path));
    $cmd = sprintf("cd %s && %s", quote($ENV{ORACLE_CHDIR}), $cmd)
        if $ENV{ORACLE_CHDIR};
    return scalar `$cmd`;
}

# Prototypes have to be read after compilation but before execution, which is
# what CHECK is for. The probe is injected rather than written into the file.
#
# Three things about the injection are load-bearing. The BEGIN comes first so
# it runs before any BEGIN in the file, which is the only moment at which
# asking perl to keep BEGIN blocks (B::save_BEGINs) still catches them. The
# CHECK block sees $oracle_file, the path exactly as perl was given it, which
# is what every COP and CV in the file names. And `#line 1 "path"` after the
# probe hands the file its own name and numbering back, so a site is reported
# on the line the source has it and CvFILE matches the path.
sub capture_with_end {
    my ($path, $body) = @_;
    my $probe = "BEGIN { require B; B::save_BEGINs() }\n"
              . "CHECK {\n    my \$oracle_file = " . perlquote($path) . ";\n$body\n}\n"
              . "#line 1 \"$path\"\n";
    open my $fh, '<', $path or return '';
    my $src = do { local $/; <$fh> };
    close $fh;
    my ($tmp) = "$path.oracle.$$.pl";
    open my $out, '>', $tmp or return '';
    print $out $probe, $src;
    close $out;
    my $cmd = "perl@{[taint()]} -c @{[quote($tmp)]} 2>&1";
    $cmd = sprintf("cd %s && %s", quote($ENV{ORACLE_CHDIR}), $cmd)
        if $ENV{ORACLE_CHDIR};
    my $res = scalar `$cmd`;
    unlink $tmp;
    return $res;
}

sub quote { my $s = shift; $s =~ s/'/'\\''/g; return "'$s'" }

# A single-quoted Perl string literal, for splicing the path into the probe.
sub perlquote { my $s = shift; $s =~ s/(['\\])/\\$1/g; return "'$s'" }

# Three corpus files carry `#!./perl -T`, and perl refuses to compile a file
# whose shebang asks for taint mode unless -T is on the command line too. The
# caller decides, because reading the shebang is the caller's job.
sub taint { return $ENV{ORACLE_TAINT} ? ' -T' : '' }

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
