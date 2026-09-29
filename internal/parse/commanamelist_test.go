// ABOUTME: A `use` list of quoted names separated by commas, without parentheses,
// ABOUTME: is as literal as a qw() list: `use feature 'evalbytes', 'unicode_eval'`.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestCommaSeparatedNameList: literalNameList read qw(), one quoted string,
// and a parenthesised list, but not the bare comma list perl allows after
// `use`. The features it named were never turned on, so a parenless
// `evalbytes "..."` after them refused. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e "use feature 'evalbytes', 'unicode_eval';
//	      my \$x = evalbytes \"s\", 1;"
//	my $x = evalbytes 's', '???';
//
// perl.git t/op/evalbytes.t:19, and lines 32-35 after it.
func TestCommaSeparatedNameList(t *testing.T) {
	src := `use feature 'evalbytes', 'unicode_eval'; my $x = evalbytes "s", 1;`
	root := parse.Parse([]byte(src))
	if containsKind(root, parse.Unknown) {
		t.Fatalf("%q: perl accepts this; got %s", src, shape(root))
	}
	call := findCall(root, "evalbytes")
	if call == nil || len(call.Children) != 1 {
		t.Errorf("%q: evalbytes takes one argument, the string; got %s", src, shape(root))
	}
}
