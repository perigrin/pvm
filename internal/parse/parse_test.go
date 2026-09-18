// ABOUTME: The skeleton's properties: empty parse, Unknown spans, round-trip, forward progress.
// ABOUTME: Round-trip is the parser-level analogue of the lexer's invariant 3, asserted over real files.

package parse_test

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// unimplementedStatement is a statement form no issue has landed yet, used by
// every test that needs the parser to DECLINE something.
//
// One constant rather than a literal per test, because these fixtures keep
// outliving their purpose. It has been `$x = 1;`, then `my $r = \@a;`, then
// `if ($c) { 1 }` -- each was Unknown when written and parses now. That is
// progress, but it silently turns "this test proves declining works" into
// "this test proves nothing", and it is only caught because the assertions
// fail loudly rather than vacuously.
//
// When nothing is left to put here, these tests are measuring an empty set.
// DELETE them then rather than inventing a construct to keep them alive: the
// harness's own no-answer bucket covers the property at that point.
// `goto &other` was the fifth and parsed as of the keyword-table commit.
// `$::x = ...` is the sixth: a leading `::` is the main-package shorthand
// and `scanVarName` does not lex it (issue 01a0afc0, the scanner-rows half).
const unimplementedStatement = "$::x = \\@a;\n"

// unimplementedStatementWithRef is the same, but containing a backslash
// reference so perl reports an srefgen the parser must hedge against.
//
// `$::x = \@a` is still declined, and the `\@a` IS the reference perl
// reports. Measured on perl 5.42.0:
//
//	$ perl -MO=Concise,-exec -e '$::x = \@a;' | grep -c srefgen
//	1
//	$ perl -e 'our @a=(1); $::x = \@a; print ref($::x)'
//	ARRAY
//
// Valid Perl that this parser declines, which is what the fixture needs.
//
// This fixture has moved five times -- `$x = \@a`, `my $r = \@a`, an `if`
// block, a `do` block, `goto &other` -- each time because the parser learned
// the form. See unimplementedStatement for when to delete these tests rather
// than keep finding new fixtures.
const unimplementedStatementWithRef = "$::x = \\@a;\n"

// TestParseEmpty: an empty input is a valid program, not an error. The root
// exists and spans nothing.
func TestParseEmpty(t *testing.T) {
	root := parse.Parse(nil)
	if root == nil {
		t.Fatal("Parse(nil) returned nil; the root must always exist")
	}
	if root.Kind != parse.SourceFile {
		t.Errorf("root kind = %v, want SourceFile", root.Kind)
	}
	if len(root.Children) != 0 {
		t.Errorf("root has %d children, want 0", len(root.Children))
	}
	if root.Start != 0 || root.End != 0 {
		t.Errorf("root spans [%d,%d), want [0,0)", root.Start, root.End)
	}
}

// TestUnknownSpansItsSource: the Unknown covers exactly the bytes it could
// not parse -- no more, so the next statement is still reachable; no less,
// so round-trip holds.
func TestUnknownSpansItsSource(t *testing.T) {
	// A statement form the parser does not implement yet. See
	// unimplementedStatement for why this is a shared constant.
	src := []byte(unimplementedStatement)

	root := parse.Parse(src)
	unknowns := collect(root, parse.Unknown)
	if len(unknowns) != 1 {
		t.Fatalf("got %d Unknown nodes, want 1: %v", len(unknowns), kinds(root))
	}
	// Through the closing brace and no further: the trailing newline is
	// trivia, belonging to no statement. Round-trip still holds because the
	// trivia node carries it.
	u := unknowns[0]
	want := strings.TrimSuffix(unimplementedStatement, "\n")
	if got := string(src[u.Start:u.End]); got != want {
		t.Errorf("Unknown spans %q, want %q", got, want)
	}
}

func kinds(n *parse.Node) []string {
	out := []string{n.Kind.String()}
	for _, c := range n.Children {
		out = append(out, kinds(c)...)
	}
	return out
}

