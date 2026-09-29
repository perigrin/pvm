// ABOUTME: Every word perl's keyword() recognises, from regen/keywords.pl in the perl.git
// ABOUTME: checkout -- the list toke.c consults before any package qualification.

package lexer

// perlKeywords is regen/keywords.pl's list after its __END__, with the
// `+` and `-` override marks removed and NULL dropped: 266 words.
// ponytail: blead's list, not 5.42's. A keyword blead added, written with an
// apostrophe after it, ends the word where 5.42 reads a package name; take
// the list from a 5.42 checkout if that spelling ever turns up.
var perlKeywords = map[string]bool{
	"ADJUST": true, "AUTOLOAD": true, "BEGIN": true, "CHECK": true,
	"DESTROY": true, "END": true, "INIT": true, "UNITCHECK": true,
	"__CLASS__": true, "__DATA__": true, "__END__": true, "__FILE__": true,
	"__LINE__": true, "__PACKAGE__": true, "__SUB__": true, "abs": true,
	"accept": true, "alarm": true, "all": true, "and": true, "any": true,
	"atan2": true, "bind": true, "binmode": true, "bless": true, "break": true,
	"caller": true, "catch": true, "chdir": true, "chmod": true, "chomp": true,
	"chop": true, "chown": true, "chr": true, "chroot": true, "class": true,
	"close": true, "closedir": true, "cmp": true, "connect": true,
	"continue": true, "cos": true, "crypt": true, "dbmclose": true,
	"dbmopen": true, "default": true, "defer": true, "defined": true,
	"delete": true, "die": true, "do": true, "dump": true, "each": true,
	"else": true, "elsif": true, "endgrent": true, "endhostent": true,
	"endnetent": true, "endprotoent": true, "endpwent": true,
	"endservent": true, "eof": true, "eq": true, "eval": true,
	"evalbytes": true, "exec": true, "exists": true, "exit": true, "exp": true,
	"fc": true, "fcntl": true, "field": true, "fileno": true, "finally": true,
	"flock": true, "for": true, "foreach": true, "fork": true, "format": true,
	"formline": true, "ge": true, "getc": true, "getgrent": true,
	"getgrgid": true, "getgrnam": true, "gethostbyaddr": true,
	"gethostbyname": true, "gethostent": true, "getlogin": true,
	"getnetbyaddr": true, "getnetbyname": true, "getnetent": true,
	"getpeername": true, "getpgrp": true, "getppid": true, "getpriority": true,
	"getprotobyname": true, "getprotobynumber": true, "getprotoent": true,
	"getpwent": true, "getpwnam": true, "getpwuid": true, "getservbyname": true,
	"getservbyport": true, "getservent": true, "getsockname": true,
	"getsockopt": true, "given": true, "glob": true, "gmtime": true,
	"goto": true, "grep": true, "gt": true, "hex": true, "if": true,
	"index": true, "int": true, "ioctl": true, "isa": true, "join": true,
	"keys": true, "kill": true, "last": true, "lc": true, "lcfirst": true,
	"le": true, "length": true, "link": true, "listen": true, "local": true,
	"localtime": true, "lock": true, "log": true, "lstat": true, "lt": true,
	"m": true, "map": true, "method": true, "mkdir": true, "msgctl": true,
	"msgget": true, "msgrcv": true, "msgsnd": true, "my": true, "ne": true,
	"next": true, "no": true, "not": true, "oct": true, "open": true,
	"opendir": true, "or": true, "ord": true, "our": true, "pack": true,
	"package": true, "pipe": true, "pop": true, "pos": true, "print": true,
	"printf": true, "prototype": true, "push": true, "q": true, "qq": true,
	"qr": true, "quotemeta": true, "qw": true, "qx": true, "rand": true,
	"read": true, "readdir": true, "readline": true, "readlink": true,
	"readpipe": true, "recv": true, "redo": true, "ref": true, "rename": true,
	"require": true, "reset": true, "return": true, "reverse": true,
	"rewinddir": true, "rindex": true, "rmdir": true, "s": true, "say": true,
	"scalar": true, "seek": true, "seekdir": true, "select": true,
	"semctl": true, "semget": true, "semop": true, "send": true,
	"setgrent": true, "sethostent": true, "setnetent": true, "setpgrp": true,
	"setpriority": true, "setprotoent": true, "setpwent": true,
	"setservent": true, "setsockopt": true, "shift": true, "shmctl": true,
	"shmget": true, "shmread": true, "shmwrite": true, "shutdown": true,
	"sin": true, "sleep": true, "socket": true, "socketpair": true,
	"sort": true, "splice": true, "split": true, "sprintf": true, "sqrt": true,
	"srand": true, "stat": true, "state": true, "study": true, "sub": true,
	"substr": true, "symlink": true, "syscall": true, "sysopen": true,
	"sysread": true, "sysseek": true, "system": true, "syswrite": true,
	"tell": true, "telldir": true, "tie": true, "tied": true, "time": true,
	"times": true, "tr": true, "truncate": true, "try": true, "uc": true,
	"ucfirst": true, "umask": true, "undef": true, "unless": true,
	"unlink": true, "unpack": true, "unshift": true, "untie": true,
	"until": true, "use": true, "utime": true, "values": true, "vec": true,
	"wait": true, "waitpid": true, "wantarray": true, "warn": true,
	"when": true, "while": true, "write": true, "x": true, "xor": true,
	"y": true,
}

// IsKeyword reports whether word is one of perl's keywords, feature-gated
// ones included: whether it is a keyword HERE is the caller's question.
func IsKeyword(word string) bool { return perlKeywords[word] }
