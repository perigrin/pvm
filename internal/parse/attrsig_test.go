// ABOUTME: Under the signatures feature a sub's attributes come before its signature,
// ABOUTME: and the declaration reads both before the body.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestAttributesBeforeSignature: perl requires the attribute list BEFORE a
// signature -- after it is "Subroutine attributes must come before the
// signature" -- and parseSubDecl read the parens first, the prototype
// order, so the declaration ended at the attributes and the signature and
// body were left over. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'use feature "signatures";
//	      sub t106 :prototype(@) ($a) { $a } sub g :lvalue ($x, $) { $x }'
//	sub t106 : prototype(@) ($a) {
//	sub g : lvalue ($x, $) {
//
// perl.git t/op/signatures.t:1174 and 1186.
func TestAttributesBeforeSignature(t *testing.T) {
	for _, sub := range []string{
		`sub t105 :prototype($) ($a) { $a || "z" }`,
		`sub t106 :prototype(@) ($a) { $a || "z" }`,
		`sub g :lvalue ($x, $) { $x }`,
		// An anonymous sub's attribute list starts right after `sub`, and
		// its signature may touch the keyword or not. perf/opcount.t:1182
		// has `sub ($, $y)`.
		`my $f = sub ($, $y) { $y };`,
		`my $f = sub($, $y) { $y };`,
	} {
		src := "use feature 'signatures';\n" + sub + "\nmy $after = 1;\n"
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", sub, shape(root))
			continue
		}
		canon := parse.Canon(root, []byte(src))
		if !strings.Contains(canon, "{$") || !strings.Contains(canon, "my $after = 1;") {
			t.Errorf("%q: the body or the next statement was lost; canon %q", sub, canon)
		}
		again := parse.Canon(parse.Parse([]byte(canon)), []byte(canon))
		if again != canon {
			t.Errorf("%q: canon %q is not a fixpoint, reparses to %q", sub, canon, again)
		}
	}
}
