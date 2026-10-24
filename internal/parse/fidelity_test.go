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

// TestCanonQuotientIsNarrow: a call's parens are not forgiven, and need not
// be.
//
// `print "x"` and `print("x")` parse to the same call, and the tree records
// which spelling the source wrote (ArgParen), so canon writes it back and
// both are faithful with nothing forgiven.
//
// A paren that changes grouping must still fail, or the quotient has
// swallowed the property it was carved out of.
func TestCanonQuotientIsNarrow(t *testing.T) {
	// Faithful: the call's own parens, present or absent in the source.
	for _, src := range []string{
		`print "x";`,
		`print("x");`,
		`ok $x, 'n';`,
		`ok($x, 'n');`,
	} {
		if ok, why := parse.Faithful(parse.Parse([]byte(src)), []byte(src)); !ok {
			t.Errorf("Faithful rejected %q, but canon writes a call's parens "+
				"as the source spelled them: %s", src, why)
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
// difference is a block's last semicolon.
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
// A case with parsed set takes its canon from the source's own parse, marks
// and all, so Faithful's verdict on that parse is held to the same answer.
var faithfulTokensCases = []struct {
	name, src, canon string
	parsed           bool
	ok               bool
	rules            []string
}{
	{"misgrouped tree", "print (1+2)*3;", "print((1 + 2) * 3);", false, false, []string{}},
	{"correct parse", "print (1+2)*3;", "print(1 + 2) * 3;", true, true, []string{}},
	{"parenless call given parens", `print "x";`, `print("x");`, false, false, []string{}},
	{"parenless call as canon writes it", `print "x";`, `print "x";`, true, true, []string{}},
	{"call with parens", `print("x");`, `print("x");`, true, true, []string{}},
	{"parenless ok", `ok $x, 'n';`, `ok $x, 'n';`, true, true, []string{}},
	{"grouping kept", `(1 + 2) * 3;`, `(1 + 2) * 3;`, true, true, []string{}},
	{"grouping lost", `(1 + 2) * 3;`, `1 + 2 * 3;`, false, false, []string{}},
	{"grouping added", `1 - 2 - 3;`, `1 - (2 - 3);`, false, false, []string{}},
	{"block without last semicolon", `sub f { 1 }`, `sub f {1;} `, true, true, []string{"last-semicolon"}},
	{"block with last semicolon", `sub f { 1; }`, `sub f {1;} `, true, true, []string{}},
	{"block argument given parens", `print {$fh} "x";`, `print({$fh;} "x");`, false, false, []string{}},
	{"block argument as canon writes it", `print {$fh} "x";`, `print {$fh;} "x";`, true, true, []string{"last-semicolon"}},
	{"nested calls given parens", "package Foo; sub create { } package main; print create Foo sub { 1 };",
		"package Foo;sub create {} package main;print(create Foo(sub {1;}));", false, false,
		[]string{}},
	{"nested calls as canon writes them", "package Foo; sub create { } package main; print create Foo sub { 1 };",
		"package Foo;sub create {} package main;print create Foo sub {1;};", true, true,
		[]string{"last-semicolon"}},
}

// caseCanon is a faithfulTokensCases canon's tokens: its source's own
// parse's, marks and all, when parsed is set, and canon's text unmarked
// otherwise.
func caseCanon(src, canon string, parsed bool) []parse.CanonToken {
	if parsed {
		return parse.CanonTokens(parse.Parse([]byte(src)), []byte(src))
	}
	return asCanon(sigTokens(canon))
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
			ok, forgiven, why := parse.FaithfulTokens(sigTokens(tc.src), caseCanon(tc.src, tc.canon, tc.parsed))
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
		// A parenthesised Term is written with one pair, however deeply
		// the source nested it.
		{`($x) x 3;`, `*( $x *) x 3 ;`},
		{`my $z = (($x));`, `my $z = *( $x *) ;`},
		// A call's argument list is not a Paren node's pair.
		{`print("x");`, `print ( "x" ) ;`},
		// A grouping the source wrote is a Paren node's pair.
		{`(1 + 2) * 3;`, `*( 1 + 2 *) * 3 ;`},
		// One canon writes for binding power is not.
		{`- -$x;`, `- ( - $x ) ;`},
		// A heredoc whose statement runs on past the opener's line is split
		// around its body, and the marks on each side must survive the split.
		{"my ($s) = first(<<END, ($u)) eq\nhello\nEND\nq(hello\n);\n",
			"my *( $s *) = first ( <<END , *( $u *) ) eq q(hello\nhello\nEND\n) ;"},
		{"my $s = first(<<END, 'u') eq\nhello\nEND\nq(hello\n) . ($x);\n",
			"my $s = first ( <<END , 'u' ) eq q(hello\nhello\nEND\n) . *( $x *) ;"},
		// A Paren node inside a call's argument list: the inner pair is the
		// node's, the call's is not. Measured on perl 5.44.0,
		// `ok(($y eq "x"))` is `seq sKPM` and `ok($y eq "x")` is `seq sKM`.
		{`ok(($y eq "x"));`, `ok ( *( $y eq "x" *) ) ;`},
	} {
		if got := markedCanon(tc.src); got != tc.want {
			t.Errorf("canon tokens of %q are %q, want %q", tc.src, got, tc.want)
		}
	}
}

// markedCanon is canon's significant tokens of src's parse, joined by
// spaces, with a paren written for a node's Paren flag written `*(` or
// `*)`.
func markedCanon(src string) string {
	var got []string
	for _, tok := range parse.CanonTokens(parse.Parse([]byte(src)), []byte(src)) {
		if tok.ParenNode {
			got = append(got, "*"+tok.Text)
			continue
		}
		got = append(got, tok.Text)
	}
	return strings.Join(got, " ")
}

// forgivenRules is every rule faithfulTokens may forgive a difference by,
// each with a fixture: two programs that differ by exactly that rewrite,
// which the pinned perl runs to the same output and compiles to the same
// ops. The canon side is canon's own spelling, and its tokens are canon's,
// marks and all (canonTokensOf).
var forgivenRules = map[string]struct{ src, canon string }{
	// A block's last statement with and without its semicolon.
	"last-semicolon": {`sub f { 1 } print f();`, `sub f { 1; } print f();`},
	// A pair nested directly inside the pair a node's Paren flag stands
	// for. Measured on perl 5.44.0 with -MO=Concise, both build `seq sKPM`:
	// the OPf_PARENS bit is boolean.
	"extra-nesting": {`sub ok { print @_; } our $y = "x"; ok((($y eq "x")));`,
		`sub ok { print @_; } our $y = "x"; ok(($y eq "x"));`},
}

// extentNegatives are canons of wrong trees, each the tree of the reading
// perl does NOT make of its source, which faithfulTokens must reject with
// nothing forgiven. Each wrong tree is psc's parse of the wrong reading's
// own spelling, so its canon carries the marks a real tree's would. perl
// 5.44.0 builds different ops for each pair (parenMeasurements); with
// -MO=Deparse,-p:
//
//	defined $x || $y         (defined($x) or $y)
//	defined($x||$y)          defined(($x or $y))
//	not $x or 1              ($x and '???')
//	not($x or 1)             not(($x || 1))
//	my $x = @_               padsv_store   (Deparse writes both `(my($x) = @_)`)
//	my($x)=@_                aassign
//	local $x = 1             sassign       (`our $x` first)
//	local($x)=1              aassign
//	$y x 3                   const[IV 3] s
//	$y x(3)                  const[IV 3] sP
//	push @a, 1 or die        (push(@a, 1) or die)
//	push(@a, 1 or die)       refused: "Experimental push on scalar is now forbidden"
//	foo 1, bar 2, 3          foo(1, bar(2, 3))     (both subs declared)
//	foo(1, bar(2), 3)        foo(1, bar(2), 3)
//
// A user sub with a `($)` prototype the file cannot see -- declared in a
// module, or after the call -- is M4 territory and is not covered here.
var extentNegatives = []struct{ src, wrong string }{
	{`defined $x || $y;`, `defined($x||$y);`},
	{`not $x or 1;`, `not($x or 1);`},
	{`my $x = @_;`, `my($x)=@_;`},
	{`our $x; local $x = 1;`, `our $x; local($x)=1;`},
	{`$y x 3;`, `$y x(3);`},
	{`my @a; push @a, 1 or die;`, `my @a; push(@a, 1 or die);`},
	{`sub foo {} sub bar {} foo 1, bar 2, 3;`, `sub foo {} sub bar {} foo(1, bar(2), 3);`},
}

// parenMeasurements are the perl 5.44.0 -MO=Concise comparisons the
// fidelity tests quote: each pair builds the same ops, or not.
var parenMeasurements = []struct {
	a, b string
	same bool
}{
	// TestFaithfulRejectsDroppedParenAfterMy: padrange + aassign against
	// padsv_store.
	{`my ($x) = @_;`, `my $x = @_;`, false},
	// TestFaithfulRejectsDroppedParenBeforeRepeat: repeat vK/DOLIST,2
	// against repeat vK/2.
	{`our $y; ($y) x 3;`, `our $y; $y x 3;`, false},
	// TestFaithfulRejectsDroppedParenAfterLocal: aassign against sassign.
	{`our $x; local ($x) = 1;`, `our $x; local $x = 1;`, false},
	// TestFaithfulDoesNotTreatCallParensAsNesting: const[IV 1] sP against
	// const[IV 1] s.
	{`print ((1))+2;`, `print (1)+2;`, false},
	// TestFaithfulNestingForgivesToOnePairNotZero: both repeat vK/DOLIST,2,
	// and neither is the string repeat.
	{`our $y; (($y)) x 3;`, `our $y; ($y) x 3;`, true},
	{`our $y; (($y)) x 3;`, `our $y; $y x 3;`, false},
	// TestFaithfulDoesNotTreatSliceParensAsNesting: a list slice's pair is
	// the slice's, and the pair inside it sets the bit: ex-list lKP against
	// lK, stat lKP against lK, padav lP against l.
	{`my $x = ((1,2,3))[1];`, `my $x = (1,2,3)[1];`, false},
	{`my $x = ((stat "/"))[2];`, `my $x = (stat "/")[2];`, false},
	{`my @a; my @x = ((@a))[0,1];`, `my @a; my @x = (@a)[0,1];`, false},
	{`my $x = ((caller))[2];`, `my $x = (caller)[2];`, false},
	{`my @a; my ($p,$q) = ((@a))[0,1];`, `my @a; my ($p,$q) = (@a)[0,1];`, false},
	// TestCanonKeepsTrailingCommaInItsPair: the comma's place moves no op.
	{`my @r = ((1),);`, `my @r = (1,);`, true},
	{`sub f {} f((0,),);`, `sub f {} f((0,));`, true},
	// TestFaithfulForgivesNestingAroundAGroupedOperand.
	{`my ($a,$b); my $x = ((($a+$b)*2));`, `my ($a,$b); my $x = (($a+$b)*2);`, true},
}

// canonTokensOf is canon's tokens of the parse of text, which must be
// canon's own spelling: its significant tokens are canon's.
func canonTokensOf(t *testing.T, text string) []parse.CanonToken {
	t.Helper()
	canon := parse.CanonTokens(parse.Parse([]byte(text)), []byte(text))
	got := make([]string, len(canon))
	for i, c := range canon {
		got[i] = c.Text
	}
	if want := sigTokens(text); !slices.Equal(got, want) {
		t.Fatalf("canon of %q is %q, not %q as written", text, got, want)
	}
	return canon
}

// tokenPair is a source and a canon to compare it against.
type tokenPair struct {
	src   string
	canon []parse.CanonToken
}

type comparator func([]string, []parse.CanonToken) (bool, []parse.Forgiveness, string)

// unlistedRules is every rule compare forgave a pair by that forgivenRules
// does not list, as "rule (source)".
func unlistedRules(compare comparator, pairs []tokenPair) []string {
	var out []string
	for _, p := range pairs {
		_, forgiven, _ := compare(sigTokens(p.src), p.canon)
		for _, f := range forgiven {
			if _, ok := forgivenRules[f.Rule]; !ok {
				out = append(out, fmt.Sprintf("%s (%q)", f.Rule, p.src))
			}
		}
	}
	return out
}

// unreplayed reports whether undoing each recorded forgiveness in canon
// fails to give src exactly, and how: a last-semicolon drops canon's `;`,
// which must carry the mark canon writes on a Block's last `;`, and an
// extra-nesting puts the source's paren back before the canon token it was
// found at. A difference faithfulTokens skips without recording one leaves
// a token behind, or misses one, here.
func unreplayed(src []string, canon []parse.CanonToken, forgiven []parse.Forgiveness) string {
	drop := map[int]bool{}
	insert := map[int][]string{}
	for _, f := range forgiven {
		switch f.Rule {
		case "extra-nesting":
			if f.SrcIndex < 0 || f.SrcIndex >= len(src) {
				return fmt.Sprintf("an extra-nesting at source token %d, past the source", f.SrcIndex)
			}
			insert[f.CanonIndex] = append(insert[f.CanonIndex], src[f.SrcIndex])
		default:
			if f.CanonIndex < 0 || f.CanonIndex >= len(canon) {
				return fmt.Sprintf("a %s at canon token %d, past canon", f.Rule, f.CanonIndex)
			}
			if c := canon[f.CanonIndex]; f.Rule == "last-semicolon" && (c.Text != ";" || !c.BlockSemi) {
				return fmt.Sprintf("a last-semicolon forgave canon token %d, %q, which is no block's marked `;`",
					f.CanonIndex, c.Text)
			}
			drop[f.CanonIndex] = true
		}
	}
	var left []string
	for i, c := range canon {
		left = append(left, insert[i]...)
		if !drop[i] {
			left = append(left, c.Text)
		}
	}
	left = append(left, insert[len(canon)]...)
	if slices.Equal(left, src) {
		return ""
	}
	return fmt.Sprintf("canon with its forgivenesses undone is %q, the source is %q", left, src)
}

// TestFaithfulForgivenessIsClosed: faithfulTokens forgives a difference by
// the listed rules and no other. Each rule's fixture is forgiven by that
// rule alone and measured equivalent by perl; the fixtures, the
// TestFaithfulTokens cases and every T1 file report no unlisted rule; on
// every faithful verdict, undoing the recorded forgivenesses gives the
// source exactly, so no difference was forgiven unrecorded; and the token
// kinds dropped before comparing are exactly the trivia.
func TestFaithfulForgivenessIsClosed(t *testing.T) {
	var pairs []tokenPair
	for rule, fx := range forgivenRules {
		canon := canonTokensOf(t, fx.canon)
		pairs = append(pairs, tokenPair{fx.src, canon})
		ok, forgiven, why := parse.FaithfulTokens(sigTokens(fx.src), canon)
		names := ruleNames(forgiven)
		if !ok || len(names) == 0 || slices.ContainsFunc(names, func(r string) bool { return r != rule }) {
			t.Errorf("%s: %q against %q gave %v %v (%s), want forgiven by %s alone",
				rule, fx.src, fx.canon, ok, names, why, rule)
		}
	}
	for _, tc := range faithfulTokensCases {
		pairs = append(pairs, tokenPair{tc.src, caseCanon(tc.src, tc.canon, tc.parsed)})
	}
	// The extent negatives: each wrong tree's canon is rejected, with
	// nothing forgiven on the way.
	for _, tc := range extentNegatives {
		canon := canonTokensOf(t, tc.wrong)
		pairs = append(pairs, tokenPair{tc.src, canon})
		if ok, forgiven, _ := parse.FaithfulTokens(sigTokens(tc.src), canon); ok || len(forgiven) > 0 {
			t.Errorf("%q against the wrong tree's canon %q gave %v forgiven by %v, want false with nothing forgiven",
				tc.src, tc.wrong, ok, ruleNames(forgiven))
		}
	}
	for _, p := range unlistedRules(parse.FaithfulTokens, pairs) {
		t.Errorf("forgiven by a rule forgivenRules does not list: %s", p)
	}
	// Each pair as it is, and with one stray token inserted anywhere in its
	// canon, under each mark or none: a skip that fires only on a difference
	// the corpus never shows is still caught, because the stray token
	// survives the replay.
	stray := []string{"+", "-", "!", "\\", "(", ")", "{", "}", "[", "]", ";", ",", "=>"}
	for _, p := range pairs {
		src, canon := sigTokens(p.src), p.canon
		variants := [][]parse.CanonToken{canon}
		for i := 0; i <= len(canon); i++ {
			for _, tok := range stray {
				for _, mark := range []parse.CanonToken{{}, {ParenNode: true}, {BlockSemi: true}} {
					mark.Text = tok
					variants = append(variants, slices.Insert(slices.Clone(canon), i, mark))
				}
			}
		}
		for _, c := range variants {
			if ok, forgiven, _ := parse.FaithfulTokens(src, c); ok {
				if why := unreplayed(src, c, forgiven); why != "" {
					t.Errorf("%q: %s", p.src, why)
				}
			}
		}
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
			want, canon := sigTokens(string(src)), parse.CanonTokens(parse.Parse(src), src)
			ok, forgiven, _ := parse.FaithfulTokens(want, canon)
			for _, f := range forgiven {
				if _, listed := forgivenRules[f.Rule]; !listed {
					t.Errorf("%s: forgiven by a rule forgivenRules does not list: %s", rel, f.Rule)
				}
			}
			if !ok {
				continue
			}
			if why := unreplayed(want, canon, forgiven); why != "" {
				t.Errorf("%s: %s", rel, why)
			}
		}
	})

	want := []lexer.Kind{lexer.Whitespace, lexer.Comment, lexer.Pod, lexer.DataSection}
	if !slices.Equal(parse.DroppedKinds, want) {
		t.Errorf("Faithful drops token kinds %v before comparing, want exactly %v", parse.DroppedKinds, want)
	}

	t.Run("perl", func(t *testing.T) {
		perl := perltest.PinnedPerl(t)
		concise := func(src string) string {
			out, _ := perl.Command("-MO=Concise", "-e", src).CombinedOutput()
			return string(out)
		}
		for rule, fx := range forgivenRules {
			a, errA := perl.Command("-e", fx.src).CombinedOutput()
			b, errB := perl.Command("-e", fx.canon).CombinedOutput()
			if errA != nil || errB != nil || string(a) != string(b) {
				t.Errorf("%s: perl runs %q to %q (%v) and %q to %q (%v); the rule forgives only an equivalence",
					rule, fx.src, a, errA, fx.canon, b, errB)
			}
			if a, b := concise(fx.src), concise(fx.canon); a != b {
				t.Errorf("%s: perl compiles %q and %q to different ops:\n%s\n%s", rule, fx.src, fx.canon, a, b)
			}
		}
		for _, tc := range extentNegatives {
			if a, b := concise(tc.src), concise(tc.wrong); a == b {
				t.Errorf("perl compiles %q and the wrong reading %q to the same ops:\n%s", tc.src, tc.wrong, a)
			}
		}
		for _, m := range parenMeasurements {
			if a, b := concise(m.a), concise(m.b); (a == b) != m.same {
				t.Errorf("perl compiles %q and %q to the same ops: %v, the tests say %v:\n%s\n%s",
					m.a, m.b, a == b, m.same, a, b)
			}
		}
	})
}

