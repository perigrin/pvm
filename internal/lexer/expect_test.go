// ABOUTME: The term-vs-operator state: what makes / division or a pattern, and x an operator or a name.
// ABOUTME: Every expectation was run against perl 5.42 with B::Deparse or B::Concise before being written.

package lexer

import "testing"

// kindsOf returns the non-whitespace token kinds of src, for assertions about
// how a construct was DIVIDED rather than what it contains.
func kindsOf(src string) []Kind {
	var out []Kind
	for _, tok := range Tokenize([]byte(src)) {
		if tok.Kind != Whitespace {
			out = append(out, tok.Kind)
		}
	}
	return out
}

// spanOfKind returns the span of the first token of kind k, or (-1,-1).
func spanOfKind(src string, k Kind) (int, int) {
	for _, tok := range Tokenize([]byte(src)) {
		if tok.Kind == k {
			return tok.Start, tok.End
		}
	}
	return -1, -1
}

// TestExpectStatesComplete pins that all eleven states exist.
//
// perl.h:5972-5985 defines them and the spec is emphatic that they all be
// modelled now: §2.15 item 3 "All eleven states ... Everything else depends
// on it", §3.9.4 "Do not economise here", §3.12 "Expect enum with all 11
// states".
//
// The reason is the same one that makes trivia tokens the right call:
// retrofitting the missing states later is a rewrite, not an addition. POD
// recognition needs XSTATE; a `{` after a bare sigil needs the block-vs-hash
// distinction; attributes need XATTRBLOCK and XATTRTERM. Exercising a subset
// is fine. Defining a subset is not.
func TestExpectStatesComplete(t *testing.T) {
	// Named exactly as perl.h spells them, so a reader can diff the two
	// lists without translating.
	want := []struct {
		state Expect
		name  string
	}{
		{XOperator, "XOPERATOR"},
		{XTerm, "XTERM"},
		{XRef, "XREF"},
		{XState, "XSTATE"},
		{XBlock, "XBLOCK"},
		{XAttrBlock, "XATTRBLOCK"},
		{XAttrTerm, "XATTRTERM"},
		{XTermBlock, "XTERMBLOCK"},
		{XBlockTerm, "XBLOCKTERM"},
		{XPostDeref, "XPOSTDEREF"},
		{XTermOrDorDor, "XTERMORDORDOR"},
	}
	if len(want) != 11 {
		t.Fatalf("the table itself lists %d states, want 11", len(want))
	}
	seen := map[Expect]bool{}
	for _, w := range want {
		if seen[w.state] {
			t.Errorf("%s duplicates another state's value: the enum is not distinct", w.name)
		}
		seen[w.state] = true
		if got := w.state.String(); got != w.name {
			t.Errorf("state %d prints as %q, want %q — the name must match perl.h "+
				"so the two can be compared without translation", int(w.state), got, w.name)
		}
	}
}

// TestExpectSlash: a slash after a TERM is division; after an OPERATOR it
// opens a pattern. Measured:
//
//	$ perl -MO=Deparse -e 'my @a=(1); my $n = @a / 2;'
//	my $n = @a / 2;
//
// t/base/lex.t:20 has the harder version, `eval '$foo{1} / 1;'`: division,
// because a hash subscript closes a term.
func TestExpectSlash(t *testing.T) {
	// After a term: division. No Quote token may appear.
	for _, src := range []string{
		`$x / 2`,
		`@a / 2`,
		`$foo{1} / 1`,
		`$a[0] / 1`,
		`f() / 2`,
	} {
		for _, k := range kindsOf(src) {
			if k == Quote {
				t.Errorf("%q: a slash after a term opened a pattern; it is division", src)
				break
			}
		}
	}

	// In term position: a pattern. Exactly one Quote token.
	for _, tc := range []struct {
		src     string
		wantEnd int
	}{
		{`/abc/`, 5},
		{`$x =~ /abc/`, 11},
		{`split /,/, $s`, 9},
	} {
		start, end := spanOfKind(tc.src, Quote)
		if start < 0 {
			t.Errorf("%q: no pattern found; a slash in term position opens one", tc.src)
			continue
		}
		if end != tc.wantEnd {
			t.Errorf("%q: pattern ends at %d, want %d", tc.src, end, tc.wantEnd)
		}
	}
}

