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
// 77 of them. The spec's §4.1 table lists a subset by hand; this is the whole
// set as perl 5.42 reports it.
var namedUnary = map[string]bool{
	"abs": true, "alarm": true, "caller": true, "chdir": true,
	"chomp": true, "chop": true, "chr": true, "chroot": true,
	"close": true, "closedir": true, "cos": true, "dbmclose": true,
	"defined": true, "each": true, "eof": true, "eval": true,
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
	"umask": true, "untie": true, "values": true, "write": true,

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

	// `return` is level 7 in perly.y and swallows its list the same way,
	// but it is a statement form here and statementKeywords owns it.
}

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

// fileTests are the `-e`, `-f`, `-d` family. perl returns UNIOP for them
// (toke.c:6255-6261 FTST), so they bind exactly like a named unary.
var fileTests = map[byte]bool{
	'e': true, 'f': true, 'd': true, 'r': true, 'w': true, 'x': true,
	's': true, 'z': true, 'l': true, 'p': true, 'S': true, 'b': true,
	'c': true, 't': true, 'u': true, 'g': true, 'k': true, 'T': true,
	'B': true, 'A': true, 'M': true, 'C': true, 'o': true, 'R': true,
	'W': true, 'X': true, 'O': true,
}
