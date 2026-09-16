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

	// Statement position: a block, even with a bareword key inside, because
	// the lexer read the brace in XState.
	for _, src := range []string{
		"{ a => 1 }",
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

// TestBlockVersusHashrefAfterListOp is the `map { ... }` half, and it is NOT
// covered here.
//
// The measurements are recorded above: a leading `;` turns two elements per
// input into four, and without one `map { a => 1 } (1,2)` is a syntax error
// because perl guessed hashref. But `map` is a list operator, and list
// operators are their own issue -- `map { ; a => 1 } (1,2)` is Unknown today
// for that reason, not because the brace decision is wrong.
//
// The skip came off when list operators landed, and the test STILL fails --
// for a third reason, now isolated. `map { ; a => 1 }` parses as
// `call(anon_hash)`: a list operator leaves a TERM expected, so the lexer's
// brace stack calls the `{` a hash constructor before the `;` inside it is
// ever seen.
//
// perl settles this with intuit_curly, which peeks PAST the brace at the
// first token -- a `;` forces a block, a bareword-then-`=>` forces a hashref.
// Neither the lexer nor the parser does that lookahead, and it is the one
// place in the grammar where a brace's meaning depends on what is INSIDE it
// rather than on what precedes it. That is its own piece of work.
//
// Kept skipped so the gap keeps its name and its measurements.
func TestBlockVersusHashrefAfterListOp(t *testing.T) {
	t.Skip("needs intuit_curly lookahead: a list operator leaves XTerm, so `map {` lexes as a hashref")

	root := parse.Parse([]byte("my @r = map { ; a => 1 } (1,2);"))
	if firstOfKind(root, parse.Block) == nil {
		t.Errorf("a leading `;` forces a block: %v", kinds(root))
	}

	root = parse.Parse([]byte("my @r = map { a => 1 }, (1,2);"))
	if firstOfKind(root, parse.AnonHash) == nil {
		t.Errorf("a bareword before `=>` makes it a hashref: %v", kinds(root))
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