// TestExpectRepetition is §0.13 rank 2, 13 corpus files.
//
//	$ perl -MO=Deparse -e 'print((1)x3)'
//	print((1) x 3);
//
// Glued or spaced, `x` in operator position is the repetition operator. The
// §0.13 measurement that makes this worth its own test: `(1) x 3` parses
// today, `((1)x3, 2)` fails, and `(1)x3` alone is SILENTLY degenerate — a
// clean tree that is not a parse of its source.
func TestExpectRepetition(t *testing.T) {
	for _, src := range []string{
		`(1)x3`,
		`(1) x 3`,
		`((1)x3, 2)`,
		`$s x 4`,
		`$s x4`,
	} {
		// The repetition operator must not be swallowed into an identifier,
		// and must not open a quote-like operator: `x` is not `q`.
		for _, k := range kindsOf(src) {
			if k == Quote {
				t.Errorf("%q: `x` in operator position opened a quote-like operator; "+
					"it is repetition", src)
				break
			}
		}
	}

	// In TERM position the same letter starts an identifier, and the test
	// would be vacuous without this half: a lexer that never treats x as an
	// operator passes the loop above.
	start, end := spanOfKind(`x3 => 1`, Word)
	if start != 0 || end != 2 {
		t.Errorf("`x3 => 1`: identifier span [%d,%d), want [0,2) — in term "+
			"position x starts a name", start, end)
	}
}

// TestExpectRegexModifiers is §0.13 rank 5, 6 corpus files, and it is rank 2
// wearing a different hat: after a quote-like operator's final delimiter the
// lexer is in OPERATOR position, so an `x` there is repetition rather than a
// modifier — except that the modifier run is consumed by the quote scanner
// first.
//
// The measured corpus failure: `s!a!b!x` is degenerate while `s!a!b!g`
// parses.
func TestExpectRegexModifiers(t *testing.T) {
	for _, tc := range []struct {
		src     string
		wantEnd int
	}{
		{`s!a!b!x`, 7},
		{`s!a!b!g`, 7},
		{`s!a!b!gi`, 8},
		{`s!a!b!xe`, 8},
		{`m!a!x`, 5},
		{`qr!a!x`, 6},
	} {
		start, end := spanOfKind(tc.src, Quote)
		if start != 0 || end != tc.wantEnd {
			t.Errorf("%q: span [%d,%d), want [0,%d) — trailing modifiers belong "+
				"to the operator", tc.src, start, end, tc.wantEnd)
		}
	}
}

// TestExpectArrayLastIndex is t/base/lex.t line 10, and the gloss in the spec
// and in this issue's own text was WRONG. Measured:
//
//	$ perl -MO=Concise,-exec -e '$x = $#[0];'
//	3  <#> aelemfast[*#] s
//
//	$ perl -e '@# = (42); print $#[0]'
//	42
//
// `$#[0]` is element 0 of the array `@#`. It is NOT "$# followed by a
// subscript" — there is no `$#` term here at all. The sigil is `$`, the
// variable name is `#`, and `[0]` subscripts it.
//
// The contrast that makes it a lexer trap: `$#x` IS last-index-of-@x. Same
// two opening bytes, different reading, decided by what follows.
func TestExpectArrayLastIndex(t *testing.T) {
	// $#[0]: the name is `#`, so the variable token covers `$#` and the
	// subscript is separate.
	start, end := spanOfKind(`$#[0]`, Variable)
	if start != 0 || end != 2 {
		t.Errorf("`$#[0]`: variable span [%d,%d), want [0,2) — `$#` names the "+
			"array @#, and `[0]` is a subscript", start, end)
	}

	// $#x: last index of @x. The whole thing is one variable token.
	start, end = spanOfKind(`$#x`, Variable)
	if start != 0 || end != 3 {
		t.Errorf("`$#x`: variable span [%d,%d), want [0,3) — this is "+
			"last-index-of-@x, one lexeme", start, end)
	}
}

