// ABOUTME: The parser's fuzz target: never panic, never lose a byte, never build an empty node.
// ABOUTME: Forward progress is structural rather than fuzzed — go test -fuzz hangs on a loop instead of failing.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// fuzzSeeds are real Perl plus the shapes that stall a statement loop. Real
// Perl mutates into interesting near-Perl; random bytes mostly do not.
var fuzzSeeds = []string{
	"my $x = 42;",
	"sub f { 1 }\nf();",
	"if ($x) { print 1 } else { print 2 }",
	"my %h = (a => 1);",
	"print <<'EOF';\nbody\nEOF\n",
	"s{a}{b}ge;",
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
	for _, src := range []string{
		"",
		"\x00\x01\x02",
		strings.Repeat("{", 100),
		strings.Repeat("}", 100),
		strings.Repeat(";", 100),
		"'",
		"\\",
	} {
		checkParseInvariants(t, src)
	}
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
		if n.Kind != parse.SourceFile && n.End == n.Start {
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
}
