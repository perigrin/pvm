// ABOUTME: A signature element may take an empty default, `($=)` or `($x=)`: an
// ABOUTME: optional parameter with no default expression.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestSignatureEmptyDefault: perly.y's sigscalarelem takes an
// optsigscalardefault after the `=`, and it may be empty. Read as an
// ordinary list, `$ =` was an assignment with no right operand and refused.
// Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'use feature "signatures"; sub t027 ($ =) { 1 }
//	      sub t119 ($ =, $a = 333) { 1 }'
//	sub t027 ($=) {
//	sub t119 ($=, $a = 333) {
//
// perl.git t/op/signatures.t:493 and 506. Only a PLACEHOLDER may: a named
// parameter with nothing after its `=` is "Optional parameter lacks default
// expression" on 5.42.0, and outside a signature an assignment with nothing
// on its right refuses as it always did.
func TestSignatureEmptyDefault(t *testing.T) {
	for _, sub := range []string{
		`sub t027 ($ =) { $a // "z" }`,
		`sub t119 ($ =, $a = 333) { $a }`,
		`sub d ($y, $ //=) { $y }`,
	} {
		src := "use feature 'signatures';\n" + sub + "\nmy $after = 1;\n"
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", sub, shape(root))
			continue
		}
		canon := parse.Canon(root, []byte(src))
		if !strings.Contains(canon, "my $after = 1;") {
			t.Errorf("%q: the next statement was lost; canon %q", sub, canon)
		}
		again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon))
		if again != canon {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", sub, canon, again)
		}
	}
	for _, src := range []string{
		"use feature 'signatures';\nsub c ($x=) { $x }\n",
		"my @a = ($x =);\n",
	} {
		if !containsKind(parse.Parse([]byte(src)), parse.Unknown) {
			t.Errorf("%q: perl rejects this; it must refuse", src)
		}
	}
}
