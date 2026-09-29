// ABOUTME: A signature sub's canon writes the signature after the name with a space,
// ABOUTME: as the source does, not as an assignment to the name.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestSignatureCanonHasNoAssignment: the signature is a parenthesised list
// child of the sub's Declaration, and canon joined it to the name with ` = `,
// the separator for a variable's initialiser. The tree was right and the
// emission was not Perl -- measured on 5.42.0:
//
//	$ perl -e 'use feature "signatures"; sub f = ($a) { $a }'
//	Illegal declaration of subroutine main::f at -e line 1.
//
// Deparse writes `sub f ($a) {`, and so must canon.
func TestSignatureCanonHasNoAssignment(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{`use feature "signatures"; sub f ($a) { $a }`, `use feature "signatures"; sub f ($a) {$a;}`},
		{`use feature "signatures"; sub h ($x, $y) { 1 }`, `use feature "signatures"; sub h ($x, $y) {1;}`},
		{`use feature "signatures"; sub e () { 1 }`, `use feature "signatures"; sub e () {1;}`},
		// One element is the element's own node with Paren set, not a List,
		// so the parentheses are canon's to write: `sub f $x = 1 {...}` is
		// not a signature at all.
		{`use feature "signatures"; sub d ($x = 1) { $x }`, `use feature "signatures"; sub d ($x = 1) {$x;}`},
		{`use feature "signatures"; sub o ($x //= 5) { $x }`, `use feature "signatures"; sub o ($x //= 5) {$x;}`},
	} {
		root := parse.Parse([]byte(tc.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", tc.src, shape(root))
			continue
		}
		got := strings.TrimSpace(parse.Canon(root, []byte(tc.src)))
		if got != tc.want {
			t.Errorf("%q:\n  got  %q\n  want %q", tc.src, got, tc.want)
		}
	}
}
