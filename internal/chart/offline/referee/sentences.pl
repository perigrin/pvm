#!/usr/bin/perl
# ABOUTME: Generates the precedence referee's sentences: one per bison (rule, token) resolution of perly.y, kept where the referees disagree.
# ABOUTME: Run by sentences.sh, which builds its inputs; writes resolutions.txt, sentences.txt and uncovered-reviewed.txt into an output directory.
use strict;
use warnings;

# Usage: sentences.pl <workdir> <outdir>. <workdir> holds sentences.sh's
# builds: perly.output (bison's report of perly.y itself), prec.output and
# noprec.output (its reports of the two referee grammars), referee-prec.y,
# referee-prec-trace (referee-prec built with YYDEBUG) and referee-noprec.
my ($work, $out) = @ARGV;
die "usage: sentences.pl <workdir> <outdir>\n" unless defined $out;

# The bound on the search: candidate sentences tried per resolution. A
# resolution whose candidates within it all fail is uncovered, reason
# `generation timed out`. A count, not a clock, so a run is reproducible.
my $BOUND = 12;
my $OTHERSTART = 1000;    # the cost of a sentence not starting with GRAMPROG

sub slurp { open my $f, '<', $_[0] or die "sentences.pl: $_[0]: $!\n"; local $/; <$f> }

# --- The grammar and automaton (referee-noprec's: the LR(0) automaton is the same with and without precedence).
my (@lhs, @rhs, %rules_of);
my $np = slurp("$work/noprec.output");
my ($gram) = $np =~ /^Grammar\n(.*?)^Terminals/ms or die "sentences.pl: no Grammar in noprec.output\n";
{
	my $l;
	for (split /\n/, $gram) {
		next unless /^\s*(\d+)\s+(?:(\S+):|\|)\s*(.*)$/;
		$l = $2 if defined $2;
		my @r = grep { $_ ne '%empty' } split ' ', $3;
		($lhs[$1], $rhs[$1]) = ($l, \@r);
		push @{ $rules_of{$l} }, $1;
	}
}
my (%goto, %items, $state);
for (split /\n/, $np) {
	if (/^State (\d+)$/) { $state = $1; next }
	next unless defined $state;
	if (/^\s+(\d+)\s+(?:\S+:|\|)\s*(.*)$/) {
		my @r = grep { $_ ne '%empty' } split ' ', $2;
		my ($dot) = grep { $r[$_] eq '.' } 0 .. $#r;
		push @{ $items{$state} }, [ $1, $dot ];
	} elsif (/^\s+(\S+)\s+(?:shift, and )?go to state (\d+)$/) {
		$goto{$state}{$1} = $2;
	}
}
my %nonterm = map { $_ => 1 } grep { defined } @lhs;

# Rules that refuse in an action, or read `error`, never stand in a yield.
my @refusing;
{
	my $i = 0;
	for (split /\n/, slurp("$work/referee-prec.y")) {
		next unless /^\w+:/;
		$i++;
		$refusing[$i] = /yyerror/ || /\berror\b/ ? 1 : 0;
	}
	die "sentences.pl: referee-prec.y has $i rules, its report $#lhs\n" unless $i == $#lhs;
}

# --- The resolutions: bison's report of perly.y numbers its rules; the referee
# grammar's report, in the same order, gives each its state.
sub conflicts {
	my ($text) = @_;
	my ($st, @c);
	for (split /\n/, $text) {
		$st = $1 if /^State (\d+)$/;
		push @c, [ $st, $1, $2, $3 ] if /Conflict between rule (\d+) and token (\S+) resolved as (shift|reduce|an error)/;
	}
	return @c;
}
my @orig = conflicts(slurp("$work/perly.output"));
my @prec = conflicts(slurp("$work/prec.output"));
die "sentences.pl: perly.y has ", scalar @orig, " resolutions, referee-prec ", scalar @prec, "\n" unless @orig == @prec;
my @res;
for my $i (0 .. $#orig) {
	my ($o, $p) = ($orig[$i], $prec[$i]);
	die "sentences.pl: resolution $i differs: rule $o->[1] $o->[2] $o->[3], rule $p->[1] $p->[2] $p->[3]\n"
	  unless $o->[2] eq $p->[2] && $o->[3] eq $p->[3];
	push @res, { rule => $o->[1], token => $o->[2], action => $o->[3] eq 'an error' ? 'error' : $o->[3], state => $p->[0], prule => $p->[1] };
}

