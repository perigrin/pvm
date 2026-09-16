// ABOUTME: Control flow: conditionals, the three loop keywords, both for forms, and labels.
// ABOUTME: A label is not an expression — `LOOP:` and `$h{LOOP}` and `LOOP ? 1 : 0` share a shape.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestConditionalChains: if/elsif/else, including an elsif with no else.
func TestConditionalChains(t *testing.T) {
	for _, src := range []string{
		"if ($c) { 1 }",
		"if ($c) { 1 } else { 2 }",
		"if ($c) { 1 } elsif ($d) { 2 }",
		"if ($c) { 1 } elsif ($d) { 2 } else { 3 }",
		"if ($c) { 1 } elsif ($d) { 2 } elsif ($e) { 3 } else { 4 }",
		"if ($c) { }",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
			continue
		}
		if firstOfKind(root, parse.Conditional) == nil {
			t.Errorf("%q must produce a Conditional: %v", src, kinds(root))
		}
	}
}

// TestNegatedAndLoopForms: unless, while and until share the block shape.
func TestNegatedAndLoopForms(t *testing.T) {
	for _, tc := range []struct {
		src  string
		kind parse.Kind
	}{
		{"unless ($c) { 1 }", parse.Conditional},
		{"unless ($c) { 1 } else { 2 }", parse.Conditional},
		{"while ($c) { 1 }", parse.Loop},
		{"until ($c) { 1 }", parse.Loop},
		{"while (1) { }", parse.Loop},
	} {
		root := parse.Parse([]byte(tc.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", tc.src, kinds(root))
			continue
		}
		if firstOfKind(root, tc.kind) == nil {
			t.Errorf("%q must produce a %v: %v", tc.src, tc.kind, kinds(root))
		}
	}
}

// TestForLoopForms: the C-style three-part head and the list form share a
// keyword and nothing else structurally.
func TestForLoopForms(t *testing.T) {
	for _, src := range []string{
		"for (my $i = 0; $i < 10; $i++) { 1 }",
		"for (;;) { 1 }",
		"for my $x (@list) { 1 }",
		"for (@list) { 1 }",
		"foreach my $x (@list) { 1 }",
		"foreach (@list) { 1 }",
		"for my $x (1 .. 10) { 1 }",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
			continue
		}
		if firstOfKind(root, parse.Loop) == nil {
			t.Errorf("%q must produce a Loop: %v", src, kinds(root))
		}
	}
}

// TestLoopLabels: a label is not an expression.
//
// `LOOP:` and `$h{LOOP}` and `LOOP ? 1 : 0` all put a bareword next to a
// colon; only the first is a label, and the difference is whether a statement
// could start here.
//
// Labels also STACK (spec §5, perly.y labfullstmt), and only the innermost
// attaches to the loop. Measured:
//
//	$ perl -MO=Deparse -e 'A: B: for (1..3){ last B }'
//	A: B: ;
//	foreach $_ (1 .. 3) { last B; }
//
//	$ perl -e 'A: B: for (1..3){ last A }'
//	Label not found for "last A"
//
// So every label is kept -- hover and go-to-definition need them -- and
// resolution against the innermost is a later concern.
func TestLoopLabels(t *testing.T) {
	root := parse.Parse([]byte("LOOP: while ($c) { last LOOP }"))
	if containsKind(root, parse.Unknown) {
		t.Fatalf("a labelled loop must parse: %v", kinds(root))
	}
	if firstOfKind(root, parse.Label) == nil {
		t.Errorf("LOOP: must produce a Label: %v", kinds(root))
	}

	// Stacked labels: all of them are kept, as siblings of what they label.
	//
	// Counted on the statement's OWN children rather than over the whole
	// tree: `last B` carries a Label too, and a recursive count would find
	// three and call it a pass for the wrong reason.
	root = parse.Parse([]byte("A: B: for (1 .. 3) { last B }"))
	if containsKind(root, parse.Unknown) {
		t.Fatalf("stacked labels must parse: %v", kinds(root))
	}
	stmt := root.Children[0]
	own := 0
	for _, c := range stmt.Children {
		if c.Kind == parse.Label {
			own++
		}
	}
	if own != 2 {
		t.Errorf("got %d Label nodes on the statement for `A: B:`, want 2: %v",
			own, kinds(root))
	}

	// The same bareword-then-colon shape in EXPRESSION position is not a
	// label. This is the assertion that fails if the parser looks for a
	// colon rather than asking whether a statement could start here.
	for _, src := range []string{
		"my $x = $h{LOOP};",
		"my $x = $c ? 1 : 0;",
	} {
		root := parse.Parse([]byte(src))
		if firstOfKind(root, parse.Label) != nil {
			t.Errorf("%q has no label, but one was parsed: %v", src, kinds(root))
		}
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must still parse: %v", src, kinds(root))
		}
	}
}

func countKind(n *parse.Node, k parse.Kind) int {
	total := 0
	if n.Kind == k {
		total++
	}
	for _, c := range n.Children {
		total += countKind(c, k)
	}
	return total
}