// TestFaithfulCallParenRuleIsGone: a call's parens are not forgiven. Canon
// writes them only where the source did or binding power needs them, so a
// pair the source lacks is a difference.
func TestFaithfulCallParenRuleIsGone(t *testing.T) {
	for _, tc := range []struct{ src, canon string }{
		{`print "x";`, `print ( "x" ) ;`},
		{`defined $x;`, `defined ( $x ) ;`},
	} {
		if ok, forgiven, _ := parse.FaithfulTokens(sigTokens(tc.src), asCanon(sigTokens(tc.canon))); ok {
			t.Errorf("FaithfulTokens(%q, %q) = true, forgiven by %v; want false", tc.src, tc.canon, ruleNames(forgiven))
		}
	}
}

// TestFaithfulLastSemicolonCoversBlockArgument: a block argument's last `;`
// is forgiven by the main loop's rule, with no paren around it, and a call
// paren canon adds after the block is not.
func TestFaithfulLastSemicolonCoversBlockArgument(t *testing.T) {
	src := `print {$fh} "x";`
	ok, forgiven, why := parse.FaithfulTokens(sigTokens(src), canonTokensOf(t, `print {$fh;} "x";`))
	if !ok || !slices.Equal(ruleNames(forgiven), []string{"last-semicolon"}) {
		t.Errorf("FaithfulTokens(%q, `print {$fh;} \"x\";`) = %v forgiven by %v (%s), want true by [last-semicolon]",
			src, ok, ruleNames(forgiven), why)
	}
	if ok, why := parse.Faithful(parse.Parse([]byte(src)), []byte(src)); !ok {
		t.Errorf("Faithful(%q) = false: %s", src, why)
	}
	if ok, _, _ := parse.FaithfulTokens(sigTokens(src), canonTokensOf(t, `print {$fh;} ("x");`)); ok {
		t.Errorf("FaithfulTokens(%q, `print {$fh;} (\"x\");`) = true, want false", src)
	}
}

