#!/usr/bin/perl
# ABOUTME: Reads a toy grammar (stdin) and writes its bison referee, its fixture header, or its sentences (stdout).
# ABOUTME: Usage: toys.pl bison|header|sentences <toy.y>; build.sh runs all three.
use strict;
use warnings;

my $mode = shift // '';
die "usage: toys.pl bison|header|sentences <toy.y\n" unless $mode =~ /^(bison|header|sentences)$/;
my ($decls, $rules) = split /^%%\n/m, do { local $/; <STDIN> }, 2;
die "toys.pl: no %% in the toy\n" unless defined $rules;
my ($max) = $decls =~ /sentences up to (\d+) terminals/ or die "toys.pl: no 'sentences up to N terminals'\n";

my (@levels, @tokens);    # @levels: [assoc, tokens...], loosest first
for (split /\n/, $decls) {
	next unless /^%(token|left|right|nonassoc)\s+(.*)/;
	push @tokens, split ' ', $2;
	push @levels, [ $1, split ' ', $2 ] unless $1 eq 'token';
}

# Rules, one alternative per line: `lhs : syms [%prec X]` or `| syms [%prec X]`, `;` alone ends a rule.
my (@alts, $lhs);    # @alts: [lhs, [syms], prec]
for (split /\n/, $rules) {
	next if /^\s*;?\s*$/;
	if (/^(\w+)\s*:(.*)$/) { $lhs = $1; $_ = $2 }
	elsif (/^\s*\|(.*)$/) { $_ = $1 }
	else                  { die "toys.pl: cannot read rule line: $_\n" }
	my ($syms, $prec) = /^(.*?)(?:%prec\s+(\w+))?\s*$/;
	my @s = split ' ', $syms;
	push @alts, [ $lhs, \@s, $prec ];
}
my %nonterm = map { $_->[0] => 1 } @alts;

if ($mode eq 'header') {
	print "level $_->[0] @$_[1 .. $#$_]\n" for @levels;
	for my $i (0 .. $#alts) {
		my ($l, $s, $p) = @{ $alts[$i] };
		print "rule $i $l:", (@$s ? " @$s" : ''), (defined $p ? " %prec $p" : ''), "\n";
	}
	exit;
}

if ($mode eq 'sentences') {
	# Every string the rules derive of up to $max terminals, by length, then every proper prefix of one.
	my %lang;    # nonterminal => length => {string => 1}
	for my $n (0 .. $max) {
		for (my $changed = 1; $changed;) {    # unit and empty rules feed one length from itself
			$changed = 0;
			for my $a (@alts) {
				for my $str (derive($a->[1], $n, \%lang)) {
					$changed = 1 unless $lang{ $a->[0] }{$n}{$str}++;
				}
			}
		}
	}
	my %out;
	for my $str (map { keys %{ $lang{ $alts[0][0] }{$_} // {} } } 1 .. $max) {
		my @w = split ' ', $str;
		$out{"@w[0 .. $_]"} = 1 for 0 .. $#w;
	}
	print "$_\n" for sort { split(' ', $a) <=> split(' ', $b) or $a cmp $b } keys %out;
	exit;
}

# The strings of exactly $n terminals that the symbols @$syms derive, from %$lang's strings so far.
sub derive {
	my ($syms, $n, $lang) = @_;
	return $n == 0 ? ('') : () unless @$syms;
	my ($first, @rest) = @$syms;
	my @out;
	for my $k (0 .. $n) {
		my @heads = $nonterm{$first} ? keys %{ $lang->{$first}{$k} // {} } : $k == 1 ? ($first) : ();
		for my $h (@heads) {
			push @out, map { join ' ', grep { $_ ne '' } $h, $_ } derive(\@rest, $n - $k, $lang);
		}
	}
	return @out;
}

# bison: the toy with a grouping trace on every rule and a driver reading one sentence per line.
print "%{\n#include <stdio.h>\n#include <stdlib.h>\n#include <string.h>\n",
  "int yylex(void);\nvoid yyerror(const char *);\nstatic void trace(int, int, int);\n%}\n%locations\n",
  "%initial-action { \@\$.first_line = \@\$.last_line = 0; }\n", $decls, "%%\n";
for my $i (0 .. $#alts) {
	my ($l, $s, $p) = @{ $alts[$i] };
	print "$l: ", (@$s ? "@$s" : '%empty'), (defined $p ? " %prec $p" : ''), " { trace($i, \@\$.first_line, \@\$.last_line); };\n";
}
print "%%\nstatic const struct { const char *name; int tok; } table[] = {\n";
print qq|\t{ "$_", $_ },\n| for @tokens;
print <<'C';
};
static char *next;
static int ntok, errs;
static char grouping[1 << 16];
static size_t glen;

/* A location is a half-open span of 0-based terminal indexes. */
int yylex(void) {
	char *w = strtok(next, " ");
	next = NULL;
	yylloc.first_line = yylloc.last_line = ntok;
	if (!w)
		return 0;
	yylloc.last_line = ++ntok;
	for (size_t i = 0; i < sizeof table / sizeof table[0]; i++)
		if (!strcmp(w, table[i].name))
			return table[i].tok;
	fprintf(stderr, "toy: unknown terminal %s\n", w);
	exit(2);
}

void yyerror(const char *s) {
	(void)s;
	errs++;
}

/* The grouping: each reduction as rule:from-to, in reduction order. */
static void trace(int rule, int from, int to) {
	int n = snprintf(grouping + glen, sizeof grouping - glen, "%s%d:%d-%d", glen ? "," : "", rule, from, to);
	if (n < 0 || (size_t)n >= sizeof grouping - glen) {
		fprintf(stderr, "toy: grouping overflows %zu bytes\n", sizeof grouping);
		exit(2);
	}
	glen += n;
}

/* One sentence per line on stdin; prints `accept<TAB>grouping<TAB>sentence` or `refuse<TAB>-<TAB>sentence`. */
int main(void) {
	char line[4096], copy[4096];
	while (fgets(line, sizeof line, stdin)) {
		line[strcspn(line, "\r\n")] = 0;
		strcpy(copy, line);
		next = copy;
		ntok = errs = 0;
		glen = 0;
		if (yyparse() == 0 && errs == 0)
			printf("accept\t%s\t%s\n", grouping, line);
		else
			printf("refuse\t-\t%s\n", line);
	}
	return 0;
}
C
