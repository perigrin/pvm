// ABOUTME: Error recovery, §6.8.2: Error and Missing nodes, HasError, and the missing tokens synthesized.
// ABOUTME: Holds perigrin's R-a′ and R-g split: catch-alls are quiet Error nodes, specific refusals stay Unknown.

package parse_test

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// invalidBuffers are buffers perl rejects, in the shapes a live buffer
// passes through while it is typed.
var invalidBuffers = []string{
	"my $x = ;",
	"my $x = ",
	"my $y = 1 + ;",
	"sub f { return 2 * }",
	"print 1+",
	"foo(1, 2;",
	"my $x = 1\nmy $y = 2;\n",
	"sub ;",
	"print $",
	"f(sub, 1);",
	"sub f {\n  my $x = 1;\n",
	"my @a = (1, 2",
	"if ($x) { foo(1; }",
	"$a $b;\n",
	"sub f { print \"abc",
	"sub f {\n  print <<EOF;\nabc\n",
	"use feature 'signatures';\nsub f ($x, $y",
	"map { $_",
	"}{",
	"{",
	"\\",
	"my $x = $a ? $b;\n",
	"f(1, [2, {3",
	"sub f { 1 +",
}

// nodesOf is every node of the tree, root first, in source order.
func nodesOf(n *parse.Node) []*parse.Node {
	out := []*parse.Node{n}
	for _, c := range n.Children {
		out = append(out, nodesOf(c)...)
	}
	return out
}

// recoveryNodes is every Error and Missing node of the tree.
func recoveryNodes(n *parse.Node) []*parse.Node {
	var out []*parse.Node
	for _, m := range nodesOf(n) {
		if m.Kind == parse.Error || m.Kind == parse.Missing {
			out = append(out, m)
		}
	}
	return out
}

// holder is the top-level child of root whose span holds offset at, a
// zero-width node's included.
func holder(root *parse.Node, at int) *parse.Node {
	for _, c := range root.Children {
		if c.Start <= at && at <= c.End && c.End > c.Start {
			return c
		}
	}
	return nil
}

// TestRecoveryMissingTokens: each missing token of §6.8.2's table is
// synthesized as a zero-width Missing node, and the statement it belongs to
// completes as a Statement rather than as one refusal.
func TestRecoveryMissingTokens(t *testing.T) {
	cases := []struct {
		name, src string
		text      string // the Missing node's Text
		code      parse.RefusalCode
		at        int
	}{
		// The next token starts a new statement on a later line.
		{"semicolon", "my $x = 1\nmy $y = 2;\n", ";", "", 9},
		// EOF with a block open.
		{"brace", "sub f {\n  my $x = 1;\n", "}", "", 20},
		// A `;` arrives with a paren open.
		{"paren", "foo(1, 2;\n", ")", "", 9},
		// A binary operator with no right operand.
		{"expression", "my $x = ;", "", parse.MissingOperand, 7},
		// `sub` with no name, and `$` with no name.
		{"identifier after sub", "f(sub, 1);", "", "", 5},
		{"identifier after $", "print $", "", "", 7},
	}
	for _, c := range cases {
		root := parse.Parse([]byte(c.src))
		var got []*parse.Node
		for _, m := range recoveryNodes(root) {
			if m.Kind == parse.Missing {
				got = append(got, m)
			}
		}
		if len(got) != 1 {
			t.Errorf("%s: %q: %d Missing nodes, want 1: %v", c.name, c.src, len(got), kinds(root))
			continue
		}
		m := got[0]
		if m.Text != c.text || m.Refusal != c.code || m.Start != c.at || m.End != c.at {
			t.Errorf("%s: %q: Missing %q %q at [%d,%d), want %q %q at %d",
				c.name, c.src, m.Text, m.Refusal, m.Start, m.End, c.text, c.code, c.at)
		}
		if h := holder(root, c.at); h == nil || h.Kind != parse.Statement {
			t.Errorf("%s: %q: the statement holding the Missing node did not complete: %v", c.name, c.src, kinds(root))
		}
	}
	// The semicolon's two statements are both read: the second is a
	// declaration, not swallowed bytes.
	root := parse.Parse([]byte("my $x = 1\nmy $y = 2;\n"))
	if n := countKind(root, parse.Declaration); n != 2 {
		t.Errorf("a synthesized `;` left %d declarations, want 2: %v", n, kinds(root))
	}
}

