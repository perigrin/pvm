// ABOUTME: The identifier character class: XID under use utf8, ASCII without it, and both package separators.
// ABOUTME: §0.13 rank 1 is 34 corpus files; the rule came from perldata.pod, not from guesswork about Unicode.

package lexer

import "testing"

// lastToken asserts that the LAST non-trivia token of src has kind k and runs
// to the end of the input.
//
// The last token rather than the only one, because most cases here prefix
// `use utf8;` and the pragma is three tokens of its own. What is under test is
// where the identifier's boundary falls, which the final span answers.
func lastToken(t *testing.T, src string, k Kind) {
	t.Helper()
	var got []Token
	for _, tok := range Tokenize([]byte(src)) {
		if tok.Kind != Whitespace {
			got = append(got, tok)
		}
	}
	if len(got) == 0 {
		t.Errorf("%q: no tokens", src)
		return
	}
	last := got[len(got)-1]
	if last.Kind != k {
		t.Errorf("%q: last token kind %v, want %v (all: %v)", src, last.Kind, k, got)
	}
	if last.End != len(src) {
		t.Errorf("%q: last token ends at %d, want %d -- the identifier is cut short",
			src, last.End, len(src))
	}
}

// TestIdentifierASCIIUnaffected: the pragma must not change ASCII lexing.
// This is the control. Without it, a test suite passes by widening the class
// to "anything", and the 34 corpus files would still fail for a different
// reason.
func TestIdentifierASCIIUnaffected(t *testing.T) {
	for _, src := range []string{
		`$foo`,
		`@bar_baz`,
		`%h2`,
		`$_x`,
		`$Foo::Bar`,
	} {
		lastToken(t, src, Variable)
		lastToken(t, "use utf8;\n"+src, Variable) // same span, pragma or not
	}
}

// TestIdentifierUTF8Corpus is §0.13 rank 1: 34 files, `t/mro/*_utf8.t` and
// most of `t/uni/`.
//
// The rule is perldata.pod:179-180 and it is an INTERSECTION, not an
// extension:
//
//	(\p{Word} & \p{XID_Start}) + [_]      then      (\p{Word} & \p{XID_Continue})*
//
// An earlier draft of this issue said "XID with perl-specific additions, and
// the additions are where the corpus files live". That was wrong and would
// have sent the implementer hunting for extensions that do not exist. All
// four forms below are plain XID letters.
func TestIdentifierUTF8Corpus(t *testing.T) {
	for _, src := range []string{
		`$Føø`,
		`@ᕘ`,
		`$Àlìcè`,
		`%Ƒ운ℭ`,
	} {
		lastToken(t, "use utf8;\n"+src, Variable)
	}

	// The package-qualified forms the corpus actually contains.
	for _, src := range []string{
		`$Føø::Bær`,
		`@ᕘ::ISA`,
		`$Àlìcè::VERSION`,
	} {
		lastToken(t, "use utf8;\n"+src, Variable)
	}
}

// TestIdentifierRequiresPragma: without `use utf8`, a byte outside the
// Latin-1 word characters cannot start or continue an identifier.
//
// The wording matters and an earlier draft got it wrong. A LATIN-1 letter
// byte is still a word character without the pragma -- perl lexes `$á` fine
// and fails later in the compiler. Measured:
//
//	$ perl -e 'my $á = 7;'
//	Can't use global $<?> in "my" ...
//	Unrecognized character \xA1 ...
//
// It is the SECOND byte of the UTF-8 sequence that is unrecognised, because
// the first (\xC3) is a Latin-1 word character. So the test asserts an Error
// token appears, not that the whole sequence is one.
func TestIdentifierRequiresPragma(t *testing.T) {
	// No pragma: the multi-byte sequence cannot lex as one identifier.
	for _, src := range []string{
		`$Føø`,
		`@ᕘ`,
	} {
		var sawError bool
		for _, tok := range Tokenize([]byte(src)) {
			if tok.Kind == Error {
				sawError = true
			}
		}
		if !sawError {
			t.Errorf("%q without use utf8: no Error token; a non-Latin-1 byte "+
				"cannot be part of an identifier", src)
		}
	}

	// With the pragma, the same input is clean.
	for _, src := range []string{
		"use utf8;\n$Føø",
		"use utf8;\n@ᕘ",
	} {
		for _, tok := range Tokenize([]byte(src)) {
			if tok.Kind == Error {
				t.Errorf("%q: Error token under use utf8; the class must accept "+
					"XID letters", src)
			}
		}
	}
}

