// ABOUTME: A declarator followed by a subscript keeps the subscript in the same statement.
// ABOUTME: Losing it produced a tree that was not a parse of its source, and it round-tripped.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestDeclaredSubscriptIsOneStatement: `local $SIG{__WARN__} = ...` is ONE
// statement holding one Index.
//
// It produced two: a Declaration spanning only `local $SIG`, then a separate
// statement whose `{__WARN__}` had become an AnonHash. `decl.go:34-37`
// already describes this failure mode for `sub {` -- the declarator stops
// mid-expression and the remainder is re-read in statement position, where a
// brace means something else entirely.
//
// Round-trip does not catch it: same leaves, same order, every byte back.
// Canonical re-emission does, which is where this was found -- the emission
// re-parses to a different tree. That disjointness is chapter 7 §7.2(c)'s
// claim and this is a measured instance of it.
func TestDeclaredSubscriptIsOneStatement(t *testing.T) {
	for _, src := range []string{
		`local $SIG{__WARN__} = sub { 1 };`,
		`local $h{k} = 1;`,
		`local $a[0] = 1;`,
		`my $h{k} = 1;`,
		`our $h{k} = 1;`,
		`state $h{k} = 1;`,
		`local $h{a}{b} = 1;`,

		// ANY assignment operator, not just `=`. Attaching the subscript
		// made this visible: with the target complete, `.= $x` fell out as
		// an Unknown beside it. Valid perl:
		//
		//	$ perl -e 'local $ENV{PATH} .= "x"; print "ok\n";'
		//	ok
		`local $ENV{PATH} .= $x;`,
		`local $h{k} ||= 1;`,
	} {
		root := parse.Parse([]byte(src))

		if n := topStatements(root); n != 1 {
			t.Errorf("%s\n  %d top-level statements, want 1 -- the subscript "+
				"was split into a statement of its own", src, n)
			continue
		}

		// The declaration's target must be an Index, not a bare Term with
		// the subscript lost.
		decl := root.Children[0].Children[0]
		if decl.Kind != parse.Declaration {
			t.Errorf("%s\n  first child is %v, want a Declaration", src, decl.Kind)
			continue
		}
		if !hasKind(decl, parse.Index) {
			t.Errorf("%s\n  the declaration holds no Index; the subscript is gone:\n%s",
				src, dumpTree(decl, 2))
		}
		if hasKind(decl, parse.AnonHash) {
			t.Errorf("%s\n  the subscript was read as an AnonHash:\n%s",
				src, dumpTree(decl, 2))
		}
	}
}
