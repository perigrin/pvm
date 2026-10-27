// ABOUTME: The parser's fuzz target: never panic, never lose a byte, never build an empty node, canon a fixpoint.
// ABOUTME: Forward progress is structural rather than fuzzed — go test -fuzz hangs on a loop instead of failing.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// fuzzPerlSeeds are the real Perl in the corpus: each parses with no Unknown,
// so each is one the canon fixpoint is asserted on.
var fuzzPerlSeeds = []string{
	"my $x = 42;",
	"sub f { 1 }\nf();",
	"if ($x) { print 1 } else { print 2 }",
	"my %h = (a => 1);",
	"print <<'EOF';\nbody\nEOF\n",
	"s{a}{b}ge;",
}

// fuzzSeeds are real Perl plus the shapes that stall a statement loop. Real
// Perl mutates into interesting near-Perl; random bytes mostly do not.
var fuzzSeeds = append(append([]string{}, fuzzPerlSeeds...),
	// The shapes that stall a naive loop: a closer with nothing open, an
	// opener that never closes, and a statement with no semicolon.
	"}",
	"{",
	"}{",
	"sub f {",
	"my $x = (1",
	// Not Perl at all. An LSP sees half-typed buffers constantly.
	"=pod\ntext\n=cut\n",
	"__END__\nnot perl ) ( $$$\n",
	"format STDOUT =\n@<<<<<\n$a\n.\n",
)

// fuzzEdgeSeeds are the degenerate inputs the ordinary suite checks beside
// the corpus: nothing, control bytes, and long runs of one delimiter.
var fuzzEdgeSeeds = []string{
	"",
	"\x00\x01\x02",
	strings.Repeat("{", 100),
	strings.Repeat("}", 100),
	strings.Repeat(";", 100),
	"'",
	"\\",
	// A word's star before a comment: FuzzParser's canon fixpoint finds.
	"fo*#",
	"A*#",
	"A::B*#",
	"x*#",
}

// FuzzParser is the target. Run it with
//
//	go test ./internal/parse/ -run '^$' -fuzz '^FuzzParser$' -fuzztime 1000000x
//
// or `make fuzz-parser`. `-run` alone does NOT fuzz -- it executes the seed
// corpus only -- which is why the acceptance criterion names the make target.
func FuzzParser(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		checkParseInvariants(t, src)
	})
}

// TestFuzzSeeds runs the invariants over the seed corpus in the ordinary test
// suite, so they are checked on every `go test` rather than only when someone
// remembers to fuzz. Crashers land in testdata/fuzz/ and are replayed here.
func TestFuzzSeeds(t *testing.T) {
	for _, src := range fuzzSeeds {
		checkParseInvariants(t, src)
	}
	for _, src := range fuzzEdgeSeeds {
		checkParseInvariants(t, src)
	}
}

// TestFuzzSeedsCanonFixpoint: canon is a fixpoint on every seed that parses
// with no Unknown, and the real-Perl seeds are among those it is asserted on.
//
// Of the 21 seeds, 19 are fixpoints and 2 are not, `}{` and `\`. Both parse
// to an Unknown, and both settle one emission later. An Unknown is the
// parser's refusal, which canon writes back verbatim inside structure the
// parser guessed around it, so the second parse can read the refused bytes
// beside punctuation the first did not have: `\` emits `\\;` because the
// lone operator's Unknown operand spans the operator's own byte, and `}{`
// emits `}{;`, whose `;` the second parse takes as another refused operand
// of the unclosed subscript. The 15 seeds with no Unknown are all fixpoints,
// and they are the ones the gate admits.
//
// The gate does not exempt the class the invariant is for. 01a0dde4's source,
// `map { $_ + 1 } @a`, parsed with no Unknown, and its emission was not a
// fixpoint.
func TestFuzzSeedsCanonFixpoint(t *testing.T) {
	admitted := map[string]bool{}
	for _, src := range append(append([]string{}, fuzzSeeds...), fuzzEdgeSeeds...) {
		admitted[src] = checkCanonFixpoint(t, src)
	}
	for _, src := range fuzzPerlSeeds {
		if !admitted[src] {
			t.Errorf("%q: parses to an Unknown, so the fixpoint is not asserted on it", src)
		}
	}
}

// checkCanonFixpoint asserts canonMatches on src when src parses with no
// Unknown, Error or Missing node, and reports whether it did. Fixpoint only,
// not that the emission parses: a refused emission can still be stable, and
// the targeted canon tests carry the parse check.
func checkCanonFixpoint(t *testing.T, src string) bool {
	t.Helper()
	if root := parse.Parse([]byte(src)); firstUnknown(root) != nil || root.HasError {
		return false
	}
	if ok, why := canonMatches([]byte(src)); !ok {
		t.Fatalf("%q: canon is not a fixpoint: %s", src, why)
	}
	return true
}

// checkParseInvariants is the tree-level form of the lexer's §7.6.2
// invariants. Forward progress is enforced structurally inside Parse, because
// `go test -fuzz` does not detect an infinite loop -- it hangs, reporting
// nothing at all. That is how the lexer's `format =` spin was found only by
// the fuzzer's own timeout rather than by a failure.
func checkParseInvariants(t *testing.T, src string) {
	t.Helper()

	// INVARIANT 1: never panic. The call itself is the assertion.
	root := parse.Parse([]byte(src))

	if root == nil {
		t.Fatalf("%q: Parse returned nil", src)
	}

	// INVARIANT 2: spans are in bounds, non-empty below the root, and cover
	// their children exactly.
	var check func(*parse.Node)
	check = func(n *parse.Node) {
		if n.Start < 0 || n.End > len(src) || n.End < n.Start {
			t.Fatalf("%q: node %v spans [%d,%d), out of bounds for %d bytes",
				src, n.Kind, n.Start, n.End, len(src))
		}
		if n.Kind != parse.SourceFile && n.Kind != parse.Missing && n.End == n.Start {
			t.Fatalf("%q: node %v is empty; the loop that made it did not advance",
				src, n.Kind)
		}
		prev := n.Start
		for _, c := range n.Children {
			if c.Start < prev {
				t.Fatalf("%q: child of %v starts at %d, before %d",
					src, n.Kind, c.Start, prev)
			}
			prev = c.End
			check(c)
		}
	}
	check(root)

	// INVARIANT 3: no byte is lost. This is the one that matters most --
	// a tree that drops source is a tree an LSP cannot render from.
	if got := leafText(root, []byte(src)); got != src {
		t.Fatalf("%q: round-trip lost bytes, got %q", src, got)
	}

	// INVARIANT 4: canon is a fixpoint on input that parses with no Unknown.
	// Gated because an Unknown is written back verbatim inside guessed
	// structure and need not re-read the same; TestFuzzSeedsCanonFixpoint
	// names the seeds that are not fixpoints for that reason.
	checkCanonFixpoint(t, src)
}