# --- Shortest yields, and shortest yields starting with a given terminal.
my %ylen;     # nonterminal => length of its shortest yield
my %ychoice;  # nonterminal => rule
{
	my $changed = 1;
	while ($changed) {
		$changed = 0;
		for my $r (1 .. $#lhs) {
			next if $refusing[$r];
			my $n = 0;
			for (@{ $rhs[$r] }) { $n = undef, last unless defined($n) && (!$nonterm{$_} || defined $ylen{$_}); $n += $nonterm{$_} ? $ylen{$_} : 1 }
			next unless defined $n;
			if (!defined $ylen{ $lhs[$r] } || $n < $ylen{ $lhs[$r] }) { ($ylen{ $lhs[$r] }, $ychoice{ $lhs[$r] }) = ($n, $r); $changed = 1 }
		}
	}
}
my %ymemo;
sub yield {    # a symbol's shortest yield, as a list of terminals
	my ($x) = @_;
	return ($x) unless $nonterm{$x};
	die "sentences.pl: $x yields nothing\n" unless defined $ychoice{$x};
	$ymemo{$x} //= [ map { yield($_) } @{ $rhs[ $ychoice{$x} ] } ];
	return @{ $ymemo{$x} };
}
sub slen { my $n = 0; for (@_) { return undef if $nonterm{$_} && !defined $ylen{$_}; $n += $nonterm{$_} ? $ylen{$_} : 1 } $n }

