// ABOUTME: intuit_curly: the brace after a list operator, and the closer-then-brace rule beside it.
// ABOUTME: Every expectation here was measured on perl 5.42.0 before it was written.

package lexer

import "testing"

// opensBlockAt reports whether the nth `{` in src was classified as opening a
// block. Counting braces rather than tokens because that is how the cases
// below read: "the second brace of `$a[0]{k}`".
func opensBlockAt(src string, n int) (opens, found bool) {
	seen := 0
	for _, tok := range Tokenize([]byte(src)) {
		if tok.End-tok.Start == 1 && src[tok.Start] == '{' {
			if seen == n {
				return tok.OpensBlock, true
			}
			seen++
		}
	}
	return false, false
}

// lastBraceOpensBlock is opensBlockAt for the final `{`, where the cases
// being compared have different brace counts before it.
func lastBraceOpensBlock(src string) (opens, found bool) {
	for _, tok := range Tokenize([]byte(src)) {
		if tok.End-tok.Start == 1 && src[tok.Start] == '{' {
			opens, found = tok.OpensBlock, true
		}
	}
	return opens, found
}

// TestIntuitCurlyLookahead: a brace after `map`, `grep` or `sort` is decided
// by what is INSIDE it, not by the token before it.
//
// Everywhere else the preceding token settles a brace, which is what the
// brace stack records. After these three both readings are grammatical, so
// perl peeks past the brace -- intuit_curly, toke.c:6698-6842. Measured:
//
//	$ perl -MO=Deparse -e 'my @b = map { $_ => 1 } @a;'
//	my(@b) = map({$_, 1;} @a);              a BLOCK
//	$ perl -MO=Deparse -e 'my @b = map { a => 1 }, @a;'
//	my(@b) = map({'a', 1}, @a);             a HASHREF
//
// A word or string then `,` or `=>` is a hash. Everything else is a block,
// including a leading `;` and including a VARIABLE before a fat comma --
// which is why the first line is a block even though it reads like a pair.
func TestIntuitCurlyLookahead(t *testing.T) {
	for _, tc := range []struct {
		src   string
		block bool
		why   string
	}{
		{"my @b = map { $_ => 1 } @a;", true, "a variable is not a key"},
		{"my @b = map { ; a => 1 } @a;", true, "a leading `;` forces a block"},
		{"my @b = map { a => 1 }, @a;", false, "bareword then `=>` is a hash"},
		{`my @b = map { "a", 1 }, @a;`, false, "string then `,` is a hash"},
		{"my @b = grep { $_ } @a;", true, "grep takes a block too"},
		{"my @b = sort { $a <=> $b } @a;", true, "so does sort"},
		{"my @b = map { } @a;", true, "empty is a block, as at statement start"},
	} {
		opens, found := opensBlockAt(tc.src, 0)
		if !found {
			t.Errorf("%q: no `{` token", tc.src)
			continue
		}
		if opens != tc.block {
			t.Errorf("%q: OpensBlock is %v, want %v -- %s",
				tc.src, opens, tc.block, tc.why)
		}
	}
}

// TestRightParenLookaheadIsParenOnly: the `{` of a subscript chain is a term
// on both halves.
//
// `$a[0]{k}` is a chain of subscripts, and the second brace must not become a
// block because a `]` preceded it. This is NOT intuit_curly's doing: a closer
// that closed a subscript leaves an operator expected, so the brace stack
// already reads the next `{` as another subscript. The lookahead is needed
// only where both readings are grammatical, and after a subscript they are
// not.
//
// Named for the rule it guards rather than for the mechanism, so a change to
// either one is caught here.
func TestRightParenLookaheadIsParenOnly(t *testing.T) {
	for _, src := range []string{
		"$a[0]{k};",
		"$h{a}{b};",
		"$x->[0]{k};",
	} {
		opens, found := opensBlockAt(src, 0)
		if !found {
			t.Errorf("%q: no `{` token", src)
			continue
		}
		if opens {
			t.Errorf("%q: a subscript brace is a term, not a block", src)
		}
	}
}

