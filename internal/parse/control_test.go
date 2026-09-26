// ABOUTME: Control flow: conditionals, the three loop keywords, both for forms, and labels.
// ABOUTME: A label is not an expression — `LOOP:` and `$h{LOOP}` and `LOOP ? 1 : 0` share a shape.

package parse_test

import (
	"strings"
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

// TestLabelledBareBlock: `SKIP: { ... }` is a statement, and the label is kept.
//
// A bare block parses and a labelled loop parses; a labelled BARE BLOCK did
// not, and it is the idiom perl's own suite uses for skipping -- present in at
// least 152 of perl.git t/'s 620 files. The issue counted 148; any single-line
// count undercounts, because `SKIP:` and its `{` are often on separate lines
// (t/io/crlf.t:36, t/comp/parser_run.t:73).
//
// The label is the whole of it. Measured at 9325864f:
//
//	SKIP: { print 1; }    Unknown=1, and the `{` read as an ANON HASH
//	{ print 1; }          Unknown=0
//
// The cause is in the LEXER, not the parser: `OpensBlock` was false on a `{`
// whose previous significant token is a label's `:`, because Operator left
// XTerm. perl's yyl_colon reaches PREBLOCK for a label instead.
//
// But a label does NOT force a block, and this is the assertion that fails if
// the fix returns XBlock unconditionally. perl runs intuit_curly after a label
// too, measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'L: {a=>1};'
//	L: +{'a', 1};
//
// A labelled anonymous HASH, in statement position.
//
// All three loop controls target a bare block's label, measured on 5.42.0:
//
//	perl -e 'SKIP: { print "a\n"; last SKIP; print "b\n"; }'          -> a
//	perl -e 'SKIP: { print "a\n"; next SKIP; print "b\n"; }'          -> a
//	perl -e 'my $n=0; SKIP: { $n++; redo SKIP if $n < 3; }'           -> 3 passes
//
// and the label's case does not matter -- `skip: { ... }` runs, so unlike
// `given`/`when` this rule is general over any word.
func TestLabelledBareBlock(t *testing.T) {
	for _, src := range []string{
		"SKIP: { print 1; }",
		"SKIP: { last SKIP; }",
		"SKIP: { next SKIP; }",
		"SKIP: { redo SKIP; }",
		"skip: { print 1; }",
		"A: B: { print 1; }",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
			continue
		}
		if firstOfKind(root, parse.Label) == nil {
			t.Errorf("%q must produce a Label: %v", src, kinds(root))
		}
		// Zero Unknowns is not enough -- `given (1) { }` came back
		// Unknown=0 and WRONG as an index call, and `SKIP: { print 1; }`
		// came back with the `{` as an AnonHash. The block must BE a block.
		if firstOfKind(root, parse.Block) == nil {
			t.Errorf("%q must produce a Block, not a hash or a subscript: %v",
				src, kinds(root))
		}
		if firstOfKind(root, parse.AnonHash) != nil {
			t.Errorf("%q read its block as an anonymous hash: %v", src, kinds(root))
		}
	}

	// A label on an EMPTY statement. perl accepts it and Deparse emits it
	// back verbatim:
	//
	//	$ perl -MO=Deparse -e 'my $x; HERE: ; goto HERE if !$x++;'
	//	my $x;
	//	HERE: ;
	//
	// It is in perl's own suite at t/class/field.t:289. Before the fix the
	// label bytes were in NO node -- the empty-statement check ran before the
	// labels were read, so the `;` reached the expression parser, refused, and
	// the Unknown started at the `;` -- which made canon of `HERE: ;` equal
	// `;`, six bytes short, and its own canon differ again.
	if root := parse.Parse([]byte("HERE: ;")); containsKind(root, parse.Unknown) {
		t.Errorf("`HERE: ;` must parse: %v", kinds(root))
	} else if firstOfKind(root, parse.Label) == nil {
		t.Errorf("`HERE: ;` must keep its Label: %v", kinds(root))
	} else if got := parse.Canon(root, []byte("HERE: ;")); !strings.Contains(got, "HERE:") {
		t.Errorf("canon of `HERE: ;` dropped the label: %q", got)
	}

	// A label does not make the next `{` a block. intuit_curly still decides.
	root := parse.Parse([]byte("L: { a => 1 };"))
	if containsKind(root, parse.Unknown) {
		t.Errorf("`L: { a => 1 };` must parse: %v", kinds(root))
	}
	if firstOfKind(root, parse.AnonHash) == nil {
		t.Errorf("`L: { a => 1 };` is a labelled anon hash, not a block: %v", kinds(root))
	}

	// Canon must re-emit the label. A canon that drops it is a canon of
	// different source, and the fixpoint then hides it because canon of canon
	// is stable at the WRONG text.
	for _, src := range []string{"SKIP: { print 1; }", "SKIP: print 1;"} {
		b := []byte(src)
		got := parse.Canon(parse.Parse(b), b)
		if !strings.Contains(got, "SKIP:") {
			t.Errorf("canon of %q dropped the label: %q", src, got)
		}
	}
}

