// ABOUTME: The parser's keyword shapes, derived from CORE.pmt, held to the sets perl was measured to give.
// ABOUTME: Named unary, list operator and niladic: a disagreement names the builtin it is about.

package parse

import (
	"bytes"
	"maps"
	"slices"
	"strings"
	"testing"
)

// The golden sets, each measured on perl 5.42.0 by deparsing every keyword as
//
//	perl -MO=Deparse,-p -e 'my $x; my $y; KW $x, $y;'
//
// and classifying it by where the comma landed:
//
//	(length($x), $y)     the comma is OUTSIDE   -> named unary
//	print($x, $y)        the comma is INSIDE    -> list operator
//
// and a niladic builtin by taking nothing at all, so that what follows it is
// an operator. `-p` matters: without it Deparse omits the parens that carry
// the answer. These were the parser's hand-kept tables, and CORE.pmt
// replaces them only because it agrees with them. Where CORE.pmt first
// disagreed, perl 5.42.0 was asked again, and the sets say what it answered.
//
// A builtin a feature gates (fc, any, isa and the rest of gatedWords) is
// left out of every set: spelled plainly with no feature on, it is a user's
// sub. So are dump and method, whose parse is their own (ownParse).
var goldenNamedUnary = []string{
	"abs", "alarm", "caller", "chdir", "chomp", "chop", "chr", "chroot",
	"close", "closedir", "cos", "dbmclose", "defined", "delete", "do", "each",
	"eof", "eval", "exists", "goto", "exit", "exp", "fileno", "getc",
	"getgrgid", "getgrnam", "gethostbyname", "getnetbyname", "getpeername",
	"getpgrp", "getprotobyname", "getpwnam", "getpwuid", "getsockname",
	"gmtime", "hex", "int", "keys", "lc", "lcfirst", "length", "localtime",
	"lock", "log", "lstat", "oct", "ord", "pop", "pos", "prototype",
	"quotemeta", "rand", "readdir", "readline", "readlink", "ref", "reset",
	"rewinddir", "rmdir", "scalar", "sethostent", "setnetent", "setprotoent",
	"setservent", "shift", "sin", "sleep", "sqrt", "srand", "stat", "study",
	"tell", "telldir", "tied", "uc", "ucfirst", "umask", "undef", "untie",
	"values", "write", "readpipe",
	// `glob` is a list operator, not a named unary as the tables had it:
	// their measurement read `glob $a, $b`'s "Too many arguments for glob"
	// as the comma being outside, but perl refuses it because the comma is
	// INSIDE, as for getprotobynumber below. toke.c reads glob with LOP, and
	// its prototype `_;` makes a list operator of a sub too:
	//
	//	perl -MO=Deparse,-p -e 'our ($a,$b); my $z = glob $a lt 5;'
	//	(my($z) = glob(($a lt 5)));        the comparison is inside
}

var goldenListOperator = []string{
	"accept", "bind", "binmode", "bless", "chmod", "chown", "connect", "crypt",
	"dbmopen", "die", "exec", "fcntl", "flock", "formline", "gethostbyaddr",
	"getnetbyaddr", "getpriority", "getprotobynumber", "getservbyname",
	"getservbyport", "getsockopt", "grep", "index", "ioctl", "kill", "link",
	"listen", "map", "mkdir", "msgctl", "msgget", "msgrcv", "msgsnd", "open",
	"opendir", "pack", "pipe", "print", "printf", "push", "read", "recv",
	"rename", "reverse", "rindex", "seek", "seekdir", "select", "semctl",
	"semget", "semop", "send", "setpgrp", "setpriority", "setsockopt",
	"shmctl", "shmget", "shmread", "shmwrite", "shutdown", "socket",
	"socketpair", "sort", "splice", "sprintf", "substr", "symlink", "syscall",
	"sysopen", "sysread", "sysseek", "system", "syswrite", "tie", "truncate",
	"unlink", "unpack", "unshift", "utime", "vec", "waitpid", "warn",
	"join", "say", "atan2", "glob",
	// return and not take the whole comma list. Each is parsed as its own
	// form, return a statement and not a unary operator, but the shape is
	// the same:
	//
	//	perl -MO=Deparse,-p -e 'sub f { return $a < 1, 2 }'
	//	(return ($a < 1), 2);
	//	perl -MO=Deparse,-p -e 'our ($a,$b); my @z = (not $a < $b, 2);'
	//	(my(@z) = (!(($a < $b), 2)));
	"return", "not",
	// `split` is a list operator too, but CORE.pmt has no line for it: it
	// has no prototype but typed positional parameters, and typed Perl has
	// no spelling for that yet (01a1113c). Until it does, the parser keeps
	// split's shape by hand, and this comparison leaves it out.
}