my (%flen, %fchoice, %firsts);    # "X T" => length / [rule, index] of X's shortest yield starting with T
{
	my $changed = 1;
	while ($changed) {
		$changed = 0;
		for my $r (1 .. $#lhs) {
			next if $refusing[$r];
			my @s = @{ $rhs[$r] };
			for my $i (0 .. $#s) {
				my $rest = slen(@s[ $i + 1 .. $#s ]);
				last unless defined $rest;
				for my $t ($nonterm{ $s[$i] } ? sort keys %{ $firsts{ $s[$i] } || {} } : ($s[$i])) {
					my $n = ($nonterm{ $s[$i] } ? $flen{"$s[$i] $t"} : 1) + $rest;
					my $k = "$lhs[$r] $t";
					if (!defined $flen{$k} || $n < $flen{$k}) { ($flen{$k}, $fchoice{$k}) = ($n, [ $r, $i ]); $firsts{ $lhs[$r] }{$t} = 1; $changed = 1 }
				}
				last unless $nonterm{ $s[$i] } && ($ylen{ $s[$i] } // 1) == 0;
			}
		}
	}
}
sub fyield {    # X's shortest yield starting with terminal T
	my ($x, $t) = @_;
	return ($t) if $x eq $t;
	my ($r, $i) = @{ $fchoice{"$x $t"} };
	my @s = @{ $rhs[$r] };
	return ((map { yield($_) } @s[ 0 .. $i - 1 ]), fyield($s[$i], $t), map { yield($_) } @s[ $i + 1 .. $#s ]);
}
sub fseq {    # a symbol string's shortest yield starting with T: [length, terminals] or undef
	my ($t, @s) = @_;
	for my $i (0 .. $#s) {
		my $rest = slen(@s[ $i + 1 .. $#s ]);
		return undef unless defined $rest;
		if ($s[$i] eq $t || $nonterm{ $s[$i] } && defined $flen{"$s[$i] $t"}) {
			my $n = ($s[$i] eq $t ? 1 : $flen{"$s[$i] $t"}) + $rest;
			return [ $n, [ (map { yield($_) } @s[ 0 .. $i - 1 ]), fyield($s[$i], $t), map { yield($_) } @s[ $i + 1 .. $#s ] ] ];
		}
		return undef unless $nonterm{ $s[$i] } && ($ylen{ $s[$i] } // 1) == 0;
	}
	return undef;
}

# --- Shortest paths over (state, rule, dot, placed) from `grammar`'s
# alternatives in state 0. A step reads the symbol after the dot (its shortest
# yield) or descends into a rule for it, paying for the rest of the rule it
# leaves. With a terminal $t, the descent may instead place $t first in that
# rest, after which every enclosing rest must be empty: then a reduction at
# the target is followed directly by $t.
sub search {
	my ($t) = @_;
	my (%dist, %pred, @bucket);
	my $push = sub {
		my ($node, $d, $from, $step) = @_;
		return if defined $dist{$node} && $dist{$node} <= $d;
		($dist{$node}, $pred{$node}) = ($d, [ $from, $step ]);
		push @{ $bucket[$d] }, $node;
	};
	for my $r (@{ $rules_of{grammar} }) {
		$push->("0 $r 0 0", $rhs[$r][0] eq 'GRAMPROG' ? 0 : $OTHERSTART, undef, undef);
	}
	for (my $d = 0; $d <= $#bucket; $d++) {
		for my $node (@{ $bucket[$d] || [] }) {
			next unless $dist{$node} == $d;
			my ($p, $r, $i, $f) = split ' ', $node;
			my @s = @{ $rhs[$r] };
			next if $i > $#s;
			my $x = $s[$i];
			next if $x eq 'error';
			my $len = $nonterm{$x} ? $ylen{$x} : 1;
			$push->(join(' ', $goto{$p}{$x}, $r, $i + 1, $f), $d + $len, $node, [ 'read', $x ]) if defined $len && defined $goto{$p}{$x};
			next unless $nonterm{$x};
			my @rest = @s[ $i + 1 .. $#s ];
			for my $r2 (@{ $rules_of{$x} }) {
				next if grep { $_ eq 'error' } @{ $rhs[$r2] };
				my $pen = $refusing[$r2] ? 50 : 0;
				if (!defined $t) {
					my $n = slen(@rest);
					$push->("$p $r2 0 0", $d + $n + $pen, $node, [ 'rest', [ map { yield($_) } @rest ] ]) if defined $n;
				} elsif ($f) {
					my $n = slen(@rest);
					$push->("$p $r2 0 1", $d + $pen, $node, [ 'rest', [] ]) if defined $n && $n == 0;
				} else {
					my $n = slen(@rest);
					$push->("$p $r2 0 0", $d + $n + $pen, $node, [ 'rest', [ map { yield($_) } @rest ] ]) if defined $n;
					my $fs = fseq($t, @rest);
					$push->("$p $r2 0 1", $d + $fs->[0] + $pen, $node, [ 'rest', $fs->[1] ]) if $fs;
				}
			}
		}
	}
	return (\%dist, \%pred);
}

# A node's sentence: what was read to reach it, then (with @mid) more, then
# the rests it left, innermost first.
sub sentence {
	my ($pred, $node, @mid) = @_;
	my (@read, @rests);
	while (defined $pred->{$node}[0]) {
		my ($from, $step) = @{ $pred->{$node} };
		if ($step->[0] eq 'read') { unshift @read, yield($step->[1]) }
		else                      { push @rests, $step->[1] }
		$node = $from;
	}
	return [ @read, @mid, map { @$_ } @rests ];
}

my ($plain_dist, $plain_pred) = search(undef);
my %tsearch;
sub tsearch { my ($t) = @_; $tsearch{$t} //= [ search($t) ]; @{ $tsearch{$t} } }

# Candidates for a (state, token): reading the token as a shift item in the
# state does, then reducing a rule complete there with the token next;
# shortest first.
sub candidates {
	my ($s, $t, @reduce) = @_;
	my @c;
	for my $it (@{ $items{$s} }) {
		my ($r, $i) = @$it;
		next unless defined $rhs[$r][$i] && $rhs[$r][$i] eq $t;
		my $node = "$s $r $i 0";
		next unless defined $plain_dist->{$node};
		my @after = @{ $rhs[$r] }[ $i + 1 .. $#{ $rhs[$r] } ];
		next unless defined slen(@after);
		push @c, sentence($plain_pred, $node, $t, map { yield($_) } @after);
	}
	my ($dist, $pred) = tsearch($t);
	for my $r (@reduce) {
		my $node = "$s $r " . scalar(@{ $rhs[$r] }) . " 1";
		push @c, sentence($pred, $node) if defined $dist->{$node};
	}
	my %seen;
	return grep { !$seen{"@$_"}++ } sort { @$a <=> @$b || "@$a" cmp "@$b" } @c;
}

# --- The referees.
my $tmp = "$work/sentence";
sub run {
	my ($toks) = @_;
	open my $f, '>', $tmp or die "sentences.pl: $tmp: $!\n";
	print $f "$_\n" for @$toks;
	close $f;
	my ($pv, $pg) = split /\n/, `"$work/referee-prec-trace" -g <"$tmp" 2>"$tmp.trace"`;
	my ($nv, $ng) = split /\n/, `"$work/referee-noprec" -g <"$tmp"`;
	die "sentences.pl: a referee failed on @$toks\n" unless defined $pg && defined $ng;
	# Each state the parser enters, with the token next in the input: the
	# lookahead if it holds one, else the next token yylex will return (a
	# state whose only action is a default reduction reads none).
	my (%seen, $la, $nread);
	open my $tr, '<', "$tmp.trace" or die;
	while (<$tr>) {
		$nread++ if /^Reading a token$/;
		$la = $1 if /^Next token is token (\S+) /;
		undef $la if /^Shifting token /;
		$seen{ "$1 " . ($la // $toks->[ $nread // 0 ] // '$end') } = 1 if /^Entering state (\d+)$/;
	}
	return { verdict => $pv, grouping => $pg, differs => ($pv ne $nv || $pg ne $ng), seen => \%seen };
}

my (@sentences, %sid, %result);
sub keep {
	my ($toks, $r) = @_;
	my $k = "@$toks";
	unless ($sid{$k}) {
		push @sentences, [ $toks, $r ];
		$sid{$k} = sprintf 's%04d', scalar @sentences;
	}
	return $sid{$k};
}
sub tryall {    # the first candidate that reaches $s with $t next and on which the referees differ
	my ($s, $t, $want, @cands) = @_;
	my $n = 0;
	for my $c (@cands) {
		last if ++$n > $BOUND;
		my $r = $result{"@$c"} //= run($c);
		next unless $r->{differs} && $r->{seen}{"$s $t"};
		next if $want && $r->{verdict} ne $want;
		return keep($c, $r);
	}
	return undef;
}

my @uncovered;
for my $x (@res) {
	my $id = tryall($x->{state}, $x->{token}, undef, candidates($x->{state}, $x->{token}, $x->{prule}));
	$x->{ids} = defined $id ? [$id] : [];
	push @uncovered, $x unless defined $id;
}

# An error resolution is covered only beside an accepted neighbour: a sentence
# referee-prec accepts through the same state, reading another token there on
# which the referees also differ.
for my $x (grep { $_->{action} eq 'error' && @{ $_->{ids} } } @res) {
	my $id;
	for my $y (grep { $_->{state} == $x->{state} && $_->{token} ne $x->{token} } @res) {
		$id = tryall($y->{state}, $y->{token}, 'accept', candidates($y->{state}, $y->{token}, $y->{prule})) and last;
	}
	if (defined $id) { push @{ $x->{ids} }, $id }
	else             { $x->{ids} = []; push @uncovered, $x }
}

open my $rf, '>', "$out/resolutions.txt" or die;
print $rf "# One line per (rule, token) resolution in bison's report of perly.y (bison-conflicts.txt), in its order:\n",
  "# rule, token, action, and the ids in sentences.txt covering it (- when none). Generated by sentences.sh.\n";
print $rf join("\t", $_->{rule}, $_->{token}, $_->{action}, @{ $_->{ids} } ? join(',', @{ $_->{ids} }) : '-'), "\n" for @res;
open my $sf, '>', "$out/sentences.txt" or die;
print $sf "# One line per sentence: id, perly.y terminals, referee-prec's verdict and grouping (referee/README in\n",
  "# internal/chart/offline/README.md). Generated by sentences.sh.\n";
for my $i (0 .. $#sentences) {
	my ($toks, $r) = @{ $sentences[$i] };
	print $sf join("\t", sprintf('s%04d', $i + 1), "@$toks", $r->{verdict}, $r->{grouping}), "\n";
}
open my $uf, '>', "$out/uncovered-reviewed.txt" or die;
print $uf "# Resolutions no sentence covers, each under its reason (spec section 8). Proposed by sentences.sh; reviewed before commit.\n";
print $uf "reason: generation timed out\n" if @uncovered;
print $uf "$_->{rule} $_->{token}\n" for sort { $a->{rule} <=> $b->{rule} || $a->{token} cmp $b->{token} } @uncovered;
printf STDERR "sentences.pl: %d resolutions, %d covered, %d uncovered, %d sentences\n", scalar @res, @res - @uncovered, scalar @uncovered, scalar @sentences;
