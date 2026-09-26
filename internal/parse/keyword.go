// ABOUTME: The parser's keyword table: named unary versus list operator, and where each binds.
// ABOUTME: Measured against perl by argument count — a unary takes one, a list operator takes all.

package parse

// namedUnary is every builtin that takes exactly ONE argument without parens,
// so `KW $x, $y` parses as `KW($x), $y`.
//
// Measured, not transcribed. Every one of perl's 251 keywords was deparsed as
//
//	perl -MO=Deparse,-p -e 'my $x; my $y; KW $x, $y;'
//
// and classified by where the comma landed:
//
//	(length($x), $y)     the comma is OUTSIDE   -> named unary
//	print($x, $y)        the comma is INSIDE    -> list operator
//
// `-p` matters: without it Deparse omits the parens that carry the answer.
//
// 81 of them. The spec's §4.1 table lists a subset by hand; this is the whole
// set as perl 5.42 reports it.
//
// The count read 77 while the table held 76 -- an off-by-one that predates
// the four added then (delete, do, exists, goto), counted rather than
// assumed. Each of those four was classified by the method above, as was
// `undef`, which the original sweep missed:
//
//	(delete $h{'a'}, $y)   comma outside  -> named unary
//	(exists $h{'a'}, $y)   comma outside  -> named unary
//	((goto $x), $y)        comma outside  -> named unary
//	((do $x), $y)          comma outside  -> named unary
//	(undef($x), $y)        comma outside  -> named unary
//
// `undef` is two operators wearing one word: the niladic one needed no entry
// here, because a word with nothing to take is already a complete term, and
// only the spelling WITH an argument refused. Declare a lexical first and perl
// optimises the unary away -- `my $x; undef $x, $y` deparses to `($x = undef,
// $y)`, no `undef(...)` in sight -- but the comma placement, which is the
// classifier this method names, is the same either way.
var namedUnary = map[string]bool{
	"abs": true, "alarm": true, "caller": true, "chdir": true,
	"chomp": true, "chop": true, "chr": true, "chroot": true,
	"close": true, "closedir": true, "cos": true, "dbmclose": true,
	"defined": true, "delete": true, "do": true, "each": true,
	"eof": true, "eval": true, "exists": true, "goto": true,
	"exit": true, "exp": true, "fileno": true, "getc": true,
	"getgrgid": true, "getgrnam": true, "gethostbyname": true,
	"getnetbyname": true, "getpeername": true, "getpgrp": true,
	"getprotobyname": true, "getpwnam": true, "getpwuid": true,
	"getsockname": true, "gmtime": true, "hex": true, "int": true,
	"keys": true, "lc": true, "lcfirst": true, "length": true,
	"localtime": true, "lock": true, "log": true, "lstat": true,
	"oct": true, "ord": true, "pop": true, "pos": true,
	"prototype": true, "quotemeta": true, "rand": true, "readdir": true,
	"readline": true, "readlink": true, "ref": true, "reset": true,
	"rewinddir": true, "rmdir": true, "scalar": true, "sethostent": true,
	"setnetent": true, "setprotoent": true, "setservent": true,
	"shift": true, "sin": true, "sleep": true, "sqrt": true,
	"srand": true, "stat": true, "study": true, "tell": true,
	"telldir": true, "tied": true, "uc": true, "ucfirst": true,
	"umask": true, "undef": true, "untie": true, "values": true,
	"write": true,

	// The audit against perl's own keyword list -- all 266 names in
	// `regen/keywords.pl` differenced against these three tables -- found
	// `glob` and `readpipe` absent from all three, and perl reads each as a
	// named unary:
	//
	//	perl -MO=Deparse,-p -e 'our ($a,$b); my $z = readpipe $a, $b;'
	//	((my($z) = `$a`), $b);               comma OUTSIDE
	//	perl -MO=Deparse,-p -e 'our ($a,$b); my $z = glob $a;'
	//	use File::Glob (); (my($z) = glob($a));
	//	perl -MO=Deparse,-p -e 'our ($a,$b); my $z = glob $a, $b;'
	//	Too many arguments for glob         one argument, so not a list op
	//
	// Both were the wrong-tree-at-Unknown-zero defect: absent from every
	// table, `glob $pattern` emitted `glob();$pattern;` -- an empty call with
	// the argument detached as its own statement. TestParenlessBuiltinArity
	// asserts the canon, which is the only instrument that saw it.
	"glob": true, "readpipe": true,

	// `my` and `local` are named unaries in perly.y too (§4.6), but they are
	// declarations here and parseDeclaration owns them. Listed in neither
	// table so the declaration path is not shadowed.
}

