// ABOUTME: Fidelity compares the emission against the SOURCE, quotienting one known difference.
// ABOUTME: Stability compares canon against itself and cannot see a tree that misread its source.

package parse_test

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/perltest"
)

// TestCanonFidelityCatchesMisgrouping: the fidelity comparison fails on a
// tree that is not a parse of its source.
//
// This is the whole reason the check exists. The stability sweep
// (canonMatches) compares canon's output against itself, so the source is
// gone after the first emission and a consistently wrong tree is still a
// fixpoint -- measured:
//
//	WRONG tree: once="print((1 + 2) * 3);"
//	            twice="print((1 + 2) * 3);"   fixpoint
//
// Fidelity puts the source back on one side. The wrong tree emits a paren
// the source does not contain, and no quotient may forgive that.
func TestCanonFidelityCatchesMisgrouping(t *testing.T) {
	src := []byte("print (1+2)*3;")

	// The wrong tree: print swallowing the whole product. Same leaves, same
	// order -- round-trip and the stability sweep both accept it.
	mul := &parse.Node{Kind: parse.Binary, Text: "*", Start: 6, End: 13, Children: []*parse.Node{
		{Kind: parse.Binary, Text: "+", Start: 7, End: 10, Children: []*parse.Node{
			{Kind: parse.Term, Text: "1", Start: 7, End: 8},
			{Kind: parse.Term, Text: "2", Start: 9, End: 10},
		}},
		{Kind: parse.Term, Text: "3", Start: 12, End: 13},
	}}
	wrong := &parse.Node{Kind: parse.SourceFile, Start: 0, End: len(src), Children: []*parse.Node{
		{Kind: parse.Statement, Start: 0, End: len(src), Children: []*parse.Node{
			{Kind: parse.Call, Text: "print", Resolved: true, Start: 0, End: 13, Children: []*parse.Node{mul}},
		}},
	}}

	if ok, _ := parse.Faithful(wrong, src); ok {
		t.Errorf("Faithful accepted a tree that is not a parse of its source; "+
			"it emitted %q from %q", parse.Canon(wrong, src), string(src))
	}

	// And the correct parse of the same source must be accepted, or the
	// check is just "always false".
	if ok, why := parse.Faithful(parse.Parse(src), src); !ok {
		t.Errorf("Faithful rejected the correct parse of %q: %s", string(src), why)
	}
}

// TestCanonQuotientIsNarrow: only the parenless-call paren is forgiven.
//
// That equivalence is real and unavoidable -- `print "x"` and `print("x")`
// parse to the SAME TREE, correctly, so a tree-faithful emitter must print
// one form for both. 519 of 658 disagreeing T1 files differed by exactly
// this and nothing else.
//
// Every other difference must still fail, or the quotient has swallowed the
// property it was carved out of.
func TestCanonQuotientIsNarrow(t *testing.T) {
	// Forgiven: the call's own parens, present or absent in the source.
	for _, src := range []string{
		`print "x";`,
		`print("x");`,
		`ok $x, 'n';`,
		`ok($x, 'n');`,
	} {
		if ok, why := parse.Faithful(parse.Parse([]byte(src)), []byte(src)); !ok {
			t.Errorf("Faithful rejected %q, but a call's parens are the one "+
				"difference it must forgive: %s", src, why)
		}
	}

	// NOT forgiven: a paren that changes grouping. These sources differ from
	// each other only by parens, and each must be accepted against ITSELF
	// and rejected against the other's tree.
	for _, tc := range []struct{ mine, theirs string }{
		{`(1 + 2) * 3;`, `1 + 2 * 3;`},
		{`1 - (2 - 3);`, `1 - 2 - 3;`},
		{`(2 ** 3) ** 2;`, `2 ** 3 ** 2;`},
	} {
		if ok, why := parse.Faithful(parse.Parse([]byte(tc.mine)), []byte(tc.mine)); !ok {
			t.Errorf("Faithful rejected %q against its own parse: %s", tc.mine, why)
		}
		// The other source's tree against this source: a grouping paren is
		// missing or extra, which must not be forgiven.
		if ok, _ := parse.Faithful(parse.Parse([]byte(tc.theirs)), []byte(tc.mine)); ok {
			t.Errorf("Faithful accepted the tree of %q against the source %q; "+
				"those group differently", tc.theirs, tc.mine)
		}
	}
}

// TestCanonFidelityRatchet: the per-file fidelity result, failing in either
// direction.
//
// This is the criterion 01a0a8d8 was written for and could not carry: its
// corpus sweep compares canon against ITSELF, so a tree that misread its
// source passes. Here the source is on one side, and the only forgiven
// difference is the paren a parenless call must acquire.
//
// The count is 1 for a file whose emission does not say what its source
// says, 0 for one that does -- so "clean" in the header is the fidelity
// rate.
func TestCanonFidelityRatchet(t *testing.T) {
	dir, files := t1Files(t)

	now := make(map[string]int, len(files))
	for _, rel := range files {
		src, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			continue
		}
		now[rel] = 0
		if ok, _ := parse.Faithful(parse.Parse(src), src); !ok {
			now[rel] = 1
		}
	}

	checkRatchet(t, filepath.Join("testdata", "t1fidelity.ratchet"),
		"Files whose canon emission does NOT say what the source says (1 = differs).\n"+t1CorpusNote, now)
}

