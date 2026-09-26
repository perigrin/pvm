// ABOUTME: Five scanner rows: literal ranges, qx, <<>>, angle globs, leading ::.
// ABOUTME: Each produced plausible wrong tokens rather than an error, which is the harder kind to notice.

package lexer_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/lexer"
)

// significant renders a source's tokens, skipping whitespace, as
// `Kind("text")` so a stream can be compared as one string.
func significant(src string) []string {
	var out []string
	for _, tok := range lexer.Tokenize([]byte(src)) {
		if tok.Kind == lexer.Whitespace {
			continue
		}
		out = append(out, tok.Kind.String()+"("+src[tok.Start:tok.End]+")")
	}
	return out
}

func streamIs(src string, want ...string) bool {
	got := significant(src)
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

// TestNumericRangeIsNotOneToken: `1..5` is three tokens, not one number.
//
// `scanNumber` consumed `isWordByte` or `.` in a loop, so a literal range
// arrived as `Number("1..5")` and the range operator was gone before the
// parser saw it. `$a..$b` was fine, so only literal endpoints were affected.
//
// The cost is measured in §4.14.2: `my @r = (1..5)` infers `@r` as `Str`,
// because `findOperatorText` (`internal/infer/infer.go:645-659`) matches the
// anonymous `.` before ever seeing `..`, while `signatures.go:182` has
// `"..": {Result: List}` waiting for a token that never arrives. perl:
// `$r[0]+1` is 2, so the elements are numbers.
func TestNumericRangeIsNotOneToken(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		{"1..5", []string{"Number(1)", "Operator(..)", "Number(5)"}},
		{"1...5", []string{"Number(1)", "Operator(...)", "Number(5)"}},
		{"0..9", []string{"Number(0)", "Operator(..)", "Number(9)"}},
	} {
		if !streamIs(c.src, c.want...) {
			t.Errorf("%q lexes as %v, want %v",
				c.src, significant(c.src), c.want)
		}
	}
}

// TestNumberFormsStillLex is the negative scenario, and the one that makes
// the range fix delicate: a `.` inside a number is part of it, and so is
// every letter in `0x1f` and `1e10`.
//
// Measured on perl 5.42.0:
//
//	$ perl -MO=Deparse -e 'my @n = (1.5, 1_000, 0x1f, 1e10, 0b11, 017);'
//	my(@n) = (1.5, 1000, 31, 10000000000.0, 3, 15);
//
// Six literals, six values. A range fix that split any of these would turn a
// number into an expression.
func TestNumberFormsStillLex(t *testing.T) {
	for _, src := range []string{
		"1.5", "1_000", "0x1f", "1e10", "0b11", "017",
		"1.5e10", "0xdead_beef", "1_000.5",
	} {
		got := significant(src)
		if len(got) != 1 {
			t.Errorf("%q lexes as %v, want one token", src, got)
		}
	}
}

// TestLeadingDecimalLiteral: a numeric literal with no digit before the point
// is ONE number, not an operator and a number.
//
// Measured: `perl -e 'my $n = .5; print $n'` prints 0.5, so `.5` is one
// number. `scanNumber` required a leading DIGIT, so the dot was emitted as
// an operator first and the parser then had no term to start on.
//
// This test was parked as a skip by the range-splitting commit, which was
// right to park it: letting `scanNumber` start on a dot collides with
// concatenation and with `..`, and deciding that inside the range fix
// would have made one number answer for two questions. Issue 01a0c13f is
// that decision, and the rule it settles on is POSITION -- a `.` before a
// digit starts a number only where a TERM is expected, which is the same
// expect-state mechanism the spec uses for a leading `%`.
//
// `TestDotIsStillConcatenation` is the other half and must stay
// green: in operator position the identical bytes are concatenation.
func TestLeadingDecimalLiteral(t *testing.T) {
	for _, c := range []struct {
		src  string
		want string
	}{
		{".5", "Number(.5)"},
		{".25", "Number(.25)"},
		{".5e10", "Number(.5e10)"},
	} {
		got := significant(c.src)
		if len(got) != 1 || got[0] != c.want {
			t.Errorf("%q lexes as %v, want [%s]", c.src, got, c.want)
		}
	}
}