// TestContinueBlock: `continue BLOCK` belongs to the loop it follows.
//
// Measured on perl 5.42.0, which is the authority on which forms take one:
//
//	while  ($i<3)  {...} continue {$i++}          ->  0 1 2
//	until  ($i>=3) {...} continue {$i++}          ->  0 1 2
//	foreach my $x (1..3) {...} continue {...}     ->  b1 c1 b2 c2 b3 c3
//	{...} continue {...}                          ->  b c
//	for (my $i=0; $i<3; $i++) {...} continue {..}  ->  SYNTAX ERROR
//
// The C-style head is the one form perl REJECTS. TestContinueRejectedFormsStay
// pins that separately, because a construct perl rejects is one we are right to
// decline.
func TestContinueBlock(t *testing.T) {
	for _, tc := range []struct {
		src  string
		kind parse.Kind
	}{
		{"while ($i < 3) { $x } continue { $i++ }", parse.Loop},
		{"until ($i >= 3) { $x } continue { $i++ }", parse.Loop},
		{"foreach my $x (1..3) { $y } continue { $z }", parse.Loop},
		{"for my $x (1..3) { $y } continue { $z }", parse.Loop},
		// A labelled loop still works. The labelled-block fix landed in
		// internal/lexer/expect.go, so a label before the loop must not
		// change where the continue block attaches.
		{"L: while ($i < 3) { $x } continue { $i++ }", parse.Loop},
		// A BARE block takes one too -- perl runs the block once, then the
		// continue: `{push @s,"b"} continue {push @s,"c"}` prints `b c`.
		{"{ $x } continue { $y }", parse.Block},
	} {
		root := parse.Parse([]byte(tc.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", tc.src, kinds(root))
			continue
		}
		if firstOfKind(root, tc.kind) == nil {
			t.Errorf("%q must produce a %v: %v", tc.src, tc.kind, kinds(root))
		}
		// Canon must place the keyword back. Without it the continue block
		// emits as a second bare Block and the word is gone from the source
		// text -- zero Unknowns and a tree that is not a parse of its bytes.
		if got := parse.Canon(root, []byte(tc.src)); !strings.Contains(got, "continue") {
			t.Errorf("canon of %q dropped `continue`: %q", tc.src, got)
		}
	}
}

// TestContinueRejectedFormsStay: a construct perl rejects is one to decline.
//
// The C-style `for` head is the only loop head that does NOT take a continue
// block. Measured on perl 5.42.0:
//
//	$ perl -e 'for (my $i=0; $i<3; $i++) { } continue { }'
//	syntax error at -e line 1, near "} continue "
//
// So `continue` must NOT be folded into that loop. And BARE `continue;` is a
// different statement form -- the jump out of a `when` block -- which is not
// implemented and must not be read as an empty loop clause.
func TestContinueRejectedFormsStay(t *testing.T) {
	// The C-style head. The continue must be a SEPARATE statement from the
	// loop, which is what refusing to fold it looks like in the tree.
	src := "for (my $i = 0; $i < 3; $i++) { $x } continue { $y }"
	root := parse.Parse([]byte(src))
	loop := firstOfKind(root, parse.Loop)
	if loop == nil {
		t.Fatalf("%q lost its Loop: %v", src, kinds(root))
	}
	for _, c := range loop.Children {
		if c.Kind == parse.Loop && c.Text == "continue" {
			t.Errorf("%q folded a continue block perl rejects into the loop", src)
		}
	}

	// Bare `continue;`. perl accepts it inside a `when` block and this parser
	// does not implement that jump, so it stays declined rather than becoming
	// a continue clause with no block.
	root = parse.Parse([]byte("continue;"))
	if n := firstOfKind(root, parse.Loop); n != nil && n.Text == "continue" {
		t.Errorf("bare `continue;` was read as a loop clause: %v", kinds(root))
	}

	// What the rejected C-style spelling DOES become, recorded rather than
	// asserted as desirable: `continue` is no longer a declined statement
	// keyword, so its block reaches the general `WORD BLOCK` reading -- the
	// same one `defer { ... }` gets. That is a statement whose shape this
	// parser can read; what perl rejects is the PAIRING with the loop, and
	// the loop no longer claims it. Printed so a later change to that reading
	// is visible here.
	t.Logf("C-style continue reads as: %v", kinds(parse.Parse([]byte(src))))
	t.Logf("bare `continue;` reads as: %v", kinds(parse.Parse([]byte("continue;"))))
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