// sigTokens is text's significant tokens, the kinds Faithful drops left out.
func sigTokens(text string) []string {
	var out []string
	for _, tok := range lexer.Tokenize([]byte(text)) {
		if !slices.Contains(parse.DroppedKinds, tok.Kind) {
			out = append(out, text[tok.Start:tok.End])
		}
	}
	return out
}

func asCanon(toks []string) []parse.CanonToken {
	out := make([]parse.CanonToken, len(toks))
	for i, s := range toks {
		out[i] = parse.CanonToken{Text: s}
	}
	return out
}

func ruleNames(fs []parse.Forgiveness) []string {
	out := []string{}
	for _, f := range fs {
		out = append(out, f.Rule)
	}
	return out
}

// faithfulTokensCases are Faithful's verdicts on the examples above, as
// token sequences, each with the rule behind every difference it forgave.
// A case with parsed set takes its canon from the source's own parse, so
// Faithful's verdict on that parse is held to the same answer.
var faithfulTokensCases = []struct {
	name, src, canon string
	parsed           bool
	ok               bool
	rules            []string
}{
	{"misgrouped tree", "print (1+2)*3;", "print((1 + 2) * 3);", false, false, []string{}},
	{"correct parse", "print (1+2)*3;", "print(1 + 2) * 3;", true, true, []string{}},
	{"parenless call", `print "x";`, `print("x");`, true, true, []string{"call-parens"}},
	{"call with parens", `print("x");`, `print("x");`, true, true, []string{}},
	{"parenless ok", `ok $x, 'n';`, `ok $x, 'n';`, true, true, []string{}},
	{"grouping kept", `(1 + 2) * 3;`, `(1 + 2) * 3;`, true, true, []string{}},
	{"grouping lost", `(1 + 2) * 3;`, `1 + 2 * 3;`, false, false, []string{}},
	{"grouping added", `1 - 2 - 3;`, `1 - (2 - 3);`, false, false, []string{}},
	{"block without last semicolon", `sub f { 1 }`, `sub f {1;} `, true, true, []string{"last-semicolon"}},
	{"block with last semicolon", `sub f { 1; }`, `sub f {1;} `, true, true, []string{}},
	{"block argument", `print {$fh} "x";`, `print({$fh;} "x");`, true, true, []string{"call-parens", "last-semicolon"}},
	{"nested call parens", "package Foo; sub create { } package main; print create Foo sub { 1 };",
		"package Foo;sub create {} package main;print(create Foo(sub {1;}));", true, true,
		[]string{"call-parens", "call-parens", "last-semicolon"}},
}

// TestFaithfulTokens: the token-sequence comparator gives Faithful's
// verdicts and names the rule behind each difference it forgives.
func TestFaithfulTokens(t *testing.T) {
	for _, tc := range faithfulTokensCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.parsed {
				root := parse.Parse([]byte(tc.src))
				if got := parse.Canon(root, []byte(tc.src)); got != tc.canon {
					t.Fatalf("canon of %q is %q, the case says %q", tc.src, got, tc.canon)
				}
				if ok, why := parse.Faithful(root, []byte(tc.src)); ok != tc.ok {
					t.Errorf("Faithful(%q) = %v (%s), want %v", tc.src, ok, why, tc.ok)
				}
			}
			ok, forgiven, why := parse.FaithfulTokens(sigTokens(tc.src), asCanon(sigTokens(tc.canon)))
			if ok != tc.ok {
				t.Errorf("FaithfulTokens(%q, %q) = %v (%s), want %v", tc.src, tc.canon, ok, why, tc.ok)
			}
			if ok && !slices.Equal(ruleNames(forgiven), tc.rules) {
				t.Errorf("FaithfulTokens(%q, %q) forgave %v, want %v", tc.src, tc.canon, ruleNames(forgiven), tc.rules)
			}
		})
	}
}

// TestCanonMarksParenNodePairs: every `(`/`)` pair canon writes because a
// node's Paren flag is set carries ParenNode, and no other pair does. Each
// want is canon's significant tokens with a marked paren written `*(` or
// `*)`.
func TestCanonMarksParenNodePairs(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		// A parenthesised List writes its own pair.
		{`my ($x) = @_;`, `my *( $x *) = @_ ;`},
		// A parenthesised Term is written as its source text, parens and
		// all; the pair its Paren flag stands for is the outermost one.
		{`($x) x 3;`, `*( $x *) x 3 ;`},
		{`my $z = (($x));`, `my $z = *( ( $x ) *) ;`},
		// A call's argument list is not a Paren node's pair.
		{`print("x");`, `print ( "x" ) ;`},
		// Nor is a grouping canon writes from binding power.
		{`(1 + 2) * 3;`, `( 1 + 2 ) * 3 ;`},
	} {
		var got []string
		for _, tok := range parse.CanonTokens(parse.Parse([]byte(tc.src)), []byte(tc.src)) {
			if tok.ParenNode {
				got = append(got, "*"+tok.Text)
				continue
			}
			got = append(got, tok.Text)
		}
		if g := strings.Join(got, " "); g != tc.want {
			t.Errorf("canon tokens of %q are %q, want %q", tc.src, g, tc.want)
		}
	}
}

