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
