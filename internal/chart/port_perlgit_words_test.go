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

// toke.c consumes a typed my's class name and emits no token for it (KEY_my, PL_in_my_stash), but croaks "No such
// class" when no stash of that name exists yet; chart/lex drops the name where the file has made the package before
// it or perl starts with it. perl.git t/op/multideref.t's form and its neighbours, each as perl -c (94e50866) gives it.
func TestTypedMy(t *testing.T) {
	checkParses(t, map[string]bool{
		"{ package Foo; our %FIELDS = (); my Foo $f; }":    true,
		"package Foo; package main; my Foo $f = 1;":        true,
		"package Foo::Bar; package main; our Foo::Bar $f;": true,
		"package Foo; package main; my Foo ($a, $b);":      true,
		"package Foo; package main; my Foo @a;":            true,
		"package Foo; package main; my Foo:: $f;":          true,
		"package Foo; package main; for my Foo $f (1) {}":  true,
		"$Foo::x = 1; my Foo $f;":                          true,
		"sub Foo::Bar::f {} my Foo::Bar $f;":               true,
		"require Foo; my Foo $f;":                          true,
		"my main $f;":                                      true,
		"my __PACKAGE__ $f;":                               true,
		"my UNIVERSAL $f;":                                 true,
		"my Nope $f;":                                      false,
		"my Carp $f;":                                      false,
		"package Foo; package main; my Foo::Bar $f;":       false,
		"my Foo $f; package Foo;":                          false,
		"my Foo::Bar $f; Foo::Bar::f();":                   false,
		"package Foo; package main; my Foo;":               false,
		"my $x = 'Foo'; Foo->new; my Foo $f;":              false,
		// a bareword or class name makes no stash; a variable, a &-sigilled word, a call with ( and a sub's name do
		"Foo::bar; my Foo $f;":          false,
		"Foo::Bar->x; my Foo $f;":       false,
		"my $x = Foo::; my Foo $f;":     false,
		"Foo::->x; my Foo $f;":          false,
		"$x->Foo::bar; my Foo $f;":      false,
		"Foo::bar(); my Foo $f;":        true,
		"Foo::bar (); my Foo $f;":       true,
		"&Foo::bar; my Foo $f;":         true,
		`\&Foo::bar; my Foo $f;`:        true,
		"@Foo::x = (); my Foo $f;":      true,
		"$Foo'x = 1; my Foo $f;":        true,
		"$main::Foo::x = 1; my Foo $f;": true,
		"$::Foo::x = 1; my Foo $f;":     true,
		`*Foo::x = \1; my Foo $f;`:      true,
		// a method name makes no stash, with ( or not; a * that is a multiply or ** is no glob's sigil
		"$x->Foo::bar(); my Foo $f;":        false,
		"$x->Foo::bar (); my Foo $f;":       false,
		"my $y = $a *Foo::bar; my Foo $f;":  false,
		"my $y = $a*Foo::bar; my Foo $f;":   false,
		"my $y = $a **Foo::bar; my Foo $f;": false,
		`* Foo::x = \1; my Foo $f;`:         true,
		`*{Foo::x} = \1; my Foo $f;`:        true,
		`* { Foo::x } = \1; my Foo $f;`:     true,
		"(*Foo::x); my Foo $f;":             true,
		"Foo::bar->(); my Foo $f;":          true,
		"Foo::bar ->(); my Foo $f;":         true,
		"Foo::bar-> (); my Foo $f;":         true,
		"Foo::Bar->x(); my Foo $f;":         false,
		"$x->$Foo::m(); my Foo $f;":         true,
		// the * after a repetition x is a glob's sigil
		"my $y = $z x *Foo::x; my Foo $f;":  true,
		"my $y = $z x*Foo::x; my Foo $f;":   true,
		"my @y = (1) x *Foo::x; my Foo $f;": true,
		// a word before -> is a class name only where a method follows; a deref makes it a call
		"Foo::bar->[0]; my Foo $f;":  true,
		"Foo::bar->{a}; my Foo $f;":  true,
		"Foo::bar->@*; my Foo $f;":   true,
		"Foo::bar->%*; my Foo $f;":   true,
		"Foo::bar->$*; my Foo $f;":   true,
		"Foo::bar->$#*; my Foo $f;":  true,
		"Foo::bar->**; my Foo $f;":   true,
		"Foo::bar->$m(); my Foo $f;": false,
		// neighbours are significant tokens: a newline or a comment between them does not count
		"Foo::bar\n(); my Foo $f;":     true,
		"Foo::bar # c\n(); my Foo $f;": true,
		"Foo::bar->\n(); my Foo $f;":   true,
		"Foo::bar ->\n(); my Foo $f;":  true,
		// a word alone in the braces after any sigil or cast is a name; a leading :: makes a block of a string
		"&{Foo::bar}; my Foo $f;":         true,
		"&{Foo::bar}(); my Foo $f;":       true,
		`\&{Foo::bar}; my Foo $f;`:        true,
		"@{Foo::x} = (); my Foo $f;":      true,
		"%{Foo::x} = (); my Foo $f;":      true,
		"${Foo::x} = 1; my Foo $f;":       true,
		`*{::Foo::x} = \1; my Foo $f;`:    false,
		`@{Foo::x . ""} = (); my Foo $f;`: false,
		// package, class, use, no and require make each parent stash too, their name read as perl reads it
		"package Foo::Bar; package main; my Foo $f;":                                            true,
		"use v5.38; use feature 'class'; no warnings; class Foo::Bar; package main; my Foo $f;": true,
		"require Foo::Bar if 0; my Foo $f;":                                                     true,
		"use File::Spec; my File $f;":                                                           true,
		"package Foo'Bar; package main; my Foo::Bar $f;":                                        true,
		"package Foo'Bar; package main; my Foo $f;":                                             true,
		"package ::Foo; package main; my Foo $f;":                                               true,
		"package main::Foo; package main; my Foo $f;":                                           true,
		"format Foo::x =\n.\nmy Foo $f;":                                                        true,
		// the class word is read as perl reads it: main:: names the same stash; a leading :: is a syntax error
		"package Foo; package main; my main::Foo $f;":           true,
		"package Foo::Bar; package main; my main::Foo::Bar $f;": true,
		"my main::main $f;":                       true,
		"my main::UNIVERSAL $f;":                  true,
		"package Foo; package main; my ::Foo $f;": false,
		"my main:: $f;":                           true,
		"my main::main:: $f;":                     true,
		// perl strips one leading :: and then any number of main::, nothing more; ' is ::
		"package main::::Foo; package main; my Foo $f;":         false,
		"$main::::Foo::x = 1; my Foo $f;":                       false,
		"$::::Foo::x = 1; my Foo $f;":                           false,
		"require main::::Foo if 0; my Foo $f;":                  false,
		"package Foo; package main; my main::::Foo $f;":         false,
		"package Foo; package main; my main::main::::Foo $f;":   false,
		"package Foo; package main; my main::'Foo $f;":          false,
		"package main::::Foo; package main; my main::::Foo $f;": true,
		"package Foo; package main; my main'main'Foo $f;":       true,
		// a name's trailing :: is a stash inside it (Foo::, or the stash named "" in main::::); the class word drops one
		"$Foo::::x = 1; my Foo:::: $f;":                           true,
		"package Foo::; package main; my Foo:::: $f;":             true,
		"package Foo; package main; my Foo:::: $f;":               false,
		"my main:::: $f;":                                         false,
		"package main::; package main; my main:::: $f;":           true,
		"$::::x = 1; my main:::: $f;":                             true,
		"package main::::Foo; package main; my main:::: $f;":      true,
		"package main::::Foo; package main; my main::::Foo:: $f;": true,
		"$::x = 1; my main:::: $f;":                               false,
		// a version after use, no or require names no stash
		"use v5; my v5 $f;":                       false,
		"use v6; my v6 $f;":                       false,
		"use v5 (); my v5 $f;":                    false,
		"no v5; my v5 $f;":                        false,
		"require v5; my v5 $f;":                   false,
		"require v5_1 if 0; my v5_1 $f;":          false,
		"require v1_0 if 0; my v1_0 $f;":          false,
		"require v_1 if 0; my v_1 $f;":            true,
		"package v5_1; package main; my v5_1 $f;": true,
		"package v5; my v5 $f;":                   true,
		// CORE::kw is read as kw
		"CORE::require Foo::Bar if 0; my Foo $f;":         true,
		"CORE::package Foo; package main; my Foo $f;":     true,
		"package Foo; package main; CORE::my Foo $f;":     true,
		"CORE::GLOBAL::require Foo::Bar if 0; my Foo $f;": false,
	})
}

