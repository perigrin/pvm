// ABOUTME: Tests of the generator's refusals on toy grammars: each check trips on a toy made from a consistent one.
// ABOUTME: Also that the committed grammar passes them, and that a refusal writes nothing.

package grammar

import (
	"bytes"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
)

// refToy is a grammar and the reference list it cites.
type refToy struct {
	g   *Grammar
	ref []Entry
}

func (r refToy) generate() error {
	_, err := Generate(r.g, r.ref)
	return err
}

// consistentToy is
//
//	s ::= A b    (perly.y:10)
//	b ::= B      (perly.y:12)
//	b ::=        (perly.y:14)
//
// with A and B classed.
func consistentToy() refToy {
	g := toy([]string{"A", "B"}, alts("s", "A b"), alts("b", "B", ""))
	for i, c := range []string{"perly.y:10", "perly.y:12", "perly.y:14"} {
		g.G1[i].Cite = c
	}
	g.Classes = []ClassDecl{
		{"A", TokenClass{Kind: ptr(lexer.Operator), Text: ptr("+")}},
		{"B", TokenClass{Kind: ptr(lexer.Word), Text: ptr("x")}},
	}
	ref := []Entry{{10, "s: A b", ""}, {12, "b: B", ""}, {14, "b:", ""}}
	return refToy{g, ref}
}

// classless is the toy with its terminals named as given and no class for any of them.
func classless(terminals ...string) refToy {
	r := consistentToy()
	r.g.G0, r.g.Classes = nil, nil
	for _, term := range terminals {
		r.g.G0 = append(r.g.G0, Rule{LHS: term})
	}
	r.g.G1 = alts("s", strings.Join(terminals, " "))
	r.g.G1[0].Cite = "perly.y:10"
	return r
}

func TestGeneratorRefusals(t *testing.T) {
	if err := consistentToy().generate(); err != nil {
		t.Fatalf("the consistent toy is refused: %v", err)
	}
	cases := []struct {
		name  string
		trip  func(*refToy)
		words []string
	}{
		{"uncited rule", func(r *refToy) { r.g.G1[1].Cite = "" }, []string{"b ::= B", "citation"}},
		{"citation outside the reference list", func(r *refToy) { r.g.G1[1].Cite = "perly.y:99" }, []string{"b ::= B", "perly.y:99"}},
		{"literal the lexer splits", func(r *refToy) { r.g.Classes[0].Key.Text = ptr("~~~") }, []string{`"~~~"`}},
		{"terminal with no class", func(r *refToy) { r.g.Classes = r.g.Classes[:1] }, []string{"B", "class"}},
		{"unreachable nonterminal", func(r *refToy) {
			r.g.G1 = append(r.g.G1, Rule{LHS: "c", RHS: []Symbol{name("A")}, Cite: "perly.y:10"})
		}, []string{"c", "unreachable"}},
		{"unproductive nonterminal", func(r *refToy) {
			r.g.G1 = append(r.g.G1, Rule{LHS: "s", RHS: []Symbol{name("A"), name("d")}, Cite: "perly.y:10"},
				Rule{LHS: "d", RHS: []Symbol{name("d"), name("A")}, Cite: "perly.y:10"})
		}, []string{"d", "unproductive"}},
		{"precedence token used by nothing", func(r *refToy) { r.g.Prec = []string{"UMINUS"} }, []string{"UMINUS", "precedence"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := consistentToy()
			c.trip(&r)
			wantErr(t, c.name, r.generate(), c.words...)
		})
	}
}

// A precedence token is used by a right-hand side or by a %prec, and a %prec alone is a use.
func TestGeneratorPrecedenceUse(t *testing.T) {
	r := consistentToy()
	r.g.Prec = []string{"A"}
	if err := r.generate(); err != nil {
		t.Errorf("a precedence token on a right-hand side is refused: %v", err)
	}
	r = consistentToy()
	r.g.Prec = []string{"UMINUS"}
	r.g.G1[1].Prec = "UMINUS"
	if err := r.generate(); err != nil {
		t.Errorf("a %%prec-only use is refused: %v", err)
	}
}