// TestFaithfulForgivenessIsClosedRejectsUnlistedRule: a comparator that
// forgives by a rule the closed set does not list is caught.
func TestFaithfulForgivenessIsClosedRejectsUnlistedRule(t *testing.T) {
	invented := func([]string, []parse.CanonToken) (bool, []parse.Forgiveness, string) {
		return true, []parse.Forgiveness{{Rule: "invented"}}, ""
	}
	got := unlistedRules(invented, []tokenPair{{`print "x";`, asCanon(sigTokens(`print("x");`))}})
	if want := []string{`invented ("print \"x\";")`}; !slices.Equal(got, want) {
		t.Errorf("unlistedRules = %q, want %q", got, want)
	}
}

// TestFaithfulTokensReportsForgivenessOnFailure: a rule that fired before
// the mismatch is reported with the failing verdict, so the closed-set
// check sees it on unfaithful files too.
func TestFaithfulTokensReportsForgivenessOnFailure(t *testing.T) {
	ok, forgiven, _ := parse.FaithfulTokens(sigTokens(`sub f { 1 } 1 + 2;`), canonTokensOf(t, `sub f { 1; } 3;`))
	if ok || !slices.Equal(ruleNames(forgiven), []string{"last-semicolon"}) {
		t.Errorf("got %v forgiven by %v, want false forgiven by [last-semicolon]", ok, ruleNames(forgiven))
	}
}

