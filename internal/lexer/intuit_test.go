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

// TestRightParenRegressionIsNamed pins WHY the closer-then-brace rule is
// still broader than perl's, and what it costs.
//
// perl narrows it to `)`. Ours fires on any closer. Narrowing it to match
// used to regress 16 T1 files; the phaser table took that to ONE, measured
// after the fact rather than assumed:
//
//	my @s = $r->@[ 2, 1 ];      postderef.t:12
//
// `@[` lexes as a Variable, so no `[` ever reaches the bracket stack, and
// its `]` pops whatever was underneath. The stack is corrupted from there,
// and the broad rule papers over the next brace. That is a postfix-slice
// sigil defect, not a brace one -- 01a0ad52 owns it, and narrowing this rule
// is 01a0b6a2's last step.
//
// What the 16 WERE is the other half of the record, and it is the case this
// test holds: without a phaser table, `BEGIN` is an ordinary bareword
// leaving XTerm, its `{` is an anonymous hash, its `}` leaves XOperator
// instead of XState, and a bare block after it is a hashref too.
//
//	BEGIN { my $x = 1; }
//	{ use Foo; }
//
// These now pass on their own merits rather than by compensation, which is
// what made narrowing cheap. The test stays because it is the guard: revert
// the phaser table and these fail, whatever the closer rule does.
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

	// The forms that do NOT need it, which is what makes the phaser case a
	// defect rather than the way Perl works. Each ends in the same bare
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
