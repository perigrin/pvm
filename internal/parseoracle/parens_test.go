// ABOUTME: Holds psc's Paren flag to perl's OPf_PARENS bit, op by op, statement by statement.
// ABOUTME: perl keeps one bit for a source paren on the op it wraps; psc must put Paren on the matching node.

package parseoracle_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/parseoracle"
	"tamarou.com/pvm/internal/perltest"
)

// TestParenWalkReadsPARENSOnNulledOps: the bit can sit on an op the
// optimiser nulled. Measured on perl 5.44.0, `f((1,2))` with `sub f {}`
// declared is `ex-list lKPM`, the only P in the statement, and `f(1,2)` has
// none. A walk that skipped ex- ops would see nothing on both.
func TestParenWalkReadsPARENSOnNulledOps(t *testing.T) {
	p := perltest.PinnedPerl(t)
	for _, tc := range []struct {
		src  string
		want []parenOp
	}{
		{"sub f {}\nf((1,2));\n", []parenOp{{2, "ex-list", ""}}},
		{"sub f {}\nf(1,2);\n", nil},
	} {
		facts, err := parseoracle.Ask(context.Background(), []byte(tc.src), parseoracle.Options{Perl: p.Path})
		if err != nil {
			t.Fatalf("Ask(%q): %v", tc.src, err)
		}
		if got := perlParens(facts); !sameParenOps(got, tc.want) {
			t.Errorf("perl's P ops for %q are %v, want %v", tc.src, got, tc.want)
		}
	}
}

// parenOp is one OPf_PARENS fact: an op name on a statement's first line,
// which is the line perl's nextstate records for every op of the statement,
// and the variable the op reads where it reads one (padsv, ex-rv2sv,
// rv2av ...), so `($x) + $y` and `$x + ($y)` differ.
type parenOp struct {
	Line int
	Op   string
	Var  string
}

// label is o's op name and variable, as the comparison matches them.
func (o parenOp) label() string {
	if o.Var == "" {
		return o.Op
	}
	return o.Op + " " + o.Var
}

// structuralParens are ops perl marks P with no source paren behind them.
// Measured on perl 5.44.0, for a block of two or more statements: `leave
// vKP` on an `if`'s, `else`'s and `do`'s block; `leaveloop vKP` and
// `lineseq vKP` on a `for` or `while`. A one-statement block is a `scope`,
// listed too.
//
// A `null` with P is NOT structural in general: it is where perl keeps a
// paren around `&&`, `||`, `//`, `and`, `or`, `?:` and `do {}`, `null
// sKP/1`, and in void context `($x && 1);` is `null vKP/1`. Only the null
// over an `if`/`unless` condition is structural, and the walk marks it:
// its logop carries OPpSTATEMENT (`and vK/STMT`).
var structuralParens = map[string]bool{
	"leave": true, "leaveloop": true, "lineseq": true, "scope": true,
}

// perlParens is perl's side: every op of the file carrying OPf_PARENS,
// less the structural ones.
func perlParens(f parseoracle.Facts) []parenOp {
	var out []parenOp
	for _, o := range f.OpFlags {
		if o.Parens && !structuralParens[o.Op] && !o.Statement {
			out = append(out, parenOp{o.Line, o.Op, o.Var})
		}
	}
	return out
}

// sameParenOps reports whether a and b hold the same facts, in any order.
func sameParenOps(a, b []parenOp) bool {
	if len(a) != len(b) {
		return false
	}
	count := map[parenOp]int{}
	for _, o := range a {
		count[o]++
	}
	for _, o := range b {
		if count[o]--; count[o] < 0 {
			return false
		}
	}
	return true
}

// parenTable maps an operator node's Text to the op perl builds for it,
// measured on perl 5.44.0 one fixture at a time (K8, chain-review round 7).
// A node kind or operator missing here is unmapped, and fails its fixture.
// perl keeps a paren around a logical operator or `?:` on the null over
// its logop: `($x && 1)` is `null sKP/1` over `and sK/1`.
var parenTable = map[string]string{
	"==": "eq", "!=": "ne", "<": "lt", ">": "gt",
	"eq": "seq", "ne": "sne",
	"+": "add", "-": "subtract", "*": "multiply", ".": "concat",
	"not": "not", "neg": "negate",
	"&&": "null", "||": "null", "//": "null", "and": "null", "or": "null", "?:": "null",
}

// parenCalls maps a call psc gave Paren to its op, measured on perl 5.44.0
// as parenTable is: `((stat "/"))[2]` is `stat lKP/1`, `((caller))[2]`
// `caller lP`, `(keys %h)` `keys sKP/1`, `((sort @a))` `sort lKP`, and
// `(do {1})` `null sKP*/1` over the block's scope.
var parenCalls = map[string]string{
	"stat": "stat", "caller": "caller", "keys": "keys", "sort": "sort", "do": "null",
}

// parenListOps maps the node a parenthesised List stands directly in to the
// list op that absorbed it, measured on perl 5.44.0: `print((1,2))` is
// `print vKP`, `[(1,2)]` `anonlist sKP*/1` and `{(a=>1)}` `anonhash sKP*/1`.
var parenListOps = map[parse.Kind]string{
	parse.AnonArray: "anonlist", parse.AnonHash: "anonhash",
}

