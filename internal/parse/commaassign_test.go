// ABOUTME: A trailing comma before an assignment operator ends a list operator's
// ABOUTME: arguments: `substr $x, 0, 1, = "a"` assigns to the substr call.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestTrailingCommaBeforeAssignment: `=` cannot begin an element, so the
// comma before it separates nothing and the list ends there; the list
// operator is then a term, and the assignment takes it as its left side.
// Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'substr $x, 0, 1, = "\x{100}";'
//	substr($x, 0, 1) = "\x{100}";
//	$ perl -MO=Deparse -e 'substr $x, 0, 1, .= "a";'
//	substr($x, 0, 1) .= 'a';
//
// perl.git t/op/utf8cache.t:70.
func TestTrailingCommaBeforeAssignment(t *testing.T) {
	for src, same := range map[string]string{
		`substr $x, 0, 1, = "\x{100}";`: `substr($x, 0, 1) = "\x{100}";`,
		`substr $x, 0, 1, .= "a";`:      `substr($x, 0, 1) .= "a";`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", src, shape(root))
			continue
		}
		got := parse.Canon(root, []byte(src))
		want := parse.Canon(parse.Parse([]byte(same)), []byte(same))
		if got != want {
			t.Errorf("%q: canon %q, want %q as for %q", src, got, want, same)
		}
	}
}
