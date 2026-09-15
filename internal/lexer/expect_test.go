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
