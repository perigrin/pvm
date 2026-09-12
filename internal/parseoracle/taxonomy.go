// ABOUTME: Categorises a file's first parse error by the Perl construct that broke it.
// ABOUTME: A raw count says "203 files fail"; a taxonomy says which 48 are heredocs, which is a work plan.

package parseoracle

import (
	"os"
	"regexp"
	"strings"

	"tamarou.com/pvm/internal/parser"
)

// Categorise names the construct behind one result's first parse error.
//
// A file that reached a verdict without declining has no first error, so it
// gets CategoryNone rather than a guess. Only the files in the no-answer
// bucket — the coverage gap the conformance plan's M1 has to move — carry a
// real category.
//
// dir is RunOptions.Dir, since the runner records corpus paths relative to the
// shim's t/ and this has to re-read the file.
func Categorise(r Result, dir string) Category {
	if r.Err != "" || r.Excluded {
		return CategoryNone
	}
	if r.Verdict.Bucket != BucketNoAnswer {
		return CategoryNone
	}

	src, err := os.ReadFile(resolve(r.Path, dir))
	if err != nil {
		return CategoryGeneral
	}
	return CategoriseSource(src)
}

// CategoriseSource categorises the first parse error in src.
//
// The construct is read from the SOURCE TEXT at the point recovery gave up,
// not from the error node's own kind, and that is forced by how this grammar
// recovers. Measured across the corpus, tree-sitter's first ERROR node
// typically begins at the top of the file — its first child is the shebang
// comment — so the node kind says only "something in this file is wrong". The
// END of that span is where the parse actually broke, so the line containing
// it is the evidence.
//
// ponytail: line-level pattern match, not a second parser. Reimplementing
// perl's grammar to classify failures of our grammar would be writing the
// parser twice, and the answer only has to be good enough to sort a work
// queue. If a category ever needs to be exact, it needs a rule in the grammar,
// not a better regexp here.
//
// This function is deliberately NOT part of the portable contract, and its
// dependency on internal/parser is not the coupling that subject.go exists to
// remove. Categorisation asks "where did THIS parser give up", which only that
// parser can answer; the comparison asks "did perl and the subject agree",
// which any implementation can answer. A second subject that wants its
// no-answer files triaged supplies its own categoriser, exactly as it supplies
// its own facts. What matters is that a verdict never depends on this.
func CategoriseSource(src []byte) Category {
	tree, err := parser.New().Parse(src)
	if err != nil || tree == nil {
		return CategoryGeneral
	}
	root := tree.RootNode()
	if root == nil {
		return CategoryGeneral
	}

	// A degenerate tree dropped source without an error node, so there is no
	// span to read. The hidden rule that leaked is the only evidence, and it
	// names an expression position rather than a construct.
	site, ok := errorSite(root, src)
	if !ok {
		if tree.IsDegenerate() {
			return CategoryGeneral
		}
		return CategoryNone
	}
	return categoriseLine(site)
}

// errorSite returns the source line on which the first error span ends, which
// is where recovery gave up.
func errorSite(root *parser.Node, src []byte) (string, bool) {
	e := firstErrorNode(root)
	if e == nil {
		return "", false
	}

	end := int(e.EndByte())
	if end > len(src) {
		end = len(src)
	}
	// The line containing the end of the error span. Recovery consumes up to
	// the token it could not place, so that token is at or just before end.
	start := strings.LastIndexByte(string(src[:end]), '\n') + 1
	lineEnd := end
	if i := strings.IndexByte(string(src[end:]), '\n'); i >= 0 {
		lineEnd = end + i
	} else {
		lineEnd = len(src)
	}
	line := strings.TrimSpace(string(src[start:lineEnd]))
	if line == "" {
		// An error ending on a blank line: fall back to the whole span's
		// tail, which still carries the construct.
		tail := string(src[max(0, end-160):end])
		return strings.TrimSpace(tail), true
	}
	return line, true
}

// firstErrorNode finds the first ERROR node in document order, descending so
// the innermost — and therefore most specific — span wins.
func firstErrorNode(n *parser.Node) *parser.Node {
	if n == nil || !n.HasError() {
		return nil
	}
	for i := 0; i < n.ChildCount(); i++ {
		c := n.Child(i)
		if c != nil && c.HasError() {
			if e := firstErrorNode(c); e != nil {
				return e
			}
		}
	}
	if n.IsError() {
		return n
	}
	return nil
}

// The taxonomy rules, in priority order. The first match wins, so a line that
// touches two constructs is filed under the higher-priority one — which is the
// right bias for a work queue: ModernFeature is P1 and the rest are P2.
var taxonomyRules = []struct {
	category Category
	pattern  *regexp.Regexp
}{
	// P1: the post-5.36 surface. This is the Perl people are writing now,
	// so a file that fails here blocks current code rather than legacy.
	{CategoryModernFeature, regexp.MustCompile(
		`\b(class|field|method|ADJUST)\b|\buse\s+v5\.(3[6-9]|4[0-9])|` +
			`\b(try|catch|finally|defer)\s*\{|\bbuiltin::|\buse\s+feature\b`)},

	// P2: quote-like operators, whose bodies cannot be lexed without first
	// knowing the operator and its delimiter. The heredoc introducer is
	// listed first because it changes where the NEXT lines are read from.
	{CategoryQuoteLike, regexp.MustCompile(
		// The heredoc introducer accepts an interpolating tag body, not just
		// a bare word: t/comp/parser.t opens `<<"${a}{`, and a rule that
		// required \w after the quote filed that under General.
		"<<[~]?(['\"`\\\\][^\\n]*|\\w)|\\b(qw|qq|qr|tr|y|q)\\s*[^\\w\\s,;)=]|" +
			"\\bformat\\b|\\b__(DATA|END)__\\b")},

	// P2: regex, where modifiers change how the pattern body itself reads.
	{CategoryRegex, regexp.MustCompile(
		`[=!]~|\bs/|\bm/|/[a-z]*[gimsxoe][a-z]*\s*[;,)]`)},

	// P2: dereference, including the postfix forms.
	{CategoryDereference, regexp.MustCompile(
		`->\s*[@%$*&]\s*\*|->\s*[\[{]|[$@%]\s*\{\s*[\\$]|[@%]\$\w|\$\$+\w`)},

	// P2: the subroutine declaration surface — prototypes, attributes,
	// signatures, and the `&` call form.
	{CategorySubroutine, regexp.MustCompile(
		`\bsub\b[^;{]*[(:]|\bmy\s+sub\b|&\$?\w+\s*\(|\bprototype\b|\bAUTOLOAD\b`)},

	// P2: control flow, including the statement-modifier and label forms.
	{CategoryControlFlow, regexp.MustCompile(
		`\b(unless|until|foreach|for|while|if|elsif|else|continue)\b|\bdo\s*\{|` +
			`^\s*\w+\s*:\s*(for|while|until|\{)|\b(last|next|redo|goto)\b`)},
}

// categoriseLine files one source line under the taxonomy.
func categoriseLine(line string) Category {
	for _, rule := range taxonomyRules {
		if rule.pattern.MatchString(line) {
			return rule.category
		}
	}
	// P3. Not a default: it is the honest answer for a line whose construct
	// none of the rules above claims, and a baseline that is entirely
	// General is a taxonomy that was never populated.
	return CategoryGeneral
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
