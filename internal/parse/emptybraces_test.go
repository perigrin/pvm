// ABOUTME: An empty `{}` at statement position is an anonymous hash, as toke.c's brace rule reads it.
// ABOUTME: A brace holding anything, even a lone `;`, stays a block.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestEmptyBracesAtStatementAreAnonHash: perl reads `{` then `}` at statement
// position, with only whitespace or comments between, as an anonymous hash.
// Measured on perl 5.42.0 with -MO=Concise, each of these builds
// `emptyavhv[t1] v/ANONHASH` and no enterloop:
//
//	{}   { }   { # c\n }   {\n}   {} ;   L: {}   sub f { 1; {} }
//
// and `{{{}}}` is two bare blocks (enterloop/leaveloop) around one emptyavhv.
// A brace with anything in it stays a block: `{ 1 }`, `{ ; }` and `{;}` are
// each an enterloop.
func TestEmptyBracesAtStatementAreAnonHash(t *testing.T) {
	// Each source's count of AnonHash; the first has `{a=>1}` as well.
	for src, want := range map[string]int{
		"my $h={a=>1};\n{}": 2,
		"{}":                1,
		"{ }":               1,
		"{ # c\n }":         1,
		"{\n}":              1,
		"{} ;":              1,
		"{};":               1,
		"L: {}":             1,
		"sub f { 1; {} }":   1,
		"if ($x) { {} }":    1,
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
		if got := len(findAll(root, parse.AnonHash)); got != want {
			t.Errorf("%q: perl builds an empty anon hash, want %d AnonHash, got %d: %v",
				src, want, got, kinds(root))
		}
	}

	// The innermost of a nest is the hash; the two around it are blocks.
	root := parse.Parse([]byte("{{{}}}"))
	if got := len(findAll(root, parse.AnonHash)); got != 1 {
		t.Errorf("{{{}}}: want 1 AnonHash, got %d: %v", got, kinds(root))
	}
	if got := len(findAll(root, parse.Block)); got != 2 {
		t.Errorf("{{{}}}: want 2 Blocks, got %d: %v", got, kinds(root))
	}

	// Anything inside the braces keeps them a block.
	for _, src := range []string{"{ 1 }", "{ ; }", "{;}"} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
		if containsKind(root, parse.AnonHash) {
			t.Errorf("%q is a block, not a hash: %v", src, kinds(root))
		}
		if !containsKind(root, parse.Block) {
			t.Errorf("%q must produce a Block: %v", src, kinds(root))
		}
	}
}
