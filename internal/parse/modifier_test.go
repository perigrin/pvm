// ABOUTME: Statement modifiers, and the `{` that is a block or a hash constructor.
// ABOUTME: The modifier binds looser than every expression operator, `or` included.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestStatementModifierForms: all six, each inverting the tree.
//
// `$y = 1 if $x` is a conditional whose body is the assignment, not an
// assignment whose value is a conditional.
func TestStatementModifierForms(t *testing.T) {
	for _, tc := range []struct {
		src  string
		kind parse.Kind
	}{
		// Bodies are assignments and increments rather than `print ...`:
		// list operators are their own issue and `print 1;` is still
		// Unknown, which would make these fail for an unrelated reason.
		{"$y = 1 if $x;", parse.Conditional},
		{"$y = 1 unless $x;", parse.Conditional},
		{"$y++ while $x;", parse.Loop},
		{"$y++ until $x;", parse.Loop},
		{"$y++ for @list;", parse.Loop},
		{"$y++ foreach @list;", parse.Loop},
	} {
		root := parse.Parse([]byte(tc.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", tc.src, kinds(root))
			continue
		}
		n := firstOfKind(root, tc.kind)
		if n == nil {
			t.Errorf("%q must produce a %v: %v", tc.src, tc.kind, kinds(root))
			continue
		}
		// The modifier owns the statement: its span covers the whole thing,
		// which is what "inverts the tree" means concretely.
		if n.Start != 0 {
			t.Errorf("%q: the modifier node starts at %d, want 0 -- it must own "+
				"the statement, not sit inside it", tc.src, n.Start)
		}
	}
}

// TestModifierBindsLoosest: looser than every expression operator.
//
// `or` is level 4, the loosest thing in the expression grammar, and the
// modifier is still below it. Measured:
//
//	$ perl -MO=Deparse -e 'print("a") or die("b") if $x;'
//	print 'a' or die 'b' if $x;
//
// So the `if` takes `print ... or die ...` entire as its body.
func TestModifierBindsLoosest(t *testing.T) {
	src := []byte("$y = 1 or $z = 2 if $x;")
	root := parse.Parse(src)
	if containsKind(root, parse.Unknown) {
		t.Fatalf("must parse: %v", kinds(root))
	}

	cond := firstOfKind(root, parse.Conditional)
	if cond == nil {
		t.Fatalf("no Conditional: %v", kinds(root))
	}
	// The `or` must be INSIDE the conditional's body, not outside it.
	if firstOfKind(cond, parse.Binary) == nil {
		t.Errorf("the `or` belongs inside the modifier's body: %v", kinds(cond))
	}
	// And the conditional owns the whole statement.
	if cond.Start != 0 {
		t.Errorf("the modifier starts at %d, want 0", cond.Start)
	}
}

// TestModifierDoesNotCrossATerminator: a finished statement cannot take one.
//
// `my $z;` ends at its semicolon, so the `foreach` on the next line starts a
// NEW statement. It is not a modifier on the declaration.
//
// Every declaration followed by a control-flow statement used to collapse
// into one Loop containing the declaration, with the loop's body falling to
// Unknown. The declaration path is the only one that reaches applyModifier
// after consuming a terminator, which is why nothing else showed it -- and
// why the existing modifier tests, which use un-terminated bodies, all
// passed.
func TestModifierDoesNotCrossATerminator(t *testing.T) {
	for _, src := range []string{
		"my $z;\nforeach my $e (@a) { $s; }\n",
		"my $z = 0;\nforeach my $e (@a) { $s; }\n",
		"my $z = 0;\nif ($c) { $s; }\n",
		"our $z;\nwhile ($c) { $s; }\n",
		"my $z = 0;\nfor my $e (@a) { $s; }\n",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
			continue
		}
		// Two statements, not one: the declaration and the control flow.
		var stmts int
		for _, n := range root.Children {
			if n.Kind == parse.Statement {
				stmts++
			}
		}
		if stmts != 2 {
			t.Errorf("%q: got %d statements, want 2 -- the control flow was "+
				"absorbed as a modifier: %v", src, stmts, kinds(root))
		}
	}

	// And a real modifier, which has no terminator before it, still works.
	root := parse.Parse([]byte("$y = 1 if $c;\n"))
	if firstOfKind(root, parse.Conditional) == nil {
		t.Errorf("a genuine modifier must still attach: %v", kinds(root))
	}
}