// TestIdentifierEveryPosition encodes the discriminating measurement from
// §0.13, and it is the test that catches a fix applied in only one place.
//
// `sub ᕘ { 1 }` alone PARSES on the current tree-sitter grammar, but
// `(shift)->SUPER::ᕘ` does not. So the defect is the character class at every
// identifier position, not a special case after `package`. A fix that only
// widens the post-`package` path passes a naive test and leaves 30 files
// failing.
func TestIdentifierEveryPosition(t *testing.T) {
	for _, tc := range []struct{ name, src string }{
		{"after package", "use utf8;\npackage Føø;"},
		{"after sub", "use utf8;\nsub ᕘ { 1 }"},
		{"after arrow", "use utf8;\n$o->ᕘ;"},
		{"after SUPER", "use utf8;\n$o->SUPER::ᕘ;"},
		{"glob assignment", "use utf8;\n*ᕘ::ᕘ_Ƒ = sub {};"},
		{"array element", "use utf8;\n$ᕘ[0];"},
		{"hash key", "use utf8;\n$h{ᕘ};"},
		{"bareword call", "use utf8;\nᕘ();"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, tok := range Tokenize([]byte(tc.src)) {
				if tok.Kind == Error {
					t.Errorf("%q: Error token at [%d,%d) %q — the class must hold "+
						"at every identifier position, not only after `package`",
						tc.src, tok.Start, tok.End, tc.src[tok.Start:tok.End])
				}
			}
		})
	}
}

// TestIdentifierPackageSeparators: §2.6.3 and §2.15 item 4 require BOTH
// separators. §0.13 rank 10, two T2 files, and it changes DELIMITATION rather
// than merely naming:
//
//	comp/package.t:17    $main'a = $'b;
//	comp/parser.t:397    sub CORE'print'foo { 43 }
//
// A lexer unaware of the apostrophe opens a string at `'a = $'`. Measured:
//
//	$ perl -e '$main::a = 5; print $main'"'"'a'
//	5
//
// The same corpus line contains `$'`, the postmatch variable, so the rule
// that an apostrophe separates only BETWEEN identifier characters is
// exercised by the corpus rather than by a fixture:
//
//	$ perl -e '"abc" =~ /b/; print $'"'"''
//	c
func TestIdentifierPackageSeparators(t *testing.T) {
	// Both separators name a package variable.
	lastToken(t, `$main::a`, Variable)
	lastToken(t, `$main'a`, Variable)
	lastToken(t, `@Foo'Bar'baz`, Variable)

	// `$'` is the postmatch variable: the apostrophe is the NAME, not a
	// separator, because no identifier character precedes it.
	lastToken(t, `$'`, Variable)

	// The corpus line that needs both readings at once. A lexer that treats
	// every apostrophe as a separator swallows `a = $` as part of a name; one
	// that treats every apostrophe as a quote opens a string. Neither is
	// right, and the span of the first token says which happened.
	const src = `$main'a = $'`
	toks := Tokenize([]byte(src))
	if len(toks) == 0 {
		t.Fatal("no tokens")
	}
	if toks[0].Kind != Variable || toks[0].End != 7 {
		t.Errorf("%q: first token %v [%d,%d), want Variable [0,7) — `$main'a` is "+
			"one package-qualified name", src, toks[0].Kind, toks[0].Start, toks[0].End)
	}
	// And no Quote token: the apostrophes are never string delimiters here.
	for _, tok := range toks {
		if tok.Kind == Quote {
			t.Errorf("%q: an apostrophe opened a string at [%d,%d); both are "+
				"package separators or variable names", src, tok.Start, tok.End)
		}
	}
}