// TestHasErrorPropagatedAtConstruction: HasError is a field the parse sets,
// read without a walk, and it is true on exactly the Error and Missing
// nodes and their ancestors.
func TestHasErrorPropagatedAtConstruction(t *testing.T) {
	var want func(*parse.Node) bool
	want = func(n *parse.Node) bool {
		has := n.Kind == parse.Error || n.Kind == parse.Missing
		for _, c := range n.Children {
			if want(c) {
				has = true
			}
		}
		return has
	}
	valid := []string{"my $x = 42;", "sub f { 1 }\nf();", "if ($x) { print 1 } else { print 2 }", ""}
	for _, src := range append(append([]string{}, valid...), invalidBuffers...) {
		root := parse.Parse([]byte(src))
		for _, n := range nodesOf(root) {
			if w := want(n); n.HasError != w {
				t.Errorf("%q: %v [%d,%d) HasError = %v, want %v", src, n.Kind, n.Start, n.End, n.HasError, w)
			}
		}
	}
	for _, src := range valid {
		if parse.Parse([]byte(src)).HasError {
			t.Errorf("%q: valid Perl has HasError at the root", src)
		}
	}
	for _, src := range invalidBuffers {
		if root := parse.Parse([]byte(src)); len(recoveryNodes(root)) > 0 && !root.HasError {
			t.Errorf("%q: holds a recovery node and the root's HasError is false", src)
		}
	}
}

// TestRecoveryCoversEveryByte: every byte of an invalid buffer is covered by
// some node -- the root's children tile it -- and the tree's text is the
// buffer's.
func TestRecoveryCoversEveryByte(t *testing.T) {
	for _, src := range invalidBuffers {
		root := parse.Parse([]byte(src))
		at := 0
		for _, c := range root.Children {
			if c.Start != at {
				t.Errorf("%q: bytes [%d,%d) belong to no node", src, at, c.Start)
			}
			at = c.End
		}
		if at != len(src) {
			t.Errorf("%q: bytes [%d,%d) belong to no node", src, at, len(src))
		}
		if got := root.SourceText([]byte(src)); got != src {
			t.Errorf("%q: the tree's text is %q", src, got)
		}
	}
}

// refusalClass is how perigrin's rulings sort a refusal code: a quiet
// catch-all recovery replaces with an Error node (R-a′), the one reported
// missing token recovery replaces with a Missing node (R-g), or a specific
// refusal that stays an Unknown (R-a).
type refusalClass int

const (
	quietCatchAll refusalClass = iota
	reportedMissing
	specificRefusal
)

// refusalClasses classifies every code in RefusalSites. A code added later
// fails TestRecoveryReplacesCatchAllRefusals until it is classified here.
var refusalClasses = map[parse.RefusalCode]refusalClass{
	parse.NotATerm:               quietCatchAll,
	parse.TrailingTokens:         quietCatchAll,
	parse.NotAnExpression:        quietCatchAll,
	parse.MissingOperand:         reportedMissing,
	parse.CallArity:              specificRefusal,
	parse.UnclosedOpener:         specificRefusal,
	parse.RefScalarSlot:          specificRefusal,
	parse.ChainClassMismatch:     specificRefusal,
	parse.FormatNotTerminated:    specificRefusal,
	parse.NonassocRepeated:       specificRefusal,
	parse.UnclosedBrace:          specificRefusal,
	parse.EmptyDeref:             specificRefusal,
	parse.TernaryNoColon:         specificRefusal,
	parse.UnimplementedStatement: specificRefusal,
	parse.AmbiguousSigil:         specificRefusal,
	parse.NestingTooDeep:         specificRefusal,
}