// Known limits (the shortcut on chart/lex's typedClass): a stash made by a string eval, an interpolated variable, an
// s///e replacement or a print's filehandle is not seen, so its typed my is refused where perl -c (94e50866) accepts each.
func TestTypedMyKnownLimit(t *testing.T) {
	for _, src := range []string{`BEGIN { eval "package Foo; 1" } my Foo $f;`, `my $s = "$Foo::x"; my Foo $f;`,
		`my $s = "@Foo::x"; my Foo $f;`, `$_ =~ s/a/Foo::bar()/e; my Foo $f;`, `print Foo::x "a"; my Foo $f;`} {
		if pipelineSource(t, src) {
			t.Errorf("%q is accepted; the known limit is gone, so drop the shortcut and the README line", src)
		}
	}
}

// A glob's name may be digits: after a * that is a glob sigil yyl_star's scan_ident reads them as the name and
// forces BAREWORD (perl -DT on `local *1 = sub { 1 };`: BAREWORD "1", PERLY_STAR), where internal/lexer emits a
// Number. perl.git t/op/method.t's `local *1 = sub { 123 }` and its neighbours, each as perl -c (94e50866) gives it:
// only decimal digits with no leading 0 name a glob.
func TestDigitGlob(t *testing.T) {
	for _, src := range []string{"local *1 = sub { 1 };", "*12 = \\&f;", "my $x = *1;"} {
		if got := terminalsAt(t, src, strings.Index(src, "*")+1); !slices.Contains(got, "BAREWORD") {
			t.Errorf("the digits after the glob * of %q admit %v, no BAREWORD", src, got)
		}
	}
	if got := terminalsAt(t, "my $x = 2 *1;", 11); slices.Contains(got, "BAREWORD") {
		t.Errorf("the 1 after a multiplying * admits %v, BAREWORD among them", got)
	}
	checkParses(t, map[string]bool{
		"local *1 = sub { 1 };": true,
		"*12 = \\&f;":           true,
		"my $x = *1;":           true,
		"print *1{CODE};":       true,
		"my $y = $z x *1;":      true,
		"my $x = 2 *1;":         true,
		"*1x = 1;":              false,
		// scan_ident takes decimal digits only, and no second digit after a leading 0 ("Numeric variables with
		// more than one digit may not start with '0'"); a glob's * before any other number is refused.
		"*1_0 = 1;":     false,
		"*1e3 = 1;":     false,
		"*01 = 1;":      false,
		"*00 = 1;":      false,
		"(*1, 2);":      true,
		"my $y = 2*01;": true,
	})
}