// parenOpOf is the op perl builds for n, a node psc gave Paren, through the
// fixed node-kind→op-name table; up is n's ancestors, its parent last.
//   - A scalar variable is `padsv` when a `my` or `state` declared it and
//     `ex-rv2sv` otherwise; an array `padav` or `rv2av`. `our ($p)` is
//     `rv2sv vKP/OURINTR`, and `($x)->[0]` `ex-padsv sPM/DREFAV`.
//   - A constant is `const`; `qw(...)` is `ex-list`, or `const` for one word.
//   - A List of two or more elements is `ex-list`, or the list op that
//     absorbed it: `return (1,2)` is `<@> return KP`, and see parenListOps.
//     `my ($x, $y);` is `list vKP`, `() = f()` `stub lPRM*` and the value
//     of `do {(1,2)}` `list sKP`.
//   - A one-element `(x)` is x itself in psc's tree, so it maps to x's op:
//     `my ($x) = @_` puts P on `padsv`, not on the list.
//   - `($x = 1)` is `sassign sKPS/2`, a list assignment `aassign`;
//     `($a[0])` is `ex-aelem sKP/2` and `($x->{a})` `ex-helem sKP/2`.
func parenOpOf(n *parse.Node, up []*parse.Node, lexical map[string]bool) (string, bool) {
	var parent *parse.Node
	if len(up) > 0 {
		parent = up[len(up)-1]
	}
	isHead := func(decl string) bool {
		return parent != nil && parent.Kind == parse.Declaration && parent.Text == decl && parent.Children[0] == n
	}
	switch n.Kind {
	case parse.Term:
		switch {
		case isQw(n):
			if len(strings.Fields(n.Text[3:len(n.Text)-1])) == 1 {
				return "const", true
			}
			return "ex-list", true
		case strings.HasPrefix(n.Text, "$") && len(n.Text) > 1:
			switch {
			case isHead("our"):
				return "rv2sv", true
			case lexical[n.Text] && parent != nil && parent.Kind == parse.Index && parent.Arrow && parent.Children[0] == n:
				return "ex-padsv", true
			case lexical[n.Text]:
				return "padsv", true
			}
			return "ex-rv2sv", true
		case strings.HasPrefix(n.Text, "@") && len(n.Text) > 1:
			if lexical[n.Text] {
				return "padav", true
			}
			return "rv2av", true
		case isConstant(n.Text):
			return "const", true
		}
	case parse.Binary:
		if n.Text == "=" && len(n.Children) == 2 {
			if l := n.Children[0]; l.Kind == parse.Term && strings.HasPrefix(l.Text, "$") && !l.Paren {
				return "sassign", true
			}
			return "aassign", true
		}
		op, ok := parenTable[n.Text]
		return op, ok
	case parse.CmpChain, parse.Unary, parse.Ternary:
		op, ok := parenTable[n.Text]
		return op, ok
	case parse.List:
		switch {
		case parent == nil:
		case parent.Kind == parse.LoopControl && parent.Text == "return":
			return "return", true
		case parent.Kind == parse.Call && parent.Text == "print":
			return "print", true
		case parenListOps[parent.Kind] != "":
			return parenListOps[parent.Kind], true
		case isHead("my") || isHead("state"):
			return "list", true
		case len(n.Children) == 0 && parent.Kind == parse.Binary && parent.Text == "=" && parent.Children[0] == n:
			return "stub", true
		case len(up) >= 3 && parent.Kind == parse.Statement && up[len(up)-2].Kind == parse.Block &&
			up[len(up)-3].Kind == parse.Call && up[len(up)-3].Text == "do":
			return "list", true
		}
		return "ex-list", true
	case parse.Call:
		op, ok := parenCalls[n.Text]
		return op, ok
	case parse.Index:
		switch {
		case n.Text == "[" && len(n.Children) == 2:
			return "ex-aelem", true
		case n.Text == "{" && len(n.Children) == 2:
			return "ex-helem", true
		}
	}
	return "", false
}

// varOf is the variable n reads, for an op the walk names one for.
func varOf(n *parse.Node, op string) string {
	switch op {
	case "padsv", "padav", "ex-rv2sv", "rv2sv", "rv2av":
		return n.Text
	}
	return ""
}

// isQw reports whether n is a `qw` list. perl's toke.c builds one as a
// parenthesised list whatever its delimiters, so it carries P with no
// Paren in psc's tree: measured on perl 5.44.0, `qw(a b)` and `qw/a b/`
// are `ex-list lKP`, `qw(a)` is `const[PV "a"] sP`, and `(qw(a b))` is
// one `ex-list lKP`, the bit being boolean.
func isQw(n *parse.Node) bool {
	return n.Kind == parse.Term && strings.HasPrefix(n.Text, "qw") && len(n.Text) >= 4 && len(n.Children) == 0
}

// isConstant reports whether text is a numeric or quoted string literal.
func isConstant(text string) bool {
	if text == "" {
		return false
	}
	if c := text[0]; c == '"' || c == '\'' || c >= '0' && c <= '9' {
		return true
	}
	return false
}