// withCode is every refusal in the tree carrying code.
func withCode(root *parse.Node, code parse.RefusalCode) []*parse.Node {
	var out []*parse.Node
	for _, n := range nodesOf(root) {
		if n.Refused() && n.Refusal == code {
			out = append(out, n)
		}
	}
	return out
}

// TestRecoveryReplacesCatchAllRefusals: each reachable code recovery
// replaces yields, on its inventory input (refusingSources), an Error or
// Missing node carrying it with HasError set, and no Unknown carrying it;
// and every code in RefusalSites is classified, CatchAll agreeing.
func TestRecoveryReplacesCatchAllRefusals(t *testing.T) {
	for code := range parse.RefusalSites {
		class, ok := refusalClasses[code]
		if !ok {
			t.Errorf("%s: not classified as quiet catch-all, reported missing token or specific", code)
			continue
		}
		if parse.CatchAll(code) != (class == quietCatchAll) {
			t.Errorf("%s: CatchAll = %v, but it is classified %d", code, parse.CatchAll(code), class)
		}
	}
	if len(refusalClasses) != len(parse.RefusalSites) {
		t.Errorf("%d codes classified, RefusalSites has %d", len(refusalClasses), len(parse.RefusalSites))
	}

	reached := map[parse.RefusalCode]bool{}
	for _, src := range refusingSources() {
		root := parse.Parse([]byte(src))
		for _, n := range nodesOf(root) {
			if !n.Refused() || refusalClasses[n.Refusal] == specificRefusal {
				continue
			}
			reached[n.Refusal] = true
			if n.Kind != parse.Error && n.Kind != parse.Missing {
				t.Errorf("%q: %s is on a %v, want an Error or Missing node", src, n.Refusal, n.Kind)
			}
			if !n.HasError {
				t.Errorf("%q: the %v carrying %s has HasError false", src, n.Kind, n.Refusal)
			}
		}
	}
	// not_an_expression is wired but unreachable (refusingSources says why).
	for _, code := range []parse.RefusalCode{parse.NotATerm, parse.TrailingTokens, parse.MissingOperand} {
		if !reached[code] {
			t.Errorf("%s: no inventory input reaches it", code)
		}
	}
}

// TestRecoveryCatchAllOnValidPerlIsUnreported: Perl perl compiles, where a
// quiet catch-all fires, yields only nodes CatchAll classifies catch-all,
// so nothing reportable (R-a′); and Perl that parses with no refusal yields
// no Error or Missing node and HasError false at the root.
func TestRecoveryCatchAllOnValidPerlIsUnreported(t *testing.T) {
	lits := readCatchAllCompiled(t)
	if len(lits) < 400 {
		t.Fatalf("catchall_compiled.txt holds %d literals; the inventory is broken, not the parser", len(lits))
	}
	quiet := 0
	for _, src := range lits {
		root := parse.Parse([]byte(src))
		for _, n := range recoveryNodes(root) {
			if !parse.CatchAll(n.Refusal) {
				t.Errorf("%q: perl compiles it, and it holds a reportable %v %q %s", src, n.Kind, n.Text, n.Refusal)
			}
			quiet++
		}
	}
	if quiet == 0 {
		t.Error("no inventory literal holds a catch-all node; the check checks nothing")
	}

	for _, src := range fuzzPerlSeeds {
		if root := parse.Parse([]byte(src)); root.HasError || len(recoveryNodes(root)) > 0 {
			t.Errorf("%q: valid Perl holds a recovery node", src)
		}
	}

	t.Run("compiled-clean", func(t *testing.T) {
		dir, _ := perlGitTFiles(t)
		paths := readCompiledSet(t)
		unclean, err := readCompiledClean(filepath.Join("testdata", "compiledclean.txt"), len(paths))
		if err != nil {
			t.Fatal(err)
		}
		roots := perlGitTRoots(dir)
		s := parse.NewSession()
		checked := 0
		for _, rel := range paths {
			if unclean[rel] {
				continue
			}
			n, err := s.ParseFileFrom(filepath.Join(dir, rel), roots...)
			if err != nil {
				continue
			}
			checked++
			if n.HasError || len(recoveryNodes(n)) > 0 {
				t.Errorf("%s: parses with no refusal on pu, and now holds a recovery node", rel)
			}
		}
		if checked < 500 {
			t.Errorf("checked %d compiled-clean files; the set is broken", checked)
		}
	})
}

