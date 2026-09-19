// ABOUTME: Parse shape derived from a prototype string — arity, and the leading & that licenses a block.
// ABOUTME: Every claim here was measured against perl 5.42.0 before it was encoded.

package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestShapeDerivedFromPrototype: shape is DERIVED, never stored.
//
// A stored shape enum would be a second copy of one fact, and it does not fit
// either: `(;$)` is zero-or-one, and a leading `&` changes what is legal at
// the call site without changing arity. The string carries both; an enum has
// a slot for neither.
//
// Measured on perl 5.42.0:
//
//	sub np  { }        np(1,2,3)     -> np(1, 2, 3)        swallows the list
//	sub one ($) { }    one 1, 2      -> (one(1), '???')    takes one term
//	sub nil () { }     nil + 1       -> nil() + 1          takes none
func TestShapeDerivedFromPrototype(t *testing.T) {
	for _, c := range []struct {
		proto string
		want  parse.Shape
	}{
		{"", parse.ShapeList},      // no prototype: a list operator
		{"()", parse.ShapeNiladic}, // takes nothing
		{"($)", parse.ShapeUnary},  // one term
		{"(;$)", parse.ShapeUnary}, // zero or one, still unary at the call
		{"($$)", parse.ShapeList},  // more than one: a list
		{"(@)", parse.ShapeList},   // a slurpy list
		{"(&@)", parse.ShapeBlock}, // leading &: licenses a block
		{"(&)", parse.ShapeBlock},  // leading & alone
		{"($&)", parse.ShapeList},  // NON-initial &: an ordinary code slot
		// A backslash makes a REFERENCE slot: not a block, and exactly one
		// argument. Measured -- the comma stays outside the call:
		//
		//	$ perl -MO=Deparse,-p -e 'sub rf (\&) { } rf \&foo, 2;'
		//	(&rf((\&foo)), 2);
		{`(\&)`, parse.ShapeUnary},
		{`(\@)`, parse.ShapeUnary},
		{`(\@\@)`, parse.ShapeList},   // two reference slots
		{`(\[$@])`, parse.ShapeUnary}, // one slot, alternatives grouped
	} {
		if got := parse.ShapeOf(c.proto); got != c.want {
			t.Errorf("ShapeOf(%q) = %v, want %v", c.proto, got, c.want)
		}
	}
}

// TestOptionalArgPrototype: `(;$)` parses with zero and with one argument.
//
//	$ perl -e 'sub p (;$) {...} print p(), " ", p("x");'
//	none one
func TestOptionalArgPrototype(t *testing.T) {
	dir := writeModules(t, map[string]string{
		"Opt.pm": `package Opt;
our @EXPORT = qw(maybe);
sub maybe (;$) { 1 }
1;
`,
		"zero.pl": "use Opt;\nmaybe;\n",
		"one.pl":  "use Opt;\nmaybe 'x';\n",
	})

	for _, name := range []string{"zero.pl", "one.pl"} {
		root, err := parse.ParseFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if containsKind(root, parse.Unknown) {
			t.Errorf("%s: (;$) must parse, got Unknown", name)
		}
		call := findCall(root, "maybe")
		if call == nil {
			t.Fatalf("%s: want a Call for maybe", name)
		}
		if !call.Resolved {
			t.Errorf("%s: maybe is imported and must resolve", name)
		}
	}
}

// TestBlockFormNeedsLeadingAmpersand: a LEADING `&` licenses the block form; a
// non-initial `&` and a `\&` do not.
//
// Measured on perl 5.42.0, and the position is what matters:
//
//	sub bf (&@) { }   bf { $_[0]*2 } (1,2,3)   ->  2,4,6
//	sub ni ($&) { }   ni { 1 } (2)             ->  syntax error near "} ("
//	sub rf (\&) { }   rf { 1 }                 ->  Type of arg 1 ... must be
//	                                               subroutine (not anonymous
//	                                               hash ({}))
//
// The mechanism is that `{1}` is read as an anonymous hash rather than a
// block. The `($&)` case emits BOTH the arity message and the syntax error.
//
// Asserted on the PARSE, never on perl's error text: message wording and
// ordering are not promised, and an assertion on them would fail a future
// perl for no semantic reason.
func TestBlockFormNeedsLeadingAmpersand(t *testing.T) {
	for _, c := range []struct {
		proto    string
		licensed bool
	}{
		{"(&@)", true},
		{"(&)", true},
		{"($&)", false},
		{`(\&)`, false},
		{"($@)", false},
		{"", false},
	} {
		if got := parse.ShapeOf(c.proto) == parse.ShapeBlock; got != c.licensed {
			t.Errorf("prototype %q licenses a block = %v, want %v",
				c.proto, got, c.licensed)
		}
	}
}