// TestExpectAngleAndAmp covers the two remaining position-dependent
// dispatches this milestone needs.
//
// Angle brackets: after a term, comparison; in term position, a readline.
// Ampersand: §0.13 rank 7, 3 corpus files. Spec §3.2's `&` row — operator
// position yields `&&`, term position yields `&name`.
func TestExpectAngleAndAmp(t *testing.T) {
	// Readline in term position is one token.
	start, end := spanOfKind(`my $l = <STDIN>;`, Readline)
	if start != 8 || end != 15 {
		t.Errorf("`<STDIN>` span [%d,%d), want [8,15) — a readline in term position",
			start, end)
	}

	// After a term, `<` is comparison -- and the shape must be one that WOULD
	// lex as a readline in term position, or the test proves nothing about
	// the state machine.
	//
	// `$a < $b` is not such a case: `< $b` has no closing `>`, so scanAngle
	// declines on shape alone and the test passes even with the state machine
	// disabled. `$a <STDIN>` is the discriminating input: identical bytes to
	// a real readline, and only the expect state says it is not one.
	//
	// Found by mutation -- making wantsTerm always return true left every
	// test in this file green, which is the "guard that cannot fail" shape.
	for _, src := range []string{
		`$a <STDIN>`,
		`$x[0] <FH>`,
		`1 <FOO>`,
	} {
		if s, _ := spanOfKind(src, Readline); s >= 0 {
			t.Errorf("%q: `<` after a term produced a readline; it is comparison, "+
				"and only the expect state distinguishes them", src)
		}
	}

	// The same discrimination for `&`: `$a & $b` is bitwise-and, and `&foo`
	// in the same position is still not a function sigil.
	for _, src := range []string{
		`$a &foo`,
		`1 &bar`,
	} {
		if s, _ := spanOfKind(src, FuncSigil); s >= 0 {
			t.Errorf("%q: `&` after a term produced a function sigil; it is "+
				"bitwise-and", src)
		}
	}

	// And `%` after a term is modulus, not a hash sigil.
	for _, src := range []string{
		`$a %foo`,
		`7 %bar`,
	} {
		for _, tok := range Tokenize([]byte(src)) {
			if tok.Kind == Variable && src[tok.Start] == '%' {
				t.Errorf("%q: `%%` after a term produced a hash variable; it is modulus", src)
			}
		}
	}

	// `&&` in operator position is one operator token, not two `&name` sigils.
	if s, e := spanOfKind(`$a && $b`, Operator); s != 3 || e != 5 {
		t.Errorf("`$a && $b`: operator span [%d,%d), want [3,5)", s, e)
	}

	// `&foo` in term position is a function sigil plus a name.
	if s, e := spanOfKind(`&foo()`, FuncSigil); s != 0 || e != 1 {
		t.Errorf("`&foo()`: function sigil span [%d,%d), want [0,1)", s, e)
	}
}

// TestMethodNameAfterArrow: `$o->s` is a method call, and the method's name
// may be spelled the same as a quote-like operator.
//
// `s`, `y`, `q`, `m` and `tr` are all legal method names, and perl reads
// every one of them as a name after `->` -- measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'my $o; my @x = ($o->s, $o->y, $o->q, $o->tr, $o->m);'
//	my(@x) = ($o->s, $o->y, $o->q, $o->tr, $o->m);
//
// Before XPostDeref was assigned, `->` left XTerm and scanQuoteLike claimed
// the name: `is($o->s, "x", 'a')` lexed as Variable `$o`, Operator `->`,
// UnknownRest `s, "x", 'a');` -- the comma delimited a substitution that ran
// to end of input, and every statement after it in the file was lost.
//
// That is what made class/accessor.t, class/field.t and comp/parser_run.t
// score WRONG through the subject contract: the anonymous hashes perl built
// on those lines were inside bytes this lexer had swallowed.
func TestMethodNameAfterArrow(t *testing.T) {
	for _, src := range []string{
		`$o->s, 1`,
		`$o->y, 1`,
		`$o->q, 1`,
		`$o->m, 1`,
		`$o->tr, 1`,
		// Whitespace and a newline between the arrow and the name do not
		// change the answer; perl deparses all three to `$o->s`.
		`$o -> s, 1`,
		"$o->\ns, 1",
	} {
		for _, tok := range Tokenize([]byte(src)) {
			switch tok.Kind {
			case Quote, UnknownRest, Error:
				t.Errorf("%q: lexed %v %q; after `->` a name is a method",
					src, tok.Kind, src[tok.Start:tok.End])
			}
		}
	}
}

