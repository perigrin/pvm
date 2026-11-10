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
my $line0 = ($decls =~ tr/\n//) + 2;    # perly.y's line of the first rules line
sub line { $line0 + (substr($rules, 0, $_[0]) =~ tr/\n//) }

# Declarations: no value types, no %union, no %type, no pure API.
$decls =~ s{/\*.*?\*/}{}gs;
$decls =~ s/^%union\s*\{.*?^\}\s*$//ms;
$decls =~ s/^%(type|define)\b.*$//mg;
$decls =~ s/<\w+>//g;
my %hasprec = map { /^%(?:left|right|nonassoc)\s+(.*)/ ? map { $_ => 1 } split ' ', $1 : () } split /\n/, $decls;
$decls =~ s/^%(left|right|nonassoc)\b/%token/mg if $noprec;
my @tokens = map { /^%(?:token|left|right|nonassoc)\s+(.*)/ ? split ' ', $1 : () } split /\n/, $decls;

# Rules: tokenize into comments, actions, words and punctuation, then re-emit.
my @alts;    # [lhs, [symbols...], perly.y line, has %prec]
my ($lhs, $cur);
pos($rules) = 0;
while (pos($rules) < length $rules) {
	if ($rules =~ /\G(?:\s+|\/\*.*?\*\/|\/\/[^\n]*)/gcs) {
		next;
	} elsif ($rules =~ /\G\{/gc) {
		$alts[-1][2] //= line($-[0]);
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
		$alts[-1][2] //= line($-[0]);
		$alts[-1][3] = 1;
		push @$cur, "%prec $1" unless $noprec;
	} elsif ($rules =~ /\G(%empty|\w+)(?:\[[\w.]+\])?/gc) {
		$alts[-1][2] //= line($-[0]);
		push @$cur, $1;
	} else {
		die "strip.pl: cannot read perly.y rules at: ", substr($rules, pos($rules), 40), "\n";
	}
}
@alts = grep { !($_->[0] eq 'catch_paren' && "@{$_->[1]}" eq 'empty') } @alts;

print "%{\n#include <stdio.h>\n#include <stdlib.h>\n#include <string.h>\n",
  "int yylex(void);\nvoid yyerror(const char *);\nstatic void trace(int, int, int);\n";
print "static int merge(int a, int b) { (void)b; return a; }\n" if $noprec;
print "%}\n", $noprec ? "%glr-parser\n" : '', $decls, "%%\n";
# shortcut: %merge runs every derivation's actions, so one refusing derivation refuses the span where every one should have to;
# unreachable at 94e50866 (each kept refusal ends its rule with nothing after it), revisit if the pin moves.
my %istoken = map { $_ => 1 } @tokens;
for my $a (@alts) {
	# The grouping trace: a rule carrying a precedence (a %prec, or its last
	# terminal's) reports its perly.y line and the terminals it covers, in one
	# final action so that a kept refusal never becomes a mid-rule action.
	my ($last) = grep { $istoken{$_} } reverse @{ $a->[1] };
	if ($a->[3] || ($last && $hasprec{$last})) {
		my $t = "trace($a->[2], \@\$.first_line, \@\$.last_line);";
		if (@{ $a->[1] } && $a->[1][-1] =~ /^\{ (.*)\}$/) { $a->[1][-1] = "{ $t $1}" }
		else                                               { push @{ $a->[1] }, "{ $t }" }
	}
	print "$a->[0]: @{$a->[1]}", ($noprec ? ' %merge <merge>' : ''), ";\n";
}
print "%%\n", <<'C';
static int errs, ntok, firsterr, eoferrs = -1, prefix;
static char grouping[1 << 16];
static size_t glen;
static const struct { const char *name; int tok; } table[] = {
C
print qq|\t{ "$_", $_ },\n| for @tokens;
print <<'C';
};

int yylex(void) {
	char line[256];
	/* A location is a half-open span of 0-based terminal indexes, GRAMPROG's 0. */
	yylloc.first_line = yylloc.last_line = ntok;
	if (!fgets(line, sizeof line, stdin)) {
		if (eoferrs < 0)
			eoferrs = errs;
		return 0;
	}
	line[strcspn(line, "\r\n")] = 0;
	yylloc.last_line = ++ntok;
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

/* The grouping: each reduction of a rule carrying a precedence, as
 * line:from-to, in reduction order, up to the first error (in prefix mode,
 * up to the end of the prefix). */
static void trace(int line, int from, int to) {
	if (errs || (prefix && eoferrs >= 0))
		return;
	int n = snprintf(grouping + glen, sizeof grouping - glen, "%s%d:%d-%d", glen ? "," : "", line, from, to);
	if (n < 0 || (size_t)n >= sizeof grouping - glen) {
		fprintf(stderr, "referee: grouping overflows %zu bytes\n", sizeof grouping);
		exit(2);
	}
	glen += n;
}

/* -p: prefix mode. A sentence is refused on any error, even one yyparse recovers from.
 * -g: a second line, the grouping, or - if empty. */
int main(int argc, char **argv) {
	int group = 0;
	for (int i = 1; i < argc; i++)
		if (!strcmp(argv[i], "-p"))
			prefix = 1;
		else if (!strcmp(argv[i], "-g"))
			group = 1;
		else {
			fprintf(stderr, "usage: referee [-p] [-g] <terminals\n");
			return 2;
		}
#if YYDEBUG
	yydebug = 1; /* sentences.sh's build: which state reads which lookahead */
#endif
	yyparse();
	if (prefix ? eoferrs == 0 : errs == 0)
		puts(prefix ? "viable" : "accept");
	else
		printf("refuse@%d\n", firsterr);
	if (group)
		puts(glen ? grouping : "-");
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
