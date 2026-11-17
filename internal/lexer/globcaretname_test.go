// ABOUTME: A caret control name after a sigil -- `*^R`, `&^R`, `*{^R}`, `$^?` -- is one name, as toke.c's scan_ident reads it.
// ABOUTME: Pins each shape against perl -c and -DT at 94e50866, with the multiply, lowercase and digit refusals.

package lexer_test

import "testing"

// caretRow is one source and its significant tokens. Every row's perl -c
// verdict is in the comment beside it, read with the pinned
// /opt/perl-94e50866-debug/bin/perl.
type caretRow struct {
	src  string
	want []string
}

func checkCaretRows(t *testing.T, rows []caretRow) {
	t.Helper()
	for _, c := range rows {
		if !streamIs(c.src, c.want...) {
			t.Errorf("%q lexes as %v, want %v", c.src, significant(c.src), c.want)
		}
	}
}

// TestGlobCaretName: a glob named by a caret control name is the star and one
// name, as `*foo` is. toke.c's yyl_star calls scan_ident, which skips space,
// then takes `^` and one isCONTROLVAR byte (A-Z or one of `[ \ ] ^ _ ?`) as
// one control character: -DT on `* ^R = 1;` shows BAREWORD PV("\22") forced
// before PERLY_STAR. A braced caret name is word characters after that byte,
// as bracedNameFollowsAt reads it for `${^...}`.
func TestGlobCaretName(t *testing.T) {
	g := func(src, name string, rest ...string) caretRow {
		return caretRow{src, append([]string{"Operator(*)", "Word(" + name + ")"}, rest...)}
	}
	eq1 := []string{"Operator(=)", "Number(1)", "Semicolon(;)"}
	checkCaretRows(t, []caretRow{
		g("*^R = 1;", "^R", eq1...), // syntax OK (re/pat.t's shape)
		g("*^W = 1;", "^W", eq1...), // syntax OK
		g("*^[ = 1;", "^[", eq1...), // syntax OK
		g("*^] = 1;", "^]", eq1...), // syntax OK
		g(`*^\ = 1;`, `^\`, eq1...), // syntax OK
		g("*^? = 1;", "^?", eq1...), // syntax OK
		g("*^_ = 1;", "^_", eq1...), // syntax OK
		g("*^^ = 1;", "^^", eq1...), // syntax OK
		g("*^ = 1;", "^", eq1...),   // syntax OK: the glob named ^
		// A bare ^ is the name across space too, as scan_ident takes the
		// caret after skipspace: -DT on `* ^;` shows BAREWORD PV("^") forced
		// before PERLY_STAR. syntax OK.
		g("* ^ = 1;", "^", eq1...),
		g("* ^;", "^", "Semicolon(;)"),
		{"*\n^;", []string{"Operator(*)", "Word(^)", "Semicolon(;)"}},
		{"* # c\n^;", []string{"Operator(*)", "Comment(# c)", "Word(^)", "Semicolon(;)"}},
		// syntax OK: the subscript follows the name.
		g("*^R{CODE};", "^R", "Operator({)", "Word(CODE)", "CloseBracket(})", "Semicolon(;)"),
		// Space and comments before the name: scan_ident's skipspace. syntax OK.
		{"* ^R = 1;", append([]string{"Operator(*)", "Word(^R)"}, eq1...)},
		{"*\n^R = 1;", append([]string{"Operator(*)", "Word(^R)"}, eq1...)},
		{"* # c\n^R = 1;", append([]string{"Operator(*)", "Comment(# c)", "Word(^R)"}, eq1...)},
		// Braced: syntax OK for every row.
		g("*{^R} = 1;", "{^R}", eq1...),
		g("*{ ^R } = 1;", "{ ^R }", eq1...),
		g("*{^[} = 1;", "{^[}", eq1...),
		g("*{^} = 1;", "{^}", eq1...),
		g("*{^^x} = 1;", "{^^x}", eq1...),
		g("*{^]x} = 1;", "{^]x}", eq1...),
		g("*{^RE} = 1;", "{^RE}", eq1...),
		{"* {^R} = 1;", append([]string{"Operator(*)", "Word({^R})"}, eq1...)},
		// In term positions after other tokens: syntax OK.
		{"my $x = *^R;", []string{"Word(my)", "Variable($x)", "Operator(=)", "Operator(*)", "Word(^R)", "Semicolon(;)"}},
		{`\*^R;`, []string{`Operator(\)`, "Operator(*)", "Word(^R)", "Semicolon(;)"}},
		{"$x x *^R;", []string{"Variable($x)", "Word(x)", "Operator(*)", "Word(^R)", "Semicolon(;)"}},
		{"print STDOUT *^R;", []string{"Word(print)", "Word(STDOUT)", "Operator(*)", "Word(^R)", "Semicolon(;)"}},
	})
}

// TestGlobCaretNameRefusals: where the star multiplies, or the byte after the
// caret is no isCONTROLVAR byte, there is no caret glob name, and perl -c
// refuses each row.
func TestGlobCaretNameRefusals(t *testing.T) {
	checkCaretRows(t, []caretRow{
		// syntax error near "*^": the star multiplies and ^ is xor.
		{"$a *^R;", []string{"Variable($a)", "Operator(*)", "Operator(^)", "Word(R)", "Semicolon(;)"}},
		{"2*^R;", []string{"Number(2)", "Operator(*)", "Operator(^)", "Word(R)", "Semicolon(;)"}},
		{"$a * ^R;", []string{"Variable($a)", "Operator(*)", "Operator(^)", "Word(R)", "Semicolon(;)"}},
		// Bareword or Number found where operator expected: the glob is *^.
		{"*^r = 1;", []string{"Operator(*)", "Word(^)", "Word(r)", "Operator(=)", "Number(1)", "Semicolon(;)"}},
		{"*^1 = 1;", []string{"Operator(*)", "Word(^)", "Number(1)", "Operator(=)", "Number(1)", "Semicolon(;)"}},
		// Bareword found before "E": an unbraced caret name is one byte.
		{"*^RE = 1;", []string{"Operator(*)", "Word(^R)", "Word(E)", "Operator(=)", "Number(1)", "Semicolon(;)"}},
		// syntax error near "{^": no name, so the brace opens a block.
		{"*{^r} = 1;", []string{"Operator(*)", "Operator({)", "Operator(^)", "Word(r)", "CloseBracket(})", "Operator(=)", "Number(1)", "Semicolon(;)"}},
		{"*{^1} = 1;", []string{"Operator(*)", "Operator({)", "Operator(^)", "Number(1)", "CloseBracket(})", "Operator(=)", "Number(1)", "Semicolon(;)"}},
		// The bare glob ^ is one byte too: syntax error, Bareword found where
		// operator expected before "r".
		{"* ^r = 1;", []string{"Operator(*)", "Word(^)", "Word(r)", "Operator(=)", "Number(1)", "Semicolon(;)"}},
	})
}

// TestGlobCaretNameAfterWord: after a word no declaration makes a sub, toke.c
// has PL_expect at XOPERATOR and the star multiplies: -DT on `A *^R;` shows
// BAREWORD then MULOP, and perl -c refuses every undeclared row with a syntax
// error. A sub declared before the word takes the glob, its prototype not
// empty: syntax OK on the two `sub A` rows. A constant's word ends a term,
// so `use constant A=>1; A *^R;` is refused too.
func TestGlobCaretNameAfterWord(t *testing.T) {
	mul := func(src string, pre ...string) caretRow {
		return caretRow{src, append(pre, "Operator(*)", "Operator(^)", "Word(R)", "Semicolon(;)")}
	}
	glob := func(pre ...string) []string { return append(pre, "Operator(*)", "Word(^R)", "Semicolon(;)") }
	a := []string{"Word(A)"}
	checkCaretRows(t, []caretRow{
		mul("A*^R;", a...),
		mul("A * ^R;", a...),
		mul("Foo::Bar *^R;", "Word(Foo::Bar)"),
		mul("use constant A=>1; A *^R;", "Word(use)", "Word(constant)", "Word(A)", "Operator(=>)", "Number(1)", "Semicolon(;)", "Word(A)"),
		{"A*{^R};", []string{"Word(A)", "Operator(*)", "Operator({)", "Operator(^)", "Word(R)", "CloseBracket(})", "Semicolon(;)"}},
		{"sub A {} A *^R;", []string{"Word(sub)", "Word(A)", "Operator({)", "CloseBracket(})", "Word(A)", "Operator(*)", "Word(^R)", "Semicolon(;)"}},
		{"sub A; A *^R;", []string{"Word(sub)", "Word(A)", "Semicolon(;)", "Word(A)", "Operator(*)", "Word(^R)", "Semicolon(;)"}},
		// A print's filehandle expects a term after it too: syntax OK.
		{"print STDOUT * ^R;", []string{"Word(print)", "Word(STDOUT)", "Operator(*)", "Word(^R)", "Semicolon(;)"}},
		{"print(STDOUT *{^R});", []string{"Word(print)", "Operator(()", "Word(STDOUT)", "Operator(*)", "Word({^R})", "CloseBracket())", "Semicolon(;)"}},
		// A sub's name is its package's: `::A`, `main::A` and `'` for `::`
		// name it in full, and a bare name is in the package in force where
		// it is written. syntax OK on these.
		{"sub main::A; A *^R;", glob("Word(sub)", "Word(main::A)", "Semicolon(;)", "Word(A)")},
		{"sub A; main::A *^R;", glob("Word(sub)", "Word(A)", "Semicolon(;)", "Word(main::A)")},
		{"sub A; ::A *^R;", glob("Word(sub)", "Word(A)", "Semicolon(;)", "Word(::A)")},
		{"sub A'B; A::B *^R;", glob("Word(sub)", "Word(A'B)", "Semicolon(;)", "Word(A::B)")},
		{"sub Foo::A; package Foo; A *^R;", glob("Word(sub)", "Word(Foo::A)", "Semicolon(;)", "Word(package)", "Word(Foo)", "Semicolon(;)", "Word(A)")},
		{"package Foo; sub A; package main; Foo::A *^R;", glob("Word(package)", "Word(Foo)", "Semicolon(;)", "Word(sub)", "Word(A)", "Semicolon(;)", "Word(package)", "Word(main)", "Semicolon(;)", "Word(Foo::A)")},
		{"package Foo { sub A {} A *^R; }", []string{"Word(package)", "Word(Foo)", "Operator({)", "Word(sub)", "Word(A)", "Operator({)", "CloseBracket(})", "Word(A)", "Operator(*)", "Word(^R)", "Semicolon(;)", "CloseBracket(})"}},
		{"sub A; package Foo { 1 } A *^R;", glob("Word(sub)", "Word(A)", "Semicolon(;)", "Word(package)", "Word(Foo)", "Operator({)", "Number(1)", "CloseBracket(})", "Word(A)")},
		// Elsewhere it is undeclared: syntax error near "*^" on each.
		mul("package Foo; sub A; package main; A *^R;", "Word(package)", "Word(Foo)", "Semicolon(;)", "Word(sub)", "Word(A)", "Semicolon(;)", "Word(package)", "Word(main)", "Semicolon(;)", "Word(A)"),
		mul("sub A; package Foo; A *^R;", "Word(sub)", "Word(A)", "Semicolon(;)", "Word(package)", "Word(Foo)", "Semicolon(;)", "Word(A)"),
		mul("package Foo 1.0 { sub A {} } A *^R;", "Word(package)", "Word(Foo)", "Number(1.0)", "Operator({)", "Word(sub)", "Word(A)", "Operator({)", "CloseBracket(})", "CloseBracket(})", "Word(A)"),
		// shortcut (wordTakesTerm): a sub declared other than by a `sub NAME`
		// the lexer has read is undeclared here, so the star multiplies. perl
		// -c: syntax OK on every row, which the chart and RD refuse.
		mul("use POSIX qw(floor); floor *^R;", "Word(use)", "Word(POSIX)", "Quote(qw(floor))", "Semicolon(;)", "Word(floor)"),
		mul("use subs 'A'; A *^R;", "Word(use)", "Word(subs)", "Quote('A')", "Semicolon(;)", "Word(A)"),
		mul("use Scalar::Util qw(blessed); blessed *^R;", "Word(use)", "Word(Scalar::Util)", "Quote(qw(blessed))", "Semicolon(;)", "Word(blessed)"),
		mul("BEGIN { *A = sub {} } A *^R;", "Word(BEGIN)", "Operator({)", "Operator(*)", "Word(A)", "Operator(=)", "Word(sub)", "Operator({)", "CloseBracket(})", "CloseBracket(})", "Word(A)"),
		mul(`BEGIN { eval "sub A {}" } A *^R;`, "Word(BEGIN)", "Operator({)", "Word(eval)", `Quote("sub A {}")`, "CloseBracket(})", "Word(A)"),
		{"use POSIX qw(floor); floor *{^R};", []string{"Word(use)", "Word(POSIX)", "Quote(qw(floor))", "Semicolon(;)", "Word(floor)", "Operator(*)", "Operator({)", "Operator(^)", "Word(R)", "CloseBracket(})", "Semicolon(;)"}},
		{"sub A; A *{^R};", []string{"Word(sub)", "Word(A)", "Semicolon(;)", "Word(A)", "Operator(*)", "Word({^R})", "Semicolon(;)"}},
	})
}

// TestGlobCaretNameOtherSigils: every sigil reads the caret name the same
// way, since each reaches scan_ident. perl -c: syntax OK on every row but
// the three refusals marked.
func TestGlobCaretNameOtherSigils(t *testing.T) {
	checkCaretRows(t, []caretRow{
		{"$^?;", []string{"Variable($^?)", "Semicolon(;)"}},
		{"$^];", []string{"Variable($^])", "Semicolon(;)"}},
		{`$^\;`, []string{`Variable($^\)`, "Semicolon(;)"}},
		{"@^?;", []string{"Variable(@^?)", "Semicolon(;)"}},
		{"%^];", []string{"Variable(%^])", "Semicolon(;)"}},
		{"$ ^R;", []string{"Variable($ ^R)", "Semicolon(;)"}},
		{"@ ^R;", []string{"Variable(@ ^R)", "Semicolon(;)"}},
		{"% ^R;", []string{"Variable(% ^R)", "Semicolon(;)"}},
		{"&^R;", []string{"FuncSigil(&)", "Word(^R)", "Semicolon(;)"}},
		{"&^R();", []string{"FuncSigil(&)", "Word(^R)", "Operator(()", "CloseBracket())", "Semicolon(;)"}},
		{"& ^R;", []string{"FuncSigil(&)", "Word(^R)", "Semicolon(;)"}},
		{"&^?;", []string{"FuncSigil(&)", "Word(^?)", "Semicolon(;)"}},
		{"&^;", []string{"FuncSigil(&)", "Word(^)", "Semicolon(;)"}},
		{"&{^R};", []string{"FuncSigil(&)", "Word({^R})", "Semicolon(;)"}},
		{"&{ ^R };", []string{"FuncSigil(&)", "Word({ ^R })", "Semicolon(;)"}},
		// Refused, Bareword found before "r": the sub is &^.
		{"&^r;", []string{"FuncSigil(&)", "Word(^)", "Word(r)", "Semicolon(;)"}},
		// Refused, syntax error near "{^": no braced name.
		{"${^r};", []string{"DerefSigil($)", "Operator({)", "Operator(^)", "Word(r)", "CloseBracket(})", "Semicolon(;)"}},
		{"${^1};", []string{"DerefSigil($)", "Operator({)", "Operator(^)", "Number(1)", "CloseBracket(})", "Semicolon(;)"}},
		// Refused ("$# is no longer supported"): -DT shows $# then ^R, as
		// toke.c's `$#` test takes no caret.
		{"$#^R;", []string{"Variable($#)", "Operator(^)", "Word(R)", "Semicolon(;)"}},
	})
}