// TestPostfixDerefStillLexes is the other half of XPostDeref, and it is why
// `->` cannot simply leave XOperator.
//
// A sigil after `->` is a postfix dereference and needs a TERM expected, or
// `%` reads as modulus and `@` as an error. perl accepts all of these:
//
//	$ perl -MO=Deparse -e 'my $r; my @a = $r->@*; my %h = $r->%*; my $s = $r->$*;'
//	my(@a) = @$r;
//	my(%h) = %$r;
//	my $s = $$r;
//
// So XPostDeref is term-ish for sigils and operator-ish for words, which is
// exactly the distinction the state was declared for.
func TestPostfixDerefStillLexes(t *testing.T) {
	for _, tc := range []struct {
		src  string
		kind Kind
	}{
		{`$r->@*`, Variable},
		{`$r->%*`, Variable},
		{`$r->$*`, Variable},
		{`$r->$m`, Variable},
		{`$r->&*`, FuncSigil},
	} {
		// The token AT the sigil, not the first of its kind: `$r` is a
		// Variable too, and spanOfKind would return that one.
		var got Kind = -1
		for _, tok := range Tokenize([]byte(tc.src)) {
			if tok.Start == 4 {
				got = tok.Kind
				break
			}
		}
		if got != tc.kind {
			t.Errorf("%q: want a %v at offset 4, got %v (tokens %v)",
				tc.src, tc.kind, got, Tokenize([]byte(tc.src)))
		}
	}
}

// TestPostfixSliceSigilStopsAtTheBracket: `$r->@[0,1]` is a sigil and a
// SUBSCRIPT, so the bracket must reach the bracket stack.
//
// The sigil scanner took `@[` as a punctuation variable named `[` -- a
// plausible-looking token, and nothing checked it. No opener was pushed, so
// the matching `]` popped whatever was underneath and the stack stayed one
// entry short for the REST OF THE FILE. The visible symptom was a brace
// misread 40 lines later in postderef.t, which is why this went unfound
// through three separate passes over brace classification.
//
// The contrast is what makes it a rule rather than a special case: `@*` and
// `%*` ARE the whole dereference and stay one token. Measured:
//
//	$ perl -MO=Deparse -e 'my $r=[1,2,3]; my @s = $r->@[0,1]; my @t = $r->@*;'
//	my(@s) = @$r[0, 1];
//	my(@t) = @$r;
func TestPostfixSliceSigilStopsAtTheBracket(t *testing.T) {
	for _, tc := range []struct {
		src   string
		sigil string
	}{
		{`my @s = $r->@[ 2, 1 ];`, "@"},
		{`my @s = $r->@{ 'a' };`, "@"},
		{`my @s = $r->%[ 0 ];`, "%"},
	} {
		var got string
		for _, tok := range Tokenize([]byte(tc.src)) {
			if tok.Kind == Variable && tok.Start > 8 {
				got = tc.src[tok.Start:tok.End]
				break
			}
		}
		if got != tc.sigil {
			t.Errorf("%q: the sigil of a postfix slice is %q alone, got %q -- "+
				"swallowing the bracket leaves its closer to pop someone "+
				"else's opener", tc.src, tc.sigil, got)
		}
	}

	// The whole-deref forms keep their star, and this is the guard against
	// "fixing" the above by splitting every sigil after `->`.
	for _, src := range []string{`$r->@*`, `$r->%*`, `$r->$*`} {
		var got string
		for _, tok := range Tokenize([]byte(src)) {
			if tok.Start == 4 {
				got = src[tok.Start:tok.End]
				break
			}
		}
		if len(got) != 2 {
			t.Errorf("%q: `%s` is the whole dereference and stays one token, "+
				"got %q", src, src[4:], got)
		}
	}

	// And the stack stays balanced afterwards, which is the property the
	// swallowed bracket actually broke. A `}` that closes a BLOCK after the
	// slice must still say so.
	src := `my @s = $r->@[0,1]; if (1) { 2 }`
	var sawBlockClose bool
	for _, tok := range Tokenize([]byte(src)) {
		if tok.Kind == CloseBracket && src[tok.Start] == '}' && tok.OpensBlock {
			sawBlockClose = true
		}
	}
	if !sawBlockClose {
		t.Errorf("%q: the `}` after a postfix slice must still close a BLOCK "+
			"-- an unbalanced stack misreads every brace that follows", src)
	}
}