// TestCanonMarksEveryParenNodePair: over T1, canon marks one pair for each
// parenthesised node it writes, but for the two whose pair is a call's (see
// CanonToken). A nested emission copied into its parent without its marks
// -- WriteString where writeFrom belongs -- loses pairs silently, and this
// count is what sees it.
func TestCanonMarksEveryParenNodePair(t *testing.T) {
	dir, files := t1Files(t)
	for _, rel := range files {
		src, err := os.ReadFile(filepath.Join(dir, rel))
		if err != nil {
			continue
		}
		root := parse.Parse(src)
		marked := 0
		for _, tok := range parse.CanonTokens(root, src) {
			if tok.ParenNode && tok.Text == "(" {
				marked++
			}
		}
		if want := parenNodes(root, src, false); marked != want {
			t.Errorf("%s: canon marks %d pairs, the tree has %d parenthesised nodes", rel, marked, want)
		}
	}
}

// parenNodes counts the parenthesised nodes canon emits under n, but for
// one whose pair the syntax requires: the argument of `->(`, the operand of
// `not (` and a list slice's list, which argParens says n is. It does not look inside a node canon
// writes as its source text rather than emitting its children: an
// Unknown, or an Index, Binary or Ternary short of its children --
// `}=>->join->()()` reads the trailing `()` as an Index holding only the
// empty List.
func parenNodes(n *parse.Node, src []byte, argParens bool) int {
	count := 0
	if n.Paren && !argParens {
		count++
	}
	switch {
	case n.Kind == parse.Unknown,
		n.Kind == parse.Index && len(n.Children) != 2,
		n.Kind == parse.Binary && len(n.Children) != 2,
		n.Kind == parse.Ternary && len(n.Children) != 3:
		return count
	}
	for i, c := range n.Children {
		call := n.Kind == parse.Binary && n.Text == "->" && i == 1 && c.Paren && src[c.Start] == '(' ||
			n.Kind == parse.Unary && n.Text == "not" && len(n.Children) == 1 && c.Paren ||
			n.Kind == parse.Index && !n.Arrow && n.Text == "[" && i == 0 && c.Paren
		count += parenNodes(c, src, call)
	}
	return count
}