// TestDotIsStillConcatenation is the other half of the rule, and
// the half that makes it a POSITION question rather than a lookahead one.
//
// In OPERATOR position -- just after a term -- a `.` is concatenation
// even when a digit follows, and perl proves the two readings are both
// live. Measured 5.42.0, the same four bytes parse two ways depending on
// nothing but what came before:
//
//	$ perl -MO=Deparse -e 'my $a = "x"; my $r = $a .5;'
//	my $r = $a . '5';                 CONCATENATION
//
//	$ perl -MO=Deparse -e 'my $a = "x"; print $a .5'
//	print $a 0.5;                     FILEHANDLE and a NUMBER
//
// The second is why `print $a .5` prints nothing: `$a` lands in print's
// filehandle slot and `.5` is its argument. That is a runtime failure
// from a lexical decision, and it is exactly the class of bug the expect
// state exists to get right.
//
// So a lexer that scanned `.` as a number wherever a digit follows would
// break the first line, and one that never did would break `.5` alone.
// Only position separates them.
func TestDotIsStillConcatenation(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		{`$a . 5`, []string{"Variable($a)", "Operator(.)", "Number(5)"}},
		{`$a .5`, []string{"Variable($a)", "Operator(.)", "Number(5)"}},
		{`$a.5`, []string{"Variable($a)", "Operator(.)", "Number(5)"}},
	} {
		got := significant(c.src)
		if len(got) != len(c.want) {
			t.Errorf("%q lexes as %v, want %v", c.src, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%q lexes as %v, want %v", c.src, got, c.want)
				break
			}
		}
	}
}

// TestLeadingDotVersionStringUnharmed pins the two regressions the
// leading-decimal rule caused, each found by a test rather than foreseen.
//
// A v-string reaches `scanNumber` looking exactly like a leading decimal:
// term expected, a dot, a digit after it. `v5` lexes as an ordinary Word,
// so nothing upstream marks the dots as part of a version.
//
// `use v5.36` is the consequential one. Merging its `.36` into a Number
// leaves `pendingVersionMajor` without the Operator it reassembles the
// version from, the signatures feature never turns on, and `sub g ($a,
// $b)` silently becomes a PROTOTYPE -- a lexical change that alters what
// the rest of the file means.
//
// `require(v5.5.630)` needs a second rule, because by its third part the
// pending major is already spent. Both are checked here so a future
// change to either rule fails on the case it breaks.
func TestLeadingDotVersionStringUnharmed(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		{`use v5.36;`, []string{
			"Word(use)", "Word(v5)", "Operator(.)", "Number(36)", "Semicolon(;)",
		}},
		// `v5.5.630` is now ONE token, and that is the v-string rule of
		// 01a0db78 rather than a loss: two dots make a string, perl agrees
		// (`length(v5.5.630)` is 3), and `scanVString` takes it whole.
		//
		// What this row pinned was that the FIRST dot stay a separator the
		// version reassembly could read. Nothing is reassembled from one
		// token, so `noteSignatures` reads the version out of the Quote
		// directly -- and `TestVStringVersionStillEnablesSignatures` is
		// where that is checked, in both directions. The claim moved; it
		// was not dropped.
		//
		// The FIRST-dot guard this row was added for is still live for the
		// ONE-dot form, which `use v5.36;` above is: `scanVString` declines
		// on one dot, and without `continuesVersionString` the `.36` would
		// merge into `Word(v5) Number(.36)`.
		{`v5.5.630`, []string{
			"Quote(v5.5.630)",
		}},
	} {
		got := significant(c.src)
		if len(got) != len(c.want) {
			t.Errorf("%q lexes as %v, want %v", c.src, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%q lexes as %v, want %v", c.src, got, c.want)
				break
			}
		}
	}
}

// TestLeadingDotRangeUnharmed keeps the range fix this file's first test
// records from being undone: `..` is the range operator wherever it
// appears, and a `.` that starts a number must not swallow the second
// dot. `1..5` and `.5..1` are both three tokens.
func TestLeadingDotRangeUnharmed(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		{`1..5`, []string{"Number(1)", "Operator(..)", "Number(5)"}},
		{`$a..5`, []string{"Variable($a)", "Operator(..)", "Number(5)"}},
	} {
		got := significant(c.src)
		if len(got) != len(c.want) {
			t.Errorf("%q lexes as %v, want %v", c.src, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%q lexes as %v, want %v", c.src, got, c.want)
				break
			}
		}
	}
}

// TestQxLexes: `qx//` is a quote-like operator, and `internal/lexer/quote.go`
// had rows for q, qq, qw, qr, m, s, tr and y but not qx.
//
// §4.14.2 records that a backtick string already lexes as one Quote whose
// Text keeps its delimiters, and that B::SoN calls the form `BacktickExpr` --
// a language fact. `qx//` is the same node reached by a different spelling.
func TestQxLexes(t *testing.T) {
	for _, src := range []string{
		"qx/ls/", "qx{ls}", "qx(ls)", "qx[ls]", "qx!ls!", "qx'ls'",
	} {
		got := significant(src)
		if len(got) != 1 || !strings.HasPrefix(got[0], "Quote(") {
			t.Errorf("%q lexes as %v, want one Quote", src, got)
		}
	}
}

// TestQxIsStillAName is qx's negative: a longer word beginning with `qx` is
// an identifier, the same rule every other quote-op keyword follows.
func TestQxIsStillAName(t *testing.T) {
	for _, src := range []string{
		"$qx", "qxray", "my $qx = 1;", "f(qx => 1);",
	} {
		for _, tok := range lexer.Tokenize([]byte(src)) {
			if tok.Kind == lexer.UnknownRest || tok.Kind == lexer.Error {
				t.Errorf("%q: %v %q -- qx followed by a word character is a name",
					src, tok.Kind, src[tok.Start:tok.End])
			}
		}
	}
}

