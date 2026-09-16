// ABOUTME: The keyword table and the brace stack: the two expect-state gaps M0 recorded and deferred.
// ABOUTME: Both decide term-vs-operator, and both were guessing before this.

package lexer

import "testing"

// TestKeywordTableClassifies pins the table against perl.
//
// The lexer needs exactly one bit per keyword: does a value follow it, or an
// operator? That is coarser than the parser's named-unary-versus-list-operator
// distinction, and deliberately so -- the expect state has two outcomes, and a
// table with more categories than outcomes invites a third that nothing reads.
//
// The 21 niladic keywords were measured rather than recalled. Every one of
// perl's 251 keywords was tried as `my $x = KEYWORD / 2 ;` under `use v5.38`,
// and exactly these 21 compiled -- meaning perl read the slash as division,
// which it does only when no argument is expected.
func TestKeywordTableClassifies(t *testing.T) {
	for _, k := range []string{
		"time", "times", "wantarray", "wait", "fork", "getppid", "getlogin",
		"getgrent", "gethostent", "getnetent", "getprotoent", "getpwent",
		"getservent", "endgrent", "endhostent", "endnetent", "endprotoent",
		"endpwent", "endservent", "setgrent", "setpwent",
	} {
		if !isNiladic(k) {
			t.Errorf("%q should be niladic: perl reads a following slash as division", k)
		}
	}

	// Keywords that DO take an argument, so a following slash opens a
	// pattern. `split` is the one the M0 comment named.
	for _, k := range []string{
		"split", "grep", "map", "sort", "return", "print", "push", "join",
		"length", "uc", "lc", "ref", "defined", "scalar", "keys", "values",
		"shift", "pop", "chomp", "die", "warn",
	} {
		if isNiladic(k) {
			t.Errorf("%q takes an argument: perl reads a following slash as a pattern", k)
		}
	}

	// Feature-gated keywords are NOT in the table. Under `use v5.38` they
	// take arguments, and without it they are ordinary identifiers; either
	// way "niladic" is the wrong answer. §0.13 rank 6 counts 5 corpus files
	// using them as plain identifiers.
	for _, k := range []string{"class", "method", "field", "try", "defer", "say", "state", "isa"} {
		if isNiladic(k) {
			t.Errorf("%q is feature-gated and must not be tabled as niladic", k)
		}
	}
}

// TestKeywordAfterNiladicIsOperator is the gap itself, as measured before the
// fix:
//
//	"my $t = time / 2;"  ->  Word "time"   UnknownRest "/ 2;"
//
// The lexer guessed term, opened a pattern at the slash, and swallowed the
// rest of the input.
func TestKeywordAfterNiladicIsOperator(t *testing.T) {
	src := "my $t = time / 2;"
	toks := Tokenize([]byte(src))

	for _, tok := range toks {
		if tok.Kind == UnknownRest || tok.Kind == Quote {
			t.Fatalf("%q: lexed %v %q; the slash after a niladic keyword is division",
				src, tok.Kind, src[tok.Start:tok.End])
		}
	}

	// And positively: the slash is an Operator on its own.
	if !hasToken(toks, []byte(src), Operator, "/") {
		t.Errorf("%q: no Operator %q in %s", src, "/", renderKinds(toks, []byte(src)))
	}
}

// TestSlashAfterArgTakingKeywordIsStillAPattern is the other direction, and
// it is the one a careless table breaks. `split /,/` must stay a pattern;
// the M0 comment chose XTerm for every Word precisely to protect it.
func TestSlashAfterArgTakingKeywordIsStillAPattern(t *testing.T) {
	src := "my @p = split /,/, $s;"
	toks := Tokenize([]byte(src))
	if !hasKind(toks, Quote) {
		t.Errorf("%q: no Quote token; the slash after split must open a pattern: %s",
			src, renderKinds(toks, []byte(src)))
	}
}

// TestBraceStackTracksBlocks: a `}` closing a block and a `}` closing a
// subscript leave different states behind. perl keeps PL_lex_brackstack for
// exactly this; the lexer had no such stack and returned XOperator for every
// CloseBracket.
func TestBraceStackTracksBlocks(t *testing.T) {
	// A subscript: after `}` an operator is expected, so `/` is division.
	src := "$h{a} / 2;"
	toks := Tokenize([]byte(src))
	if hasKind(toks, Quote) || hasKind(toks, UnknownRest) {
		t.Errorf("%q: the slash after a subscript is division: %s",
			src, renderKinds(toks, []byte(src)))
	}
}

// TestTermAfterBlockClose is the measured gap:
//
//	"if ($x) { 1 }\n%h = ();"  ->  CloseBracket "}"  Operator "%"  Word "h"
//
// A statement after a block starts in TERM position, so `%h` is a hash and
// `/foo/` is a pattern. The lexer read both as operators.
func TestTermAfterBlockClose(t *testing.T) {
	src := "if ($x) { 1 }\n%h = ();"
	toks := Tokenize([]byte(src))
	if !hasToken(toks, []byte(src), Variable, "%h") {
		t.Errorf("%q: %%h is a hash variable, not modulus: %s",
			src, renderKinds(toks, []byte(src)))
	}

	src = "if ($x) { 1 }\n/foo/ and die;"
	toks = Tokenize([]byte(src))
	if !hasKind(toks, Quote) {
		t.Errorf("%q: /foo/ is a pattern, not division: %s",
			src, renderKinds(toks, []byte(src)))
	}
}

// TestModulusAfterValueStillWorks guards the other direction. A brace stack
// that answered "block" everywhere would turn real modulus into a hash.
func TestModulusAfterValueStillWorks(t *testing.T) {
	src := "my $n = 5 % 2;"
	toks := Tokenize([]byte(src))
	if hasToken(toks, []byte(src), Variable, "% 2") || !hasToken(toks, []byte(src), Operator, "%") {
		t.Errorf("%q: %% after a number is modulus: %s", src, renderKinds(toks, []byte(src)))
	}
}

// --- helpers ---

func hasToken(toks []Token, src []byte, k Kind, text string) bool {
	for _, tok := range toks {
		if tok.Kind == k && string(src[tok.Start:tok.End]) == text {
			return true
		}
	}
	return false
}

func hasKind(toks []Token, k Kind) bool {
	for _, tok := range toks {
		if tok.Kind == k {
			return true
		}
	}
	return false
}

func renderKinds(toks []Token, src []byte) string {
	out := ""
	for _, tok := range toks {
		if tok.Kind == Whitespace {
			continue
		}
		out += tok.Kind.String() + "(" + string(src[tok.Start:tok.End]) + ") "
	}
	return out
}