// TestFaithfulForgivesOnlyExtraNesting: a source pair nested directly
// inside the pair a node's Paren flag stands for is forgiven, however
// deep; a pair canon does not write at all is not. perl keeps such a
// paren as one boolean bit (forgivenRules' extra-nesting fixture).
func TestFaithfulForgivesOnlyExtraNesting(t *testing.T) {
	for _, src := range []string{
		`my $z = (($x));`,
		`my $z = ((($x)));`,
		`ok((($y eq "x")));`,
		`my @r = ((1, 2)) x 2;`,
	} {
		ok, why := parse.Faithful(parse.Parse([]byte(src)), []byte(src))
		if !ok {
			t.Errorf("Faithful(%q) = false (%s), want true", src, why)
		}
		_, forgiven, _ := parse.FaithfulTokens(sigTokens(src), parse.CanonTokens(parse.Parse([]byte(src)), []byte(src)))
		if names := ruleNames(forgiven); len(names) == 0 ||
			slices.ContainsFunc(names, func(r string) bool { return r != "extra-nesting" }) {
			t.Errorf("FaithfulTokens(%q) forgave %v, want extra-nesting alone", src, names)
		}
	}
	// No Paren on the node: canon writes no pair to nest inside.
	if ok, _, _ := parse.FaithfulTokens(sigTokens(`my $z = (($x));`), canonTokensOf(t, `my $z = $x;`)); ok {
		t.Errorf("FaithfulTokens(`my $z = (($x));`, `my $z = $x;`) = true, want false")
	}
	// The same pair unmarked, as a call's or a binding-power one is.
	if ok, _, _ := parse.FaithfulTokens(sigTokens(`my $z = (($x));`), asCanon(sigTokens(`my $z = ($x);`))); ok {
		t.Errorf("FaithfulTokens(`my $z = (($x));`, unmarked `my $z = ($x);`) = true, want false")
	}
}

