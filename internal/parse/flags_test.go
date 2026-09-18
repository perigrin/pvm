// ABOUTME: Arrow, Paren, Fat and Handle: four facts the parser knew and had nowhere to write down.
// ABOUTME: Each is a §4.14.2 row marked PARSER because of the field list, not because the parse is hard.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// find returns the first node of kind k in the tree, or nil.
func find(n *parse.Node, k parse.Kind) *parse.Node {
	if n.Kind == k {
		return n
	}
	for _, c := range n.Children {
		if got := find(c, k); got != nil {
			return got
		}
	}
	return nil
}

// findAll returns every node of kind k.
func findAll(n *parse.Node, k parse.Kind) []*parse.Node {
	var out []*parse.Node
	if n.Kind == k {
		out = append(out, n)
	}
	for _, c := range n.Children {
		out = append(out, findAll(c, k)...)
	}
	return out
}

// TestArrowBitDistinguishesSubscripts: `$h{k}` and `$h->{k}` are one
// operation with one node (§4.14's `Index{Base,Idx,Arrow}`), and the arrow
// is the field that separates them.
//
// They read DIFFERENT VARIABLES, which is why the bit is not cosmetic.
// Measured on perl 5.42.0:
//
//	%h = (k => 'hash'); $h = {k => 'ref'};
//	print $h{k};     # hash    -- the hash %h
//	print $h->{k};   # ref     -- the hashref $h
//
// `internal/parse/expr.go` already promised the flag in a comment -- "an
// `Arrow` flag, not two unrelated shapes" -- while `Node` had nowhere to put
// it. 25.3% of T1's index nodes are the arrow form.
func TestArrowBitDistinguishesSubscripts(t *testing.T) {
	plain := find(parse.Parse([]byte(`$h{k};`)), parse.Index)
	arrow := find(parse.Parse([]byte(`$h->{k};`)), parse.Index)
	if plain == nil || arrow == nil {
		t.Fatalf("both forms must produce an Index: plain=%v arrow=%v",
			plain, arrow)
	}
	if plain.Arrow {
		t.Error("`$h{k}` has no arrow")
	}
	if !arrow.Arrow {
		t.Error("`$h->{k}` has an arrow, and the tree must say so")
	}

	for _, c := range []struct {
		src  string
		want bool
	}{
		{`$a[0];`, false},
		{`$a->[0];`, true},
		{`$h{a}{b};`, false},
		{`f()->{k};`, true},
	} {
		idx := find(parse.Parse([]byte(c.src)), parse.Index)
		if idx == nil {
			t.Errorf("%q: no Index node", c.src)
			continue
		}
		if idx.Arrow != c.want {
			t.Errorf("%q: Arrow = %v, want %v", c.src, idx.Arrow, c.want)
		}
	}

	// In a CHAIN only the first subscript has an arrow, and the flag is per
	// node rather than per statement. `$x->{a}[0]` is two Index nodes: the
	// `{a}` is reached through `->`, the `[0]` is not. perl agrees -- the
	// arrow after a subscript is optional and almost never written.
	chain := findAll(parse.Parse([]byte(`$x->{a}[0];`)), parse.Index)
	if len(chain) != 2 {
		t.Fatalf("`$x->{a}[0]` has %d Index nodes, want 2", len(chain))
	}
	// findAll is pre-order, so the outer `[0]` comes first.
	if chain[0].Text != "[" || chain[0].Arrow {
		t.Errorf("the outer `[0]` is not reached through an arrow; "+
			"got %q arrow=%v", chain[0].Text, chain[0].Arrow)
	}
	if chain[1].Text != "{" || !chain[1].Arrow {
		t.Errorf("the inner `{a}` IS reached through an arrow; "+
			"got %q arrow=%v", chain[1].Text, chain[1].Arrow)
	}
}

// TestParenBitSurvives: `($x)` is distinguishable from `$x` without
// re-deriving anything from spans.
//
// `finishList` returns the inner node with a widened span, so the only
// evidence was the span itself -- and a census found 721 T1 nodes whose span
// starts `(` with a first child at the same offset, so the naive span test
// has 721 false positives.
func TestParenBitSurvives(t *testing.T) {
	for _, c := range []struct {
		src  string
		want bool
	}{
		{`my $x = ($y);`, true},
		{`my $x = $y;`, false},
		{`my $x = (($y));`, true},
		{`f($y);`, false}, // a call's parens are the call's, not a group
	} {
		root := parse.Parse([]byte(c.src))
		var got bool
		for _, n := range findAll(root, parse.Term) {
			if n.Paren {
				got = true
			}
		}
		if got != c.want {
			t.Errorf("%q: a parenthesised term is %v, want %v: %v",
				c.src, got, c.want, kinds(root))
		}
	}
}