// pscParens is psc's side: the op every Paren node maps to, on the first
// line of the statement that holds it, and a problem for each Paren node
// whose kind the table does not map.
func pscParens(src []byte, root *parse.Node) (ops []parenOp, unmapped []string) {
	lexical := declaredLexicals(root)
	var up []*parse.Node
	var walk func(n *parse.Node, stmt int)
	walk = func(n *parse.Node, stmt int) {
		if n.Kind == parse.Statement {
			stmt = lineOf(src, n.Start)
		}
		if n.Paren || isQw(n) {
			if op, ok := parenOpOf(n, up, lexical); ok {
				ops = append(ops, parenOp{stmt, op, varOf(n, op)})
			} else {
				unmapped = append(unmapped, fmt.Sprintf("line %d: unmapped: psc's Paren on a %s %q has no row in the node-kind→op table",
					stmt, n.Kind, src[n.Start:n.End]))
			}
		}
		up = append(up, n)
		for _, c := range n.Children {
			walk(c, stmt)
		}
		up = up[:len(up)-1]
	}
	walk(root, 1)
	return ops, unmapped
}

// declaredLexicals is every variable a `my` or `state` declares anywhere in
// the tree: `my $x`, `my ($a, $b) = ...` and `my $v = ...` alike.
func declaredLexicals(root *parse.Node) map[string]bool {
	out := map[string]bool{}
	var vars func(n *parse.Node)
	vars = func(n *parse.Node) {
		if n.Kind == parse.Term && len(n.Text) > 1 && strings.ContainsRune("$@%", rune(n.Text[0])) {
			out[n.Text] = true
		}
		if n.Kind == parse.List {
			for _, c := range n.Children {
				vars(c)
			}
		}
	}
	var walk func(n *parse.Node)
	walk = func(n *parse.Node) {
		if n.Kind == parse.Declaration && (n.Text == "my" || n.Text == "state") && len(n.Children) > 0 {
			head := n.Children[0]
			if head.Kind == parse.Binary && head.Text == "=" && len(head.Children) > 0 {
				head = head.Children[0]
			}
			vars(head)
		}
		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(root)
	return out
}

// lineOf is the 1-based line of byte offset at in src.
func lineOf(src []byte, at int) int {
	return 1 + bytes.Count(src[:at], []byte("\n"))
}

// compareParens compares perl's P ops with psc's, statement by statement.
// On each line the ops both sides name cancel; then each leftover pair is a
// paren psc put on the wrong node (misattached), a leftover of perl's is a
// paren psc dropped, and a leftover of psc's is a paren psc invented --
// unless gaps lists it: an op perl folded away and so keeps no bit for, or
// one perl's fold gave a bit no source paren stands for.
//
// The match is by op name, and variable name for an op that reads one,
// per statement line -- not by the op's position in the tree. Two P ops
// of one name on one line that read no variable, `(1) + 2` against `1 +
// (2)`, are not told apart: a Paren moved between them passes.
//
// Zero on both sides is agreement only when the fixture says so (none):
// a source with a `(` in it whose two sides are both empty may be a walk
// that read nothing, and is never taken for a match by default.
func compareParens(src string, perl, psc, gaps []parenOp, none bool) []string {
	if len(perl) == 0 && len(psc) == 0 {
		if strings.Contains(src, "(") && !none {
			return []string{"neither perl nor psc reports a paren, though the source has a `(`: record `none` once perl is measured to keep no P bit"}
		}
		return nil
	}
	type sides struct{ perl, psc []string }
	lines := map[int]*sides{}
	at := func(line int) *sides {
		if lines[line] == nil {
			lines[line] = &sides{}
		}
		return lines[line]
	}
	for _, o := range perl {
		at(o.Line).perl = append(at(o.Line).perl, o.label())
	}
	for _, o := range psc {
		at(o.Line).psc = append(at(o.Line).psc, o.label())
	}
	excused := map[parenOp]int{}
	for _, g := range gaps {
		excused[parenOp{g.Line, g.label(), ""}]++
	}
	excuse := func(line int, ops []string) []string {
		var kept []string
		for _, op := range ops {
			if k := (parenOp{line, op, ""}); excused[k] > 0 {
				excused[k]--
				continue
			}
			kept = append(kept, op)
		}
		return kept
	}
	var nums []int
	for line := range lines {
		nums = append(nums, line)
	}
	sort.Ints(nums)
	var problems []string
	for _, line := range nums {
		s := lines[line]
		left, right := cancel(s.perl, s.psc)
		invented := excuse(line, right)
		left = excuse(line, left)
		for len(left) > 0 && len(invented) > 0 {
			problems = append(problems, fmt.Sprintf("line %d: misattached: perl puts P on %s, psc puts Paren on %s", line, left[0], invented[0]))
			left, invented = left[1:], invented[1:]
		}
		for _, op := range left {
			problems = append(problems, fmt.Sprintf("line %d: dropped: perl puts P on %s, psc has no Paren there", line, op))
		}
		for _, op := range invented {
			problems = append(problems, fmt.Sprintf("line %d: invented: psc puts Paren on %s, perl keeps no P there", line, op))
		}
	}
	return problems
}

// cancel removes the names a and b share, one for one, and returns what is
// left of each, sorted.
func cancel(a, b []string) (left, right []string) {
	count := map[string]int{}
	for _, op := range b {
		count[op]++
	}
	for _, op := range a {
		if count[op] > 0 {
			count[op]--
			continue
		}
		left = append(left, op)
	}
	for op, k := range count {
		for ; k > 0; k-- {
			right = append(right, op)
		}
	}
	sort.Strings(left)
	sort.Strings(right)
	return left, right
}

// findNode is the first node of root, depth first, that match accepts.
func findNode(root *parse.Node, match func(*parse.Node) bool) *parse.Node {
	if match(root) {
		return root
	}
	for _, c := range root.Children {
		if n := findNode(c, match); n != nil {
			return n
		}
	}
	return nil
}

// pscCompare parses src, lets edit change the tree, and compares psc's
// Paren nodes with perl's P ops.
func pscCompare(t *testing.T, src string, edit func(root *parse.Node), perl, gaps []parenOp, none bool) []string {
	t.Helper()
	root := parse.Parse([]byte(src))
	if edit != nil {
		edit(root)
	}
	psc, unmapped := pscParens([]byte(src), root)
	if len(unmapped) > 0 {
		t.Fatalf("unmapped: %q", unmapped)
	}
	return compareParens(src, perl, psc, gaps, none)
}

// hasProblem reports whether one of problems contains want.
func hasProblem(problems []string, want string) bool {
	for _, p := range problems {
		if strings.Contains(p, want) {
			return true
		}
	}
	return false
}

// TestParenComparisonReportsMisattachedParen: perl's facts are hand-built
// from perl 5.44.0's Concise, `ok(($y eq "x"))` is `seq sKPM`. psc's tree
// with Paren on the comparison matches; moved onto $y, it is misattached.
func TestParenComparisonReportsMisattachedParen(t *testing.T) {
	src := "sub ok {}\nour $y;\nok(($y eq \"x\"));\n"
	perl := []parenOp{{3, "seq", ""}}
	if got := pscCompare(t, src, nil, perl, nil, false); len(got) != 0 {
		t.Errorf("a matching pair reports %q, want nothing", got)
	}
	moved := func(root *parse.Node) {
		cmp := findNode(root, func(n *parse.Node) bool { return n.Paren && n.Text == "eq" })
		if cmp == nil {
			t.Fatal("no parenthesised eq in psc's tree")
		}
		cmp.Paren, cmp.Children[0].Paren = false, true
	}
	got := pscCompare(t, src, moved, perl, nil, false)
	if want := "line 3: misattached: perl puts P on seq, psc puts Paren on ex-rv2sv $y"; len(got) != 1 || got[0] != want {
		t.Errorf("Paren moved onto $y reports %q, want exactly %q", got, want)
	}
}

// TestParenComparisonReportsDroppedParen: `my $z = ($x)` is `padsv sP` on
// perl 5.44.0; psc's tree without Paren on $x has dropped it.
func TestParenComparisonReportsDroppedParen(t *testing.T) {
	src := "my $x;\nmy $z = ($x);\n"
	perl := []parenOp{{2, "padsv", "$x"}}
	if got := pscCompare(t, src, nil, perl, nil, false); len(got) != 0 {
		t.Errorf("psc's own tree reports %q, want nothing", got)
	}
	drop := func(root *parse.Node) {
		findNode(root, func(n *parse.Node) bool { return n.Paren }).Paren = false
	}
	got := pscCompare(t, src, drop, perl, nil, false)
	if want := "line 2: dropped: perl puts P on padsv $x, psc has no Paren there"; len(got) != 1 || got[0] != want {
		t.Errorf("a dropped Paren reports %q, want exactly %q", got, want)
	}
}

// TestParenComparisonReportsInventedParen: `print (1)+2` is `const[IV 1]
// s` on perl 5.44.0, no P: print's pair is its call parens. A Paren psc
// put on the 1 is invented. `$h{(1)}` keeps no P either, but its const is
// folded into a multideref, a gap, and is not reported.
func TestParenComparisonReportsInventedParen(t *testing.T) {
	src := "print (1)+2;\n"
	invent := func(root *parse.Node) {
		findNode(root, func(n *parse.Node) bool { return n.Kind == parse.Term && n.Text == "1" }).Paren = true
	}
	got := pscCompare(t, src, invent, nil, nil, true)
	if want := "line 1: invented: psc puts Paren on const, perl keeps no P there"; len(got) != 1 || got[0] != want {
		t.Errorf("an invented Paren reports %q, want exactly %q", got, want)
	}
	gap := "my %h;\nmy $v = $h{(1)};\n"
	if got := pscCompare(t, gap, nil, nil, []parenOp{{2, "const", ""}}, true); len(got) != 0 {
		t.Errorf("a gaps.txt fold reports %q, want nothing", got)
	}
}

// TestParenComparisonRejectsEmptyFactsForParenthesisedSource: zero P ops on
// both sides of a source with a paren is not a match unless the fixture
// records that perl keeps none.
func TestParenComparisonRejectsEmptyFactsForParenthesisedSource(t *testing.T) {
	src := "sub f {}\nf(1,2);\n"
	if got := pscCompare(t, src, nil, nil, nil, false); !hasProblem(got, "neither perl nor psc reports a paren") {
		t.Errorf("0 = 0 for a parenthesised source reports %q, want a failure", got)
	}
	if got := pscCompare(t, src, nil, nil, nil, true); len(got) != 0 {
		t.Errorf("a fixture recording none reports %q, want nothing", got)
	}
}

// TestParenUnmatchedOpsAreNotAgreement: `-(1)` folds to `const[IV -1]
// s/FOLD` on perl 5.44.0, so perl has no P op at all. psc's Paren on the 1
// is then unmatched, and that is an invented paren unless gaps.txt lists
// the fold.
func TestParenUnmatchedOpsAreNotAgreement(t *testing.T) {
	src := "my $v = -(1);\n"
	if got := pscCompare(t, src, nil, nil, nil, true); !hasProblem(got, "invented: psc puts Paren on const") {
		t.Errorf("an unmatched Paren reports %q, want it invented", got)
	}
	if got := pscCompare(t, src, nil, nil, []parenOp{{1, "const", ""}}, true); len(got) != 0 {
		t.Errorf("an unmatched Paren listed as a gap reports %q, want nothing", got)
	}
}

// TestParenOneElementListMapsToElement: `my ($x) = @_` is, on perl 5.44.0,
// `padsv[$x:1,2] sPRM*/LVINTRO` under an `ex-list lK` with no P. psc's
// tree holds `($x)` as $x with Paren, which maps to padsv and matches; a
// table that mapped it to ex-list would be misattached.
func TestParenOneElementListMapsToElement(t *testing.T) {
	src := "my ($x) = @_;\n"
	perl := []parenOp{{1, "padsv", "$x"}}
	root := parse.Parse([]byte(src))
	psc, unmapped := pscParens([]byte(src), root)
	if len(unmapped) > 0 || !sameParenOps(psc, perl) {
		t.Fatalf("psc's side is %v (unmapped %q), want %v", psc, unmapped, perl)
	}
	if got := compareParens(src, perl, psc, nil, false); len(got) != 0 {
		t.Errorf("my ($x) reports %q, want nothing", got)
	}
	if got := compareParens(src, perl, []parenOp{{1, "ex-list", ""}}, nil, false); !hasProblem(got, "misattached: perl puts P on padsv $x, psc puts Paren on ex-list") {
		t.Errorf("an ex-list mapping reports %q, want a misattachment", got)
	}
}

// TestParenUnmappedNodeKindFails: a Paren on a node kind the table has no
// row for is unmapped, never agreement. `([1])` is an anonymous array.
func TestParenUnmappedNodeKindFails(t *testing.T) {
	src := "my $r = ([1]);\n"
	_, unmapped := pscParens([]byte(src), parse.Parse([]byte(src)))
	if !hasProblem(unmapped, "unmapped") {
		t.Errorf("a Paren on an anon_array reports %q, want it unmapped", unmapped)
	}
}

const parenDir = "testdata/parens"

// parenFixture is one snippet of fixtures.txt or gaps.txt.
type parenFixture struct {
	Name     string
	Classes  []string
	Src      string
	Perl     []parenOp // the ops perl marks P, as recorded
	None     bool      // recorded: perl marks no op P
	Gaps     []parenOp // gaps.txt only: ops perl folded away
	PerlGaps []parenOp // gaps.txt only: P bits a fold gave, no source paren
	Evidence []string  // the `#` lines: perl's Concise
}

// loadParenFixtures reads a fixture file: `=== NAME [CLASS ...]`, the
// source, `--- perl` with `LINE OP [VAR]` rows or `none`, and in gaps.txt
// a `--- gap` or `--- perl-gap` section of `LINE OP` rows. Text before the
// first `===` is commentary.
func loadParenFixtures(path string) ([]parenFixture, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []parenFixture
	var fx *parenFixture
	section := ""
	for _, line := range strings.Split(string(data), "\n") {
		switch {
		case strings.HasPrefix(line, "=== "):
			out = append(out, parenFixture{})
			fx, section = &out[len(out)-1], "src"
			fx.Name = strings.TrimPrefix(line, "=== ")
			if i := strings.Index(fx.Name, " ["); i >= 0 && strings.HasSuffix(fx.Name, "]") {
				fx.Classes = strings.Fields(fx.Name[i+2 : len(fx.Name)-1])
				fx.Name = fx.Name[:i]
			}
		case fx == nil:
		case strings.HasPrefix(line, "--- "):
			section = strings.TrimPrefix(line, "--- ")
			if section != "perl" && section != "gap" && section != "perl-gap" {
				return nil, fmt.Errorf("%s: %s: unknown section %q", path, fx.Name, line)
			}
		case section == "src":
			fx.Src += line + "\n"
		case strings.HasPrefix(line, "#"):
			fx.Evidence = append(fx.Evidence, line)
		case strings.TrimSpace(line) == "":
		case section == "perl" && line == "none":
			fx.None = true
		default:
			var o parenOp
			f := strings.Fields(line)
			line, err := strconv.Atoi(f[0])
			if err != nil || len(f) < 2 || len(f) > 3 {
				return nil, fmt.Errorf("%s: %s: bad row %q", path, fx.Name, strings.Join(f, " "))
			}
			o.Line, o.Op = line, f[1]
			if len(f) == 3 {
				o.Var = f[2]
			}
			switch section {
			case "gap":
				fx.Gaps = append(fx.Gaps, o)
			case "perl-gap":
				fx.PerlGaps = append(fx.PerlGaps, o)
			default:
				fx.Perl = append(fx.Perl, o)
			}
		}
	}
	for _, fx := range out {
		if fx.None == (len(fx.Perl) > 0) {
			return nil, fmt.Errorf("%s: %s: `--- perl` must list P ops or say none, not both or neither", path, fx.Name)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("%s holds no fixture", path)
	}
	return out, nil
}

// loadAllParenFixtures is fixtures.txt and gaps.txt; only gaps.txt rows
// may excuse an op.
func loadAllParenFixtures(t *testing.T) []parenFixture {
	t.Helper()
	fixtures, err := loadParenFixtures(filepath.Join(parenDir, "fixtures.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fx := range fixtures {
		if len(fx.Gaps)+len(fx.PerlGaps) > 0 {
			t.Fatalf("fixtures.txt: %s has a gap section; gaps belong in gaps.txt", fx.Name)
		}
	}
	gaps, err := loadParenFixtures(filepath.Join(parenDir, "gaps.txt"))
	if err != nil {
		t.Fatal(err)
	}
	return append(fixtures, gaps...)
}

// askParens runs perl on src and returns its facts, failing with perl's
// stderr when it compiled nothing.
func askParens(perl, src string) (parseoracle.Facts, error) {
	facts, err := parseoracle.Ask(context.Background(), []byte(src), parseoracle.Options{Perl: perl})
	if err != nil {
		return facts, err
	}
	if !facts.OK || len(facts.OpFlags) == 0 {
		return facts, fmt.Errorf("perl compiled no ops for the snippet: %s", facts.Stderr)
	}
	return facts, nil
}

// checkParenFixture holds fx to perl and psc: perl must still mark P on
// the ops the fixture records, and psc's Paren nodes must match them.
func checkParenFixture(perl string, fx parenFixture) ([]string, error) {
	facts, err := askParens(perl, fx.Src)
	if err != nil {
		return nil, err
	}
	var problems []string
	got := perlParens(facts)
	if !sameParenOps(got, fx.Perl) {
		problems = append(problems, fmt.Sprintf("the fixture records P on %v, perl reports %v", fx.Perl, got))
	}
	psc, unmapped := pscParens([]byte(fx.Src), parse.Parse([]byte(fx.Src)))
	problems = append(problems, unmapped...)
	return append(problems, compareParens(fx.Src, got, psc, append(fx.Gaps, fx.PerlGaps...), fx.None)...), nil
}

// TestParenFlagMatchesPerlOPfPARENS: on every fixture, psc's Paren flag
// matches perl's OPf_PARENS bit op by op.
func TestParenFlagMatchesPerlOPfPARENS(t *testing.T) {
	p := perltest.PinnedPerl(t)
	for _, fx := range loadAllParenFixtures(t) {
		problems, err := checkParenFixture(p.Path, fx)
		if err != nil {
			t.Errorf("%s: %v", fx.Name, err)
		}
		for _, problem := range problems {
			t.Errorf("%s: %s", fx.Name, problem)
		}
	}
}

// parenClasses are the measured classes the fixture set must cover, each
// by at least one fixture, so the check cannot pass on an empty or
// lopsided set.
var parenClasses = []string{
	"list-op", "one-element-list", "scalar-op", "doubled", "call-arg",
	"ex-list", "qw", "absorbed", "block", "loop", "single-form", "logical",
}

// classProblems names each class in parenClasses no fixture carries.
func classProblems(fixtures []parenFixture) []string {
	have := map[string]bool{}
	for _, fx := range fixtures {
		for _, c := range fx.Classes {
			have[c] = true
		}
	}
	var problems []string
	for _, c := range parenClasses {
		if !have[c] {
			problems = append(problems, fmt.Sprintf("no fixture of class %s", c))
		}
	}
	return problems
}

// TestParenFixturesCoverEveryClass: every measured class has a fixture,
// and a set missing one is reported.
func TestParenFixturesCoverEveryClass(t *testing.T) {
	for _, problem := range classProblems(loadAllParenFixtures(t)) {
		t.Error(problem)
	}
	if got := classProblems([]parenFixture{{Classes: []string{"qw"}}}); !hasProblem(got, "no fixture of class list-op") {
		t.Errorf("a set with only a qw fixture reports %q, want list-op missing", got)
	}
}

// gapProblems re-checks a gap row against perl's facts for its snippet,
// not against its pasted text: the op psc's Paren maps to must be absent
// from the statement's ops, or one of that name there must carry FOLD, or
// every one of that name must be a nulled op without P on a line where a
// multideref absorbed it -- `($_[0])->[0]`'s `ex-aelem sK/2` under
// `multideref($_[0][0])`.
func gapProblems(facts parseoracle.Facts, gap parenOp) []string {
	present, absorbed, multideref := false, true, false
	for _, o := range facts.OpFlags {
		if o.Line == gap.Line && o.Op == "multideref" {
			multideref = true
		}
		if o.Line == gap.Line && o.Op == gap.Op {
			if o.Folded {
				return nil
			}
			present = true
			absorbed = absorbed && strings.HasPrefix(o.Op, "ex-") && !o.Parens
		}
	}
	if present && !(absorbed && multideref) {
		return []string{fmt.Sprintf("line %d: perl still emits %s, present and unfolded: not a fold, so not a gap", gap.Line, gap.Op)}
	}
	return nil
}

// TestParenFoldedOpsAreGaps: each gaps.txt row is re-checked against
// perl, is reported neither invented nor a defect, and re-emits
// faithfully, canon writing the paren.
func TestParenFoldedOpsAreGaps(t *testing.T) {
	p := perltest.PinnedPerl(t)
	gaps, err := loadParenFixtures(filepath.Join(parenDir, "gaps.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fx := range gaps {
		if len(fx.Gaps)+len(fx.PerlGaps) == 0 {
			t.Errorf("%s: a gaps.txt row with no `--- gap` or `--- perl-gap` section", fx.Name)
		}
		facts, err := askParens(p.Path, fx.Src)
		if err != nil {
			t.Errorf("%s: %v", fx.Name, err)
			continue
		}
		for _, gap := range fx.Gaps {
			for _, problem := range gapProblems(facts, gap) {
				t.Errorf("%s: %s", fx.Name, problem)
			}
		}
		for _, gap := range fx.PerlGaps {
			for _, problem := range perlGapProblems(facts, gap) {
				t.Errorf("%s: %s", fx.Name, problem)
			}
		}
		problems, err := checkParenFixture(p.Path, fx)
		if err != nil {
			t.Errorf("%s: %v", fx.Name, err)
		}
		for _, problem := range problems {
			t.Errorf("%s: %s", fx.Name, problem)
		}
		if ok, why := parse.Faithful(parse.Parse([]byte(fx.Src)), []byte(fx.Src)); !ok {
			t.Errorf("%s: canon does not re-emit the paren: %s", fx.Name, why)
		}
	}
}

// perlGapProblems re-checks a perl-gap row: perl must still mark the op P
// on that line, or the row excuses nothing and is stale.
func perlGapProblems(facts parseoracle.Facts, gap parenOp) []string {
	for _, o := range perlParens(facts) {
		if o.Line == gap.Line && o.label() == gap.label() {
			return nil
		}
	}
	return []string{fmt.Sprintf("line %d: perl no longer marks %s P: the perl-gap row is stale", gap.Line, gap.label())}
}

// TestParenGapRowWithUnfoldedOpFails: `print (1)+2` is `const[IV 1] s` on
// perl 5.44.0, present and unfolded, so a gap row for it fails the
// re-check rather than excusing an invented paren.
func TestParenGapRowWithUnfoldedOpFails(t *testing.T) {
	p := perltest.PinnedPerl(t)
	facts, err := askParens(p.Path, "print (1)+2;\n")
	if err != nil {
		t.Fatal(err)
	}
	if got := gapProblems(facts, parenOp{1, "const", ""}); !hasProblem(got, "present and unfolded") {
		t.Errorf("a gap row for an unfolded const reports %q, want it rejected", got)
	}
}

// TestParenKnownDefectsEmpty: defects.txt holds no row, and every gaps.txt
// row quotes perl's Concise as evidence of the fold.
func TestParenKnownDefectsEmpty(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(parenDir, "defects.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if line = strings.TrimSpace(line); line != "" && !strings.HasPrefix(line, "#") {
			t.Errorf("defects.txt holds a known defect: %s", line)
		}
	}
	gaps, err := loadParenFixtures(filepath.Join(parenDir, "gaps.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, fx := range gaps {
		quoted := false
		for _, e := range fx.Evidence {
			quoted = quoted || concise.MatchString(e)
		}
		if !quoted {
			t.Errorf("gaps.txt: %s quotes no Concise line as evidence of the fold", fx.Name)
		}
	}
}

// concise matches a quoted B::Concise op line: `# 5  <$> const[IV 1] s`.
var concise = regexp.MustCompile(`^#\s+\S+\s+<.>\s+\S`)

// TestParenStructuralOpsAreExcluded: perl marks P on the leave of every
// block and the leaveloop and lineseq of every loop with no source paren
// behind them. The walk reports them, and the check excludes them, so the
// block and loop fixtures report no dropped paren.
func TestParenStructuralOpsAreExcluded(t *testing.T) {
	p := perltest.PinnedPerl(t)
	src := "our $x;\nif ($x) {\n  my $y = 1;\n  $y++;\n}\nfor my $i (1, 2) {\n  my $y = 2;\n  $y++;\n}\n"
	facts, err := askParens(p.Path, src)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, o := range facts.OpFlags {
		if o.Parens {
			seen[o.Op] = true
		}
	}
	for _, op := range []string{"leave", "leaveloop", "lineseq", "null"} {
		if !seen[op] {
			t.Errorf("the walk reports no %s with P; the exclusion is untested", op)
		}
	}
	if got := perlParens(facts); len(got) != 0 {
		t.Errorf("perl's side keeps %v, want the structural ops excluded", got)
	}
	psc, _ := pscParens([]byte(src), parse.Parse([]byte(src)))
	if got := compareParens(src, perlParens(facts), psc, nil, true); len(got) != 0 {
		t.Errorf("a block and a loop report %q, want nothing", got)
	}
}

// TestParenFlagCallArgParensAreNotAParen: a call's own parens are not a
// paren on its argument. perl 5.44.0 measures `print ((1))+2` as
// `const[IV 1] sP` and `print (1)+2` as `const[IV 1] s`; psc's Paren
// matches each.
func TestParenFlagCallArgParensAreNotAParen(t *testing.T) {
	p := perltest.PinnedPerl(t)
	for _, fx := range []parenFixture{
		{Name: "doubled", Src: "print ((1))+2;\n", Perl: []parenOp{{1, "const", ""}}},
		{Name: "single", Src: "print (1)+2;\n", None: true},
	} {
		problems, err := checkParenFixture(p.Path, fx)
		if err != nil {
			t.Fatalf("%s: %v", fx.Name, err)
		}
		for _, problem := range problems {
			t.Errorf("%s: %s", fx.Name, problem)
		}
	}
}

// TestParenFlagFailsWhenPerlProducesNoOps: a snippet perl cannot compile
// fails with perl's stderr, rather than passing as a fixture with no P.
func TestParenFlagFailsWhenPerlProducesNoOps(t *testing.T) {
	p := perltest.PinnedPerl(t)
	_, err := checkParenFixture(p.Path, parenFixture{Name: "broken", Src: "my $x = ;\n", None: true})
	if err == nil || !strings.Contains(err.Error(), "syntax error") {
		t.Errorf("a snippet perl rejects gives %v, want a failure carrying perl's syntax error", err)
	}
}

// TestParenWalkSurvivesLargeOutput: perl -c writes "syntax OK" to the
// stream the walk prints to. A file with enough ops to overflow a buffered
// stdout had that line land inside an op row -- `500 ent/tmp/oracle-
// sJWmnMnw.pl syntax OK` -- and the oracle errored on comp/proto.t,
// comp/require.t and comp/use.t.
func TestParenWalkSurvivesLargeOutput(t *testing.T) {
	p := perltest.PinnedPerl(t)
	// The BEGIN is comp/use.t's. Without one, measured, perl 5.44.0
	// flushed the walk's lines before "syntax OK" even when buffered.
	var src strings.Builder
	src.WriteString("BEGIN { $INC{\"feature.pm\"} = 1 }\n")
	for i := 0; i < 2000; i++ {
		fmt.Fprintf(&src, "my $x%d = (%d);\n", i, i)
	}
	facts, err := askParens(p.Path, src.String())
	if err != nil {
		t.Fatal(err)
	}
	if got := len(perlParens(facts)); got != 2000 {
		t.Errorf("perl's side has %d P ops, want 2000, one const per statement", got)
	}
	for _, o := range facts.OpFlags {
		if !opName.MatchString(o.Op) {
			t.Fatalf("line %d: op row %q is not an op name", o.Line, o.Op)
		}
	}
}

// opName matches an op name as the walk prints it: `const`, `ex-rv2sv`.
var opName = regexp.MustCompile(`^[a-z_][a-z0-9_-]*$`)

// TestParenLogicalNullIsNotStructural: perl keeps a paren around `&&`,
// `||`, `//`, `and`, `or`, `?:` and `do {}` on the null op over the
// logop, `null sKP/1` on perl 5.44.0 -- and in void context `($x && 1);`
// is `null vKP/1`, the flags of the null over an `if`. Only the `if`'s,
// whose logop carries OPpSTATEMENT (`and vK/STMT`), is structural. A
// Paren psc dropped from the `&&` is reported.
func TestParenLogicalNullIsNotStructural(t *testing.T) {
	p := perltest.PinnedPerl(t)
	src := "our $x;\nmy $v = ($x && 1) + ($x);\n($x && 1);\nif ($x) {\n  my $y = 1;\n  $y++;\n}\n"
	facts, err := askParens(p.Path, src)
	if err != nil {
		t.Fatal(err)
	}
	want := []parenOp{{2, "null", ""}, {2, "ex-rv2sv", "$x"}, {3, "null", ""}}
	if got := perlParens(facts); !sameParenOps(got, want) {
		t.Fatalf("perl's side is %v, want %v: only the if's null is structural", got, want)
	}
	drop := func(root *parse.Node) {
		findNode(root, func(n *parse.Node) bool { return n.Paren && n.Text == "&&" }).Paren = false
	}
	if got := pscCompare(t, src, drop, perlParens(facts), nil, false); !hasProblem(got, "line 2: dropped: perl puts P on null") {
		t.Errorf("a Paren dropped from && reports %q, want it dropped", got)
	}
}

// TestParenComparisonKeysVariables: a variable's op carries its name, so
// a Paren moved between two variables of one op name on one line is
// misattached. `my $z = ($x) + $y` is `padsv[$x] sP` on perl 5.44.0.
func TestParenComparisonKeysVariables(t *testing.T) {
	p := perltest.PinnedPerl(t)
	src := "my ($x, $y);\nmy $z = ($x) + $y;\n"
	facts, err := askParens(p.Path, src)
	if err != nil {
		t.Fatal(err)
	}
	perl := perlParens(facts)
	if want := []parenOp{{1, "list", ""}, {2, "padsv", "$x"}}; !sameParenOps(perl, want) {
		t.Fatalf("perl's side is %v, want %v", perl, want)
	}
	if got := pscCompare(t, src, nil, perl, nil, false); len(got) != 0 {
		t.Errorf("psc's own tree reports %q, want nothing", got)
	}
	moved := func(root *parse.Node) {
		x := findNode(root, func(n *parse.Node) bool { return n.Paren && n.Text == "$x" })
		y := findNode(root, func(n *parse.Node) bool { return n.Kind == parse.Term && n.Text == "$y" && n.Start > x.Start })
		x.Paren, y.Paren = false, true
	}
	if got := pscCompare(t, src, moved, perl, nil, false); !hasProblem(got, "misattached: perl puts P on padsv $x, psc puts Paren on padsv $y") {
		t.Errorf("a Paren moved from $x to $y reports %q, want a misattachment", got)
	}
}
