// ABOUTME: `WORD BLOCK ARG` is one statement, not a call subscripted by an anonymous hash.
// ABOUTME: The shape `defer { ... } print "b";` needs, and which perl never reads as an Index.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestWordBlockArgIsAStatement: a brace group after an unknown bareword opens
// a BLOCK, and what follows it belongs to the same statement.
//
// Perl gives this shape three readings and which one applies is a symbol
// table question. Measured on 5.42.0 with -MO=Deparse:
//
//	zzz {a};                   zzz { 'a' }      undeclared: indirect method
//	sub zzz {} zzz {a};        zzz({'a'})       declared: call, anon hash arg
//	sub zzz(&) {} zzz { 1 };   &zzz(sub { 1; }) prototyped: call, code ref
//
// All three agree on the STRUCTURE -- the brace group and what follows belong
// to one statement -- and none of them is a call SUBSCRIPTED by a hash. That
// last is what this parser produced: `zzz { 1 } print "b"` read `zzz(){1}`
// and then had a `print` with no operator before it, so the statement fell to
// Unknown.
//
// Which of the three readings is right is M4's, once prototypes resolve. The
// statement BOUNDARY is M1's, and it is the same in all three.
func TestWordBlockArgIsAStatement(t *testing.T) {
	for _, src := range []string{
		`zzz { 1 } print "b";`,
		`defer { 1 };`,
		`defer { 1 } print "b";`,
		`sub f { defer { print "D\n" } print "body\n" }`,
		`use feature 'defer'; sub f { defer { print "D\n" } print "body\n" }`,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}

// TestWordBlockIsNotAnIndexCall: the brace group is a Block, so canon emits it
// as one.
//
// An Unknown count cannot see this failure. `my $x = defer { 1 } print "b";`
// has ZERO Unknowns today and canons to `defer(){1}` -- the index call perl
// never produces. Zero-Unknown-and-wrong is the more expensive failure of the
// two, because nothing counts it.
func TestWordBlockIsNotAnIndexCall(t *testing.T) {
	for _, src := range []string{
		`zzz { 1 };`,
		`defer { 1 };`,
		`my $x = defer { 1 } print "b";`,
	} {
		root := parse.Parse([]byte(src))
		got := parse.Canon(root, []byte(src))
		// `WORD(){...}` is the index-call spelling: an empty argument list
		// followed by a subscript. A block-taking word emits `WORD {...}`.
		if strings.Contains(got, `(){`) {
			t.Errorf("%q canons as an index call: %q", src, got)
		}
	}
}