// TestParenBitCarriesContext is why the bit exists rather than being
// cosmetic: two of Perl's rules turn on it.
//
// §4.10, list repeat versus string repeat. Measured:
//
//	$ perl -e 'my @a = ("a") x 3; print scalar @a'   3
//	$ perl -e 'my @a = "a" x 3;   print scalar @a'   1
//
// §4.12.2, list versus scalar assignment. Measured with `sub f {(1,2,3)}`:
//
//	my ($x) = f();   $x is 1, the first element
//	my $y = f();     $y is 3, the last
func TestParenBitCarriesContext(t *testing.T) {
	// `('a') x 3` -- the left operand is parenthesised, so this is a list
	// repeat. `'a' x 3` is a string repeat.
	listRepeat := parse.Parse([]byte(`my @a = ('a') x 3;`))
	strRepeat := parse.Parse([]byte(`my @a = 'a' x 3;`))

	leftOf := func(root *parse.Node) *parse.Node {
		for _, b := range findAll(root, parse.Binary) {
			if b.Text == "x" && len(b.Children) > 0 {
				return b.Children[0]
			}
		}
		return nil
	}
	l, s := leftOf(listRepeat), leftOf(strRepeat)
	if l == nil || s == nil {
		t.Fatalf("both must produce a Binary \"x\": list=%v str=%v", l, s)
	}
	if !l.Paren {
		t.Error("`('a') x 3`: the left operand is parenthesised, which is what " +
			"makes it a LIST repeat (§4.10)")
	}
	if s.Paren {
		t.Error("`'a' x 3`: the left operand is not parenthesised")
	}

	// `my ($x) = f()` versus `my $x = f()`. The list form's target is
	// parenthesised, and that is the whole difference.
	listDecl := parse.Parse([]byte(`my ($x) = f();`))
	scalarDecl := parse.Parse([]byte(`my $x = f();`))

	var listParen, scalarParen bool
	for _, n := range findAll(listDecl, parse.Term) {
		if n.Paren {
			listParen = true
		}
	}
	for _, n := range findAll(scalarDecl, parse.Term) {
		if n.Paren {
			scalarParen = true
		}
	}
	if !listParen {
		t.Errorf("`my ($x) = f()`: the target is parenthesised, which is what "+
			"makes it a LIST assignment (§4.12.2): %v", kinds(listDecl))
	}
	if scalarParen {
		t.Errorf("`my $x = f()`: the target is not parenthesised: %v",
			kinds(scalarDecl))
	}
}

// TestFatBitSurvivesInAggregates: `(a => 1)` and `(a, 1)` must differ.
//
// §4.5.4: `=>` quotes the word to its left, so without the bit the lowering
// cannot tell a string from a call. `parseList` consumed `,` and `=>` alike,
// so inside a list, anon array or anon hash the fact was destroyed -- while
// it survived in call arguments, which made the loss inconsistent as well as
// lossy.
func TestFatBitSurvivesInAggregates(t *testing.T) {
	for _, c := range []struct {
		src  string
		kind parse.Kind
		want bool
	}{
		{`my @x = (a => 1);`, parse.List, true},
		{`my @x = (a, 1);`, parse.List, false},
		{`my $r = [a => 1];`, parse.AnonArray, true},
		{`my $r = [a, 1];`, parse.AnonArray, false},
		{`my $r = {a => 1};`, parse.AnonHash, true},
		{`my $r = {a, 1};`, parse.AnonHash, false},
	} {
		root := parse.Parse([]byte(c.src))
		agg := find(root, c.kind)
		if agg == nil {
			t.Errorf("%q: no %v node: %v", c.src, c.kind, kinds(root))
			continue
		}
		var got bool
		for _, child := range agg.Children {
			if child.Fat {
				got = true
			}
		}
		if got != c.want {
			t.Errorf("%q: a fat-comma element is %v, want %v",
				c.src, got, c.want)
		}
	}
}

// TestHandleBitIsSet: `print $fh "x"` marks its handle, so the lowering does
// not have to count children to find it.
//
// `parseFilehandleSlot` decides the slot at construction and emitted a plain
// Term, so the only evidence was child shape: two children means a handle,
// one means a comma list. `internal/infer/infer.go:1155-1165` records what
// that cost -- counting the handle as argument 1 made every typed-handle
// print a false Str mismatch.
func TestHandleBitIsSet(t *testing.T) {
	for _, c := range []struct {
		src  string
		want bool
	}{
		{`print $fh "x";`, true},
		{`print STDERR "x";`, true},
		{`print $x, "x";`, false}, // a comma list, not a handle
		{`print "x";`, false},
	} {
		root := parse.Parse([]byte(c.src))
		call := find(root, parse.Call)
		if call == nil {
			t.Errorf("%q: no Call node: %v", c.src, kinds(root))
			continue
		}
		var got bool
		for _, child := range call.Children {
			if child.Handle {
				got = true
			}
		}
		if got != c.want {
			t.Errorf("%q: a marked handle is %v, want %v: %v",
				c.src, got, c.want, kinds(root))
		}
	}
}

// TestFlagsAreFalseWhereInapplicable is the negative scenario for all four.
//
// `Resolved` set the precedent -- its doc says "only meaningful for Call;
// false everywhere else and not read there" -- and a flag that leaks onto
// nodes it does not describe is worse than no flag, because a consumer
// cannot tell a real answer from a stray one.
func TestFlagsAreFalseWhereInapplicable(t *testing.T) {
	const src = `my $x = 1; my @a = (1, 2); if ($x) { f(); }`

	root := parse.Parse([]byte(src))
	for _, n := range allNodes(root) {
		switch n.Kind {
		case parse.Index:
			// Arrow is meaningful here; the others are not.
			if n.Paren || n.Fat || n.Handle {
				t.Errorf("Index carries a flag that is not its own: "+
					"Paren=%v Fat=%v Handle=%v", n.Paren, n.Fat, n.Handle)
			}
		default:
			if n.Arrow {
				t.Errorf("%v carries Arrow, which only an Index may", n.Kind)
			}
		}
	}
}
