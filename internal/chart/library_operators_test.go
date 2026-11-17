// ABOUTME: Tests library-declared infix operators read as perl's PLUGIN_* terminals where a use brings them into scope.
// ABOUTME: The fixture testdata/libops/declarations/My/Ops.tpm declares one in each of perl's eleven buckets and one between two.

package chart

import (
	"os"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/forest"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/parse"
)

// libopsScope is src's Scope with the libops fixture modules declared and nothing else loadable.
func libopsScope(t *testing.T, src string) forest.Scope {
	t.Helper()
	t.Cleanup(parse.OverlayDeclarations(os.DirFS("testdata/libops")))
	root := parse.ParseWithLoader([]byte(src), func(string) ([]byte, bool) { return nil, false })
	return forest.Scope{Imports: parse.Imports(root), LibraryOperators: parse.LibraryOperators(root)}
}

// grouping is src's one derivation under its scope with every binary operator's operands parenthesized, "refused"
// where forest.Parse refuses it, or "ambiguous" where more than one derivation is left.
func grouping(t *testing.T, src string) string {
	t.Helper()
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	f, _ := forest.Parse(tab, []byte(src), libopsScope(t, src))
	if f == nil {
		return "refused"
	}
	trees, _ := f.Derivations(f.Root, 2)
	if len(trees) != 1 {
		return "ambiguous"
	}
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	var render func(tr *forest.Tree) string
	render = func(tr *forest.Tree) string {
		if len(tr.Kids) == 0 {
			if tr.To == tr.From || !tab.Symbols[tr.Sym].Terminal {
				return "" // the slipper's ;, or an empty rule
			}
			r := s.Raw[s.Tokens[tr.From].Index]
			return src[r.Start:r.End]
		}
		var parts []string
		for _, k := range tr.Kids {
			if p := render(k); p != "" {
				parts = append(parts, p)
			}
		}
		k := tr.Kids
		if len(k) == 3 && len(k[1].Kids) == 0 && tab.Symbols[k[1].Sym].Terminal && len(k[0].Kids) > 0 && len(k[2].Kids) > 0 {
			return "(" + strings.Join(parts, " ") + ")"
		}
		return strings.Join(parts, " ")
	}
	return render(trees[0])
}

// TestLibraryOperatorsAsPerl: with My::Ops in scope each of its operators groups at the bucket toke.c's
// tokentype_for_plugop gives its level, against the core operators either side; before its use, or with no use, it is
// refused, as perl -c refuses a word or an unknown character where an operator is expected.
func TestLibraryOperatorsAsPerl(t *testing.T) {
	const use = "use My::Ops; "
	for src, want := range map[string]string{
		"1 lowop 2 or 3;":        "((1 lowop 2) or 3)", // perly.y's termbinop: a term each side
		"$x lowop $y = 3;":       "($x lowop ($y = 3))",
		"1 orlow 2 and 3;":       "(1 orlow (2 and 3))",
		"1 or 2 orlow 3;":        "((1 or 2) orlow 3)",
		"1 or 2 andlow 3;":       "(1 or (2 andlow 3))",
		"$x andlow $y = 3;":      "($x andlow ($y = 3))",
		"1 and $x assign 2;":     "(1 and ($x assign 2))",
		"$x assign $y || 1;":     "($x assign ($y || 1))",
		"$x assign $y assign 1;": "($x assign ($y assign 1))",
		"$x = 1 oror 2;":         "($x = (1 oror 2))",
		"1 oror 2 && 3;":         "(1 oror (2 && 3))",
		"1 || 2 andand 3;":       "(1 || (2 andand 3))",
		"1 andand 2 | 3;":        "(1 andand (2 | 3))",
		"1 == 2 same 3;":         "(1 == (2 same 3))",
		"1 same 2 + 3;":          "(1 same (2 + 3))",
		"1 plus 2 * 3;":          "(1 plus (2 * 3))",
		"1 * 2 plus 3;":          "((1 * 2) plus 3)",
		"1 plus 2 + 3;":          "((1 plus 2) + 3)",
		"1 + 2 mul 3;":           "(1 + (2 mul 3))",
		"1 mul 2 ** 3;":          "(1 mul (2 ** 3))",
		"1 * 2 pow 3;":           "(1 * (2 pow 3))",
		"1 pow 2 pow 3;":         "(1 pow (2 pow 3))",
		"1 ** 2 high 3;":         "(1 ** (2 high 3))",
		"1 CORE 2;":              "(1 CORE 2)", // a word a class names (PRIVATEREF), still a plugin's: yyl_keylookup
		// The parity limit: recursive descent places ⊕ between + and *, so it groups (1 ⊕ (2 * 3)) there; perl has no
		// bucket between, and tokentype_for_plugop's <= puts it at PLUGIN_MUL_OP, so it groups ((1 ⊕ 2) * 3) here.
		"1 ⊕ 2 * 3;": "((1 ⊕ 2) * 3)",
		// The parity limit at LOW: recursive descent groups ($a lowop ($b or $c)).
		"$a lowop $b or $c;": "(($a lowop $b) or $c)",
		// The parity limit at =='s level (EQUALITY, ORDERING, :equiv(==)): recursive descent groups (($a < $b) same $c).
		"$a < $b same $c;": "($a < ($b same $c))",
	} {
		if got := grouping(t, use+src); !strings.Contains(got, want) {
			t.Errorf("%q: got %s, want %s", use+src, got, want)
		}
	}
	for _, src := range []string{
		"1 plus 2;", "1 ⊕ 2;", "1 plus 2; use My::Ops;", "use My::Ops; 1 nosuch 2;",
		"use My::Ops; 1 plus;", "use My::Ops; plus 2;",
		"use My::Ops; 1 lowop 2 lowop 3;", // %nonassoc
		// BEGIN runs the import at the use's ;, so the operator is not in scope inside its own use.
		"use My::Ops (1 plus 2);", "use My::Ops 1 plus 2;", "BEGIN { require My::Ops; My::Ops->import; 1 plus 2 }",
		// Recursive descent accepts these at =='s level; PLUGIN_REL_OP and PLUGIN_HIGH_OP are %nonassoc in perly.y.
		"use My::Ops; $a same $b < $c;", "use My::Ops; 1 same 2 same 3;", "use My::Ops; 1 high 2 high 3;",
	} {
		if got := grouping(t, src); got != "refused" {
			t.Errorf("%q: got %s, want refused", src, got)
		}
	}
}