// Classes are declared by terminal name: a rule added ahead of the terminals renumbers them,
// and each class stays on its terminal.
func TestClassesFollowTheirTerminals(t *testing.T) {
	admits := func(r refToy) map[string]string {
		tab, err := Generate(r.g, r.ref)
		if err != nil {
			t.Fatal(err)
		}
		classes, err := r.g.ResolveClasses(tab)
		if err != nil {
			t.Fatal(err)
		}
		got := map[string]string{}
		for _, c := range classes {
			got[*c.Text] = tab.Symbols[c.Terminals[0]].Name
		}
		return got
	}
	want := map[string]string{"+": "A", "x": "B"}
	before := consistentToy()
	if got := admits(before); !maps.Equal(got, want) {
		t.Fatalf("classes admit %v, want %v", got, want)
	}
	after := consistentToy()
	after.g.G1 = append(after.g.G1, Rule{LHS: "c", RHS: []Symbol{name("A"), name("B")}, Cite: "perly.y:10"})
	after.g.G1[0].RHS = append(after.g.G1[0].RHS, name("c")) // s ::= A b c: a nonterminal before the terminals
	if got := admits(after); !maps.Equal(got, want) {
		t.Errorf("after a rule is added, classes admit %v, want %v", got, want)
	}
}

func TestClassNamesAreTerminals(t *testing.T) {
	for name, term := range map[string]string{"unknown": "NOSUCH", "nonterminal": "b"} {
		r := consistentToy()
		r.g.Classes = append(r.g.Classes, ClassDecl{term, TokenClass{Kind: ptr(lexer.Word)}})
		wantErr(t, name, r.generate(), term, "class")
	}
}

func TestGeneratorRefusalExceptionsExact(t *testing.T) {
	for _, term := range []string{"PLUGIN_ADD_OP", "GRAMPROG", "GRAMSUBSIGNATURE", "PLUGEXPR", "PLUGSTMT"} {
		if err := classless(term).generate(); err != nil {
			t.Errorf("classless %s is refused: %v", term, err)
		}
	}
	for _, term := range []string{"PLUGINS", "PLUGIN", "GRAMMAR", "GRAM", "PLUGEXPRESSION", "PLUGSTMTS", "XPLUGEXPR", "NOPLUGIN_ADD_OP"} {
		wantErr(t, "classless "+term, classless(term).generate(), term, "class")
	}
}

func TestGeneratorRefusesSplitLiteral(t *testing.T) {
	literal := func(kind lexer.Kind, text string) error {
		r := consistentToy()
		r.g.Classes[0].Key.Kind, r.g.Classes[0].Key.Text = &kind, &text
		return r.generate()
	}
	// ~~~ lexes as two tokens, and "x " as the word x: the right kind, other text.
	wantErr(t, "~~~", literal(lexer.Operator, "~~~"), "~~~")
	wantErr(t, `"x "`, literal(lexer.Word, "x "), `"x "`)
	// ~~ is one operator between two terms, and x= one word.
	for _, ok := range []struct {
		kind lexer.Kind
		text string
	}{{lexer.Operator, "~~"}, {lexer.Operator, "**"}, {lexer.Word, "x="}, {lexer.Operator, "/"}} {
		if err := literal(ok.kind, ok.text); err != nil {
			t.Errorf("%v %q is refused: %v", ok.kind, ok.text, err)
		}
	}
	// A class with a text and no kind is held to the text alone, in any kind.
	textOnly := func(text string) error {
		r := consistentToy()
		r.g.Classes[0].Key = TokenClass{Text: &text}
		return r.generate()
	}
	wantErr(t, "text-only ~~~", textOnly("~~~"), "~~~")
	if err := textOnly("~~"); err != nil {
		t.Errorf("text-only ~~ is refused: %v", err)
	}
	// __END__ lexes as a word between two terms and as a data section only from a statement's start:
	// the lexer's checkpoint and resume reach it, a string around it does not.
	if err := literal(lexer.DataSection, "__END__"); err != nil {
		t.Errorf("a literal that lexes through checkpoint and resume is refused: %v", err)
	}
	wantErr(t, "__END__ as an operator", literal(lexer.Operator, "__END__"), "__END__")
}