var goldenNiladic = []string{
	"time", "times", "wantarray", "wait", "fork", "getppid", "getlogin",
	"getgrent", "gethostent", "getnetent", "getprotoent", "getpwent",
	"getservent", "setgrent", "setpwent", "endgrent", "endhostent",
	"endnetent", "endprotoent", "endpwent", "endservent",
	// __FILE__, __LINE__ and __PACKAGE__ are terms, and continue, outside a
	// when block, takes nothing either: what follows each is an operator.
	//
	//	perl -MO=Deparse,-p -e 'my $z = __LINE__ + 1, 2;'
	//	((my($z) = 2), '???');               __LINE__ + 1, folded
	//	perl -MO=Deparse,-p -e 'my @z = (__FILE__ . "x", 2);'
	//	(my(@z) = ('-ex', 2));
	//	perl -MO=Deparse,-p -e 'my @z = (continue + 1);'
	//	(my(@z) = ((continue) + 1));
	"__FILE__", "__LINE__", "__PACKAGE__", "continue",
}

func goldenShapes() map[string]Shape {
	golden := map[string]Shape{}
	for _, set := range []struct {
		names []string
		shape Shape
	}{{goldenNamedUnary, ShapeUnary}, {goldenListOperator, ShapeList}, {goldenNiladic, ShapeNiladic}} {
		for _, name := range set.names {
			golden[name] = set.shape
		}
	}
	return golden
}

// TestCoreDerivedShapes: the named unaries, list operators and niladic
// builtins CORE.pmt derives are the ones perl was measured to give.
func TestCoreDerivedShapes(t *testing.T) {
	for _, d := range shapeDisagreements(plainKeywordShapes(coreShapes()), goldenShapes()) {
		t.Error(d)
	}
}

// TestCoreShapesUnknownNameHasNoShape: a name CORE.pmt does not declare is no
// keyword, so it parses as an ordinary sub call, never a named unary or a
// niladic.
func TestCoreShapesUnknownNameHasNoShape(t *testing.T) {
	p := &parser{}
	if s, ok := p.plainKeywordShape("frobnicate"); ok {
		t.Errorf("frobnicate derives the shape %v; CORE.pmt does not declare it", s)
	}
	if p.isPerlKeyword("frobnicate") {
		t.Error("frobnicate is a keyword; CORE.pmt does not declare it")
	}
	// A sub declared with no prototype takes the whole list, where a named
	// unary would leave the comma outside and a niladic take nothing:
	// measured, `sub frobnicate {} frobnicate $x, $y` deparses as
	// `frobnicate($x, $y)`.
	src := []byte("sub frobnicate {} frobnicate $x, $y;")
	if got := Canon(Parse(src), src); !strings.Contains(got, "frobnicate($x , $y)") {
		t.Errorf("frobnicate takes the whole list: %s", got)
	}
	// Bare, it is a call to a sub this parse has not seen, unresolved,
	// where a niladic builtin is resolved.
	call := Parse([]byte("frobnicate;"))
	for call.Kind != Call && len(call.Children) > 0 {
		call = call.Children[len(call.Children)-1]
	}
	if call.Kind != Call || call.Resolved {
		t.Errorf("frobnicate; is an unresolved call, got %v resolved=%v", call.Kind, call.Resolved)
	}
}

// TestCoreDerivedShapesCatchesMismatch: the comparison is not tautological.
// A golden set short a name, and a CORE.pmt line with a wrong `:unary`, are
// each reported as a disagreement naming the builtin.
func TestCoreDerivedShapesCatchesMismatch(t *testing.T) {
	derived := plainKeywordShapes(coreShapes())

	golden := goldenShapes()
	delete(golden, "abs")
	if got := shapeDisagreements(derived, golden); len(got) != 1 || !strings.Contains(got[0], "abs") {
		t.Errorf("a golden set without abs: got %q, want one disagreement naming abs", got)
	}

	src, _ := declaration("CORE")
	// delete has no prototype, so only `:unary` makes it a named unary;
	// without it, it derives a list operator.
	wrong := bytes.ReplaceAll(src, []byte("multi sub delete :unary"), []byte("multi sub delete"))
	if bytes.Equal(wrong, src) {
		t.Fatal("CORE.pmt has no `multi sub delete :unary` line to unmark")
	}
	protos, sigs, err := coreProtos(wrong)
	if err != nil {
		t.Fatalf("CORE.pmt with delete unmarked: %v", err)
	}
	got := shapeDisagreements(plainKeywordShapes(deriveShapes(protos, sigs)), goldenShapes())
	if len(got) != 1 || !strings.Contains(got[0], "delete") {
		t.Errorf("delete without :unary: got %q, want one disagreement naming delete", got)
	}
}

