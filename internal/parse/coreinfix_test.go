// ABOUTME: A word operator may be spelled with CORE:: -- `$a CORE::eq $b` -- and binds
// ABOUTME: exactly as the bare word does; canon keeps the spelling.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestCoreQualifiedInfix: measured on 5.42.0 with -MO=Deparse, each of eq ne
// lt gt le ge cmp x and or xor isa takes a CORE:: prefix and means the bare
// operator: `$r = $a CORE::eq $b` is `$r = $a eq $b`, and `$r = $a CORE::or
// $b` is `$b unless $r = $a`. PerlOnJava unit/core_qualified_infix.t.
func TestCoreQualifiedInfix(t *testing.T) {
	for qualified, bare := range map[string]string{
		`ok($dot CORE::eq '.', 'x');`:     `ok($dot eq '.', 'x');`,
		`ok(2 CORE::lt 3, 'y');`:          `ok(2 lt 3, 'y');`,
		`ok((2 CORE::cmp 10) gt 0, 'z');`: `ok((2 cmp 10) gt 0, 'z');`,
		`$r = $a CORE::or $b;`:            `$r = $a or $b;`,
		`$r = $a CORE::x 3;`:              `$r = $a x 3;`,
	} {
		root := parse.Parse([]byte(qualified))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", qualified, shape(root))
			continue
		}
		got := strings.ReplaceAll(shape(root), "CORE::", "")
		if want := shape(parse.Parse([]byte(bare))); got != want {
			t.Errorf("%q: shape %s, want %s as for %q", qualified, got, want, bare)
		}
		canon := parse.Canon(root, []byte(qualified))
		if !strings.Contains(canon, "CORE::") {
			t.Errorf("%q: canon %q drops the CORE:: spelling", qualified, canon)
		}
	}
}