// rejects fails t unless FaithfulTokens rejects src against the canon of
// canon's own spelling.
func rejects(t *testing.T, src, canon string) {
	t.Helper()
	if ok, forgiven, _ := parse.FaithfulTokens(sigTokens(src), canonTokensOf(t, canon)); ok {
		t.Errorf("FaithfulTokens(%q, %q) = true, forgiven by %v; want false", src, canon, ruleNames(forgiven))
	}
}

// TestFaithfulRejectsDroppedParenAfterMy: `my ($x) = @_` is a list
// assignment and `my $x = @_` a scalar one. Measured on perl 5.44.0
// (parenMeasurements): padrange + aassign against padsv_store.
func TestFaithfulRejectsDroppedParenAfterMy(t *testing.T) {
	rejects(t, `my ($x) = @_;`, `my $x = @_;`)
}

// TestFaithfulRejectsDroppedParenBeforeRepeat: `($y) x 3` is a list repeat
// and `$y x 3` a string one. Measured on perl 5.44.0
// (parenMeasurements): repeat vK/DOLIST,2 against repeat vK/2.
func TestFaithfulRejectsDroppedParenBeforeRepeat(t *testing.T) {
	rejects(t, `($y) x 3;`, `$y x 3;`)
}

// TestFaithfulRejectsDroppedParenAfterLocal: for `our $x`, `local ($x) = 1`
// is a list assignment and `local $x = 1` a scalar one. Measured on perl
// 5.44.0 (parenMeasurements): aassign against sassign.
func TestFaithfulRejectsDroppedParenAfterLocal(t *testing.T) {
	rejects(t, `our $x; local ($x) = 1;`, `our $x; local $x = 1;`)
}

// TestFaithfulDoesNotTreatCallParensAsNesting: a call's argument pair is
// not a Paren node's, so a pair inside it is not extra nesting. Measured
// on perl 5.44.0 (parenMeasurements), `print ((1))+2` is `const[IV 1] sP`
// and `print (1)+2` is `const[IV 1] s`.
func TestFaithfulDoesNotTreatCallParensAsNesting(t *testing.T) {
	rejects(t, `print((1))+2;`, `print(1) + 2;`)
}

// TestFaithfulDoesNotForgiveInnerPairAroundSubterm: a pair inside the
// node's pair but around one of its operands is not nesting: it wraps a
// different node.
func TestFaithfulDoesNotForgiveInnerPairAroundSubterm(t *testing.T) {
	rejects(t, `((1) + 2);`, `(1 + 2);`)
}

// TestFaithfulNestingForgivesToOnePairNotZero: doubled parens in a
// context-changing position are one pair, never none. Measured on perl
// 5.44.0 (parenMeasurements), `(($y)) x 3` and `($y) x 3` are both repeat
// vK/DOLIST,2; `$y x 3` is the string repeat.
func TestFaithfulNestingForgivesToOnePairNotZero(t *testing.T) {
	rejects(t, `(($y)) x 3;`, `$y x 3;`)
	if ok, why := parse.Faithful(parse.Parse([]byte(`(($y)) x 3;`)), []byte(`(($y)) x 3;`)); !ok {
		t.Errorf("Faithful(`(($y)) x 3;`) = false (%s), want true", why)
	}
}

// TestFaithfulDistinguishesMyListFromMyComma: `my ($p, $q) = (1, 2)`
// declares both names, and `my $p, $q` declares one and evaluates the
// other. Canon writes each as written, and neither is faithful against the
// other's canon.
func TestFaithfulDistinguishesMyListFromMyComma(t *testing.T) {
	list, comma := `my ($p, $q) = (1, 2);`, `my $p, $q;`
	for _, src := range []string{list, comma} {
		if got := parse.Canon(parse.Parse([]byte(src)), []byte(src)); got != src {
			t.Errorf("Canon(%q) = %q, want it as written", src, got)
		}
	}
	rejects(t, list, comma)
	rejects(t, comma, list)
}