// shapeDisagreements is each builtin whose derived shape is not its golden
// one, in name order, naming the builtin first.
func shapeDisagreements(derived, golden map[string]Shape) []string {
	var out []string
	names := maps.Clone(golden)
	maps.Copy(names, derived)
	for _, name := range slices.Sorted(maps.Keys(names)) {
		d, dok := derived[name]
		g, gok := golden[name]
		switch {
		case !gok:
			out = append(out, name+": CORE.pmt derives "+d.String()+", the measured sets give it none")
		case !dok:
			out = append(out, name+": CORE.pmt derives no keyword shape, perl was measured to give "+g.String())
		case d != g:
			out = append(out, name+": CORE.pmt derives "+d.String()+", perl was measured to give "+g.String())
		}
	}
	return out
}

// TestCoreShapesReachTheParser: a call parses in the shape CORE.pmt
// derives, where the hand-kept tables said otherwise. glob and
// getprotobynumber are list operators, so a comparison after the argument
// is inside the call. Measured on perl 5.42.0 with -MO=Deparse,-p:
//
//	my @f = glob $a lt 5;               (my(@f) = glob(($a lt 5)))
//	my $p = getprotobynumber $a < 5;    (my $p = getprotobynumber(($a < 5)))
func TestCoreShapesReachTheParser(t *testing.T) {
	for src, want := range map[string]string{
		"my @f = glob $a lt 5;":            "my @f = glob($a lt 5);",
		"my $p = getprotobynumber $a < 5;": "my $p = getprotobynumber($a < 5);",
	} {
		if got := Canon(Parse([]byte(src)), []byte(src)); got != want {
			t.Errorf("%s\n  got  %s\n  want %s", src, got, want)
		}
	}
}

// TestGatedBuiltinsTakeTheirShapeUnderTheirFeature: a builtin a feature
// gates takes its CORE.pmt shape with the feature on, and is an unknown
// word's call with it off. Measured on perl 5.42.0 with -MO=Deparse,-p:
//
//	use feature "keyword_any"; my $r = any { $_ } @a;   CORE::any({$_;} @a)
//	use feature "keyword_all"; my $r = all { $_ } @a;   CORE::all({$_;} @a)
//	my $r = any { $_ } @a;                              (any { $_ } @a)
//	use feature "current_sub"; my $s = __SUB__ + 1;     (__SUB__ + 1)
//	use v5.16; my $s = __SUB__ + 1;                     (__SUB__ + 1)
//	my $s = __SUB__ + 1;                                ('__SUB__' + 1)
//
// Unfeatured, `any` and `all` are a block and a list after a word perl does
// not know, and `__SUB__` a bareword string.
func TestGatedBuiltinsTakeTheirShapeUnderTheirFeature(t *testing.T) {
	for _, c := range []struct {
		src, canon, call string
		resolved         bool
	}{
		{"use feature 'keyword_any'; my $r = any { $_ } @a;", "use feature 'keyword_any'; my $r = any({$_;}@a);", "any", true},
		{"use feature 'keyword_all'; my $r = all { $_ } @a;", "use feature 'keyword_all'; my $r = all({$_;}@a);", "all", true},
		{"my $r = any { $_ } @a;", "my $r = any {$_;}@a;", "any", false},
		{"my $r = all { $_ } @a;", "my $r = all {$_;}@a;", "all", false},
		{"use feature 'current_sub'; my $s = __SUB__ + 1;", "use feature 'current_sub'; my $s = __SUB__() + 1;", "__SUB__", true},
		{"use v5.16; my $s = __SUB__ + 1;", "use v5.16; my $s = __SUB__() + 1;", "__SUB__", true},
		{"my $s = CORE::__SUB__ + 1;", "my $s = CORE::__SUB__() + 1;", "CORE::__SUB__", true},
		{"my $s = __SUB__ + 1;", "my $s = __SUB__() + 1;", "__SUB__", false},
	} {
		root := Parse([]byte(c.src))
		if got := Canon(root, []byte(c.src)); got != c.canon {
			t.Errorf("%s\n  got  %s\n  want %s", c.src, got, c.canon)
		}
		call := findCallNamed(root, c.call)
		if call == nil {
			t.Errorf("%s: no call to %s", c.src, c.call)
		} else if call.Resolved != c.resolved {
			t.Errorf("%s: %s resolved=%v, want %v", c.src, c.call, call.Resolved, c.resolved)
		}
	}
}

func findCallNamed(n *Node, name string) *Node {
	if n.Kind == Call && n.Text == name {
		return n
	}
	for _, c := range n.Children {
		if got := findCallNamed(c, name); got != nil {
			return got
		}
	}
	return nil
}
