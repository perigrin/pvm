// ABOUTME: `given`, `when` and `default` are keyword statement forms, not barewords with braces.
// ABOUTME: Deprecated in modern perl, still compiled by it, and still exercised by perl's own op/switch.t.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestGivenWhenDefault: the switch statement forms parse, and parse as
// statements rather than as calls subscripted by an anonymous hash.
//
// The Unknown count alone cannot see this failure, which is why the shape is
// asserted too. Measured at 2dc301fd, before this landed:
//
//	given (1) { print "a" }                     Unknown=0, canon `given(1){print("a")};`
//	when (1) { print "a" } print "b";           Unknown=1
//
// The first has ZERO Unknowns and is still wrong: the tree is
// Index(Call(given, 1), Call(print)) -- a call SUBSCRIPTED by a hash, the one
// reading perl never produces. It is the same zero-Unknown-and-wrong defect
// `TestWordBlockIsNotAnIndexCall` was written for, reached through the paren
// cliff instead of the parenless path.
//
// Every spelling here was compiled on 5.42.0 before being asserted:
//
//	perl -e 'use feature "switch"; no warnings; given (1) { when (1) {...} }'
//	  -> one
func TestGivenWhenDefault(t *testing.T) {
	for _, src := range []string{
		// The three forms alone. `default` takes NO parens at all.
		`given ($x) { 1 }`,
		`when (1) { 1 }`,
		`default { 1 }`,

		// Two clauses inside one given, which is the shape op/switch.t has
		// and the one the issue measured at Unknown=2.
		`given ($x) { when (1) { print "a" } default { print "b" } }`,
		`given ($x) { when (1) { print "a" } when (2) { print "b" } }`,
		`use feature 'switch'; my $x = 1; given ($x) { when (1) { print "one\n" } default { print "other\n" } }`,

		// A clause followed by an ordinary statement. This is the case that
		// refused: the block is complete at its `}` and takes no `;`, so
		// what follows starts a new statement.
		`when (1) { print "a" } print "b";`,
		`default { print "a" } print "b";`,
		`given ($x) { when (1) { print "a" }; default { print "b" } }`,

		// `when`'s argument is an ordinary expression, not a shape of its
		// own. All four compile on 5.42.0; the smartmatch that interprets
		// them is runtime semantics, not syntax.
		`given ($x) { when ([1,2,3]) { 1 } }`,
		`given ($x) { when (/24/) { 1 } }`,
		`given ($x) { when (@list[0..2]) { 1 } }`,
		`given ($x) { when ($a && $b) { 1 } }`,

		// given in EXPRESSION position, which op/switch.t contains.
		`my @l = do { given ($_) { when (1) { "a" } } };`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
			continue
		}
		if firstOfKind(root, parse.Conditional) == nil {
			t.Errorf("%q must produce a Conditional: %v", src, kinds(root))
		}
	}
}

// TestSwitchIsNotAnIndexCall: the brace group is a Block, so canon emits it as
// one and never as `WORD(){...}`.
//
// `given(1){print("a")}` round-trips and has no Unknown, and is still a parse
// of a program perl does not have. Only the emitted shape shows it.
func TestSwitchIsNotAnIndexCall(t *testing.T) {
	for _, src := range []string{
		`given ($x) { 1 }`,
		`when (1) { 1 }`,
		`default { 1 }`,
		`given ($x) { when (1) { print "a" } default { print "b" } }`,
	} {
		root := parse.Parse([]byte(src))
		got := parse.Canon(root, []byte(src))
		if strings.Contains(got, `(){`) {
			t.Errorf("%q canons as an index call: %q", src, got)
		}
	}
}

// TestSwitchKeywordsAreNotCalls: `given` and `when` are keywords in this
// position only. Perl rejects the same shape on an arbitrary word --
// `zzz (1) { print "a" }` is a syntax error on 5.42.0, declared or not -- so
// the reading must be gated on the keyword rather than added to the general
// `WORD BLOCK` machinery, which would claim a form perl refuses.
//
//	$ perl -MO=Deparse -e 'zzz (1) { print "a" } print "b";'
//	syntax error at -e line 1, near ") {"
//	$ perl -MO=Deparse -e 'sub zzz {} zzz (1) { print "a" } print "b";'
//	syntax error at -e line 1, near ") {"
func TestSwitchKeywordsAreNotCalls(t *testing.T) {
	// An arbitrary word with parens and a block stays what it was: this
	// parser does not newly CLAIM the form, and the assertion is that the
	// switch fix did not widen to it.
	root := parse.Parse([]byte(`given ($x) { 1 }`))
	c := firstOfKind(root, parse.Conditional)
	if c == nil {
		t.Fatalf("given must be a Conditional: %v", kinds(root))
	}
	if c.Text != "given" {
		t.Errorf("Conditional Text = %q, want \"given\"", c.Text)
	}
}