// TestFaithfulDoesNotTreatSliceParensAsNesting: a list slice's pair is
// syntax, as a call's is, so a pair inside it is not extra nesting and
// canon writes the slice's pair unmarked. Measured on perl 5.44.0
// (parenMeasurements), `((1,2,3))[1]` is `ex-list lKP` and `(1,2,3)[1]` is
// `ex-list lK`; `((@a))[0,1]` is `padav lP` and `(@a)[0,1]` is `padav l`.
// psc's tree holds one Paren for both pairs, so canon writes one, and the
// doubled source is unfaithful (01a16602-5f1a).
func TestFaithfulDoesNotTreatSliceParensAsNesting(t *testing.T) {
	for _, src := range []string{
		`my $x = ((1,2,3))[1];`,
		`my $x = (((1,2,3)))[1];`,
		`my $x = ((stat "/"))[2];`,
		`my @a; my @x = ((@a))[0,1];`,
		`my $x = ((caller))[2];`,
		`my @a; my ($p,$q) = ((@a))[0,1];`,
	} {
		if ok, _ := parse.Faithful(parse.Parse([]byte(src)), []byte(src)); ok {
			t.Errorf("Faithful(%q) = true, want false", src)
		}
	}
	for _, tc := range []struct{ src, want string }{
		{`my $x = (1,2,3)[1];`, `my $x = ( 1 , 2 , 3 ) [ 1 ] ;`},
		{`my @a; my @x = (@a)[0];`, `my @a ; my @x = ( @a ) [ 0 ] ;`},
	} {
		if got := markedCanon(tc.src); got != tc.want {
			t.Errorf("canon tokens of %q are %q, want %q", tc.src, got, tc.want)
		}
		if ok, why := parse.Faithful(parse.Parse([]byte(tc.src)), []byte(tc.src)); !ok {
			t.Errorf("Faithful(%q) = false (%s), want true", tc.src, why)
		}
	}
}

// TestCanonKeepsTrailingCommaInItsPair: a trailing comma is written inside
// the pair the source wrote it in. Measured on perl 5.44.0
// (parenMeasurements), `((1),)` and `(1,)` build the same ops, as do
// `f((0,),)` and `f((0,))`; each re-emits as written.
func TestCanonKeepsTrailingCommaInItsPair(t *testing.T) {
	for _, src := range []string{`my @r = ((1),);`, `sub f {} f((0,),);`, `sub f {} f((0),);`, `my @r = (1,);`} {
		if ok, why := parse.Faithful(parse.Parse([]byte(src)), []byte(src)); !ok {
			t.Errorf("Faithful(%q) = false (%s), want true", src, why)
		}
	}
}

// TestFaithfulForgivesNestingAroundAGroupedOperand: an extra pair whose
// first token is itself a grouped operand's `(` is still extra nesting.
// Measured on perl 5.44.0 (parenMeasurements), `((($a+$b)*2))` and
// `(($a+$b)*2)` build the same ops. Paired greedily, the source's second
// `(` would be matched to canon's operand pair and the nesting missed.
func TestFaithfulForgivesNestingAroundAGroupedOperand(t *testing.T) {
	src := `my ($a,$b); my $x = ((($a+$b)*2));`
	if ok, why := parse.Faithful(parse.Parse([]byte(src)), []byte(src)); !ok {
		t.Errorf("Faithful(%q) = false (%s), want true", src, why)
	}
	// Not when canon nests the same way: the source's pair is canon's.
	if ok, forgiven, _ := parse.FaithfulTokens(sigTokens(`((1));`),
		[]parse.CanonToken{{Text: "(", ParenNode: true}, {Text: "(", ParenNode: true}, {Text: "1"},
			{Text: ")", ParenNode: true}, {Text: ")", ParenNode: true}, {Text: ";"}}); !ok || len(forgiven) != 0 {
		t.Errorf("((1)) against a doubly marked canon gave %v forgiven by %v, want true with nothing forgiven",
			ok, ruleNames(forgiven))
	}
}

// TestFaithfulNestingNeedsAnOpeningParen: only a `(` opens a nested pair.
// A word right inside a marked pair, whose own call pair closes right
// before it, is matched as itself.
func TestFaithfulNestingNeedsAnOpeningParen(t *testing.T) {
	canon := []parse.CanonToken{{Text: "(", ParenNode: true}, {Text: "f"}, {Text: "("}, {Text: "1"},
		{Text: ")"}, {Text: ")", ParenNode: true}, {Text: ";"}}
	if ok, forgiven, why := parse.FaithfulTokens(sigTokens(`(f(1));`), canon); !ok || len(forgiven) != 0 {
		t.Errorf("(f(1)) against its own canon gave %v forgiven by %v (%s), want true with nothing forgiven",
			ok, ruleNames(forgiven), why)
	}
}

// hashSemicolons are a `;` written inside braces perl reads as an anonymous
// hash or a subscript, each source against that canon. Measured on perl
// 5.44.0, every source compiles and every canon is a syntax error near
// ";}": the `;` is no spelling of the same program, so it is never forgiven.
var hashSemicolons = []struct{ src, canon string }{
	{`$h{a};`, `$h{a;};`},
	{`$h->{a};`, `$h->{a;};`},
	{`@h{a};`, `@h{a;};`},
	{`$h = {a => 1};`, `$h = {a => 1;};`},
}

// realBlocks are code blocks whose last statement the source leaves without
// its `;`. Canon writes one, and last-semicolon forgives it: perl 5.44.0
// compiles each source and its canon alike.
var realBlocks = []string{
	"my @m = map { $_ + 1 } (1, 2);",
	"my @g = grep { $_ > 1 } (1, 2);",
	"my @s = sort { $a <=> $b } (2, 1);",
	"my $d = do { 5 + 6 };",
	"my $f = sub { 1 };",
	"my $e = eval { 7 * 8 };",
	`print {*STDOUT} "x";`,
	"{ print 1 }",
	"sub f { require Carp }",
}