// TestRightParenStillOpensABlock is the other side: the closer-then-brace
// rule is what assigns XBlock at all, and every conditional and loop in the
// corpus depends on it.
//
// perl's own version is toke.c's yyl_rightparen (7156):
//
//	s = skipspace(s);
//	if (*s == '{')
//	    PREBLOCK(PERLY_PAREN_CLOSE);
func TestRightParenStillOpensABlock(t *testing.T) {
	for _, src := range []string{
		"if ($c) { 1 }",
		"while ($c) { 1 }",
		"for (;;) { 1 }",
		"foreach my $x (@a) { 1 }",
		"unless ($c) { 1 }",
	} {
		opens, found := opensBlockAt(src, 0)
		if !found {
			t.Errorf("%q: no `{` token", src)
			continue
		}
		if !opens {
			t.Errorf("%q: the brace after `)` opens a block", src)
		}
	}
}

// TestRightParenRegressionIsNamed holds the cases that made the
// closer-then-brace rule too broad to narrow, and it is now the guard on
// them rather than the record of a debt.
//
// perl's rule is `)` alone (toke.c yyl_rightparen, 7156). Ours fired on
// every closer, and narrowing it cost 16 T1 files. Those 16 were not one
// bug: they were two, each hiding the other, and the broad rule was papering
// over both.
//
// FIRST, the phasers. Without a phaser table `BEGIN` is an ordinary bareword
// leaving XTerm, so its `{` is an anonymous hash, its `}` leaves XOperator
// instead of XState, and a bare block after it is read as a hashref too:
//
//	BEGIN { my $x = 1; }
//	{ use Foo; }
//
// That took the cost from 16 files to one.
//
// SECOND, the postfix slice, which was the remaining one:
//
//	my @s = $r->@[ 2, 1 ];      postderef.t:12
//
// `@[` lexed as a Variable, so no `[` reached the bracket stack and its `]`
// popped someone else's opener -- leaving the stack short for the rest of
// the file. TestPostfixSliceSigilStopsAtTheBracket owns that one.
//
// With both fixed the rule is narrowed and costs nothing; postderef.t even
// improved, 3 -> 2. This test keeps the phaser cases because they fail the
// moment that table is reverted, whatever the closer rule does.
func TestRightParenRegressionIsNamed(t *testing.T) {
	// The five phasers. The bare block is the LAST brace: the phaser's own
	// body comes first.
	for _, w := range []string{"BEGIN", "END", "CHECK", "INIT", "UNITCHECK"} {
		src := w + " { my $x = 1; }\n{ use Foo; }\n"
		opens, found := lastBraceOpensBlock(src)
		if !found {
			t.Errorf("%s: no `{` token", w)
			continue
		}
		if !opens {
			t.Errorf("%s: the bare block after a phaser must open a block. "+
				"A phaser's own brace is a block, which is what leaves the "+
				"machine in XState for this one -- see isPhaser.", w)
		}
	}

	// The forms that ALWAYS worked, which is what made the phaser case a
	// defect rather than the way Perl works: a sub body, a conditional and a
	// paren list each leave the machine in a state that reads the following
	// brace correctly, and a phaser did not. Each ends in the same bare
	// block, so the assertion is about the LAST brace -- the preceding forms
	// have different brace counts of their own.
	for _, src := range []string{
		"sub f { 1 }\n{ use Foo; }\n",
		"if (1) { 2 }\n{ use Foo; }\n",
		"my %h = (a => 1);\n{ use Foo; }\n",
	} {
		opens, found := lastBraceOpensBlock(src)
		if !found {
			t.Errorf("%q: no `{` token", src)
			continue
		}
		if !opens {
			t.Errorf("%q: a bare block here needs no compensation", src)
		}
	}
}

