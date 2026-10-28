// ABOUTME: Holds R-a′ and R-g on Perl that perl compiles: catch-all nodes publish nothing, and no missing_operand fires.
// ABOUTME: Reads compiled.txt (perl.git t/) and testdata/t1compiled.txt (T1); never runs perl. Skips without both corpora.

package psc

import (
	"bufio"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// compiledList reads a list of corpus-relative paths, one per line, with #
// comments. An empty list is an error, never a corpus of 0 files.
func compiledList(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); line != "" && !strings.HasPrefix(line, "#") {
			out = append(out, line)
		}
	}
	if err := sc.Err(); err != nil || len(out) == 0 {
		t.Fatalf("%s: %d paths, %v", path, len(out), err)
	}
	return out
}

// catchAllSpans is the span of every catch-all node in n's tree.
func catchAllSpans(n *parse.Node) []span {
	var out []span
	if n.Refused() && parse.CatchAll(n.Refusal) {
		out = append(out, span{uint32(n.Start), uint32(n.End)})
	}
	for _, c := range n.Children {
		out = append(out, catchAllSpans(c)...)
	}
	return out
}

// isSyntaxCode reports whether code is a syntax diagnostic's: a refusal
// code or a missing token's.
func isSyntaxCode(code string) bool {
	_, ok := parse.RefusalSites[parse.RefusalCode(code)]
	return ok || code == CodeMissingToken
}

// TestParseDiagnosticsCatchAllQuietOnCorpus: on every perl-compilable corpus
// file, through both of OpenDocument's parse paths (a file: URI with its
// incRoots, and an untitled: URI through parse.Parse), the catch-all nodes
// are present yet publish no syntax diagnostic, and no missing_operand
// diagnostic fires (R-a′, R-g). Any other syntax diagnostic is logged by
// file, since perl compiles every one of these files.
func TestParseDiagnosticsCatchAllQuietOnCorpus(t *testing.T) {
	t1, perl5 := os.Getenv("T1_CORPUS"), os.Getenv("PERL5_CORPUS")
	if t1 == "" || perl5 == "" {
		t.Skip("set T1_CORPUS and PERL5_CORPUS")
	}
	for _, c := range []struct{ name, root, list string }{
		{"perl5", filepath.Join(perl5, "t"), filepath.Join("..", "parseoracle", "testdata", "ratchet", "compiled.txt")},
		{"t1", t1, filepath.Join("testdata", "t1compiled.txt")},
	} {
		files := compiledList(t, c.list)
		s := NewLSPServer()
		var catchAll, syntaxDiags, missingOp, other int
		for _, rel := range files {
			path := filepath.Join(c.root, rel)
			src, err := os.ReadFile(path)
			if err != nil {
				t.Errorf("%s: %v", c.name, err)
				continue
			}
			roots, _ := incRoots(path)
			for _, p := range []struct {
				uri  string
				tree *parse.Node
			}{
				{FileURI(path), parse.NewSession().ParseSourceFrom(path, src, roots...)},
				{"untitled:" + rel, parse.Parse(src)},
			} {
				spans := catchAllSpans(p.tree)
				catchAll += len(spans)
				if err := s.OpenDocument(p.uri, src); err != nil {
					t.Fatal(err)
				}
				for _, d := range s.Diagnostics(p.uri) {
					if !isSyntaxCode(d.Code) {
						continue
					}
					switch {
					case parse.CatchAll(parse.RefusalCode(d.Code)) || slices.ContainsFunc(spans, func(g span) bool { return g.overlaps(d) }):
						syntaxDiags++
						t.Errorf("%s %s: catch-all published %s at %d", c.name, p.uri, d.Code, d.StartByte)
					case d.Code == string(parse.MissingOperand):
						missingOp++
						t.Errorf("%s %s: missing_operand at %d on Perl perl compiles", c.name, p.uri, d.StartByte)
					default:
						other++
						t.Logf("%s %s: %s at %d: %s", c.name, p.uri, d.Code, d.StartByte, d.Message)
					}
				}
				s.CloseDocument(p.uri)
			}
		}
		t.Logf("corpus=%s files=%d catchall=%d syntaxdiags=%d missingop=%d other=%d", c.name, len(files), catchAll, syntaxDiags, missingOp, other)
		if catchAll == 0 {
			t.Errorf("%s: no catch-all node seen; the quiet check exercised nothing", c.name)
		}
	}
}
