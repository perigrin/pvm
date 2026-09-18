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

// TestLeadingDotNumber records a form perl accepts and this lexer splits.
//
// Measured: `perl -e 'my $n = .5; print $n'` prints 0.5, so `.5` is one
// number. `scanNumber` requires a leading DIGIT (`scan.go:127`), so the dot
// is emitted as an operator first.
//
// Skipped rather than asserted: fixing it means letting scanNumber start on
// `.`, which collides with concatenation (`$a . $b`) and with the range this
// commit is splitting. That is its own decision and its own measurement,
// and doing it inside the range fix would make one number answer for two.
//
// Recorded here so the next person measures it rather than rediscovering it.
func TestLeadingDotNumber(t *testing.T) {
	t.Skip("`.5` splits into Operator(.) and Number(5); needs scanNumber to " +
		"start on a dot, which collides with concatenation and with `..` -- " +
		"its own issue, not part of the range fix")
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