// TestBlockVersusHashref is spec §4.9.2's decision.
//
// The issue's original example does not discriminate: `map { $_ => 1 }` and
// `map { ; $_ => 1 }` produce IDENTICAL optrees, both blocks, zero anonhash
// (measured). perl already reads that brace as a block, so the `;` changes
// nothing.
//
// Where it bites is a BAREWORD key, which is what intuit_curly's heuristic
// keys on -- a word or string followed by `,` or `=>` means hashref:
//
//	$ perl -e 'my @r = map { a => 1 }, (1,2); print scalar @r'
//	2                       # hashref: one element per input
//	$ perl -e 'my @r = map { ; a => 1 } (1,2); print scalar @r'
//	4                       # block: two elements per input
//	$ perl -e 'my @r = map { a => 1 } (1,2);'
//	syntax error            # hashref guessed, then no comma follows
//
// Two per input versus four is the whole difference, and it is decided by
// one leading semicolon.
func TestBlockVersusHashref(t *testing.T) {
	// Term position: a hashref.
	for _, src := range []string{
		"my $h = { a => 1 };",
		"my $h = {};",
		"$x = { a => 1 };",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
			continue
		}
		if firstOfKind(root, parse.AnonHash) == nil {
			t.Errorf("%q: a brace in term position is a hashref: %v", src, kinds(root))
		}
	}

	// Where a BLOCK is required, a brace opens one whatever is inside it.
	// perl reads the contents as statements and says so -- measured:
	//
	//	$ perl -MO=Deparse -e 'if ($c) { a => 1 }'
	//	if ($c) {
	//	    '???', '???';
	//	}
	//
	// `{ a => 1 }` at STATEMENT START is not in this list any more: perl
	// deparses that one as `+{'a', 1}`, an anonymous hash, and this test
	// asserted the opposite. Its reason was "because the lexer read the
	// brace in XState" -- a description of the implementation rather than of
	// Perl, which is exactly the kind of test that freezes a defect.
	// TestStatementStartBrace in wrongtree_test.go now owns that case.
	for _, src := range []string{
		"if ($c) { a => 1 }",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
			continue
		}
		if firstOfKind(root, parse.AnonHash) != nil {
			t.Errorf("%q: a brace in statement position is a block: %v", src, kinds(root))
		}
		if firstOfKind(root, parse.Block) == nil {
			t.Errorf("%q must produce a Block: %v", src, kinds(root))
		}
	}
}

// TestBlockVersusHashrefAfterListOp is the `map { ... }` half: the brace
// after a list operator, decided by LOOKAHEAD rather than by position.
//
// Everywhere else a brace is settled by what came before it, which is what
// the lexer's brace stack records. After `map`, `grep` and `sort` both
// readings are grammatical, so perl peeks past the brace at the first thing
// inside -- intuit_curly, toke.c:6698-6842, now `internal/lexer/intuit.go`.
//
// The measurements that pin the rule, on perl 5.42.0:
//
//	$ perl -e 'my @r = map { a => 1 }, (1,2); print scalar @r'
//	2                       # hashref: one element per input
//	$ perl -e 'my @r = map { ; a => 1 } (1,2); print scalar @r'
//	4                       # block: two elements per input
//	$ perl -e 'my @r = map { a => 1 } (1,2);'
//	syntax error            # hashref guessed, then no comma follows
//
// Two per input versus four, decided by one leading semicolon.
func TestBlockVersusHashrefAfterListOp(t *testing.T) {
	root := parse.Parse([]byte("my @r = map { ; a => 1 } (1,2);"))
	if firstOfKind(root, parse.Block) == nil {
		t.Errorf("a leading `;` forces a block: %v", kinds(root))
	}

	root = parse.Parse([]byte("my @r = map { a => 1 }, (1,2);"))
	if firstOfKind(root, parse.AnonHash) == nil {
		t.Errorf("a bareword before `=>` makes it a hashref: %v", kinds(root))
	}
}