// TestLabelColonOpensBlock: a label's `:` leaves a statement boundary, so the
// brace after it is classified as a block -- but intuit_curly still decides.
//
// perl's yyl_colon reaches PREBLOCK for a label, which is XSTATE. Without
// that the `:` left XTerm, `SKIP: { print 1; }` lexed its brace as an
// anonymous hash, and the `}` then reported a closed SUBSCRIPT -- so every
// brace after it in the file was misclassified too. Measured over perl.git t/
// as 284 of the 1,118 bare-brace Unknowns and 9 whole files.
//
// A label does NOT force a block. Measured on perl 5.42.0:
//
//	$ perl -MO=Deparse -e 'L: {a=>1};'
//	L: +{'a', 1};
//
// A labelled anonymous HASH, so the same two-way branch the WORD case makes
// applies here, and the cases below assert both halves.
//
// The discrimination is POSITIONAL, not punctuational. Four shapes put a
// bareword next to a colon and only one is a label; the others are held here
// so a fix that looked for a colon rather than for a statement boundary fails.
func TestLabelColonOpensBlock(t *testing.T) {
	for _, tc := range []struct {
		src   string
		block bool
		why   string
	}{
		{"SKIP: { print 1; }", true, "a labelled bare block"},
		{"SKIP:\n{ print 1; }", true, "the label and its brace on separate lines"},
		{"SKIP : { print 1; }", true, "space before the colon"},
		{"skip: { print 1; }", true, "perl accepts a lowercase label"},
		{"print 1; SKIP: { print 2; }", true, "a label mid-file"},
		{"if (1) { SKIP: { print 2; } }", true, "a label first inside a block"},
		// A labelled hash is NOT distinguished here, and that is this layer's
		// documented boundary rather than a gap: trackBrackets is
		// "intuit_curly minus its lookahead", and the lookahead is the
		// parser's decision (spec §4.9.2, and see trackBrackets' own
		// comment). A statement-start `{ a => 1 };` gets the same `true`, so
		// the labelled form matching the unlabelled one is exactly right --
		// the parser's braceOpensAnonHash refines both, and
		// TestLabelledBareBlock holds that end.
		{"{ a => 1 };", true, "the unlabelled control: this layer says block too"},
		{"L: { a => 1 };", true, "so the labelled form must agree with it"},
	} {
		opens, found := lastBraceOpensBlock(tc.src)
		if !found {
			t.Errorf("%q: no `{` token", tc.src)
			continue
		}
		if opens != tc.block {
			t.Errorf("%q: OpensBlock is %v, want %v -- %s",
				tc.src, opens, tc.block, tc.why)
		}
	}

	// A label STACKS, and the second word must qualify as one too. `A: B: {`
	// works because the first `:` returns XState, which is where the second
	// word then stands -- no second rule.
	if opens, found := lastBraceOpensBlock("A: B: { print 1; }"); !found || !opens {
		t.Errorf("`A: B: { ... }` must open a block: opens=%v found=%v", opens, found)
	}

	// The shapes that are NOT labels, and the ternary is the one that breaks
	// loudly: a `:` misread as a label there turns the hashref after it into a
	// block.
	if opens, found := lastBraceOpensBlock("my $x = $c ? 1 : { a => 1 };"); !found || opens {
		t.Errorf("a ternary's `:` is not a label -- `{ a => 1 }` after it is a "+
			"hashref: opens=%v found=%v", opens, found)
	}
	// An attribute's brace IS a block, for its own reason (afterDeclName), so
	// this stays true whichever rule reaches it -- what must not happen is the
	// `:lvalue` colon being taken for a label, and `f` stands in term position
	// rather than at a statement boundary, which is what denies it.
	if opens, found := lastBraceOpensBlock("sub f :lvalue { 1 }"); !found || !opens {
		t.Errorf("a sub body after an attribute is still a block: opens=%v found=%v",
			opens, found)
	}
}
