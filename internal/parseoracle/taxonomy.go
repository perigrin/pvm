// ABOUTME: Categorises a file's first parse error by the Perl construct that broke it.
// ABOUTME: A raw count says "203 files fail"; a taxonomy says which 48 are heredocs, which is a work plan.

package parseoracle

import (
	"bytes"
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

	site, ok := errorSite(root, src)
	if !ok {
		// A degenerate tree with no leaked hidden rule (a collapsed varname
		// is the other signal) has no position to read at all.
		if tree.IsDegenerate() {
			return CategoryGeneral
		}
		return CategoryNone
	}
	return categoriseLine(site)
}

// errorSite returns the source line where the parse broke.
//
// Three kinds of tree carry a defect, and each leaves its evidence in a
// different place. An ERROR or MISSING node marks a span, and the line where
// that span ends is where recovery gave up. A tree whose root carries
// HasError but holds no such node is one where recovery halted outright:
// the source_file node stops before the source does, and the first unparsed
// line is the site — measured on the corpus, eight no-answer files were
// uncategorised for exactly this shape. A degenerate tree recorded no error
// at all, so the line holding the hidden rule that leaked is the only
// position it offers.
func errorSite(root *parser.Node, src []byte) (string, bool) {
	if e := firstErrorNode(root); e != nil {
		return lineAt(src, int(e.EndByte())), true
	}
	if root.HasError() {
		end := int(root.EndByte())
		if end < len(bytes.TrimRight(src, " \t\r\n")) {
			// Skip the whitespace between the last parsed token and the
			// first unparsed one, so a halt at a line end reads the next
			// line rather than an empty one.
			for end < len(src) && (src[end] == '\n' || src[end] == ' ' || src[end] == '\t' || src[end] == '\r') {
				end++
			}
			return lineAt(src, end), true
		}
	}
	if h := firstHiddenNode(root); h != nil {
		return lineAt(src, int(h.StartByte())), true
	}
	return "", false
}

// lineAt returns the trimmed source line containing byte offset pos. Recovery
// consumes up to the token it could not place, so that token is at or just
// before an error span's end.
func lineAt(src []byte, pos int) string {
	if pos > len(src) {
		pos = len(src)
	}
	start := strings.LastIndexByte(string(src[:pos]), '\n') + 1
	lineEnd := len(src)
	if i := strings.IndexByte(string(src[pos:]), '\n'); i >= 0 {
		lineEnd = pos + i
	}
	line := strings.TrimSpace(string(src[start:lineEnd]))
	if line == "" {
		// An error ending on a blank line: fall back to the whole span's
		// tail, which still carries the construct.
		tail := string(src[max(0, pos-160):pos])
		return strings.TrimSpace(tail)
	}
	return line
}

// firstErrorNode finds the first ERROR or MISSING node in document order,
// descending so the innermost — and therefore most specific — span wins. A
// MISSING node is a token recovery inserted rather than a span it gave up
// on, and a tree can carry one as its only defect.
func firstErrorNode(n *parser.Node) *parser.Node {
	if n == nil || !n.HasError() {
		return nil
	}
	if n.IsMissing() {
		return n
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

// firstHiddenNode finds the first named node whose kind is a hidden rule,
// which is the signal Tree.IsDegenerate reads: a hidden rule is inlined in a
// successful parse and surfaces only where recovery bailed out mid-rule.
func firstHiddenNode(n *parser.Node) *parser.Node {
	if n == nil {
		return nil
	}
	if n.IsNamed() && strings.HasPrefix(n.Kind(), "_") {
		return n
	}
	for i := 0; i < n.NamedChildCount(); i++ {
		if h := firstHiddenNode(n.NamedChild(i)); h != nil {
			return h
		}
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

	// P2: an identifier the lexer cannot read, which fails before any
	// construct rule sees the line and so is claimed ahead of them. A
	// non-ASCII rune counts only in identifier position — after a sigil,
	// `::`, `->`, or a `package`/`sub` keyword — so a name inside a string
	// on the same line does not claim it. The `'` separator likewise needs
	// a word on both sides, which leaves `$'` (the postmatch variable) alone.
	{CategoryIdentifier, regexp.MustCompile(
		`(?:[$@%&*]\{?|::|->|\b(?:package|sub)\s+)[\w:]*[^\x00-\x7F]|` +
			`(?:[$@%&*]|\b(?:package|sub)\s+)[\w:]+'\w`)},

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

	// P2: an operator the lexer cannot separate from its operand. This
	// sits ahead of ControlFlow because `if (foo && 1)` breaks on `foo &&`,
	// not on `if`. The repetition rule wants the whole quoted operand so
	// that `'x1234'` — a string that merely starts with x — is not read as
	// `'` followed by `x1234`; `(1) x 3` with spaces parses cleanly and is
	// not claimed. The bareword rule excludes a sigil, `>` or `:` before
	// the word, so `$x &&`, `->m &&` and `A::b &&` are left to their own
	// constructs.
	{CategoryOperator, regexp.MustCompile(
		`\)x\d|'[^']*'x\d|"[^"]*"x\d|\s[&|^]\.\s|` +
			`(?:^|[^\w$@%&*>:'"-])[A-Za-z_]\w*\s*&&`)},

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
