// ABOUTME: Error recovery, §6.8.2: Error and Missing nodes, HasError, and the missing tokens synthesized.
// ABOUTME: Holds perigrin's R-a′ and R-g split: catch-alls are quiet Error nodes, specific refusals stay Unknown.

package parse_test

import (
	"bufio"
	"fmt"
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

// operandAbsentButValid are the seven test literals that parsed to
// missing_operand and that perl compiles (perl 5.44.0 -c): `state` and
// `field` alone are bareword calls without their features, `infix isa`
// is `'isa'->infix` without isa's, and a comma or fat comma at end of
// input ends a valid list. perigrin's R-g′ exempts them.
var operandAbsentButValid = []string{
	"state", "field", "infix isa", "infix ,", "infix =>", " files, ", "substr $x, 0, 1, ",
}

// TestRecoveryMissingOperandExemptions: none of R-g′'s seven builds a
// Missing node or a reportable refusal, and six carry no error at all.
// `infix isa`'s reading, `'isa'->infix`, is the unknown-class indirect
// method call parseIndirect declines while the sub table is open, so it is
// the quiet catch-all that decline is everywhere -- not a missing operand.
func TestRecoveryMissingOperandExemptions(t *testing.T) {
	for _, src := range operandAbsentButValid {
		root := parse.Parse([]byte(src))
		for _, n := range recoveryNodes(root) {
			if n.Kind == parse.Missing || !parse.CatchAll(n.Refusal) || src != "infix isa" {
				t.Errorf("%q: perl compiles it, and it holds %v %s: %v", src, n.Kind, n.Refusal, kinds(root))
			}
		}
		if firstUnknown(root) != nil && src != "infix isa" {
			t.Errorf("%q: perl compiles it, and it refuses: %v", src, kinds(root))
		}
	}
	// Under their features the same words are keywords, and perl 5.44.0
	// rejects each alone: `syntax error at EOF`.
	for _, src := range []string{"use v5.36; state", "use v5.36; infix isa"} {
		if root := parse.Parse([]byte(src)); len(withCode(root, parse.MissingOperand)) != 1 {
			t.Errorf("%q: want one missing_operand: %v", src, kinds(root))
		}
	}
}

// TestRecoveryMissingOperandIsReported: an operator with no operand before
// `;`, `}` or end of input yields one zero-width Missing node carrying
// missing_operand and no not_a_term on the punctuation (R-g).
func TestRecoveryMissingOperandIsReported(t *testing.T) {
	for _, src := range append(append([]string{}, operandBeforeTerminator...), "print 1+") {
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

// TestRecoveryOperatorAtEOFInsideBracket: an operator with no operand at
// end of file, inside a bracket still open, is one well-formed error. The
// Binary keeps two children, its right operand the unclosed_opener refusal
// on the operator's token, which holds both the Missing operand and the
// missing closer; nothing else refuses.
func TestRecoveryOperatorAtEOFInsideBracket(t *testing.T) {
	for _, c := range []struct{ src, closers string }{
		{"f(1 +", ")"},
		{"my $r = [1, 2 *", "]"},
		{"sub f { 2 *", "}"},
	} {
		root := parse.Parse([]byte(c.src))
		bin := firstOfKind(root, parse.Binary)
		for b := bin; b != nil && b.Text == ","; b = firstOfKind(b.Children[len(b.Children)-1], parse.Binary) {
			bin = b.Children[len(b.Children)-1]
		}
		if bin == nil || len(bin.Children) != 2 {
			t.Errorf("%q: the operator's Binary is %v, want two children: %v", c.src, bin, kinds(root))
			continue
		}
		right := bin.Children[1]
		if right.Kind != parse.Unknown || right.Refusal != parse.UnclosedOpener {
			t.Errorf("%q: right operand is %v %s, want the unclosed_opener Unknown", c.src, right.Kind, right.Refusal)
		}
		if got := len(withCode(root, parse.UnclosedOpener)) + len(withCode(root, parse.MissingOperand)); got != 2 {
			t.Errorf("%q: %d refusals, want the Unknown and the Missing operand inside it", c.src, got)
		}
		if got := closersIn(root); got != c.closers {
			t.Errorf("%q: closers %q, want %q", c.src, got, c.closers)
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

// errorNodes is every Error node of the tree.
func errorNodes(root *parse.Node) []*parse.Node {
	var out []*parse.Node
	for _, n := range nodesOf(root) {
		if n.Kind == parse.Error {
			out = append(out, n)
		}
	}
	return out
}

// tiles reports the first node whose children overlap or leave it, or the
// tree's text differing from src; "" when neither.
func tiles(root *parse.Node, src string) string {
	for _, n := range nodesOf(root) {
		at := n.Start
		for _, c := range n.Children {
			if c.Start < at || c.End > n.End {
				return fmt.Sprintf("%v [%d,%d) in %v [%d,%d)", c.Kind, c.Start, c.End, n.Kind, n.Start, n.End)
			}
			at = c.End
		}
	}
	if root.SourceText([]byte(src)) != src {
		return "the tree's text is not the buffer's"
	}
	return ""
}

// TestRecoveryPanicModeBounded: a catch-all's skip stops at the nearest
// §5.13.4 sync point -- a `;` at its depth (taken), the `}` closing the
// block it is in, a line-start statement keyword, a column-0 `sub` -- and
// never skips more than 100 tokens; what follows the sync point parses.
func TestRecoveryPanicModeBounded(t *testing.T) {
	for _, c := range []struct{ name, src, region string }{
		{"semicolon", "$a $b $c;\nfoo();\n", "$a $b $c;"},
		{"nearest point, a `;` before a later sub", "$a $b;\n$c;\nsub g { 1 }\n", "$a $b;"},
		{"a `;` inside parens is not one", "$a $b (1; 2);\nfoo();\n", "$a $b (1; 2);"},
		{"line-start keyword", "$a $b $c\nmy $y = 1;\n", "$a $b $c"},
		{"indented line-start keyword", "$a $b $c\n    local $x = 1;\n", "$a $b $c"},
		{"a line-start modifier after an expression is not one", "$a $b $c\n    if $x;\nfoo();\n", "$a $b $c\n    if $x;"},
		{"inside an open paren a line-start keyword is not one", "$a $b (1,\nmy $x);\nfoo();\n", "$a $b (1,\nmy $x);"},
		{"column-0 sub", "$a $b $c\nsub g { 1 }\n", "$a $b $c"},
		{"column-0 package", "$a $b $c\npackage Q;\n", "$a $b $c"},
		{"the block's closing brace", "sub f { $a $b }\nsub g { 1 }\n", "$a $b"},
		{"a block form ends at its brace", "$a $b { 1 }\nfoo();\n", "$a $b { 1 }"},
		{"a stray closer is skipped", "$a ) ] $b;\nfoo();\n", "$a ) ] $b;"},
		// The parse stopped inside the subscript: its `}` is the statement's,
		// not the block's, and ends no block form (Safe.pm).
		{"a bracket the statement opened before the skip", "for (@x) {\n    next if ++$h{refaddr $i} > 1;\n    g();\n}\n", "next if ++$h{refaddr $i} > 1;"},	} {
		root := parse.Parse([]byte(c.src))
		errs := errorNodes(root)
		if len(errs) != 1 {
			t.Errorf("%s: %q: %d Error nodes, want 1: %v", c.name, c.src, len(errs), kinds(root))
			continue
		}
		start := strings.Index(c.src, c.region)
		if e := errs[0]; e.Start != start || e.End != start+len(c.region) {
			t.Errorf("%s: %q: the skip covers %q, want %q", c.name, c.src, c.src[e.Start:e.End], c.region)
		}
		if n := len(recoveryNodes(root)); n != 1 {
			t.Errorf("%s: %q: %d recovery nodes, want the one Error, so what follows the sync point parses: %v", c.name, c.src, n, kinds(root))
		}
		if d := tiles(root, c.src); d != "" {
			t.Errorf("%s: %q: %s", c.name, c.src, d)
		}
	}

	// The cap: 150 tokens with no sync point skip 100 after the term that
	// failed, and the outer loop resumes there.
	src := strings.Repeat("$a ", 150) + ";\n"
	errs := errorNodes(parse.Parse([]byte(src)))
	if len(errs) != 2 {
		t.Fatalf("150 tokens: %d Error nodes, want 2", len(errs))
	}
	if got := errs[0].End - errs[0].Start; got != 101*3-1 {
		t.Errorf("150 tokens: the first skip spans %d bytes, want 101 tokens' %d", got, 101*3-1)
	}
}

// TestRecoverySkippedRegionIsOneErrorNode: everything one skip passes over
// is one Error node, carrying the catch-all code that started it (R-a′),
// where it was a cascade of one-token refusals.
func TestRecoverySkippedRegionIsOneErrorNode(t *testing.T) {
	for _, c := range []struct {
		src, region string
		code        parse.RefusalCode
	}{
		{"$a $b $c $d $e;\nfoo();\n", "$a $b $c $d $e;", parse.TrailingTokens},
		{") ) ) ) ;\nfoo();\n", ") ) ) ) ;", parse.TrailingTokens},
		{"} } } }\nmy $x = 1;\n", "} } } }", parse.NotATerm},
		{"sub f {\n  ) ) ) )\n}\nfoo();\n", ") ) ) )", parse.TrailingTokens},
		{")\nmy $x = 1;\n", ")", parse.TrailingTokens},
	} {
		root := parse.Parse([]byte(c.src))
		errs := errorNodes(root)
		if len(errs) != 1 {
			t.Errorf("%q: %d Error nodes, want 1: %v", c.src, len(errs), kinds(root))
			continue
		}
		e := errs[0]
		if got := c.src[e.Start:e.End]; got != c.region || e.Refusal != c.code {
			t.Errorf("%q: Error %s over %q, want %s over %q", c.src, e.Refusal, got, c.code, c.region)
		}
		if !e.HasError || !root.HasError {
			t.Errorf("%q: HasError is not set on the Error node and the root", c.src)
		}
	}

	// A specific refusal in the region is never hidden: its Unknown stays a
	// child of the Error node, with its code.
	src := "$a ? $b $c;\n"
	root := parse.Parse([]byte(src))
	errs := errorNodes(root)
	if len(errs) != 1 || len(errs[0].Children) != 1 || errs[0].Children[0].Kind != parse.Unknown || errs[0].Children[0].Refusal != parse.TernaryNoColon {
		t.Errorf("%q: want one Error holding the ternary_no_colon Unknown: %v", src, kinds(root))
	}
}

// subsSource is n subs at column 0, each a few statements and a loop.
func subsSource(n int) string {
	var b strings.Builder
	for i := range n {
		fmt.Fprintf(&b, "sub s%d {\n    my ($self, @v) = @_;\n    my $n = scalar @v;\n    for my $x (@v) {\n        $n += $x;\n    }\n    return $n;\n}\n\n", i)
	}
	return b.String()
}

// TestRecoveryErrorStaysInItsSub: an error inside one sub leaves every
// other sub's subtree equal to the error-free parse (§6.8.1 "Local"): by
// TreeDiff before the error, and after it by TreeDiffShifted by the length
// change.
func TestRecoveryErrorStaysInItsSub(t *testing.T) {
	clean := subsSource(6)
	stmt := "    my $n = scalar @v;\n"
	at := strings.Index(clean, "sub s3 {")
	at += strings.Index(clean[at:], stmt)
	want := parse.Parse([]byte(clean))
	for _, c := range []struct{ name, broken string }{
		{"trailing tokens", "    my $n = scalar @v $self;\n"},
		{"stray closer", "    my $n = ) scalar @v;\n"},
		{"cascade", "    ) ) ] ) $n\n"},
		{"missing operand", "    my $n = ;\n"},
		{"missing semicolon", "    my $n = scalar @v\n"},
		{"lost brace", stmt + "    {\n"},
		{"extra closing brace", "    my $n = scalar @v; }\n"},
	} {
		src := clean[:at] + c.broken + clean[at+len(stmt):]
		delta := len(src) - len(clean)
		got := parse.Parse([]byte(src))
		if len(recoveryNodes(got)) == 0 && firstUnknown(got) == nil {
			t.Errorf("%s: the injected error was not noticed", c.name)
		}
		byStart := map[int]*parse.Node{}
		for _, n := range got.Children {
			byStart[n.Start] = n
		}
		for _, sub := range want.Children {
			if sub.Kind != parse.Statement || sub.Start <= at && at < sub.End {
				continue
			}
			shift := 0
			if sub.Start > at {
				shift = delta
			}
			other := byStart[sub.Start+shift]
			if other == nil {
				t.Errorf("%s: the sub at byte %d is no longer a top-level statement: %v", c.name, sub.Start, kinds(got))
				continue
			}
			if d := parse.TreeDiffShifted(sub, other, shift); d != "" {
				t.Errorf("%s: the sub at byte %d differs from the error-free parse: %s", c.name, sub.Start, d)
			}
		}
	}
}

// TestRecoveryCapOnSyncFreeInput: 10,000 tokens with no sync point
// terminate, each skip spans at most 100 tokens after the one that failed,
// and the tree holds at most one Error node per skip, so at most 100.
func TestRecoveryCapOnSyncFreeInput(t *testing.T) {
	for _, tok := range []string{"$a ", ") ", "} ", "] "} {
		src := strings.Repeat(tok, 10000)
		root := parse.Parse([]byte(src))
		errs := errorNodes(root)
		if len(errs) == 0 || len(errs) > 100 {
			t.Errorf("%q x 10,000: %d Error nodes, want 1 to 100", tok, len(errs))
		}
		for _, e := range errs {
			if e.End-e.Start > 101*len(tok) {
				t.Errorf("%q x 10,000: an Error spans %d bytes, more than 101 tokens", tok, e.End-e.Start)
				break
			}
		}
		if d := tiles(root, src); d != "" {
			t.Errorf("%q x 10,000: %s", tok, d)
		}
	}
}

// TestRecoveryAlwaysAdvances: a sync point the statement parser rejects
// again at once still costs each recovery at least one token, so the parse
// terminates with every byte in a node and no empty Error node.
func TestRecoveryAlwaysAdvances(t *testing.T) {
	for _, word := range []string{"my", "our", "sub", "package", "if", "for", "use", "else", ")", "}"} {
		for _, sep := range []string{"\n", " ", "\n    "} {
			src := strings.Repeat(word+sep, 500)
			root := parse.Parse([]byte(src))
			if d := tiles(root, src); d != "" {
				t.Errorf("%q: %s", word+sep, d)
			}
			for _, e := range errorNodes(root) {
				if e.End <= e.Start {
					t.Errorf("%q: an empty Error node at %d", word+sep, e.Start)
					break
				}
			}
		}
	}
}

// TestRecoverySyncIgnoresSubInsideHeredocPODOrString: a column-0 `sub` in a
// heredoc body, POD or a string is one token's text, not a sync point, so
// neither the skip nor the brace recovery stops at it.
func TestRecoverySyncIgnoresSubInsideHeredocPODOrString(t *testing.T) {
	for _, c := range []struct {
		src   string
		decls int
	}{
		{"$a $b, <<EOF, 1;\nsub g { 1 }\nEOF\n", 0},
		{"$a $b,\n\n=pod\n\nsub g { 1 }\n\n=cut\n\n$c;\n", 0},
		{"$a $b \"\nsub g { 1 }\n\";\n", 0},
		{"$a $b q{\nsub g { 1 }\n};\n", 0},
		{"sub f {\n    print <<EOF;\nsub g { 1 }\nEOF\n", 1},
		{"sub f {\n    my $s = '\nsub g { 1 }\n';\n", 1},
	} {
		root := parse.Parse([]byte(c.src))
		if n := countKind(root, parse.Declaration); n != c.decls+strings.Count(c.src, "my ") {
			t.Errorf("%q: %d declarations: the `sub` inside was read as one: %v", c.src, n, kinds(root))
		}
		if n := len(errorNodes(root)); n > 1 {
			t.Errorf("%q: %d Error nodes, want at most 1: %v", c.src, n, kinds(root))
		}
		if c.decls == 1 && closersIn(root) != "}" {
			t.Errorf("%q: closers %q, want the one `}` at end of input", c.src, closersIn(root))
		}
	}
}

// TestRecoveryMissingBraceBeforeLaterSubs: a sub missing its `}` before ten
// column-0 subs closes at the first of them with a Missing `}`, and the ten
// stay at top level, each equal to its error-free parse.
func TestRecoveryMissingBraceBeforeLaterSubs(t *testing.T) {
	var subs strings.Builder
	for i := range 10 {
		fmt.Fprintf(&subs, "sub g%d {\n    my $x = %d;\n    return $x;\n}\n", i, i)
	}
	head := "sub f {\n    my $x = 1;\n    if ($x) {\n        print $x;\n    }\n\n"
	src := head + subs.String()
	root := parse.Parse([]byte(src))
	if got := closersIn(root); got != "}" {
		t.Errorf("closers %q, want one `}`: %v", got, kinds(root))
	}
	if f := root.Children[0]; f.End > len(head) {
		t.Errorf("sub f spans [%d,%d), past the column-0 sub at %d", f.Start, f.End, len(head))
	}
	if n := len(withCode(root, parse.UnclosedOpener)); n != 0 {
		t.Errorf("%d unclosed_opener refusals; the brace closes at the column-0 sub", n)
	}
	top := map[int]*parse.Node{}
	for _, n := range root.Children {
		top[n.Start] = n
	}
	subs10 := 0
	for _, sub := range parse.Parse([]byte(subs.String())).Children {
		if sub.Kind != parse.Statement {
			continue
		}
		subs10++
		got := top[sub.Start+len(head)]
		if got == nil {
			t.Errorf("the sub at byte %d is not at top level: %v", sub.Start+len(head), kinds(root))
			continue
		}
		if d := parse.TreeDiffShifted(sub, got, len(head)); d != "" {
			t.Errorf("the sub at byte %d: %s", got.Start, d)
		}
	}
	if subs10 != 10 {
		t.Fatalf("the error-free parse holds %d subs, want 10", subs10)
	}
}

// TestRecoveryLostBraceInsideAnExpression: a column-0 `sub` the expression
// parser takes as an anonymous sub, or reads inside a paren, closes nothing:
// the brace left open gains its Missing `}` at end of input, as before
// panic mode, in the refusal holding the last token. perl -c rejects both ("Missing right curly", "syntax
// error near sub g").
func TestRecoveryLostBraceInsideAnExpression(t *testing.T) {
	for _, src := range []string{
		"sub f {\n  my $cb = 1 &&\nsub { 1 };\n  return $cb;\n",
		"sub f {\n  my $x = (1,\nsub g { 1 }\n",
	} {
		root := parse.Parse([]byte(src))
		if !strings.Contains(closersIn(root), "}") || !root.HasError {
			t.Errorf("%q: closers %q, HasError %v: the open brace is not refused: %v", src, closersIn(root), root.HasError, kinds(root))
		}
	}
}

// TestRecoveryModifierAfterExpressionIsNotASyncPoint: a statement modifier
// first on its line continues the statement on the line before, so a skip
// passes it: Perl perl compiles, declined by a catch-all, keeps its one
// Error rather than splitting into fragments read as code.
func TestRecoveryModifierAfterExpressionIsNotASyncPoint(t *testing.T) {
	for _, word := range []string{"if", "unless", "while", "until", "for", "foreach"} {
		src := "use Carp;\nmy $loaded;\ncroak \"x\"\n\t\t" + word + " $loaded;\n"
		root := parse.Parse([]byte(src))
		at := strings.Index(src, "croak")
		end := strings.LastIndex(src, ";") + 1
		for _, c := range root.Children {
			if c.Start > at && c.Start < end {
				t.Errorf("%q: the statement is split at byte %d: %v", src, c.Start, kinds(root))
				break
			}
		}
		if n := len(errorNodes(root)); n > 1 {
			t.Errorf("%q: %d Error nodes, want at most 1", src, n)
		}
	}
}

// TestRecoveryErrorAtEOF: an error in the last statement with no sync point
// before end of input is one Error node ending at len(src).
func TestRecoveryErrorAtEOF(t *testing.T) {
	for _, src := range []string{
		"my $x = 1;\nfoo $a $b",
		"my $x = 1;\n$a $b $c $d",
		"sub f { 1 }\n) ) )",
		"my $x = 1;\n$a $b (1, 2",
	} {
		root := parse.Parse([]byte(src))
		errs := errorNodes(root)
		if len(errs) != 1 {
			t.Errorf("%q: %d Error nodes, want 1: %v", src, len(errs), kinds(root))
			continue
		}
		if errs[0].End != len(src) {
			t.Errorf("%q: the Error ends at %d, want %d", src, errs[0].End, len(src))
		}
	}
}

// TestRecoveryEdit5kOneErrorAtMostThreeNodes: edit5k.pl with one injected
// syntax error, of each shape a live buffer passes through, holds at most 3
// Error nodes in the whole tree.
func TestRecoveryEdit5kOneErrorAtMostThreeNodes(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "psc", "testdata", "latency", "edit5k.pl"))
	if err != nil {
		t.Fatal(err)
	}
	clean := string(raw)
	if n := len(errorNodes(parse.Parse(raw))); n != 0 {
		t.Fatalf("edit5k.pl holds %d Error nodes before any is injected", n)
	}
	// The 40th of its 80 `sub scale` bodies, mid-file.
	stmt := "    my @kept = grep { defined && $_ > 0 } @values;\n"
	if n := strings.Count(clean, stmt); n < 40 {
		t.Fatalf("edit5k.pl holds the statement %d times, want at least 40", n)
	}
	at := -1
	for range 40 {
		at += 1 + strings.Index(clean[at+1:], stmt)
	}
	for _, c := range []struct{ name, broken string }{
		{"trailing tokens", "    my @kept = grep { defined && $_ > 0 } @values $x;\n"},
		{"stray closer", "    my @kept = ) grep { defined && $_ > 0 } @values;\n"},
		{"cascade", "    ) ) ] ) @values\n"},
		{"missing operand", "    my @kept = ;\n"},
		{"missing semicolon", "    my @kept = grep { defined && $_ > 0 } @values\n"},
		{"unclosed brace", "    my @kept = grep { defined && $_ > 0 @values;\n"},
		{"lost brace", stmt + "    {\n"},
		{"extra closing brace", "    my @kept = grep { defined && $_ > 0 } @values; }\n"},
	} {
		src := clean[:at] + c.broken + clean[at+len(stmt):]
		root := parse.Parse([]byte(src))
		if n := len(errorNodes(root)); n > 3 {
			t.Errorf("%s: %d Error nodes, want at most 3", c.name, n)
		}
		if len(recoveryNodes(root)) == 0 && firstUnknown(root) == nil {
			t.Errorf("%s: the injected error was not noticed", c.name)
		}
	}
}
