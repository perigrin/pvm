// ABOUTME: Named unaries and list operators: where precedence stops being a table.
// ABOUTME: Every classification here was measured by deparsing, not transcribed from the spec.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestNamedUnaryPrecedence: level 19 — tighter than comparison, looser than
// arithmetic.
//
// Measured on the optree, because Deparse round-trips the surface form and
// hides the grouping:
//
//	$ perl -MO=Concise -e 'my $x="abc"; print length $x + 1'
//	a  <1> length[t3] sK/1 ->b
//	9     <2> add[t2] sK/2 ->a          length(add($x, 1))
//
// So the `+` binds tighter and ends up INSIDE the length.
func TestNamedUnaryPrecedence(t *testing.T) {
	root := parse.Parse([]byte("length $x + 1;"))
	if containsKind(root, parse.Unknown) {
		t.Fatalf("must parse: %v", kinds(root))
	}
	call := firstOfKind(root, parse.Call)
	if call == nil {
		t.Fatalf("no Call node: %v", kinds(root))
	}
	// The addition is the argument, not the other way round.
	if firstOfKind(call, parse.Binary) == nil {
		t.Errorf("`length $x + 1` is length($x + 1): %v", kinds(call))
	}

	// And looser than comparison: `length $x < 5` is `length($x) < 5`.
	// Measured: perl -MO=Deparse gives `length $x < 5`, and the optree puts
	// lt outermost.
	root = parse.Parse([]byte("length $x < 5;"))
	if containsKind(root, parse.Unknown) {
		t.Fatalf("must parse: %v", kinds(root))
	}
	top := root.Children[0].Children[0]
	if top.Kind != parse.CmpChain && top.Kind != parse.Binary {
		t.Errorf("the comparison is outermost: got %v", top.Kind)
	}
}

// TestNamedUnaryParenForm: the paren cliff of §4.8.1.
//
// `toke.c`'s UNI3 returns FUNC1 when the next character is `(` and UNIOP
// otherwise, so the two forms group differently. Measured:
//
//	print length $x + 1      length[t3] outermost   -> length($x + 1)
//	print length ($x) + 1    add[t3] outermost      -> length($x) + 1
func TestNamedUnaryParenForm(t *testing.T) {
	withParens := parse.Parse([]byte("length ($x) + 1;"))
	if containsKind(withParens, parse.Unknown) {
		t.Fatalf("paren form must parse: %v", kinds(withParens))
	}
	// The `+` is OUTSIDE the call here.
	top := withParens.Children[0].Children[0]
	if top.Kind != parse.Binary || top.Text != "+" {
		t.Errorf("`length ($x) + 1` is (length($x)) + 1, so `+` is outermost: "+
			"got %v %q", top.Kind, top.Text)
	}

	// Without parens the call is outermost, asserted above.
	noParens := parse.Parse([]byte("length $x + 1;"))
	topNo := noParens.Children[0].Children[0]
	if topNo.Kind == parse.Binary && topNo.Text == "+" {
		t.Error("`length $x + 1` is length($x + 1); `+` must NOT be outermost")
	}
}

// TestListOperatorExtent: swallows to the statement end, or to a closing
// paren.
//
// Measured:
//
//	perl -MO=Deparse -e 'print join ",", @a, "x";'
//	print join(',', @a, 'x');              // everything
//
//	perl -MO=Deparse -e 'my $s = (join ",", @a) . "!";'
//	my $s = join(',', @a) . '!';           // stopped at the paren
func TestListOperatorExtent(t *testing.T) {
	root := parse.Parse([]byte(`print $a, $b, $c;`))
	if containsKind(root, parse.Unknown) {
		t.Fatalf("must parse: %v", kinds(root))
	}
	call := firstOfKind(root, parse.Call)
	if call == nil {
		t.Fatalf("no Call: %v", kinds(root))
	}
	// All three arguments belong to the call.
	if n := countKind(call, parse.Term); n < 3 {
		t.Errorf("a list operator takes the whole list: got %d terms in the "+
			"call, want 3: %v", n, kinds(call))
	}

	// A closing paren stops it.
	root = parse.Parse([]byte(`my $s = (join $a, $b) . $c;`))
	if containsKind(root, parse.Unknown) {
		t.Fatalf("must parse: %v", kinds(root))
	}
	call = firstOfKind(root, parse.Call)
	if call == nil {
		t.Fatalf("no Call: %v", kinds(root))
	}
	if call.End > len("my $s = (join $a, $b)") {
		t.Errorf("the list operator ran past its closing paren: ends at %d",
			call.End)
	}
}

// TestFeatureGatedKeywords: §0.13 rank 6, 5 corpus files.
//
// `try`, `defer`, `method` and `true` are keywords only where their feature
// is enabled; elsewhere they are ordinary identifiers. Measured:
//
//	$ perl -e 'sub try { 42 } print try(), "\n"'      42
//	$ perl -e 'my %h = (method => 1); print $h{method}'   1
//
// A table that does not know this turns valid old code into a parse error.
func TestFeatureGatedKeywords(t *testing.T) {
	for _, src := range []string{
		"my $x = $h{method};",
		"my $x = $h{try};",
		"my %h = (defer => 1);",
		"my $x = try();",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q uses a feature-gated word as an identifier and must "+
				"parse: %v", src, kinds(root))
		}
	}
}

// TestUnknownSubIsUnresolvedCall: §4.8.3's rule.
//
// A call to a sub this parser has never seen is a Call with Resolved false,
// NOT an Unknown. The spec is explicit:
//
//	The expression parser's job is to produce a `Call` node with
//	`Resolved: false` and let a later pass decide.
//
// Unknown is for constructs with no known shape. A call has a known shape and
// an unknown callee, which is a different thing and scores differently: the
// adapter reports it as an Unresolved site, which is `wider`, not WRONG.
func TestUnknownSubIsUnresolvedCall(t *testing.T) {
	root := parse.Parse([]byte("zzznotakeyword($x);"))
	if containsKind(root, parse.Unknown) {
		t.Fatalf("an unknown sub call is a Call, not Unknown: %v", kinds(root))
	}
	call := firstOfKind(root, parse.Call)
	if call == nil {
		t.Fatalf("no Call node: %v", kinds(root))
	}
	if call.Resolved {
		t.Error("a call to an undeclared sub must be Resolved:false")
	}

	// A known builtin IS resolved.
	root = parse.Parse([]byte("length($x);"))
	call = firstOfKind(root, parse.Call)
	if call == nil {
		t.Fatalf("no Call node for a builtin: %v", kinds(root))
	}
	if !call.Resolved {
		t.Error("a builtin call is Resolved:true")
	}
}

// TestAmpersandCallForms: the four shapes of §4.8.3.
//
//	foo(@args)     normal call, prototype applied
//	&foo(@args)    call, prototype IGNORED
//	&foo           call, passes the CALLER'S @_ through
//	\&foo          a code reference, no call
func TestAmpersandCallForms(t *testing.T) {
	for _, src := range []string{
		"foo($x);",
		"&foo($x);",
		"&foo;",
		"my $r = \\&foo;",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}

	// `\&foo` is a reference, not a call.
	root := parse.Parse([]byte("my $r = \\&foo;"))
	if u := firstOfKind(root, parse.Unary); u == nil || u.Text != "ref" {
		t.Errorf("`\\&foo` is a reference: %v", kinds(root))
	}
}