// TestDoubleDiamond: `<<>>` is the 5.22 readline that treats every argument
// as a filename, never as a command.
//
// It lexed as two shift operators, so a `while (<<>>)` loop read as an
// expression that could not parse.
func TestDoubleDiamond(t *testing.T) {
	got := significant("<<>>")
	if len(got) != 1 || !strings.HasPrefix(got[0], "Readline(") {
		t.Errorf("`<<>>` lexes as %v, want one Readline", got)
	}
}

// TestShiftStillShifts is the double diamond's negative. `<<` in OPERATOR
// position is a left shift, and `<<EOF` is a heredoc; neither may become a
// readline.
func TestShiftStillShifts(t *testing.T) {
	for _, c := range []struct{ src, want string }{
		{"$a << $c", "Operator(<<)"},
		{"$a >> $c", "Operator(>>)"},
	} {
		var found bool
		for _, g := range significant(c.src) {
			if g == c.want {
				found = true
			}
		}
		if !found {
			t.Errorf("%q lexes as %v, want a %s", c.src, significant(c.src), c.want)
		}
	}
	// A heredoc opener must still open a heredoc.
	if got := significant("<<EOF"); len(got) == 0 ||
		!strings.HasPrefix(got[0], "HeredocOpen(") {
		t.Errorf("`<<EOF` lexes as %v, want a HeredocOpen", got)
	}
}

// TestAngleGlob: `<*.c>` is a glob, and `scanAngle` accepted only word
// bytes, `$` and `:` between the angles -- so a `*` ended the candidate and
// the whole form fell through to five separate operators.
func TestAngleGlob(t *testing.T) {
	for _, src := range []string{"<*.c>", "<*>", "<~/*.txt>"} {
		got := significant(src)
		if len(got) != 1 {
			t.Errorf("%q lexes as %v, want one token", src, got)
		}
	}
}

// TestComparisonStillCompares is the glob's negative, and it is the sharper
// half: `<` in operator position is a comparison and must stay one.
func TestComparisonStillCompares(t *testing.T) {
	for _, c := range []struct{ src, want string }{
		{"$a < $c", "Operator(<)"},
		{"$a > $c", "Operator(>)"},
		{"$a <= $c", "Operator(<=)"},
		{"$a <=> $c", "Operator(<=>)"},
	} {
		var found bool
		for _, g := range significant(c.src) {
			if g == c.want {
				found = true
			}
		}
		if !found {
			t.Errorf("%q lexes as %v, want a %s", c.src, significant(c.src), c.want)
		}
	}
	// A readline must still lex: the glob fix must widen, not replace.
	for _, src := range []string{"<STDIN>", "<$fh>", "<>"} {
		got := significant(src)
		if len(got) != 1 || !strings.HasPrefix(got[0], "Readline(") {
			t.Errorf("%q lexes as %v, want one Readline", src, got)
		}
	}
}

// TestLeadingPackageSeparator: `$::x` is `$main::x`.
//
// `scanVarName` reached its identifier scanner, which requires a leading
// identifier byte, and `:` is not one -- so `$::x` lexed as
// `Variable("$:")`, a punctuation variable named `:`, followed by stray
// tokens. That is a plausible wrong answer rather than an error, which is
// why nothing caught it.
//
// Measured: `our $x = 5; print $::x` prints 5.
func TestLeadingPackageSeparator(t *testing.T) {
	for _, src := range []string{
		"$::x", "@::a", "%::h", "$::Foo::x",
	} {
		got := significant(src)
		if len(got) != 1 {
			t.Errorf("%q lexes as %v, want one token", src, got)
		}
	}

	// `&::f` is NOT one token, and that is the existing design rather than a
	// gap here: the lexer emits `&` as its own FuncSigil so that `&f` and
	// `&&` stay distinguishable, and `internal/parse/term.go:97` joins the
	// two. `&foo` is two tokens for the same reason. What matters is that
	// the name after the sigil carries its package separator.
	if got := significant("&::f"); len(got) != 3 ||
		!strings.HasSuffix(got[len(got)-1], "(f)") {
		t.Errorf("`&::f` lexes as %v; the sigil is its own token by design, "+
			"and the name must follow it", got)
	}
}

// TestPunctuationVariablesStillLexHere is the leading-`::` negative. `$:` is
// a real punctuation variable -- the format line-break characters -- and
// widening the name scanner must not swallow what follows it.
func TestPunctuationVariablesStillLexHere(t *testing.T) {
	for _, c := range []struct {
		src  string
		want []string
	}{
		{"$: = 1", []string{"Variable($:)", "Operator(=)", "Number(1)"}},
		{"$:", []string{"Variable($:)"}},
	} {
		if !streamIs(c.src, c.want...) {
			t.Errorf("%q lexes as %v, want %v",
				c.src, significant(c.src), c.want)
		}
	}
}
