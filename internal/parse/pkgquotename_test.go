// ABOUTME: After `package` a quote-operator word is the package's name:
// ABOUTME: `package q;` declares the package q, not an unterminated q//.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestPackageNameIsQuoteLikeWord: measured on 5.44.0, perl compiles
// `package W;`, `package W 1.2;` and `package W { 1 }` for every quote
// operator word W. Deparse keeps the name:
//
//	$ perl -MO=Deparse -e 'package qq { 1 }'
//	{
//	    package qq;
//	    '???';
//	}
//
// The first five sources are Guacamole v0.008's
// internal/chart/testdata/guacamole/parses/Statements-PackageStatement-005.pl
// through -009.pl, byte for byte.
func TestPackageNameIsQuoteLikeWord(t *testing.T) {
	var srcs []string
	for _, w := range []string{"q", "qq", "qw", "qx", "qr"} {
		srcs = append(srcs, "use feature ':5.36';\npackage "+w+";")
	}
	for _, w := range []string{"m", "q", "qq", "qr", "qw", "qx", "s", "tr", "y"} {
		srcs = append(srcs, "package "+w+";", "package "+w+" 1.2;", "package "+w+" { 1 }")
	}
	for _, src := range srcs {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) || countUnknown(root) != 0 {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		d := firstOfKind(root, parse.Declaration)
		if d == nil {
			t.Errorf("%q: no Declaration: %s", src, shape(root))
			continue
		}
		name := strings.Fields(src[strings.LastIndex(src, "package ")+len("package "):])[0]
		name = strings.TrimSuffix(name, ";")
		if got := shape(d); got != "(package "+name+")" && !strings.HasPrefix(got, "(package "+name+" ") {
			t.Errorf("%q: the declaration does not name %q: %s", src, name, shape(d))
		}
	}
}