// TestQuoteOpNameAfterSubOrMethod: the name of a `sub` or a `method` may be
// spelled the same as a quote-like operator, and perl reads it as a name.
//
//	$ perl -MO=Deparse -e 'sub y { 1 } sub s { 2 } sub q { 3 }'
//	sub y { 1; }
//	sub s { 2; }
//	sub q { 3; }
//
// `method y { return @y; }` in t/class/field.t:166 is the same shape, and
// scanQuoteLike was reading that `y` with the following `{` as the start of
// a transliteration -- which swallowed the rest of the class body.
//
// The lexer already knows it is at a declaration's name: noteSubName sets
// sawSubWord on `sub` and `method` for the prototype scanner. This asks the
// same fact one token earlier.
func TestQuoteOpNameAfterSubOrMethod(t *testing.T) {
	for _, src := range []string{
		`sub y { 1 }`,
		`sub s { 2 }`,
		`sub q { 3 }`,
		`sub tr { 4 }`,
		`sub m { 5 }`,
		// `method` declares only with class syntax in effect -- the
		// feature, as t/class/field.t has, or a class already declared,
		// which is the spelling here since a quote in the source would trip
		// this test's own check. Without it op/method.t calls a sub named
		// method.
		`class C; method y { return $y; }`,
		`class C; method s { 1 }`,
		// Trivia between the keyword and the name changes nothing.
		`sub  y  { 1 }`,
	} {
		for _, tok := range Tokenize([]byte(src)) {
			switch tok.Kind {
			case Quote, UnknownRest, Error:
				t.Errorf("%q: lexed %v %q; a declaration's name is a name",
					src, tok.Kind, src[tok.Start:tok.End])
			}
		}
	}
}

// TestQuoteOpStillLexesElsewhere is the mutation guard for both declines:
// the same keywords in TERM position are still quote operators, and the
// machine is back to normal after the declaration's name.
func TestQuoteOpStillLexesElsewhere(t *testing.T) {
	for _, src := range []string{
		`$x =~ s/a/b/`,
		`$x =~ y/a/b/`,
		`$x =~ tr/a/b/`,
		`$x =~ m/a/`,
		`my $v = q(a)`,
		`sub f { $x =~ s/a/b/ }`,
		`$o->f; $x =~ s/a/b/;`,
	} {
		if s, _ := spanOfKind(src, Quote); s < 0 {
			t.Errorf("%q: no Quote token; the keyword is still a quote "+
				"operator here (tokens %v)", src, Tokenize([]byte(src)))
		}
	}
}

// TestMethodNeedsClassSyntax: without class syntax `method` is a name, so
// op/method.t's `method Pack ("a")` keeps its parens as arguments rather
// than a prototype; `use feature 'class'` turns the keyword on. Measured on
// 5.42.0, `sub Pack::method { "m" } my $x = method Pack ("a");` deparses as
// `my $x = 'Pack'->method('a');`.
func TestMethodNeedsClassSyntax(t *testing.T) {
	for _, tok := range Tokenize([]byte(`my $x = method Pack ("a");`)) {
		if tok.Kind == Prototype {
			t.Errorf("without class syntax `method` does not declare; got a Prototype")
		}
	}
	found := false
	for _, tok := range Tokenize([]byte(`use feature 'class'; method m ($$) { 1 }`)) {
		if tok.Kind == Prototype {
			found = true
		}
	}
	if !found {
		t.Errorf("under the class feature `method m ($$)` declares with a prototype")
	}
}