// forgivenRules is every rule faithfulTokens may forgive a difference by,
// each with a fixture: two programs that differ by exactly that rewrite,
// which the pinned perl runs to the same output.
var forgivenRules = map[string]struct{ src, canon string }{
	// `print "x"` and `print("x")` parse to the same tree.
	"call-parens": {`print "x";`, `print("x");`},
	// A block's last statement with and without its semicolon.
	"last-semicolon": {`sub f { 1 } print f();`, `sub f { 1; } print f();`},
}

type comparator func([]string, []parse.CanonToken) (bool, []parse.Forgiveness, string)

// unlistedRules is every rule compare forgave a pair by that forgivenRules
// does not list, as "rule (source)".
func unlistedRules(compare comparator, pairs [][2]string) []string {
	var out []string
	for _, p := range pairs {
		_, forgiven, _ := compare(sigTokens(p[0]), asCanon(sigTokens(p[1])))
		for _, f := range forgiven {
			if _, ok := forgivenRules[f.Rule]; !ok {
				out = append(out, fmt.Sprintf("%s (%q)", f.Rule, p[0]))
			}
		}
	}
	return out
}

// TestFaithfulForgivenessIsClosed: faithfulTokens forgives a difference by
// the listed rules and no other. Each rule's fixture is forgiven by that
// rule alone and measured equivalent by perl; the fixtures, the
// TestFaithfulTokens cases and every T1 file report no unlisted rule; and
// the token kinds dropped before comparing are exactly the trivia.
func TestFaithfulForgivenessIsClosed(t *testing.T) {
	var pairs [][2]string
	for rule, fx := range forgivenRules {
		pairs = append(pairs, [2]string{fx.src, fx.canon})
		ok, forgiven, why := parse.FaithfulTokens(sigTokens(fx.src), asCanon(sigTokens(fx.canon)))
		if !ok || !slices.Equal(ruleNames(forgiven), []string{rule}) {
			t.Errorf("%s: %q against %q gave %v %v (%s), want forgiven by %s alone",
				rule, fx.src, fx.canon, ok, ruleNames(forgiven), why, rule)
		}
	}
	for _, tc := range faithfulTokensCases {
		pairs = append(pairs, [2]string{tc.src, tc.canon})
	}
	for _, p := range unlistedRules(parse.FaithfulTokens, pairs) {
		t.Errorf("forgiven by a rule forgivenRules does not list: %s", p)
	}

	// Every T1 file, against its own parse's canon. A subtest, so a machine
	// without the corpus skips this half alone.
	t.Run("T1", func(t *testing.T) {
		dir, files := t1Files(t)
		for _, rel := range files {
			src, err := os.ReadFile(filepath.Join(dir, rel))
			if err != nil {
				continue
			}
			_, forgiven, _ := parse.FaithfulTokens(sigTokens(string(src)), parse.CanonTokens(parse.Parse(src), src))
			for _, f := range forgiven {
				if _, ok := forgivenRules[f.Rule]; !ok {
					t.Errorf("%s: forgiven by a rule forgivenRules does not list: %s", rel, f.Rule)
				}
			}
		}
	})

	want := []lexer.Kind{lexer.Whitespace, lexer.Comment, lexer.Pod, lexer.DataSection}
	if !slices.Equal(parse.DroppedKinds, want) {
		t.Errorf("Faithful drops token kinds %v before comparing, want exactly %v", parse.DroppedKinds, want)
	}

	t.Run("perl", func(t *testing.T) {
		perl := perltest.PinnedPerl(t)
		for rule, fx := range forgivenRules {
			a, errA := perl.Command("-e", fx.src).CombinedOutput()
			b, errB := perl.Command("-e", fx.canon).CombinedOutput()
			if errA != nil || errB != nil || string(a) != string(b) {
				t.Errorf("%s: perl runs %q to %q (%v) and %q to %q (%v); the rule forgives only an equivalence",
					rule, fx.src, a, errA, fx.canon, b, errB)
			}
		}
	})
}

// TestFaithfulForgivenessIsClosedRejectsUnlistedRule: a comparator that
// forgives by a rule the closed set does not list is caught.
func TestFaithfulForgivenessIsClosedRejectsUnlistedRule(t *testing.T) {
	invented := func([]string, []parse.CanonToken) (bool, []parse.Forgiveness, string) {
		return true, []parse.Forgiveness{{Rule: "invented"}}, ""
	}
	got := unlistedRules(invented, [][2]string{{`print "x";`, `print("x");`}})
	if want := []string{`invented ("print \"x\";")`}; !slices.Equal(got, want) {
		t.Errorf("unlistedRules = %q, want %q", got, want)
	}
}