// TestFaithfulForgivesSemicolonOnlyInBlocks: at canon-token level, a `;`
// inside an anonymous hash or a subscript is never forgiven. Canon of each
// source's own tree marks no `;`, and the `;` such braces would take is
// unmarked, so last-semicolon cannot reach it.
func TestFaithfulForgivesSemicolonOnlyInBlocks(t *testing.T) {
	for _, tc := range hashSemicolons {
		src := []byte(tc.src)
		for _, tok := range parse.CanonTokens(parse.Parse(src), src) {
			if tok.BlockSemi {
				t.Errorf("canon of %q marks a %q as a block's last `;`", tc.src, tok.Text)
			}
		}
		if ok, forgiven, _ := parse.FaithfulTokens(sigTokens(tc.src), asCanon(sigTokens(tc.canon))); ok || len(forgiven) > 0 {
			t.Errorf("FaithfulTokens(%q, %q) = %v forgiven by %v, want false with nothing forgiven",
				tc.src, tc.canon, ok, ruleNames(forgiven))
		}
	}
	t.Run("perl", func(t *testing.T) {
		perl := perltest.PinnedPerl(t)
		for _, tc := range hashSemicolons {
			if out, err := perl.Command("-c", "-e", tc.src).CombinedOutput(); err != nil {
				t.Errorf("perl -c %q: %v\n%s", tc.src, err, out)
			}
			if out, err := perl.Command("-c", "-e", tc.canon).CombinedOutput(); err == nil ||
				!strings.Contains(string(out), `near ";}"`) {
				t.Errorf("perl -c %q: %v, want a syntax error near \";}\"\n%s", tc.canon, err, out)
			}
		}
	})
}

// TestFaithfulLastSemicolonOnlyInBlocks: `$h{a;}` against `$h{a}`, and a
// hash constructor written `{a => 1;}`, are unfaithful; a real block's last
// `;` -- map, grep, sort, do, sub, eval, a print block and a bare block --
// is still forgiven, by last-semicolon alone, on the canon of its own tree.
func TestFaithfulLastSemicolonOnlyInBlocks(t *testing.T) {
	for _, tc := range hashSemicolons {
		if ok, _, _ := parse.FaithfulTokens(sigTokens(tc.src), asCanon(sigTokens(tc.canon))); ok {
			t.Errorf("FaithfulTokens(%q, %q) = true, want false", tc.src, tc.canon)
		}
	}
	for _, src := range realBlocks {
		canon := parse.CanonTokens(parse.Parse([]byte(src)), []byte(src))
		ok, forgiven, why := parse.FaithfulTokens(sigTokens(src), canon)
		names := ruleNames(forgiven)
		if !ok || len(names) == 0 || slices.ContainsFunc(names, func(r string) bool { return r != "last-semicolon" }) {
			t.Errorf("%q against its canon %q: %v forgiven by %v (%s), want true by last-semicolon alone",
				src, parse.Canon(parse.Parse([]byte(src)), []byte(src)), ok, names, why)
		}
	}
	// Every token canon marks is a `;` right before a `}`, here and over
	// T1: a mark left over from an earlier statement would sit elsewhere,
	// and so would one on the `;` a heredoc body follows.
	marksBeforeBrace := func(t *testing.T, name string, canon []parse.CanonToken) {
		for i, tok := range canon {
			if tok.BlockSemi && (tok.Text != ";" || i+1 == len(canon) || canon[i+1].Text != "}") {
				t.Errorf("%s: canon token %d, %q, is marked a block's last `;`", name, i, tok.Text)
			}
		}
	}
	for _, src := range append(slices.Clone(realBlocks), "{ 1; is f(10), 4; }", "sub g {\n  print <<E\nbody\nE\n}\n") {
		marksBeforeBrace(t, src, parse.CanonTokens(parse.Parse([]byte(src)), []byte(src)))
	}
	// A heredoc body canon writes into a quote that runs on past the
	// opener, ahead of the statement's `;`: that `;` still ends the block,
	// and is marked.
	runOn := "sub first { $_[0] } sub h {\n  first(<<END, 'u') eq\nhello\nEND\nq(hello\n)\n}\n"
	if canon := parse.CanonTokens(parse.Parse([]byte(runOn)), []byte(runOn)); !canon[len(canon)-2].BlockSemi {
		t.Errorf("canon of %q leaves the `;` before its last `}` unmarked: %v", runOn, canon)
	}
	t.Run("T1", func(t *testing.T) {
		dir, files := t1Files(t)
		for _, rel := range files {
			src, err := os.ReadFile(filepath.Join(dir, rel))
			if err != nil {
				t.Errorf("%s: %v", rel, err)
				continue
			}
			marksBeforeBrace(t, rel, parse.CanonTokens(parse.Parse(src), src))
		}
	})
	t.Run("perl", func(t *testing.T) {
		perl := perltest.PinnedPerl(t)
		for _, src := range realBlocks {
			canon := parse.Canon(parse.Parse([]byte(src)), []byte(src))
			for _, text := range []string{src, canon} {
				if out, err := perl.Command("-c", "-e", text).CombinedOutput(); err != nil {
					t.Errorf("perl -c %q: %v\n%s", text, err, out)
				}
			}
		}
	})
}