// TestMapBlockForms: the block is a SLOT, and the list after it belongs to
// the same call.
//
// The failure this guards against is subtler than a misread brace, and it is
// what the brace being wrong actually cost: with the `{` read as a hashref,
// `map { $_ => 1 } @a` produced a call holding an AnonHash and then ORPHANED
// `@a` into a statement of its own. Every byte was in the tree and it
// round-tripped, so nothing caught it -- the same failure mode as the four
// trees of 5b415101.
//
// perl's shape, and the one asserted here:
//
//	$ perl -MO=Deparse -e 'my %h = map { $_ => 1 } @a;'
//	my(%h) = map({$_, 1;} @a);
func TestMapBlockForms(t *testing.T) {
	for _, tc := range []struct {
		src  string
		call string
	}{
		{"my %h = map { $_ => 1 } @a;", "map"},
		{"my @b = grep { $_ } @a;", "grep"},
		{"my @b = sort { $a <=> $b } @a;", "sort"},
		{"my @b = map { ; $_ } @a;", "map"},
	} {
		root := parse.Parse([]byte(tc.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", tc.src, kinds(root))
			continue
		}
		call := firstOfKind(root, parse.Call)
		if call == nil || call.Text != tc.call {
			t.Errorf("%q: want a %s call: %v", tc.src, tc.call, kinds(root))
			continue
		}
		// The block, then the list, both children of the call. Two children
		// is the whole assertion: one would mean the list was orphaned.
		if len(call.Children) != 2 {
			t.Errorf("%q: %s takes a block AND its list, got %d child(ren): %v",
				tc.src, tc.call, len(call.Children), kinds(root))
			continue
		}
		if call.Children[0].Kind != parse.Block {
			t.Errorf("%q: the first child is the block, got %v",
				tc.src, call.Children[0].Kind)
		}
	}

	// `map +{ ... }, @a` is the OTHER reading and must stay one: a hashref,
	// with a comma after it, so the block slot takes nothing.
	//
	//	$ perl -MO=Deparse -e 'my @b = map +{ x => $_ }, @a;'
	//	my(@b) = map({'x', $_}, @a);
	root := parse.Parse([]byte("my @b = map +{ x => $_ }, @a;"))
	if firstOfKind(root, parse.AnonHash) == nil {
		t.Errorf("`map +{...}` is a hashref, not a block: %v", kinds(root))
	}
}

// TestEmptyStatement: a lone `;` is a statement, not a parse failure.
//
// Valid Perl, and it is what makes `map { ; $_ }` a block rather than a
// hashref -- so this milestone reaches it through the list-operator brace
// even where nobody writes a bare `;` on purpose.
//
// Before this, the expression parser was handed the `;`, returned nil, and
// skipToStatementEnd ran PAST the enclosing `}` hunting a terminator it had
// already gone by. The block lost its closing brace into an Unknown:
//
//	map { ; $_ } @a    ->  block "{ ; $_ } @a;" holding unknown "; $_ }"
func TestEmptyStatement(t *testing.T) {
	for _, src := range []string{
		";",
		"{};",
		"{ ; }",
		"my $x = 1;;",
	} {
		root := parse.Parse([]byte(src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q must parse: %v", src, kinds(root))
		}
	}
}

// TestBraceDecisionUsesExpectState: the decision is READ from the lexer, not
// made twice.
//
// The lexer's brace stack already classified every `{`, because a `}` leaves
// a different expect state depending on what it closed. A parser that
// re-derived it could disagree with the token stream that produced it, and
// then the two halves of the same program would be parsed under different
// assumptions.
func TestBraceDecisionUsesExpectState(t *testing.T) {
	// A block after a conditional's `)`, which the lexer marks via the
	// `)`-then-`{` rule rather than by anything the parser can see locally.
	root := parse.Parse([]byte("if ($c) { a => 1 }"))
	if firstOfKind(root, parse.AnonHash) != nil {
		t.Errorf("a conditional's body is a block even with a bareword key: %v",
			kinds(root))
	}
	if firstOfKind(root, parse.Block) == nil {
		t.Errorf("the conditional's body must be a Block: %v", kinds(root))
	}
}