// readCatchAllCompiled reads testdata/catchall_compiled.txt.
func readCatchAllCompiled(t *testing.T) []string {
	t.Helper()
	f, err := os.Open(filepath.Join("testdata", "catchall_compiled.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		s, err := strconv.Unquote(line)
		if err != nil {
			t.Fatalf("catchall_compiled.txt: %q: %v", line, err)
		}
		out = append(out, s)
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// TestRecoveryUnknownStaysDistinctFromError: a specific refusal -- a
// perlRefuses code, unimplemented_statement, ambiguous_sigil,
// nesting_too_deep -- is still an Unknown, and the Unknown alone sets no
// HasError: only a Missing closer it gained does.
func TestRecoveryUnknownStaysDistinctFromError(t *testing.T) {
	reached := 0
	for _, src := range refusingSources() {
		root := parse.Parse([]byte(src))
		for _, n := range nodesOf(root) {
			if !n.Refused() || refusalClasses[n.Refusal] != specificRefusal {
				continue
			}
			reached++
			if n.Kind != parse.Unknown {
				t.Errorf("%q: specific refusal %s is a %v, want Unknown", src, n.Refusal, n.Kind)
			}
			missingChild := false
			for _, c := range n.Children {
				missingChild = missingChild || c.HasError
			}
			if n.HasError != missingChild {
				t.Errorf("%q: the %s Unknown has HasError %v, and a Missing child: %v", src, n.Refusal, n.HasError, missingChild)
			}
		}
	}
	if reached < 9 {
		t.Errorf("reached %d specific refusals; the inventory is broken", reached)
	}
	// The nesting_too_deep fixture is balanced, so nothing in it is missing.
	if root := parse.Parse([]byte(refusingSources()[len(refusingSources())-1])); root.HasError {
		t.Error("a balanced nesting_too_deep refusal set HasError")
	}
}

// operandBeforeTerminator are R-g's moves: an operator with no operand
// before `;` or `}` was not_a_term on the punctuation, and is missing_operand.
var operandBeforeTerminator = []string{
	"my $x = ;",
	"my $y = 1 + ;",
	"sub f { return 2 * }",
}

// TestRecoveryKeepsRefusalCodes: every refusal code that fires on its
// inventory input still fires with the same code -- on an Unknown for a
// specific refusal, on an Error or Missing node for one recovery replaces,
// with a Missing closer where §6.8.2 synthesizes one. The exception is an
// operator before `;` or `}`, which R-g moves from not_a_term to
// missing_operand.
func TestRecoveryKeepsRefusalCodes(t *testing.T) {
	for _, src := range refusingSources() {
		want := firstRefusalCode(t, src)
		root := parse.Parse([]byte(src))
		got := withCode(root, want)
		if len(got) == 0 {
			t.Errorf("%q: no longer refuses with %s", src, want)
			continue
		}
		n := got[0]
		isUnknown := n.Kind == parse.Unknown
		if isUnknown != (refusalClasses[want] == specificRefusal) {
			t.Errorf("%q: %s is on a %v", src, want, n.Kind)
		}
	}
	for _, src := range []string{"sub f {\n", "map { $_"} {
		root := parse.Parse([]byte(src))
		closer := false
		for _, m := range recoveryNodes(root) {
			closer = closer || m.Kind == parse.Missing && m.Text == "}"
		}
		if !closer {
			t.Errorf("%q: no Missing `}` beside its refusal: %v", src, kinds(root))
		}
	}
	for _, src := range operandBeforeTerminator {
		root := parse.Parse([]byte(src))
		if len(withCode(root, parse.NotATerm)) != 0 || len(withCode(root, parse.MissingOperand)) != 1 {
			t.Errorf("%q: want one missing_operand and no not_a_term: %v", src, kinds(root))
		}
	}
}

// TestRecoveryMissingOperandIsReported: an operator with no operand before
// `;`, `}` or end of input yields one zero-width Missing node carrying
// missing_operand and no not_a_term on the punctuation (R-g). A buffer that
// is exactly `state` or `field` yields one too, although perl -c compiles
// each as a bareword call: the cost R-g accepts, recorded for perigrin.
func TestRecoveryMissingOperandIsReported(t *testing.T) {
	for _, src := range append(append([]string{}, operandBeforeTerminator...), "print 1+", "state", "field") {
		root := parse.Parse([]byte(src))
		got := withCode(root, parse.MissingOperand)
		if len(got) != 1 {
			t.Errorf("%q: %d missing_operand nodes, want 1: %v", src, len(got), kinds(root))
			continue
		}
		if m := got[0]; m.Kind != parse.Missing || m.Start != m.End || len(m.Children) != 0 {
			t.Errorf("%q: missing_operand is a %v [%d,%d), want a zero-width Missing node", src, m.Kind, m.Start, m.End)
		}
		if n := withCode(root, parse.NotATerm); len(n) != 0 {
			t.Errorf("%q: not_a_term on the punctuation, %v", src, kinds(root))
		}
	}
	if parse.CatchAll(parse.MissingOperand) {
		t.Error("CatchAll(missing_operand) is true; R-g reports it")
	}
	for _, code := range []parse.RefusalCode{parse.NotATerm, parse.TrailingTokens, parse.NotAnExpression} {
		if !parse.CatchAll(code) {
			t.Errorf("CatchAll(%s) is false", code)
		}
	}
	n := 0
	for code := range parse.RefusalSites {
		if parse.CatchAll(code) {
			n++
		}
	}
	if n != 3 {
		t.Errorf("CatchAll is true for %d codes, want exactly 3", n)
	}
}

// TestRatchetHelpersCountCatchAllAsUnknown: countUnknown and
// firstUnknownLine count a catch-all Error node exactly as they counted the
// Unknown it replaced, so a ratchet's trailing_tokens cannot drop to 0 and
// pass only-tighten as an improvement.
func TestRatchetHelpersCountCatchAllAsUnknown(t *testing.T) {
	src := []byte("my $x = 1;\n$a $b;\n")
	root := parse.Parse(src)
	if len(withCode(root, parse.TrailingTokens)) != 1 || withCode(root, parse.TrailingTokens)[0].Kind != parse.Error {
		t.Fatalf("%q: want one trailing_tokens Error node: %v", src, kinds(root))
	}
	if got := countUnknown(root); got != 1 {
		t.Errorf("countUnknown = %d, want 1", got)
	}
	if got := firstUnknownLine(root, src); got != 2 {
		t.Errorf("firstUnknownLine = %d, want 2", got)
	}

	// t1.ratchet's own one.
	t.Run("t1", func(t *testing.T) {
		const rel = "source_filter_subroutine_lexical_storage.t"
		dir, _ := t1Files(t)
		n, err := parse.NewSession().ParseFileFrom(filepath.Join(dir, rel), t1Roots(t)...)
		if err != nil {
			t.Fatal(err)
		}
		if got := countUnknown(n); got != 1 {
			t.Errorf("%s: countUnknown = %d, t1.ratchet pins 1", rel, got)
		}
		if len(withCode(n, parse.TrailingTokens)) != 1 {
			t.Errorf("%s: want its one refusal a trailing_tokens Error", rel)
		}
	})
}

// TestRecoveryEmptyAndTriviaOnlyInput: nothing to recover from has no
// recovery node.
func TestRecoveryEmptyAndTriviaOnlyInput(t *testing.T) {
	for _, src := range []string{"", "  \n\t\n", "# a comment\n", "=pod\n\ntext\n\n=cut\n", "__END__\nnot perl ) (\n"} {
		root := parse.Parse([]byte(src))
		if root.HasError || len(recoveryNodes(root)) > 0 {
			t.Errorf("%q: holds a recovery node: %v", src, kinds(root))
		}
	}
}

// closersIn is the Text of every Missing closer in the tree, in order.
func closersIn(root *parse.Node) string {
	var b strings.Builder
	for _, m := range recoveryNodes(root) {
		if m.Kind == parse.Missing && strings.ContainsAny(m.Text, "})]") {
			b.WriteString(m.Text)
		}
	}
	return b.String()
}

// TestRecoveryEOFMidConstruct: end of input inside a sub body, a signature,
// a list, an unterminated string and an unterminated heredoc gives a total
// tree with one missing closer per open construct.
func TestRecoveryEOFMidConstruct(t *testing.T) {
	for _, c := range []struct{ src, want string }{
		{"sub f {\n  my $x = 1;\n", "}"},
		{"use feature 'signatures';\nsub f ($x, $y", ")"},
		{"my @a = (1, 2", ")"},
		{"f(1, [2, {3", "}])"},
		{"sub f { print \"abc", "}"},
		{"sub f {\n  print <<EOF;\nabc\n", "}"},
		{"sub f { g(<<EOF", ")}"},
	} {
		root := parse.Parse([]byte(c.src))
		if got := closersIn(root); got != c.want {
			t.Errorf("%q: closers %q, want %q: %v", c.src, got, c.want, kinds(root))
		}
		if root.SourceText([]byte(c.src)) != c.src {
			t.Errorf("%q: the tree is not total", c.src)
		}
	}
}

// TestRecoveryDeepNestingBounded: 999 open brackets, the deepest maxDepth
// accepts, synthesize one closer each; 10,000 terminate as one
// nesting_too_deep refusal with no closer synthesized past maxDepth.
func TestRecoveryDeepNestingBounded(t *testing.T) {
	for _, open := range []string{"{", "("} {
		closer := map[string]string{"{": "}", "(": ")"}[open]
		root := parse.Parse([]byte(strings.Repeat(open, 999)))
		if got := closersIn(root); got != strings.Repeat(closer, 999) {
			t.Errorf("999 %q: %d closers, want 999 %q", open, len(got), closer)
		}
		if n := len(withCode(root, parse.NestingTooDeep)); n != 0 {
			t.Errorf("999 %q: refused nesting_too_deep", open)
		}

		root = parse.Parse([]byte(strings.Repeat(open, 10000)))
		if n := len(withCode(root, parse.NestingTooDeep)); n != 1 {
			t.Errorf("10,000 %q: %d nesting_too_deep refusals, want 1", open, n)
		}
		// The decline is a stack guard, not a syntax error: nothing is
		// synthesized past it.
		if got := closersIn(root); got != "" {
			t.Errorf("10,000 %q: %d closers synthesized past maxDepth", open, len(got))
		}
	}
}

// TestRecoveryNodeSpansInBounds: on invalid input every node lies in the
// buffer, a Missing node is zero-width and childless, and siblings do not
// overlap.
func TestRecoveryNodeSpansInBounds(t *testing.T) {
	bufs := append(append(append([]string{}, invalidBuffers...), fuzzSeeds...), fuzzEdgeSeeds...)
	for _, src := range bufs {
		root := parse.Parse([]byte(src))
		for _, n := range nodesOf(root) {
			if n.Start < 0 || n.Start > n.End || n.End > len(src) {
				t.Errorf("%q: %v spans [%d,%d), out of bounds for %d bytes", src, n.Kind, n.Start, n.End, len(src))
			}
			if n.Kind == parse.Missing && (n.Start != n.End || len(n.Children) != 0) {
				t.Errorf("%q: Missing node [%d,%d) with %d children", src, n.Start, n.End, len(n.Children))
			}
			prev := n.Start
			for _, c := range n.Children {
				if c.Start < prev || c.End > n.End {
					t.Errorf("%q: %v [%d,%d) overlaps its sibling or leaves its %v [%d,%d)", src, c.Kind, c.Start, c.End, n.Kind, n.Start, n.End)
				}
				prev = c.End
			}
		}
	}
}