func TestGeneratorNullableIsProductive(t *testing.T) {
	// b is productive only through its empty alternative.
	if err := consistentToy().generate(); err != nil {
		t.Errorf("a nonterminal productive through an empty alternative is refused: %v", err)
	}
	r := consistentToy()
	r.g.G1 = r.g.G1[:2] // without b's empty alternative, b ::= B is still productive
	if err := r.generate(); err != nil {
		t.Errorf("b ::= B alone is refused: %v", err)
	}
	r = consistentToy()
	r.g.G1[1].RHS = []Symbol{name("b")} // b ::= b | (empty): productive
	if err := r.generate(); err != nil {
		t.Errorf("b ::= b | empty is refused: %v", err)
	}
	r.g.G1 = r.g.G1[:2] // b ::= b alone derives nothing
	wantErr(t, "b ::= b", r.generate(), "b", "unproductive")
}

func TestCommittedGrammarConsistent(t *testing.T) {
	g, ref, err := Committed()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Generate(g, ref); err != nil {
		t.Errorf("the committed grammar is refused:\n%v", err)
	}
}

// files reads every file in dir by name.
func files(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	m := map[string][]byte{}
	for _, e := range entries {
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		m[e.Name()] = b
	}
	return m
}

func TestGeneratorRefusalWritesNothing(t *testing.T) {
	dir := t.TempDir()
	ok := consistentToy()
	if err := GenerateTables(dir, ok.g, ok.ref); err != nil {
		t.Fatal(err)
	}
	before := files(t, dir)
	if len(before) == 0 {
		t.Fatal("a consistent grammar wrote no tables")
	}
	bad := consistentToy()
	bad.g.G1[1].Cite = ""
	bad.g.G1 = append(bad.g.G1, Rule{LHS: "z", RHS: []Symbol{name("A")}, Cite: "perly.y:10"}) // would change every table
	wantErr(t, "refused generation", GenerateTables(dir, bad.g, bad.ref), "b ::= B")
	after := files(t, dir)
	for name, data := range before {
		if !bytes.Equal(after[name], data) {
			t.Errorf("%s changed after a refusal", name)
		}
	}
	if len(after) != len(before) {
		t.Errorf("a refusal left %d files, want %d: %v", len(after), len(before), slices.Sorted(maps.Keys(after)))
	}
	// A consistent grammar still rewrites them.
	if err := os.WriteFile(filepath.Join(dir, "rules.txt"), []byte("stale\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := GenerateTables(dir, ok.g, ok.ref); err != nil {
		t.Fatal(err)
	}
	if got := files(t, dir); !bytes.Equal(got["rules.txt"], before["rules.txt"]) {
		t.Errorf("a consistent grammar did not rewrite rules.txt: %q", got["rules.txt"])
	}
}

// perlyYPath is the pinned perl's perly.y, which the offline scripts also read.
const perlyYPath = "/opt/corpora/perl5/perly.y"

// The slots of perlyClasses are the lexed terminals perly.y declares, in the order it first
// declares them (%token, %left, %right, %nonassoc), less the ones needing no class.
func TestClassSlotsAreThePerlyYTokens(t *testing.T) {
	y, err := os.ReadFile(perlyYPath)
	if err != nil {
		t.Skipf("perly.y is not on this machine: %v", err)
	}
	decls, _, _ := strings.Cut(string(y), "\n%%")
	var want []string
	seen := map[string]bool{}
	for _, l := range regexp.MustCompile(`(?m)^%(?:token|left|right|nonassoc)\b(.*)$`).FindAllStringSubmatch(decls, -1) {
		for _, name := range strings.Fields(regexp.MustCompile(`<\w+>`).ReplaceAllString(l[1], " ")) {
			if seen[name] || strings.HasPrefix(name, "GRAM") || strings.HasPrefix(name, "PLUGIN_") ||
				slices.Contains([]string{"PLUGEXPR", "PLUGSTMT", "PREC_LOW", "UMINUS"}, name) {
				continue
			}
			seen[name] = true
			want = append(want, name)
		}
	}
	src, err := os.ReadFile("perly.go")
	if err != nil {
		t.Fatal(err)
	}
	_, classes, _ := strings.Cut(string(src), "var perlyClasses")
	var got []string
	for _, m := range regexp.MustCompile(`(?m)^\t(?://\s*|\{")([A-Za-z_0-9]+)`).FindAllStringSubmatch(classes, -1) {
		if len(got) == 0 || got[len(got)-1] != m[1] { // a terminal may be declared more than once
			got = append(got, m[1])
		}
	}
	if !slices.Equal(got, want) {
		t.Errorf("perlyClasses has %d slots, perly.y declares %d lexed tokens; first difference: %s", len(got), len(want), firstDifference("perly.go", "perly.y", want, got))
	}
}
