#!/usr/bin/perl
# ABOUTME: Rewrites perly.y (stdin) into a bison referee grammar with a C driver (stdout); argument prec or noprec.
# ABOUTME: Strips every action but yyerror calls, omits catch_paren: empty; noprec drops precedence and builds GLR.
use strict;
use warnings;

my $mode = shift // '';
die "usage: strip.pl prec|noprec <perly.y\n" unless $mode =~ /^(no)?prec$/;
my $noprec = $mode eq 'noprec';

my ($decls, $rules) = split /^%%.*\n/m, do { local $/; <STDIN> }, 2;
die "strip.pl: no %% in perly.y\n" unless defined $rules;

# Declarations: no value types, no %union, no %type, no pure API.
$decls =~ s{/\*.*?\*/}{}gs;
$decls =~ s/^%union\s*\{.*?^\}\s*$//ms;
$decls =~ s/^%(type|define)\b.*$//mg;
$decls =~ s/<\w+>//g;
$decls =~ s/^%(left|right|nonassoc)\b/%token/mg if $noprec;
my @tokens = map { /^%(?:token|left|right|nonassoc)\s+(.*)/ ? split ' ', $1 : () } split /\n/, $decls;

# Rules: tokenize into comments, actions, words and punctuation, then re-emit.
my @alts;    # [lhs, [symbols...]]
my ($lhs, $cur);
pos($rules) = 0;
while (pos($rules) < length $rules) {
	if ($rules =~ /\G(?:\s+|\/\*.*?\*\/|\/\/[^\n]*)/gcs) {
		next;
	} elsif ($rules =~ /\G\{/gc) {
		my $body = action($rules);
		# Conditional refusals belong to catch_paren: empty, which is omitted.
		push @$cur, '{ yyerror("refused in action"); }'
		  if $body =~ /\byyerror\b/ && $body !~ /\bif\s*\(/;
	} elsif ($rules =~ /\G(\w+)(?:\[[\w.]+\])?(?:\s|(?>\/\*.*?\*\/))*:/gcs) {
		$lhs = $1;
		push @alts, [ $lhs, $cur = [] ];
	} elsif ($rules =~ /\G\|/gc) {
		push @alts, [ $lhs, $cur = [] ];
	} elsif ($rules =~ /\G(?:;|\[[\w.]+\])/gc) {    # a name on a stripped action goes with it
	} elsif ($rules =~ /\G%prec\s+(\w+)/gc) {
		push @$cur, "%prec $1" unless $noprec;
	} elsif ($rules =~ /\G(%empty|\w+)(?:\[[\w.]+\])?/gc) {
		push @$cur, $1;
	} else {
		die "strip.pl: cannot read perly.y rules at: ", substr($rules, pos($rules), 40), "\n";
	}
}
@alts = grep { !($_->[0] eq 'catch_paren' && "@{$_->[1]}" eq 'empty') } @alts;

print "%{\n#include <stdio.h>\n#include <stdlib.h>\n#include <string.h>\n",
  "int yylex(void);\nvoid yyerror(const char *);\n";
print "static int merge(int a, int b) { (void)b; return a; }\n" if $noprec;
print "%}\n", $noprec ? "%glr-parser\n" : '', $decls, "%%\n";
# shortcut: %merge runs every derivation's actions, so a refusal in any one refuses the span; revisit if a case needs bison's choice.
for my $a (@alts) {
	print "$a->[0]: @{$a->[1]}", ($noprec ? ' %merge <merge>' : ''), ";\n";
}
print "%%\n", <<'C';
static int errs, ntok, firsterr, eoferrs = -1;
static const struct { const char *name; int tok; } table[] = {
C
print qq|\t{ "$_", $_ },\n| for @tokens;
print <<'C';
};

int yylex(void) {
	char line[256];
	if (!fgets(line, sizeof line, stdin)) {
		if (eoferrs < 0)
			eoferrs = errs;
		return 0;
	}
	line[strcspn(line, "\r\n")] = 0;
	ntok++;
	for (size_t i = 0; i < sizeof table / sizeof table[0]; i++)
		if (!strcmp(line, table[i].name))
			return table[i].tok;
	fprintf(stderr, "referee: unknown terminal %s\n", line);
	exit(2);
}

void yyerror(const char *s) {
	(void)s;
	if (!errs++)
		firsterr = ntok;
}

/* -p: prefix mode. A sentence is refused on any error, even one yyparse recovers from. */
int main(int argc, char **argv) {
	int prefix = argc > 1 && !strcmp(argv[1], "-p");
	yyparse();
	if (prefix ? eoferrs == 0 : errs == 0)
		puts(prefix ? "viable" : "accept");
	else
		printf("refuse@%d\n", firsterr);
	return 0;
}
C

sub action {
	my $depth = 1;
	my $start = pos($_[0]);
	while ($depth) {
		$_[0] =~ /\G(?:"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'|\/\*.*?\*\/|[^{}"'\/]+|\/|(\{)|\})/gcs
		  or die "strip.pl: unterminated action\n";
		$depth += defined $1 ? 1 : $& eq '}' ? -1 : 0;
	}
	return substr $_[0], $start, pos($_[0]) - $start;
}