// TestTreeRoundTrips: concatenating every leaf reproduces the input. This is
// invariant 3 lifted to the tree, and it is what lets an LSP show correct
// text for a construct nobody has taught the parser yet.
func TestTreeRoundTrips(t *testing.T) {
	for _, src := range []string{
		"",
		"\n",
		"$x = 1;\n",
		"# just a comment\n",
		"sub f { 1 }\n\nf();\n",
		"my $s = \"unterminated\n",
		"\x00\xff garbage \x01\n",
	} {
		root := parse.Parse([]byte(src))
		if got := leafText(root, []byte(src)); got != src {
			t.Errorf("round-trip failed\n in: %q\nout: %q", src, got)
		}
	}
}

// TestTreeRoundTripsCorpus: the same property over every real file we have.
// A hand-picked list proves the cases I thought of; the corpus proves the
// ones I did not. The lexer found a 141-byte loss this way that no unit test
// had reached.
func TestTreeRoundTripsCorpus(t *testing.T) {
	files := corpusFiles(t)
	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		root := parse.Parse(src)
		if got := leafText(root, src); got != string(src) {
			t.Errorf("%s: round-trip lost %d bytes", path, len(src)-len(got))
		}
	}
	t.Logf("round-tripped %d files", len(files))
}

// TestParserForwardProgress: every parse step consumes at least one token.
// Asserted structurally rather than by timeout: a timeout test that passes
// tells you nothing about why, and a loop that makes no progress is a bug
// with a precise definition.
func TestParserForwardProgress(t *testing.T) {
	// Inputs chosen to stall a naive loop: a lone closer with nothing open,
	// a statement that never ends, and an empty block.
	for _, src := range []string{"}", "{", "}{", ";;;", "{}", "sub", "=", "\n}\n"} {
		root := parse.Parse([]byte(src))
		if got := leafText(root, []byte(src)); got != src {
			t.Errorf("%q: round-trip failed, output %q", src, got)
		}
		// Every node below the root must span at least one byte, or the loop
		// that made it did not advance.
		for _, n := range allNodes(root) {
			if n.Kind == parse.SourceFile {
				continue
			}
			if n.End <= n.Start {
				t.Errorf("%q: node %v spans [%d,%d), which is empty",
					src, n.Kind, n.Start, n.End)
			}
		}
	}
}

// --- helpers ---

func allNodes(n *parse.Node) []*parse.Node {
	out := []*parse.Node{n}
	for _, c := range n.Children {
		out = append(out, allNodes(c)...)
	}
	return out
}

func collect(n *parse.Node, k parse.Kind) []*parse.Node {
	var out []*parse.Node
	for _, m := range allNodes(n) {
		if m.Kind == k {
			out = append(out, m)
		}
	}
	return out
}

// leafText reconstructs the input from the tree.
//
// Delegates to Node.SourceText rather than re-implementing the walk: a test
// helper with its own traversal can agree with a broken tree, and this
// property is exactly the one that must not be asserted against itself.
func leafText(n *parse.Node, src []byte) string {
	return n.SourceText(src)
}

// corpusFiles walks the whole perl5 t/ directory, or skips.
//
// The whole tree rather than the 56 T2 files, matching the lexer's ratchet:
// round-trip is cheap to check and the wide sweep is what found a 141-byte
// loss in comp/parser.t that no unit test had reached. Coverage targets are
// scoped to T2; an invariant is not.
// corpusFilesIn returns the .t files in one corpus subdirectory.
func corpusFilesIn(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	want := string(filepath.Separator) + dir + string(filepath.Separator)
	for _, f := range corpusFiles(t) {
		if strings.Contains(f, want) {
			out = append(out, f)
		}
	}
	return out
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return src
}

func corpusFiles(t *testing.T) []string {
	t.Helper()
	root := os.Getenv("PERL5_CORPUS")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skipf("no PERL5_CORPUS and no home directory: %v", err)
		}
		root = filepath.Join(home, "dev", "perl5")
	}
	tDir := filepath.Join(root, "t")
	if _, err := os.Stat(tDir); err != nil {
		t.Skipf("perl5 corpus not present at %s: %v", tDir, err)
	}

	var files []string
	err := filepath.Walk(tDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".t") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Skipf("walking the corpus: %v", err)
	}
	sort.Strings(files)
	if len(files) == 0 {
		t.Skipf("no corpus files under %s", tDir)
	}
	return files
}