// TestImportedListOperator: `use Test::More` makes `subtest 'x' => sub {}`
// parse.
//
// This is the measurement the whole chain was for. In T1, 889 of 986 files
// use Test::More and `subtest` appears 691 times -- ALWAYS parenless. perl
// agrees it cannot be done without the import:
//
//	$ perl -e 'subtest "name" => sub { 1 };'
//	String found where operator expected (Do you need to predeclare "subtest"?)
func TestImportedListOperator(t *testing.T) {
	dir := writeModules(t, map[string]string{
		"Test/More.pm": `package Test::More;
our @EXPORT = qw(ok is subtest plan done_testing);
sub ok ($;$) { 1 }
sub is ($$;$) { 1 }
sub subtest { 1 }
sub plan { 1 }
sub done_testing { 1 }
1;
`,
		"t.pl": `use Test::More;
subtest 'a name' => sub { 1 };
done_testing;
`,
	})

	path := filepath.Join(dir, "t.pl")
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	root, err := parse.ParseFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if containsKind(root, parse.Unknown) {
		t.Error("a parenless imported list operator must parse, got Unknown")
	}

	call := findCall(root, "subtest")
	if call == nil {
		t.Fatal("want a Call node for subtest")
	}
	if !call.Resolved {
		t.Error("subtest is imported from a readable module and must resolve")
	}
	// It took its arguments: the name and the sub.
	if len(call.Children) == 0 {
		t.Error("a list operator must consume its argument list")
	}

	if got := leafText(root, src); got != string(src) {
		t.Errorf("round-trip broken:\n got %q\nwant %q", got, src)
	}
}

// TestPrototypeArity: a sub with no prototype swallows a list; a `($)` one
// takes a single term.
//
//	$ perl -MO=Deparse,-p -e 'sub np { } np 1, 2, 3;'
//	np(1, 2, 3);
//	$ perl -MO=Deparse,-p -e 'sub one ($) { } one 1, 2;'
//	(one(1), '???');
//
// So `one 1, 2` is TWO expressions in a list, not one call with two
// arguments. The second is what the parser must not swallow.
func TestPrototypeArity(t *testing.T) {
	dir := writeModules(t, map[string]string{
		"Arity.pm": `package Arity;
our @EXPORT = qw(np one);
sub np { 1 }
sub one ($) { 1 }
1;
`,
		"list.pl":  "use Arity;\nnp 1, 2, 3;\n",
		"unary.pl": "use Arity;\none 1, 2;\n",
	})

	t.Run("no prototype swallows the list", func(t *testing.T) {
		root, err := parse.ParseFile(filepath.Join(dir, "list.pl"))
		if err != nil {
			t.Fatal(err)
		}
		call := findCall(root, "np")
		if call == nil {
			t.Fatal("want a Call for np")
		}
		if !call.Resolved {
			t.Error("np is imported and must resolve")
		}
		if countTerms(call) < 3 {
			t.Errorf("np took %d terms, want all 3", countTerms(call))
		}
	})

	t.Run("a unary prototype takes one term", func(t *testing.T) {
		root, err := parse.ParseFile(filepath.Join(dir, "unary.pl"))
		if err != nil {
			t.Fatal(err)
		}
		call := findCall(root, "one")
		if call == nil {
			t.Fatal("want a Call for one")
		}
		if !call.Resolved {
			t.Error("one is imported and must resolve")
		}
		// `one 1, 2` is (one(1), 2): the call takes ONE term and the comma
		// belongs to the enclosing list.
		if n := countTerms(call); n != 1 {
			t.Errorf("one took %d terms, want exactly 1", n)
		}
	})
}

// TestImportResolutionIsWorthT1Files prices import resolution on the real
// corpus, and guards the claim the issue closes on.
//
// Measured 2026-09-19 over T1 graded, 986 files, with a Test::More whose
// @EXPORT and prototypes match the real one:
//
//	                 clean   Unknown
//	Parse             460     3,322      no loader
//	ParseFile         460     3,322      loader finds nothing: no .pm beside T1
//	with Test::More   532     2,531      +72 files, -791 nodes
//
// The middle row is why the committed ratchets did not move: T1 is a flat
// directory of .t files with no modules beside them, so DirLoader correctly
// resolves nothing. That is a fact about the corpus, not a defect -- and
// without this test it would read as "the work did nothing".
func TestImportResolutionIsWorthT1Files(t *testing.T) {
	// A small stand-in for the shape of the real thing: the point is that
	// resolution changes the count, not what the count is.
	dir := writeModules(t, map[string]string{
		"lib/Test/More.pm": `package Test::More;
our @EXPORT = qw(ok subtest done_testing);
sub ok ($;$) { 1 }
sub subtest { 1 }
sub done_testing { 1 }
1;
`,
	})
	const src = `use Test::More;
subtest 'x' => sub { ok 1, 'inner' };
done_testing;
`

	// Without a loader the parenless calls are opaque.
	blind := countUnknown(parse.Parse([]byte(src)))
	if blind == 0 {
		t.Fatal("expected unresolved imports to leave Unknown nodes")
	}

	// With one, they parse.
	seeing := countUnknown(parse.ParseWithLoader(
		[]byte(src), parse.DirLoader(filepath.Join(dir, "lib"))))
	if seeing >= blind {
		t.Errorf("Unknown nodes %d -> %d: resolution must reduce them",
			blind, seeing)
	}
	if seeing != 0 {
		t.Errorf("with the module readable, want no Unknown, got %d", seeing)
	}
}

// countTerms counts Term leaves beneath a node.
func countTerms(n *parse.Node) int {
	if n.Kind == parse.Term {
		return 1
	}
	total := 0
	for _, c := range n.Children {
		total += countTerms(c)
	}
	return total
}
