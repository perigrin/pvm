// ABOUTME: Tests the word readings perl.git t/ showed the token table missing: quote-like-named lexical subs, and the
// ABOUTME: { after a word and a * (uni/parser.t's `is *{$_[0]}`), which toke.c reads as a glob's block after a call.

package chart

import (
	"slices"
	"strings"
	"testing"
)

// internal/lexer emits a Word for a quote-like's name once a lexical sub of that name is declared, and toke.c reads
// the lexical sub before the quote-like (it checks the pad first), so the word is a call. perl -c (94e50866) accepts
// each; the first two are op/lexsub.t's forms and tr's op/lvref.t's s.
func TestQuoteOpNamedSubs(t *testing.T) {
	rows := map[string]bool{}
	for _, src := range []string{
		"my sub s { 1 } s(1);", "my sub q { 1 } q();", "my sub y { 1 } my @a = (y, 1);", "my sub tr { 1 } tr(1);",
		"my sub qw { 1 } qw(1);", "my sub m { 1 } m(1);", "my sub qq { 1 } qq(1);", "my sub qr { 1 } qr(1);",
		"my sub qx { 1 } qx(1);",
	} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at %d (%s), expecting %v; perl accepts it", src, v.at, v.by, v.expected)
		}
		rows[src] = true
	}
	checkParses(t, rows)
}

// A quote-like the lexer reads as a quote keeps its quote reading and gains no word's: perl -c accepts each as a quote.
func TestQuoteOpNamedSubsLeaveQuotes(t *testing.T) {
	for _, tc := range []struct {
		src string
		at  int
	}{{"my $x = q(a);", 8}, {"$x =~ s/a/b/;", 6}, {"y/a/b/;", 0}} {
		if v := recognizeSource(t, []byte(tc.src)); !v.accepted {
			t.Errorf("%q is refused at %d; perl accepts it", tc.src, v.at)
		}
		got := terminalsAt(t, tc.src, tc.at)
		for _, w := range []string{"BAREWORD", "PRIVATEREF", "NOAMP", "FUNC0SUB", "METHCALL0"} {
			if slices.Contains(got, w) {
				t.Errorf("%q's quote admits %v, a word's %s among them", tc.src, got, w)
			}
		}
	}
}

// After a call toke.c leaves XTERM, so a * is a glob's sigil and the { after it a block (yyl_star's PREREF); after a
// word scope leaves undeclared it leaves XOPERATOR, a multiplication and an anonymous hash. The token table cannot see
// scope, so that { carries both and resolve picks. perl -c accepts each row: uni/parser.t 777's form, then the forms
// the SCOPESUBS gate found after a declared sub, a filehandle and a method's class, then the undeclared word's.
func TestPerlGitWords(t *testing.T) {
	rows := map[string]bool{}
	for _, src := range []string{
		"sub is {} is *{$_[0]};", "is *{$_[0]};", "sub foo {} foo *{$x};", `print STDOUT *{"x"};`, `new Foo *{"x"};`,
		// a quote-like's name has a text of its own, so the { after its * needs a class of its own (quoteOpWords)
		`my sub q { 1 } q *{"x"};`, `my sub s { 1 } s *{$x};`,
	} {
		if got := terminalsAt(t, src, strings.Index(src, "*{")+1); !slices.Contains(got, "PERLY_BRACE_OPEN") {
			t.Errorf("the { of %q admits %v, no PERLY_BRACE_OPEN", src, got)
		}
		rows[src] = true
	}
	checkParses(t, rows)
}

// The block reading is the * after a word's alone: a { after = or return, or after a * that follows a term, stays an
// anonymous hash, as toke.c reads it (XTERM or XOPERATOR, not PREREF).
func TestPerlGitWordsBounded(t *testing.T) {
	for _, src := range []string{"my $y = {a => 1};", "return {a => 1};", "my $y = $x * {a => 1};", "my $y = f() * {a => 1};"} {
		if got := terminalsAt(t, src, strings.LastIndex(src, "{")); !slices.Equal(got, []string{"HASHBRACK"}) {
			t.Errorf("the { of %q admits %v, want only HASHBRACK", src, got)
		}
	}
}

// A lexical sub overrides every keyword where toke.c does not expect an operator (it checks the pad first,
// yyl_keylookup): internal/lexer marks that word LexSub and chart/lex reads it as a word no class names. perl.git
// t/op/lexsub.t:72's form and its neighbours, each as perl -c (94e50866) gives it.
func TestLexSubOverridesKeyword(t *testing.T) {
	checkParses(t, map[string]bool{
		"{ our sub if() { 42 } my $x = if if if; }": true,
		"our sub if() { 42 } my $x = if if if;":     true,
		"my sub if { 42 } my $x = if(1);":           true,
		"my sub if { 42 } my @a = (if, 1);":         true,
		"my sub unless { 1 } my $z = unless(2);":    true,
		"my sub if { 1 } my $z = 1 if 1;":           true,
		"my $x = if if if;":                         false,
		"{ our sub if() { 42 } } my $x = if if if;": false,
		"my sub if { 1 } if (1) { 2 }":              false,
	})
}