// TestLibraryOperatorsLimits pins what the chart accepts and perl -c refuses under an XS module registering My::Ops's
// operators, each a stated shortcut: its verdict changes only with the shortcut.
func TestLibraryOperatorsLimits(t *testing.T) {
	for src, want := range map[string]string{
		// shortcut at forest.Scope.LibraryOperators: file-wide, so a block's end or a no does not end the scope.
		"{ use My::Ops; } 1 plus 2;":             "(1 plus 2)",
		"sub f { use My::Ops; } 1 plus 2;":       "", // accepted, with two derivations as `sub f { } 1 + 2;` has
		"if (1) { use My::Ops; } 1 plus 2;":      "(1 plus 2)",
		"use My::Ops; { no My::Ops; 1 plus 2; }": "(1 plus 2)",
		// shortcut in perly.go: a library operator's word reads as one only where an operator is expected, where
		// toke.c asks the plugin about every word.
		"use My::Ops; my $x = plus;": "(my $x = plus)",
		"use My::Ops; foo(plus);":    "foo ( plus )",
		"use My::Ops; print plus;":   "print plus",
		// shortcut in perly.go: a library operator spelled as a keyword reads as the keyword.
		"use My::Ops; print 1 plus 2;": "print (1 plus 2)",
	} {
		if got := grouping(t, src); got == "refused" || !strings.Contains(got, want) {
			t.Errorf("%q: got %s, want %s", src, got, want)
		}
	}
}

// TestLibraryOperatorsLeaveCoreOperators: a core operator's token admits no PLUGIN_* terminal, whether or not a library
// operator is in scope, and its grouping is perl's under My::Ops.
func TestLibraryOperatorsLeaveCoreOperators(t *testing.T) {
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	const core = "$x = 1 + 2 * 3 ** 4 x 5 . 6 eq 7 == 8 < 9 lt 10 && 11 || 12 // 13 and 14 or 15 xor 16 | 17 & 18 << 19 =~ 20 .. 21 cmp 22 <=> 23;"
	for _, src := range []string{core, "use My::Ops; " + core, "$x x= 2;", "use My::Ops; $x x= 2;"} {
		s, _ := lex.Lex([]byte(src), tab.Tokens)
		for _, tok := range s.Tokens {
			for _, id := range tab.TerminalSets[tok.TerminalSet] {
				if name := tab.Symbols[id].Name; strings.HasPrefix(name, "PLUGIN_") {
					r := s.Raw[tok.Index]
					t.Errorf("%q: %q admits %s", src, src[r.Start:r.End], name)
				}
			}
		}
	}
	if got, want := grouping(t, "use My::Ops; 1 + 2 * 3;"), "(1 + (2 * 3))"; !strings.Contains(got, want) {
		t.Errorf("1 + 2 * 3 under My::Ops: got %s, want %s", got, want)
	}
}

// TestLibraryOperatorsWithoutDeclaration: a use of a module that cannot be loaded, or that declares no infix operator,
// brings in no PLUGIN_* reading and does not panic; a sub of the operator's name is a call, which perl -c refuses there.
func TestLibraryOperatorsWithoutDeclaration(t *testing.T) {
	for _, src := range []string{
		"use No::Such::Module; 1 plus 2;", "use My::Plain; 1 plus 2;", "sub plus {} 1 plus 2;", "use My::Ops (); 1 plus 2;",
	} {
		if got := grouping(t, src); got != "refused" {
			t.Errorf("%q: got %s, want refused", src, got)
		}
	}
	if got := grouping(t, "use My::Plain; plus 1, 2;"); got == "refused" {
		t.Error("use My::Plain; plus 1, 2; is refused: My::Plain's plus is a call")
	}
}
