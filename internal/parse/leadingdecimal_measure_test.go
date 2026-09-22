// ABOUTME: Measures how many Unknowns in base/num.t and base/lex.t are leading-decimal literals.
// ABOUTME: Issue 01a0c13f asks for the number BEFORE the fix, so the delta after is measured, not asserted.
package parse_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// reLeadingDecimal matches a numeric literal written with no digit before
// the point, as it appears in SOURCE.
//
// The leading character class is what makes it a LEADING decimal rather
// than the tail of `1.5` or a method arrow: a `.` after a digit, letter,
// underscore or another `.` is something else. It is deliberately loose
// about what precedes otherwise -- `(`, `,`, `=`, a space after an
// operator -- because the classification here only has to be good enough
// to bucket Unknown spans, and every candidate is printed for reading.
var reLeadingDecimal = regexp.MustCompile(`(^|[^0-9A-Za-z_.])\.[0-9]`)

// TestMeasureLeadingDecimalUnknowns reports, per file, how many Unknown
// nodes have a leading decimal inside their span.
//
// It ASSERTS NOTHING about the count. The issue's first step is to
// record the number before any fix, and a test that pinned it would fail
// the moment the fix landed -- which is the point of the fix. So this
// prints and always passes; `TestLeadingDecimalLiteral` is the one that
// pins behaviour.
//
// An Unknown's span is the evidence rather than the line: a leading
// decimal that the lexer reads as an operator swallows a region, and the
// region is what the parser reports. Counting source occurrences instead
// would report every `.5` in the file including those in comments and
// strings, which are not parse failures at all.
func TestMeasureLeadingDecimalUnknowns(t *testing.T) {
	root := os.Getenv("PERL5_CORPUS")
	if root == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Skipf("no PERL5_CORPUS and no home directory: %v", err)
		}
		root = filepath.Join(home, "dev", "perl5")
	}

	for _, name := range []string{"base/num.t", "base/lex.t"} {
		path := filepath.Join(root, "t", name)
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Skipf("perl5 corpus not present: %v", err)
		}
		src := string(raw)

		var total, withDecimal int
		for _, u := range collectUnknownNodes(parse.Parse(raw)) {
			total++
			if u.Start < 0 || u.End > len(src) || u.Start >= u.End {
				continue
			}
			if reLeadingDecimal.MatchString(src[u.Start:u.End]) {
				withDecimal++
				span := strings.TrimSpace(src[u.Start:u.End])
				if len(span) > 60 {
					span = span[:60] + "..."
				}
				t.Logf("  %s: %q", name, span)
			}
		}
		t.Logf("%-12s %d Unknown(s), %d containing a leading decimal",
			name, total, withDecimal)
	}
}

// collectUnknownNodes gathers every Unknown in a tree.
//
// A local copy rather than a call to `collectUnknown` in
// refusal_test.go, because that one lives in package `parse` and this
// file is in `parse_test` -- the external test package, which is where a
// measurement belongs: it uses only the exported API, so it cannot
// accidentally depend on an internal the fix is about to change.
func collectUnknownNodes(n *parse.Node) []*parse.Node {
	if n == nil {
		return nil
	}
	var out []*parse.Node
	if n.Kind == parse.Unknown {
		out = append(out, n)
	}
	for _, c := range n.Children {
		out = append(out, collectUnknownNodes(c)...)
	}
	return out
}
