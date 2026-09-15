// ABOUTME: The M0 gate: t/base/lex.t tokenizes, round-trips, and matches a committed golden token stream.
// ABOUTME: Round-trip alone cannot gate this — it holds for a lexer that returns one token for the whole file.

package lexer

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// updateGolden rewrites the committed token stream. A flag rather than an
// environment variable so a `go test` invocation records in its own command
// line that it meant to, and so a plain run can never do it by accident --
// the same discipline the parse-fidelity ratchet uses.
var updateGolden = flag.Bool("lexer.update", false,
	"rewrite the committed golden token stream for t/base/lex.t")

// lexDotT returns the contents of t/base/lex.t, or skips.
//
// $PERL5_CORPUS or ~/dev/perl5, matching parseoracle.CorpusRoot so the two
// halves of this repository look for the corpus in the same place. The skip
// matters: ci.yml runs `go test ./...` with no corpus checkout, and only
// parse-fidelity.yml clones perl5.
func lexDotT(t *testing.T) (string, []byte) {
	t.Helper()
	root := os.Getenv("PERL5_CORPUS")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skipf("no PERL5_CORPUS and no home directory: %v", err)
		}
		root = filepath.Join(home, "dev", "perl5")
	}
	path := filepath.Join(root, "t", "base", "lex.t")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("t/base/lex.t not readable at %s: %v\n"+
			"set PERL5_CORPUS to a perl5 checkout to run the M0 gate", path, err)
	}
	return path, src
}

// TestLexDotTNoErrorTokens: the file lexes cleanly.
//
// Spec §7.3.2: "Treat lex.t as a milestone in its own right, not as one file
// among 620. If it parses, the lexer is real." 15,633 bytes, 129 tests,
// written specifically to abuse a lexer.
func TestLexDotTNoErrorTokens(t *testing.T) {
	path, src := lexDotT(t)
	var bad []string
	for _, tok := range Tokenize(src) {
		if tok.Kind == Error || tok.Kind == UnknownRest {
			bad = append(bad, fmt.Sprintf("%s at line %d: %q",
				tok.Kind, lineOf(src, tok.Start), excerpt(src, tok)))
			if len(bad) == 10 {
				bad = append(bad, "...")
				break
			}
		}
	}
	if len(bad) > 0 {
		t.Errorf("%s does not lex cleanly:\n  %s", path, strings.Join(bad, "\n  "))
	}
}

// TestLexDotTRoundTrips is invariant 3 on the hardest file in the corpus.
//
// Necessary and NOT sufficient: the property holds for any partition of the
// bytes, including one token spanning the file. The golden stream below is
// what actually gates delimitation.
func TestLexDotTRoundTrips(t *testing.T) {
	path, src := lexDotT(t)
	var sb strings.Builder
	for _, tok := range Tokenize(src) {
		sb.Write(src[tok.Start:tok.End])
	}
	if sb.String() != string(src) {
		t.Errorf("%s does not round-trip: %d bytes in, %d out", path, len(src), sb.Len())
	}
}

// TestLexDotTGoldenStream is the real gate.
//
// Round-trip cannot fail on misdelimitation -- a lexer that returns
//
//	[]Token{{Kind: Whitespace, Start: 0, End: len(src)}}
//
// satisfies it, and so does every near-miss that matters: `=cute` ending POD
// early, the /e heredoc body lexed as barewords, `$main'a` opening a string.
// All of them reproduce the input exactly and emit no error.
//
// So the gate is the token stream itself: kind and span for every token,
// committed, compared exactly. That is what makes §7.3.2's "if it parses, the
// lexer is real" mean anything at the round-trip bar.
func TestLexDotTGoldenStream(t *testing.T) {
	_, src := lexDotT(t)
	got := renderTokens(src, Tokenize(src))

	golden := filepath.Join("testdata", "lex.t.tokens")
	if *updateGolden {
		if err := os.MkdirAll(filepath.Dir(golden), 0o755); err != nil {
			t.Fatalf("creating testdata: %v", err)
		}
		if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
			t.Fatalf("writing %s: %v", golden, err)
		}
		t.Logf("rewrote %s; commit it in the same commit as the change that moved it", golden)
		return
	}

	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("reading %s: %v\nrun with -lexer.update to create it", golden, err)
	}
	if got != string(want) {
		t.Errorf("the token stream moved. First difference:\n%s\n\n"+
			"If this is a deliberate improvement, re-run with -lexer.update and "+
			"commit the golden file in the same commit.", firstDiff(string(want), got))
	}
}

// TestGoldenRequiresUpdateFlag: a golden file that rewrites itself on
// mismatch is decoration. Assert the flag defaults off, so a plain `go test`
// can only compare.
func TestGoldenRequiresUpdateFlag(t *testing.T) {
	f := flag.Lookup("lexer.update")
	if f == nil {
		t.Fatal("no -lexer.update flag: the golden file has no deliberate update path")
	}
	if f.DefValue != "false" {
		t.Errorf("-lexer.update defaults to %q; a golden file that rewrites itself "+
			"on mismatch records nothing", f.DefValue)
	}
}

