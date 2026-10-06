// ABOUTME: Under `use feature 'defer'`, `defer BLOCK` is a whole statement: a block
// ABOUTME: after it is the next statement, never an argument.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestDeferIsAStatement: measured on 5.42.0,
//
//	$ perl -MO=Deparse -e 'use feature "defer"; no warnings;
//	      { defer { $l .= "o" } { defer { $l .= "i," } } $l .= "m"; }'
//	{
//	    defer { $l .= 'o'; }
//	    { defer { $l .= 'i,'; } }
//	    $l .= 'm';
//	}
//
// Read as a word taking a block, the `{` after the defer's block became an
// anonymous-hash argument: `defer {1;}{2};`. PerlOnJava unit/defer.t:83.
func TestDeferIsAStatement(t *testing.T) {
	src := `use feature "defer"; { defer { $l .= "o" } { defer { $l .= "i," } } $l .= "m"; }`
	root := parse.Parse([]byte(src))
	if containsKind(root, parse.Unknown) || containsKind(root, parse.AnonHash) {
		t.Errorf("%q: want two defers and a bare block; got %s", src, shape(root))
	}
	canon := parse.Canon(root, []byte(src))
	if again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon)); again != canon {
		t.Errorf("canon %q is not a fixpoint, reparses to %q", canon, again)
	}
	if ok, why := parse.Faithful(root, []byte(src)); !ok {
		t.Errorf("canon %q is not faithful: %s", canon, why)
	}
}