// listOperator is every builtin that swallows the whole comma list, so
// `KW $x, $y` parses as `KW($x, $y)`.
//
// Same measurement, 82 of them.
var listOperator = map[string]bool{
	"accept": true, "bind": true, "binmode": true, "bless": true,
	"chmod": true, "chown": true, "connect": true, "crypt": true,
	"dbmopen": true, "die": true, "exec": true, "fcntl": true,
	"flock": true, "formline": true, "gethostbyaddr": true,
	"getnetbyaddr": true, "getpriority": true, "getprotobynumber": true,
	"getservbyname": true, "getservbyport": true, "getsockopt": true,
	"grep": true, "index": true, "ioctl": true, "kill": true,
	"link": true, "listen": true, "map": true, "mkdir": true,
	"msgctl": true, "msgget": true, "msgrcv": true, "msgsnd": true,
	"open": true, "opendir": true, "pack": true, "pipe": true,
	"print": true, "printf": true, "push": true, "read": true,
	"recv": true, "rename": true, "reverse": true, "rindex": true,
	"seek": true, "seekdir": true, "select": true, "semctl": true,
	"semget": true, "semop": true, "send": true, "setpgrp": true,
	"setpriority": true, "setsockopt": true, "shmctl": true,
	"shmget": true, "shmread": true, "shmwrite": true, "shutdown": true,
	"socket": true, "socketpair": true, "sort": true, "splice": true,
	"sprintf": true, "substr": true, "symlink": true, "syscall": true,
	"sysopen": true, "sysread": true, "sysseek": true, "system": true,
	"syswrite": true, "tie": true, "truncate": true, "unlink": true,
	"unpack": true, "unshift": true, "utime": true, "vec": true,
	"waitpid": true, "warn": true,

	// `split` and `join` are list operators too. The sweep missed them
	// because perl constant-folds `join ",", $x, $y` and rewrites
	// `split /$x/, $y` into a three-argument form, so neither deparsed to
	// the shape the classifier looked for. Both verified by hand:
	//
	//	perl -MO=Deparse -e 'my @a=(1,2); print join ",", @a, "x";'
	//	print join(',', @a, 'x');
	"split": true, "join": true,

	// `say` is feature-gated, so the sweep -- which ran without `use v5.38`
	// for the classification half -- saw it as a plain bareword. It is a
	// list operator with a filehandle slot exactly like `print`.
	"say": true,

	// `atan2` was absent from all three tables, so `atan2 $var, 1` emitted
	// `atan2();$var , 1;` -- an empty call with the whole argument list
	// detached, at Unknown=0. It is a two-argument builtin and the comma is
	// inside:
	//
	//	perl -MO=Deparse,-p -e 'our ($a,$b); my $z = atan2 $a, $b;'
	//	(my($z) = atan2($a, $b));
	"atan2": true,

	// `return` is level 7 in perly.y and swallows its list the same way,
	// but it is a statement form here and statementKeywords owns it.
}

// The audit of these three tables against perl's own keyword list.
//
// The tables above were built by deparsing keywords one at a time, but the
// SET they were drawn from was never differenced against perl's. `atan2` and
// `fc` were found by accident, which is the reason for this pass: all 266
// names in `/home/perigrin/dev/perl5/regen/keywords.pl` were differenced
// against namedUnary, listOperator and niladicParse, and each of the 78
// absent names was then put to perl in expression position:
//
//	perl -MO=Deparse,-p -e 'our ($a,$b); my $z = KW $a, $b;'
//
// Most of the 78 perl itself rejects there (`if`, `while`, `sub`, `else`,
// `__DATA__`, the phasers) or reads as an infix operator (`eq`, `cmp`, `lt`,
// `xor`) or a quote-like (`q`, `qw`, `m`, `s`, `tr`). Of the names perl
// ACCEPTS and hands an argument to, these are owned elsewhere and correctly
// absent from all three tables:
//
//	my, our, local        parseDeclaration
//	return, last, next, redo, require   statement forms
//	not                   a unary operator, not a call
//
// That left exactly four genuine gaps: `atan2` (list operator, added above),
// `glob` and `readpipe` (named unaries, added above), and `fc`.
//
// `fc` is TABLED DELIBERATELY NOT, and filed as 01a0dfbd. It is
// feature-gated, and unenabled perl
// does not read it as a builtin at all:
//
//	perl -MO=Deparse,-p -e 'our $a; my $z = fc $a;'
//	(my($z) = $a->fc);                     a METHOD call
//	perl -MO=Deparse,-p -e 'use v5.42; our $a; my $z = fc $a, $b;'
//	(fc($a), $b);                          the builtin, a named unary
//
// Tabling it unconditionally would make us disagree with perl on unfeatured
// source, which is the larger half of the corpus. The disagreement either way
// is recorded and measured: ONE file in perl.git t/ has a parenless `fc` at
// all (`re/anyof.t:858`, `$mod_cp = ord fc $char;`), so neither reading
// changes the ratchet. `evalbytes`, `break`, `default` and `isa` are gated
// the same way and absent for the same reason; `isa` is in fact an infix
// operator under `use feature 'isa'`, not a call at all.
//
// The audit's measurement lives in TestParenlessBuiltinArity and
// TestParenlessArityNoDetachedCall: an absent name emits `NAME()` with its
// arguments detached, at Unknown=0, so the canon is the only instrument that
// can see the gap.

// niladicParse is a builtin taking no argument at all, so nothing follows it.
// The lexer has its own copy for the expect state; this one is the parser's,
// because a niladic in expression position still needs a node.
//
// Kept as a separate map rather than imported from the lexer: the lexer's
// question is "does an operator come next", the parser's is "does this take
// an operand", and they are the same set today only by coincidence.
var niladicParse = map[string]bool{
	"time": true, "times": true, "wantarray": true, "wait": true,
	"fork": true, "getppid": true, "getlogin": true,
	"getgrent": true, "gethostent": true, "getnetent": true,
	"getprotoent": true, "getpwent": true, "getservent": true,
	"setgrent": true, "setpwent": true,
	"endgrent": true, "endhostent": true, "endnetent": true,
	"endprotoent": true, "endpwent": true, "endservent": true,
}

// The `-e`, `-f`, `-d` family lives in `internal/lexer`, which forms each of
// them as ONE operator token and answers `lexer.IsFileTest`. perl returns
// UNIOP for them (toke.c:6255-6261 FTST), so they bind exactly like a named
// unary, and `parseTerm` reads them at `bpNamedUnary` for that reason.