// TestLexDotTTraps pins each construct §7.3.2 names, at the line it occupies,
// so a failure says WHICH trap broke rather than only that the stream moved.
//
// Line numbers verified against the pinned corpus revision.
func TestLexDotTTraps(t *testing.T) {
	_, src := lexDotT(t)
	toks := Tokenize(src)

	for _, tc := range []struct {
		name string
		line int
		kind Kind
	}{
		// $x = $#[0]; -- `$#` names the array @#, and `[0]` subscripts it.
		{"array named hash", 10, Variable},
		// $x = '\\'; # '; -- a quoted backslash, then a comment holding a quote.
		{"quoted backslash", 12, Quote},
		// eval '$foo{1} / 1;'; -- division after a hash subscript, inside a string.
		{"division in string", 20, Quote},
		// print <<'EOF'; -- non-interpolating heredoc.
		{"single quoted heredoc", 36, HeredocOpen},
		// print <<EOF; -- interpolating.
		{"bare heredoc", 41, HeredocOpen},
		// eval <<\EOE, print $@; -- backslash-quoted, mid-expression.
		{"backslash heredoc", 45, HeredocOpen},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var found bool
			for _, tok := range toks {
				if lineOf(src, tok.Start) == tc.line && tok.Kind == tc.kind {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("no %v token on line %d: %q", tc.kind, tc.line,
					lineAt(src, tc.line))
			}
		})
	}
}

// TestLexDotTSkipsWithoutCorpus: the gate must skip cleanly where the corpus
// is absent, because ci.yml runs `go test ./...` without one. A test that
// FAILS there would make every CI run red for an environmental reason, which
// is the defect class this project keeps finding.
func TestLexDotTSkipsWithoutCorpus(t *testing.T) {
	t.Setenv("PERL5_CORPUS", filepath.Join(t.TempDir(), "absent"))
	// Run the resolver in a subtest so its Skip does not stop this one.
	var skipped bool
	t.Run("resolve", func(st *testing.T) {
		defer func() { skipped = st.Skipped() }()
		lexDotT(st)
	})
	if !skipped {
		t.Error("the gate did not skip with the corpus absent; it must skip " +
			"rather than fail, or a missing checkout reads as a lexer defect")
	}
}

// renderTokens formats a token stream as one line per token: kind, span, and
// the source text with newlines escaped so a diff stays line-oriented.
func renderTokens(src []byte, toks []Token) string {
	var sb strings.Builder
	for _, tok := range toks {
		fmt.Fprintf(&sb, "%-12s %6d %6d  %s\n",
			tok.Kind, tok.Start, tok.End, escapeSpan(src[tok.Start:tok.End]))
	}
	return sb.String()
}

// escapeSpan renders a token's text on one line, truncated. The text is here
// so a golden diff is readable without cross-referencing offsets.
//
// SPACES ARE ESCAPED, which looks like over-decoration and is not. A
// whitespace token's text IS a space, so an unescaped rendering ends the line
// with one -- and the repository's trailing-whitespace pre-commit hook then
// rewrites the golden file, silently, on the way into the commit. Measured:
// it stripped 1,059 of 3,463 lines and the gate failed on its own committed
// data.
//
// Escaping here rather than exempting the file from the hook: a golden file
// that needs a hook exemption to survive is a golden file someone will later
// regenerate wrongly.
func escapeSpan(b []byte) string {
	const max = 60
	var sb strings.Builder
	for i, c := range b {
		if i >= max {
			sb.WriteString("...")
			break
		}
		switch c {
		case '\n':
			sb.WriteString(`\n`)
		case '\t':
			sb.WriteString(`\t`)
		case '\r':
			sb.WriteString(`\r`)
		case ' ':
			sb.WriteString(`\s`)
		default:
			sb.WriteByte(c)
		}
	}
	return sb.String()
}

// firstDiff returns the first differing line of two rendered streams, with a
// little context. A 3,000-line diff is unreadable; the first difference is
// almost always the only one that matters.
func firstDiff(want, got string) string {
	w := strings.Split(want, "\n")
	g := strings.Split(got, "\n")
	for i := 0; i < len(w) || i < len(g); i++ {
		var wl, gl string
		if i < len(w) {
			wl = w[i]
		}
		if i < len(g) {
			gl = g[i]
		}
		if wl != gl {
			return fmt.Sprintf("  line %d\n  want: %s\n  got:  %s", i+1, wl, gl)
		}
	}
	return "  (streams are equal but compared unequal: a trailing-newline difference)"
}

// lineOf returns the 1-based line number of a byte offset.
func lineOf(src []byte, off int) int {
	n := 1
	for i := 0; i < off && i < len(src); i++ {
		if src[i] == '\n' {
			n++
		}
	}
	return n
}

// lineAt returns the text of a 1-based line.
func lineAt(src []byte, want int) string {
	n, start := 1, 0
	for i := 0; i < len(src); i++ {
		if n == want {
			end := i
			for end < len(src) && src[end] != '\n' {
				end++
			}
			return string(src[start:end])
		}
		if src[i] == '\n' {
			n++
			start = i + 1
		}
	}
	return ""
}

// excerpt returns a short rendering of a token's source, for error messages.
func excerpt(src []byte, tok Token) string {
	end := tok.End
	if end > tok.Start+40 {
		end = tok.Start + 40
	}
	return escapeSpan(src[tok.Start:end])
}
